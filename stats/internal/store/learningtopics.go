package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
)

// MaxLearningTopicShows is how many times one topic may be recorded before
// RecordLearningTopic refuses it (cap-in-store).
const MaxLearningTopicShows = 3

// ErrInvalidLearningTopic is returned by RecordLearningTopic when the topic
// is not a lowercase kebab slug of at most 64 bytes, or the note is empty
// (topic-slug).
var ErrInvalidLearningTopic = errors.New("store: invalid learning topic")

// ErrLearningTopicExhausted is returned by RecordLearningTopic when the
// topic has already been recorded MaxLearningTopicShows times.
var ErrLearningTopicExhausted = errors.New("store: learning topic exhausted")

var learningTopicSlug = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// LearningTopic is one learning note shown by a run: its topic slug, a
// one-line note, and the change and project it was shown in
// (topics-in-flowd).
type LearningTopic struct {
	Topic      string
	Note       string
	Change     string
	ProjectKey string
}

// LearningTopicCount is one row of ListLearningTopics: a topic, how many
// times it was recorded, and its newest note.
type LearningTopicCount struct {
	Topic    string
	Count    int
	LastNote string
}

// RecordLearningTopic records one showing of in.Topic. It returns
// ErrInvalidLearningTopic for a malformed slug or an empty note, and
// ErrLearningTopicExhausted -- writing nothing -- when the topic is
// already at MaxLearningTopicShows. The count and the insert run in one
// transaction under a per-topic advisory lock, so two concurrent records
// cannot both pass the cap.
func (s *Store) RecordLearningTopic(ctx context.Context, in LearningTopic) error {
	if len(in.Topic) > 64 || !learningTopicSlug.MatchString(in.Topic) {
		return fmt.Errorf("%w: %q", ErrInvalidLearningTopic, in.Topic)
	}
	if in.Note == "" {
		return fmt.Errorf("%w: empty note", ErrInvalidLearningTopic)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("store: begin RecordLearningTopic: %w", err)
	}
	// Rollback after a successful Commit is a documented no-op in pgx.
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, in.Topic); err != nil {
		return fmt.Errorf("store: record learning topic: lock: %w", err)
	}
	var count int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM flow_learning_topics WHERE topic = $1
	`, in.Topic).Scan(&count); err != nil {
		return fmt.Errorf("store: record learning topic: count: %w", err)
	}
	if count >= MaxLearningTopicShows {
		return fmt.Errorf("%w: %q shown %d times", ErrLearningTopicExhausted, in.Topic, count)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO flow_learning_topics (topic, note, change, project_key)
		VALUES ($1, $2, $3, $4)
	`, in.Topic, in.Note, in.Change, in.ProjectKey); err != nil {
		return fmt.Errorf("store: record learning topic: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("store: commit RecordLearningTopic: %w", err)
	}
	return nil
}

// ListLearningTopics returns every recorded topic once, ordered by topic,
// with its count and newest note -- what the composing session reads to
// skip exhausted topics before choosing one.
func (s *Store) ListLearningTopics(ctx context.Context) ([]LearningTopicCount, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT topic, count(*),
			(array_agg(note ORDER BY created_at DESC, id DESC))[1]
		FROM flow_learning_topics
		GROUP BY topic
		ORDER BY topic
	`)
	if err != nil {
		return nil, fmt.Errorf("store: list learning topics: %w", err)
	}
	defer rows.Close()

	out := []LearningTopicCount{}
	for rows.Next() {
		var c LearningTopicCount
		if err := rows.Scan(&c.Topic, &c.Count, &c.LastNote); err != nil {
			return nil, fmt.Errorf("store: list learning topics: scan: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list learning topics: %w", err)
	}
	return out, nil
}
