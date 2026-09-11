package store_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestSeedPricingRatesRoundTrip is step 3's own guard: every rate
// SeedPricingRates returns must actually be usable by Price once
// SeedPricing has written it -- a wrong column, a swapped rate, or a rate
// that silently failed to persist would otherwise pass PutPricing's own
// unit tests (which construct a PricingRate by hand) while never actually
// working end to end.
func TestSeedPricingRatesRoundTrip(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	if err := st.SeedPricing(ctx); err != nil {
		t.Fatalf("SeedPricing: %v", err)
	}

	for _, rate := range store.SeedPricingRates() {
		rate := rate
		t.Run(rate.Model, func(t *testing.T) {
			projectKey := fmt.Sprintf("proj-seed-%s-%d", rate.Model, time.Now().UnixNano())
			seedChange(t, st, projectKey, "kan-1")

			in := baseBeginInput(projectKey, "kan-1", "/flow", "SDD + TDD per task")
			// Started well after this seed's effective_from, so the seeded
			// row is the one in effect.
			in.StartedAt = time.Date(2026, 8, 13, 10, 0, 0, 0, time.UTC)
			run, err := st.BeginStage(ctx, in)
			if err != nil {
				t.Fatalf("BeginStage: %v", err)
			}
			patch := fmt.Sprintf(`{"models":{%q:{"tokens":{"main":{"input":1000000}}}}}`, rate.Model)
			if err := st.MergeMetrics(ctx, run.ID, json.RawMessage(patch)); err != nil {
				t.Fatalf("MergeMetrics: %v", err)
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
			if diff := bag.CostUSD - rate.InputPerMTok; diff > 1e-9 || diff < -1e-9 {
				t.Errorf("%s: cost_usd = %v, want %v (1 Mtok input at this model's own seeded rate)", rate.Model, bag.CostUSD, rate.InputPerMTok)
			}
		})
	}
}

// TestSeedPricingIsIdempotent proves step 3's "flowd upserts them at
// startup ... whose ON CONFLICT already makes re-seeding a no-op" claim
// directly: calling SeedPricing twice must not error (a naive INSERT
// without ON CONFLICT would violate the (model, effective_from) primary
// key on the second call).
func TestSeedPricingIsIdempotent(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()

	if err := st.SeedPricing(ctx); err != nil {
		t.Fatalf("SeedPricing (first): %v", err)
	}
	if err := st.SeedPricing(ctx); err != nil {
		t.Fatalf("SeedPricing (second): %v", err)
	}
}

// TestSeedPricingRatesCarryFable pins the row for the operator's own
// session model -- absent, every fable-orchestrated run priced its parent
// at zero (design.md's fable-pricing-row). Rates are the published table
// read on 2026-09-02; the cache-read rate is the one column that is not
// 0.1x input on this model.
func TestSeedPricingRatesCarryFable(t *testing.T) {
	var fable *store.PricingRate
	for _, rate := range store.SeedPricingRates() {
		if rate.Model == "claude-fable-5-1" {
			rate := rate
			fable = &rate
		}
	}
	if fable == nil {
		t.Fatal("no seeded rate for claude-fable-5-1")
	}
	if fable.InputPerMTok != 10 || fable.OutputPerMTok != 50 {
		t.Errorf("input/output = %v/%v, want 10/50", fable.InputPerMTok, fable.OutputPerMTok)
	}
	if fable.CacheWrite5mPerMTok != 12.5 || fable.CacheWrite1hPerMTok == nil || *fable.CacheWrite1hPerMTok != 20 {
		t.Errorf("cache writes = %v/%v, want 12.5/20", fable.CacheWrite5mPerMTok, fable.CacheWrite1hPerMTok)
	}
	if fable.CacheReadPerMTok != 0.25 {
		t.Errorf("cache read = %v, want 0.25", fable.CacheReadPerMTok)
	}
	if fable.FastInputPerMTok != nil || fable.FastOutputPerMTok != nil {
		t.Error("fast rate present, want nil (fast mode is Opus 5 / 4.8 only)")
	}
}

// TestSeedPricingRatesOmitFastForModelsWithNone pins the table's own
// documented gap: claude-sonnet-5 and claude-haiku-4-5 have no published
// fast-mode rate (the source table's "—" cells), and this seed must not
// invent one for them.
func TestSeedPricingRatesOmitFastForModelsWithNone(t *testing.T) {
	for _, rate := range store.SeedPricingRates() {
		switch rate.Model {
		case "claude-sonnet-5", "claude-haiku-4-5", "claude-fable-5-1":
			if rate.FastInputPerMTok != nil || rate.FastOutputPerMTok != nil {
				t.Errorf("%s: fast rate present, want nil (no fast-mode rate is published for this model)", rate.Model)
			}
		case "claude-opus-5", "claude-opus-4-8":
			if rate.FastInputPerMTok == nil || rate.FastOutputPerMTok == nil {
				t.Errorf("%s: fast rate missing, want both FastInputPerMTok and FastOutputPerMTok set", rate.Model)
			}
		}
	}
}

// TestSeedPricingRatesCarriesGLMFlash pins the kan-479 seed row: the one
// model this machine's ZCode rollouts record, at the discounted plan rates
// published at docs.z.ai/guides/overview/pricing (read 2026-09-09) -- and
// with a nil 1h cache-write rate, which is what makes the model's single
// published cache rate priceable for ZCode's unknown-split cache writes.
func TestSeedPricingRatesCarriesGLMFlash(t *testing.T) {
	var found *store.PricingRate
	for i, rate := range store.SeedPricingRates() {
		if rate.Model == "glm-5.3-flash" {
			found = &store.SeedPricingRates()[i]
			break
		}
	}
	if found == nil {
		t.Fatal("SeedPricingRates carries no glm-5.3-flash row")
	}
	want := map[string]float64{
		"input": 0.075, "output": 0.25, "cache write": 0, "cache read": 0.015,
	}
	if found.InputPerMTok != want["input"] {
		t.Errorf("InputPerMTok = %v, want %v", found.InputPerMTok, want["input"])
	}
	if found.OutputPerMTok != want["output"] {
		t.Errorf("OutputPerMTok = %v, want %v", found.OutputPerMTok, want["output"])
	}
	if found.CacheWritePerMTok != want["cache write"] || found.CacheWrite5mPerMTok != want["cache write"] {
		t.Errorf("cache write rates = (%v, %v), want (0, 0): cache storage is limited-time free",
			found.CacheWritePerMTok, found.CacheWrite5mPerMTok)
	}
	if found.CacheReadPerMTok != want["cache read"] {
		t.Errorf("CacheReadPerMTok = %v, want %v", found.CacheReadPerMTok, want["cache read"])
	}
	if found.CacheWrite1hPerMTok != nil {
		t.Errorf("CacheWrite1hPerMTok = %v, want nil: Z.ai publishes one cache rate", *found.CacheWrite1hPerMTok)
	}
	if found.FastInputPerMTok != nil || found.FastOutputPerMTok != nil {
		t.Errorf("fast rates = (%v, %v), want (nil, nil): no fast-mode rate is published for this model",
			found.FastInputPerMTok, found.FastOutputPerMTok)
	}
}
