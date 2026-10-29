package store_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// statsFixture seeds changes, stage runs, dispatches, findings and task
// counts around one experiment whose start T is a week before now, so the
// after window is [T, now) and the before window [T-7d, T).
type statsFixture struct {
	t       *testing.T
	st      *store.Store
	pool    *pgxpool.Pool
	project string
	now, T  time.Time
}

func newStatsFixture(t *testing.T) statsFixture {
	t.Helper()
	st, pool := newRecordStore(t)
	now := time.Now().UTC()
	return statsFixture{t: t, st: st, pool: pool, project: fmt.Sprintf("proj-exp-%d", now.UnixNano()), now: now, T: now.Add(-7 * 24 * time.Hour)}
}

// open opens an experiment over cells and moves its start to fx.T.
func (fx statsFixture) open(cells map[string][]string) int64 {
	fx.t.Helper()
	ctx := context.Background()
	e, err := fx.st.OpenExperiment(ctx, store.Experiment{Name: "t", Cells: cells})
	if err != nil {
		fx.t.Fatalf("OpenExperiment: %v", err)
	}
	if _, err := fx.pool.Exec(ctx, `UPDATE experiments SET started_at = $2 WHERE id = $1`, e.ID, fx.T); err != nil {
		fx.t.Fatalf("move experiment start: %v", err)
	}
	return e.ID
}

// change seeds a change whose stage runs start at each of starts (the
// first is its cohort instant), every run 30 minutes long and carrying
// metrics.
func (fx statsFixture) change(name, metrics string, starts ...time.Time) {
	fx.t.Helper()
	seedChange(fx.t, fx.st, fx.project, name)
	for i, at := range starts {
		if _, err := fx.pool.Exec(context.Background(), `
			INSERT INTO stage_runs (change_id, harness, command, stage, attempt, started_at, ended_at, metrics)
			SELECT c.id, 'claude-code', 'flow', 'flow.implement', $3, $4::timestamptz, $4::timestamptz + interval '30 minutes', $5::jsonb
			FROM changes c WHERE c.project_key = $1 AND c.name = $2
		`, fx.project, name, i+1, at, metrics); err != nil {
			fx.t.Fatalf("seed stage run %s: %v", name, err)
		}
	}
}

// dispatch records one dispatch lasting minutes with tokens in its metrics
// and returns its seq.
func (fx statsFixture) dispatch(change, role, key, outcome, cause string, at time.Time, minutes int, tokens string) int {
	fx.t.Helper()
	d := baseDispatch(role, "opus")
	end := at.Add(time.Duration(minutes) * time.Minute)
	d.Key, d.Outcome, d.StartedAt, d.EndedAt = key, outcome, at, &end
	if tokens != "" {
		d.Metrics = []byte(`{"tokens":` + tokens + `}`)
	}
	out, err := fx.st.RecordDispatch(context.Background(), fx.project, change, d)
	if err != nil {
		fx.t.Fatalf("RecordDispatch %s %s: %v", change, key, err)
	}
	// RecordDispatch is a begin, and a begin carries no cause; the end
	// that would is not this fixture's subject.
	if cause != "" {
		if _, err := fx.pool.Exec(context.Background(), `UPDATE dispatches SET cause = $2 WHERE id = $1`, out.ID, cause); err != nil {
			fx.t.Fatalf("set cause %s %s: %v", change, key, err)
		}
	}
	return out.Seq
}

func (fx statsFixture) finding(change, ref, severity string, round int, seq *int) {
	fx.t.Helper()
	f := baseFinding(ref, round)
	f.Severity, f.DispatchSeq = severity, seq
	if _, _, err := fx.st.UpsertFinding(context.Background(), fx.project, change, f); err != nil {
		fx.t.Fatalf("UpsertFinding %s %s: %v", change, ref, err)
	}
}

func (fx statsFixture) corrections(change string, n int) {
	fx.t.Helper()
	if _, err := fx.st.RecordTaskCount(context.Background(), fx.project, change, records.TaskCount{TotalTasks: 5, Corrections: n}); err != nil {
		fx.t.Fatalf("RecordTaskCount %s: %v", change, err)
	}
}

func (fx statsFixture) stats(id int64) store.ExperimentStatsResult {
	fx.t.Helper()
	got, err := fx.st.ExperimentStats(context.Background(), id)
	if err != nil {
		fx.t.Fatalf("ExperimentStats: %v", err)
	}
	return got
}

func (fx statsFixture) day(d float64) time.Time {
	return fx.T.Add(time.Duration(d * 24 * float64(time.Hour)))
}

