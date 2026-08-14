// Readers for the metrics bag's known keys -- one test per absence, wrong
// type and unknown-key case, per this task's own non-negotiable rule:
// every reader returns null for an absent key and a null for a key
// present with the wrong JSON type, never a coerced 0. This is the
// absence-is-never-zero rule at the last layer that can still break it.
import { describe, expect, it } from "vitest";
import {
  readCacheCreationTokens,
  readCacheReadTokens,
  readCostUsd,
  readEffort,
  readFastMode,
  readInputTokens,
  readMainTokens,
  readModels,
  readOtherMetrics,
  readOutputTokens,
  readSidechainTokens,
} from "./metrics";
// The one committed copy of the wire shape internal/harvest.MetricsPatch
// actually marshals -- generated from that Go type by
// internal/harvest/wireshape_test.go's TestSPAMetricsFixtureMatchesTheWireShape,
// never hand-written here. Regenerate it from the Go side
// (UPDATE_METRICS_FIXTURE=1 go test ./internal/harvest/ -run
// TestSPAMetricsFixtureMatchesTheWireShape) rather than editing this file
// directly -- editing it directly would recreate exactly the private-copy
// problem this fixture exists to close (task 24's lesson, generalised: see
// wireshape_test.go's own doc comment).
import metricsPatchFixture from "./testdata/metricsPatchFixture.json";

describe("readCostUsd", () => {
  it("reads a present numeric cost_usd", () => {
    expect(readCostUsd({ cost_usd: 4.5 })).toBe(4.5);
  });

  it("returns null when cost_usd is absent", () => {
    expect(readCostUsd({})).toBeNull();
  });

  it("returns null when cost_usd is JSON null", () => {
    expect(readCostUsd({ cost_usd: null })).toBeNull();
  });

  it("returns null when cost_usd holds a string instead of a number", () => {
    expect(readCostUsd({ cost_usd: "4.5" })).toBeNull();
  });

  it("returns null when the bag itself is not an object", () => {
    expect(readCostUsd(null)).toBeNull();
    expect(readCostUsd(undefined)).toBeNull();
    expect(readCostUsd("not-an-object")).toBeNull();
  });
});

// readModels reads the "models" bucket the harvester actually writes
// (internal/harvest/attribute.go's MetricsPatch/ModelBucket: a stage run's
// metrics bag carries `models.<model>.tokens`, keyed by every model that
// touched the run -- never a single top-level scalar "model"). This
// fixture is built from that real shape, not from the retired scalar the
// old readModel used to read: metrics.test.ts's own private-copy problem
// (task 24's own rationale, harvestshape_test.go) is exactly a fixture
// that encodes a shape nothing writes passing against code that reads it.
describe("readModels", () => {
  it("reads a single model as a one-element array", () => {
    expect(readModels({ models: { "claude-sonnet-5": { tokens: { main: { input: 100 } } } } })).toEqual([
      "claude-sonnet-5",
    ]);
  });

  it("reports every model of a mixed-model run, not just one -- the exact misattribution this round exists to fix", () => {
    expect(
      readModels({
        models: {
          "claude-sonnet-5": { tokens: { main: { input: 100 } } },
          "claude-opus-5": { tokens: { main: { input: 200 } } },
        },
      }),
    ).toEqual(["claude-opus-5", "claude-sonnet-5"]);
  });

  it("returns null when the bag has no models bucket at all", () => {
    expect(readModels({})).toBeNull();
  });

  it("returns null when models is present but carries no keys", () => {
    expect(readModels({ models: {} })).toBeNull();
  });

  it("returns null when models holds the wrong JSON type", () => {
    expect(readModels({ models: "claude-sonnet-5" })).toBeNull();
    expect(readModels({ models: null })).toBeNull();
    expect(readModels({ models: ["claude-sonnet-5"] })).toBeNull();
  });

  it("ignores the retired top-level scalar 'model' key -- nothing writes it, and the models bucket is the only source now", () => {
    expect(readModels({ model: "claude-sonnet-5" })).toBeNull();
  });
});

describe("readEffort", () => {
  it("reads a present string effort", () => {
    expect(readEffort({ effort: "high" })).toBe("high");
  });

  it("returns null when effort is absent", () => {
    expect(readEffort({})).toBeNull();
  });

  it("returns null when effort holds the wrong type", () => {
    expect(readEffort({ effort: 1 })).toBeNull();
  });
});

describe("readFastMode", () => {
  it("reads a present boolean fast_mode", () => {
    expect(readFastMode({ fast_mode: true })).toBe(true);
    expect(readFastMode({ fast_mode: false })).toBe(false);
  });

  it("returns null when fast_mode is absent", () => {
    expect(readFastMode({})).toBeNull();
  });

  it("returns null when fast_mode holds the wrong type", () => {
    expect(readFastMode({ fast_mode: "true" })).toBeNull();
  });
});

