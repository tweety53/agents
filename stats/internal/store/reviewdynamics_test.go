package store_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// dynamicsFixture seeds the three changes TestReviewDynamicsSeries and
// TestReviewDynamicsHeatSummaryMarkers read: A and B reviewed in the week
// of 2026-06-01, C in the week of 2026-06-08.
type dynamicsFixture struct {
	t       *testing.T
	st      *store.Store
	project string
}

var (
	week1 = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	week2 = time.Date(2026, 6, 8, 0, 0, 0, 0, time.UTC)
)

func (fx dynamicsFixture) dispatch(change, role, slot, model, effort, key, outcome string, at time.Time) int {
	fx.t.Helper()
	d := baseDispatch(role, model)
	d.Slot, d.Effort, d.Key, d.Outcome, d.StartedAt = slot, effort, key, outcome, at
	out, err := fx.st.RecordDispatch(context.Background(), fx.project, change, d)
	if err != nil {
		fx.t.Fatalf("RecordDispatch %s %s: %v", change, key, err)
	}
	return out.Seq
}

func (fx dynamicsFixture) finding(change, ref, slot, severity string, round int, seq *int) {
	fx.t.Helper()
	f := baseFinding(ref, round)
	f.Slot, f.Severity, f.DispatchSeq = slot, severity, seq
	if _, _, err := fx.st.UpsertFinding(context.Background(), fx.project, change, f); err != nil {
		fx.t.Fatalf("UpsertFinding %s %s: %v", change, ref, err)
	}
}

func seedDynamics(t *testing.T) (*store.Store, string) {
	t.Helper()
	st, _ := newRecordStore(t)
	fx := dynamicsFixture{t: t, st: st, project: fmt.Sprintf("proj-dynamics-%d", time.Now().UnixNano())}
	for _, c := range []string{"a", "b", "c"} {
		seedChange(t, st, fx.project, c)
	}
	day := func(base time.Time, d int) time.Time { return base.Add(time.Duration(d)*24*time.Hour + 9*time.Hour) }

	// Change A, week 1: one bundled panel dispatch, findings over rounds 0
	// and 2, two implementer dispatches (one blocked), three verifier runs.
	a1 := fx.dispatch("a", "reviewer", "primary+principles", "opus", "medium", "panel-0-primary+principles", "completed", day(week1, 0))
	fx.finding("a", "A1", "primary", "Critical", 0, &a1)
	fx.finding("a", "A2", "principles", "major", 0, &a1)
	fx.finding("a", "A3", "primary", "minor", 2, &a1)
	fx.dispatch("a", "implementer", "", "opus", "low", "task-1-implementer", "completed", day(week1, 0).Add(-2*time.Hour))
	fx.dispatch("a", "implementer", "", "opus", "low", "task-2-implementer", "blocked", day(week1, 0).Add(-time.Hour))
	fx.dispatch("a", "verifier", "", "opus", "low", "visual-verify", "completed", day(week1, 1))
	fx.dispatch("a", "verifier", "", "opus", "low", "visual-verify-fix-1", "completed", day(week1, 1).Add(time.Hour))
	fx.dispatch("a", "verifier", "", "opus", "low", "visual-verify-fix-2", "completed", day(week1, 1).Add(2*time.Hour))

	// Change B, week 1: one clean test-sanity dispatch, nothing else.
	fx.dispatch("b", "reviewer", "test-sanity", "sonnet", "low", "panel-0-test-sanity", "completed", day(week1, 2))

	// Change C, week 2: one primary dispatch, one finding no dispatch
	// carries, one verifier run that passed first time.
	fx.dispatch("c", "reviewer", "primary", "sonnet", "high", "panel-0-primary", "completed", day(week2, 1))
	fx.finding("c", "C1", "primary", "Minor", 0, nil)
	fx.dispatch("c", "verifier", "", "opus", "low", "visual-verify", "completed", day(week2, 2))

	return st, fx.project
}

type pointKey struct {
	bucket         time.Time
	metric, series string
}

