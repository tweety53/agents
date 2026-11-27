package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TraceSpan is one timed row of a change as `flow stats trace` draws it:
// Kind is request, stage, dispatch, guard or verify; Lane the track it
// belongs on; End nil while the row is still open; Split a stage run's or
// dispatch's stored time split (0115_time_splits.sql), nil where none is.
type TraceSpan struct {
	Kind    string
	Lane    string
	Name    string
	AgentID string
	Outcome string
	Start   time.Time
	End     *time.Time
	Split   json.RawMessage
}

// TraceSpans returns every timed row of one change, ordered by kind, then
// start, then name: its stage runs, requests and dispatches (one lane per
// role and dispatch key, the seq standing in for a missing key), and its
// guard and verify runs, whose start is their recorded end less their
// duration. A guard run that waited (wait_ms) adds a span on the
// "guard wait" lane, drawn from its start for the length of its wait --
// the wait's total, not where inside the run it fell -- and one whose
// bootstrap was stamped (bootstrap_ms) a span on the "guard bootstrap" lane
// ending at its start. A guard run belongs to the change when its worktree names the
// change's worktree, as Bottlenecks claims it. ErrChangeNotFound when the
// store has no such change.
func (s *Store) TraceSpans(ctx context.Context, projectKey, change string) ([]TraceSpan, error) {
	id, err := s.changeID(ctx, projectKey, change)
	if err != nil {
		return nil, err
	}
	q, err := s.pool.Query(ctx, `
		SELECT 'stage', 'stage', sr.stage || ' #' || sr.attempt, '', COALESCE(sr.outcome, ''), sr.started_at, sr.ended_at, sr.time_split
		FROM stage_runs sr WHERE sr.change_id = $1
		UNION ALL
		SELECT 'request', 'request', 'request ' || r.n || CASE WHEN r.title <> '' THEN ': ' || r.title ELSE '' END,
			'', COALESCE(r.outcome, ''), r.received_at, r.ended_at, NULL::jsonb
		FROM requests r WHERE r.change_id = $1
		UNION ALL
		SELECT 'dispatch', l.lane, l.lane, COALESCE(d.agent_id, ''), COALESCE(d.outcome, ''), d.started_at, d.ended_at, d.time_split
		FROM dispatches d CROSS JOIN LATERAL (SELECT d.role || ' ' || COALESCE(d.dispatch_key, '#' || d.seq) AS lane) l
		WHERE d.change_id = $1
		UNION ALL
		SELECT 'guard', 'guard', g.guard, '', g.outcome, g.recorded_at - g.duration_ms * interval '1 millisecond', g.recorded_at, NULL
		FROM guard_runs g WHERE g.project_key = $2
		  AND (strpos(g.worktree, '-worktrees/' || $3 || '/') > 0
		       OR right(g.worktree, length('-worktrees/' || $3)) = '-worktrees/' || $3)
		UNION ALL
		SELECT 'guard', 'guard wait', g.guard || ' wait', '', '', g.recorded_at - g.duration_ms * interval '1 millisecond',
			g.recorded_at - (g.duration_ms - g.wait_ms) * interval '1 millisecond', NULL
		FROM guard_runs g WHERE g.project_key = $2 AND g.wait_ms > 0
		  AND (strpos(g.worktree, '-worktrees/' || $3 || '/') > 0
		       OR right(g.worktree, length('-worktrees/' || $3)) = '-worktrees/' || $3)
		UNION ALL
		SELECT 'guard', 'guard bootstrap', g.guard || ' bootstrap', '', '',
			g.recorded_at - (g.duration_ms + g.bootstrap_ms) * interval '1 millisecond',
			g.recorded_at - g.duration_ms * interval '1 millisecond', NULL
		FROM guard_runs g WHERE g.project_key = $2 AND g.bootstrap_ms IS NOT NULL
		  AND (strpos(g.worktree, '-worktrees/' || $3 || '/') > 0
		       OR right(g.worktree, length('-worktrees/' || $3)) = '-worktrees/' || $3)
		UNION ALL
		SELECT 'verify', 'verify', v.label, '', 'exit ' || v.exit_code, v.recorded_at - v.seconds * interval '1 second', v.recorded_at, NULL
		FROM verify_runs v WHERE v.change_id = $1
		ORDER BY 1, 6, 3
	`, id, projectKey, change)
	if err != nil {
		return nil, fmt.Errorf("store: trace spans: %w", err)
	}
	defer q.Close()
	var out []TraceSpan
	for q.Next() {
		var sp TraceSpan
		var split []byte
		if err := q.Scan(&sp.Kind, &sp.Lane, &sp.Name, &sp.AgentID, &sp.Outcome, &sp.Start, &sp.End, &split); err != nil {
			return nil, fmt.Errorf("store: trace spans: scan: %w", err)
		}
		sp.Split = storedSplit(split)
		out = append(out, sp)
	}
	if err := q.Err(); err != nil {
		return nil, fmt.Errorf("store: trace spans: %w", err)
	}
	return out, nil
}