// The token readers all read out of the nested "tokens" object -- but not
// as flat scalars. internal/harvest.TokenDelta (attribute.go) writes
// "tokens" as {main: Bucket, sidechain: Bucket}, each Bucket an object of
// its own ({input, output, cache_creation, cache_read, thinking, ...}),
// so there is no flat tokens.input/tokens.main for any bag a real harvest
// batch ever produces. metricsPatchFixture is that real shape, sourced
// from internal/harvest itself (see the import's doc comment) rather than
// hand-written here -- this is F2's own fix: a hand-written flat literal
// passed against readers that could never work against a real bag, and
// nobody noticed because the fixture and the readers shared the same
// private, wrong assumption.
describe("token readers", () => {
  const bag = { tokens: metricsPatchFixture.tokens };
  const main = metricsPatchFixture.tokens.main;
  const sidechain = metricsPatchFixture.tokens.sidechain;

  it("sum a present numeric field across both buckets, for each per-field key", () => {
    expect(readInputTokens(bag)).toBe(main.input + sidechain.input);
    expect(readOutputTokens(bag)).toBe(main.output + sidechain.output);
    expect(readCacheReadTokens(bag)).toBe(main.cache_read + sidechain.cache_read);
    expect(readCacheCreationTokens(bag)).toBe(main.cache_creation + sidechain.cache_creation);
  });

  it("sum each bucket's own fields into that bucket's grand total", () => {
    expect(readMainTokens(bag)).toBe(main.input + main.output + main.cache_creation + main.cache_read + main.thinking);
    expect(readSidechainTokens(bag)).toBe(
      sidechain.input + sidechain.output + sidechain.cache_creation + sidechain.cache_read + sidechain.thinking,
    );
  });

  it("sum only the bucket that carries a field, when the other bucket never recorded it", () => {
    expect(readInputTokens({ tokens: { main: { input: 7 } } })).toBe(7);
    expect(readInputTokens({ tokens: { sidechain: { input: 9 } } })).toBe(9);
  });

  it("return null for a per-field key absent from both buckets, not a coerced 0", () => {
    expect(readInputTokens({ tokens: { main: { output: 1 }, sidechain: { output: 2 } } })).toBeNull();
  });

  it("return null when a bucket's field is JSON null", () => {
    expect(readInputTokens({ tokens: { main: { input: null } } })).toBeNull();
  });

  it("return null when a bucket's field holds the wrong type", () => {
    expect(readInputTokens({ tokens: { main: { input: "100" } } })).toBeNull();
  });

  it("return null for a bucket total when that bucket is absent, even if the other bucket is present", () => {
    expect(readMainTokens({ tokens: { sidechain: { input: 1 } } })).toBeNull();
    expect(readSidechainTokens({ tokens: { main: { input: 1 } } })).toBeNull();
  });

  it("return null for every token reader when the nested tokens object is missing entirely -- a run with wall-clock time but no token metrics", () => {
    expect(readInputTokens({})).toBeNull();
    expect(readOutputTokens({})).toBeNull();
    expect(readCacheReadTokens({})).toBeNull();
    expect(readCacheCreationTokens({})).toBeNull();
    expect(readMainTokens({})).toBeNull();
    expect(readSidechainTokens({})).toBeNull();
  });

  it("return null when tokens itself is not an object", () => {
    expect(readInputTokens({ tokens: "not-an-object" })).toBeNull();
  });

  it("return null when a bucket is present but carries no recognised fields at all", () => {
    expect(readMainTokens({ tokens: { main: {} } })).toBeNull();
  });
});

describe("readOtherMetrics", () => {
  it("passes through a key this build does not recognise", () => {
    expect(readOtherMetrics({ cost_usd: 1, unknown_future_key: "surprise" })).toEqual({
      unknown_future_key: "surprise",
    });
  });

  it("omits every key with a dedicated reader", () => {
    const bag = {
      cost_usd: 1,
      model: "claude-sonnet-5",
      effort: "high",
      fast_mode: true,
      tokens: { input: 1 },
      pricing_version: "v3",
    };
    expect(readOtherMetrics(bag)).toEqual({ pricing_version: "v3" });
  });

  it("returns an empty object for an absent or malformed bag", () => {
    expect(readOtherMetrics(null)).toEqual({});
    expect(readOtherMetrics(undefined)).toEqual({});
    expect(readOtherMetrics("not-an-object")).toEqual({});
  });
});
