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

// runStage begins, merges metrics into, and ends a stage run in one call,
// for tests that only care about the resulting row's shape.
func runStage(t *testing.T, st *store.Store, in store.BeginStageInput, metrics json.RawMessage, endedAt time.Time, outcome string) store.StageRun {
	t.Helper()
	ctx := context.Background()

	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if metrics != nil {
		if err := st.MergeMetrics(ctx, run.ID, metrics); err != nil {
			t.Fatalf("MergeMetrics: %v", err)
		}
	}
	if err := st.EndStage(ctx, run.ID, endedAt, outcome); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	return got
}

func TestAggregateRestrictsByPeriodInSQL(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-period-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	inside := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	inside.StartedAt = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	runStage(t, st, inside, json.RawMessage(`{"cost_usd":1.5}`), inside.StartedAt.Add(time.Minute), "completed")

	before := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	before.StartedAt = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	runStage(t, st, before, json.RawMessage(`{"cost_usd":9}`), before.StartedAt.Add(time.Minute), "completed")

	after := baseBeginInput(projectKey, "kan-1", "/flow", "finish")
	after.StartedAt = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	runStage(t, st, after, json.RawMessage(`{"cost_usd":9}`), after.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.CostPerChange(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("CostPerChange returned %d groups, want 1 (only the stage started inside the period)", len(rows))
	}
	if rows[0].Stage != "SDD + TDD per task" {
		t.Errorf("CostPerChange returned stage %q, want the one whose run started inside the period", rows[0].Stage)
	}
}

func TestAggregateEmptyPeriodReturnsEmptyNotError(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-empty-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	run.StartedAt = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	runStage(t, st, run, json.RawMessage(`{"cost_usd":1}`), run.StartedAt.Add(time.Minute), "completed")

	emptyPeriod := store.Period{
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	rows, err := st.CostPerChange(ctx, emptyPeriod, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange(empty period): %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("CostPerChange(empty period) returned %d rows, want 0", len(rows))
	}

	board, err := st.LiveStateBoard(ctx, emptyPeriod, &projectKey)
	if err != nil {
		t.Fatalf("LiveStateBoard(empty period): %v", err)
	}
	if len(board) != 0 {
		t.Errorf("LiveStateBoard(empty period) returned %d rows, want 0", len(board))
	}

	leaderboard, err := st.StageLeaderboard(ctx, emptyPeriod, &projectKey, nil)
	if err != nil {
		t.Fatalf("StageLeaderboard(empty period): %v", err)
	}
	if len(leaderboard) != 0 {
		t.Errorf("StageLeaderboard(empty period) returned %d rows, want 0", len(leaderboard))
	}
}

// TestAggregateExcludesUnavailableTokensFromAverages seeds three stage runs
// in the same command/stage group: two carry a measured tokens.input, one
// carries no tokens key at all (a harness with no transcript). A buggy
// aggregation that coalesces missing tokens to zero before averaging would
// report a mean pulled down toward zero by the unmeasured run; the correct
// mean is computed over the two measured runs only, and the unmeasured run
// is counted separately rather than folded into the average.
func TestAggregateExcludesUnavailableTokensFromAverages(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-unavail-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	base := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	base.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	measuredA := base
	measuredA.SessionID = ptr("s-a")
	runStage(t, st, measuredA, json.RawMessage(`{"tokens":{"main":{"input":100}}}`), base.StartedAt.Add(time.Minute), "completed")

	measuredB := base
	measuredB.SessionID = ptr("s-b")
	runStage(t, st, measuredB, json.RawMessage(`{"tokens":{"main":{"input":100}}}`), base.StartedAt.Add(time.Minute), "completed")

	// A Cursor/Codex-style run: no transcript, so no tokens key at all.
	unmeasured := base
	unmeasured.SessionID = ptr("s-c")
	unmeasured.Harness = "cursor"
	runStage(t, st, unmeasured, json.RawMessage(`{"tokens_available":false}`), base.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.CostPerChange(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("CostPerChange returned %d groups, want 1", len(rows))
	}
	row := rows[0]

	if row.RunCount != 3 {
		t.Errorf("RunCount = %d, want 3 (every stage run counts)", row.RunCount)
	}
	if row.MeasuredRuns != 2 {
		t.Errorf("MeasuredRuns = %d, want 2 (only the runs carrying a tokens key)", row.MeasuredRuns)
	}
	if row.MeanTokensInput == nil {
		t.Fatalf("MeanTokensInput is nil, want a value computed over the 2 measured runs")
	}
	if *row.MeanTokensInput != 100 {
		t.Errorf("MeanTokensInput = %v, want 100 (average of the two measured runs, not pulled toward 0 by the unmeasured one)", *row.MeanTokensInput)
	}
	if row.TotalTokensInput == nil || *row.TotalTokensInput != 200 {
		t.Errorf("TotalTokensInput = %v, want 200", row.TotalTokensInput)
	}
}

// TestAggregateSeparatesMainFromSidechain seeds one run whose metrics carry
// both tokens.main and tokens.sidechain, and asserts CostPerChange totals
// them into separate fields rather than folding subagent cost into its
// parent's.
func TestAggregateSeparatesMainFromSidechain(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-mainside-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, in, json.RawMessage(`{"tokens":{"main":{"input":1000},"sidechain":{"input":4000}}}`), in.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.CostPerChange(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("CostPerChange returned %d groups, want 1", len(rows))
	}
	row := rows[0]

	if row.MainTokens == nil || *row.MainTokens != 1000 {
		t.Errorf("MainTokens = %v, want 1000", row.MainTokens)
	}
	if row.SidechainTokens == nil || *row.SidechainTokens != 4000 {
		t.Errorf("SidechainTokens = %v, want 4000", row.SidechainTokens)
	}
}

// costRun begins and ends a stage run with cost_usd set directly in its
// metrics, bypassing Price -- these tests are checking the aggregation
// arithmetic over a known cost, not Price's derivation of one.
func costRun(t *testing.T, st *store.Store, in store.BeginStageInput, costUSD float64) store.StageRun {
	t.Helper()
	ctx := context.Background()

	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	patch, err := json.Marshal(map[string]any{"cost_usd": costUSD})
	if err != nil {
		t.Fatalf("marshal cost_usd patch: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, patch); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, in.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	return got
}

// TestStageLeaderboardComputesMeanMedianP90 seeds five stage runs of the
// same command/stage with cost_usd 10, 20, 30, 40 and 50, and checks the
// leaderboard's mean, median and p90 against values hand-computed from
// that exact set (percentile_cont with linear interpolation): mean 30,
// median 30 (the middle value), p90 46 (interpolated 60% of the way from
// 40 to 50, since rank (5-1)*0.9 = 3.6 lands between the 4th and 5th
// values in sorted order).
func TestStageLeaderboardComputesMeanMedianP90(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-leaderboard-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	costs := []float64{10, 20, 30, 40, 50}
	for i, cost := range costs {
		in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
		in.SessionID = ptr(fmt.Sprintf("session-%d", i))
		in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Hour)
		costRun(t, st, in, cost)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.StageLeaderboard(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("StageLeaderboard: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("StageLeaderboard returned %d rows, want 1", len(rows))
	}
	row := rows[0]

	const eps = 1e-9
	if row.RunCount != 5 {
		t.Errorf("RunCount = %d, want 5", row.RunCount)
	}
	if diff := row.MeanCostUSD - 30; diff > eps || diff < -eps {
		t.Errorf("MeanCostUSD = %v, want 30", row.MeanCostUSD)
	}
	if diff := row.MedianCostUSD - 30; diff > eps || diff < -eps {
		t.Errorf("MedianCostUSD = %v, want 30", row.MedianCostUSD)
	}
	if diff := row.P90CostUSD - 46; diff > eps || diff < -eps {
		t.Errorf("P90CostUSD = %v, want 46", row.P90CostUSD)
	}
}

// TestTrendOverTimeBucketsByDay seeds three runs on one day and two on the
// next, and checks that TrendOverTime's per-day bucket, count and total
// cost match those hand-computed groupings exactly.
func TestTrendOverTimeBucketsByDay(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-trend-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	day1 := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)

	day1Costs := []float64{10, 20, 30} // total 60
	for i, cost := range day1Costs {
		in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
		in.SessionID = ptr(fmt.Sprintf("day1-%d", i))
		in.StartedAt = day1.Add(time.Duration(i) * time.Hour)
		costRun(t, st, in, cost)
	}
	day2Costs := []float64{40, 50} // total 90
	for i, cost := range day2Costs {
		in := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
		in.SessionID = ptr(fmt.Sprintf("day2-%d", i))
		in.StartedAt = day2.Add(time.Duration(i) * time.Hour)
		costRun(t, st, in, cost)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	points, err := st.TrendOverTime(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("TrendOverTime: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("TrendOverTime returned %d points, want 2", len(points))
	}

	const eps = 1e-9
	if !points[0].Day.Equal(day1) {
		t.Errorf("points[0].Day = %v, want %v", points[0].Day, day1)
	}
	if points[0].RunCount != 3 {
		t.Errorf("points[0].RunCount = %d, want 3", points[0].RunCount)
	}
	if points[0].TotalCostUSD == nil || *points[0].TotalCostUSD < 60-eps || *points[0].TotalCostUSD > 60+eps {
		t.Errorf("points[0].TotalCostUSD = %v, want 60", points[0].TotalCostUSD)
	}
	if !points[1].Day.Equal(day2) {
		t.Errorf("points[1].Day = %v, want %v", points[1].Day, day2)
	}
	if points[1].RunCount != 2 {
		t.Errorf("points[1].RunCount = %d, want 2", points[1].RunCount)
	}
	if points[1].TotalCostUSD == nil || *points[1].TotalCostUSD < 90-eps || *points[1].TotalCostUSD > 90+eps {
		t.Errorf("points[1].TotalCostUSD = %v, want 90", points[1].TotalCostUSD)
	}
}

// TestPanelEconomicsComputesTokensAndFindingsPerMTok seeds one stage run
// with a known token total and a known findings-by-severity bag, and
// checks TokensTotal and FindingsPerMTok against values hand-computed from
// them, not just FindingsTotal's row count as before.
// TestAggregateOtherViewsRunWithoutError covers the two views not exercised
// with hand-computed arithmetic elsewhere: LiveStateBoard (a projection of
// changes, already covered thoroughly by changes_test.go) and
// CacheEfficiency's ratio, computed here from a known cache-read and
// cache-creation total.
func TestAggregateOtherViewsRunWithoutError(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-smoke-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"tokens":{"main":{"cache_read":900000,"cache_creation":100000}}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, in.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	if _, err := st.LiveStateBoard(ctx, period, nil); err != nil {
		t.Errorf("LiveStateBoard: %v", err)
	}

	rows, err := st.CacheEfficiency(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CacheEfficiency: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("CacheEfficiency returned %d rows, want 1", len(rows))
	}
	if rows[0].CacheReadTotal == nil || *rows[0].CacheReadTotal != 900000 {
		t.Errorf("CacheReadTotal = %v, want 900000", rows[0].CacheReadTotal)
	}
	if rows[0].CacheCreationTotal == nil || *rows[0].CacheCreationTotal != 100000 {
		t.Errorf("CacheCreationTotal = %v, want 100000", rows[0].CacheCreationTotal)
	}
	const eps = 1e-9
	if rows[0].Ratio == nil {
		t.Fatalf("Ratio is nil, want 9")
	}
	if diff := *rows[0].Ratio - 9; diff > eps || diff < -eps {
		t.Errorf("Ratio = %v, want 9 (900000/100000)", *rows[0].Ratio)
	}
}

// TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero is F6's
// own test: a stage run that recorded real cache-read usage but exactly
// zero cache-creation -- a plausible, real shape, not a malformed one:
// every message hit an already-warm cache and created nothing new -- must
// leave Ratio nil rather than computing cache_read / 0, which in float64
// division is +Inf, not a panic, so nothing before this test would have
// caught a regression that dropped the `!= 0` guard: the query would keep
// succeeding and returning a row, just with an unusable ratio silently
// smuggled through as a number.
func TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-cache-ratio-zero-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"tokens":{"main":{"cache_read":500000,"cache_creation":0}}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, in.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	rows, err := st.CacheEfficiency(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CacheEfficiency: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("CacheEfficiency returned %d rows, want 1", len(rows))
	}
	if rows[0].CacheReadTotal == nil || *rows[0].CacheReadTotal != 500000 {
		t.Errorf("CacheReadTotal = %v, want 500000", rows[0].CacheReadTotal)
	}
	if rows[0].CacheCreationTotal == nil || *rows[0].CacheCreationTotal != 0 {
		t.Errorf("CacheCreationTotal = %v, want 0", rows[0].CacheCreationTotal)
	}
	if rows[0].Ratio != nil {
		t.Errorf("Ratio = %v, want nil: cache-read over zero cache-creation must never become a (non-nil) NaN or Inf ratio", *rows[0].Ratio)
	}
}

// TestCostPerChangeModelFilterReportsThatModelsOwnNumbers is task 21, step
// 3b's own scenario: a two-model review-panel run costing $61.10 across
// Opus ($41.20) and Sonnet ($19.90), filtered to Sonnet, must report
// exactly Sonnet's own $19.90 -- not the run's $61.10 (attributing the
// Opus parent to Sonnet) and not $0 (reading as "Sonnet was never used").
// Both wrong answers are asserted against directly, not merely the right
// one, since either would otherwise pass a looser assertion.
func TestCostPerChangeModelFilterReportsThatModelsOwnNumbers(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-modelfilter-cost-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, in, json.RawMessage(`{
		"tokens": {"main": {"input": 1000}, "sidechain": {"input": 4000}},
		"cost_usd": 61.10,
		"models": {
			"claude-opus-5":   {"tokens": {"main": {"input": 1000}}, "cost_usd": 41.20},
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 4000}}, "cost_usd": 19.90}
		}
	}`), in.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	const eps = 1e-9

	unfiltered, err := st.CostPerChange(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange(unfiltered): %v", err)
	}
	if len(unfiltered) != 1 || unfiltered[0].TotalCostUSD == nil {
		t.Fatalf("CostPerChange(unfiltered) = %+v, want one row with a total cost", unfiltered)
	}
	if diff := *unfiltered[0].TotalCostUSD - 61.10; diff > eps || diff < -eps {
		t.Errorf("unfiltered TotalCostUSD = %v, want 61.10 (the whole run)", *unfiltered[0].TotalCostUSD)
	}

	filtered, err := st.CostPerChange(ctx, period, &projectKey, ptr("claude-sonnet-5"))
	if err != nil {
		t.Fatalf("CostPerChange(model=claude-sonnet-5): %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("CostPerChange(model=claude-sonnet-5) returned %d rows, want 1", len(filtered))
	}
	row := filtered[0]
	if row.TotalCostUSD == nil {
		t.Fatalf("filtered TotalCostUSD is nil, want 19.90")
	}
	if diff := *row.TotalCostUSD - 61.10; diff > -eps && diff < eps {
		t.Fatalf("filtered TotalCostUSD = 61.10, the whole run's cost -- the Opus parent's cost was attributed to Sonnet")
	}
	if *row.TotalCostUSD == 0 {
		t.Fatalf("filtered TotalCostUSD = 0 -- reads as \"Sonnet was never used\", when it was")
	}
	if diff := *row.TotalCostUSD - 19.90; diff > eps || diff < -eps {
		t.Errorf("filtered TotalCostUSD = %v, want 19.90 (Sonnet's own bucket)", *row.TotalCostUSD)
	}
	if row.SidechainTokens == nil || *row.SidechainTokens != 4000 {
		t.Errorf("filtered SidechainTokens = %v, want 4000 (Sonnet's own bucket, not the run's main+sidechain total)", row.SidechainTokens)
	}
	if row.MainTokens != nil {
		t.Errorf("filtered MainTokens = %v, want nil (Sonnet's own bucket recorded no main tokens)", row.MainTokens)
	}

	// A filter matching no run in scope returns empty, never an error.
	none, err := st.CostPerChange(ctx, period, &projectKey, ptr("claude-haiku-5"))
	if err != nil {
		t.Fatalf("CostPerChange(model=claude-haiku-5): %v", err)
	}
	if len(none) != 0 {
		t.Errorf("CostPerChange(model=claude-haiku-5) returned %d rows, want 0", len(none))
	}
}

// TestModelFilterHonouredByEveryAggregation seeds one stage run recording
// exactly one model, and iterates every one of the four model-filterable
// aggregations -- rather than spot-checking one or two -- asserting each
// one returns at least one row for the model the run actually used and
// zero rows for a model it did not, so a method added to this file later
// without wiring the filter through fails here rather than in a view
// nobody queried.
func TestModelFilterHonouredByEveryAggregation(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-modelfilter-every-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1") // baseChange sets ReviewPanelRoster "light"

	in := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, in, json.RawMessage(`{
		"tokens": {"main": {"input": 100}},
		"cost_usd": 10,
		"findings_by_severity": {"high": 1},
		"models": {"claude-opus-5": {"tokens": {"main": {"input": 100}}, "cost_usd": 10}}
	}`), in.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	matching := ptr("claude-opus-5")
	nonMatching := ptr("claude-sonnet-5")

	checks := []struct {
		name string
		rows func(model *string) (int, error)
	}{
		{"CostPerChange", func(m *string) (int, error) {
			r, err := st.CostPerChange(ctx, period, &projectKey, m)
			return len(r), err
		}},
		{"StageLeaderboard", func(m *string) (int, error) {
			r, err := st.StageLeaderboard(ctx, period, &projectKey, m)
			return len(r), err
		}},
		{"TrendOverTime", func(m *string) (int, error) {
			r, err := st.TrendOverTime(ctx, period, &projectKey, m)
			return len(r), err
		}},
		{"CacheEfficiency", func(m *string) (int, error) {
			r, err := st.CacheEfficiency(ctx, period, &projectKey, m)
			return len(r), err
		}},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			n, err := c.rows(matching)
			if err != nil {
				t.Fatalf("%s(model=%s): %v", c.name, *matching, err)
			}
			if n == 0 {
				t.Errorf("%s(model=%s) returned 0 rows, want at least 1 (the run recorded this model)", c.name, *matching)
			}

			n, err = c.rows(nonMatching)
			if err != nil {
				t.Fatalf("%s(model=%s): %v", c.name, *nonMatching, err)
			}
			if n != 0 {
				t.Errorf("%s(model=%s) returned %d rows, want 0 (the run never recorded this model)", c.name, *nonMatching, n)
			}
		})
	}
}

// TestCountRunsWithoutModel seeds one stage run recording a model and two
// that record none (one with an explicit tokens_available:false, matching
// a non-Claude harness; one with no metrics patch at all), and asserts the
// count covers exactly the two runs with no recorded model -- unaffected
// by which model a caller is about to filter for, since a run with no
// model could not have matched any filter.
func TestCountRunsWithoutModel(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nomodel-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	withModel := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	withModel.SessionID = ptr("s-with")
	withModel.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, withModel, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":5}}}`), withModel.StartedAt.Add(time.Minute), "completed")

	noModelA := baseBeginInput(projectKey, "kan-1", "/flow", "finish")
	noModelA.SessionID = ptr("s-none-a")
	noModelA.Harness = "cursor"
	noModelA.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, noModelA, json.RawMessage(`{"tokens_available":false}`), noModelA.StartedAt.Add(time.Minute), "completed")

	noModelB := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	noModelB.SessionID = ptr("s-none-b")
	noModelB.StartedAt = time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)
	runStage(t, st, noModelB, nil, noModelB.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	count, err := st.CountRunsWithoutModel(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("CountRunsWithoutModel: %v", err)
	}
	if count != 2 {
		t.Errorf("CountRunsWithoutModel = %d, want 2 (the two runs recording no model)", count)
	}
}

// TestCountRunsWithoutModelCoversAnExplicitEmptyModelsObject is F5's own
// test: a run whose metrics bag carries a "models" key that is present
// but an empty object (`{}`) -- distinct from the key being absent
// entirely, and distinct from it being JSON null, both already covered by
// TestCountRunsWithoutModel -- must still count as a run with no recorded
// model. The SQL's third disjunct (`sr.metrics->'models' = '{}'::jsonb`)
// exists for exactly this shape; nothing before this test seeded it, so a
// regression that dropped that disjunct passed the whole suite.
func TestCountRunsWithoutModelCoversAnExplicitEmptyModelsObject(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nomodel-empty-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	emptyModels := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	emptyModels.SessionID = ptr("s-empty-models")
	emptyModels.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, emptyModels, json.RawMessage(`{"models":{}}`), emptyModels.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	count, err := st.CountRunsWithoutModel(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("CountRunsWithoutModel: %v", err)
	}
	if count != 1 {
		t.Errorf("CountRunsWithoutModel = %d, want 1 (the run whose \"models\" key is present but empty)", count)
	}
}

// TestCountRunsWithoutModelIsGenuinelyZero asserts the count reports 0,
// not an absent or error result, when every run in scope did record a
// model -- the genuinely-zero case the plan calls out by name, distinct
// from "no filter was applied" (which this method is never called for at
// all -- the API layer, not this method, encodes that distinction).
func TestCountRunsWithoutModelIsGenuinelyZero(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nomodel-zero-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, in, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":5}}}`), in.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	count, err := st.CountRunsWithoutModel(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("CountRunsWithoutModel: %v", err)
	}
	if count != 0 {
		t.Errorf("CountRunsWithoutModel = %d, want 0", count)
	}
}

// TestListModels seeds two in-period runs recording three distinct models
// between them, one run outside the period recording a fourth, and one
// in-period run recording none, and asserts ListModels returns exactly the
// distinct models actually used inside the period, sorted -- never the
// out-of-period model, and never a fabricated entry for the run with none.
func TestListModels(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-listmodels-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	a := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	a.SessionID = ptr("s-a")
	a.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, a, json.RawMessage(`{"models":{"claude-sonnet-5":{"cost_usd":1}}}`), a.StartedAt.Add(time.Minute), "completed")

	b := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	b.SessionID = ptr("s-b")
	b.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, b, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":2},"claude-sonnet-5":{"cost_usd":1}}}`), b.StartedAt.Add(time.Minute), "completed")

	outsidePeriod := baseBeginInput(projectKey, "kan-1", "/flow", "finish")
	outsidePeriod.SessionID = ptr("s-outside")
	outsidePeriod.StartedAt = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	runStage(t, st, outsidePeriod, json.RawMessage(`{"models":{"claude-haiku-5":{"cost_usd":1}}}`), outsidePeriod.StartedAt.Add(time.Minute), "completed")

	noModel := baseBeginInput(projectKey, "kan-1", "/flow", "finish")
	noModel.SessionID = ptr("s-nomodel")
	noModel.StartedAt = time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC)
	runStage(t, st, noModel, nil, noModel.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	models, err := st.ListModels(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	want := []string{"claude-opus-5", "claude-sonnet-5"}
	if len(models) != len(want) {
		t.Fatalf("ListModels = %v, want %v", models, want)
	}
	for i, m := range want {
		if models[i] != m {
			t.Errorf("ListModels[%d] = %q, want %q", i, models[i], m)
		}
	}
}

// --- TestAllRecordedRunsUnmeasured* --------------------------------------
//
// Task 5's third arm: three separate cases pinning that "no runs", "runs
// recorded but none measured" and "a run measured as a real zero" cannot
// collapse into one another (tasks.md's own "must not collapse" rule).

// TestAllRecordedRunsUnmeasuredWhenNoneCarryTokens seeds two runs in
// period, neither carrying a "tokens" key -- one with no metrics merged
// at all, one that recorded a model but no tokens -- and asserts the
// period reads as entirely unmeasured, not as empty.
func TestAllRecordedRunsUnmeasuredWhenNoneCarryTokens(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-allunmeasured-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	noMetrics := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	noMetrics.SessionID = ptr("s-no-metrics")
	noMetrics.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, noMetrics, nil, noMetrics.StartedAt.Add(time.Minute), "completed")

	modelNoTokens := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	modelNoTokens.SessionID = ptr("s-model-no-tokens")
	modelNoTokens.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, modelNoTokens, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":5}}}`), modelNoTokens.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	unmeasured, err := st.AllRecordedRunsUnmeasured(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("AllRecordedRunsUnmeasured: %v", err)
	}
	if !unmeasured {
		t.Errorf("AllRecordedRunsUnmeasured = false, want true: neither seeded run carries a \"tokens\" key")
	}
}

// TestAllRecordedRunsUnmeasuredFalseWhenARunIsMeasuredAsZero seeds one
// unmeasured run alongside one whose tokens bag is present but totals a
// real zero, and asserts the period does NOT read as all-unmeasured: a
// single measured run, even one measured at zero, is enough to disprove
// "none was measured" for the whole period.
func TestAllRecordedRunsUnmeasuredFalseWhenARunIsMeasuredAsZero(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-allunmeasured-zero-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	unmeasured := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
	unmeasured.SessionID = ptr("s-unmeasured")
	unmeasured.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, unmeasured, nil, unmeasured.StartedAt.Add(time.Minute), "completed")

	measuredZero := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	measuredZero.SessionID = ptr("s-measured-zero")
	measuredZero.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, measuredZero, json.RawMessage(`{"tokens":{"main":{"input":0}}}`), measuredZero.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	got, err := st.AllRecordedRunsUnmeasured(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("AllRecordedRunsUnmeasured: %v", err)
	}
	if got {
		t.Errorf("AllRecordedRunsUnmeasured = true, want false: one run carries a \"tokens\" key, even measured at zero")
	}
}

