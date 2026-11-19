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
// one is given, oldest first -- and every guard run's headless model call
// (pair-judge, plan-check's decider) as one more row, its Role "headless"
// and its Key the role the call ran as.
func (s *Store) RoleDispatches(ctx context.Context, period Period, project *string) ([]RoleDispatch, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT * FROM (
		SELECT d.role, COALESCE(d.dispatch_key, '') AS key, d.model, d.effort,
			d.started_at, d.ended_at, (d.metrics->>'cost_usd')::float8, COALESCE(d.outcome, '')
		FROM dispatches d
		JOIN changes c ON c.id = d.change_id
		WHERE d.started_at >= $1 AND d.started_at < $2
		  AND ($3::text IS NULL OR c.project_key = $3)
		UNION ALL
		SELECT 'headless', g.headless_role, g.headless_model, COALESCE(g.headless_effort, 'default'),
			g.recorded_at - make_interval(secs => COALESCE(g.headless_ms, 0) / 1000.0), g.recorded_at,
			g.headless_cost_usd, 'completed'
		FROM guard_runs g
		WHERE g.headless_role IS NOT NULL
		  AND g.recorded_at >= $1 AND g.recorded_at < $2
		  AND ($3::text IS NULL OR g.project_key = $3)
		) r
		ORDER BY 5, 2
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
