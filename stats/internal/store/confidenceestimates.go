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

// RecordConfidenceEstimate validates and records one confidence estimate. Its
// review status is always to-review: whatever the caller sent is not read.
func (s *Store) RecordConfidenceEstimate(ctx context.Context, projectKey string, in records.ConfidenceEstimate) (records.ConfidenceEstimate, error) {
	if err := validateConfidenceEstimate(in); err != nil {
		return records.ConfidenceEstimate{}, err
	}
	out := in
	out.Project, out.Reason = projectKey, ""
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO confidence_estimates (project_key, change, stage, task, score, gaps, tags, disposition)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, recorded_at, review_status
	`, projectKey, in.Change, in.Stage, nullIfEmpty(in.Task), in.Score, in.Gaps, in.Tags, nullIfEmpty(in.Disposition),
	).Scan(&out.ID, &out.RecordedAt, &out.ReviewStatus); err != nil {
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

// ErrConfidenceEstimateNotFound reports a review of an estimate id the
// project has no row for; the handler maps it to 404.
var ErrConfidenceEstimateNotFound = errors.New("store: no such confidence estimate")

func validConfidenceReviewStatus(status string) error {
	switch status {
	case "to-review", "approved", "declined":
		return nil
	}
	return fmt.Errorf("%w: review status must be to-review, approved or declined, got %q", ErrConfidenceEstimateInvalid, status)
}

// ReviewConfidenceEstimate sets an estimate's review status: to-review,
// approved or declined, with the operator's optional reason (empty clears
// it). Nothing else about the row changes.
func (s *Store) ReviewConfidenceEstimate(ctx context.Context, projectKey string, id int64, status, reason string) error {
	if err := validConfidenceReviewStatus(status); err != nil {
		return err
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE confidence_estimates SET review_status = $3, review_reason = NULLIF($4, '')
		WHERE project_key = $1 AND id = $2
	`, projectKey, id, status, reason)
	if err != nil {
		return fmt.Errorf("store: review confidence estimate %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %d", ErrConfidenceEstimateNotFound, id)
	}
	return nil
}

// ConfidencePage is one page of the Confidence dashboard: the rows carrying
// the chosen review status, how many rows that list holds in all, and the
// count of every status (for the status chips).
type ConfidencePage struct {
	Rows                         []records.ConfidenceEstimate
	Total                        int
	ToReview, Approved, Declined int
}

// ConfidenceEstimatesPage reads one page of the estimates across all
// projects carrying the review status, newest first. Only limit rows are read.
func (s *Store) ConfidenceEstimatesPage(ctx context.Context, status string, limit, offset int) (ConfidencePage, error) {
	var out ConfidencePage
	if err := validConfidenceReviewStatus(status); err != nil {
		return out, err
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT
		  count(*) FILTER (WHERE review_status = $1),
		  count(*) FILTER (WHERE review_status = 'to-review'),
		  count(*) FILTER (WHERE review_status = 'approved'),
		  count(*) FILTER (WHERE review_status = 'declined')
		FROM confidence_estimates
	`, status).Scan(&out.Total, &out.ToReview, &out.Approved, &out.Declined); err != nil {
		return out, fmt.Errorf("store: count confidence estimates: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, project_key, change, stage, COALESCE(task, ''), score, gaps, tags, COALESCE(disposition, ''), recorded_at, review_status, COALESCE(review_reason, '')
		FROM confidence_estimates
		WHERE review_status = $1
		ORDER BY recorded_at DESC, id DESC
		LIMIT $2 OFFSET $3
	`, status, limit, offset)
	if err != nil {
		return out, fmt.Errorf("store: confidence estimates page: %w", err)
	}
	defer rows.Close()
	out.Rows = []records.ConfidenceEstimate{}
	for rows.Next() {
		var e records.ConfidenceEstimate
		if err := rows.Scan(&e.ID, &e.Project, &e.Change, &e.Stage, &e.Task, &e.Score, &e.Gaps, &e.Tags, &e.Disposition, &e.RecordedAt, &e.ReviewStatus, &e.Reason); err != nil {
			return out, fmt.Errorf("store: confidence estimates page: scan: %w", err)
		}
		out.Rows = append(out.Rows, e)
	}
	return out, rows.Err()
}