// TestAllRecordedRunsUnmeasuredFalseWhenNoRunsInPeriod asserts a period
// with no stage runs at all reads as false here -- "no runs" and "runs
// recorded but none measured" are the two different arms this method must
// never conflate; the caller (internal/api) uses the existing recorded
// signal, not this method, to report the "no runs" arm.
func TestAllRecordedRunsUnmeasuredFalseWhenNoRunsInPeriod(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-allunmeasured-empty-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	got, err := st.AllRecordedRunsUnmeasured(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("AllRecordedRunsUnmeasured: %v", err)
	}
	if got {
		t.Errorf("AllRecordedRunsUnmeasured = true, want false: no runs exist in this period at all")
	}
}

// TestReviewersCountsBySeverityAndMarksExperimental pins design.md's
// "reviewers" view: one row per findings.slot joined to dispatches on
// (change_id, slot) within the period, with an experimental slot (an
// "exp-" prefix) badged and carrying the newest decision row's roster
// description for that slot -- task 16's own worked shape.
//
// "primary" gets two dispatches and three findings (one Critical, one
// Important, one deferred Minor); "exp-failure-modes" gets one dispatch and
// two findings (one withdrawn Critical, one fixed Minor). Every count and
// share below is hand-computed from exactly that seed.
func TestReviewersCountsBySeverityAndMarksExperimental(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-reviewers-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	stageRunID := run.ID

	recordDispatch := func(slot string) records.Dispatch {
		t.Helper()
		d := baseDispatch("reviewer", "sonnet")
		d.Slot = slot
		d.StageRunID = &stageRunID
		out, err := st.RecordDispatch(ctx, projectKey, "kan-1", d)
		if err != nil {
			t.Fatalf("RecordDispatch %s: %v", slot, err)
		}
		return out
	}
	primary1 := recordDispatch("primary")
	recordDispatch("primary")
	exp1 := recordDispatch("exp-failure-modes")

	recordFinding := func(ref, slot, severity, status string, dispatchSeq int) {
		t.Helper()
		f := baseFinding(ref, 1)
		f.Slot = slot
		f.Severity = severity
		f.Status = status
		f.DispatchSeq = &dispatchSeq
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
			t.Fatalf("UpsertFinding %s: %v", ref, err)
		}
	}
	recordFinding("F1", "primary", "Critical", "fixed", primary1.Seq)
	recordFinding("F2", "primary", "Important", "fixed", primary1.Seq)
	recordFinding("F3", "primary", "Minor", "deferred cosmetic, out of scope", primary1.Seq)
	recordFinding("F4", "exp-failure-modes", "Critical", "withdrawn not applicable here", exp1.Seq)
	recordFinding("F5", "exp-failure-modes", "Minor", "fixed", exp1.Seq)

	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-1", records.Decision{
		SessionToken: "mf-reviewers-test",
		Decision: json.RawMessage(`{
			"panel": {
				"roster": [
					{"slot": "primary", "experimental": false},
					{"slot": "exp-failure-modes", "experimental": true,
					 "prompt": "skills/flow/experimental/failure-modes.md",
					 "description": "What the diff does under error, timeout and partial write"}
				]
			}
		}`),
	}); err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.Reviewers(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("Reviewers: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("Reviewers returned %d rows, want 2", len(rows))
	}

	bySlot := map[string]store.ReviewerRow{}
	for _, r := range rows {
		bySlot[r.Slot] = r
	}

	const eps = 1e-9

	primary, ok := bySlot["primary"]
	if !ok {
		t.Fatalf("no row for slot %q", "primary")
	}
	if primary.Experimental {
		t.Errorf("primary Experimental = true, want false")
	}
	if primary.Description != "" {
		t.Errorf("primary Description = %q, want empty (its roster entry carries none)", primary.Description)
	}
	if primary.Dispatches != 2 {
		t.Errorf("primary Dispatches = %d, want 2", primary.Dispatches)
	}
	if primary.Changes != 1 {
		t.Errorf("primary Changes = %d, want 1", primary.Changes)
	}
	if primary.Critical != 1 || primary.Important != 1 || primary.Minor != 1 {
		t.Errorf("primary severity counts = %+v, want Critical=1 Important=1 Minor=1", primary)
	}
	if diff := primary.FindingsPerDispatch - 1.5; diff > eps || diff < -eps {
		t.Errorf("primary FindingsPerDispatch = %v, want 1.5 (3 findings / 2 dispatches)", primary.FindingsPerDispatch)
	}
	if diff := primary.DeferredShare - (1.0 / 3.0); diff > eps || diff < -eps {
		t.Errorf("primary DeferredShare = %v, want 1/3", primary.DeferredShare)
	}
	if primary.WithdrawnShare != 0 {
		t.Errorf("primary WithdrawnShare = %v, want 0", primary.WithdrawnShare)
	}

	exp, ok := bySlot["exp-failure-modes"]
	if !ok {
		t.Fatalf("no row for slot %q", "exp-failure-modes")
	}
	if !exp.Experimental {
		t.Errorf("exp-failure-modes Experimental = false, want true")
	}
	const wantDescription = "What the diff does under error, timeout and partial write"
	if exp.Description != wantDescription {
		t.Errorf("exp-failure-modes Description = %q, want %q", exp.Description, wantDescription)
	}
	if exp.Dispatches != 1 {
		t.Errorf("exp-failure-modes Dispatches = %d, want 1", exp.Dispatches)
	}
	if exp.Changes != 1 {
		t.Errorf("exp-failure-modes Changes = %d, want 1", exp.Changes)
	}
	if exp.Critical != 1 || exp.Important != 0 || exp.Minor != 1 {
		t.Errorf("exp-failure-modes severity counts = %+v, want Critical=1 Important=0 Minor=1", exp)
	}
	if diff := exp.FindingsPerDispatch - 2.0; diff > eps || diff < -eps {
		t.Errorf("exp-failure-modes FindingsPerDispatch = %v, want 2.0", exp.FindingsPerDispatch)
	}
	if exp.DeferredShare != 0 {
		t.Errorf("exp-failure-modes DeferredShare = %v, want 0", exp.DeferredShare)
	}
	if diff := exp.WithdrawnShare - 0.5; diff > eps || diff < -eps {
		t.Errorf("exp-failure-modes WithdrawnShare = %v, want 0.5", exp.WithdrawnShare)
	}
}

