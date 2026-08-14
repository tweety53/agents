package store_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// seedChange puts a minimal, valid change under projectKey/name and returns
// nothing -- callers key every stage run call by that same project/name
// pair, exactly as BeginStage requires.
func seedChange(t *testing.T, st *store.Store, projectKey, name string) {
	t.Helper()
	ctx := context.Background()
	c := baseChange(projectKey, name)
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("seed change %s/%s: %v", projectKey, name, err)
	}
}

func baseBeginInput(projectKey, name, command, stage string) store.BeginStageInput {
	return store.BeginStageInput{
		ProjectKey: projectKey,
		ChangeName: name,
		Harness:    "claude-code",
		SessionID:  ptr("session-1"),
		Command:    command,
		Stage:      stage,
		StartedAt:  time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC),
	}
}

func TestBeginStageAllocatesAttempts(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-attempts-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")

	first, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("first BeginStage: %v", err)
	}
	if first.Attempt != 1 {
		t.Errorf("first BeginStage attempt = %d, want 1", first.Attempt)
	}

	second, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("second BeginStage: %v", err)
	}
	if second.Attempt != 2 {
		t.Errorf("second BeginStage attempt = %d, want 2", second.Attempt)
	}
	if second.ID == first.ID {
		t.Errorf("second BeginStage returned the same row id as the first: %d", second.ID)
	}

	// A different stage of the same command starts its own attempt series.
	otherStage := in
	otherStage.Stage = "review panel"
	third, err := st.BeginStage(ctx, otherStage)
	if err != nil {
		t.Fatalf("BeginStage for a different stage: %v", err)
	}
	if third.Attempt != 1 {
		t.Errorf("BeginStage attempt for a different stage = %d, want 1 (its own series)", third.Attempt)
	}
}

func TestBeginStageUnknownChangeIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	_, err := st.BeginStage(ctx, baseBeginInput("no-such-project", "no-such-change", "/myflow-do", "SDD + TDD per task"))
	if !errors.Is(err, store.ErrChangeNotFound) {
		t.Fatalf("BeginStage(unknown change) error = %v, want errors.Is(_, store.ErrChangeNotFound)", err)
	}
}

// TestConcurrentBeginStageDoesNotCollide reproduces the failure a
// check-then-insert attempt allocator would have: many goroutines call
// BeginStage for the exact same (change, command, stage) triple at once.
// If the next attempt were computed by a separate read followed by a
// gated insert, two goroutines could both read the same current maximum
// and insert the same attempt number, tripping the unique constraint for
// one of them (a bug) or -- worse, if the code swallowed that error --
// silently losing one attempt's row. BeginStage instead allocates the
// attempt inside the insert and retries on the constraint's own
// unique-violation, so every goroutine must come away with an attempt
// number, and those numbers must be exactly {1..N} with no duplicate and
// no gap.
func TestConcurrentBeginStageDoesNotCollide(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-concurrent-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	const writers = 20
	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")

	var wg sync.WaitGroup
	attempts := make([]int, writers)
	errs := make([]error, writers)

	for i := range writers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			run, err := st.BeginStage(callCtx, in)
			errs[i] = err
			if err == nil {
				attempts[i] = run.Attempt
			}
		}(i)
	}
	wg.Wait()

	seen := make(map[int]int)
	for i, err := range errs {
		if err != nil {
			t.Fatalf("writer %d: BeginStage: %v", i, err)
		}
		seen[attempts[i]]++
	}
	for attempt := 1; attempt <= writers; attempt++ {
		if seen[attempt] != 1 {
			t.Errorf("attempt %d was allocated %d times, want exactly 1", attempt, seen[attempt])
		}
	}
	if len(seen) != writers {
		t.Errorf("got %d distinct attempt numbers, want %d", len(seen), writers)
	}
}