func metricNames(r store.ExperimentStatsResult) []string {
	out := []string{}
	for _, m := range r.Metrics {
		out = append(out, m.Name)
	}
	return out
}

func metricOf(t *testing.T, r store.ExperimentStatsResult, name string) store.ExperimentMetric {
	t.Helper()
	for _, m := range r.Metrics {
		if m.Name == name {
			return m
		}
	}
	t.Fatalf("metric %q missing from %v", name, metricNames(r))
	return store.ExperimentMetric{}
}

// wantMeans asserts metric name's before and after means; a nil want is
// "no figure".
func wantMeans(t *testing.T, r store.ExperimentStatsResult, name string, before, after *float64) {
	t.Helper()
	m := metricOf(t, r, name)
	for _, c := range []struct {
		side      string
		got, want *float64
	}{{"before", m.Before, before}, {"after", m.After, after}} {
		switch {
		case c.want == nil && c.got != nil:
			t.Errorf("%s %s = %v, want none", name, c.side, *c.got)
		case c.want != nil && c.got == nil:
			t.Errorf("%s %s = none, want %v", name, c.side, *c.want)
		case c.want != nil && math.Abs(*c.got-*c.want) > 1e-9:
			t.Errorf("%s %s = %v, want %v", name, c.side, *c.got, *c.want)
		}
	}
}

func f64(v float64) *float64 { return &v }

// weekOf is the Monday 00:00 UTC opening at's ISO week.
func weekOf(at time.Time) time.Time {
	d := at.UTC().Truncate(24 * time.Hour)
	return d.AddDate(0, 0, -((int(d.Weekday()) + 6) % 7))
}

// TestExperimentStatsPlannerFamily pins the planner metrics, the equal
// before/after windows and that a change counts once, in the window its
// first stage run started in.
func TestExperimentStatsPlannerFamily(t *testing.T) {
	fx := newStatsFixture(t)
	id := fx.open(map[string][]string{"plan-designer": {"opus/Plan@xhigh"}})
	tok := func(n int) string {
		return fmt.Sprintf(`{"main":{"input":%d,"output":0},"sidechain":{"output":%d,"cache_creation":0,"cache_read":99999}}`, n/2, n/2)
	}

	// Before: b1 (rounds 1, two round-0 Crit/Imp, corrections 2, one
	// plan-designer dispatch of 2000 tokens over 30 min), b2 (nothing).
	fx.change("b1", `{}`, fx.day(-3))
	fx.finding("b1", "B1", "Critical", 0, nil)
	fx.finding("b1", "B2", "major", 0, nil)
	fx.finding("b1", "B3", "minor", 1, nil)
	fx.corrections("b1", 1)
	fx.corrections("b1", 2)
	fx.dispatch("b1", "planner", "design-1", "completed", "", fx.day(-3), 30, tok(2000))
	fx.change("b2", `{}`, fx.day(-2))
	// b2's only task-count row predates 0055: NULL corrections, skipped.
	if _, err := fx.pool.Exec(context.Background(), `
		INSERT INTO change_task_counts (change_id, total_tasks, corrections)
		SELECT id, 4, NULL FROM changes WHERE project_key = $1 AND name = 'b2'`, fx.project); err != nil {
		t.Fatalf("seed legacy task count: %v", err)
	}
	// After: a1 (corrections 3, a plan-designer dispatch of 4000 tokens
	// over 60 min); a planner dispatch of another role never counts.
	fx.change("a1", `{}`, fx.day(1))
	fx.finding("a1", "A1", "minor", 0, nil)
	fx.corrections("a1", 3)
	fx.dispatch("a1", "planner", "design-1", "completed", "", fx.day(1), 60, tok(4000))
	fx.dispatch("a1", "planner", "pipeline-fix-1-plan", "completed", "", fx.day(1), 500, tok(900000))
	// Outside both windows: old's first run predates the before window,
	// its later run inside the after window does not move it.
	fx.change("old", `{}`, fx.day(-10), fx.day(2))
	fx.corrections("old", 9)

	got := fx.stats(id)
	if want := []string{"reviewRounds", "round0CritImp", "planCorrections", "tokens", "wallClockMin"}; !reflect.DeepEqual(metricNames(got), want) {
		t.Fatalf("metrics = %v, want %v", metricNames(got), want)
	}
	if got.Runs != (store.ExperimentRuns{Before: 2, After: 1}) {
		t.Errorf("runs = %+v, want before 2 after 1", got.Runs)
	}
	if !got.BeforeFrom.Equal(fx.T.Add(-got.End.Sub(fx.T))) || got.End.Before(fx.now) {
		t.Errorf("windows = [%v, %v, %v), want equal lengths around %v", got.BeforeFrom, fx.T, got.End, fx.T)
	}
	wantMeans(t, got, "reviewRounds", f64(0.5), f64(0))
	wantMeans(t, got, "round0CritImp", f64(1), f64(0))
	wantMeans(t, got, "planCorrections", f64(2), f64(3))
	wantMeans(t, got, "tokens", f64(2000), f64(4000))
	wantMeans(t, got, "wallClockMin", f64(30), f64(60))

	// Weekly series: each week's mean over the changes whose first run
	// fell in it.
	perChange := map[time.Time][]float64{}
	for _, c := range []struct {
		at time.Time
		v  float64
	}{{fx.day(-3), 2}, {fx.day(1), 3}} {
		perChange[weekOf(c.at)] = append(perChange[weekOf(c.at)], c.v)
	}
	want := []store.ExperimentWeek{}
	for w, vs := range perChange {
		sum := 0.0
		for _, v := range vs {
			sum += v
		}
		want = append(want, store.ExperimentWeek{Week: w, Value: sum / float64(len(vs))})
	}
	sort.Slice(want, func(i, j int) bool { return want[i].Week.Before(want[j].Week) })
	series := metricOf(t, got, "planCorrections").Series
	if len(series) != len(want) {
		t.Fatalf("planCorrections series = %+v, want %+v", series, want)
	}
	for i := range want {
		if !series[i].Week.Equal(want[i].Week) || math.Abs(series[i].Value-want[i].Value) > 1e-9 {
			t.Errorf("planCorrections series[%d] = %+v, want %+v", i, series[i], want[i])
		}
	}
}