// TestReviewersSplitsBundledSlots pins design.md's
// "compound-slot-one-row-per-bundle": a bundle is one dispatches row whose
// slot is its roles "+"-joined, and Reviewers must split it into one row
// per role rather than one row for the compound string, joining findings
// by dispatch_id (not by the raw slot column, which never equals a single
// role's name for a bundled dispatch).
//
// One dispatch, slot "primary+principles", carries one "primary" finding
// and none for "principles" -- expect two rows: "principles" with
// Dispatches=1 and zero findings, "primary" with Dispatches=1 and one
// finding.
func TestReviewersSplitsBundledSlots(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-reviewers-bundle-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/flow", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	stageRunID := run.ID

	d := baseDispatch("reviewer", "sonnet")
	d.Slot = "primary+principles"
	d.StageRunID = &stageRunID
	bundle, err := st.RecordDispatch(ctx, projectKey, "kan-1", d)
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}

	f := baseFinding("F1", 1)
	f.Slot = "primary"
	f.Severity = "Important"
	f.Status = "fixed"
	f.DispatchSeq = &bundle.Seq
	if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.Reviewers(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("Reviewers: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("Reviewers returned %d rows, want 2 (one per role of the bundled slot)", len(rows))
	}

	bySlot := map[string]store.ReviewerRow{}
	for _, r := range rows {
		bySlot[r.Slot] = r
	}

	principles, ok := bySlot["principles"]
	if !ok {
		t.Fatalf("no row for slot %q -- the bundled dispatch's second role was not split out", "principles")
	}
	if principles.Dispatches != 1 {
		t.Errorf("principles Dispatches = %d, want 1", principles.Dispatches)
	}
	if principles.Critical != 0 || principles.Important != 0 || principles.Minor != 0 {
		t.Errorf("principles severity counts = %+v, want all zero (no finding recorded under this role)", principles)
	}

	primary, ok := bySlot["primary"]
	if !ok {
		t.Fatalf("no row for slot %q", "primary")
	}
	if primary.Dispatches != 1 {
		t.Errorf("primary Dispatches = %d, want 1", primary.Dispatches)
	}
	if primary.Important != 1 {
		t.Errorf("primary Important = %d, want 1", primary.Important)
	}
}

