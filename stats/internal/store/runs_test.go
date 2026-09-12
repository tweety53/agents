package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

func mergeMetrics(t *testing.T, st *store.Store, id int64, s string) {
	t.Helper()
	if err := st.MergeMetrics(context.Background(), id, json.RawMessage(s)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
}

func TestListRunsGroupsByChangeAndSessionToken(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-runs-%d", time.Now().UnixNano())
	t0 := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)

	// Plan session first, unattached.
	plan, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey: projectKey, MainCheckoutPath: "/tmp/" + projectKey, JiraKey: "KAN-950",
		Harness: "claude-code", SessionToken: ptr("fp-runs-plan"), Command: "/flow-plan", Stage: "plan.session", StartedAt: t0,
	})
	if err != nil {
		t.Fatalf("plan BeginStage: %v", err)
	}
	mergeMetrics(t, st, plan.ID, `{"tokens":{"main":{"input":100,"output":10,"cache_read":1000}},"cost_usd":0.5,
		"signals":{"main":{"compactions":1,"turns":4,"tool_calls_total":6,"tool_errors":1,"denials":0,"api_errors":0,"context_end":"1100"}}}`)
	if err := st.EndStage(ctx, plan.ID, t0.Add(20*time.Minute), "staged"); err != nil {
		t.Fatalf("plan EndStage: %v", err)
	}

	c := baseChange(projectKey, "kan-950-runs")
	c.JiraIssue = ptr("KAN-950")
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	// Creating run: kickoff, design-approval (human gate), sdd-tdd with one dispatch.
	const token = "mf-runs-1"
	begin := func(stage string, at time.Time) store.StageRun {
		t.Helper()
		in := baseBeginInput(projectKey, "kan-950-runs", "/flow", stage)
		in.SessionToken = ptr(token)
		in.StartedAt = at
		run, err := st.BeginStage(ctx, in)
		if err != nil {
			t.Fatalf("BeginStage %s: %v", stage, err)
		}
		return run
	}
	kick := begin("flow.kickoff", t0.Add(time.Hour))
	if err := st.EndStage(ctx, kick.ID, t0.Add(61*time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}
	gate := begin("flow.design-approval", t0.Add(61*time.Minute))
	if err := st.EndStage(ctx, gate.ID, t0.Add(71*time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}
	impl := begin("flow.sdd-tdd", t0.Add(71*time.Minute))
	mergeMetrics(t, st, impl.ID, `{"tokens":{"main":{"input":200,"output":20,"cache_read":2000},"sidechain":{"input":300,"output":30,"cache_read":3000}},"cost_usd":2.0,
		"signals":{"main":{"compactions":0,"turns":10,"tool_calls_total":20},"sidechain":{"turns":5,"tool_calls_total":9}},
		"dispatches":{"agent-r1":{"cost_usd":1.25,"tokens":{"sidechain":{"input":300,"output":30,"cache_read":3000}},"agent_type":"flow-medium","spawn_depth":"1",
			"signals":{"turns":5,"tool_calls_total":9,"served_models":{"claude-sonnet-5":5},"served_efforts":{"medium":5}}}}}`)
	if err := st.EndStage(ctx, impl.ID, t0.Add(101*time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}
	d := baseDispatch("implementer", "claude-opus-5")
	d.SessionToken = token
	d.AgentID = "agent-r1"
	d.Effort = "medium"
	d.StartedAt = t0.Add(75 * time.Minute)
	d.Metrics = json.RawMessage(`{"tokens":{"sidechain":{"input":300,"output":30,"cache_read":3000}},"signals":{"sidechain":{"turns":5,"tool_calls_total":9,"served_models":{"claude-sonnet-5":5},"served_efforts":{"medium":5}}}}`)
	implOut, err := st.RecordDispatch(ctx, projectKey, "kan-950-runs", d)
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	for i, status := range []string{"fixed", "fixed", "open"} {
		f := baseFinding(fmt.Sprintf("F%d", i+1), 1)
		f.DispatchSeq = &implOut.Seq
		f.Severity = "Minor"
		f.Status = status
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-950-runs", f); err != nil {
			t.Fatalf("UpsertFinding: %v", err)
		}
	}
	if _, err := st.InsertSuiteRun(ctx, projectKey, records.SuiteRun{Suite: "go", Host: "ci", DurationMs: 1000, ExitCode: 1, RanAt: t0.Add(80 * time.Minute)}); err != nil {
		t.Fatalf("InsertSuiteRun: %v", err)
	}
	if _, err := st.InsertSuiteRun(ctx, projectKey, records.SuiteRun{Suite: "go", Host: "ci", DurationMs: 1000, ExitCode: 0, RanAt: t0.Add(90 * time.Minute)}); err != nil {
		t.Fatalf("InsertSuiteRun: %v", err)
	}
	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-950-runs", records.Decision{
		SessionToken: token,
		Decision:     json.RawMessage(`{"class":"regular","execution":"subagent","implementer":{"model":"claude-opus-5","effort":"medium"},"panel":{"roster":[{"slot":"primary"}],"compact":true}}`),
	}); err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}

	// A fix run two hours later.
	fixIn := baseBeginInput(projectKey, "kan-950-runs", "/flow", "flow.document-fix")
	fixIn.SessionToken = ptr("mf-runs-2")
	fixIn.StartedAt = t0.Add(4 * time.Hour)
	fix, err := st.BeginStage(ctx, fixIn)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, fix.ID, t0.Add(4*time.Hour+10*time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}

	// decisions.recorded_at takes the column's own now() default rather
	// than a caller-supplied timestamp (RecordDecision), so the period's
	// upper bound has to reach real "now" too, not just t0's fictional
	// window -- every other row here is dated from t0, and only the
	// decision is dated from the actual clock.
	rows, err := st.ListRuns(ctx, store.Period{From: t0.Add(-time.Hour), To: time.Now().Add(time.Hour)}, &projectKey, nil)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d change groups, want 1: %+v", len(rows), rows)
	}
	g := rows[0]
	if g.Change == nil || *g.Change != "kan-950-runs" || g.JiraKey == nil || *g.JiraKey != "KAN-950" {
		t.Errorf("group identity = %+v", g)
	}
	if len(g.Runs) != 3 {
		t.Fatalf("got %d runs, want 3 (plan, creating, fix)", len(g.Runs))
	}
	if g.Runs[0].Kind != "plan" || g.Runs[1].Kind != "flow" || g.Runs[2].Kind != "flow" {
		t.Errorf("kinds = %s/%s/%s", g.Runs[0].Kind, g.Runs[1].Kind, g.Runs[2].Kind)
	}
	if g.FixIterations != 1 {
		t.Errorf("FixIterations = %d, want 1", g.FixIterations)
	}
	// idle: plan end 10:20 -> creating start 11:00 = 40m; creating end 11:41 -> fix start 14:00 = 139m.
	if g.IdleBetweenRunsMs != (40+139)*60*1000 {
		t.Errorf("IdleBetweenRunsMs = %d", g.IdleBetweenRunsMs)
	}

	cr := g.Runs[1]
	if cr.Totals.InputTokens != 500 || cr.Totals.CacheRead != 5000 || cr.Totals.CostUSD == nil || *cr.Totals.CostUSD != 2.0 {
		t.Errorf("creating run totals = %+v", cr.Totals)
	}
	if cr.Main.InputTokens != 200 || cr.Main.Turns != 10 || cr.Main.CostUSD == nil || *cr.Main.CostUSD != 0.75 {
		t.Errorf("creating run main = %+v", cr.Main)
	}
	if cr.Totals.HumanGateMs != 10*60*1000 || cr.Totals.WallClockMs != 41*60*1000 {
		t.Errorf("gate/wall = %d/%d", cr.Totals.HumanGateMs, cr.Totals.WallClockMs)
	}
	if cr.Decision == nil || cr.Decision.Execution != "subagent" || cr.Decision.Implementer != "claude-opus-5/medium" || cr.Decision.Panel != "1 slot, compact" {
		t.Errorf("decision = %+v", cr.Decision)
	}
	if len(cr.Dispatches) != 1 {
		t.Fatalf("dispatches = %+v", cr.Dispatches)
	}
	dr := cr.Dispatches[0]
	if dr.DeclaredModel != "claude-opus-5" || dr.ServedModels["claude-sonnet-5"] != 5 || !dr.Mismatch {
		t.Errorf("dispatch model mismatch not flagged: %+v", dr)
	}
	if dr.Totals.InputTokens != 300 || dr.Totals.Turns != 5 || dr.Totals.CostUSD == nil || *dr.Totals.CostUSD != 1.25 || dr.AgentType != "flow-medium" {
		t.Errorf("dispatch totals = %+v", dr)
	}
	if dr.FindingsRaised != 3 || dr.FindingsByStatus["fixed"] != 2 || dr.FindingsByStatus["open"] != 1 {
		t.Errorf("dispatch findings = %d %v", dr.FindingsRaised, dr.FindingsByStatus)
	}
	if cr.FanOutMax != 1 || cr.SuiteRuns != 2 || cr.SuiteFirstPass == nil || *cr.SuiteFirstPass {
		t.Errorf("fan-out/suite = %d/%d/%v, want 1/2/false", cr.FanOutMax, cr.SuiteRuns, cr.SuiteFirstPass)
	}

	pr := g.Runs[0]
	if pr.Totals.Compactions != 1 || pr.Totals.ContextEnd == nil || *pr.Totals.ContextEnd != 1100 || pr.Decision != nil {
		t.Errorf("plan run = %+v", pr)
	}
}

