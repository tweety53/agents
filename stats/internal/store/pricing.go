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

// dispatchBucket is the shape Price reads out of one entry in the metrics
// bag's "dispatches" object (internal/harvest.DispatchBucket's own
// encoding) -- modelBucket's shape plus the dispatch's own recorded
// model, which is what Price resolves a pricing rate against: a
// dispatch's own tokens, priced at its own model's rate, never the
// model that happens to head this run's "models" bag. Model is read as
// a plain string, empty when the dispatch's meta sidecar was never
// found (DispatchBucket's own doc comment, attribute.go) -- absence
// Price treats exactly like any other unpriceable dispatch, below.
type dispatchBucket struct {
	Tokens modelBucketTokens `json:"tokens"`
	Model  string            `json:"model"`
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
//
// Price also prices every entry under the metrics bag's "dispatches" key
// (KAN-201, myflow-stats-views spec.md's "Per-dispatch cost SHALL be
// derived through the same pricing path every other cost figure uses"),
// writing "dispatches.<agentId>.cost_usd" through the identical rate
// resolution and chargeableTokens.cost arithmetic used for "models.
// <model>" above -- never an implied average rate scaled from a model
// bucket's blended total, which is what this task replaces. A dispatch's
// own tokens are already inside its recorded model's "models.<model>"
// bucket -- the same usage viewed a second way -- so a dispatch's cost is
// never added into the top-level "cost_usd": doing so would roughly
// double every run's total. A dispatch that cannot be priced (no
// recorded model, no rate in effect for it, or its own tokens carry
// cache_creation_unknown or an unrated cache/fast-mode component --
// chargeableTokens.cost's own doc comment) simply gets no cost_usd; this
// never affects Price's own return value, since a dispatch is a second,
// narrower view of tokens the "models" loop above has already reported
// on at its own granularity -- Price already surfaces incompleteness
// there, and a dispatch failing to price is not new information the
// caller needs told to it a second time. A "dispatches" value that does
// not even decode is likewise never fatal to the whole call (F6): it
// degrades to no dispatch costs this pass, so the models loop's own
// already-computed modelsPatch and top-level cost_usd still get written
// -- losing a computed total to an unrelated field's bad data is exactly
// the "worse than writing nothing" outcome this doc comment already
// warns about, one paragraph up.
//
// Dispatch pricing runs regardless of whether the "models" loop above
// priced anything at all (F28, pass 7 of this change's own review panel):
// a run whose entire "models" bag is unpriceable -- no rate in effect for
// any model recorded there -- no longer returns before the dispatches
// loop runs. A dispatch resolves its own pricing rate from its own
// sidecar-declared model, which routinely differs from whatever the
// "models" bag happened to record, so refusing to even attempt dispatch
// pricing whenever the models bag was entirely unpriceable defeated this
// method's own per-dispatch attribution exactly when the two model
// sources diverge -- the case per-dispatch attribution exists to expose.
// ErrPricingNotFound is still returned naming every unpriceable model in
// "models" whenever there is at least one, and the top-level "cost_usd"
// is still written only when every model bucket in "models" priced --
// both unchanged by whether any dispatch also priced.
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

	// rateCache memoizes pricingRateInEffect per model across both the
	// "models" loop below and the "dispatches" loop that follows it: a
	// dispatch's own recorded model is, in the ordinary case, one of the
	// same models this run's "models" bag already recorded (a dispatch's
	// tokens are a subset of its model's own bucket), so re-querying
	// Postgres a second time for the same (model, run.StartedAt) pair
	// would be pure duplicate work for what is, by construction, almost
	// always the same answer. Only successful lookups are cached --
	// errors are resolved fresh each time, which costs nothing extra
	// since an error path returns or moves on without a second lookup of
	// the same model anyway.
	rateCache := make(map[string]PricingRate)
	resolveRate := func(model string) (PricingRate, error) {
		if r, ok := rateCache[model]; ok {
			return r, nil
		}
		r, err := s.pricingRateInEffect(ctx, model, run.StartedAt)
		if err != nil {
			return PricingRate{}, err
		}
		rateCache[model] = r
		return r, nil
	}

	modelsPatch := make(map[string]any, len(candidates))
	var missing []string
	var total float64
	priced := 0
	for _, c := range candidates {
		rate, cost, ok, err := priceCandidate(resolveRate, c.model, c.tokens, fast, &missing)
		if err != nil {
			return err
		}
		if !ok {
			// A rate was in effect for this model, but not for every
			// component this bucket actually recorded -- an unknown cache
			// split, a 1-hour write against a pre-0007 row with no 1h
			// rate, or fast speed against a model with no fast rate --
			// or no rate was in effect at all. Either way, this bucket
			// is unpriceable, not partially priced (chargeableTokens.cost's
			// own doc comment explains each case; priceCandidate has
			// already appended c.model to missing).
			continue
		}
		modelsPatch[c.model] = map[string]any{
			"cost_usd":        cost,
			"pricing_version": pricingVersion(rate),
		}
		total += cost
		priced++
	}

	// dispatchesPatch prices each "dispatches.<agentId>" bucket through
	// the identical rate resolution and cost arithmetic as "models.
	// <model>" above, applied to that dispatch's own recorded model and
	// own tokens (this method's own doc comment explains why). A
	// dispatch that cannot be priced is simply omitted -- never added to
	// missing, never affecting priced/total, and never causing this
	// method to fail: it is a second, narrower view of tokens the
	// "models" loop above has already accounted for.
	dispatchesPatch := make(map[string]any)
	if dispatchesRaw, ok := bag["dispatches"]; ok {
		var dispatches map[string]dispatchBucket
		if err := json.Unmarshal(dispatchesRaw, &dispatches); err != nil {
			// Malformed, not fatal (F6, pass 1 of this change's own
			// review panel): this method's own doc comment says a
			// partial sum is "the single most dangerous output Price
			// could produce -- worse than writing nothing", and failing
			// here, before patchFields is even built below, would throw
			// away the models loop's already-computed modelsPatch and
			// total along with it -- exactly the partial-versus-nothing
			// mistake that sentence warns against, just triggered by a
			// different field. Before this dispatches loop existed, an
			// unrecognised "dispatches" key was simply never read at
			// all; a value this loop cannot decode gets the same
			// treatment now -- skipped, degrading to "no dispatch costs
			// priced this pass" -- rather than aborting the whole call.
			// No writer in this codebase currently produces a malformed
			// "dispatches" value (encodePatches, harvest/watcher.go,
			// always emits a well-formed object), so this path is
			// presently unreachable defensive robustness, not a
			// response to an observed failure.
			dispatches = nil
		}
		// Deterministic order, for the same reason candidates is sorted
		// above -- a stable map key list rather than one that depends on
		// Go's randomised iteration order. Ranges over a nil map (the
		// decode-failure case above) exactly like an empty one.
		agentIDs := make([]string, 0, len(dispatches))
		for id := range dispatches {
			agentIDs = append(agentIDs, id)
		}
		sort.Strings(agentIDs)
		for _, id := range agentIDs {
			db := dispatches[id]
			if db.Model == "" {
				// No sidecar was ever found for this dispatch's transcript
				// (DispatchBucket's own doc comment) -- absence is never a
				// value, so this dispatch is left unpriced rather than
				// priced against a neighbouring dispatch's model.
				continue
			}
			combined := combineMainAndSidechain(db.Tokens.Main, db.Tokens.Sidechain)
			if !combined.hasCharge() {
				// This dispatch's harvest recorded no chargeable token
				// field for it (or no "tokens" key at all) -- not an error,
				// just nothing to price.
				continue
			}
			// A dispatch's own failure to resolve or price is never added
			// to missing (this method's own doc comment explains why) --
			// dispatchMissing is a throwaway, never read, that exists only
			// because priceCandidate needs somewhere to append to; any
			// resolveRate failure here -- ErrPricingNotFound or a genuine
			// store error -- is treated the same, skip this dispatch,
			// never abort the whole call (F13, pass 2 of this change's own
			// review panel). Aborting here, before patchFields is even
			// built below, would discard the models loop's already-computed
			// modelsPatch and total along with it -- the identical "worse
			// than writing nothing" mistake F6's fix prevents for a
			// malformed dispatches value, just reached through a different
			// trigger.
			var dispatchMissing []string
			_, cost, ok, _ := priceCandidate(resolveRate, db.Model, combined, fast, &dispatchMissing)
			if !ok {
				continue
			}
			dispatchesPatch[id] = map[string]any{"cost_usd": cost}
		}
	}

	// patchFields writes "models" only when at least one model bucket
	// actually priced (F28, pass 7 of this change's own review panel): a
	// run whose entire "models" bag is unpriceable no longer aborts before
	// this dispatches loop even runs, so modelsPatch can legitimately be
	// empty here while dispatchesPatch still carries real, priceable
	// dispatch costs -- the dispatch's own sidecar-declared model is
	// resolved independently of whatever the models bag itself could
	// price. Omitting an empty "models" key (rather than writing "models":
	// {}) mirrors how "dispatches" is already omitted below when empty.
	patchFields := map[string]any{}
	if len(modelsPatch) > 0 {
		patchFields["models"] = modelsPatch
	}
	if len(missing) == 0 {
		patchFields["cost_usd"] = total
	}
	if len(dispatchesPatch) > 0 {
		patchFields["dispatches"] = dispatchesPatch
	}
	if len(patchFields) > 0 {
		patch, err := json.Marshal(patchFields)
		if err != nil {
			return fmt.Errorf("store: price stage run %d: encode patch: %w", stageRunID, err)
		}
		if err := s.MergeMetrics(ctx, stageRunID, patch); err != nil {
			return err
		}
	}

	if len(missing) > 0 {
		// At least one candidate model bucket had no rate in effect, or
		// carried a component its rate could not charge -- reported here
		// exactly as before F28's fix, regardless of whether a dispatch
		// elsewhere in this same run did get priced above: a dispatch is
		// a second, narrower view of tokens the "models" loop has already
		// reported on, and its pricing (or lack of it) is never new
		// information this return value needs to carry a second time
		// (this method's own doc comment).
		return fmt.Errorf("%w: %s", ErrPricingNotFound, strings.Join(missing, ", "))
	}
	return nil
}

