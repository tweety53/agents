# design — kan-479-harvest-zcode-rollout-transcripts

## Context

- Metrics come from one place: flowd harvesting Claude Code transcripts
  (`stats/internal/harvest/watcher.go` over `DefaultTranscriptsRoot`, Claude-shape parser in
  `transcript.go`). `internal/api/stages.go` stamps `tokens_available: false` at stage end for
  every run whose recorded harness is not `claudeCodeHarness`.
- Rollout format, measured against `~/.zcode/cli/rollout/model-io-sess_*.jsonl` (2026-09-09):
  every line `type: "model_io"`; `sessionId` is `sess_<uuid>`; `startedAt`/`completedAt` are
  RFC 3339 with milliseconds; `model.modelId` is `GLM-5.3-Flash` (a lowercase `glm-5.3-flash`
  spelling also appears in `request.body.model`); `response.usage` carries `inputTokens`,
  `outputTokens`, `totalTokens` (derived, ignored), `cacheReadTokens`, `cacheWriteTokens`;
  Bash commands sit in `request.messages[*].toolCalls[*].input.command` (history replayed on
  every request) and `response.toolCalls[*].input.command`.
- Auxiliary calls (title generation) are distinguishable only via `querySource` /
  `x-zcode-session-type` (`main_turn`/`main` vs `other`) — an unversioned enum.
