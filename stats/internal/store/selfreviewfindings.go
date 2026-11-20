package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrSelfReviewFindingInvalid reports a RecordSelfReviewFinding whose
// disposition is outside fixed/filed/declined/noted/suggested, whose ref does not fit its
// disposition, whose change, angle or note is empty, whose blast radius
// or is negative. The handler maps it
// to 400: a caller's mistake, never a store failure.
var ErrSelfReviewFindingInvalid = errors.New("store: invalid self-review finding")

// validateSelfReviewFinding is the boundary for the rules the table's CHECK
// cannot state: which ref shape each disposition carries.
func validateSelfReviewFinding(f records.SelfReviewFinding) error {
	if f.Change == "" || f.Angle == "" || f.Note == "" {
		return fmt.Errorf("%w: change, angle and note are required", ErrSelfReviewFindingInvalid)
	}
	switch f.Severity {
	case "", "critical", "important", "minor":
	default:
		return fmt.Errorf("%w: severity must be critical, important or minor, got %q", ErrSelfReviewFindingInvalid, f.Severity)
	}
	if f.BlastRadius != nil && *f.BlastRadius < 0 {
		return fmt.Errorf("%w: blast radius must be 0 or more, got %d", ErrSelfReviewFindingInvalid, *f.BlastRadius)
	}
	switch f.Disposition {
	case "fixed":
		if !records.SelfReviewSha.MatchString(f.Ref) {
			return fmt.Errorf("%w: fixed needs a 7-40 character lowercase hex sha as ref, got %q", ErrSelfReviewFindingInvalid, f.Ref)
		}
	case "filed":
		if !records.IssueKey.MatchString(f.Ref) {
			return fmt.Errorf("%w: filed needs an issue key as ref, got %q", ErrSelfReviewFindingInvalid, f.Ref)
		}
	case "suggested":
		if f.Severity == "minor" {
			return fmt.Errorf("%w: a minor finding is never suggested, only critical or important ones", ErrSelfReviewFindingInvalid)
		}
		if f.Ref != "" || strings.TrimSpace(f.Detail) == "" {
			return fmt.Errorf("%w: suggested carries no ref and needs a detail, got ref %q", ErrSelfReviewFindingInvalid, f.Ref)
		}
	case "declined", "noted":
		if f.Ref != "" {
			return fmt.Errorf("%w: %s carries no ref, got %q", ErrSelfReviewFindingInvalid, f.Disposition, f.Ref)
		}
	default:
		return fmt.Errorf("%w: disposition must be fixed, filed, declined, noted or suggested, got %q", ErrSelfReviewFindingInvalid, f.Disposition)
	}
	return nil
}

// RecordSelfReviewFinding validates and records one self-review finding,
// once: a repeat of a recorded finding -- same project, change, angle and
// disposition, and the same ref and the same note --
// writes no second row and returns the recorded one, its missing blast
// radius filled from the repeat.
func (s *Store) RecordSelfReviewFinding(ctx context.Context, projectKey string, in records.SelfReviewFinding) (records.SelfReviewFinding, error) {
	in.Severity = strings.ToLower(in.Severity)
	if err := validateSelfReviewFinding(in); err != nil {
		return records.SelfReviewFinding{}, err
	}
	out := in
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO self_review_findings (project_key, change, angle, note, disposition, ref, blast_radius, severity, detail, review_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, CASE WHEN $5 = 'suggested' THEN 'to-review' END)
		ON CONFLICT (project_key, change, angle, disposition, (COALESCE(ref, '')), (md5(note)))
		DO UPDATE SET blast_radius = COALESCE(self_review_findings.blast_radius, EXCLUDED.blast_radius),
		              severity = COALESCE(self_review_findings.severity, EXCLUDED.severity),
		              detail = COALESCE(self_review_findings.detail, EXCLUDED.detail)
		RETURNING id, note, blast_radius, COALESCE(severity, ''), COALESCE(detail, ''), COALESCE(review_status, ''), recorded_at
	`, projectKey, in.Change, in.Angle, in.Note, in.Disposition, nullIfEmpty(in.Ref), in.BlastRadius, nullIfEmpty(in.Severity), nullIfEmpty(in.Detail),
	).Scan(&out.ID, &out.Note, &out.BlastRadius, &out.Severity, &out.Detail, &out.ReviewStatus, &out.RecordedAt); err != nil {
		return records.SelfReviewFinding{}, fmt.Errorf("store: record self-review finding for %s/%s: %w", projectKey, in.Change, err)
	}
	return out, nil
}

// ListSelfReviewFindings reads one change's self-review findings in the
// order they were recorded.
func (s *Store) ListSelfReviewFindings(ctx context.Context, projectKey, change string) ([]records.SelfReviewFinding, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, change, angle, note, disposition, COALESCE(ref, ''), blast_radius, COALESCE(severity, ''), COALESCE(detail, ''), COALESCE(review_status, ''), recorded_at
		FROM self_review_findings
		WHERE project_key = $1 AND change = $2
		ORDER BY id
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: list self-review findings for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()

	out := []records.SelfReviewFinding{}
	for rows.Next() {
		var f records.SelfReviewFinding
		if err := rows.Scan(&f.ID, &f.Change, &f.Angle, &f.Note, &f.Disposition, &f.Ref, &f.BlastRadius, &f.Severity, &f.Detail, &f.ReviewStatus, &f.RecordedAt); err != nil {
			return nil, fmt.Errorf("store: list self-review findings for %s/%s: scan: %w", projectKey, change, err)
		}
		out = append(out, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list self-review findings for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}

// SelfReviewFindingRow is one row of the self-review stats view: a finding
// with the project it was recorded under.
type SelfReviewFindingRow struct {
	ProjectKey string
	records.SelfReviewFinding
}

// SelfReviewFindingsInPeriod lists every self-review finding recorded within
// period, optionally restricted to one project, newest first.
func (s *Store) SelfReviewFindingsInPeriod(ctx context.Context, period Period, project *string) ([]SelfReviewFindingRow, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT project_key, id, change, angle, note, disposition, COALESCE(ref, ''), blast_radius, COALESCE(severity, ''), COALESCE(detail, ''), COALESCE(review_status, ''), recorded_at
		FROM self_review_findings
		WHERE recorded_at >= $1 AND recorded_at < $2
		  AND ($3::text IS NULL OR project_key = $3)
		ORDER BY recorded_at DESC, id DESC
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: self-review findings in period: %w", err)
	}
	defer rows.Close()

	out := []SelfReviewFindingRow{}
	for rows.Next() {
		var r SelfReviewFindingRow
		if err := rows.Scan(&r.ProjectKey, &r.ID, &r.Change, &r.Angle, &r.Note, &r.Disposition, &r.Ref, &r.BlastRadius, &r.Severity, &r.Detail, &r.ReviewStatus, &r.RecordedAt); err != nil {
			return nil, fmt.Errorf("store: self-review findings in period: scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: self-review findings in period: %w", err)
	}
	return out, nil
}

// ErrSelfReviewFindingNotSuggested reports a review of a row that does not exist or is not a
// suggestion; the handler maps it to 404.
var ErrSelfReviewFindingNotSuggested = errors.New("store: no such suggested self-review finding")

// ReviewSelfReviewFinding sets a suggested finding's review status: to-review, approved or
// declined. Nothing else about the row changes.
func (s *Store) ReviewSelfReviewFinding(ctx context.Context, projectKey string, id int64, status string) error {
	switch status {
	case "to-review", "approved", "declined":
	default:
		return fmt.Errorf("%w: review status must be to-review, approved or declined, got %q", ErrSelfReviewFindingInvalid, status)
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE self_review_findings SET review_status = $3
		WHERE project_key = $1 AND id = $2 AND disposition = 'suggested'
	`, projectKey, id, status)
	if err != nil {
		return fmt.Errorf("store: review self-review finding %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: %d", ErrSelfReviewFindingNotSuggested, id)
	}
	return nil
}

