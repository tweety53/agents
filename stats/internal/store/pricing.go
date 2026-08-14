package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrPricingNotFound is returned by Price when no pricing row is in effect,
// for the stage run's model, at or before its start instant.
var ErrPricingNotFound = errors.New("store: no pricing in effect for model at stage run start")

// ErrTokensUnavailable is returned by Price when a stage run's metrics
// carry no usable token data to price -- because the harness recorded no
// model, or recorded no tokens bag at all. Absence is not zero: Price
// refuses to compute a cost from data that was never measured rather than
// silently pricing it at zero.
var ErrTokensUnavailable = errors.New("store: token metrics unavailable for this stage run")

// tokensPerMillion is the unit pricing.*_per_mtok rates are expressed in.
const tokensPerMillion = 1_000_000.0

// PricingRate is one row of the pricing table: the per-token-type rate for
// one model, in effect from effectiveFrom onward until superseded by a
// later row for the same model.
//
// CacheWritePerMTok is the original, single cache-write rate (0003
// _stage_runs.sql) -- kept because the column still exists (0007
// _pricing_rate_shape.sql leaves it in place) but no longer read by Price,
// which now charges CacheWrite5mPerMTok and CacheWrite1hPerMTok
// separately: the two rates Anthropic actually charges (task 23), which a
// single collapsed rate cannot represent correctly. CacheWrite1hPerMTok,
// FastInputPerMTok and FastOutputPerMTok are nullable, matching their
// nullable columns: a nil here means no rate was ever published for that
// component (a model with no fast-mode rate at all, or a pre-0007 row
// that predates the 1-hour split), and Price treats that exactly like a
// missing pricing row for that bucket -- unpriceable, never priced at 0
// or at a neighbouring rate.
type PricingRate struct {
	Model               string
	EffectiveFrom       time.Time
	InputPerMTok        float64
	OutputPerMTok       float64
	CacheWritePerMTok   float64
	CacheWrite5mPerMTok float64
	CacheWrite1hPerMTok *float64
	CacheReadPerMTok    float64
	FastInputPerMTok    *float64
	FastOutputPerMTok   *float64
}

