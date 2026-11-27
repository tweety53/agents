package store

import (
	"context"
	"fmt"
)

// BottleneckRow is one name's time within one dimension of the bottleneck
// ranking: how many runs it had, their summed and P50/P95 duration in
// milliseconds. Dimension is stage, role, role-class, guard, guard-wait,
// guard-bootstrap, hook, verify, llm, tool or human.
type BottleneckRow struct {
	Dimension string
	Name      string
	Runs      int
	TotalMs   int64
	P50Ms     int64
	P95Ms     int64
}

// Bottlenecks ranks where the last `last` changes (by their latest stage
// run, optionally one project's) spent their time, from rows flowd already
// holds: ended stage runs by stage key, ended dispatches by role and by
// dispatch_class(role), guard runs by guard (a run belongs to a change when
// its worktree names the change's worktree, as ChangeGuardRuns claims it) --
// its work, duration_ms less wait_ms, under guard and the wait, where there
// was one, under guard-wait, and the bootstrap before it, where stamped,
// under guard-bootstrap --, hook runs by hook (hook_runs carry no project
// or worktree: a run counts once when it ended inside any kept change's
// first-start-to-last-end window),
// verify runs by label, and the stored transcript time splits
// (0115_time_splits.sql) -- every dispatch's, named agent, and every stage
// run's parent split (0117_parent_splits.sql: its session's own transcript,
// whichever agent marked it; a stage run's own time_split is not read, a
// subagent's re-counting its dispatch), named parent -- as llm and human by those names and tool by
// tool, one run per split. WallMs is the summed wall clock of those
// changes -- each change's first stage-run start to its last stage-run end
// -- the denominator a row's share is read against. Rows come ordered by
// dimension, then total descending.
func (s *Store) Bottlenecks(ctx context.Context, project *string, last int) (rows []BottleneckRow, wallMs int64, err error) {
	q, err := s.pool.Query(ctx, `
		WITH picked AS (
			SELECT c.id, c.project_key, c.name,
				MIN(sr.started_at) AS first, MAX(COALESCE(sr.ended_at, sr.started_at)) AS last,
				EXTRACT(EPOCH FROM MAX(COALESCE(sr.ended_at, sr.started_at)) - MIN(sr.started_at)) * 1000 AS wall
			FROM changes c JOIN stage_runs sr ON sr.change_id = c.id
			WHERE $1::text IS NULL OR c.project_key = $1
			GROUP BY c.id
			ORDER BY MAX(sr.started_at) DESC
			LIMIT $2
		), spans AS (
			SELECT 'stage' AS dim, sr.stage AS name,
				EXTRACT(EPOCH FROM sr.ended_at - sr.started_at) * 1000 AS ms
			FROM stage_runs sr JOIN picked p ON p.id = sr.change_id
			WHERE sr.ended_at IS NOT NULL
			UNION ALL
			SELECT d.dim, d.name, EXTRACT(EPOCH FROM x.ended_at - x.started_at) * 1000
			FROM dispatches x JOIN picked p ON p.id = x.change_id
			CROSS JOIN LATERAL (VALUES ('role', x.role), ('role-class', dispatch_class(x.role))) d(dim, name)
			WHERE x.ended_at IS NOT NULL
			UNION ALL
			SELECT w.dim, g.guard, w.ms
			FROM guard_runs g JOIN picked p ON g.project_key = p.project_key
			  AND (strpos(g.worktree, '-worktrees/' || p.name || '/') > 0
			       OR right(g.worktree, length('-worktrees/' || p.name)) = '-worktrees/' || p.name)
			CROSS JOIN LATERAL (VALUES ('guard', g.duration_ms - g.wait_ms), ('guard-wait', NULLIF(g.wait_ms, 0)),
				('guard-bootstrap', g.bootstrap_ms)) w(dim, ms)
			WHERE w.ms IS NOT NULL
			UNION ALL
			SELECT 'hook', h.hook, h.duration_us / 1000.0
			FROM hook_runs h
			WHERE EXISTS (SELECT 1 FROM picked p WHERE h.ran_at BETWEEN p.first AND p.last)
			UNION ALL
			SELECT 'verify', v.label, v.seconds * 1000
			FROM verify_runs v JOIN picked p ON p.id = v.change_id
			UNION ALL
			SELECT t.dim, t.name, t.ms
			FROM (
				SELECT x.time_split AS ts, 'agent' AS who
				FROM dispatches x JOIN picked p ON p.id = x.change_id
				WHERE jsonb_typeof(x.time_split) = 'object'
				UNION ALL
				SELECT sr.parent_split, 'parent'
				FROM stage_runs sr JOIN picked p ON p.id = sr.change_id
				WHERE jsonb_typeof(sr.parent_split) = 'object'
			) s
			CROSS JOIN LATERAL (
				VALUES ('llm', s.who, (s.ts->>'llmMs')::numeric), ('human', s.who, (s.ts->>'humanMs')::numeric)
				UNION ALL
				SELECT 'tool', e.key, e.value::numeric FROM jsonb_each_text(COALESCE(s.ts->'byTool', '{}')) e
			) t(dim, name, ms)
		)
		SELECT dim, name, COUNT(*)::int, ROUND(SUM(ms))::bigint,
			ROUND(percentile_cont(0.5) WITHIN GROUP (ORDER BY ms))::bigint,
			ROUND(percentile_cont(0.95) WITHIN GROUP (ORDER BY ms))::bigint,
			(SELECT COALESCE(ROUND(SUM(wall)), 0)::bigint FROM picked)
		FROM spans
		GROUP BY dim, name
		ORDER BY dim, SUM(ms) DESC, name
	`, project, last)
	if err != nil {
		return nil, 0, fmt.Errorf("store: bottlenecks: %w", err)
	}
	defer q.Close()
	for q.Next() {
		var r BottleneckRow
		if err := q.Scan(&r.Dimension, &r.Name, &r.Runs, &r.TotalMs, &r.P50Ms, &r.P95Ms, &wallMs); err != nil {
			return nil, 0, fmt.Errorf("store: bottlenecks: scan: %w", err)
		}
		rows = append(rows, r)
	}
	if err := q.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: bottlenecks: %w", err)
	}
	return rows, wallMs, nil
}