// SelfReviewPage is one page of the self-review dashboard: the rows of the chosen list, how many
// rows that list holds in all, and the suggestion counts by review status (for the status chips).
type SelfReviewPage struct {
	Rows                         []SelfReviewFindingRow
	Total                        int
	ToReview, Approved, Declined int
}

// SelfReviewFindingsPage reads one page of the findings recorded within period, newest first,
// optionally restricted to one project. show is "fixed" (the fixed findings) or "to-review" (the
// suggestions with the given review status). Only limit rows are read.
func (s *Store) SelfReviewFindingsPage(ctx context.Context, period Period, project *string, show, status string, limit, offset int) (SelfReviewPage, error) {
	var out SelfReviewPage
	disposition, reviewStatus := "fixed", ""
	switch show {
	case "fixed":
	case "to-review":
		disposition, reviewStatus = "suggested", status
		if status != "to-review" && status != "approved" && status != "declined" {
			return out, fmt.Errorf("%w: status must be to-review, approved or declined, got %q", ErrSelfReviewFindingInvalid, status)
		}
	default:
		return out, fmt.Errorf("%w: show must be fixed or to-review, got %q", ErrSelfReviewFindingInvalid, show)
	}
	if err := s.pool.QueryRow(ctx, `
		SELECT
		  count(*) FILTER (WHERE disposition = $4 AND ($5 = '' OR review_status = $5)),
		  count(*) FILTER (WHERE review_status = 'to-review'),
		  count(*) FILTER (WHERE review_status = 'approved'),
		  count(*) FILTER (WHERE review_status = 'declined')
		FROM self_review_findings
		WHERE recorded_at >= $1 AND recorded_at < $2 AND ($3::text IS NULL OR project_key = $3)
	`, period.From, period.To, project, disposition, reviewStatus).Scan(&out.Total, &out.ToReview, &out.Approved, &out.Declined); err != nil {
		return out, fmt.Errorf("store: count self-review findings: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT project_key, id, change, angle, note, disposition, COALESCE(ref, ''), blast_radius, COALESCE(severity, ''), COALESCE(detail, ''), COALESCE(review_status, ''), recorded_at
		FROM self_review_findings
		WHERE recorded_at >= $1 AND recorded_at < $2 AND ($3::text IS NULL OR project_key = $3)
		  AND disposition = $4 AND ($5 = '' OR review_status = $5)
		ORDER BY recorded_at DESC, id DESC
		LIMIT $6 OFFSET $7
	`, period.From, period.To, project, disposition, reviewStatus, limit, offset)
	if err != nil {
		return out, fmt.Errorf("store: self-review findings page: %w", err)
	}
	defer rows.Close()
	out.Rows = []SelfReviewFindingRow{}
	for rows.Next() {
		var r SelfReviewFindingRow
		if err := rows.Scan(&r.ProjectKey, &r.ID, &r.Change, &r.Angle, &r.Note, &r.Disposition, &r.Ref, &r.BlastRadius, &r.Severity, &r.Detail, &r.ReviewStatus, &r.RecordedAt); err != nil {
			return out, fmt.Errorf("store: self-review findings page: scan: %w", err)
		}
		out.Rows = append(out.Rows, r)
	}
	return out, rows.Err()
}