// TestReviewDynamicsSeries pins every series formula of
// Store.ReviewDynamics (cohort-by-change, impl-vv-metrics,
// bucket-from-range) over a two-week fixture.
func TestReviewDynamicsSeries(t *testing.T) {
	st, project := seedDynamics(t)
	ctx := context.Background()
	period := store.Period{From: week1, To: week1.Add(15 * 24 * time.Hour)}

	got, err := st.ReviewDynamics(ctx, period, &project)
	if err != nil {
		t.Fatalf("ReviewDynamics: %v", err)
	}
	if got.Bucket != "week" {
		t.Errorf("Bucket = %q, want week for a 15-day period", got.Bucket)
	}
	points := map[pointKey]store.DynamicsPoint{}
	for _, p := range got.Series {
		points[pointKey{p.Bucket.UTC(), p.Metric, p.Series}] = p
	}
	want := []struct {
		bucket         time.Time
		metric, series string
		value          float64
		basis          int
	}{
		{week1, "findings-per-change", "critical", 0.5, 2},
		{week1, "findings-per-change", "important", 0.5, 2},
		{week1, "findings-per-change", "minor", 0.5, 2},
		{week2, "findings-per-change", "minor", 1, 1},
		{week2, "findings-per-change", "critical", 0, 1},
		{week1, "fix-rounds", "mean", 1, 2},
		{week1, "fix-rounds", "max", 2, 2},
		{week1, "changes", "changes", 2, 2},
		{week2, "changes", "changes", 1, 1},
		{week1, "per-dispatch-slot", "primary", 2, 1},
		{week1, "per-dispatch-slot", "principles", 1, 1},
		{week1, "per-dispatch-slot", "test-sanity", 0, 1},
		{week2, "per-dispatch-slot", "primary", 0, 1},
		{week1, "per-dispatch-model", "opus", 3, 1},
		{week1, "per-dispatch-model", "sonnet", 0, 1},
		{week1, "per-dispatch-effort", "medium", 3, 1},
		{week1, "per-dispatch-effort", "low", 0, 1},
		{week1, "impl-findings-per-task", "opus · low", 1.5, 2},
		{week1, "impl-blocked-share", "blocked", 0.5, 2},
		{week1, "vv-rounds", "mean", 2, 1},
		{week1, "vv-rounds", "max", 2, 1},
		{week2, "vv-rounds", "mean", 0, 1},
		{week1, "vv-first-pass", "first-pass", 0, 1},
		{week2, "vv-first-pass", "first-pass", 1, 1},
	}
	for _, w := range want {
		p, ok := points[pointKey{w.bucket, w.metric, w.series}]
		if !ok {
			t.Errorf("no point %s %s/%s", w.bucket.Format("2006-01-02"), w.metric, w.series)
			continue
		}
		if math.Abs(p.Value-w.value) > 1e-9 || p.Basis != w.basis {
			t.Errorf("%s %s/%s = %v (basis %d), want %v (basis %d)",
				w.bucket.Format("2006-01-02"), w.metric, w.series, p.Value, p.Basis, w.value, w.basis)
		}
	}

	short, err := st.ReviewDynamics(ctx, store.Period{From: week1, To: week1.Add(14 * 24 * time.Hour)}, &project)
	if err != nil {
		t.Fatalf("ReviewDynamics 14 days: %v", err)
	}
	if short.Bucket != "day" {
		t.Errorf("Bucket = %q, want day for a 14-day period", short.Bucket)
	}

	other := "proj-dynamics-none"
	empty, err := st.ReviewDynamics(ctx, period, &other)
	if err != nil {
		t.Fatalf("ReviewDynamics other project: %v", err)
	}
	if len(empty.Series) != 0 || len(empty.Heat) != 0 {
		t.Errorf("another project's dynamics = %d points, %d heat rows; want none", len(empty.Series), len(empty.Heat))
	}
}

