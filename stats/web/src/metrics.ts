// metrics.ts reads the myflow-run-telemetry metrics bag (design.md, "The
// metrics bag") out of a StageRunDTO's `metrics` field -- deliberately typed
// `unknown` in api.ts, since the bag is an open JSONB document with no
// closed schema, and typing it as an interface there would re-close what
// the store went out of its way to leave open.
//
// **Every reader here returns `null` for a key the bag does not carry, for
// a key whose value is JSON `null`, and for a key whose value has the
// wrong JSON type -- never `0`, and never a coerced value.** This is the
// absence-is-never-zero rule (tasks.md's own Global Constraints, this
// task's single most important rule) at the last layer that can still
// break it: a run with wall-clock time but no token metrics must read as
// "unavailable" for every token figure, not as a measured zero.
//
// Token figures live under one nested `tokens` object (design.md: "every
// token figure lives under one tokens object"); the top-level readers
// (cost, model, effort, fast_mode) read the bag directly.

/**
 * Narrows an arbitrary JSON value to a plain object, or null if it isn't
 * one. Exported (F10, pass 2 of this change's own review panel) so
 * StageRunTable.tsx's per-dispatch and per-model readers -- which narrow
 * a nested object one level down from the whole metrics bag this file's
 * own readers work on, not a different shape -- share this single
 * definition instead of carrying a byte-for-byte duplicate under a
 * second name.
 */
