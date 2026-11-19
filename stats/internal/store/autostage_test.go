package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// autoRuns returns the change's auto-opened stage runs.
func autoRuns(t *testing.T, st *store.Store, projectKey string) []store.StageRun {
	t.Helper()
	runs, _, err := st.QueryStageRuns(context.Background(), store.Query{
		Filters: []store.Filter{{Field: "project_key", Op: store.OpEq, Value: projectKey}},
	})
	if err != nil {
		t.Fatalf("QueryStageRuns: %v", err)
	}
	var out []store.StageRun
	for _, r := range runs {
		if r.AutoOpened {
			out = append(out, r)
		}
	}
	return out
}

func autoDispatch(token, key string, at time.Time) records.Dispatch {
	d := baseDispatch("implementer", "opus")
	d.SessionToken = token
	d.SessionID = "session-auto"
	d.Key = key
	d.AgentID = "agent-" + key
	d.StartedAt = at
	return d
}

// A dispatch begun with no stage open lands in an auto-opened run whose
// stage is derived from its key and that carries its session.
func TestRecordDispatchOpensAutoStageRun(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	for key, stage := range map[string]string{
		"self-review-kan-1-review-2": "flow.self-review",
		"pipeline-fix-3-plan":        "flow.pipeline-fix",
		"task-6-implementer":         "flow.sdd-tdd",
		"panel-2-principles":         "flow.review-panel",
		"full-suite-fix-1":           "flow.review-panel",
		"visual-verify-1":            "flow.visual-verify",
		"verify-lint":                "flow.verify",
		"something-else":             "flow.sdd-tdd",
	} {
		projectKey := fmt.Sprintf("proj-auto-open-%s-%d", key, time.Now().UnixNano())
		seedChange(t, st, projectKey, "kan-1")
		got, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-1", key, time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)))
		if err != nil {
			t.Fatalf("RecordDispatch %s: %v", key, err)
		}
		runs := autoRuns(t, st, projectKey)
		if len(runs) != 1 {
			t.Fatalf("%s: %d auto runs, want 1", key, len(runs))
		}
		r := runs[0]
		if r.Stage != stage || r.Command != "/flow" || r.EndedAt != nil ||
			r.SessionToken == nil || *r.SessionToken != "mf-auto-1" || r.SessionID == nil || *r.SessionID != "session-auto" {
			t.Errorf("%s: auto run %+v, want open %s /flow carrying the dispatch's session", key, r, stage)
		}
		if got.StageRunID == nil || *got.StageRunID != r.ID {
			t.Errorf("%s: dispatch stage_run_id = %v, want %d", key, got.StageRunID, r.ID)
		}
	}
}

// A dispatch begun while a stage is open opens nothing: it links to that run.
func TestRecordDispatchInsideOpenStageOpensNothing(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-auto-inside-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	begin := baseBeginInput(projectKey, "kan-1", "/flow", "flow.sdd-tdd")
	begin.SessionToken = ptr("mf-auto-2")
	run, err := st.BeginStage(ctx, begin)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	got, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-2", "task-1-implementer", time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	if runs := autoRuns(t, st, projectKey); len(runs) != 0 || got.StageRunID == nil || *got.StageRunID != run.ID {
		t.Errorf("auto runs %+v, stage_run_id %v; want none, linked to %d", runs, got.StageRunID, run.ID)
	}
}

// A dispatch back-dated to before the end of a recorded stage run opens no
// auto run: the harvester has committed past that time and would never fill
// it. One begun at or after the latest end still opens one.
func TestRecordDispatchBackdatedOpensNoAutoStageRun(t *testing.T) {
	t.Parallel()
	st, pool := newRecordStore(t)
	ctx := context.Background()
	day := func(h, m int) time.Time { return time.Date(2026, 8, 22, h, m, 0, 0, time.UTC) }
	for _, tc := range []struct {
		at   time.Time
		want int
	}{{day(9, 30), 0}, {day(10, 0), 1}, {day(10, 30), 1}} {
		projectKey := fmt.Sprintf("proj-auto-backdated-%d", time.Now().UnixNano())
		seedChange(t, st, projectKey, "kan-1")
		var recorded int64
		if err := pool.QueryRow(ctx, `
			INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, metrics, auto_opened)
			SELECT id, 'claude-code', '/flow', 'flow.verify', 1, $2, $3, '{}'::jsonb, false
			FROM changes WHERE project_key = $1 AND name = 'kan-1' RETURNING id`, projectKey, day(9, 0), day(10, 0)).Scan(&recorded); err != nil {
			t.Fatalf("seed stage run: %v", err)
		}
		got, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-1", "task-6-implementer", tc.at))
		if err != nil {
			t.Fatalf("RecordDispatch at %s: %v", tc.at, err)
		}
		runs := autoRuns(t, st, projectKey)
		// At 09:30 the begin lies inside the recorded run and links to it
		// (dispatch_stage_run); afterwards it lands in a new auto run.
		wantID := recorded
		if len(runs) == 1 {
			wantID = runs[0].ID
		}
		if len(runs) != tc.want || got.StageRunID == nil || *got.StageRunID != wantID {
			t.Errorf("dispatch at %s: auto runs %+v, stage_run_id %v; want %d, linked to %d", tc.at, runs, got.StageRunID, tc.want, wantID)
		}
	}
}

