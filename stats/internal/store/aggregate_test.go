package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
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

	board, err := st.LiveStateBoard(ctx, emptyPeriod, &projectKey, nil)
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

	// A run on a harness with no transcript, so no tokens key at all.
	unmeasured := base
	unmeasured.SessionID = ptr("s-c")
	unmeasured.Harness = "no-transcript-harness"
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
	if diff := row.TotalCostUSD - 150; diff > eps || diff < -eps {
		t.Errorf("TotalCostUSD = %v, want 150", row.TotalCostUSD)
	}
}

// TestStageLeaderboardKeepsTopTenByTotalCost: the leaderboard ranks stages
// by total cost and keeps only the ten costliest (design.md's
// leaderboard-top10-total) -- eleven stages in, the cheapest is gone and
// the rest arrive total-descending.
func TestStageLeaderboardKeepsTopTenByTotalCost(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-leaderboard-top-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	// Stage i costs i+1 in total, as two runs, so the rank is by the sum
	// and not by any one run.
	for i := range 11 {
		for j := range 2 {
			in := baseBeginInput(projectKey, "kan-1", "/flow", fmt.Sprintf("stage-%02d", i))
			in.SessionID = ptr(fmt.Sprintf("session-%d-%d", i, j))
			in.StartedAt = time.Date(2026, 6, 10, i, j, 0, 0, time.UTC)
			costRun(t, st, in, float64(i+1)/2)
		}
	}

	rows, err := st.StageLeaderboard(ctx, store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil {
		t.Fatalf("StageLeaderboard: %v", err)
	}
	if len(rows) != 10 {
		t.Fatalf("StageLeaderboard returned %d rows, want 10", len(rows))
	}
	for k, row := range rows {
		if want := fmt.Sprintf("stage-%02d", 10-k); row.Stage != want {
			t.Errorf("rows[%d].Stage = %q, want %q (total-cost descending)", k, row.Stage, want)
		}
		if want := float64(11 - k); row.TotalCostUSD != want {
			t.Errorf("rows[%d].TotalCostUSD = %v, want %v", k, row.TotalCostUSD, want)
		}
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

// TestWeeklyCostsBucketsByWednesdayAndCategory seeds runs either side of
// a Wednesday 00:00 UTC boundary, across every category plus one stage in
// none, and checks each week's overall and per-category totals -- a loop
// change's run counting as self-review whatever its stage.
func TestWeeklyCostsBucketsByWednesdayAndCategory(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-weekly-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "self-review-kan-1")
	seedChange(t, st, projectKey, "pipeline-fix-kan-1-1")

	runs := []struct {
		change string
		stage  string
		at     time.Time
		cost   float64
	}{
		{"kan-1", "flow.sdd-tdd", time.Date(2026, 6, 9, 23, 0, 0, 0, time.UTC), 5}, // Tue: week of Wed 06-03
		// Every planning stage carries its own share of the week's $10.
		{"kan-1", "flow.brainstorm", time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), 5},
		{"kan-1", "plan.session", time.Date(2026, 6, 10, 0, 10, 0, 0, time.UTC), 1},
		{"kan-1", "flow.design-approval", time.Date(2026, 6, 10, 0, 20, 0, 0, time.UTC), 1},
		{"kan-1", "flow.create-artifacts", time.Date(2026, 6, 10, 0, 30, 0, 0, time.UTC), 1},
		{"kan-1", "flow.writing-plans", time.Date(2026, 6, 10, 0, 40, 0, 0, time.UTC), 1},
		{"kan-1", "flow.decide", time.Date(2026, 6, 10, 0, 50, 0, 0, time.UTC), 1},
		{"kan-1", "flow.sdd-tdd", time.Date(2026, 6, 10, 1, 0, 0, 0, time.UTC), 20},
		{"kan-1", "flow.review-panel", time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC), 30},
		{"kan-1", "flow.verify", time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC), 7}, // overall only
		{"self-review-kan-1", "flow.sdd-tdd", time.Date(2026, 6, 13, 0, 0, 0, 0, time.UTC), 20},
		{"pipeline-fix-kan-1-1", "flow.brainstorm", time.Date(2026, 6, 13, 1, 0, 0, 0, time.UTC), 20},
		{"kan-1", "flow.self-review", time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC), 3},
	}
	for i, r := range runs {
		in := baseBeginInput(projectKey, r.change, "/flow", r.stage)
		in.SessionID = ptr(fmt.Sprintf("weekly-%d", i))
		in.StartedAt = r.at
		costRun(t, st, in, r.cost)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	points, err := st.WeeklyCosts(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("WeeklyCosts: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("WeeklyCosts returned %d points, want 2", len(points))
	}

	const eps = 1e-9
	check := func(label string, got, want *float64) {
		t.Helper()
		switch {
		case want == nil && got != nil:
			t.Errorf("%s = %v, want nil", label, *got)
		case want != nil && (got == nil || *got < *want-eps || *got > *want+eps):
			t.Errorf("%s = %v, want %v", label, got, *want)
		}
	}

	if want := time.Date(2026, 6, 3, 0, 0, 0, 0, time.UTC); !points[0].Week.Equal(want) {
		t.Errorf("points[0].Week = %v, want %v", points[0].Week, want)
	}
	check("points[0].OverallUSD", points[0].OverallUSD, ptr(5.0))
	check("points[0].PlanningUSD", points[0].PlanningUSD, nil)
	check("points[0].ImplementUSD", points[0].ImplementUSD, ptr(5.0))
	check("points[0].ReviewUSD", points[0].ReviewUSD, nil)
	check("points[0].SelfReviewUSD", points[0].SelfReviewUSD, nil)

	if want := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC); !points[1].Week.Equal(want) {
		t.Errorf("points[1].Week = %v, want %v", points[1].Week, want)
	}
	check("points[1].OverallUSD", points[1].OverallUSD, ptr(110.0))
	check("points[1].PlanningUSD", points[1].PlanningUSD, ptr(10.0))
	check("points[1].ImplementUSD", points[1].ImplementUSD, ptr(20.0))
	check("points[1].ReviewUSD", points[1].ReviewUSD, ptr(30.0))
	check("points[1].SelfReviewUSD", points[1].SelfReviewUSD, ptr(43.0))
}

