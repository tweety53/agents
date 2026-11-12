package store_test

// Integration tests for PriceDispatch (dispatch-cost-on-row-metrics): a
// dispatch row's own metrics.tokens priced at its served model, at the rate
// in effect when it started. They live outside pricing_test.go because that
// file is package store, which cannot import internal/storetest (storetest
// imports store).

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/tweety53/agents/stats/internal/store"
)

// One million output tokens: 10 USD at the first rate, 20 at the second.
const dispatchTokens = `{"tokens":{"main":{"output":600000},"sidechain":{"output":400000}}}`

var (
	firstRateFrom  = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	secondRateFrom = time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
)

// pricedDispatchStore seeds a change, a first claude-opus-5-5 rate in effect
// from firstRateFrom and a dearer one from secondRateFrom, and returns a
// recorder of dispatch rows started at baseDispatch's 2026-08-22.
func pricedDispatchStore(t *testing.T) (*store.Store, *pgxpool.Pool, func(key, metrics string) int64) {
	t.Helper()
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-dispatch-cost-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	for _, r := range []store.PricingRate{
		{Model: "claude-opus-5-5", EffectiveFrom: firstRateFrom, OutputPerMTok: 10},
		{Model: "claude-opus-5-5", EffectiveFrom: secondRateFrom, OutputPerMTok: 20},
	} {
		if err := st.PutPricing(ctx, r); err != nil {
			t.Fatalf("PutPricing: %v", err)
		}
	}
	record := func(key, metrics string) int64 {
		t.Helper()
		d := baseDispatch("implementer", "opus")
		d.Key, d.AgentID = key, "inline"
		if metrics != "" {
			d.Metrics = json.RawMessage(metrics)
		}
		out, err := st.RecordDispatch(ctx, projectKey, "kan-1", d)
		if err != nil {
			t.Fatalf("RecordDispatch(%s): %v", key, err)
		}
		return out.ID
	}
	return st, pool, record
}

// dispatchCost reads a dispatch row's metrics.cost_usd, nil when unset.
func dispatchCost(t *testing.T, pool *pgxpool.Pool, id int64) *float64 {
	t.Helper()
	var cost *float64
	if err := pool.QueryRow(context.Background(),
		`SELECT (metrics->>'cost_usd')::float8 FROM dispatches WHERE id = $1`, id).Scan(&cost); err != nil {
		t.Fatalf("read cost of dispatch %d: %v", id, err)
	}
	return cost
}

func wantCost(t *testing.T, got *float64, want float64) {
	t.Helper()
	if got == nil {
		t.Errorf("cost_usd unset, want %v", want)
	} else if math.Abs(*got-want) > 1e-9 {
		t.Errorf("cost_usd = %v, want %v", *got, want)
	}
}

// TestDispatchRowPricedAtItsServedModel: an inline row labelled opus whose
// metrics name claude-opus-5-5 is priced at that model's rate in effect at
// its start, never at the label or at today's dearer rate.
func TestDispatchRowPricedAtItsServedModel(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	id := record("task-1-implementer", `{"modelVersion":"claude-opus-5-5",`+dispatchTokens[1:])
	if err := st.PriceDispatch(context.Background(), id); err != nil {
		t.Fatalf("PriceDispatch: %v", err)
	}
	wantCost(t, dispatchCost(t, pool, id), 10)
}

// TestDispatchRowCostIsSetNeverAdded: pricing the same row twice leaves the
// same cost -- every harvest pass reprices, so an added cost would inflate.
func TestDispatchRowCostIsSetNeverAdded(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	id := record("task-1-implementer", `{"modelVersion":"claude-opus-5-5",`+dispatchTokens[1:])
	for range 2 {
		if err := st.PriceDispatch(context.Background(), id); err != nil {
			t.Fatalf("PriceDispatch: %v", err)
		}
	}
	wantCost(t, dispatchCost(t, pool, id), 10)
}

// TestDispatchRowWithNoRateStaysUnpriced: a served model with no pricing
// row, or no tokens at all, leaves no cost_usd and is not an error; a cost an
// earlier pass wrote is dropped once the row can no longer be priced, so it
// never reads as complete.
func TestDispatchRowWithNoRateStaysUnpriced(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	ctx := context.Background()
	for _, c := range []struct{ name, metrics string }{
		{"no rate", `{"modelVersion":"claude-unrated",` + dispatchTokens[1:]},
		{"no tokens", `{"modelVersion":"claude-opus-5-5"}`},
		{"no metrics", ""},
		{"stale cost, no rate", `{"modelVersion":"claude-unrated","cost_usd":3,` + dispatchTokens[1:]},
	} {
		t.Run(c.name, func(t *testing.T) {
			id := record("task-"+c.name, c.metrics)
			if err := st.PriceDispatch(ctx, id); err != nil {
				t.Fatalf("PriceDispatch: %v", err)
			}
			if got := dispatchCost(t, pool, id); got != nil {
				t.Errorf("cost_usd = %v, want none", *got)
			}
		})
	}
}

