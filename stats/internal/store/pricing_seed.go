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

// SeedPricingRates is the published rate table this change's own
// plan-provenance tags cite. The Claude rows were read from
// https://platform.claude.com/docs/en/about-claude/pricing on 2026-08-14,
// with the claude-fable-5-1 row read from the same page on 2026-09-02. The
// GLM-5.3-Flash row (kan-479, keyed to the canonical lowercase form of
// the model id -- design decision model-id-canonical-lowercase) was read from
// https://docs.z.ai/guides/overview/pricing on 2026-09-09: the launch-
// discounted prices this machine's Z.ai coding plan actually charges
// (input $0.15, cached-input read $0.03, output $0.50 at list, all at a
// 50% launch discount at read time), with cache-write storage
// limited-time free. Prices are USD per million tokens (per_mtok).
//
// These figures are not re-derived, adjusted, or supplemented from
// memory -- and a model or rate this table does not name (claude-sonnet-5,
// claude-haiku-4-5 and claude-fable-5-1 have no fast-mode rate at all,
// per the source table's own "—" cells; glm-5.3-flash has no fast rate
// either, and no separate 1-hour cache rate, which is what lets the
// pricer price ZCode's unknown-split cache writes exactly --
// chargeableTokens.cost's flat-rate rule) is priced as unavailable by
// Store.Price rather than at an invented rate: see PricingRate's own doc
// comment, and chargeableTokens.cost, for how a nil rate here is treated
// at pricing time.
func SeedPricingRates() []PricingRate {
	return []PricingRate{
		{
			Model:               "glm-5.3-flash",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        0.075,
			OutputPerMTok:       0.25,
			CacheWritePerMTok:   0,
			CacheWrite5mPerMTok: 0,
			CacheReadPerMTok:    0.015,
			// CacheWrite1hPerMTok nil: Z.ai publishes one cache-write rate
			// (limited-time free), and no fast-mode rate is published for
			// this model -- both left nil deliberately.
		},
		{
			Model:               "claude-fable-5-1",
			EffectiveFrom:       pricingSeedEffectiveFrom,
			InputPerMTok:        10,
			OutputPerMTok:       50,
			CacheWritePerMTok:   12.50, // legacy column; superseded by the 5m/1h split below.
			CacheWrite5mPerMTok: 12.50,
			CacheWrite1hPerMTok: fastRate(20),
			CacheReadPerMTok:    0.25, // 0.025x input on this model, not the 0.1x every other row uses.
			// No fast-mode rate published for this model (Opus 5 / 4.8 only).
		},
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