// TestDecisionsJoinsRunTotals seeds one recorded decision together with the
// stage run, dispatches and findings its own session token shares, and
// asserts Decisions (design.md's "Stats views › decisions") joins every one
// of them into a single row: the decision's own fields (class, override,
// execution, implementer, panel), the run's wall-clock (stage_runs), its
// tokens and cost (dispatches.metrics, the same "tokens" keys CostPerChange
// reads), its fallback/timed-out dispatch counts (dispatches.outcome), and
// its findings by severity and fix round (findings, joined through
// dispatch_id).
func TestDecisionsJoinsRunTotals(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-decisions-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const token = "mf-decisions-test"

	runIn := baseBeginInput(projectKey, "kan-1", "/flow", "flow.implement")
	runIn.SessionToken = ptr(token)
	runIn.StartedAt = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, runIn)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, runIn.StartedAt.Add(10*time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	implD := baseDispatch("implementer", "sonnet")
	implD.SessionToken = token
	implD.Outcome = "completed"
	implD.Metrics = json.RawMessage(`{"tokens":{"main":{"input":100,"output":50,"cache_read":20}},"cost_usd":1.25}`)
	implOut, err := st.RecordDispatch(ctx, projectKey, "kan-1", implD)
	if err != nil {
		t.Fatalf("RecordDispatch implementer: %v", err)
	}

	fallbackD := baseDispatch("panel", "sonnet")
	fallbackD.SessionToken = token
	fallbackD.Outcome = "fallback"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", fallbackD); err != nil {
		t.Fatalf("RecordDispatch fallback: %v", err)
	}

	timedOutD := baseDispatch("panel", "sonnet")
	timedOutD.SessionToken = token
	timedOutD.Outcome = "timed-out"
	if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", timedOutD); err != nil {
		t.Fatalf("RecordDispatch timed-out: %v", err)
	}

	recordFinding := func(ref string, round int, severity string) {
		t.Helper()
		f := baseFinding(ref, round)
		f.DispatchSeq = &implOut.Seq
		f.Severity = severity
		f.Status = "fixed"
		if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
			t.Fatalf("UpsertFinding %s: %v", ref, err)
		}
	}
	recordFinding("F1", 1, "Critical")
	recordFinding("F2", 1, "Important")
	recordFinding("F3", 2, "Minor")

	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-1", records.Decision{
		SessionToken: token,
		Decision: json.RawMessage(`{
			"class": "regular", "classMechanical": "small",
			"override": "two tasks touch the harvest attribution path",
			"execution": "inline",
			"implementer": "skipped — inline",
			"panel": {
				"compact": true, "rerun": "delta",
				"roster": [
					{"slot": "primary", "experimental": false},
					{"slot": "exp-failure-modes", "experimental": true,
					 "prompt": "skills/flow/experimental/failure-modes.md",
					 "description": "What the diff does under error, timeout and partial write"}
				]
			}
		}`),
	}); err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}

	period := store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}
	rows, err := st.Decisions(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("Decisions: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("Decisions returned %d rows, want 1", len(rows))
	}
	got := rows[0]

	if got.Project != projectKey || got.Change != "kan-1" {
		t.Errorf("Project/Change = %q/%q, want %q/%q", got.Project, got.Change, projectKey, "kan-1")
	}
	if got.Class != "regular" {
		t.Errorf("Class = %q, want %q", got.Class, "regular")
	}
	if !got.Overridden {
		t.Errorf("Overridden = false, want true (decision carries an override reason)")
	}
	if got.Execution != "inline" {
		t.Errorf("Execution = %q, want %q", got.Execution, "inline")
	}
	if got.ImplementerModel != "skipped — inline" {
		t.Errorf("ImplementerModel = %q, want %q", got.ImplementerModel, "skipped — inline")
	}
	if got.RosterSize != 2 {
		t.Errorf("RosterSize = %d, want 2", got.RosterSize)
	}
	if !got.Compact {
		t.Errorf("Compact = false, want true")
	}
	if got.ExperimentalSlot != "exp-failure-modes" {
		t.Errorf("ExperimentalSlot = %q, want %q", got.ExperimentalSlot, "exp-failure-modes")
	}
	if got.Rerun != "delta" {
		t.Errorf("Rerun = %q, want %q", got.Rerun, "delta")
	}
	if got.WallClockSeconds != 600 {
		t.Errorf("WallClockSeconds = %v, want 600", got.WallClockSeconds)
	}
	if got.InputTokens != 100 || got.OutputTokens != 50 || got.CacheReadTokens != 20 {
		t.Errorf("tokens = %d/%d/%d, want 100/50/20", got.InputTokens, got.OutputTokens, got.CacheReadTokens)
	}
	if got.CostUsd != 1.25 {
		t.Errorf("CostUsd = %v, want 1.25", got.CostUsd)
	}
	if got.Critical != 1 || got.Important != 1 || got.Minor != 1 {
		t.Errorf("severity counts = %+v, want Critical=1 Important=1 Minor=1", got)
	}
	if got.FixRounds != 2 {
		t.Errorf("FixRounds = %d, want 2 (max finding round)", got.FixRounds)
	}
	if got.Fallbacks != 1 {
		t.Errorf("Fallbacks = %d, want 1", got.Fallbacks)
	}
	if got.TimedOut != 1 {
		t.Errorf("TimedOut = %d, want 1", got.TimedOut)
	}
}

