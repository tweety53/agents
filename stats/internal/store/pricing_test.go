package store

// Unit tests for chargeableTokens.cost's unknown-split cache-write rule
// (kan-479): ZCode reports a single collapsed cacheWriteTokens with no
// 5m/1h split, which lands in CacheCreationUnknown. Where the rate carries
// exactly one cache-write rate -- a 1h rate equal to the 5m rate -- pricing
// the unknown split at that one rate is exact, never a guess. Where the two rates differ (every seeded Claude row), the task-23
// refusal stands: pricing the rest while dropping the unknown portion would
// understate the total, and guessing a rate would invent one.

import (
	"math"
	"testing"
)

func TestCostPricesUnknownSplitAtFlatRate(t *testing.T) {
	splitRate := PricingRate{
		Model:               "claude-test-split",
		InputPerMTok:        1,
		OutputPerMTok:       2,
		CacheWrite5mPerMTok: 6.25,
		CacheWrite1hPerMTok: fastRate(10),
		CacheReadPerMTok:    0.1,
	}
	equalSplitRate := splitRate
	equalSplitRate.CacheWrite1hPerMTok = fastRate(6.25)

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

	t.Run("a nil 1h rate still refuses", func(t *testing.T) {
		// A row with no 1h rate never said its 5m rate is the only one --
		// pricing the unknown split at the 5m rate would be a guess.
		no1h := splitRate
		no1h.CacheWrite1hPerMTok = nil
		if _, ok := (chargeableTokens{CacheCreationUnknown: ptr(500_000)}).cost(no1h, false); ok {
			t.Fatal("ok = true, want false: a row with no 1h rate cannot price an unknown split")
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
