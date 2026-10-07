package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/tweety53/agents/stats/internal/records"
)

// ErrSelfReviewFindingInvalid reports a RecordSelfReviewFinding whose
// disposition is outside fixed/filed/declined, whose ref does not fit its
// disposition, whose change, angle or note is empty, or whose blast radius
// is negative. The handler maps it
// to 400: a caller's mistake, never a store failure.
var ErrSelfReviewFindingInvalid = errors.New("store: invalid self-review finding")

// validateSelfReviewFinding is the boundary for the rules the table's CHECK
// cannot state: which ref shape each disposition carries.
func validateSelfReviewFinding(f records.SelfReviewFinding) error {
	if f.Change == "" || f.Angle == "" || f.Note == "" {
		return fmt.Errorf("%w: change, angle and note are required", ErrSelfReviewFindingInvalid)
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
	case "declined":
		if f.Ref != "" {
			return fmt.Errorf("%w: declined carries no ref, got %q", ErrSelfReviewFindingInvalid, f.Ref)
		}
	default:
		return fmt.Errorf("%w: disposition must be fixed, filed or declined, got %q", ErrSelfReviewFindingInvalid, f.Disposition)
	}
	return nil
}

// RecordSelfReviewFinding validates and inserts one self-review finding.
func (s *Store) RecordSelfReviewFinding(ctx context.Context, projectKey string, in records.SelfReviewFinding) (records.SelfReviewFinding, error) {
	if err := validateSelfReviewFinding(in); err != nil {
		return records.SelfReviewFinding{}, err
	}
	out := in
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO self_review_findings (project_key, change, angle, note, disposition, ref, blast_radius)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, recorded_at
	`, projectKey, in.Change, in.Angle, in.Note, in.Disposition, nullIfEmpty(in.Ref), in.BlastRadius,
	).Scan(&out.ID, &out.RecordedAt); err != nil {
		return records.SelfReviewFinding{}, fmt.Errorf("store: record self-review finding for %s/%s: %w", projectKey, in.Change, err)
	}
	return out, nil
}

// ListSelfReviewFindings reads one change's self-review findings in the
// order they were recorded.
func (s *Store) ListSelfReviewFindings(ctx context.Context, projectKey, change string) ([]records.SelfReviewFinding, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, change, angle, note, disposition, COALESCE(ref, ''), blast_radius, recorded_at
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
		if err := rows.Scan(&f.ID, &f.Change, &f.Angle, &f.Note, &f.Disposition, &f.Ref, &f.BlastRadius, &f.RecordedAt); err != nil {
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
		SELECT project_key, id, change, angle, note, disposition, COALESCE(ref, ''), blast_radius, recorded_at
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
		if err := rows.Scan(&r.ProjectKey, &r.ID, &r.Change, &r.Angle, &r.Note, &r.Disposition, &r.Ref, &r.BlastRadius, &r.RecordedAt); err != nil {
			return nil, fmt.Errorf("store: self-review findings in period: scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: self-review findings in period: %w", err)
	}
	return out, nil
}