// TestWeeklyCostsHonoursProjectModelAndPeriod checks the three filters
// WeeklyCosts shares with TrendOverTime: another project's run, a run
// starting exactly at period.To, and -- with a model -- that model's own
// cost_usd rather than the run's total, and no run lacking that model.
func TestWeeklyCostsHonoursProjectModelAndPeriod(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-weekly-f-%d", time.Now().UnixNano())
	otherKey := projectKey + "-other"
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, otherKey, "kan-1")

	period := store.Period{
		From: time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 6, 17, 0, 0, 0, 0, time.UTC),
	}
	run := func(project string, i int, at time.Time, metrics string) {
		t.Helper()
		in := baseBeginInput(project, "kan-1", "/flow", "flow.review-panel")
		in.SessionID = ptr(fmt.Sprintf("weekly-f-%d", i))
		in.StartedAt = at
		r, err := st.BeginStage(ctx, in)
		if err != nil {
			t.Fatalf("BeginStage: %v", err)
		}
		if err := st.MergeMetrics(ctx, r.ID, json.RawMessage(metrics)); err != nil {
			t.Fatalf("MergeMetrics: %v", err)
		}
	}
	run(projectKey, 0, period.From, `{"cost_usd": 10, "models": {"m-a": {"cost_usd": 4}, "m-b": {"cost_usd": 6}}}`)
	run(projectKey, 1, period.From.Add(time.Hour), `{"cost_usd": 3, "models": {"m-b": {"cost_usd": 3}}}`)
	run(projectKey, 2, period.To, `{"cost_usd": 100, "models": {"m-a": {"cost_usd": 100}}}`)
	run(otherKey, 3, period.From, `{"cost_usd": 1000, "models": {"m-a": {"cost_usd": 1000}}}`)
	// A week whose only run lacks m-a: no point under the m-a filter.
	run(projectKey, 4, period.From.Add(-24*time.Hour), `{"cost_usd": 2, "models": {"m-b": {"cost_usd": 2}}}`)
	period.From = period.From.Add(-7 * 24 * time.Hour)

	for _, tc := range []struct {
		name   string
		model  *string
		points int
		want   float64
	}{
		{"all models", nil, 2, 13},
		{"model m-a", ptr("m-a"), 1, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			points, err := st.WeeklyCosts(ctx, period, &projectKey, tc.model)
			if err != nil {
				t.Fatalf("WeeklyCosts: %v", err)
			}
			if len(points) != tc.points {
				t.Fatalf("WeeklyCosts returned %d points, want %d", len(points), tc.points)
			}
			last := points[len(points)-1]
			for label, got := range map[string]*float64{"OverallUSD": last.OverallUSD, "ReviewUSD": last.ReviewUSD} {
				if got == nil || *got != tc.want {
					t.Errorf("%s = %v, want %v", label, got, tc.want)
				}
			}
		})
	}
}

// TestWeeklyCostsCountsVisualVerify pins visual-verify-category and
// other-category: a flow.visual-verify run counts in its own column (a
// loop change's still as self-review), and a run in no named phase counts
// in Other, nil in a week whose runs all fall in named phases.
func TestWeeklyCostsCountsVisualVerify(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-weekly-vv-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "self-review-kan-1")

	runs := []struct {
		change string
		stage  string
		at     time.Time
		cost   float64
	}{
		{"kan-1", "flow.sdd-tdd", time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC), 5}, // week of 06-03: named phases only
		{"kan-1", "flow.visual-verify", time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC), 2},
		{"self-review-kan-1", "flow.visual-verify", time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC), 4},
		{"kan-1", "flow.verify", time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC), 1},
	}
	for i, r := range runs {
		in := baseBeginInput(projectKey, r.change, "/flow", r.stage)
		in.SessionID = ptr(fmt.Sprintf("weekly-vv-%d", i))
		in.StartedAt = r.at
		costRun(t, st, in, r.cost)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	points, err := st.WeeklyCosts(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("WeeklyCosts: %v", err)
	}
	if len(points) != 2 {
		t.Fatalf("WeeklyCosts returned %d points, want 2", len(points))
	}
	check := func(label string, got, want *float64) {
		t.Helper()
		switch {
		case want == nil && got != nil:
			t.Errorf("%s = %v, want nil", label, *got)
		case want != nil && (got == nil || *got != *want):
			t.Errorf("%s = %v, want %v", label, got, *want)
		}
	}
	check("points[0].VisualVerifyUSD", points[0].VisualVerifyUSD, nil)
	check("points[0].OtherUSD", points[0].OtherUSD, nil)
	check("points[1].VisualVerifyUSD", points[1].VisualVerifyUSD, ptr(2.0))
	check("points[1].SelfReviewUSD", points[1].SelfReviewUSD, ptr(4.0))
	check("points[1].OtherUSD", points[1].OtherUSD, ptr(1.0))
	check("points[1].OverallUSD", points[1].OverallUSD, ptr(7.0))
}