func TestMergeMetricsPreservesOtherKeys(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-merge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// The harvester writes token keys first.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":100,"output":50}}}`)); err != nil {
		t.Fatalf("MergeMetrics (tokens): %v", err)
	}
	// A stage-end mark writes outcome keys afterward.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"duration_ms":4200,"fast_mode":false}`)); err != nil {
		t.Fatalf("MergeMetrics (duration): %v", err)
	}
	// A later merge overwrites a key it shares, and leaves the rest alone.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":150,"output":50,"cache_read":10}}}`)); err != nil {
		t.Fatalf("MergeMetrics (updated tokens): %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}

	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}

	if _, ok := bag["duration_ms"]; !ok {
		t.Errorf("metrics lost duration_ms after a later merge touched only tokens: %s", got.Metrics)
	}
	if _, ok := bag["fast_mode"]; !ok {
		t.Errorf("metrics lost fast_mode after a later merge touched only tokens: %s", got.Metrics)
	}

	var bagTokens struct {
		Main struct {
			Input     float64 `json:"input"`
			Output    float64 `json:"output"`
			CacheRead float64 `json:"cache_read"`
		} `json:"main"`
	}
	if err := json.Unmarshal(bag["tokens"], &bagTokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	if bagTokens.Main.Input != 150 {
		t.Errorf("tokens.main.input = %v, want 150 (last write wins on a shared key)", bagTokens.Main.Input)
	}
	if bagTokens.Main.CacheRead != 10 {
		t.Errorf("tokens.main.cache_read = %v, want 10", bagTokens.Main.CacheRead)
	}
}