// PutPricing records one pricing rate, upserting on (model, effective_from)
// so republishing the same rate is a no-op rather than a duplicate row.
func (s *Store) PutPricing(ctx context.Context, r PricingRate) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO pricing (
			model, effective_from, input_per_mtok, output_per_mtok,
			cache_write_per_mtok, cache_write_5m_per_mtok, cache_write_1h_per_mtok,
			cache_read_per_mtok, fast_input_per_mtok, fast_output_per_mtok
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (model, effective_from) DO UPDATE SET
			input_per_mtok          = EXCLUDED.input_per_mtok,
			output_per_mtok         = EXCLUDED.output_per_mtok,
			cache_write_per_mtok    = EXCLUDED.cache_write_per_mtok,
			cache_write_5m_per_mtok = EXCLUDED.cache_write_5m_per_mtok,
			cache_write_1h_per_mtok = EXCLUDED.cache_write_1h_per_mtok,
			cache_read_per_mtok     = EXCLUDED.cache_read_per_mtok,
			fast_input_per_mtok     = EXCLUDED.fast_input_per_mtok,
			fast_output_per_mtok    = EXCLUDED.fast_output_per_mtok
	`, r.Model, r.EffectiveFrom, r.InputPerMTok, r.OutputPerMTok,
		r.CacheWritePerMTok, r.CacheWrite5mPerMTok, r.CacheWrite1hPerMTok,
		r.CacheReadPerMTok, r.FastInputPerMTok, r.FastOutputPerMTok)
	if err != nil {
		return fmt.Errorf("store: put pricing for %s @ %s: %w", r.Model, r.EffectiveFrom, err)
	}
	return nil
}

// chargeableTokens is the shape Price reads out of one model bucket's
// "tokens.main" or "tokens.sidechain" object (internal/harvest's Bucket
// type, the harvester's own encoding) -- the fields a PricingRate
// actually prices. "thinking" and the top-level, collapsed
// "cache_creation" total are both deliberately absent: "thinking" has no
// rate at all (unchanged from before task 22); the collapsed total is
// real, recorded usage but cannot say which of the two cache-write rates
// applied to it, and Price now reads only the split fields below plus
// CacheCreationUnknown, which exists precisely so that ambiguity is
// visible to Price rather than silently discarded. Each field is
// optional -- absence means that token type was never recorded and
// contributes nothing to the cost, which is a different fact from a
// recorded zero but is priced the same way: no charge for tokens that
// were not spent.
type chargeableTokens struct {
	Input                *float64 `json:"input"`
	Output               *float64 `json:"output"`
	CacheCreation5m      *float64 `json:"cache_creation_5m"`
	CacheCreation1h      *float64 `json:"cache_creation_1h"`
	CacheCreationUnknown *float64 `json:"cache_creation_unknown"`
	CacheRead            *float64 `json:"cache_read"`
}

// hasCharge reports whether t carries at least one of the chargeable
// fields, distinguishing "recorded some usage worth pricing" (including
// cache-creation usage whose split is not known, which is real spend
// this method still counts as a pricing *candidate* -- see cost's own
// doc comment for why it then always fails to price) from a bucket that
// carries only unbilled bookkeeping (thinking tokens, or nothing at all).
func (t chargeableTokens) hasCharge() bool {
	return t.Input != nil || t.Output != nil || t.CacheCreation5m != nil ||
		t.CacheCreation1h != nil || t.CacheCreationUnknown != nil || t.CacheRead != nil
}

// cost computes t's price under rate, charging only the fields t
// actually recorded, using the fast-mode input/output rate when fast is
// true. It returns ok=false -- refusing to price t at all -- in exactly
// three cases, every one of them a place guessing would silently invent a
// number rather than surface an absence:
//
//   - t carries CacheCreationUnknown: some of this bucket's cache-creation
//     usage never recorded which of the two rates applied, and pricing
//     the rest while dropping that portion would produce a total that
//     reads as complete but understates the true cost by exactly the
//     unpriced amount -- the same "partial sum looks like a correct one"
//     failure Price's own top-level cost_usd rule already guards against,
//     one level down.
//   - t carries CacheCreation1h but rate has no CacheWrite1hPerMTok (a
//     pre-0007 pricing row, or one nobody has published a 1-hour rate
//     for yet): there is no rate to charge that portion at.
//   - fast is true but rate has no fast-mode rate for one or both of
//     input/output (Sonnet 5 and Haiku 4.5 have none at all, per
//     pricing_seed.go's table): a run recorded as fast-speed against a
//     model with no fast rate is unpriceable, not silently priced at the
//     standard rate.
func (t chargeableTokens) cost(rate PricingRate, fast bool) (cost float64, ok bool) {
	if t.CacheCreationUnknown != nil {
		return 0, false
	}

	inputRate, outputRate := rate.InputPerMTok, rate.OutputPerMTok
	if fast {
		if rate.FastInputPerMTok == nil || rate.FastOutputPerMTok == nil {
			return 0, false
		}
		inputRate, outputRate = *rate.FastInputPerMTok, *rate.FastOutputPerMTok
	}

	if t.Input != nil {
		cost += *t.Input / tokensPerMillion * inputRate
	}
	if t.Output != nil {
		cost += *t.Output / tokensPerMillion * outputRate
	}
	if t.CacheCreation5m != nil {
		cost += *t.CacheCreation5m / tokensPerMillion * rate.CacheWrite5mPerMTok
	}
	if t.CacheCreation1h != nil {
		if rate.CacheWrite1hPerMTok == nil {
			return 0, false
		}
		cost += *t.CacheCreation1h / tokensPerMillion * *rate.CacheWrite1hPerMTok
	}
	if t.CacheRead != nil {
		cost += *t.CacheRead / tokensPerMillion * rate.CacheReadPerMTok
	}
	return cost, true
}

// sumOptional adds two optional token counts, staying nil only when both
// sides are nil -- the same "absence is not zero, but two absences plus
// a presence is still a presence" rule chargeableTokens.hasCharge relies
// on to decide whether a combined main+sidechain bucket has anything
// chargeable in it at all.
func sumOptional(a, b *float64) *float64 {
	if a == nil && b == nil {
		return nil
	}
	var v float64
	if a != nil {
		v += *a
	}
	if b != nil {
		v += *b
	}
	return &v
}

// combineMainAndSidechain merges one model's main and sidechain buckets
// into the single chargeableTokens Price prices against -- a model's
// cost does not distinguish which thread its tokens came from, unlike
// CostPerChange's main/sidechain split, which exists precisely to keep
// that distinction visible elsewhere.
func combineMainAndSidechain(main, sidechain *chargeableTokens) chargeableTokens {
	var m, sc chargeableTokens
	if main != nil {
		m = *main
	}
	if sidechain != nil {
		sc = *sidechain
	}
	return chargeableTokens{
		Input:                sumOptional(m.Input, sc.Input),
		Output:               sumOptional(m.Output, sc.Output),
		CacheCreation5m:      sumOptional(m.CacheCreation5m, sc.CacheCreation5m),
		CacheCreation1h:      sumOptional(m.CacheCreation1h, sc.CacheCreation1h),
		CacheCreationUnknown: sumOptional(m.CacheCreationUnknown, sc.CacheCreationUnknown),
		CacheRead:            sumOptional(m.CacheRead, sc.CacheRead),
	}
}

// modelBucketTokens is the shape Price reads out of one model bucket's
// "tokens" key (internal/harvest.TokenDelta's own encoding: "main" and
// "sidechain", each optional -- a bucket recording only one kind of
// usage must not fabricate a recorded zero for the other).
type modelBucketTokens struct {
	Main      *chargeableTokens `json:"main"`
	Sidechain *chargeableTokens `json:"sidechain"`
}

// modelBucket is the shape Price reads out of one entry in the metrics
// bag's "models" object (internal/harvest.ModelBucket's own encoding).
type modelBucket struct {
	Tokens modelBucketTokens `json:"tokens"`
}

// Price resolves the pricing row in effect, at the stage run's start
// instant, for every model its metrics bag's "models" key recorded, and
// prices each model's bucket separately -- writing "models.<model>.cost_usd"
// and "models.<model>.pricing_version" per model -- through MergeMetrics,
// so the token counts and every other existing key are left exactly as
// they were. History remains re-priceable because the token counts are
// never overwritten, only read.
//
// The top-level "cost_usd" is written as the sum of every bucket Price
// could price, but only when every model bucket in this run priced
// successfully. A run with one unpriceable bucket (no rate in effect for
// that model) still gets every other bucket's cost_usd written, and
// still gets no top-level cost_usd at all: a partial sum would read
// exactly like a complete one at every layer that reads cost_usd above
// this method, which makes it the single most dangerous output Price
// could produce -- worse than writing nothing.
//
// Price returns ErrTokensUnavailable if the stage run's metrics carry no
// "models" key, or every model bucket in it carries no chargeable token
// field -- a harness with no transcript, or a stage priced before the
// harvester has run. It returns ErrPricingNotFound, naming every model
// with no rate in effect, when at least one model bucket could not be
// priced; the models that could still get their buckets written (this
// method's own doc comment above), so this error reports an incomplete
// result, not a failed one.
func (s *Store) Price(ctx context.Context, stageRunID int64) error {
	run, err := s.GetStageRun(ctx, stageRunID)
	if err != nil {
		return err
	}

	var bag map[string]json.RawMessage
	if err := json.Unmarshal(run.Metrics, &bag); err != nil {
		return fmt.Errorf("store: price stage run %d: decode metrics: %w", stageRunID, err)
	}

	modelsRaw, ok := bag["models"]
	if !ok {
		return fmt.Errorf("%w: stage run %d has no models recorded", ErrTokensUnavailable, stageRunID)
	}
	var models map[string]modelBucket
	if err := json.Unmarshal(modelsRaw, &models); err != nil {
		return fmt.Errorf("store: price stage run %d: decode models: %w", stageRunID, err)
	}

	// speed is a whole-run property (harvest.MetricsPatch's own "speed"
	// key, last-write-wins, populated by internal/harvest since task 23),
	// not per-model -- a stage run's session runs at one speed setting at
	// a time. Absent entirely on a run harvested before this task, or on
	// one whose harness never records it, which decodes to "" here and is
	// read as standard speed, exactly as it always priced before fast
	// mode existed.
	var speed string
	if speedRaw, ok := bag["speed"]; ok {
		if err := json.Unmarshal(speedRaw, &speed); err != nil {
			return fmt.Errorf("store: price stage run %d: decode speed: %w", stageRunID, err)
		}
	}
	fast := speed == "fast"

	type candidate struct {
		model  string
		tokens chargeableTokens
	}
	var candidates []candidate
	for model, mb := range models {
		combined := combineMainAndSidechain(mb.Tokens.Main, mb.Tokens.Sidechain)
		if !combined.hasCharge() {
			// This model bucket carries only unbilled bookkeeping (or
			// nothing at all) -- not a model to price, and not an error on
			// its own; see this method's own doc comment for when the
			// whole call refuses instead.
			continue
		}
		candidates = append(candidates, candidate{model: model, tokens: combined})
	}
	if len(candidates) == 0 {
		return fmt.Errorf("%w: stage run %d has no chargeable token field in any model bucket", ErrTokensUnavailable, stageRunID)
	}
	// Deterministic order, purely so a caller reading ErrPricingNotFound's
	// message (or a test asserting against it) sees a stable model list
	// rather than one that depends on Go's randomised map iteration.
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].model < candidates[j].model })

	modelsPatch := make(map[string]any, len(candidates))
	var missing []string
	var total float64
	priced := 0
	for _, c := range candidates {
		rate, err := s.pricingRateInEffect(ctx, c.model, run.StartedAt)
		if err != nil {
			if errors.Is(err, ErrPricingNotFound) {
				missing = append(missing, c.model)
				continue
			}
			return err
		}
		cost, ok := c.tokens.cost(rate, fast)
		if !ok {
			// A rate was in effect for this model, but not for every
			// component this bucket actually recorded -- an unknown cache
			// split, a 1-hour write against a pre-0007 row with no 1h
			// rate, or fast speed against a model with no fast rate.
			// Treated identically to no rate at all: this bucket is
			// unpriceable, not partially priced (chargeableTokens.cost's
			// own doc comment explains each case).
			missing = append(missing, c.model)
			continue
		}
		modelsPatch[c.model] = map[string]any{
			"cost_usd":        cost,
			"pricing_version": pricingVersion(rate),
		}
		total += cost
		priced++
	}

	if priced == 0 {
		// Nothing in this run could be priced at all -- every candidate
		// model lacked a rate. Nothing is written; this is a straight
		// failure, not a partial result.
		return fmt.Errorf("%w: %s", ErrPricingNotFound, strings.Join(missing, ", "))
	}

	patchFields := map[string]any{"models": modelsPatch}
	if len(missing) == 0 {
		patchFields["cost_usd"] = total
	}
	patch, err := json.Marshal(patchFields)
	if err != nil {
		return fmt.Errorf("store: price stage run %d: encode patch: %w", stageRunID, err)
	}
	if err := s.MergeMetrics(ctx, stageRunID, patch); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w: %s", ErrPricingNotFound, strings.Join(missing, ", "))
	}
	return nil
}

// pricingVersion names the exact pricing row a cost was derived from, so a
// later reader can tell which rate produced a stored figure without
// re-resolving it.
func pricingVersion(r PricingRate) string {
	return fmt.Sprintf("%s@%s", r.Model, r.EffectiveFrom.UTC().Format(time.RFC3339))
}

// pricingRateInEffect returns the most recent pricing row for model whose
// effective_from is at or before at -- the row "in effect" at that instant.
func (s *Store) pricingRateInEffect(ctx context.Context, model string, at time.Time) (PricingRate, error) {
	var r PricingRate
	err := s.pool.QueryRow(ctx, `
		SELECT model, effective_from, input_per_mtok, output_per_mtok,
		       cache_write_per_mtok, cache_write_5m_per_mtok, cache_write_1h_per_mtok,
		       cache_read_per_mtok, fast_input_per_mtok, fast_output_per_mtok
		FROM pricing
		WHERE model = $1 AND effective_from <= $2
		ORDER BY effective_from DESC
		LIMIT 1
	`, model, at).Scan(
		&r.Model, &r.EffectiveFrom, &r.InputPerMTok, &r.OutputPerMTok,
		&r.CacheWritePerMTok, &r.CacheWrite5mPerMTok, &r.CacheWrite1hPerMTok,
		&r.CacheReadPerMTok, &r.FastInputPerMTok, &r.FastOutputPerMTok,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PricingRate{}, fmt.Errorf("%w: %q at %s", ErrPricingNotFound, model, at)
		}
		return PricingRate{}, fmt.Errorf("store: resolve pricing for %s at %s: %w", model, at, err)
	}
	return r, nil
}
