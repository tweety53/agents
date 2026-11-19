package store

import (
	"context"
	"fmt"
	"time"
)

// RoleDispatch is one dispatch as the role-costs view reads it: its stored
// role and key (the view derives the allowed-pairs role from them), its
// pair, its span and its cost. CostUSD and EndedAt are nil when unrecorded,
// never 0; Outcome is empty while the dispatch runs.
type RoleDispatch struct {
	Role      string
	Key       string
	Model     string
	Effort    string
	StartedAt time.Time
	EndedAt   *time.Time
	CostUSD   *float64
	Outcome   string
}

// RoleDispatches returns every dispatch started in period, of project when
// one is given, oldest first.
func (s *Store) RoleDispatches(ctx context.Context, period Period, project *string) ([]RoleDispatch, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT d.role, COALESCE(d.dispatch_key, ''), d.model, d.effort,
			d.started_at, d.ended_at, (d.metrics->>'cost_usd')::float8, COALESCE(d.outcome, '')
		FROM dispatches d
		JOIN changes c ON c.id = d.change_id
		WHERE d.started_at >= $1 AND d.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		ORDER BY d.started_at, d.id
	`, period.From, period.To, project)
	if err != nil {
		return nil, fmt.Errorf("store: role dispatches: %w", err)
	}
	defer rows.Close()
	var out []RoleDispatch
	for rows.Next() {
		var r RoleDispatch
		if err := rows.Scan(&r.Role, &r.Key, &r.Model, &r.Effort, &r.StartedAt, &r.EndedAt, &r.CostUSD, &r.Outcome); err != nil {
			return nil, fmt.Errorf("store: role dispatches: scan: %w", err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: role dispatches: %w", err)
	}
	return out, nil
}