func TestListRunsListsAnUnattachedPlanSessionUnderItsJiraKey(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-runs-orphan-%d", time.Now().UnixNano())
	t0 := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey: projectKey, MainCheckoutPath: "/tmp/" + projectKey, JiraKey: "KAN-951",
		Harness: "claude-code", SessionToken: ptr("fp-orphan"), Command: "/flow-plan", Stage: "plan.session", StartedAt: t0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, run.ID, t0.Add(time.Minute), "abandoned"); err != nil {
		t.Fatal(err)
	}
	rows, err := st.ListRuns(ctx, store.Period{From: t0.Add(-time.Hour), To: t0.Add(time.Hour)}, &projectKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Change != nil || rows[0].JiraKey == nil || *rows[0].JiraKey != "KAN-951" || len(rows[0].Runs) != 1 {
		t.Fatalf("rows = %+v", rows)
	}
}

// TestFanOutMaxTieBreaksCloseBeforeOpen covers the edge fanOutMax's own
// sort comparator names in its "close before open" tie-break: a third
// dispatch that starts at the exact instant a second one ends must not be
// counted as briefly overlapping it. Two of the three dispatches do
// genuinely overlap, so the maximum is 2, never 3.
func TestFanOutMaxTieBreaksCloseBeforeOpen(t *testing.T) {
	t0 := time.Date(2026, 9, 3, 9, 0, 0, 0, time.UTC)
	end := func(at time.Time) *time.Time { return &at }
	got := store.FanOutMaxForTest([]store.DispatchWindowForTest{
		{StartedAt: t0, EndedAt: end(t0.Add(10 * time.Minute))},
		{StartedAt: t0.Add(5 * time.Minute), EndedAt: end(t0.Add(15 * time.Minute))},
		{StartedAt: t0.Add(15 * time.Minute), EndedAt: end(t0.Add(20 * time.Minute))},
	})
	if got != 2 {
		t.Errorf("FanOutMax = %d, want 2", got)
	}
}