// TestReviewDynamicsHeatSummaryMarkers pins the heatmap rows, the summary
// against the previous range of the same length (previous-range) and the
// prompt-change markers (markers-by-area).
func TestReviewDynamicsHeatSummaryMarkers(t *testing.T) {
	st, project := seedDynamics(t)
	ctx := context.Background()
	slot := "primary"
	for _, pc := range []store.PromptChange{
		{SHA: "s1:skills/flow/primary-reviewer-prompt.md", At: week2.Add(time.Hour), Kind: "edit", Area: "review",
			Slot: &slot, Target: "skills/flow/primary-reviewer-prompt.md", Summary: "in range"},
		{SHA: "s2:skills/flow/visual-verify-verifier.md", At: week2.Add(30 * 24 * time.Hour), Kind: "edit", Area: "vv",
			Target: "skills/flow/visual-verify-verifier.md", Summary: "later"},
	} {
		if err := st.UpsertPromptChange(ctx, pc); err != nil {
			t.Fatalf("UpsertPromptChange: %v", err)
		}
	}

	// Week 2 alone, so week 1 is the previous range.
	got, err := st.ReviewDynamics(ctx, store.Period{From: week2, To: week2.Add(7 * 24 * time.Hour)}, &project)
	if err != nil {
		t.Fatalf("ReviewDynamics: %v", err)
	}
	if got.Bucket != "day" {
		t.Errorf("Bucket = %q, want day", got.Bucket)
	}

	wantHeat := []store.HeatCell{{Slot: "primary", Model: "sonnet", Effort: "high", Dispatches: 1}}
	if fmt.Sprint(got.Heat) != fmt.Sprint(wantHeat) {
		t.Errorf("Heat = %+v, want %+v", got.Heat, wantHeat)
	}
	wantPrev := []store.HeatCell{
		{Slot: "primary", Model: "opus", Effort: "medium", Dispatches: 1, Critical: 1, Minor: 1},
		{Slot: "principles", Model: "opus", Effort: "medium", Dispatches: 1, Important: 1},
		{Slot: "test-sanity", Model: "sonnet", Effort: "low", Dispatches: 1},
	}
	if fmt.Sprint(got.PreviousHeat) != fmt.Sprint(wantPrev) {
		t.Errorf("PreviousHeat = %+v, want %+v", got.PreviousHeat, wantPrev)
	}

	s, p := got.Summary, got.PreviousSummary
	if s.Changes != 1 || s.FindingsPerChange != 1 || s.CriticalImportantPerChange != 0 || s.MeanFixRounds != 0 {
		t.Errorf("Summary = %+v, want 1 change, 1 finding, 0 critical+important, 0 rounds", s)
	}
	if s.ImplFindingsPerTask != nil {
		t.Errorf("Summary.ImplFindingsPerTask = %v, want nil (no implementer dispatch in range)", *s.ImplFindingsPerTask)
	}
	if s.VVFirstPassShare == nil || *s.VVFirstPassShare != 1 {
		t.Errorf("Summary.VVFirstPassShare = %v, want 1", s.VVFirstPassShare)
	}
	if p.Changes != 2 || p.FindingsPerChange != 1.5 || p.CriticalImportantPerChange != 1 || p.MeanFixRounds != 1 {
		t.Errorf("PreviousSummary = %+v, want 2 changes, 1.5 findings, 1 critical+important, 1 round", p)
	}
	if p.ImplFindingsPerTask == nil || *p.ImplFindingsPerTask != 1.5 {
		t.Errorf("PreviousSummary.ImplFindingsPerTask = %v, want 1.5", p.ImplFindingsPerTask)
	}
	if p.VVFirstPassShare == nil || *p.VVFirstPassShare != 0 {
		t.Errorf("PreviousSummary.VVFirstPassShare = %v, want 0", p.VVFirstPassShare)
	}

	if len(got.Markers) != 1 || got.Markers[0].Summary != "in range" {
		t.Errorf("Markers = %+v, want the one in-range change", got.Markers)
	}
	if got.LastPromptChangeAt == nil || !got.LastPromptChangeAt.Equal(week2.Add(30*24*time.Hour)) {
		t.Errorf("LastPromptChangeAt = %v, want the newest change whatever the range", got.LastPromptChangeAt)
	}
}

