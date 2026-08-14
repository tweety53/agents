package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

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

	inside := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	inside.StartedAt = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	runStage(t, st, inside, json.RawMessage(`{"cost_usd":1.5}`), inside.StartedAt.Add(time.Minute), "completed")

	before := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	before.StartedAt = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	runStage(t, st, before, json.RawMessage(`{"cost_usd":9}`), before.StartedAt.Add(time.Minute), "completed")

	after := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
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

	run := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	base := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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
		in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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
		in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
		in.SessionID = ptr(fmt.Sprintf("day1-%d", i))
		in.StartedAt = day1.Add(time.Duration(i) * time.Hour)
		costRun(t, st, in, cost)
	}
	day2Costs := []float64{40, 50} // total 90
	for i, cost := range day2Costs {
		in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
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
func TestPanelEconomicsComputesTokensAndFindingsPerMTok(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-panel-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1") // baseChange sets ReviewPanelRoster "light"

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	// total tokens = 1,000,000 exactly, so FindingsPerMTok reduces to
	// FindingsTotal itself -- a deliberately round number for a hand-check.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"tokens": {"main": {"input": 700000, "output": 300000}},
		"findings_by_severity": {"high": 2, "low": 1}
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
	rows, err := st.PanelEconomics(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("PanelEconomics: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("PanelEconomics returned %d rows, want 1", len(rows))
	}
	row := rows[0]

	if row.ReviewPanelRoster != "light" {
		t.Errorf("ReviewPanelRoster = %q, want light", row.ReviewPanelRoster)
	}
	if row.FindingsTotal != 3 {
		t.Errorf("FindingsTotal = %d, want 3", row.FindingsTotal)
	}
	if row.TokensTotal == nil || *row.TokensTotal != 1000000 {
		t.Errorf("TokensTotal = %v, want 1000000", row.TokensTotal)
	}
	const eps = 1e-9
	if row.FindingsPerMTok == nil {
		t.Fatalf("FindingsPerMTok is nil, want 3")
	}
	if diff := *row.FindingsPerMTok - 3; diff > eps || diff < -eps {
		t.Errorf("FindingsPerMTok = %v, want 3 (3 findings over exactly 1 Mtok)", *row.FindingsPerMTok)
	}
}

// TestModelComparisonComputesMeanCostAndRework seeds two stage runs of the
// same model/command/stage triple -- attempt 1 and attempt 2, a fix
// re-run -- with known costs, and checks MeanCostUSD and ReworkAttempts
// against hand-computed values: mean (10+20)/2 = 15, and exactly one of
// the two runs (attempt 2) counted as rework.
func TestModelComparisonComputesMeanCostAndRework(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-modelcmp-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	first, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage(first): %v", err)
	}
	if first.Attempt != 1 {
		t.Fatalf("first attempt = %d, want 1", first.Attempt)
	}
	if err := st.MergeMetrics(ctx, first.ID, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":10}}}`)); err != nil {
		t.Fatalf("MergeMetrics(first): %v", err)
	}
	if err := st.EndStage(ctx, first.ID, in.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage(first): %v", err)
	}

	second, err := st.BeginStage(ctx, in) // a fix re-run: attempt 2
	if err != nil {
		t.Fatalf("BeginStage(second): %v", err)
	}
	if second.Attempt != 2 {
		t.Fatalf("second attempt = %d, want 2", second.Attempt)
	}
	if err := st.MergeMetrics(ctx, second.ID, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":20}}}`)); err != nil {
		t.Fatalf("MergeMetrics(second): %v", err)
	}
	if err := st.EndStage(ctx, second.ID, in.StartedAt.Add(2*time.Hour), "completed"); err != nil {
		t.Fatalf("EndStage(second): %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.ModelComparison(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("ModelComparison: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("ModelComparison returned %d rows, want 1", len(rows))
	}
	row := rows[0]

	if row.Model != "claude-opus-5" {
		t.Errorf("Model = %q, want claude-opus-5", row.Model)
	}
	if row.RunCount != 2 {
		t.Errorf("RunCount = %d, want 2", row.RunCount)
	}
	const eps = 1e-9
	if row.MeanCostUSD == nil {
		t.Fatalf("MeanCostUSD is nil, want 15")
	}
	if diff := *row.MeanCostUSD - 15; diff > eps || diff < -eps {
		t.Errorf("MeanCostUSD = %v, want 15", *row.MeanCostUSD)
	}
	if row.ReworkAttempts != 1 {
		t.Errorf("ReworkAttempts = %d, want 1 (only attempt 2 counts as rework)", row.ReworkAttempts)
	}
}

// TestModelComparisonProducesOneRowPerModelInAMixedRun seeds a single
// stage run whose metrics carry a "models" bucket for two models (the
// review panel's own parent-model-vs-reviewer-model split -- see
// tasks.md task 22's "why this task exists"), each with its own
// cost_usd, and asserts ModelComparison produces two rows, one per
// model, each carrying that model's own mean cost. Under the retired
// scalar `metrics->>'model'` grouping this run collapses to exactly one
// row (whichever model was last written), so this test fails against
// that implementation -- which is what makes it a real test of task 22's
// fix rather than a restatement of it.
func TestModelComparisonProducesOneRowPerModelInAMixedRun(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-modelcmp-mixed-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)

	run := runStage(t, st, in, json.RawMessage(`{
		"tokens": {"main": {"input": 100}, "sidechain": {"input": 400}},
		"models": {
			"claude-opus-5":   {"tokens": {"main": {"input": 100}}, "cost_usd": 10},
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 400}}, "cost_usd": 5}
		}
	}`), in.StartedAt.Add(time.Minute), "completed")

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.ModelComparison(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("ModelComparison: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("ModelComparison returned %d rows, want 2 (one per model in the mixed run)", len(rows))
	}

	byModel := map[string]store.ModelComparisonRow{}
	for _, r := range rows {
		byModel[r.Model] = r
	}
	opus, ok := byModel["claude-opus-5"]
	if !ok {
		t.Fatalf("no row for claude-opus-5: %v", rows)
	}
	if opus.RunCount != 1 {
		t.Errorf("claude-opus-5 RunCount = %d, want 1", opus.RunCount)
	}
	if opus.MeanCostUSD == nil || *opus.MeanCostUSD != 10 {
		t.Errorf("claude-opus-5 MeanCostUSD = %v, want 10", opus.MeanCostUSD)
	}

	sonnet, ok := byModel["claude-sonnet-5"]
	if !ok {
		t.Fatalf("no row for claude-sonnet-5: %v", rows)
	}
	if sonnet.RunCount != 1 {
		t.Errorf("claude-sonnet-5 RunCount = %d, want 1", sonnet.RunCount)
	}
	if sonnet.MeanCostUSD == nil || *sonnet.MeanCostUSD != 5 {
		t.Errorf("claude-sonnet-5 MeanCostUSD = %v, want 5", sonnet.MeanCostUSD)
	}

	// The models bucket's own token totals sum to the run's top-level
	// "tokens" total -- the same guarantee attribute_test.go's
	// TestAttributeBucketsTokensPerModel checks at the harvester layer,
	// re-checked here against what actually landed in Postgres, unmerged
	// by any store-side transformation.
	var bag struct {
		Tokens struct {
			Main struct {
				Input float64 `json:"input"`
			} `json:"main"`
			Sidechain struct {
				Input float64 `json:"input"`
			} `json:"sidechain"`
		} `json:"tokens"`
		Models map[string]struct {
			Tokens struct {
				Main struct {
					Input float64 `json:"input"`
				} `json:"main"`
				Sidechain struct {
					Input float64 `json:"input"`
				} `json:"sidechain"`
			} `json:"tokens"`
		} `json:"models"`
	}
	if err := json.Unmarshal(run.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	sumMain := bag.Models["claude-opus-5"].Tokens.Main.Input + bag.Models["claude-sonnet-5"].Tokens.Main.Input
	sumSidechain := bag.Models["claude-opus-5"].Tokens.Sidechain.Input + bag.Models["claude-sonnet-5"].Tokens.Sidechain.Input
	if sumMain != bag.Tokens.Main.Input {
		t.Errorf("models main.input sum = %v, want top-level tokens.main.input = %v", sumMain, bag.Tokens.Main.Input)
	}
	if sumSidechain != bag.Tokens.Sidechain.Input {
		t.Errorf("models sidechain.input sum = %v, want top-level tokens.sidechain.input = %v", sumSidechain, bag.Tokens.Sidechain.Input)
	}
}

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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	if rows, err := st.ReworkRate(ctx, period, &projectKey, nil); err != nil {
		t.Errorf("ReworkRate: %v", err)
	} else if len(rows) != 1 || rows[0].TotalAttempts != 1 {
		t.Errorf("ReworkRate rows = %+v, want one row with TotalAttempts 1", rows)
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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
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
// exactly one model, and iterates every one of the seven model-filterable
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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
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
		{"PanelEconomics", func(m *string) (int, error) {
			r, err := st.PanelEconomics(ctx, period, &projectKey, m)
			return len(r), err
		}},
		{"ModelComparison", func(m *string) (int, error) {
			r, err := st.ModelComparison(ctx, period, &projectKey, m)
			return len(r), err
		}},
		{"ReworkRate", func(m *string) (int, error) { r, err := st.ReworkRate(ctx, period, &projectKey, m); return len(r), err }},
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

	withModel := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	withModel.SessionID = ptr("s-with")
	withModel.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, withModel, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":5}}}`), withModel.StartedAt.Add(time.Minute), "completed")

	noModelA := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
	noModelA.SessionID = ptr("s-none-a")
	noModelA.Harness = "cursor"
	noModelA.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, noModelA, json.RawMessage(`{"tokens_available":false}`), noModelA.StartedAt.Add(time.Minute), "completed")

	noModelB := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
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

	emptyModels := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
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

	a := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	a.SessionID = ptr("s-a")
	a.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	runStage(t, st, a, json.RawMessage(`{"models":{"claude-sonnet-5":{"cost_usd":1}}}`), a.StartedAt.Add(time.Minute), "completed")

	b := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	b.SessionID = ptr("s-b")
	b.StartedAt = time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC)
	runStage(t, st, b, json.RawMessage(`{"models":{"claude-opus-5":{"cost_usd":2},"claude-sonnet-5":{"cost_usd":1}}}`), b.StartedAt.Add(time.Minute), "completed")

	outsidePeriod := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
	outsidePeriod.SessionID = ptr("s-outside")
	outsidePeriod.StartedAt = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	runStage(t, st, outsidePeriod, json.RawMessage(`{"models":{"claude-haiku-5":{"cost_usd":1}}}`), outsidePeriod.StartedAt.Add(time.Minute), "completed")

	noModel := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
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