- ZCode reports a single `cacheWriteTokens` with no 5m/1h split; the pricer refuses
  unknown-split cache writes (`chargeableTokens.cost`, task 23's rule) — correct where the two
  rates differ, silently omission-shaped where the model has one flat rate.
- Z.ai published rates (docs.z.ai/guides/overview/pricing, read 2026-09-09): GLM-5.3-Flash —
  input $0.15, cached-input read $0.03, cache-write storage limited-time free, output $0.50,
  all at a 50% launch discount ($0.075 / $0.015 / free / $0.25). No fast-mode rates published.
- `spectre/specs/` is empty; no capability spec edits. Historical rows: the harness stamp is
  immutable, so kan-252's mislabeled rows stay wrong; item "pipeline.md roster" prevents
  recurrence. Old zcode tokens already persisted as give-ups get a fresh bounded retry through
  the existing `scanRetriedTokens` path once rollouts are scanned. Historical
  `tokens_available: false` stamps on already-ended runs stay (metrics merge is additive; no
  history rewrite).

## Shape

### Harvest: one Watcher, many sources

- `internal/harvest` gains `Source` — `{Root string; ReadNew func(path string, offset int64)
  ([]Record, []CommandRecord, int64, error); ReadAllCmds func(path string) ([]CommandRecord,
  error)}` — and `Watcher` holds `[]Source` instead of `root string`. Discovery stays the one
  shared recursive `*.jsonl` walk per source root. `NewWatcher` takes the source list; flowd
  wires Claude + rollout sources into one Watcher, one goroutine, one set of token-binding
  bookkeeping.
- `DefaultZcodeRolloutRoot()`: `FLOW_ZCODE_ROLLOUTS_DIR` when set, else `~/.zcode/cli/rollout`
  — mirroring `DefaultTranscriptsRoot`/`FLOW_TRANSCRIPTS_DIR`; like it, deliberately not
  workspace-isolated.
- Rollout parser: one line = one API call = one `Record`. `Timestamp` = `completedAt` (the
  response-arrival instant — what a Claude transcript line's timestamp represents; half-open
  window semantics unchanged). `SessionID` = the line's `sessionId`. `Model` = `model.modelId`
  verbatim at the `Record`; the **bucket key** derived from it is canonicalized lowercase at
  attribution (fix 1, decision `model-id-canonical-lowercase`). `Usage`: `inputTokens`→Input, `outputTokens`→Output, `cacheReadTokens`→CacheRead,
  `cacheWriteTokens`→CacheCreation with `CacheSplitKnown=false` (lands in
  `cache_creation_unknown` — real measured spend, visible to the pricer). No `message.id`
  dedup: one line is one response. `Effort`/`Speed`/`AgentID`/`IsSidechain` are always zero —
  the format carries none.
- Commands from both tool-call locations where `name == "Bash"`, same `CommandRecord` shape.
  Same tolerance rules as the Claude parser: unrecognised shapes skip, never error; a partial
  trailing line never counts as consumed.

### Attribution and binding — unchanged

- Attribution, token binding, batch withholding, give-up bookkeeping: all source-agnostic,
  untouched. `BindSession` sets `session_id = sess_<uuid>` on runs carrying the token; window
  matching works as today. Mark commands replayed in request history match
  `isSessionMarkCommand` unchanged; repeats come from the same session, so no false ambiguity.
- `scanRetriedTokens` calls the per-source `ReadAllCmds`, so rollout files join the
  retried-give-up recovery scan. Existing rollout files are read from offset 0 on first scan;
  usage only lands where a stage window matches.

### Pricing

- `chargeableTokens.cost`: `cache_creation_unknown` becomes priceable when the rate carries a
  single cache-write rate in effect — `CacheWrite1hPerMTok` nil, or non-nil and equal to the
  5m rate — priced at that one rate (exact, never a guess). Where the rates differ (every
  seeded Claude row), the refusal stands unchanged.
- Seed row: `glm-5.3-flash` at the discounted plan rates — input 0.075, cache-read 0.015,
  cache-write 0 (limited-time free), output 0.25 — `cache_write_1h` nil (Z.ai publishes one
  cache rate). Model key is the canonical lowercase form of `model.modelId` (fix 1, decision
  `model-id-canonical-lowercase`). No `GLM-5.3` row: unmeasured on
  this machine; a later row is an insert.

### API and contract

- `internal/api/stages.go`: `claudeCodeHarness` becomes a `harvestedHarnesses` set
  (`claude-code`, `zcode`) consulted at the same `openRun.Harness` point; end marks for those
  harnesses stop stamping `tokens_available: false`.
- `skills/flow-contracts/pipeline.md`: the `-harness` roster sentence gains `zcode`.

## Decisions

### One Watcher over a list of sources, not a second Watcher

**ID:** one-watcher-many-sources
**Status:** active
**Chosen:** `Watcher` holds `[]Source`; flowd wires two sources into one Watcher — the
per-source read abstraction is required either way, and one goroutine keeps one set of
token-binding bookkeeping.
**Considered:** a second `harvest.Watcher` instance over the rollout root — needs the same
abstraction (the parser is hardcoded in the read path), plus duplicated binding bookkeeping
and a second loop to reason about; ruled out.

### Attribute every rollout line; no `querySource` filter

**ID:** no-auxiliary-call-filter
**Status:** active
**Chosen:** harvest every `model_io` line; the residual — an auxiliary call (title
generation) inside an open stage window attributes there, bounded by title-generation-sized
usage — is stated here rather than handled in code.
**Considered:** filtering on `querySource == "main_turn"` — keys harvest on an unversioned
enum whose rename would silently zero every zcode figure; ruled out.

### Per-dispatch figures stay Claude-only

**ID:** dispatch-grain-claude-only
**Status:** active
**Chosen:** stage-run grain fully works for zcode; `DispatchAttributor` finds nothing for
zcode sessions and the limitation is stated, not silently absent.
**Considered:** inferring zcode dispatch windows from tool-call shapes — no stable agent
identifier in the rollout format to key on, no measured demand; ruled out.

### Discounted GLM rates in the seed row

**ID:** glm-discounted-rates
**Status:** active
**Chosen:** input 0.075 / cache-read 0.015 / cache-write 0 / output 0.25 — what this machine's
coding plan actually charges; a later rate change is a new `effective_from` row, never a
migration.
**Considered:** list rates ($0.15/$0.03/free/$0.50) — immune to the discount ending but
overstate cost ~2× while it runs; ruled out for this machine's store.

### Model bucket keys are canonical lowercase

**ID:** model-id-canonical-lowercase
**Status:** active
**Chosen:** the models-bag key minted at attribution is `strings.ToLower(r.Model)` — one
canonical bucket per model regardless of how the harness spells the id, and the pricing seed
keys its row to the same canonical form. Rationale: measured rollout files carry both
`GLM-5.3-Flash` and `glm-5.3-flash` as `modelId` within one session; verbatim keys split the
usage and leave the lowercase share unpriced, which withholds the run's top-level cost
entirely.
**Considered:** verbatim keys + a case-insensitive pricing lookup — prices correctly but the
per-model view still shows two rows for one model; ruled out. Normalizing at the parsers —
would make `Record.Model` lie about the source bytes; ruled out.

### Unknown-split cache writes priced at a flat rate only

**ID:** unknown-split-flat-rate
**Status:** active
**Chosen:** price `cache_creation_unknown` when the rate row's cache-write columns agree on
one value — 1h == 5m, or 1h nil with the collapsed legacy column equal to the 5m rate (the row
was published flat); refuse otherwise. Amended from the original "1h nil or 1h == 5m" during
implementation: a pre-0007 row carrying only the collapsed column (its zero 5m rate an unset,
not a published free) must still refuse, which is what the existing store-level pin
`TestPriceDispatchWithUnknownCacheSplitGetsNoCost` protects.
**Considered:** leaving cache-write cost omitted for zcode — the exact "silently omitted"
outcome the ticket names; ruled out.

## Open questions

### Operator design approval

**ID:** operator-design-approval
**Status:** answered by the operator's 2026-09-09 fix-run instructions
**Why it is open:** the merged convergence-and-approval confirm was asked twice and no answer
came back through the channel; the run proceeded on the operator's written run instructions
and the harness's continue directive. The five decisions above and every section are exactly
what the operator reviews at the `IN_PROGRESS` gate.
**What it affects:** a revise answer at review becomes a `/flow <fix>` round against the
relevant decision — most cheaply `glm-discounted-rates` (a seed-row edit).