// TestExperimentStatsImplementerFamily pins verify reds (verify% rows
// ended blocked by a test failure) and findings per implementer task.
func TestExperimentStatsImplementerFamily(t *testing.T) {
	fx := newStatsFixture(t)
	id := fx.open(map[string][]string{"implementer": {"sonnet/flow-low"}})

	fx.change("c1", `{}`, fx.day(-1))
	fx.dispatch("c1", "implementer", "task-1-implementer", "completed", "", fx.day(-1), 10, "")
	fx.dispatch("c1", "implementer", "task-2-implementer", "completed", "", fx.day(-1), 10, "")
	for _, ref := range []string{"F1", "F2", "F3"} {
		fx.finding("c1", ref, "minor", 0, nil)
	}
	fx.dispatch("c1", "verifier", "verify", "blocked", "test-failure", fx.day(-1), 5, "")
	fx.dispatch("c1", "verifier", "verify-fix-1", "blocked", "environment", fx.day(-1), 5, "")
	fx.dispatch("c1", "verifier", "visual-verify", "blocked", "test-failure", fx.day(-1), 5, "")

	fx.change("c2", `{}`, fx.day(2))
	fx.dispatch("c2", "implementer", "task-1-implementer", "completed", "", fx.day(2), 10, "")
	fx.finding("c2", "G1", "minor", 0, nil)
	fx.dispatch("c2", "verifier", "verify", "blocked", "test-failure", fx.day(2), 5, "")
	fx.dispatch("c2", "verifier", "verify-fix-1", "blocked", "test-failure", fx.day(2), 5, "")

	got := fx.stats(id)
	if want := []string{"reviewRounds", "findingsPerTask", "verifyReds", "tokens", "wallClockMin"}; !reflect.DeepEqual(metricNames(got), want) {
		t.Fatalf("metrics = %v, want %v", metricNames(got), want)
	}
	wantMeans(t, got, "verifyReds", f64(1), f64(2))
	wantMeans(t, got, "findingsPerTask", f64(1.5), f64(1))
	wantMeans(t, got, "wallClockMin", f64(20), f64(10))
}

