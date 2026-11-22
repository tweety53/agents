package store

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrConfidenceEstimateInvalid reports a RecordConfidenceEstimate the rules
// of validateConfidenceEstimate refuse. The handler maps it to 400: a
// caller's mistake, never a store failure.
var ErrConfidenceEstimateInvalid = errors.New("store: invalid confidence estimate")

var confidenceTag = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// validateConfidenceEstimate is the boundary: a score of 1-99 (100 has no
// gap and is not kept), at least one non-blank gap and one kebab-case tag, a
// plan row with a disposition and no task, an implement row with a task and
// no disposition.
func validateConfidenceEstimate(e records.ConfidenceEstimate) error {
	bad := func(format string, a ...any) error {
		return fmt.Errorf("%w: "+format, append([]any{ErrConfidenceEstimateInvalid}, a...)...)
	}
	if e.Change == "" {
		return bad("change is required")
	}
	if e.Score < 1 || e.Score > 99 {
		return bad("score must be 1-99, got %d", e.Score)
	}
	if len(e.Gaps) == 0 {
		return bad("at least one gap is required")
	}
	for _, g := range e.Gaps {
		if strings.TrimSpace(g) == "" {
			return bad("a gap is blank")
		}
	}
	if len(e.Tags) == 0 {
		return bad("at least one tag is required")
	}
	for _, t := range e.Tags {
		if !confidenceTag.MatchString(t) {
			return bad("tag %q is not lowercase kebab-case", t)
		}
	}
	switch e.Stage {
	case "plan":
		if e.Task != "" || (e.Disposition != "raised" && e.Disposition != "later") {
			return bad("a plan estimate carries no task and a disposition of raised or later, got task %q, disposition %q", e.Task, e.Disposition)
		}
	case "implement":
		if e.Task == "" || e.Disposition != "" {
			return bad("an implement estimate carries a task and no disposition, got task %q, disposition %q", e.Task, e.Disposition)
		}
	default:
		return bad("stage must be plan or implement, got %q", e.Stage)
	}
	return nil
}

// RecordConfidenceEstimate validates and records one confidence estimate.
func (s *Store) RecordConfidenceEstimate(ctx context.Context, projectKey string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error) {
	if err := validateConfidenceEstimate(in); err != nil {
		return records.ConfidenceEstimate{}, err
	}
	out := in
	out.Project = projectKey
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO confidence_estimates (project_key, change, stage, task, score, gaps, tags, disposition)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, recorded_at
	`, projectKey, in.Change, in.Stage, nullIfEmpty(in.Task), in.Score, in.Gaps, in.Tags, nullIfEmpty(in.Disposition),
	).Scan(&out.ID, &out.RecordedAt); err != nil {
		return records.ConfidenceEstimate{}, fmt.Errorf("store: record confidence estimate for %s/%s: %w", projectKey, in.Change, err)
	}
	return out, nil
}

// ConfidenceTags lists every tag in use across all projects, most used
// first -- the vocabulary a filing step reuses before it adds a tag.
func (s *Store) ConfidenceTags(ctx context.Context) ([]records.ConfidenceTag, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t, count(*) FROM confidence_estimates, unnest(tags) AS t
		GROUP BY t ORDER BY count(*) DESC, t
	`)
	if err != nil {
		return nil, fmt.Errorf("store: confidence tags: %w", err)
	}
	defer rows.Close()
	out := []records.ConfidenceTag{}
	for rows.Next() {
		var t records.ConfidenceTag
		if err := rows.Scan(&t.Tag, &t.Count); err != nil {
			return nil, fmt.Errorf("store: confidence tags: scan: %w", err)
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ConfidenceEstimatesPage reads one page of the estimates across all
// projects, newest first, restricted to those carrying tag unless it is
// empty, and how many rows that list holds in all.
func (s *Store) ConfidenceEstimatesPage(ctx context.Context, tag string, limit, offset int) ([]records.ConfidenceEstimate, int, error) {
	var total int
	if err := s.pool.QueryRow(ctx, `SELECT count(*) FROM confidence_estimates WHERE $1 = '' OR $1 = ANY(tags)`, tag).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store: count confidence estimates: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, project_key, change, stage, COALESCE(task, ''), score, gaps, tags, COALESCE(disposition, ''), recorded_at
		FROM confidence_estimates
		WHERE $1 = '' OR $1 = ANY(tags)
		ORDER BY recorded_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, tag, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store: confidence estimates page: %w", err)
	}
	defer rows.Close()
	out := []records.ConfidenceEstimate{}
	for rows.Next() {
		var e records.ConfidenceEstimate
		if err := rows.Scan(&e.ID, &e.Project, &e.Change, &e.Stage, &e.Task, &e.Score, &e.Gaps, &e.Tags, &e.Disposition, &e.RecordedAt); err != nil {
			return nil, 0, fmt.Errorf("store: confidence estimates page: scan: %w", err)
		}
		out = append(out, e)
	}
	return out, total, rows.Err()
}
