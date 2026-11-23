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

// validateConfidenceEstimate is the boundary: a score of 1-100; below 100 at least one non-blank gap and one
// kebab-case tag, at 100 neither (it has nothing to close); a
// plan row with a disposition and no task, an implement row with a task and
// no disposition.
func validateConfidenceEstimate(e records.ConfidenceEstimate) error {
	bad := func(format string, a ...any) error {
		return fmt.Errorf("%w: "+format, append([]any{ErrConfidenceEstimateInvalid}, a...)...)
	}
	if e.Change == "" {
		return bad("change is required")
	}
	if e.Score < 1 || e.Score > 100 {
		return bad("score must be 1-100, got %d", e.Score)
	}
	if e.Score == 100 && (len(e.Gaps) > 0 || len(e.Tags) > 0) {
		return bad("a 100%% estimate carries no gap and no tag")
	}
	if e.Score < 100 && len(e.Gaps) == 0 {
		return bad("at least one gap is required")
	}
	for _, g := range e.Gaps {
		if strings.TrimSpace(g) == "" {
			return bad("a gap is blank")
		}
	}
	if e.Score < 100 && len(e.Tags) == 0 {
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
	if out.Gaps == nil {
		out.Gaps = []string{}
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	in = out
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

// ConfidenceTags lists every tag in use across all projects, alphabetically (a popularity order makes the
// top tag the one an agent takes without reading) -- the vocabulary a filing step reuses before it adds a tag.
func (s *Store) ConfidenceTags(ctx context.Context) ([]records.ConfidenceTag, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t, count(*) FROM confidence_estimates, unnest(tags) AS t
		GROUP BY t ORDER BY t
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

// ConfidencePage is one page of the Confidence dashboard (estimates below 100%): the rows carrying
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
		FROM confidence_estimates WHERE score < 100
	`, status).Scan(&out.Total, &out.ToReview, &out.Approved, &out.Declined); err != nil {
		return out, fmt.Errorf("store: count confidence estimates: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT id, project_key, change, stage, COALESCE(task, ''), score, score_after, gaps_after, gaps, tags, COALESCE(disposition, ''), recorded_at, review_status, COALESCE(review_reason, '')
		FROM confidence_estimates
		WHERE review_status = $1 AND score < 100
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
		if err := rows.Scan(&e.ID, &e.Project, &e.Change, &e.Stage, &e.Task, &e.Score, &e.ScoreAfter, &e.GapsAfter, &e.Gaps, &e.Tags, &e.Disposition, &e.RecordedAt, &e.ReviewStatus, &e.Reason); err != nil {
			return out, fmt.Errorf("store: confidence estimates page: scan: %w", err)
		}
		out.Rows = append(out.Rows, e)
	}
	return out, rows.Err()
}

// ConfidenceEstimatesForChange reads one change's estimates, oldest first,
// whatever their review status -- what the run page shows.
func (s *Store) ConfidenceEstimatesForChange(ctx context.Context, projectKey, change string) ([]records.ConfidenceEstimate, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, project_key, change, stage, COALESCE(task, ''), score, score_after, gaps_after, gaps, tags, COALESCE(disposition, ''), recorded_at, review_status, COALESCE(review_reason, '')
		FROM confidence_estimates
		WHERE project_key = $1 AND change = $2
		ORDER BY recorded_at, id
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: confidence estimates for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()
	out := []records.ConfidenceEstimate{}
	for rows.Next() {
		var e records.ConfidenceEstimate
		if err := rows.Scan(&e.ID, &e.Project, &e.Change, &e.Stage, &e.Task, &e.Score, &e.ScoreAfter, &e.GapsAfter, &e.Gaps, &e.Tags, &e.Disposition, &e.RecordedAt, &e.ReviewStatus, &e.Reason); err != nil {
			return nil, fmt.Errorf("store: confidence estimates for %s/%s: scan: %w", projectKey, change, err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// SetConfidenceAfter sets score_after on the latest implement estimate of project/change whose task is task:
// what the implementer held once its tests were green, with the gaps still open below 100%. ErrConfidenceEstimateNotFound when the group filed none.
func (s *Store) SetConfidenceAfter(ctx context.Context, projectKey, change, task string, score int, gaps []string) error {
	bad := func(format string, a ...any) error {
		return fmt.Errorf("%w: "+format, append([]any{ErrConfidenceEstimateInvalid}, a...)...)
	}
	switch {
	case score < 1 || score > 100:
		return bad("score must be 1-100, got %d", score)
	case score == 100 && len(gaps) > 0:
		return bad("a 100%% estimate carries no gap")
	case score < 100 && len(gaps) == 0:
		return bad("at least one gap is required below 100%%: what would close it")
	}
	for _, g := range gaps {
		if strings.TrimSpace(g) == "" {
			return bad("a gap is blank")
		}
	}
	if gaps == nil {
		gaps = []string{}
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE confidence_estimates SET score_after = $4, gaps_after = $5
		WHERE id = (SELECT max(id) FROM confidence_estimates WHERE project_key = $1 AND change = $2 AND stage = 'implement' AND task = $3)
	`, projectKey, change, task, score, gaps)
	if err != nil {
		return fmt.Errorf("store: set confidence after for %s/%s task %s: %w", projectKey, change, task, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %s/%s task %s", ErrConfidenceEstimateNotFound, projectKey, change, task)
	}
	return nil
}
