package store

import (
	"context"
	"fmt"
	"time"
)

// HeadlessRow is one headless model call (pair-judge, plan-triage, the
// plan-check decider) as a change's run page lists it. CostUSD is nil when
// the call's cost went unrecorded, never 0.
type HeadlessRow struct {
	Role       string    `json:"role"`
	Model      string    `json:"model"`
	Effort     string    `json:"effort"`
	CostUSD    *float64  `json:"costUSD"`
	DurationMS int       `json:"durationMS"`
	RecordedAt time.Time `json:"recordedAt"`
}

// ChangeHeadless returns the headless calls that belong to a change, oldest
// first. A guard run carries a worktree path, not a change, so a call is the
// change's when it ran in the change's worktree, or in the project's main
// checkout while one of the change's stage runs was open (from the first
// run's start to the last one's end, or now while one is still running).
func (s *Store) ChangeHeadless(ctx context.Context, projectKey, change string) ([]HeadlessRow, error) {
	rows, err := s.pool.Query(ctx, `
		WITH c AS (SELECT id, project_key, name FROM changes WHERE project_key = $1 AND name = $2),
		span AS (
			SELECT min(s.started_at) AS f,
			       CASE WHEN bool_or(s.ended_at IS NULL) THEN now() ELSE max(s.ended_at) END AS t
			FROM stage_runs s JOIN c ON c.id = s.change_id
		)
		SELECT g.headless_role, g.headless_model, COALESCE(g.headless_effort, 'default'),
		       g.headless_cost_usd, COALESCE(g.headless_ms, 0), g.recorded_at
		FROM guard_runs g, c, span
		WHERE g.headless_role IS NOT NULL AND g.project_key = c.project_key
		  AND (strpos(g.worktree, '-worktrees/' || c.name) > 0
		       OR (strpos(g.worktree, '-worktrees/') = 0 AND g.recorded_at BETWEEN span.f AND span.t + interval '1 minute'))
		ORDER BY g.recorded_at, g.id
	`, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: change headless for %s/%s: %w", projectKey, change, err)
	}
	defer rows.Close()
	var out []HeadlessRow
	for rows.Next() {
		var r HeadlessRow
		if err := rows.Scan(&r.Role, &r.Model, &r.Effort, &r.CostUSD, &r.DurationMS, &r.RecordedAt); err != nil {
			return nil, fmt.Errorf("store: change headless for %s/%s: scan: %w", projectKey, change, err)
		}
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: change headless for %s/%s: %w", projectKey, change, err)
	}
	return out, nil
}
