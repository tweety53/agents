package store

import (
	"context"
	"fmt"
	"time"
)

// PromptChange is one change to a tunable prompt: a tuning
// /flow-prompt-tuning applied (Kind "tuning") or a commit that edited the
// prompt file (Kind "edit"). SHA is "<commit>:<target>", unique per row;
// Area is review, impl, vv or fix and names the charts that draw it as a
// marker (prompt-changes-table, markers-by-area).
type PromptChange struct {
	SHA     string
	At      time.Time
	Kind    string
	Area    string
	Slot    *string
	Target  string
	Summary string
	CaseID  *string
}

// UpsertPromptChange records in, keyed by its SHA. A tuning row is the more
// specific fact, so it replaces an edit row of the same SHA and an edit
// never demotes a tuning back (prompt-changes-table). An area or kind
// outside the table's CHECK is refused by the database.
func (s *Store) UpsertPromptChange(ctx context.Context, in PromptChange) error {
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO prompt_changes (sha, at, kind, area, slot, target, summary, case_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (sha) DO UPDATE SET
			kind = CASE WHEN prompt_changes.kind = 'tuning' THEN 'tuning' ELSE EXCLUDED.kind END,
			case_id = COALESCE(EXCLUDED.case_id, prompt_changes.case_id),
			at = CASE WHEN EXCLUDED.kind = 'tuning' THEN EXCLUDED.at ELSE prompt_changes.at END,
			area = CASE WHEN EXCLUDED.kind = 'tuning' THEN EXCLUDED.area ELSE prompt_changes.area END,
			slot = CASE WHEN EXCLUDED.kind = 'tuning' THEN EXCLUDED.slot ELSE prompt_changes.slot END,
			target = CASE WHEN EXCLUDED.kind = 'tuning' THEN EXCLUDED.target ELSE prompt_changes.target END,
			summary = CASE WHEN EXCLUDED.kind = 'tuning' THEN EXCLUDED.summary ELSE prompt_changes.summary END
	`, in.SHA, in.At, in.Kind, in.Area, in.Slot, in.Target, in.Summary, in.CaseID); err != nil {
		return fmt.Errorf("store: upsert prompt change %q: %w", in.SHA, err)
	}
	return nil
}

// ListPromptChanges returns the prompt changes with at in [from, to),
// ordered by at.
func (s *Store) ListPromptChanges(ctx context.Context, from, to time.Time) ([]PromptChange, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT sha, at, kind, area, slot, target, summary, case_id
		FROM prompt_changes
		WHERE at >= $1 AND at < $2
		ORDER BY at, sha
	`, from, to)
	if err != nil {
		return nil, fmt.Errorf("store: list prompt changes: %w", err)
	}
	defer rows.Close()

	out := []PromptChange{}
	for rows.Next() {
		var c PromptChange
		if err := rows.Scan(&c.SHA, &c.At, &c.Kind, &c.Area, &c.Slot, &c.Target, &c.Summary, &c.CaseID); err != nil {
			return nil, fmt.Errorf("store: list prompt changes: scan: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: list prompt changes: %w", err)
	}
	return out, nil
}

// LastPromptChangeAt returns the newest prompt change's at, whatever the
// period, or nil when none is recorded -- what the view's "Since last
// prompt change" preset starts from (view-owned-range).
func (s *Store) LastPromptChangeAt(ctx context.Context) (*time.Time, error) {
	var at *time.Time
	if err := s.pool.QueryRow(ctx, `SELECT max(at) FROM prompt_changes`).Scan(&at); err != nil {
		return nil, fmt.Errorf("store: last prompt change: %w", err)
	}
	return at, nil
}