// The auto run ends with its last open dispatch, and takes the latest
// run's command.
func TestEndDispatchClosesAutoStageRun(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-auto-end-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	begin := baseBeginInput(projectKey, "kan-1", "/flow-fast", "flow.decide")
	begin.SessionToken = ptr("mf-auto-3")
	run, err := st.BeginStage(ctx, begin)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, begin.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	t0 := time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)
	for _, key := range []string{"task-6-implementer", "task-6-reviewer"} {
		if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-3", key, t0)); err != nil {
			t.Fatalf("RecordDispatch %s: %v", key, err)
		}
	}
	end := func(key string, at time.Time) {
		t.Helper()
		if _, err := st.EndDispatch(ctx, projectKey, "kan-1", records.DispatchEnd{
			SessionToken: "mf-auto-3", Key: key, Outcome: "completed", CommitSHA: "abc1234", EndedAt: at,
		}); err != nil {
			t.Fatalf("EndDispatch %s: %v", key, err)
		}
	}
	end("task-6-reviewer", t0.Add(5*time.Minute))
	runs := autoRuns(t, st, projectKey)
	if len(runs) != 1 || runs[0].EndedAt != nil || runs[0].Command != "/flow-fast" {
		t.Fatalf("after the first end: auto runs %+v, want one open /flow-fast run", runs)
	}
	end("task-6-implementer", t0.Add(10*time.Minute))
	runs = autoRuns(t, st, projectKey)
	if len(runs) != 1 || runs[0].EndedAt == nil || !runs[0].EndedAt.Equal(t0.Add(10*time.Minute)) ||
		runs[0].Outcome == nil || *runs[0].Outcome != "completed" {
		t.Errorf("after the last end: auto runs %+v, want one completed at %s", runs, t0.Add(10*time.Minute))
	}
}

// An explicit stage begin closes a dangling auto run, from another session
// token too.
func TestBeginStageClosesDanglingAutoStageRun(t *testing.T) {
	t.Parallel()
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-auto-supersede-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	t0 := time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-4", "task-1-implementer", t0)); err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	begin := baseBeginInput(projectKey, "kan-1", "/flow", "flow.verify")
	begin.SessionToken = ptr("mf-auto-other")
	begin.StartedAt = t0.Add(time.Hour)
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	runs := autoRuns(t, st, projectKey)
	if len(runs) != 1 || runs[0].EndedAt == nil || !runs[0].EndedAt.Equal(begin.StartedAt) {
		t.Errorf("auto runs %+v, want one closed at the explicit begin %s", runs, begin.StartedAt)
	}
}

// TestAutoStageRunCarriesTheOpenInteractiveRequest pins sr3-auto-run-request
// (kan-1070 self-review, F3): a run auto-opened on a /flow-interactive change
// carries the change's one open request; with that request ended, two open,
// or a /flow change, it carries none.
func TestAutoStageRunCarriesTheOpenInteractiveRequest(t *testing.T) {
	t.Parallel()
	st, pool := newRecordStore(t)
	ctx := context.Background()
	day := func(h, m int) time.Time { return time.Date(2026, 10, 31, h, m, 0, 0, time.UTC) }
	for _, tc := range []struct {
		name, command string
		setup         func(projectKey string)
		want          *int
	}{
		{"interactive, one open", "/flow-interactive", func(pk string) {
			mustBegin(t, st, pk, 6)
		}, ptr(6)},
		{"interactive, request ended", "/flow-interactive", func(pk string) {
			mustBegin(t, st, pk, 6)
			if _, err := st.EndRequest(ctx, pk, "kan-1", records.RequestEnd{N: 6, Outcome: "answered"}); err != nil {
				t.Fatalf("EndRequest: %v", err)
			}
		}, nil},
		{"interactive, two open", "/flow-interactive", func(pk string) {
			mustBegin(t, st, pk, 6)
			mustBegin(t, st, pk, 7)
		}, nil},
		{"flow, one open", "/flow", func(pk string) {
			mustBegin(t, st, pk, 6)
		}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projectKey := fmt.Sprintf("proj-auto-request-%d", time.Now().UnixNano())
			seedChange(t, st, projectKey, "kan-1")
			if _, err := pool.Exec(ctx, `
				INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, metrics, auto_opened)
				SELECT id, 'claude-code', $2, 'flow.load-context', 1, $3, $4, '{}'::jsonb, false
				FROM changes WHERE project_key = $1 AND name = 'kan-1'`, projectKey, tc.command, day(12, 58), day(12, 59)); err != nil {
				t.Fatalf("seed stage run: %v", err)
			}
			tc.setup(projectKey)
			if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("fi-auto-1", "r6-draft", day(13, 0))); err != nil {
				t.Fatalf("RecordDispatch: %v", err)
			}
			runs := autoRuns(t, st, projectKey)
			if len(runs) != 1 {
				t.Fatalf("%d auto runs, want 1", len(runs))
			}
			got := runs[0].RequestN
			if (got == nil) != (tc.want == nil) || (got != nil && *got != *tc.want) {
				t.Errorf("auto run request_n = %v, want %v", deref(got), deref(tc.want))
			}
		})
	}
}