// TestReviewDynamicsModelVersions pins model-version-split: per-model
// series and the heatmap exist keyed by family (versions merged) and by
// version, a dispatch's version read from its harvested modelVersion, its
// full model ID, or "<family> (version unrecorded)".
func TestReviewDynamicsModelVersions(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	fx := dynamicsFixture{t: t, st: st, project: fmt.Sprintf("proj-versions-%d", time.Now().UnixNano())}
	seedChange(t, st, fx.project, "v")
	at := week1.Add(9 * time.Hour)

	fx.dispatch("v", "reviewer", "primary", "opus", "low", "panel-0-a", "completed", at)
	fx.dispatch("v", "reviewer", "primary", "claude-opus-5", "low", "panel-0-b", "completed", at.Add(time.Hour))
	fx.dispatch("v", "reviewer", "primary", "opus", "low", "panel-0-c", "completed", at.Add(2*time.Hour))
	fx.dispatch("v", "implementer", "", "opus", "low", "task-1-implementer", "completed", at.Add(-time.Hour))
	for _, key := range []string{"panel-0-a", "task-1-implementer"} {
		if _, err := pool.Exec(ctx, `UPDATE dispatches SET metrics = metrics || '{"modelVersion":"claude-opus-5-5"}'
			WHERE dispatch_key = $1 AND change_id = (SELECT id FROM changes WHERE name = 'v' AND project_key = $2)`, key, fx.project); err != nil {
			t.Fatalf("stamp %s: %v", key, err)
		}
	}

	got, err := st.ReviewDynamics(ctx, store.Period{From: week1, To: week1.Add(7 * 24 * time.Hour)}, &fx.project)
	if err != nil {
		t.Fatalf("ReviewDynamics: %v", err)
	}
	series := map[string]map[string]int{}
	for _, p := range got.Series {
		if series[p.Metric] == nil {
			series[p.Metric] = map[string]int{}
		}
		series[p.Metric][p.Series] += p.Basis
	}
	if fmt.Sprint(series["per-dispatch-model"]) != "map[opus:3]" {
		t.Errorf("per-dispatch-model = %v, want one opus series over all three dispatches", series["per-dispatch-model"])
	}
	if fmt.Sprint(series["per-dispatch-model-version"]) != "map[claude-opus-5:1 claude-opus-5-5:1 opus (version unrecorded):1]" {
		t.Errorf("per-dispatch-model-version = %v", series["per-dispatch-model-version"])
	}
	if _, ok := series["impl-findings-per-task"]["opus · low"]; !ok {
		t.Errorf("impl-findings-per-task = %v, want opus · low", series["impl-findings-per-task"])
	}
	if _, ok := series["impl-findings-per-task-version"]["claude-opus-5-5 · low"]; !ok {
		t.Errorf("impl-findings-per-task-version = %v, want claude-opus-5-5 · low", series["impl-findings-per-task-version"])
	}
	if len(got.Heat) != 1 || got.Heat[0].Model != "opus" || got.Heat[0].Dispatches != 3 {
		t.Errorf("Heat = %+v, want one opus cell of 3 dispatches", got.Heat)
	}
	if len(got.HeatVersion) != 3 {
		t.Errorf("HeatVersion = %+v, want three version cells", got.HeatVersion)
	}

	for in, want := range map[string]string{
		"opus": "opus", "claude-opus-5-5": "opus", "Opus": "opus", " Opus ": "opus", "claude-haiku-5-5": "haiku", "fable": "fable", "claude-sonnet-5": "sonnet",
		"glm-5.3-flash": "glm", "GLM-5.3-Flash (inline session)": "glm", "builtin:zai-coding-plan/GLM-5.3-Flash": "glm",
	} {
		var got string
		if err := pool.QueryRow(ctx, "SELECT model_family($1)", in).Scan(&got); err != nil || got != want {
			t.Errorf("model_family(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
}
