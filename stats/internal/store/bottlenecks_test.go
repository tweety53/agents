package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestBottlenecksRanksEachDimension seeds two changes' stage runs,
// dispatches, guard runs and verify runs, and pins the ranking: -last keeps
// only the most recent change, each dimension is ordered by total, P50/P95
// come from the run durations -- a guard's its work, duration less wait_ms,
// the wait ranked apart as guard-wait and a stamped
// bootstrap_ms as guard-bootstrap --, hook runs that fell inside a kept
// change's wall clock by hook, the stored time splits as llm, tool (per
// tool) and human -- named agent for a dispatch's, parent for a stage run's
// linked to a parent transcript; a subagent-linked stage run's and a JSON
// null split add nothing -- and the wall clock is the kept change's first
// start to last end.
func TestBottlenecksRanksEachDimension(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	pk := fmt.Sprintf("proj-bneck-%d", time.Now().UnixNano())
	seedChange(t, st, pk, "old")
	seedChange(t, st, pk, "new")
	pool := store.PoolOf(st)
	id := func(name string) int64 {
		var v int64
		if err := pool.QueryRow(ctx, `SELECT id FROM changes WHERE project_key=$1 AND name=$2`, pk, name).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	exec := func(sql string, args ...any) {
		t.Helper()
		if _, err := pool.Exec(ctx, sql, args...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	at := func(s int) time.Time { return t0.Add(time.Duration(s) * time.Second) }
	stage := func(c int64, key string, n, from, to int) {
		exec(`INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at)
			VALUES ($1,'claude-code','/flow',$2,$3,$4,$5)`, c, key, n, at(from), at(to))
	}
	oldID, newID := id("old"), id("new")
	stage(oldID, "flow.implement", 1, 0, 1000) // older change: excluded by last=1
	stage(newID, "flow.implement", 1, 2000, 2100)
	stage(newID, "flow.implement", 2, 2100, 2400)
	stage(newID, "flow.verify", 1, 2400, 2500)
	exec(`INSERT INTO dispatches (change_id, seq, role, model, started_at, ended_at) VALUES
		($1,1,'panel-reviewer','opus',$2,$3), ($1,2,'operator-asked','opus',$3,$4)`, newID, at(2000), at(2060), at(2100))
	exec(`UPDATE dispatches SET time_split = $2 WHERE change_id = $1 AND seq = 1`, newID,
		`{"llmMs":40000,"toolMs":20000,"humanMs":0,"byTool":{"Bash go":15000,"Read":5000}}`)
	exec(`UPDATE dispatches SET time_split = 'null' WHERE change_id = $1 AND seq = 2`, newID)
	exec(`UPDATE stage_runs SET transcript_path = '/t/s.jsonl', time_split = $2 WHERE change_id = $1 AND stage = 'flow.verify'`, newID,
		`{"llmMs":60000,"toolMs":30000,"humanMs":10000,"byTool":{"Bash go":30000}}`)
	exec(`UPDATE stage_runs SET transcript_path = '/t/s/subagents/agent-a.jsonl', time_split = '{"llmMs":1}'
		WHERE change_id = $1 AND stage = 'flow.implement' AND attempt = 1`, newID)
	exec(`INSERT INTO guard_runs (project_key, guard, worktree, exit_code, outcome, duration_ms, recorded_at) VALUES
		($1,'check-a','/x/p-worktrees/new',0,'clear',500,$2),
		($1,'check-a','/x/p-worktrees/new/sub',0,'clear',1500,$2),
		($1,'check-a','/x/p-worktrees/old',0,'clear',99999,$2)`, pk, at(2050))
	exec(`UPDATE guard_runs SET wait_ms = 400, bootstrap_ms = 30 WHERE project_key = $1 AND duration_ms = 500`, pk)
	exec(`INSERT INTO verify_runs (change_id, kind, label, seconds, tests, failed, exit_code, recorded_at)
		VALUES ($1,'run-verify-list','lint',30,1,0,0,$2)`, newID, at(2450))
	hook := fmt.Sprintf("hook-%d", time.Now().UnixNano()) // hook_runs carry no project: a name of this test's own
	if err := st.RecordHookRuns(ctx, []records.HookRun{
		{Hook: hook, DurationUS: 4000, At: at(2010)},
		{Hook: hook, DurationUS: 8000, At: at(2490)},
		{Hook: hook, DurationUS: 99_000_000, At: at(500)}, // inside the old change only
	}); err != nil {
		t.Fatal(err)
	}

	rows, wall, err := st.Bottlenecks(ctx, &pk, 1)
	if err != nil {
		t.Fatal(err)
	}
	if wall != 500_000 {
		t.Errorf("wall = %d, want 500000", wall)
	}
	got := map[string]store.BottleneckRow{}
	for _, r := range rows {
		if r.Dimension == "hook" && r.Name != hook {
			continue // another test's, or the live machine's, hook runs
		}
		got[r.Dimension+"/"+r.Name] = r
	}
	want := map[string]store.BottleneckRow{
		"stage/flow.implement":    {Runs: 2, TotalMs: 400_000, P50Ms: 200_000, P95Ms: 290_000},
		"stage/flow.verify":       {Runs: 1, TotalMs: 100_000, P50Ms: 100_000, P95Ms: 100_000},
		"role/panel-reviewer":     {Runs: 1, TotalMs: 60_000, P50Ms: 60_000, P95Ms: 60_000},
		"role/operator-asked":     {Runs: 1, TotalMs: 40_000, P50Ms: 40_000, P95Ms: 40_000},
		"role-class/reviewer":     {Runs: 1, TotalMs: 60_000, P50Ms: 60_000, P95Ms: 60_000},
		"role-class/implementer":  {Runs: 1, TotalMs: 40_000, P50Ms: 40_000, P95Ms: 40_000},
		"guard/check-a":           {Runs: 2, TotalMs: 1600, P50Ms: 800, P95Ms: 1430},
		"guard-wait/check-a":      {Runs: 1, TotalMs: 400, P50Ms: 400, P95Ms: 400},
		"guard-bootstrap/check-a": {Runs: 1, TotalMs: 30, P50Ms: 30, P95Ms: 30},
		"verify/lint":             {Runs: 1, TotalMs: 30_000, P50Ms: 30_000, P95Ms: 30_000},
		"hook/" + hook:            {Runs: 2, TotalMs: 12, P50Ms: 6, P95Ms: 8},
		"llm/agent":               {Runs: 1, TotalMs: 40_000, P50Ms: 40_000, P95Ms: 40_000},
		"llm/parent":              {Runs: 1, TotalMs: 60_000, P50Ms: 60_000, P95Ms: 60_000},
		"human/agent":             {Runs: 1, TotalMs: 0, P50Ms: 0, P95Ms: 0},
		"human/parent":            {Runs: 1, TotalMs: 10_000, P50Ms: 10_000, P95Ms: 10_000},
		"tool/Bash go":            {Runs: 2, TotalMs: 45_000, P50Ms: 22_500, P95Ms: 29_250},
		"tool/Read":               {Runs: 1, TotalMs: 5000, P50Ms: 5000, P95Ms: 5000},
	}
	if len(got) != len(want) {
		t.Errorf("rows = %+v, want %d rows", rows, len(want))
	}
	for k, w := range want {
		g := got[k]
		if g.Runs != w.Runs || g.TotalMs != w.TotalMs || g.P50Ms != w.P50Ms || g.P95Ms != w.P95Ms {
			t.Errorf("%s = %+v, want %+v", k, g, w)
		}
	}
	if rows[0].Dimension != "guard" {
		t.Errorf("first row %+v, want dimension-ordered", rows[0])
	}
}