func mustBegin(t *testing.T, st *store.Store, projectKey string, n int) {
	t.Helper()
	if _, err := st.BeginRequest(context.Background(), projectKey, "kan-1", records.RequestBegin{N: n, Title: fmt.Sprintf("r%d", n)}); err != nil {
		t.Fatalf("BeginRequest %d: %v", n, err)
	}
}

func deref(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestReBeginRelinksToTheRunHoldingTheNewBegin(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-rebegin-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	day := func(h, m int) time.Time { return time.Date(2026, 8, 22, h, m, 0, 0, time.UTC) }
	var a, b int64
	ins := `INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, metrics, auto_opened, session_token)
		SELECT id, 'claude-code', '/flow', 'flow.sdd-tdd', $4, $2, $3, '{}'::jsonb, false, 'mf-rb'
		FROM changes WHERE project_key = $1 AND name = 'kan-1' RETURNING id`
	if err := pool.QueryRow(ctx, ins, projectKey, day(9, 0), day(10, 0), 1).Scan(&a); err != nil {
		t.Fatalf("seed run a: %v", err)
	}
	if err := pool.QueryRow(ctx, ins, projectKey, day(11, 0), nil, 2).Scan(&b); err != nil {
		t.Fatalf("seed run b: %v", err)
	}
	first, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-rb", "task-1-reviewer", day(9, 30)))
	if err != nil {
		t.Fatalf("first begin: %v", err)
	}
	if first.StageRunID == nil || *first.StageRunID != a {
		t.Fatalf("first begin linked %v, want %d", first.StageRunID, a)
	}
	if _, err := pool.Exec(ctx, `UPDATE dispatches SET outcome='aborted', ended_at=$2 WHERE id=$1`, first.ID, day(9, 31)); err != nil {
		t.Fatalf("abort: %v", err)
	}
	again, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-rb", "task-1-reviewer", day(11, 30)))
	if err != nil {
		t.Fatalf("re-begin: %v", err)
	}
	if again.ID != first.ID {
		t.Fatalf("re-begin made a new row %d", again.ID)
	}
	if again.StageRunID == nil || *again.StageRunID != b {
		t.Errorf("re-begun at %s: stage_run_id %v, want %d (the run holding the new begin); still linked to run %d ended %s",
			again.StartedAt.Format(time.Kitchen), relinkDeref(again.StageRunID), b, a, day(10, 0).Format(time.Kitchen))
	}
}

func relinkDeref(p *int64) any {
	if p == nil {
		return nil
	}
	return *p
}

func TestRetimeRelinksToTheRunHoldingTheNewBegin(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-retime-link-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	at := func(h, m int) time.Time { return time.Date(2026, 8, 22, h, m, 0, 0, time.UTC) }
	ins := `INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, metrics, auto_opened, session_token)
		SELECT id, 'claude-code', '/flow', $4, 1, $2, $3, '{}'::jsonb, false, 'mf-rt'
		FROM changes WHERE project_key = $1 AND name = 'kan-1' RETURNING id`
	var a, b int64
	if err := pool.QueryRow(ctx, ins, projectKey, at(9, 0), at(10, 0), "flow.sdd-tdd").Scan(&a); err != nil {
		t.Fatalf("seed a: %v", err)
	}
	if err := pool.QueryRow(ctx, ins, projectKey, at(11, 0), at(12, 0), "flow.verify").Scan(&b); err != nil {
		t.Fatalf("seed b: %v", err)
	}
	d, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-rt", "verify-lint", at(9, 30)))
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	if d.StageRunID == nil || *d.StageRunID != a {
		t.Fatalf("begin linked %v, want %d", d.StageRunID, a)
	}
	got, err := st.RetimeDispatch(ctx, projectKey, "kan-1", records.DispatchRetime{ID: d.ID, StartedAt: at(11, 10), EndedAt: at(11, 20), Reason: "late begin"})
	if err != nil {
		t.Fatalf("retime: %v", err)
	}
	if got.StageRunID == nil || *got.StageRunID != b {
		var v any
		if got.StageRunID != nil {
			v = *got.StageRunID
		}
		t.Errorf("retimed into run %d's window: stage_run_id %v -- STALE LINK", b, v)
	}
}
