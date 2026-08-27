package store

import (
	"context"
	"fmt"
	"time"
)

// pricingSeedEffectiveFrom is set well before this store's own recorded
// history began: the store started empty in August 2026 (tasks.md's own
// "Baseline" section), so one row per model, effective from this instant,
// already covers every stage run this store will ever have to price
// retroactively. The pricing table is keyed (model, effective_from), so a
// future rate change is a later row -- an insert, never a migration.
var pricingSeedEffectiveFrom = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// fastRate is a small constructor for the *float64 fields FastInputPerMTok,
// FastOutputPerMTok and CacheWrite1hPerMTok expect -- named for what it is
// used for here, not because it means anything beyond "a published rate
// exists for this component".
func fastRate(v float64) *float64 { return &v }

// SeedPricingRates is the published Anthropic rate table this task's own
// plan-provenance tag cites: read from
// https://platform.claude.com/docs/en/about-claude/pricing on 2026-08-14.
// Prices are USD per million tokens (per_mtok).
//
// These figures are not re-derived, adjusted, or supplemented from
// memory -- they are exactly the table task 23's plan carries, and a
// model or rate this table does not name (claude-sonnet-5 and
// claude-haiku-4-5 have no fast-mode rate at all, per the source table's
// own "—" cells) is priced as unavailable by Store.Price rather than at
// an invented rate: see PricingRate's own doc comment, and
// chargeableTokens.cost, for how a nil rate here is treated at pricing
// time.
func SeedPricingRates() []PricingRate {
	return []PricingRate{
		{
			Model:               "claude-opus-5",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        5,
			OutputPerMTok:       25,
			CacheWritePerMTok:   6.25, // legacy column; superseded by the 5m/1h split below.
			CacheWrite5mPerMTok: 6.25,
			CacheWrite1hPerMTok: fastRate(10),
			CacheReadPerMTok:    0.50,
			FastInputPerMTok:    fastRate(10),
			FastOutputPerMTok:   fastRate(50),
		},
		{
			Model:               "claude-opus-4-8",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        5,
			OutputPerMTok:       25,
			CacheWritePerMTok:   6.25,
			CacheWrite5mPerMTok: 6.25,
			CacheWrite1hPerMTok: fastRate(10),
			CacheReadPerMTok:    0.50,
			FastInputPerMTok:    fastRate(10),
			FastOutputPerMTok:   fastRate(50),
		},
		{
			Model:               "claude-sonnet-5",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        2,
			OutputPerMTok:       10,
			CacheWritePerMTok:   2.50,
			CacheWrite5mPerMTok: 2.50,
			CacheWrite1hPerMTok: fastRate(4),
			CacheReadPerMTok:    0.20,
			// No fast-mode rate published for this model -- left nil
			// deliberately, per SeedPricingRates' own doc comment.
		},
		{
			Model:               "claude-haiku-4-5",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        1,
			OutputPerMTok:       5,
			CacheWritePerMTok:   1.25,
			CacheWrite5mPerMTok: 1.25,
			CacheWrite1hPerMTok: fastRate(2),
			CacheReadPerMTok:    0.10,
			// No fast-mode rate published for this model either.
		},
	}
}

// SeedPricing upserts every rate SeedPricingRates returns, through the
// existing PutPricing -- whose ON CONFLICT (model, effective_from) already
// makes republishing the same rate a no-op, so calling this at every
// flowd startup (cmd/flowd/main.go) is safe regardless of how many
// times it has already run.
func (s *Store) SeedPricing(ctx context.Context) error {
	for _, r := range SeedPricingRates() {
		if err := s.PutPricing(ctx, r); err != nil {
			return fmt.Errorf("store: seed pricing for %s: %w", r.Model, err)
		}
	}
	return nil
}