export function asObject(value: unknown): Record<string, unknown> | null {
  if (value === null || value === undefined || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  return value as Record<string, unknown>;
}

function readNumberKey(obj: Record<string, unknown> | null, key: string): number | null {
  if (!obj) return null;
  const v = obj[key];
  if (v === null || v === undefined || typeof v !== "number") return null;
  return v;
}

function readStringKey(obj: Record<string, unknown> | null, key: string): string | null {
  if (!obj) return null;
  const v = obj[key];
  if (v === null || v === undefined || typeof v !== "string") return null;
  return v;
}

function readBooleanKey(obj: Record<string, unknown> | null, key: string): boolean | null {
  if (!obj) return null;
  const v = obj[key];
  if (v === null || v === undefined || typeof v !== "boolean") return null;
  return v;
}

function tokensObject(bag: unknown): Record<string, unknown> | null {
  const obj = asObject(bag);
  if (!obj) return null;
  return asObject(obj.tokens);
}

/**
 * `tokens.<bucketKey>` (`main` or `sidechain`), narrowed to an object --
 * or null if the parent `tokens` object or the bucket itself is absent or
 * malformed. `internal/harvest.TokenDelta` (attribute.go) writes both
 * buckets as objects (`{input, output, cache_creation, cache_read,
 * thinking, ...}`), never as flat numbers -- there is no top-level
 * `tokens.main` or `tokens.input` scalar for any real bag.
 */
function bucketObject(bag: unknown, bucketKey: "main" | "sidechain"): Record<string, unknown> | null {
  const tokens = tokensObject(bag);
  if (!tokens) return null;
  return asObject(tokens[bucketKey]);
}

/**
 * Sums one field (`input`, `output`, `cache_read` or `cache_creation`)
 * across both the `main` and `sidechain` buckets -- `internal/harvest`'s
 * own per-field total, mirrored from `internal/store/aggregate.go`'s
 * `COALESCE(main->>key, 0) + COALESCE(sidechain->>key, 0)`. Absence is
 * never zero: if the field is present on neither bucket, this returns
 * null rather than 0 -- a run that recorded no input tokens at all reads
 * distinctly from a run whose input tokens were recorded and happened to
 * total zero. A field recorded on only one bucket sums to that bucket's
 * own value, since a bucket that never touched the run correctly
 * contributes nothing.
 */
function readSummedBucketField(bag: unknown, key: string): number | null {
  const main = readNumberKey(bucketObject(bag, "main"), key);
  const sidechain = readNumberKey(bucketObject(bag, "sidechain"), key);
  if (main === null && sidechain === null) return null;
  return (main ?? 0) + (sidechain ?? 0);
}

// BUCKET_TOTAL_FIELDS is every field readBucketTotal sums to produce one
// bucket's own grand total -- internal/harvest.Bucket's shape (Input,
// Output, CacheCreation, CacheRead, Thinking), matching
// harvestshape_test.go's wantMain/wantSidechain exactly. CacheCreation5m,
// CacheCreation1h and CacheCreationUnknown are deliberately excluded: the
// collapsed CacheCreation total already includes their contribution
// (Bucket's own doc comment, attribute.go), so summing both would
// double-count.
const BUCKET_TOTAL_FIELDS = ["input", "output", "cache_creation", "cache_read", "thinking"];

/**
 * Sums every one of a single bucket's (`main` or `sidechain`) own fields
 * into one grand total -- the bucket's whole contribution to the run,
 * regardless of which kind of token it was. Absence is never zero: a
 * present-but-empty bucket, or one whose fields never got past
 * `readNumberKey`, returns null rather than 0.
 */
function readBucketTotal(bag: unknown, bucketKey: "main" | "sidechain"): number | null {
  const bucket = bucketObject(bag, bucketKey);
  if (!bucket) return null;
  let sum = 0;
  let any = false;
  for (const field of BUCKET_TOTAL_FIELDS) {
    const v = readNumberKey(bucket, field);
    if (v !== null) {
      sum += v;
      any = true;
    }
  }
  return any ? sum : null;
}

/** `cost_usd`, computed from pricing at write time. */
export function readCostUsd(bag: unknown): number | null {
  return readNumberKey(asObject(bag), "cost_usd");
}

/**
 * The distinct models a stage run used, read from the `models` bucket
 * `internal/harvest/attribute.go`'s MetricsPatch actually writes
 * (`models.<model>.tokens`, one entry per model that touched the run) --
 * never the top-level scalar `model`, which task 22 retired and nothing
 * has written since. Returns every key, sorted, rather than one: a mixed-
 * model run (the common case for a review-panel stage) has more than one,
 * and reporting only one of them is the exact misattribution this round
 * exists to fix, resurfacing here if this reader picked a "first" or
 * "last" model instead. Returns null when the bucket is absent, present
 * with no keys, or the wrong JSON type -- never an empty array, so a
 * caller's null-check and empty-check cannot silently diverge.
 */
export function readModels(bag: unknown): string[] | null {
  const obj = asObject(bag);
  if (!obj) return null;
  const models = asObject(obj.models);
  if (!models) return null;
  const keys = Object.keys(models);
  if (keys.length === 0) return null;
  return keys.sort();
}

/** `effort`, from the transcript's `effort`. */
export function readEffort(bag: unknown): string | null {
  return readStringKey(asObject(bag), "effort");
}

/** `fast_mode`, from the CLI invocation's environment. */
export function readFastMode(bag: unknown): boolean | null {
  return readBooleanKey(asObject(bag), "fast_mode");
}

/**
 * `input`, summed across `tokens.main` and `tokens.sidechain` --
 * `internal/harvest.TokenDelta` writes both as `Bucket` objects, never a
 * flat `tokens.input` scalar, so this must read the nested shape rather
 * than the top-level key its name might suggest.
 */
export function readInputTokens(bag: unknown): number | null {
  return readSummedBucketField(bag, "input");
}

/** `output`, summed across `tokens.main` and `tokens.sidechain`. */
export function readOutputTokens(bag: unknown): number | null {
  return readSummedBucketField(bag, "output");
}

/** `cache_read`, summed across `tokens.main` and `tokens.sidechain`. */
export function readCacheReadTokens(bag: unknown): number | null {
  return readSummedBucketField(bag, "cache_read");
}

/** `cache_creation`, summed across `tokens.main` and `tokens.sidechain`. */
export function readCacheCreationTokens(bag: unknown): number | null {
  return readSummedBucketField(bag, "cache_creation");
}

/**
 * `tokens.main`'s own grand total, from the transcript split on
 * `isSidechain` -- `internal/harvest.TokenDelta` writes `main` as a
 * `Bucket` object (`{input, output, cache_creation, cache_read,
 * thinking}`), never a flat number, so this sums the bucket's own fields
 * rather than reading a scalar that no real bag carries.
 */
export function readMainTokens(bag: unknown): number | null {
  return readBucketTotal(bag, "main");
}

/** `tokens.sidechain`'s own grand total -- readMainTokens's counterpart. */
export function readSidechainTokens(bag: unknown): number | null {
  return readBucketTotal(bag, "sidechain");
}

/**
 * Every top-level key in the bag this file does not otherwise expose a
 * dedicated reader for, so an unrecognised metric (a key design.md's table
 * does not name, or one added by a later myflow change without a matching
 * reader here) is still displayed rather than silently dropped. Returns an
 * empty object for an absent or malformed bag, never null -- callers spread
 * this directly into a detail view.
 */
const KNOWN_TOP_LEVEL_KEYS = new Set(["cost_usd", "model", "models", "effort", "fast_mode", "tokens", "dispatches"]);

export function readOtherMetrics(bag: unknown): Record<string, unknown> {
  const obj = asObject(bag);
  if (!obj) return {};
  const out: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    if (KNOWN_TOP_LEVEL_KEYS.has(key)) continue;
    out[key] = value;
  }
  return out;
}