// TestListRunsFiltersByChangeName asserts the change parameter narrows
// ListRuns to that one change's group, and a change name the project holds
// no runs under returns zero groups rather than every change's.
func TestListRunsFiltersByChangeName(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-runs-filter-%d", time.Now().UnixNano())
	t0 := time.Date(2026, 9, 4, 9, 0, 0, 0, time.UTC)

	seedChange(t, st, projectKey, "kan-960-a")
	seedChange(t, st, projectKey, "kan-961-b")

	beginIn := func(name, token string) store.BeginStageInput {
		in := baseBeginInput(projectKey, name, "/flow", "flow.kickoff")
		in.SessionToken = ptr(token)
		in.StartedAt = t0
		return in
	}
	runA, err := st.BeginStage(ctx, beginIn("kan-960-a", "mf-filter-a"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, runA.ID, t0.Add(time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}
	runB, err := st.BeginStage(ctx, beginIn("kan-961-b", "mf-filter-b"))
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, runB.ID, t0.Add(time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}

	period := store.Period{From: t0.Add(-time.Hour), To: t0.Add(time.Hour)}

	rows, err := st.ListRuns(ctx, period, &projectKey, ptr("kan-960-a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Change == nil || *rows[0].Change != "kan-960-a" {
		t.Fatalf("filtered rows = %+v", rows)
	}

	none, err := st.ListRuns(ctx, period, &projectKey, ptr("kan-962-nonexistent"))
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Fatalf("got %d groups for a change with no runs, want 0: %+v", len(none), none)
	}
}

// TestListRunsDetachesTokenlessStageRuns is the controller's ruling for a
// token-less stage run: it must carry nothing of the change's other
// token-less rows. Two token-less stage runs plus one token-less dispatch
// must come back as two runs, each with zero dispatches -- not one run
// inheriting the dispatch and the other duplicating it -- and neither
// counts toward FixIterations.
func TestListRunsDetachesTokenlessStageRuns(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-runs-notoken-%d", time.Now().UnixNano())
	t0 := time.Date(2026, 9, 5, 9, 0, 0, 0, time.UTC)
	seedChange(t, st, projectKey, "kan-970-notoken")

	first := baseBeginInput(projectKey, "kan-970-notoken", "/flow", "flow.kickoff")
	first.StartedAt = t0
	run1, err := st.BeginStage(ctx, first)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, run1.ID, t0.Add(time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}

	second := baseBeginInput(projectKey, "kan-970-notoken", "/flow", "flow.document-fix")
	second.StartedAt = t0.Add(time.Hour)
	run2, err := st.BeginStage(ctx, second)
	if err != nil {
		t.Fatal(err)
	}
	if err := st.EndStage(ctx, run2.ID, t0.Add(time.Hour+time.Minute), "completed"); err != nil {
		t.Fatal(err)
	}

	d := baseDispatch("implementer", "claude-opus-5")
	d.StartedAt = t0.Add(30 * time.Second)
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-970-notoken", d); err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}

	rows, err := st.ListRuns(ctx, store.Period{From: t0.Add(-time.Hour), To: t0.Add(2 * time.Hour)}, &projectKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("got %d change groups, want 1: %+v", len(rows), rows)
	}
	g := rows[0]
	if len(g.Runs) != 2 {
		t.Fatalf("got %d runs, want 2: %+v", len(g.Runs), g.Runs)
	}
	for _, run := range g.Runs {
		if len(run.Dispatches) != 0 {
			t.Errorf("run %+v carries %d dispatches, want 0", run, len(run.Dispatches))
		}
	}
	if g.FixIterations != 0 {
		t.Errorf("FixIterations = %d, want 0", g.FixIterations)
	}
}
