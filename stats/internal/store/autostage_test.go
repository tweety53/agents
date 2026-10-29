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

// A dispatch begun while a stage is open opens nothing.
func TestRecordDispatchInsideOpenStageOpensNothing(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-auto-inside-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	begin := baseBeginInput(projectKey, "kan-1", "/flow", "flow.sdd-tdd")
	begin.SessionToken = ptr("mf-auto-2")
	if _, err := st.BeginStage(ctx, begin); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	got, err := st.RecordDispatch(ctx, projectKey, "kan-1", autoDispatch("mf-auto-2", "task-1-implementer", time.Date(2026, 8, 22, 9, 0, 0, 0, time.UTC)))
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	if runs := autoRuns(t, st, projectKey); len(runs) != 0 || got.StageRunID != nil {
		t.Errorf("auto runs %+v, stage_run_id %v; want none", runs, got.StageRunID)
	}
}

// The auto run ends with its last open dispatch, and takes the latest
// run's command.
func TestEndDispatchClosesAutoStageRun(t *testing.T) {
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
