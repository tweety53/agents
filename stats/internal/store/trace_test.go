package store_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestTraceSpansOfOneChange seeds one change's stage runs, request,
// dispatches, guard runs and verify runs next to another change's, and pins
// the spans `flow stats trace` draws: only the asked change's rows (guard
// runs claimed by worktree), a guard or verify run's start derived from its
// recorded end minus its duration, an open run's end left nil, ordered by
// kind, then start.
func TestTraceSpansOfOneChange(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	pk := fmt.Sprintf("proj-trace-%d", time.Now().UnixNano())
	seedChange(t, st, pk, "other")
	seedChange(t, st, pk, "mine")
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
	mine, other := id("mine"), id("other")
	exec(`INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, outcome) VALUES
		($1,'claude-code','/flow','flow.implement',1,$3,$4,'completed'),
		($1,'claude-code','/flow','flow.verify',1,$4,NULL,NULL),
		($2,'claude-code','/flow','flow.implement',1,$3,$4,'completed')`, mine, other, at(0), at(100))
	exec(`INSERT INTO requests (change_id, n, title, received_at, ended_at, outcome) VALUES
		($1,1,'add trace',$2,$3,'landed'), ($1,2,'',$3,NULL,NULL)`, mine, at(0), at(150))
	exec(`INSERT INTO dispatches (change_id, seq, role, model, started_at, ended_at, dispatch_key, agent_id, outcome) VALUES
		($1,1,'operator-asked','opus',$2,$3,'operator-1','a1','done'),
		($1,2,'panel-reviewer','opus',$3,NULL,NULL,NULL,NULL)`, mine, at(10), at(90))
	exec(`INSERT INTO guard_runs (project_key, guard, worktree, exit_code, outcome, duration_ms, recorded_at) VALUES
		($1,'check-a','/x/p-worktrees/mine',0,'clear',1500,$2),
		($1,'check-a','/x/p-worktrees/other',1,'fired',500,$2)`, pk, at(20))
	exec(`UPDATE guard_runs SET wait_ms = 1000 WHERE project_key = $1 AND duration_ms = 1500`, pk)
	exec(`INSERT INTO verify_runs (change_id, kind, label, seconds, tests, failed, exit_code, recorded_at) VALUES
		($1,'run-verify-list','lint',30,1,0,0,$3), ($2,'run-verify-list','lint',9,1,0,0,$3)`, mine, other, at(130))

	spans, err := st.TraceSpans(ctx, pk, "mine")
	if err != nil {
		t.Fatal(err)
	}
	ptr := func(s int) *time.Time { v := at(s); return &v }
	want := []store.TraceSpan{
		{Kind: "dispatch", Lane: "operator-asked operator-1", Name: "operator-asked operator-1", AgentID: "a1", Outcome: "done", Start: at(10), End: ptr(90)},
		{Kind: "dispatch", Lane: "panel-reviewer #2", Name: "panel-reviewer #2", Start: at(90)},
		{Kind: "guard", Lane: "guard", Name: "check-a", Outcome: "clear", Start: at(20).Add(-1500 * time.Millisecond), End: ptr(20)},
		{Kind: "guard", Lane: "guard wait", Name: "check-a wait", Start: at(20).Add(-1500 * time.Millisecond), End: func() *time.Time { v := at(20).Add(-500 * time.Millisecond); return &v }()},
		{Kind: "request", Lane: "request", Name: "request 1: add trace", Outcome: "landed", Start: at(0), End: ptr(150)},
		{Kind: "request", Lane: "request", Name: "request 2", Start: at(150)},
		{Kind: "stage", Lane: "stage", Name: "flow.implement #1", Outcome: "completed", Start: at(0), End: ptr(100)},
		{Kind: "stage", Lane: "stage", Name: "flow.verify #1", Start: at(100)},
		{Kind: "verify", Lane: "verify", Name: "lint", Outcome: "exit 0", Start: at(100), End: ptr(130)},
	}
	for i := range spans {
		spans[i].Start = spans[i].Start.UTC()
		if spans[i].End != nil {
			e := spans[i].End.UTC()
			spans[i].End = &e
		}
	}
	if !reflect.DeepEqual(spans, want) {
		t.Errorf("spans:\n got %+v\nwant %+v", spans, want)
	}
	if _, err := st.TraceSpans(ctx, pk, "absent"); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("absent change err = %v, want ErrChangeNotFound", err)
	}
}