// TestDecisionsRendersGrouping covers task 23 (design.md's "Stats views ›
// decisions"): the decisions view's Grouping, Dispatches and
// ImplementerGroups columns, projected from panel.grouping,
// panel.dispatches and the top-level groups field. A free grouping with two
// dispatch groups and two implementer-merge groups renders each as its
// roles/ids '+'-joined within a group and ' · '-joined across groups,
// whether the groups are {slots|bundles, model, effort} objects or the bare
// id arrays older rows recorded; a decision whose panel is still the bare
// "default" string (the toggle was off) renders Grouping as "default" with
// no dispatches to show, and a nil groups field (inline execution) renders
// as empty.
func TestDecisionsRendersGrouping(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-decisions-grouping-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-free")
	seedChange(t, st, projectKey, "kan-legacy")
	seedChange(t, st, projectKey, "kan-default")

	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-free", records.Decision{
		SessionToken: "mf-decisions-grouping-free",
		Decision: json.RawMessage(`{
			"class": "big", "execution": "sdd",
			"implementer": {"model": "opus", "effort": "high"},
			"panel": {
				"compact": false, "rerun": "full",
				"roster": [
					{"slot": "primary", "experimental": false},
					{"slot": "principles", "experimental": false},
					{"slot": "code-review-low", "experimental": false},
					{"slot": "mutation", "experimental": false}
				],
				"grouping": "free",
				"dispatches": [
					{"slots": ["primary", "principles"], "model": "opus", "effort": "high"},
					{"slots": ["code-review-low", "mutation"], "model": "sonnet", "effort": "high"}
				],
				"grouping_reason": "reading roles together, mutating roles together"
			},
			"groups": [
				{"bundles": [1, 2], "model": "opus", "effort": "high"},
				{"bundles": [3], "model": "sonnet", "effort": "medium"}
			],
			"groups_reason": "tasks 1-2 share a file, task 3 stands alone on sonnet: mechanical"
		}`),
	}); err != nil {
		t.Fatalf("RecordDecision kan-free: %v", err)
	}

	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-legacy", records.Decision{
		SessionToken: "mf-decisions-grouping-legacy",
		Decision: json.RawMessage(`{
			"class": "big", "execution": "sdd",
			"implementer": {"model": "opus", "effort": "high"},
			"panel": {
				"compact": false, "rerun": "full",
				"roster": [{"slot": "primary", "model": "opus", "effort": "high", "experimental": false}],
				"grouping": "free",
				"dispatches": [["primary", "principles"], ["code-review-low", "mutation"]],
				"grouping_reason": "legacy per-slot row"
			},
			"groups": [[1, 2], [3]],
			"groups_reason": "legacy bare arrays"
		}`),
	}); err != nil {
		t.Fatalf("RecordDecision kan-legacy: %v", err)
	}

	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-default", records.Decision{
		SessionToken: "mf-decisions-grouping-default",
		Decision: json.RawMessage(`{
			"class": "small", "execution": "inline",
			"implementer": "skipped — inline",
			"panel": "default",
			"groups": null
		}`),
	}); err != nil {
		t.Fatalf("RecordDecision kan-default: %v", err)
	}

	period := store.Period{From: time.Now().Add(-time.Hour), To: time.Now().Add(time.Hour)}
	rows, err := st.Decisions(ctx, period, &projectKey)
	if err != nil {
		t.Fatalf("Decisions: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("Decisions returned %d rows, want 3", len(rows))
	}

	byChange := map[string]store.DecisionRow{}
	for _, r := range rows {
		byChange[r.Change] = r
	}

	free, ok := byChange["kan-free"]
	if !ok {
		t.Fatalf("no row for kan-free")
	}
	if free.Grouping != "free" {
		t.Errorf("kan-free Grouping = %q, want %q", free.Grouping, "free")
	}
	if free.Dispatches != "primary+principles · code-review-low+mutation" {
		t.Errorf("kan-free Dispatches = %q, want %q", free.Dispatches, "primary+principles · code-review-low+mutation")
	}
	if free.ImplementerGroups != "1+2 · 3" {
		t.Errorf("kan-free ImplementerGroups = %q, want %q", free.ImplementerGroups, "1+2 · 3")
	}

	legacy, ok := byChange["kan-legacy"]
	if !ok {
		t.Fatalf("no row for kan-legacy")
	}
	if legacy.Dispatches != "primary+principles · code-review-low+mutation" {
		t.Errorf("kan-legacy Dispatches = %q, want %q", legacy.Dispatches, "primary+principles · code-review-low+mutation")
	}
	if legacy.ImplementerGroups != "1+2 · 3" {
		t.Errorf("kan-legacy ImplementerGroups = %q, want %q", legacy.ImplementerGroups, "1+2 · 3")
	}

	def, ok := byChange["kan-default"]
	if !ok {
		t.Fatalf("no row for kan-default")
	}
	if def.Grouping != "default" {
		t.Errorf("kan-default Grouping = %q, want %q", def.Grouping, "default")
	}
	if def.Dispatches != "" {
		t.Errorf("kan-default Dispatches = %q, want empty", def.Dispatches)
	}
	if def.ImplementerGroups != "" {
		t.Errorf("kan-default ImplementerGroups = %q, want empty", def.ImplementerGroups)
	}
}
