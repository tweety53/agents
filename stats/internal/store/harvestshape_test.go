package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/harvest"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestAggregationsReadTheHarvesterSMetricsShape is task 24's seam test. It
// builds its fixture by marshalling internal/harvest's own MetricsPatch
// value -- never a hand-written JSON literal standing in for it -- commits
// it through the ordinary store path (MergeMetrics, exactly as Watcher
// does), and asserts every aggregation that reads the metrics bag's
// "tokens" key returns real numbers over it, instead of erroring outright
// or reading NULL.
//
// Sourcing the fixture from internal/harvest is the entire point: a
// hand-written literal that happens to match today's MetricsPatch shape is
// a second, private copy of the contract between the two packages, and
// this task's own defect ("How it survived twenty-one review passes",
// tasks.md task 24) is exactly what happens when the two copies drift. A
// future change to MetricsPatch that breaks an aggregation must fail here,
// in this repository, rather than silently in a view nobody queried.
func TestAggregationsReadTheHarvesterSMetricsShape(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-harvestshape-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1") // baseChange sets ReviewPanelRoster "light"

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	in.StartedAt = time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// The exact shape a harvest batch commits -- internal/harvest's own
	// type, encoded by its own json tags, not a literal standing in for it.
	patch := harvest.MetricsPatch{
		Tokens: harvest.TokenDelta{
			Main:      harvest.Bucket{Input: 600000, Output: 300000, CacheCreation: 50000, CacheRead: 950000, Thinking: 10000},
			Sidechain: harvest.Bucket{Input: 400000},
		},
	}
	patchJSON, err := json.Marshal(patch)
	if err != nil {
		t.Fatalf("marshal MetricsPatch: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, patchJSON); err != nil {
		t.Fatalf("MergeMetrics(harvester-shaped patch): %v", err)
	}
	if err := st.EndStage(ctx, run.ID, in.StartedAt.Add(time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	period := store.Period{
		From: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
	}

	// Every expectation below is derived from patch itself, not
	// re-typed as a second set of magic numbers -- the same discipline
	// this test exists to enforce one layer up.
	wantMain := patch.Tokens.Main.Input + patch.Tokens.Main.Output + patch.Tokens.Main.CacheCreation + patch.Tokens.Main.CacheRead + patch.Tokens.Main.Thinking
	wantSidechain := patch.Tokens.Sidechain.Input + patch.Tokens.Sidechain.Output + patch.Tokens.Sidechain.CacheCreation + patch.Tokens.Sidechain.CacheRead + patch.Tokens.Sidechain.Thinking
	wantTotalInput := patch.Tokens.Main.Input + patch.Tokens.Sidechain.Input

	costRows, err := st.CostPerChange(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CostPerChange: %v (must not error against a run whose metrics carry the harvester's real, nested tokens shape)", err)
	}
	if len(costRows) != 1 {
		t.Fatalf("CostPerChange returned %d groups, want 1", len(costRows))
	}
	cost := costRows[0]
	if cost.TotalTokensInput == nil || *cost.TotalTokensInput != wantTotalInput {
		t.Errorf("TotalTokensInput = %v, want %v (main.input + sidechain.input)", cost.TotalTokensInput, wantTotalInput)
	}
	if cost.MainTokens == nil || *cost.MainTokens != wantMain {
		t.Errorf("MainTokens = %v, want %v (main bucket's own fields summed)", cost.MainTokens, wantMain)
	}
	if cost.SidechainTokens == nil || *cost.SidechainTokens != wantSidechain {
		t.Errorf("SidechainTokens = %v, want %v (sidechain bucket's own fields summed)", cost.SidechainTokens, wantSidechain)
	}

	cacheRows, err := st.CacheEfficiency(ctx, period, &projectKey, nil)
	if err != nil {
		t.Fatalf("CacheEfficiency: %v", err)
	}
	if len(cacheRows) != 1 {
		t.Fatalf("CacheEfficiency returned %d rows, want 1", len(cacheRows))
	}
	wantCacheRead := patch.Tokens.Main.CacheRead + patch.Tokens.Sidechain.CacheRead
	wantCacheCreation := patch.Tokens.Main.CacheCreation + patch.Tokens.Sidechain.CacheCreation
	cache := cacheRows[0]
	if cache.CacheReadTotal == nil || *cache.CacheReadTotal != wantCacheRead {
		t.Errorf("CacheReadTotal = %v, want %v", cache.CacheReadTotal, wantCacheRead)
	}
	if cache.CacheCreationTotal == nil || *cache.CacheCreationTotal != wantCacheCreation {
		t.Errorf("CacheCreationTotal = %v, want %v", cache.CacheCreationTotal, wantCacheCreation)
	}
	if cache.Ratio == nil {
		t.Errorf("CacheEfficiency Ratio is nil, want a computed ratio over real cache totals")
	}
}