// TestExperimentStatsReviewerAndVisual pins findings per reviewer dispatch,
// late-round findings and the visual verifier's mismatches and re-runs.
func TestExperimentStatsReviewerAndVisual(t *testing.T) {
	fx := newStatsFixture(t)
	id := fx.open(map[string][]string{"panel-reviewer": {"opus/flow-low"}, "visual-verifier": {"opus/flow-low"}})

	fx.change("r1", `{}`, fx.day(-1))
	p := fx.dispatch("r1", "reviewer", "panel-0-primary", "completed", "", fx.day(-1), 10, "")
	fx.finding("r1", "P1", "major", 0, &p)
	fx.finding("r1", "P2", "minor", 0, &p)
	x := fx.dispatch("r1", "reviewer", "panel-0-test-sanity+call-efficiency", "completed", "", fx.day(-1), 10, "")
	fx.finding("r1", "X1", "minor", 0, &x)
	fx.finding("r1", "L1", "minor", 2, nil)
	v := fx.dispatch("r1", "verifier", "visual-verify", "completed", "", fx.day(-1), 10, "")
	fx.finding("r1", "V1", "minor", 0, &v)
	fx.dispatch("r1", "verifier", "visual-verify-fix-1", "completed", "", fx.day(-1), 10, "")
	fx.dispatch("r1", "verifier", "visual-verify-fix-2", "completed", "", fx.day(-1), 10, "")

	fx.change("r2", `{}`, fx.day(2))
	fx.dispatch("r2", "reviewer", "panel-0-primary+principles", "completed", "", fx.day(2), 10, "")
	q := fx.dispatch("r2", "reviewer", "panel-1-primary", "completed", "", fx.day(2), 10, "")
	fx.finding("r2", "Q1", "minor", 1, &q)

	got := fx.stats(id)
	wantMeans(t, got, "findingsPerDispatch", f64(2), f64(0))
	wantMeans(t, got, "lateRoundFindings", f64(1), f64(0))
	wantMeans(t, got, "vvMismatches", f64(1), f64(0))
	wantMeans(t, got, "vvReruns", f64(2), f64(0))
}

// TestExperimentStatsMainCountsMatchingRunsOnly pins that a main cell's
// after figures count only changes whose harvested main thread ran its
// model (alias or full id) at its effort; the rest are excluded.
func TestExperimentStatsMainCountsMatchingRunsOnly(t *testing.T) {
	fx := newStatsFixture(t)
	id := fx.open(map[string][]string{"main": {"opus@medium"}})
	run := func(model, effort string) string {
		return fmt.Sprintf(`{"effort":%q,"models":{%q:{"tokens":{"main":{"input":10,"output":90}}}},"tokens":{"main":{"input":10,"output":90}}}`, effort, model)
	}

	fx.change("m0", run("claude-opus-5-5", "high"), fx.day(-1)) // before: never filtered
	fx.change("m1", run("claude-opus-5-5", "medium"), fx.day(1))
	fx.change("m2", run("claude-opus-5-5", "high"), fx.day(1))
	fx.change("m3", run("opus", "medium"), fx.day(2))
	fx.change("m4", run("claude-sonnet-5", "medium"), fx.day(2))
	fx.change("m5", `{}`, fx.day(3)) // nothing harvested: no evidence it ran the pair
	fx.dispatch("m1", "verifier", "verify", "blocked", "test-failure", fx.day(1), 5, "")
	fx.dispatch("m2", "verifier", "verify", "blocked", "test-failure", fx.day(1), 5, "")

	got := fx.stats(id)
	if got.Runs != (store.ExperimentRuns{Before: 1, After: 2}) || got.Excluded != 3 {
		t.Fatalf("runs = %+v excluded %d, want before 1 after 2, excluded 3", got.Runs, got.Excluded)
	}
	wantMeans(t, got, "verifyReds", f64(0), f64(0.5))
	wantMeans(t, got, "tokens", f64(100), f64(100))
	wantMeans(t, got, "wallClockMin", f64(30), f64(30))
}

// TestExperimentStatsCombinedUnion pins that a two-role experiment
// reports the union of both families, each metric once.
func TestExperimentStatsCombinedUnion(t *testing.T) {
	fx := newStatsFixture(t)
	id := fx.open(map[string][]string{"plan-designer": {"opus/Plan@xhigh"}, "implementer": {"sonnet/flow-low"}})
	got := fx.stats(id)
	names := metricNames(got)
	sorted := append([]string(nil), names...)
	sort.Strings(sorted)
	want := []string{"findingsPerTask", "planCorrections", "reviewRounds", "round0CritImp", "tokens", "verifyReds", "wallClockMin"}
	if !reflect.DeepEqual(sorted, want) {
		t.Fatalf("metrics = %v, want each of %v once", names, want)
	}
	for _, m := range got.Metrics {
		if m.Before != nil || m.After != nil || len(m.Series) != 0 {
			t.Errorf("metric %s with no runs = %+v, want no figures", m.Name, m)
		}
	}
	if _, err := fx.st.ExperimentStats(context.Background(), id+1000); !errors.Is(err, store.ErrExperimentNotFound) {
		t.Errorf("ExperimentStats(unknown) = %v, want ErrExperimentNotFound", err)
	}
}