// TestMergeMetricsDeepMergesNestedSiblings reproduces the finding against
// the shallow `metrics || patch` implementation: two writers each touch a
// different sub-key of the same nested "tokens" object -- one the
// harvester's main-thread token counts, the other its sidechain counts,
// exactly as two separate CommitHarvestBatch calls would. A shallow
// top-level concatenation replaces "tokens" wholesale on the second write,
// discarding the first writer's main bucket. The merge must combine the
// nested object's keys instead, so neither writer needs to know what the
// other already stored there.
func TestMergeMetricsDeepMergesNestedSiblings(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-deepmerge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":5,"output":9}}}`)); err != nil {
		t.Fatalf("MergeMetrics (main): %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"sidechain":{"input":0}}}`)); err != nil {
		t.Fatalf("MergeMetrics (sidechain): %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}

	var bag struct {
		Tokens struct {
			Main struct {
				Input  *float64 `json:"input"`
				Output *float64 `json:"output"`
			} `json:"main"`
			Sidechain struct {
				Input *float64 `json:"input"`
			} `json:"sidechain"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}

	if bag.Tokens.Main.Input == nil || *bag.Tokens.Main.Input != 5 {
		t.Errorf("tokens.main.input = %v, want 5 -- a later write to tokens.sidechain must not erase it", bag.Tokens.Main.Input)
	}
	if bag.Tokens.Main.Output == nil || *bag.Tokens.Main.Output != 9 {
		t.Errorf("tokens.main.output = %v, want 9 -- a later write to tokens.sidechain must not erase it", bag.Tokens.Main.Output)
	}
	if bag.Tokens.Sidechain.Input == nil || *bag.Tokens.Sidechain.Input != 0 {
		t.Errorf("tokens.sidechain.input = %v, want 0 (a recorded zero, not absence)", bag.Tokens.Sidechain.Input)
	}
}

// TestMergeMetricsNonObjectValueReplaces asserts the other half of the
// merge rule: when patch's value at a key is not itself an object, it
// replaces whatever was stored there -- including an object -- rather than
// being merged into it. There is nothing under a non-object value to
// preserve.
func TestMergeMetricsNonObjectValueReplaces(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nonobjreplace-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":{"main":{"input":5,"output":9}}}`)); err != nil {
		t.Fatalf("MergeMetrics (object): %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"tokens":false}`)); err != nil {
		t.Fatalf("MergeMetrics (non-object replaces): %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if string(bag["tokens"]) != "false" {
		t.Errorf(`tokens = %s, want "false" -- a non-object patch value must replace the stored object entirely`, bag["tokens"])
	}
}

// TestMergeMetricsRejectsNilPatch asserts that a nil patch is refused with
// a typed error before it ever reaches SQL. A nil json.RawMessage
// marshals to SQL NULL, and jsonb_deep_merge(metrics, NULL) returns NULL,
// which stage_runs.metrics' NOT NULL constraint would then reject as a raw
// Postgres error -- MergeMetrics catches this in Go instead.
func TestMergeMetricsRejectsNilPatch(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nilpatch-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	err = st.MergeMetrics(ctx, run.ID, nil)
	if !errors.Is(err, store.ErrNilMetricsPatch) {
		t.Fatalf("MergeMetrics(nil patch) error = %v, want errors.Is(_, store.ErrNilMetricsPatch)", err)
	}

	// The stage run's metrics must be untouched, not corrupted or nulled.
	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if string(got.Metrics) != "{}" {
		t.Errorf("metrics = %s after a rejected nil patch, want unchanged {}", got.Metrics)
	}
}

func TestMergeMetricsUnknownStageRunIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	err := st.MergeMetrics(ctx, 99999999, json.RawMessage(`{"a":1}`))
	if !errors.Is(err, store.ErrStageRunNotFound) {
		t.Fatalf("MergeMetrics(unknown id) error = %v, want errors.Is(_, store.ErrStageRunNotFound)", err)
	}
}

func TestEndStageRecordsOutcome(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-end-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	endedAt := run.StartedAt.Add(5 * time.Minute)
	if err := st.EndStage(ctx, run.ID, endedAt, "completed"); err != nil {
		t.Fatalf("EndStage: %v", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	if got.EndedAt == nil || !got.EndedAt.Equal(endedAt) {
		t.Errorf("EndedAt = %v, want %v", got.EndedAt, endedAt)
	}
	if got.Outcome == nil || *got.Outcome != "completed" {
		t.Errorf("Outcome = %v, want completed", got.Outcome)
	}
}

func TestEndStageUnknownStageRunIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	err := st.EndStage(ctx, 99999999, time.Now(), "completed")
	if !errors.Is(err, store.ErrStageRunNotFound) {
		t.Fatalf("EndStage(unknown id) error = %v, want errors.Is(_, store.ErrStageRunNotFound)", err)
	}
}

// TestSweepAbandonedClosesSilentStages asserts that SweepAbandoned closes
// only stage runs that are both open (no end mark) and started before the
// silence cutoff -- never a stage run that already ended, and never one
// that started after the cutoff.
func TestSweepAbandonedClosesSilentStages(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-sweep-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	silent := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	silent.StartedAt = time.Now().Add(-2 * time.Hour)
	silentRun, err := st.BeginStage(ctx, silent)
	if err != nil {
		t.Fatalf("BeginStage(silent): %v", err)
	}

	ended := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	ended.StartedAt = time.Now().Add(-2 * time.Hour)
	endedRun, err := st.BeginStage(ctx, ended)
	if err != nil {
		t.Fatalf("BeginStage(ended): %v", err)
	}
	if err := st.EndStage(ctx, endedRun.ID, time.Now().Add(-90*time.Minute), "completed"); err != nil {
		t.Fatalf("EndStage(ended): %v", err)
	}

	recent := baseBeginInput(projectKey, "kan-1", "/myflow-do", "finish")
	recent.StartedAt = time.Now().Add(-1 * time.Minute)
	recentRun, err := st.BeginStage(ctx, recent)
	if err != nil {
		t.Fatalf("BeginStage(recent): %v", err)
	}

	cutoff := time.Now().Add(-1 * time.Hour)
	n, err := st.SweepAbandoned(ctx, cutoff)
	if err != nil {
		t.Fatalf("SweepAbandoned: %v", err)
	}
	if n != 1 {
		t.Fatalf("SweepAbandoned closed %d rows, want 1 (only the silent, still-open one)", n)
	}

	got, err := st.GetStageRun(ctx, silentRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(silent): %v", err)
	}
	if got.Outcome == nil || *got.Outcome != "abandoned" {
		t.Errorf("silent stage run outcome = %v, want abandoned", got.Outcome)
	}
	if got.EndedAt == nil {
		t.Errorf("silent stage run has no EndedAt after being swept")
	}

	gotEnded, err := st.GetStageRun(ctx, endedRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(ended): %v", err)
	}
	if gotEnded.Outcome == nil || *gotEnded.Outcome != "completed" {
		t.Errorf("already-ended stage run outcome changed to %v, want it left as completed", gotEnded.Outcome)
	}

	gotRecent, err := st.GetStageRun(ctx, recentRun.ID)
	if err != nil {
		t.Fatalf("GetStageRun(recent): %v", err)
	}
	if gotRecent.Outcome != nil {
		t.Errorf("recent, still-open stage run outcome = %v, want nil (not swept: started after cutoff)", gotRecent.Outcome)
	}
}

func TestPriceFreezesCostAndVersion(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.StartedAt = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	// cache_creation_1h (not the collapsed "cache_creation") is what Price
	// now reads: this run's cache-creation usage is entirely a 1-hour
	// write, which task 23 exists to price at its own rate rather than
	// the 5-minute one. "cache_creation" itself is still present,
	// unchanged in meaning, purely to prove Price leaves it untouched.
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"input": 1000000, "output": 500000, "cache_creation": 200000, "cache_creation_1h": 200000, "cache_read": 4000000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics(tokens): %v", err)
	}

	oldRate := store.PricingRate{
		Model:               "claude-opus-5",
		EffectiveFrom:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok:        3,
		OutputPerMTok:       15,
		CacheWritePerMTok:   3.75,
		CacheWrite5mPerMTok: 3.75,
		CacheWrite1hPerMTok: ptr(3.75),
		CacheReadPerMTok:    0.3,
	}
	if err := st.PutPricing(ctx, oldRate); err != nil {
		t.Fatalf("PutPricing(old): %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD        float64 `json:"cost_usd"`
			PricingVersion string  `json:"pricing_version"`
			Tokens         struct {
				Main struct {
					Input         float64 `json:"input"`
					Output        float64 `json:"output"`
					CacheCreation float64 `json:"cache_creation"`
					CacheRead     float64 `json:"cache_read"`
				} `json:"main"`
			} `json:"tokens"`
		} `json:"models"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal priced metrics: %v", err)
	}

	wantCost := 1*3 + 0.5*15 + 0.2*3.75 + 4*0.3
	opus, ok := bag.Models["claude-opus-5"]
	if !ok {
		t.Fatalf("models.claude-opus-5 missing from priced metrics: %s", priced.Metrics)
	}
	if diff := opus.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want %v", opus.CostUSD, wantCost)
	}
	wantVersion := "claude-opus-5@2026-01-01T00:00:00Z"
	if opus.PricingVersion != wantVersion {
		t.Errorf("models.claude-opus-5.pricing_version = %q, want %q", opus.PricingVersion, wantVersion)
	}
	// The single-model run's whole cost is this one bucket's cost, so the
	// top-level total (their sum) matches it exactly.
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v", bag.CostUSD, wantCost)
	}
	// Token counts survive pricing untouched, so history can be re-priced.
	if opus.Tokens.Main.Input != 1000000 || opus.Tokens.Main.Output != 500000 ||
		opus.Tokens.Main.CacheCreation != 200000 || opus.Tokens.Main.CacheRead != 4000000 {
		t.Errorf("token counts changed after Price: %+v", opus.Tokens)
	}

	// A later, higher rate takes effect for a subsequent Price call, but a
	// figure already frozen and never re-priced does not move under the
	// reader: re-fetching the row without calling Price again must not
	// reflect the new rate.
	newRate := oldRate
	newRate.EffectiveFrom = time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	newRate.InputPerMTok = 30
	if err := st.PutPricing(ctx, newRate); err != nil {
		t.Fatalf("PutPricing(new): %v", err)
	}

	stillOld, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun after new pricing published: %v", err)
	}
	var stillBag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(stillOld.Metrics, &stillBag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if diff := stillBag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd moved to %v after a newer pricing row was published without re-pricing; want it to stay frozen at %v", stillBag.CostUSD, wantCost)
	}
}

// TestPriceTwoModelsAgainstTwoRates covers step 3's central case: a
// mixed-model run (the review panel's own parent-vs-reviewer split) is
// priced per bucket against each model's own rate, and the top-level
// cost_usd is their sum -- a mixed run's real total, not one model's
// rate applied to every token.
func TestPriceTwoModelsAgainstTwoRates(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-twomodel-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel")
	in.StartedAt = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	run, err := st.BeginStage(ctx, in)
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5":   {"tokens": {"main": {"input": 1000000}}},
			"claude-sonnet-5": {"tokens": {"sidechain": {"input": 2000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing(opus): %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 1, OutputPerMTok: 5, CacheWritePerMTok: 1.25, CacheReadPerMTok: 0.1,
	}); err != nil {
		t.Fatalf("PutPricing(sonnet): %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
		Models  map[string]struct {
			CostUSD float64 `json:"cost_usd"`
		} `json:"models"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	const eps = 1e-9
	wantOpus := 1.0 * 3
	wantSonnet := 2.0 * 1
	if diff := bag.Models["claude-opus-5"].CostUSD - wantOpus; diff > eps || diff < -eps {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want %v", bag.Models["claude-opus-5"].CostUSD, wantOpus)
	}
	if diff := bag.Models["claude-sonnet-5"].CostUSD - wantSonnet; diff > eps || diff < -eps {
		t.Errorf("models.claude-sonnet-5.cost_usd = %v, want %v", bag.Models["claude-sonnet-5"].CostUSD, wantSonnet)
	}
	wantTotal := wantOpus + wantSonnet
	if diff := bag.CostUSD - wantTotal; diff > eps || diff < -eps {
		t.Errorf("cost_usd = %v, want %v (the sum of both models' own bucket cost, not one rate applied to every token)", bag.CostUSD, wantTotal)
	}
}

// TestPriceOneUnpriceableBucketOmitsTopLevelTotal is step 3's other
// central case: when one of a mixed run's model buckets has no rate in
// effect, Price writes the bucket it could price, omits the one it could
// not, and omits the top-level cost_usd entirely rather than writing a
// partial sum that reads like a complete one -- a partial total is
// indistinguishable from a correct one at every layer above it, which
// makes it the most dangerous possible output here.
func TestPriceOneUnpriceableBucketOmitsTopLevelTotal(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-partial-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "review panel"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {"tokens": {"main": {"input": 1000000}}},
			"no-such-model": {"tokens": {"sidechain": {"input": 500000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(one bucket unpriceable) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd key at all when one model bucket could not be priced", bag["cost_usd"])
	}

	var models map[string]json.RawMessage
	if err := json.Unmarshal(bag["models"], &models); err != nil {
		t.Fatalf("unmarshal models: %v", err)
	}
	var opus struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["claude-opus-5"], &opus); err != nil {
		t.Fatalf("unmarshal models.claude-opus-5: %v", err)
	}
	if opus.CostUSD != 3 {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want 3 (the bucket that could be priced is still written)", opus.CostUSD)
	}
	var noSuch struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["no-such-model"], &noSuch); err != nil {
		t.Fatalf("unmarshal models.no-such-model: %v", err)
	}
	if noSuch.CostUSD != nil {
		t.Errorf("models.no-such-model.cost_usd = %v, want absent (no rate was ever in effect for it)", *noSuch.CostUSD)
	}
}

func TestPriceWithoutTokensIsUnavailable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-unavail-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	// Harness with no transcript: no models, no tokens recorded at all.

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrTokensUnavailable) {
		t.Fatalf("Price(no tokens) error = %v, want errors.Is(_, store.ErrTokensUnavailable)", err)
	}
}

// TestPriceWithoutChargeableFieldIsUnavailable covers the case where a
// "models" bucket is present but every model in it carries only
// thread-attribution bookkeeping (thinking tokens; TokenDelta's own
// Bucket type carries a "thinking" field that is real, recorded usage
// but never billed) with none of the four chargeable fields
// (input/output/cache_creation/cache_read). Without a check for this,
// Price's cost accumulator never advances past its zero initialiser and
// a literal cost_usd: 0 is written, indistinguishable from a genuinely
// free stage. Price must refuse instead.
func TestPriceWithoutChargeableFieldIsUnavailable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-nocharge-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"models":{"claude-opus-5":{"tokens":{"main":{"thinking":50}}}}}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 3, OutputPerMTok: 15, CacheWritePerMTok: 3.75, CacheReadPerMTok: 0.3,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrTokensUnavailable) {
		t.Fatalf("Price(tokens present, no chargeable field) error = %v, want errors.Is(_, store.ErrTokensUnavailable)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd was written (%s) for a run with no chargeable token field; want no cost_usd key at all", bag["cost_usd"])
	}
}

func TestPriceWithoutPricingRowIsNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-nopricing-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{"models":{"no-such-model":{"tokens":{"main":{"input":10}}}}}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(no pricing row) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}
}

// TestPriceChargesFiveMinuteAndOneHourWritesAtTheirOwnRates is task 23's
// central defect, priced directly: a run whose cache-creation usage is
// split between a 5-minute write and a 1-hour write must charge each
// portion at its own rate, not one collapsed rate applied to the whole
// total -- the 37.5% understatement this task's plan-provenance note
// measures if the two ever get collapsed onto the cheaper 5m rate.
func TestPriceChargesFiveMinuteAndOneHourWritesAtTheirOwnRates(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-split-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"cache_creation": 3000000, "cache_creation_5m": 1000000, "cache_creation_1h": 2000000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok: 0.50,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// 1 Mtok @ 6.25 (5m) + 2 Mtok @ 10 (1h) = 6.25 + 20 = 26.25. Collapsing
	// onto the 5m rate alone would give 3*6.25 = 18.75, a 28.6% understatement
	// of this bucket's true cost -- the exact failure mode this test exists
	// to catch.
	wantCost := 1*6.25 + 2*10.0
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v", bag.CostUSD, wantCost)
	}
}

// TestPriceUnknownCacheSplitIsUnpriceable is the absent-split guard: a
// bucket carrying cache_creation_unknown (this store's encoding for "the
// harvester recorded cache-creation usage but not which rate applied",
// internal/harvest.Bucket's own doc comment) must never be priced by
// guessing a rate -- the whole model bucket is treated exactly like one
// with no pricing row at all: unpriceable, and the top-level cost_usd is
// omitted.
func TestPriceUnknownCacheSplitIsUnpriceable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-unknownsplit-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"models": {
			"claude-opus-5": {
				"tokens": {"main": {"input": 1000000, "cache_creation": 500000, "cache_creation_unknown": 500000}}
			}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok: 0.50,
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(unknown cache split) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd: this run's only model bucket has an unpriceable cache split", bag["cost_usd"])
	}
	var models map[string]json.RawMessage
	if err := json.Unmarshal(bag["models"], &models); err != nil {
		t.Fatalf("unmarshal models: %v", err)
	}
	var opus struct {
		CostUSD *float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(models["claude-opus-5"], &opus); err != nil {
		t.Fatalf("unmarshal models.claude-opus-5: %v", err)
	}
	if opus.CostUSD != nil {
		t.Errorf("models.claude-opus-5.cost_usd = %v, want absent -- guessing which rate applied to the unknown-split portion is exactly what this rule forbids", *opus.CostUSD)
	}
}

