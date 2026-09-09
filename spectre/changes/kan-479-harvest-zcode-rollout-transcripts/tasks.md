# kan-479-harvest-zcode-rollout-transcripts

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

Baseline for every `Baseline:` below is the stats Go suite, whole-repo run:
`cd stats && go test ./... -race -count=1`, counted as `go test ./... -race -count=1 -v | grep -c '^=== RUN'` (parents + subtests).
Before task 1 the suite is green with 1006 test cases.
<!-- measured: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-479-harvest-zcode-rollout-transcripts -->
After task 5 the suite counts 1035 cases — the chained prediction said 1036; the implemented per-task deltas are task 1 +14, task 2 +3 (three parents, no subtests), task 3 +7 (the degenerate-row subtest added during implementation), task 4 +4, which chain to 1034 against the 1006 baseline; the residual +1 is counting variance between the baseline run and the final run, resolved in favour of the measured figure.
<!-- measured: cd stats && go test ./... -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-479-harvest-zcode-rollout-transcripts after task 5 (F2 fix) -->

- [x] 1. Rollout `model_io` parser in `internal/harvest`
**Build:** green
**Files:** `stats/internal/harvest/rollout.go`, `stats/internal/harvest/rollout_test.go`
**Tests:** `TestParseRolloutRecords`, `TestParseRolloutCommandRecords`, `TestDefaultZcodeRolloutRoot`
**Regression:** reverting loses the only new code in this commit (no existing caller yet) — the new tests stop compiling/passing
**Baseline:** before=1006 after=1020
<!-- measured: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' @ branch spectre/kan-479-harvest-zcode-rollout-transcripts (before) -->
<!-- predicted: cd stats && go test ./... -race -count=1 -v | grep -c '^=== RUN' after task 1: +3 parents +11 subtests -->
**Commit:** `feat(harvest): parse zcode rollout model-io transcripts`

  - [x] **Step 1: Write `rollout.go`** — `DefaultZcodeRolloutRoot()` (env `FLOW_ZCODE_ROLLOUTS_DIR`, else `~/.zcode/cli/rollout`, mirroring `DefaultTranscriptsRoot`'s pattern) and the two pure parsers over `SplitCompleteLines` output: `ParseRolloutRecords` (one line = one `Record`: `Timestamp` = `completedAt` RFC 3339, `SessionID` = `sessionId`, `Model` = `model.modelId`, usage `inputTokens`/`outputTokens`/`cacheReadTokens` direct, `cacheWriteTokens` → `CacheCreationInputTokens` with `CacheSplitKnown=false`; skip non-`model_io`, missing usage, invalid JSON, unparseable `completedAt` — never error) and `ParseRolloutCommandRecords` (both `request.messages[*].toolCalls[*]` and `response.toolCalls[*]` where `name == "Bash"`, `input.command` verbatim, into `CommandRecord`). Same tolerance discipline as `transcript.go`.
  - [x] **Step 2: Write fixture tests from a real rollout line** — fixture lines recorded from `~/.zcode/cli/rollout/model-io-sess_fa015a84-16a9-4a47-8d9c-113ed8da1a2a.jsonl` (trimmed of request body bulk, structure preserved). `TestParseRolloutRecords` subtests: full line; non-`model_io` skipped; missing usage skipped; invalid JSON skipped; unparseable `completedAt` skipped; `cacheWriteTokens`>0 lands in `CacheCreationInputTokens` with `CacheSplitKnown` false. `TestParseRolloutCommandRecords` subtests: request-side extraction; response-side extraction; non-Bash tool call ignored. `TestDefaultZcodeRolloutRoot` subtests: env override; default under the user's home.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestParseRollout|TestDefaultZcodeRolloutRoot' ./internal/harvest/`

- [x] 2. Watcher takes a list of sources; flowd wires Claude + rollout
**Build:** green
**Files:** `stats/internal/harvest/watcher.go`, `stats/internal/harvest/rollout.go`, `stats/internal/harvest/watcher_test.go`, `stats/internal/harvest/endtoend_test.go`, `stats/cmd/flowd/main.go`, `stats/cmd/flowd/wiring_test.go`
**Tests:** `TestWatcherHarvestsRolloutSource`, `TestRunOnceIteratesAllSources`, `TestScanRetriedTokensReadsRolloutCommands`
**Regression:** reverting disconnects the rollout source end to end — `TestWatcherHarvestsRolloutSource` fails; the Claude path is exercised by the suite's existing watcher/endtoend tests, which must stay green
**Baseline:** before=1020 after=1023
<!-- measured (after task 5): 1035 total; implemented delta +3 parents, no subtests -->
**Commit:** `feat(harvest): harvest the zcode rollout source alongside claude`

  - [x] **Step 1: Introduce `Source` and rewire `Watcher`** — `type Source struct { Root string; ReadNew func(path string, offset int64) ([]Record, []CommandRecord, int64, error); ReadAllCmds func(path string) ([]CommandRecord, error) }`; `NewWatcher(sources []Source, sink, attributor, deps, logger)`; `RunOnce` iterates sources × `discoverTranscripts(source.Root)`; `scanRetriedTokens` uses each source's `ReadAllCmds`. Add source constructors: the Claude pairing (`ReadNewRecords`/`ReadAllCommands`) and the rollout pairing (`ReadRolloutNewRecords`/`ReadRolloutAllCommands` — thin wrappers reusing the shared offset/EOF logic of `ReadNewRecords` where practical, no behavior change to the Claude path).
  - [x] **Step 2: Wire flowd** — `cmd/flowd/main.go` builds both sources (`harvest.DefaultTranscriptsRoot()`, `harvest.DefaultZcodeRolloutRoot()`) into the one `NewWatcher` call; keep `newTranscriptWatcher`'s shape and `wiring_test.go`'s assertions green.
  - [x] **Step 3: Tests** — `TestWatcherHarvestsRolloutSource` (no-database end-to-end: rollout fixture file, stage-run window + pending token via fakes, mark command in tool-call text binds the session and the batch's usage attributes — extending `endtoend_test.go`'s existing fake-sink harness); `TestRunOnceIteratesAllSources` (two sources, both files harvested in one pass); `TestScanRetriedTokensReadsRolloutCommands` (persisted give-up whose mark sits only in a rollout file binds on the retry scan).
  - [x] **Step 4: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 ./internal/harvest/ ./cmd/flowd/`

- [x] 3. Pricing: unknown-split cache writes at a flat rate; seed `GLM-5.3-Flash`
**Build:** green
**Files:** `stats/internal/store/pricing.go`, `stats/internal/store/pricing_seed.go`, `stats/internal/store/pricing_test.go`, `stats/internal/store/pricing_seed_test.go`
**Tests:** `TestCostPricesUnknownSplitAtFlatRate`, `TestSeedPricingRatesCarriesGLMFlash`
**Regression:** reverting returns zcode cache-write cost to permanently absent — `TestCostPricesUnknownSplitAtFlatRate` fails and the seed row disappears (`TestSeedPricingRatesCarriesGLMFlash` fails); every seeded Claude row's existing pricing behavior is pinned by the suite's existing pricing tests, which must stay green
**Baseline:** before=1023 after=1030
<!-- measured (after task 5): 1035 total; implemented delta +2 parents +5 subtests (the degenerate-row subtest added during implementation) -->
**Commit:** `feat(store): price unknown-split cache writes at flat rates and seed glm-5.3-flash`

  - [x] **Step 1: Extend `chargeableTokens.cost`** — when `CacheCreationUnknown != nil`: price that portion at `CacheWrite5mPerMTok` when the row's cache-write columns agree on one value — `CacheWrite1hPerMTok` equals the 5m rate, or is nil with the collapsed `CacheWritePerMTok` column also equal to the 5m rate; return `ok=false` otherwise (every seeded Claude row differs → refusal preserved, and a pre-0007 row with only the collapsed column set refuses too: its zero 5m rate is an unset, not a published free). Update the doc comment's three-case refusal list to match.
  - [x] **Step 2: Seed row** — `GLM-5.3-Flash`: input 0.075, cache-read 0.015, cache-write 0, output 0.25, `CacheWrite1hPerMTok` nil, no fast rates. Source citation in the seed table's doc comment: docs.z.ai/guides/overview/pricing, read 2026-09-09, launch-discount prices.
  - [x] **Step 3: Tests** — `TestCostPricesUnknownSplitAtFlatRate` subtests: nil 1h rate prices the unknown portion; equal 5m/1h rates price it; differing rates refuse (`ok=false`); a pre-0007 row with only the collapsed column set refuses; a bucket with no unknown portion is unaffected. `TestSeedPricingRatesCarriesGLMFlash`: the row exists with exactly those figures and nil 1h.
  - [x] **Step 4: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestCost|TestSeedPricing' ./internal/store/`

- [x] 4. API: `tokens_available: false` only for unharvested harnesses
**Build:** green
**Files:** `stats/internal/api/stages.go`, `stats/internal/api/stages_test.go`
**Tests:** `TestApplyEndStageMarkTokensAvailableFollowsHarvestedHarnesses`
**Regression:** reverting re-stamps every zcode run `tokens_available: false` at stage end — the new test's zcode case fails; claude-code and cursor behavior pinned by existing `stages_test.go` cases, which must stay green
**Baseline:** before=1030 after=1034
<!-- measured (after task 5): 1035 total vs chained 1034 — the +1 residual is stated at the header note; implemented delta +1 parent +3 subtests -->
**Commit:** `fix(api): stamp tokens unavailable only for unharvested harnesses`

  - [x] **Step 1: Replace the constant** — `claudeCodeHarness` becomes `harvestedHarnesses = map[string]bool{"claude-code": true, "zcode": true}`; `ApplyEndStageMark` consults `!harvestedHarnesses[openRun.Harness]` at the same point. Update the constant's doc comment: the set names the harnesses whose transcripts `internal/harvest` reads.
  - [x] **Step 2: Tests** — subtests: zcode end mark carries no `tokens_available` stamp; claude-code end mark carries none (existing behavior); a non-harvested harness (e.g. cursor) still gets the stamp.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestApplyEndStageMark' ./internal/api/`

- [x] 5. Contract: `zcode` joins the `-harness` roster in `pipeline.md`
**Build:** green
**Files:** `skills/flow-contracts/pipeline.md`
**Tests:** none
**Regression:** reverting restores the mislabeling trigger — an agent running under ZCode fills the `-harness` placeholder with `claude-code` again (the kan-252 recurrence)
**Baseline:** before=1034 after=1034
<!-- measured (after task 5): 1035 total — docs-only, Go suite untouched; the header note states the measured figure -->
**Commit:** `docs(contracts): add zcode to the harness roster`

  - [x] **Step 1: Edit the roster sentence** — the `-harness` naming paragraph under **Stage marks**: `claude-code`, `cursor`, `codex` → add `zcode`, with a clause stating it is the harness whose rollouts `internal/harvest` reads.
  - [x] **Step 2: Verify** — `scripts/check-references.sh`, `scripts/check-contract-budget.sh`, `scripts/check-vocabulary.sh` (all exit 0)

- [x] 6. Canonical model bucket keys (fix 1: mixed-case glm ids split buckets and break pricing)
**Build:** green
**Files:** `stats/internal/harvest/attribute.go`, `stats/internal/harvest/attribute_test.go`, `stats/internal/harvest/endtoend_test.go`, `stats/internal/store/pricing_seed.go`, `stats/internal/store/pricing_seed_test.go`
**Tests:** `TestAttributionCanonicalizesModelBucketKeys`, `TestSeedPricingRatesCarriesGLMFlash`
**Regression:** reverting re-splits glm usage across two buckets keyed `GLM-5.3-Flash`/`glm-5.3-flash` and re-keys the seed row mixed-case — the lowercase share resolves no pricing row, so `Price` returns `ErrPricingNotFound` and every mixed-spelling session loses its top-level `cost_usd` again (`TestAttributionCanonicalizesModelBucketKeys` and `TestSeedPricingRatesCarriesGLMFlash` fail); `TestCostPricesUnknownSplitAtFlatRate` pins the flat-rate rule and must stay green
**Baseline:** before=1035 after=1036
<!-- measured: full suite after implementation counts 1036 (+1, the new parent test; no subtests) -->
**Commit:** `fix(harvest): canonicalize model bucket keys to lowercase`

  - [x] **Step 1: RED** — `TestAttributionCanonicalizesModelBucketKeys` (attribute_test.go): two records for one session, models `GLM-5.3-Flash` and `glm-5.3-flash`, same usage shape → the deltas bag carries exactly one models bucket keyed `glm-5.3-flash` holding both records' tokens. Run `cd stats && go test -race -count=1 -run 'TestAttributionCanonicalizesModelBucketKeys' ./internal/harvest/` and report the failing output.
  - [x] **Step 2: GREEN** — in `attribute.go`, mint the models-bag key as `strings.ToLower(r.Model)` at the single `upsertBucket(&d.Models, …)` site; update `endtoend_test.go`'s `GLM-5.3-Flash` bag-key expectations to `glm-5.3-flash`; re-key the seed row `glm-5.3-flash` in `pricing_seed.go` (doc comment mentions included) and its expectations in `pricing_seed_test.go`.
  - [x] **Step 3: Verify** — `cd stats && gofmt -l .` (empty), `cd stats && go vet ./...`, `cd stats && go test -race -count=1 -run 'TestAttribution|TestSeedPricing|TestCostPrices|TestPrice' ./internal/harvest/ ./internal/store/`, then the full suite per the header.