// TestWeeklyModelEffortCostsSplitsDispatchAndSession pins
// dispatch-served-effort and session-line: priced dispatches chart by
// served model and effort (default when none was served), and each model
// bucket's remainder after its dispatches charts as that model's session.
func TestWeeklyModelEffortCostsSplitsDispatchAndSession(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-weekly-me-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run := func(i int, at time.Time, metrics string) {
		t.Helper()
		in := baseBeginInput(projectKey, "kan-1", "/flow", "flow.sdd-tdd")
		in.SessionID = ptr(fmt.Sprintf("weekly-me-%d", i))
		in.StartedAt = at
		r, err := st.BeginStage(ctx, in)
		if err != nil {
			t.Fatalf("BeginStage: %v", err)
		}
		if err := st.MergeMetrics(ctx, r.ID, json.RawMessage(metrics)); err != nil {
			t.Fatalf("MergeMetrics: %v", err)
		}
	}
	run(0, time.Date(2026, 6, 11, 0, 0, 0, 0, time.UTC), `{
		"cost_usd": 12,
		"models": {"claude-opus-5-5": {"cost_usd": 10}, "claude-sonnet-5-5": {"cost_usd": 2}},
		"dispatches": {
			"a1": {"model": "opus", "cost_usd": 3, "signals": {"served_models": {"claude-opus-5-5": 2}, "served_efforts": {"low": 2}}},
			"a2": {"model": "sonnet", "cost_usd": 2, "signals": {"served_models": {"claude-sonnet-5-5": 2}, "served_efforts": {"medium": 2}}},
			"a3": {"model": "opus", "cost_usd": 1, "signals": {"served_models": {"claude-opus-5-5": 1}}},
			"a4": {"model": "opus"}
		}
	}`)
	// Outside the period: contributes nothing.
	run(1, time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC), `{"cost_usd": 50, "models": {"claude-opus-5-5": {"cost_usd": 50}}}`)
	// A run with an unpriceable bucket carries no total: Overall excludes
	// it, so no line counts its priced bucket either.
	run(3, time.Date(2026, 6, 12, 0, 0, 0, 0, time.UTC), `{
		"models": {"claude-opus-5-5": {"cost_usd": 7}, "no-such-model": {"tokens": {"main": {"input": 1}}}}
	}`)
	// The week after: dispatches costing more than their model bucket floor
	// the session at 0, which is dropped; the most-turns served model and
	// effort win, a turns tie breaking by name.
	run(2, time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC), `{
		"cost_usd": 1,
		"models": {"claude-opus-5-5": {"cost_usd": 1}},
		"dispatches": {
			"b1": {"model": "opus", "cost_usd": 3, "signals": {
				"served_models": {"claude-sonnet-5-5": 1, "claude-opus-5-5": 4},
				"served_efforts": {"medium": 2, "high": 2}}},
			"b2": {"model": "opus", "cost_usd": 4, "signals": {
				"served_models": {"claude-sonnet-5-5": 3, "claude-opus-5-5": 3},
				"served_efforts": {"low": 1, "xhigh": 5}}}
		}
	}`)

	// An unattached plan session: Overall excludes it, so no line counts it.
	plan, err := st.BeginStage(ctx, store.BeginStageInput{
		ProjectKey: projectKey, MainCheckoutPath: "/tmp/" + projectKey, JiraKey: "KAN-900",
		Harness: "claude-code", Command: "/flow-plan", Stage: "plan.session",
		StartedAt: time.Date(2026, 6, 11, 1, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("BeginStage plan session: %v", err)
	}
	if err := st.MergeMetrics(ctx, plan.ID, json.RawMessage(`{"cost_usd": 9, "models": {"claude-haiku-5-5": {"cost_usd": 9}}}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	points, err := st.WeeklyModelEffortCosts(ctx, period)
	if err != nil {
		t.Fatalf("WeeklyModelEffortCosts: %v", err)
	}
	week := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	want := []store.WeeklyModelEffortPoint{
		{Week: week, Model: "claude-opus-5-5", Effort: "default", CostUSD: 1},
		{Week: week, Model: "claude-opus-5-5", Effort: "low", CostUSD: 3},
		{Week: week, Model: "claude-opus-5-5", Effort: "session", CostUSD: 6},
		{Week: week, Model: "claude-sonnet-5-5", Effort: "medium", CostUSD: 2},
		{Week: week.AddDate(0, 0, 7), Model: "claude-opus-5-5", Effort: "high", CostUSD: 3},
		{Week: week.AddDate(0, 0, 7), Model: "claude-opus-5-5", Effort: "xhigh", CostUSD: 4},
	}
	if len(points) != len(want) {
		t.Fatalf("WeeklyModelEffortCosts returned %d rows %+v, want %d", len(points), points, len(want))
	}
	for i, w := range want {
		g := points[i]
		if !g.Week.Equal(w.Week) || g.Model != w.Model || g.Effort != w.Effort || g.CostUSD != w.CostUSD {
			t.Errorf("row %d = %+v, want %+v", i, g, w)
		}
	}
}

// TestPanelEconomicsComputesTokensAndFindingsPerMTok seeds one stage run
// with a known token total and a known findings-by-severity bag, and
// checks TokensTotal and FindingsPerMTok against values hand-computed from
// them, not just FindingsTotal's row count as before.
// TestAggregateOtherViewsRunWithoutError covers the view not exercised
// with hand-computed arithmetic elsewhere: LiveStateBoard (a projection of
// changes, already covered thoroughly by changes_test.go).
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

	if _, err := st.LiveStateBoard(ctx, period, nil, nil); err != nil {
		t.Errorf("LiveStateBoard: %v", err)
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
	noModelA.Harness = "no-transcript-harness"
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
		d.StartedAt = in.StartedAt
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
	d.StartedAt = in.StartedAt
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
	seedRoster(t, st, projectKey, "kan-1", "primary", "principles")

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

// TestReviewersIncludesExtraDispatchSlots seeds the always-on
// test-sanity+call-efficiency bundle, whose slots a decision records under
// extra_dispatch.slots and never under panel.roster, and asserts both roles
// surface as rows, each with its finding and its description.
func TestReviewersIncludesExtraDispatchSlots(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-reviewers-extra-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	d := baseDispatch("reviewer", "sonnet")
	d.Slot = "test-sanity+call-efficiency"
	d.StartedAt = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	bundle, err := st.RecordDispatch(ctx, projectKey, "kan-1", d)
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	f := baseFinding("F1", 1)
	f.Slot = "test-sanity"
	f.Severity = "Minor"
	f.Status = "fixed"
	f.DispatchSeq = &bundle.Seq
	if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}
	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-1", records.Decision{
		SessionToken: "mf-extra-" + projectKey,
		Decision:     json.RawMessage(`{"panel": "default", "extra_dispatch": {"slots": ["test-sanity", "call-efficiency"], "model": "sonnet", "effort": "low"}}`),
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
	bySlot := map[string]store.ReviewerRow{}
	for _, r := range rows {
		bySlot[r.Slot] = r
	}
	for _, slot := range []string{"test-sanity", "call-efficiency"} {
		r, ok := bySlot[slot]
		if !ok {
			t.Fatalf("no row for extra-dispatch slot %q (rows: %+v)", slot, rows)
		}
		if r.Dispatches != 1 {
			t.Errorf("%s Dispatches = %d, want 1", slot, r.Dispatches)
		}
		if r.Description == "" {
			t.Errorf("%s Description is empty", slot)
		}
	}
	if got := bySlot["test-sanity"].Minor; got != 1 {
		t.Errorf("test-sanity Minor = %d, want 1", got)
	}
}

// TestReviewersCountsMutationExtraDispatch seeds the always-on
// mutation+failure-modes bundle (kan-1028), whose slots a decision records
// under extra_dispatch_mutation.slots and no roster names, and asserts both
// roles surface as rows, failure-modes with its finding, each with the
// description review-panel.md gives it.
func TestReviewersCountsMutationExtraDispatch(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-reviewers-extramf-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	d := baseDispatch("reviewer", "opus")
	d.Slot = "mutation+failure-modes"
	d.StartedAt = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	bundle, err := st.RecordDispatch(ctx, projectKey, "kan-1", d)
	if err != nil {
		t.Fatalf("RecordDispatch: %v", err)
	}
	f := baseFinding("F1", 1)
	f.Slot = "failure-modes"
	f.Severity = "Important"
	f.Status = "fixed"
	f.DispatchSeq = &bundle.Seq
	if _, _, err := st.UpsertFinding(ctx, projectKey, "kan-1", f); err != nil {
		t.Fatalf("UpsertFinding: %v", err)
	}
	if _, _, err := st.RecordDecision(ctx, projectKey, "kan-1", records.Decision{
		SessionToken: "mf-extramf-" + projectKey,
		Decision:     json.RawMessage(`{"panel": "default", "extra_dispatch_mutation": {"slots": ["mutation", "failure-modes"], "model": "opus", "effort": "low"}}`),
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
	bySlot := map[string]store.ReviewerRow{}
	for _, r := range rows {
		bySlot[r.Slot] = r
	}
	for slot, desc := range map[string]string{
		"mutation":      "Mutation — sabotage-proofing",
		"failure-modes": "Failure-modes — error return, timeout, partial write, concurrent re-entry",
	} {
		r, ok := bySlot[slot]
		if !ok {
			t.Fatalf("no row for extra-dispatch slot %q (rows: %+v)", slot, rows)
		}
		if r.Dispatches != 1 {
			t.Errorf("%s Dispatches = %d, want 1", slot, r.Dispatches)
		}
		if r.Description != desc {
			t.Errorf("%s Description = %q, want %q", slot, r.Description, desc)
		}
	}
	if got := bySlot["failure-modes"].Important; got != 1 {
		t.Errorf("failure-modes Important = %d, want 1", got)
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
	// Cost lives on the stage run, where Price writes it -- a dispatch's
	// own metrics bag never carries cost_usd.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"cost_usd":1.25}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	implD := baseDispatch("implementer", "sonnet")
	implD.SessionToken = token
	implD.Outcome = "completed"
	implD.Metrics = json.RawMessage(`{"tokens":{"main":{"input":100,"output":50,"cache_read":20}}}`)
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
			"fixer": {"model": "haiku", "effort": "low", "reason": "one-line fixes"},
			"panel": {
				"compact": true, "rerun": "delta",
				"rerun_dispatch": {"model": "haiku", "effort": "low", "reason": "unused by pass 1"},
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
	if got.Fixer != "haiku/low" {
		t.Errorf("Fixer = %q, want %q", got.Fixer, "haiku/low")
	}
	if got.RerunDispatch != "haiku/low" {
		t.Errorf("RerunDispatch = %q, want %q", got.RerunDispatch, "haiku/low")
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
// roles/ids '+'-joined within a group and ' · '-joined across groups; a
// decision whose panel is the bare "default" string (a micro class)
// renders Grouping as "default" with no dispatches to show, and a nil
// groups field (inline execution) renders as empty.
func TestDecisionsRendersGrouping(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-decisions-grouping-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-free")
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
					{"slot": "bugbot", "experimental": false},
					{"slot": "mutation", "experimental": false}
				],
				"grouping": "free",
				"dispatches": [
					{"slots": ["primary", "principles"], "model": "opus", "effort": "high"},
					{"slots": ["bugbot", "mutation"], "model": "sonnet", "effort": "high"}
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
	if len(rows) != 2 {
		t.Fatalf("Decisions returned %d rows, want 2", len(rows))
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
	if free.Dispatches != "primary+principles · bugbot+mutation" {
		t.Errorf("kan-free Dispatches = %q, want %q", free.Dispatches, "primary+principles · bugbot+mutation")
	}
	if free.ImplementerGroups != "1+2 · 3" {
		t.Errorf("kan-free ImplementerGroups = %q, want %q", free.ImplementerGroups, "1+2 · 3")
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

// TestLiveStateBoardCarriesPlannedAndCurrentTasks: the board's task-count
// columns are the first and latest observations of the change's
// change_task_counts series (KAN-415) -- and a change never observed reads
// nil, nil, the board's absence-is-never-zero case, never a count of zero.
func TestLiveStateBoardCarriesPlannedAndCurrentTasks(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-growth-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")

	if _, err := st.RecordTaskCount(ctx, projectKey, "kan-1", records.TaskCount{TotalTasks: 22}); err != nil {
		t.Fatalf("RecordTaskCount(22): %v", err)
	}
	if _, err := st.RecordTaskCount(ctx, projectKey, "kan-1", records.TaskCount{TotalTasks: 46}); err != nil {
		t.Fatalf("RecordTaskCount(46): %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC),
	}
	rows, err := st.LiveStateBoard(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("LiveStateBoard: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("LiveStateBoard returned %d rows, want 2", len(rows))
	}

	byName := map[string]LiveStateByName{}
	for _, row := range rows {
		byName[row.Name] = LiveStateByName{Planned: row.PlannedTasks, Current: row.CurrentTasks}
	}

	grown := byName["kan-1"]
	if grown.Planned == nil || *grown.Planned != 22 {
		t.Errorf("kan-1 PlannedTasks = %v, want 22", grown.Planned)
	}
	if grown.Current == nil || *grown.Current != 46 {
		t.Errorf("kan-1 CurrentTasks = %v, want 46", grown.Current)
	}

	unobserved := byName["kan-2"]
	if unobserved.Planned != nil || unobserved.Current != nil {
		t.Errorf("kan-2 PlannedTasks/CurrentTasks = %v/%v, want nil/nil",
			unobserved.Planned, unobserved.Current)
	}
}

// LiveStateByName keys a board row's two task-count figures by change name
// inside TestLiveStateBoardCarriesPlannedAndCurrentTasks -- the board's own
// row order is by updated_at, which two same-second seeds do not
// distinguish.
type LiveStateByName struct {
	Planned *int
	Current *int
}

// TestReviewersScopesByDispatchStart pins that the view attributes a
// dispatch to the period containing ITS OWN started_at and never requires
// a stage_run_id: nothing that records a dispatch (flow record dispatch
// begin) sets that column, so a join through stage_runs left the live view
// empty from its first day.
func TestReviewersScopesByDispatchStart(t *testing.T) {
	st, _ := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-reviewers-nostage-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	record := func(role, slot string, startedAt time.Time) {
		t.Helper()
		d := baseDispatch(role, "sonnet")
		d.Slot = slot
		d.StartedAt = startedAt
		if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", d); err != nil {
			t.Fatalf("RecordDispatch: %v", err)
		}
	}
	record("reviewer", "primary", time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC))
	record("reviewer", "primary", time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)) // at the half-open end: excluded
	// A panel-fix dispatch a session tagged with a slot of its own: no roster
	// names "fix", so it is not a reviewer and never a row.
	record("panel-fix", "fix", time.Date(2026, 6, 10, 10, 0, 0, 0, time.UTC))
	seedRoster(t, st, projectKey, "kan-1", "primary")

	rows, err := st.Reviewers(ctx, store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil {
		t.Fatalf("Reviewers: %v", err)
	}
	if len(rows) != 1 || rows[0].Slot != "primary" || rows[0].Dispatches != 1 {
		t.Fatalf("rows = %+v, want one primary row with 1 dispatch", rows)
	}
}

// seedRoster records one decision whose panel roster names slots, so the
// reviewers view -- which shows a slot only when some roster names it --
// has a roster to consult.
func seedRoster(t *testing.T, st *store.Store, projectKey, change string, slots ...string) {
	t.Helper()
	entries := make([]string, len(slots))
	for i, s := range slots {
		entries[i] = fmt.Sprintf(`{"slot": %q}`, s)
	}
	if _, _, err := st.RecordDecision(context.Background(), projectKey, change, records.Decision{
		SessionToken: "mf-roster-" + projectKey,
		Decision:     json.RawMessage(`{"panel": {"roster": [` + strings.Join(entries, ",") + `]}}`),
	}); err != nil {
		t.Fatalf("RecordDecision: %v", err)
	}
}

// TestLiveStateBoardCarriesWholeLifeRunTotals: each board row carries its
// change's run totals over the change's whole life (design.md's
// board-run-totals) -- a run is the stage rows sharing a session token, a
// token-less row is a run of its own, and a run started long before the
// board's period still counts.
func TestLiveStateBoardCarriesWholeLifeRunTotals(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-runs-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	stage := func(token *string, started time.Time, minutes int, metrics string) {
		t.Helper()
		in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
		in.SessionToken = token
		in.StartedAt = started
		run, err := st.BeginStage(ctx, in)
		if err != nil {
			t.Fatalf("BeginStage: %v", err)
		}
		if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(metrics)); err != nil {
			t.Fatalf("MergeMetrics: %v", err)
		}
		if err := st.EndStage(ctx, run.ID, started.Add(time.Duration(minutes)*time.Minute), "completed"); err != nil {
			t.Fatalf("EndStage: %v", err)
		}
	}
	day := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	stage(ptr("mf-a"), day, 1, `{"cost_usd":1.0,"signals":{"main":{"tool_errors":2}}}`)
	stage(ptr("mf-a"), day.Add(5*time.Minute), 5, `{"cost_usd":0.5,"signals":{"sidechain":{"api_errors":1}}}`)
	stage(nil, day.Add(-72*time.Hour), 30, `{"cost_usd":2.0,"signals":{"main":{"api_errors":1}}}`)

	rows, err := st.LiveStateBoard(ctx, store.Period{
		From: time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil {
		t.Fatalf("LiveStateBoard: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("LiveStateBoard returned %d rows, want 1", len(rows))
	}
	got := rows[0]
	if got.RunCount != 2 {
		t.Errorf("RunCount = %d, want 2 (one token run, one token-less run)", got.RunCount)
	}
	if diff := got.CostUsd - 3.5; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("CostUsd = %v, want 3.5", got.CostUsd)
	}
	if want := int64(40 * time.Minute / time.Millisecond); got.WallClockMs != want {
		t.Errorf("WallClockMs = %d, want %d (10 min token run + 30 min token-less run)", got.WallClockMs, want)
	}
	if got.Errors != 4 {
		t.Errorf("Errors = %d, want 4 (tool and API errors, main and sidechain)", got.Errors)
	}
}

// TestLiveStateBoardCarriesOpenStageDoneTasksAndActivity pins the three
// live facts the board's Running now cards read: the change's open stage
// run, how many tasks carry a committed implementer dispatch, and what the
// open run's session is doing. A change with no open run and no
// implementer dispatch carries none of them.
func TestLiveStateBoardCarriesOpenStageDoneTasksAndActivity(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-live-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")

	day := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	closed := baseBeginInput(projectKey, "kan-1", "/flow", "flow.load-context")
	closed.StartedAt = day
	run, err := st.BeginStage(ctx, closed)
	if err != nil {
		t.Fatalf("BeginStage(closed): %v", err)
	}
	if err := st.EndStage(ctx, run.ID, day.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	open := baseBeginInput(projectKey, "kan-1", "/flow", "flow.sdd-tdd")
	open.SessionID = ptr("s-1")
	open.StartedAt = day.Add(2 * time.Minute)
	if _, err := st.BeginStage(ctx, open); err != nil {
		t.Fatalf("BeginStage(open): %v", err)
	}

	dispatch := func(task, commit string) {
		t.Helper()
		d := baseDispatch("implementer", "opus")
		d.TaskID = task
		d.CommitSHA = commit
		if task == "3" {
			d.AgentID = "inline"
		}
		if _, err := st.RecordDispatch(ctx, projectKey, "kan-1", d); err != nil {
			t.Fatalf("RecordDispatch: %v", err)
		}
	}
	dispatch("1", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	dispatch("2", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	dispatch("3", "")

	if err := st.SetSessionActivity(ctx, "s-1", "question", "Proceed?"); err != nil {
		t.Fatalf("SetSessionActivity: %v", err)
	}

	rows, err := st.LiveStateBoard(ctx, store.Period{
		From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil {
		t.Fatalf("LiveStateBoard: %v", err)
	}
	byName := map[string]store.LiveStateRow{}
	for _, row := range rows {
		byName[row.Name] = row
	}

	live := byName["kan-1"]
	if live.CurrentStage == nil || *live.CurrentStage != "flow.sdd-tdd" {
		t.Errorf("CurrentStage = %v, want flow.sdd-tdd", live.CurrentStage)
	}
	if live.CurrentStageStartedAt == nil || !live.CurrentStageStartedAt.Equal(day.Add(2*time.Minute)) {
		t.Errorf("CurrentStageStartedAt = %v, want %v", live.CurrentStageStartedAt, day.Add(2*time.Minute))
	}
	if live.LastStage == nil || *live.LastStage != "flow.sdd-tdd" || live.LastStageEndedAt != nil {
		t.Errorf("LastStage = %v ended %v, want the open flow.sdd-tdd, not ended", live.LastStage, live.LastStageEndedAt)
	}
	if !slices.Equal(live.CommittedTasks, []string{"1", "2"}) {
		t.Errorf("CommittedTasks = %v, want [1 2]", live.CommittedTasks)
	}
	if len(live.Subagents) != 3 || live.Subagents[0].Role != "implementer" || live.Subagents[0].Model != "opus" ||
		live.Subagents[2].TaskID == nil || *live.Subagents[2].TaskID != "3" {
		t.Errorf("Subagents = %+v, want the three unended implementer dispatches, task 3 last", live.Subagents)
	}
	if len(live.Subagents) == 3 && (live.Subagents[0].Inline || !live.Subagents[2].Inline) {
		t.Errorf("Subagents inline = %v/%v, want only task 3's main-session dispatch inline", live.Subagents[0].Inline, live.Subagents[2].Inline)
	}
	if live.OpenStageDispatches != 3 {
		t.Errorf("OpenStageDispatches = %d, want the 3 dispatches begun in the open stage", live.OpenStageDispatches)
	}
	if live.DoneTasks == nil || *live.DoneTasks != 2 {
		t.Errorf("DoneTasks = %v, want 2", live.DoneTasks)
	}
	if live.Activity == nil || *live.Activity != "question" {
		t.Errorf("Activity = %v, want question", live.Activity)
	}
	if live.ActivityDetail == nil || *live.ActivityDetail != "Proceed?" {
		t.Errorf("ActivityDetail = %v, want Proceed?", live.ActivityDetail)
	}
	if live.ActivitySince == nil {
		t.Errorf("ActivitySince = nil, want the activity's time")
	}

	idle := byName["kan-2"]
	if len(idle.Subagents) != 0 || len(idle.CommittedTasks) != 0 {
		t.Errorf("kan-2 Subagents = %v CommittedTasks = %v, want none", idle.Subagents, idle.CommittedTasks)
	}
	if idle.LastStage != nil || idle.LastStageEndedAt != nil {
		t.Errorf("kan-2 LastStage = %v, want nil (no stage run)", idle.LastStage)
	}
	if idle.CurrentStage != nil || idle.CurrentStageStartedAt != nil || idle.DoneTasks != nil ||
		idle.Activity != nil || idle.ActivityDetail != nil || idle.ActivitySince != nil {
		t.Errorf("kan-2 = %+v, want every live field nil", idle)
	}
}

// TestLiveStateBoardActivityBetweenStages pins that a change with no open
// stage run still carries its latest run's session activity, so a main
// session working between stages reads as working.
func TestLiveStateBoardActivityBetweenStages(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-between-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	day := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	in := baseBeginInput(projectKey, "kan-1", "/flow", "flow.visual-verify")
	in.SessionID = ptr("s-between")
	in.StartedAt = day
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.EndStage(ctx, run.ID, day.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}
	if err := st.SetSessionActivity(ctx, "s-between", "working", ""); err != nil {
		t.Fatalf("SetSessionActivity: %v", err)
	}
	rows, err := st.LiveStateBoard(ctx, store.Period{
		From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil || len(rows) != 1 {
		t.Fatalf("LiveStateBoard = %v, %v; want one row", rows, err)
	}
	if rows[0].CurrentStage != nil || rows[0].Activity == nil || *rows[0].Activity != "working" {
		t.Errorf("CurrentStage = %v Activity = %v, want no open stage and working", rows[0].CurrentStage, rows[0].Activity)
	}
	if rows[0].SessionSeenAt != nil {
		t.Errorf("SessionSeenAt = %v before any transcript was harvested, want nil", rows[0].SessionSeenAt)
	}
	// A harvested subagent transcript of the session marks it seen.
	if _, err := st.CommitHarvestBatch(ctx, "/p/s-between/subagents/agent-1.jsonl", 0, false, 10, nil); err != nil {
		t.Fatalf("CommitHarvestBatch: %v", err)
	}
	rows, err = st.LiveStateBoard(ctx, store.Period{
		From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, nil)
	if err != nil || len(rows) != 1 || rows[0].SessionSeenAt == nil {
		t.Fatalf("after harvest: rows = %+v err = %v, want SessionSeenAt set", rows, err)
	}
}

// TestLiveStateBoardLiveFieldEdges pins the edges the board's live fields
// turn on: the newest of two open runs is the one shown, a repeated report of
// the same activity keeps its since while a new state moves it, the period's
// end is exclusive, and "change" scopes the board to one change.
func TestLiveStateBoardLiveFieldEdges(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-edges-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	seedChange(t, st, projectKey, "kan-2")

	day := time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
	for i, stage := range []string{"flow.sdd-tdd", "flow.review-panel"} {
		in := baseBeginInput(projectKey, "kan-1", "/flow", stage)
		in.SessionID = ptr(fmt.Sprintf("edge-%s-%d", projectKey, i))
		in.StartedAt = day.Add(time.Duration(i) * time.Minute)
		if _, err := st.BeginStage(ctx, in); err != nil {
			t.Fatalf("BeginStage(%s): %v", stage, err)
		}
	}
	session := fmt.Sprintf("edge-%s-1", projectKey)
	all := store.Period{From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)}
	board := func(period store.Period, change *string) map[string]store.LiveStateRow {
		t.Helper()
		rows, err := st.LiveStateBoard(ctx, period, &projectKey, change)
		if err != nil {
			t.Fatalf("LiveStateBoard: %v", err)
		}
		out := map[string]store.LiveStateRow{}
		for _, r := range rows {
			out[r.Name] = r
		}
		return out
	}
	since := func() time.Time {
		t.Helper()
		r := board(all, nil)["kan-1"]
		if r.ActivitySince == nil {
			t.Fatalf("ActivitySince = nil")
		}
		return *r.ActivitySince
	}

	if got := board(all, nil)["kan-1"].CurrentStage; got == nil || *got != "flow.review-panel" {
		t.Errorf("CurrentStage = %v, want the newer open run flow.review-panel", got)
	}

	if err := st.SetSessionActivity(ctx, session, "working", "Bash"); err != nil {
		t.Fatalf("SetSessionActivity: %v", err)
	}
	first := since()
	time.Sleep(20 * time.Millisecond)
	if err := st.SetSessionActivity(ctx, session, "working", "Edit"); err != nil {
		t.Fatalf("SetSessionActivity: %v", err)
	}
	if got := since(); !got.Equal(first) {
		t.Errorf("ActivitySince after a repeated working = %v, want unchanged %v", got, first)
	}
	time.Sleep(20 * time.Millisecond)
	if err := st.SetSessionActivity(ctx, session, "question", "Proceed?"); err != nil {
		t.Fatalf("SetSessionActivity: %v", err)
	}
	if got := since(); !got.After(first) {
		t.Errorf("ActivitySince after a new state = %v, want after %v", got, first)
	}

	updated := board(all, nil)["kan-1"].UpdatedAt
	if _, ok := board(store.Period{From: all.From, To: updated}, nil)["kan-1"]; ok {
		t.Errorf("a change updated exactly at the period's end is on the board; the end is exclusive")
	}
	if _, ok := board(store.Period{From: all.From, To: updated.Add(time.Microsecond)}, nil)["kan-1"]; !ok {
		t.Errorf("a change updated just before the period's end is missing")
	}

	one := board(all, ptr("kan-1"))
	if _, ok := one["kan-1"]; !ok || len(one) != 1 {
		t.Errorf("board scoped to kan-1 = %d rows, want kan-1 alone", len(one))
	}
}

// TestLiveStateBoardCarriesSelfReviewRoots pins that a change's board row
// names the worktrees of its self-review sub-run, self-review-<name>, even
// when that sub-run lives in another project.
func TestLiveStateBoardCarriesSelfReviewRoots(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-board-srroots-%d", time.Now().UnixNano())
	name := fmt.Sprintf("kan-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, name)
	sub := baseChange(projectKey+"-tooling", "self-review-"+name)
	sub.Repos = []store.Repo{{RepoRoot: "/tmp/sr-wt"}}
	if err := st.PutChange(ctx, sub); err != nil {
		t.Fatalf("PutChange(sub): %v", err)
	}
	rows, err := st.LiveStateBoard(ctx, store.Period{
		From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
	}, &projectKey, &name)
	if err != nil {
		t.Fatalf("LiveStateBoard: %v", err)
	}
	// The sub-run's worktree, then its project's main checkout, which keeps the archived plan.
	if len(rows) != 1 || len(rows[0].SelfReviewRoots) != 2 || rows[0].SelfReviewRoots[0] != "/tmp/sr-wt" || !strings.HasSuffix(rows[0].SelfReviewRoots[1], projectKey+"-tooling") {
		t.Fatalf("rows = %+v, want SelfReviewRoots [/tmp/sr-wt <tooling main checkout>]", rows)
	}
	if n := len(rows[0].RepoRoots); n == 0 || !strings.HasSuffix(rows[0].RepoRoots[n-1], projectKey) {
		t.Errorf("RepoRoots = %v, want the project's main checkout last", rows[0].RepoRoots)
	}
}