// priceCandidate resolves model's pricing rate through resolveRate and
// prices tokens against it -- the sequence Price's "models" and
// "dispatches" loops both need (F32, pass 7 of this change's own review
// panel: hand-copying it into both loops let a future pricing-rule change
// applied to one silently miss the other). ok reports whether tokens
// priced; rate and cost are only meaningful when ok is true.
//
// err is non-nil only for a genuine resolveRate failure that is not
// ErrPricingNotFound (a real store error) -- deliberately left for the
// caller to decide what to do with, since the two loops disagree: the
// "models" loop aborts the whole Price call on it, while the "dispatches"
// loop skips just that dispatch (F13, pass 2 of this change's own review
// panel). Every other case that leaves tokens unpriced -- no rate in
// effect (ErrPricingNotFound), or a component this bucket recorded that
// the rate cannot charge (chargeableTokens.cost's own doc comment) --
// returns ok=false, err=nil and appends model to *missing so the caller's
// own missing-models bookkeeping (or a throwaway slice, for the
// dispatches loop, which never reports its own failures through missing)
// stays in exactly one place.
func priceCandidate(resolveRate func(string) (PricingRate, error), model string, tokens chargeableTokens, fast bool, missing *[]string) (rate PricingRate, cost float64, ok bool, err error) {
	rate, err = resolveRate(model)
	if err != nil {
		if errors.Is(err, ErrPricingNotFound) {
			*missing = append(*missing, model)
			return PricingRate{}, 0, false, nil
		}
		return PricingRate{}, 0, false, err
	}
	cost, ok = tokens.cost(rate, fast)
	if !ok {
		*missing = append(*missing, model)
		return PricingRate{}, 0, false, nil
	}
	return rate, cost, true, nil
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