// TestPriceFastModeUsesFastRate covers task 23's third defect: a run
// recorded at "fast" speed must be charged the model's fast input/output
// rate, not the standard one.
func TestPriceFastModeUsesFastRate(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-fast-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"speed": "fast",
		"models": {
			"claude-opus-5": {"tokens": {"main": {"input": 1000000, "output": 1000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-opus-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 5, OutputPerMTok: 25,
		CacheWritePerMTok: 6.25, CacheWrite5mPerMTok: 6.25, CacheWrite1hPerMTok: ptr(10.0),
		CacheReadPerMTok:  0.50,
		FastInputPerMTok:  ptr(10.0),
		FastOutputPerMTok: ptr(50.0),
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	if err := st.Price(ctx, run.ID); err != nil {
		t.Fatalf("Price: %v", err)
	}

	priced, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag struct {
		CostUSD float64 `json:"cost_usd"`
	}
	if err := json.Unmarshal(priced.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	wantCost := 1*10.0 + 1*50.0 // fast rates, not the standard 5/25.
	if diff := bag.CostUSD - wantCost; diff > 1e-9 || diff < -1e-9 {
		t.Errorf("cost_usd = %v, want %v (fast-mode input/output rate)", bag.CostUSD, wantCost)
	}
}

// TestPriceFastModeWithoutFastRateIsUnpriceable covers the other half:
// a run recorded at "fast" speed against a model with no published
// fast-mode rate (Sonnet 5 and Haiku 4.5, per pricing_seed.go's own
// table) must not be silently priced at the standard rate.
func TestPriceFastModeWithoutFastRateIsUnpriceable(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-price-fastnorate-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")

	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(`{
		"speed": "fast",
		"models": {
			"claude-sonnet-5": {"tokens": {"main": {"input": 1000000}}}
		}
	}`)); err != nil {
		t.Fatalf("MergeMetrics: %v", err)
	}
	if err := st.PutPricing(ctx, store.PricingRate{
		Model: "claude-sonnet-5", EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		InputPerMTok: 2, OutputPerMTok: 10,
		CacheWritePerMTok: 2.5, CacheWrite5mPerMTok: 2.5, CacheWrite1hPerMTok: ptr(4.0),
		CacheReadPerMTok: 0.20,
		// No fast rate published for this model -- FastInputPerMTok and
		// FastOutputPerMTok stay nil.
	}); err != nil {
		t.Fatalf("PutPricing: %v", err)
	}

	err = st.Price(ctx, run.ID)
	if !errors.Is(err, store.ErrPricingNotFound) {
		t.Fatalf("Price(fast speed, no fast rate) error = %v, want errors.Is(_, store.ErrPricingNotFound)", err)
	}

	got, err := st.GetStageRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("GetStageRun: %v", err)
	}
	var bag map[string]json.RawMessage
	if err := json.Unmarshal(got.Metrics, &bag); err != nil {
		t.Fatalf("unmarshal metrics: %v", err)
	}
	if _, ok := bag["cost_usd"]; ok {
		t.Errorf("cost_usd = %s, want no top-level cost_usd: claude-sonnet-5 has no fast-mode rate", bag["cost_usd"])
	}
}

// TestStageRunsNullRepoRootSkipsFKCheck verifies the design's claim, tagged
// unverified in design.md, that a NULL repo_root inserts cleanly with no
// matching change_repos row at all -- because Postgres' default FK match
// semantics (MATCH SIMPLE) skip the check entirely whenever any column of
// a composite foreign key is NULL.
func TestStageRunsNullRepoRootSkipsFKCheck(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-nullrepo-%d", time.Now().UnixNano())
	// No repos on this change at all.
	c := baseChange(projectKey, "kan-1")
	c.Repos = nil
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.RepoRoot = nil
	if _, err := st.BeginStage(ctx, in); err != nil {
		t.Fatalf("BeginStage with nil RepoRoot and no change_repos rows: %v", err)
	}
}

// TestStageRunsRepoRootMustMatchChangeRepos verifies the other half of the
// same unverified DDL claim: a non-NULL repo_root that names a repository
// the change does not have is rejected by the foreign key, rather than
// silently accepted.
func TestStageRunsRepoRootMustMatchChangeRepos(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-badrepo-%d", time.Now().UnixNano())
	c := baseChange(projectKey, "kan-1")
	c.Repos = []store.Repo{{RepoRoot: "/repo/a"}}
	if err := st.PutChange(ctx, c); err != nil {
		t.Fatalf("PutChange: %v", err)
	}

	in := baseBeginInput(projectKey, "kan-1", "/myflow-do", "SDD + TDD per task")
	in.RepoRoot = ptr("/repo/does-not-exist")
	_, err := st.BeginStage(ctx, in)
	if err == nil {
		t.Fatalf("BeginStage with a repo_root absent from change_repos succeeded, want a foreign-key rejection")
	}
}