// TestMergeDispatchMetricsPricesTheRow: the harvester's tokens patch prices
// the row as it lands, so inline and fixer rows carry their own cost.
func TestMergeDispatchMetricsPricesTheRow(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	id := record("task-1-implementer", `{"modelVersion":"claude-opus-5-5"}`)
	if err := st.MergeDispatchMetrics(context.Background(), id, json.RawMessage(dispatchTokens)); err != nil {
		t.Fatalf("MergeDispatchMetrics: %v", err)
	}
	wantCost(t, dispatchCost(t, pool, id), 10)
}

// TestRepriceSinceRepricesDispatchRows: a dispatch row whose cost_usd was
// written under a rate since removed is repriced from a date before it, and
// counts into priced; an unpriceable row counts into skipped.
func TestRepriceSinceRepricesDispatchRows(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	ctx := context.Background()
	id := record("task-1-implementer", `{"modelVersion":"claude-opus-5-5","cost_usd":99,`+dispatchTokens[1:])
	record("task-2-implementer", `{"modelVersion":"claude-unrated",`+dispatchTokens[1:])

	priced, skipped, err := st.RepriceSince(ctx, firstRateFrom)
	if err != nil {
		t.Fatalf("RepriceSince: %v", err)
	}
	// Skipped also holds the stage run RecordDispatch auto-opened, which
	// carries no tokens.
	if priced != 1 || skipped != 2 {
		t.Errorf("RepriceSince = priced %d, skipped %d, want 1, 2", priced, skipped)
	}
	wantCost(t, dispatchCost(t, pool, id), 10)
}

// TestRestampRepricesTheRow: both restamps replace a row's tokens, so the
// cost priced from the old tokens is replaced with the new tokens' cost.
func TestRestampRepricesTheRow(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	ctx := context.Background()
	const doubled = `{"tokens":{"main":{"output":2000000}}}`
	for _, c := range []struct {
		name    string
		restamp func(id int64) error
	}{
		{"agent", func(id int64) error { return st.RestampDispatch(ctx, id, "agent-x", json.RawMessage(doubled)) }},
		{"inline", func(id int64) error { return st.RestampInlineDispatch(ctx, id, json.RawMessage(doubled)) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			id := record("task-restamp-"+c.name, `{"modelVersion":"claude-opus-5-5",`+dispatchTokens[1:])
			if err := st.PriceDispatch(ctx, id); err != nil {
				t.Fatalf("PriceDispatch: %v", err)
			}
			if err := c.restamp(id); err != nil {
				t.Fatalf("restamp: %v", err)
			}
			wantCost(t, dispatchCost(t, pool, id), 20)
		})
	}
}

// TestDropDispatchPricesTheKeptRow: the tokens moved onto the kept row are
// priced there.
func TestDropDispatchPricesTheKeptRow(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	keep := record("task-keep", `{"modelVersion":"claude-opus-5-5"}`)
	drop := record("task-drop", `{"modelVersion":"claude-opus-5-5",`+dispatchTokens[1:])
	if err := st.DropDispatch(context.Background(), drop, keep); err != nil {
		t.Fatalf("DropDispatch: %v", err)
	}
	wantCost(t, dispatchCost(t, pool, keep), 10)
}

// TestBackfillModelVersionRepricesTheRow: a row its label could not price is
// priced once the backfill stamps the model that served it.
func TestBackfillModelVersionRepricesTheRow(t *testing.T) {
	st, pool, record := pricedDispatchStore(t)
	ctx := context.Background()
	id := record("task-backfill", dispatchTokens)
	agent := fmt.Sprintf("agent-backfill-%d", id)
	if _, err := pool.Exec(ctx, `UPDATE dispatches SET agent_id = $2 WHERE id = $1`, id, agent); err != nil {
		t.Fatalf("set agent: %v", err)
	}
	if _, err := st.BackfillModelVersions(ctx, func(a string) (string, bool) { return "claude-opus-5-5", a == agent }); err != nil {
		t.Fatalf("BackfillModelVersions: %v", err)
	}
	wantCost(t, dispatchCost(t, pool, id), 10)
}
