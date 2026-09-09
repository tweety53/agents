package store

// Unit tests for chargeableTokens.cost's unknown-split cache-write rule
// (kan-479): ZCode reports a single collapsed cacheWriteTokens with no
// 5m/1h split, which lands in CacheCreationUnknown. Where the rate carries
// exactly one cache-write rate -- a nil 1h rate, or a 1h rate equal to the
// 5m rate -- pricing the unknown split at that one rate is exact, never a
// guess. Where the two rates differ (every seeded Claude row), the task-23
// refusal stands: pricing the rest while dropping the unknown portion would
// understate the total, and guessing a rate would invent one.

import (
	"math"
	"testing"
)

func TestCostPricesUnknownSplitAtFlatRate(t *testing.T) {
	flatRate := PricingRate{
		Model:               "glm-test-flat",
		InputPerMTok:        1,
		OutputPerMTok:       2,
		CacheWritePerMTok:   3,
		CacheWrite5mPerMTok: 3,
		CacheReadPerMTok:    0.1,
		// CacheWrite1hPerMTok nil: the provider publishes one cache-write rate.
	}
	splitRate := PricingRate{
		Model:               "claude-test-split",
		InputPerMTok:        1,
		OutputPerMTok:       2,
		CacheWritePerMTok:   6.25,
		CacheWrite5mPerMTok: 6.25,
		CacheWrite1hPerMTok: fastRate(10),
		CacheReadPerMTok:    0.1,
	}
	equalSplitRate := splitRate
	equalSplitRate.CacheWrite1hPerMTok = fastRate(6.25)

	t.Run("nil 1h rate prices the unknown portion at the 5m rate", func(t *testing.T) {
		cost, ok := chargeableTokens{CacheCreationUnknown: ptr(1_000_000)}.cost(flatRate, false)
		if !ok {
			t.Fatal("ok = false, want true: one published rate means the unknown split is exact")
		}
		if math.Abs(cost-3) > 1e-9 {
			t.Errorf("cost = %v, want 3 (1M tokens at the single cache-write rate)", cost)
		}
	})

	t.Run("equal 5m and 1h rates price the unknown portion at that rate", func(t *testing.T) {
		cost, ok := chargeableTokens{CacheCreationUnknown: ptr(500_000)}.cost(equalSplitRate, false)
		if !ok {
			t.Fatal("ok = false, want true: equal rates are one rate in effect")
		}
		if math.Abs(cost-6.25*0.5) > 1e-9 {
			t.Errorf("cost = %v, want %v", cost, 6.25*0.5)
		}
	})

	t.Run("differing 5m and 1h rates still refuse", func(t *testing.T) {
		_, ok := chargeableTokens{CacheCreationUnknown: ptr(1_000_000)}.cost(splitRate, false)
		if ok {
			t.Fatal("ok = true, want false: two different rates cannot price an unknown split without guessing")
		}
	})

	t.Run("a pre-0007 row with only the collapsed column set still refuses", func(t *testing.T) {
		// The 5m column's zero here is an unset, not a published free --
		// pricing the unknown split at it would be silently zero, and the
		// collapsed column alone does not say which split rate applied.
		degenerate := PricingRate{
			Model:             "claude-pre-0007",
			InputPerMTok:      1,
			OutputPerMTok:     5,
			CacheWritePerMTok: 1.25,
			CacheReadPerMTok:  0.1,
		}
		if _, ok := (chargeableTokens{CacheCreationUnknown: ptr(500_000)}).cost(degenerate, false); ok {
			t.Fatal("ok = true, want false: a row whose split columns were never filled prices nothing")
		}
	})

	t.Run("a bucket with no unknown portion is unaffected", func(t *testing.T) {
		cost, ok := chargeableTokens{Input: ptr(2_000_000)}.cost(splitRate, false)
		if !ok {
			t.Fatal("ok = false, want true")
		}
		if math.Abs(cost-2) > 1e-9 {
			t.Errorf("cost = %v, want 2", cost)
		}
	})
}

func ptr(v float64) *float64 { return &v }
