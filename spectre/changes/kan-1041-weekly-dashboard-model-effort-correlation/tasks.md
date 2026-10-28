# kan-1041-weekly-dashboard-model-effort-correlation

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Tasks appended:** 1
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** Weekly costs gains Visual verify and Other phase lines, a second chart of cost by model · effort (dispatches by served model and effort, plus the parent session per model), and permanent `$cost ▲n%` labels on every dot; dispatch buckets get priced, and `flow pricing reprice` backfills the window.

**Spec:** `design.md` beside this file; the approved prototype `/Users/tweety53/Projects/agents-worktrees/kan-1041-weekly-dashboard-model-effort-correlation/.superpowers/sdd/weekly-model-effort-preview.html` (https://claude.ai/artifact/7TZTUL4b6gJqu1nHzX7i5j, version 2) is the layout and behaviour reference.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1041-weekly-dashboard-model-effort-correlation`. Every `**Files:**` path is relative to its root. Never touch the main checkout.
- Never stop the dev workspace's `flowd` or `flow-postgres`, and never write to the dev `flow` database; store tests use the test database their harness already creates.
- No new dependency; the charts stay inline SVG.
- Go lint per task: `cd stats && make -s web-dist && go vet ./... && gofmt -l .` must print nothing. SPA lint: `cd stats/web && npx tsc -b`.

- [x] 1. Store: price a dispatch by its served model
  - [x] **Step 1: Write the failing test** `TestPriceDispatchUsesServedModel` in `stats/internal/store/stageruns_test.go`, beside the existing `Price` tests there: a stage run whose `metrics` carry a `models` bucket for `claude-opus-5-5` and a `dispatches` bucket `{"a1": {"model": "opus", "tokens": {...sidechain output 1000...}, "signals": {"served_models": {"claude-opus-5-5": 3}, "served_efforts": {"low": 3}}}, "a2": {"model": "opus", "tokens": {...}}}`, with the seeded `claude-opus-5-5` rate; after `Price`, `dispatches.a1.cost_usd` equals the rate applied to its tokens and `dispatches.a2` (no served model, alias unpriced) has no `cost_usd`; a third bucket with `served_models` `{"claude-opus-5-5": 1, "claude-sonnet-5-5": 4}` prices at the sonnet rate. Run `cd stats && go test ./internal/store/ -run 'TestPriceDispatchUsesServedModel'` — RED.
  - [x] **Step 2: Implement** in `stats/internal/store/pricing.go`: decode `signals.served_models` on `dispatchBucket`; the dispatch loop resolves its rate from the most-turns key (ties by name), falling back to `db.Model`; the `db.Model == ""` skip applies only when both are empty. Update `Price`'s doc comment paragraph on dispatch pricing to say which model it resolves.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestPrice'`.

**Files:** `stats/internal/store/pricing.go`, `stats/internal/store/stageruns_test.go`
**Tests:** `TestPriceDispatchUsesServedModel`
**Regression:** `TestPriceDispatchUsesServedModel` fails, because the dispatch resolves its rate from the alias `opus`, which no pricing row matches, and gets no `cost_usd`.
**Baseline:** before=324 after=325 <!-- measured: grep -rhE '^func Test' stats/internal/store | wc -l @ a9f37e89 -->
**Build:** green
**After:** none
**Commit:** `fix(store): price a dispatch by the model that served it`

**Decision:** dispatch-served-model

- [x] 2. Store: Visual verify category and the model·effort weekly series
  - [x] **Step 1: Write the failing tests** in `stats/internal/store/aggregate_test.go`, on the fixtures `TestWeeklyCostsBucketsByWednesdayAndCategory` uses:
    - `TestWeeklyCostsCountsVisualVerify` — a `flow.visual-verify` run of $2 in one week yields `VisualVerifyUSD` $2 that week and nil in a week without one; a `self-review-*` change's `flow.visual-verify` run counts as self-review. A `flow.verify` run of \$1 yields `OtherUSD` \$1 and is not in any named phase; a week whose runs all fall in named phases has nil `OtherUSD`.
    - `TestWeeklyModelEffortCostsSplitsDispatchAndSession` — one week, one stage run whose `models.claude-opus-5-5.cost_usd` is $10 with two priced dispatches ($3 served `claude-opus-5-5`/`low`, $2 served `claude-sonnet-5-5`/`medium`) and one priced dispatch with no `served_efforts` ($1, `claude-opus-5-5`), plus a `models.claude-sonnet-5-5.cost_usd` of $2: rows (raw model ids) are `claude-opus-5-5/low $3`, `claude-opus-5-5/default $1`, `claude-opus-5-5/session $6`, `claude-sonnet-5-5/medium $2`; the zero `claude-sonnet-5-5/session` row is not returned; ordered by week, model, effort; a run outside the period contributes nothing.
    Run `cd stats && go test ./internal/store/ -run 'TestWeeklyCostsCountsVisualVerify|TestWeeklyModelEffortCostsSplitsDispatchAndSession'` — RED.
  - [x] **Step 2: Implement** in `stats/internal/store/aggregate.go`: `weeklyCostStages.visualVerify = "flow.visual-verify"`, a `visual-verify` CASE arm after self-review and the matching `SUM ... FILTER` column, `WeeklyCostPoint.VisualVerifyUSD`, and `OtherUSD` as `SUM(cost) FILTER (WHERE category IS NULL)`; and `func (s *Store) WeeklyModelEffortCosts(ctx context.Context, period Period) ([]WeeklyModelEffortPoint, error)` with `WeeklyModelEffortPoint{Week time.Time; Model, Effort string; CostUSD float64}` — one SQL statement: dispatch rows from `jsonb_each(metrics->'dispatches')` where the bucket has `cost_usd`, model the most-turns `served_models` key else the bucket `model`, effort the most-turns `served_efforts` key else `default`; session rows per stage run and `models` key with `cost_usd`, that cost minus the dispatch costs of the same model in the run, `GREATEST(0, …)`, effort `session`; summed per week, model, effort, rows with a zero sum dropped. Ties in most-turns by key name, as `dispatch-served-model`. Doc comments cite the decisions. Extend `TestWeeklyCostStagesAreDeclared` (`stats/internal/store/weeklycosts_internal_test.go`) to cover `visualVerify`.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestWeekly'`.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/weeklycosts_internal_test.go`
**Tests:** `TestWeeklyCostsCountsVisualVerify`, `TestWeeklyModelEffortCostsSplitsDispatchAndSession`
**Regression:** `TestWeeklyCostsCountsVisualVerify` fails, because visual-verify cost lands in overall only and no Other column exists; `TestWeeklyModelEffortCostsSplitsDispatchAndSession` fails, because no model·effort series exists.
**Baseline:** before=325 after=327 <!-- measured: grep -rhE '^func Test' stats/internal/store | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 1
**Commit:** `feat(store): weekly Visual verify cost and the model·effort series`

**Decision:** visual-verify-category
**Decision:** dispatch-served-effort
**Decision:** other-category
**Decision:** session-line

- [x] 3. API: visualVerifyUsd and the weekly-model-effort-costs view
  - [x] **Step 1: Write the failing tests** in `stats/internal/api/stats_test.go`, on the fake store the existing weekly-costs view test uses: `TestStatsWeeklyCostsCarriesVisualVerify` — the `weekly-costs` row carries `visualVerifyUsd` and `otherUsd`; `TestStatsWeeklyModelEffortCosts` — `GET /api/v1/stats/weekly-model-effort-costs?from=…&to=…` answers rows `{week: "YYYY-MM-DD", model, effort, costUsd}` from `WeeklyModelEffortCosts`, and ignores `project`/`model` query parameters. Run `cd stats && go test ./internal/api/ -run 'TestStatsWeekly'` — RED.
  - [x] **Step 2: Implement** in `stats/internal/api/stats.go`: `VisualVerifyUSD` and `OtherUSD` on `weeklyCostPointDTO` (`json:"visualVerifyUsd"`, `json:"otherUsd"`); `WeeklyModelEffortCosts` on the store interface; `viewWeeklyModelEffortCosts viewName = "weekly-model-effort-costs"` in the view list and switch, with `weeklyModelEffortPointDTO` and its converter. Add the method to every fake store the package's tests use.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/api/ -run 'TestStats'`.

**Files:** `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`
**Tests:** `TestStatsWeeklyCostsCarriesVisualVerify`, `TestStatsWeeklyModelEffortCosts`
**Regression:** `TestStatsWeeklyModelEffortCosts` fails, because the view name is unknown and answers 400.
**Baseline:** before=179 after=181 <!-- measured: grep -rhE '^func Test' stats/internal/api | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 2
**Commit:** `feat(api): weekly-model-effort-costs view, visualVerifyUsd and otherUsd`

**Decision:** model-effort-view

- [x] 4. Store: RepriceSince
  - [x] **Step 1: Write the failing test** `TestRepriceSinceCountsPricedAndSkipped` in `stats/internal/store/stageruns_test.go`: three stage runs — one before `since`, one after with priceable tokens and an unpriced served-model dispatch, one after with no `models` — `RepriceSince(ctx, since)` returns `priced=1 skipped=1`, the second run's dispatch now has `cost_usd`, the first run is untouched. Run `cd stats && go test ./internal/store/ -run 'TestRepriceSince'` — RED.
  - [x] **Step 2: Implement** `func (s *Store) RepriceSince(ctx context.Context, since time.Time) (priced, skipped int, err error)` in `stats/internal/store/pricing.go`: select the ids of stage runs with `started_at >= since`, call `Price` on each; `errors.Is` `ErrTokensUnavailable` or `ErrPricingNotFound` counts as skipped, any other error returns.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestRepriceSince|TestPrice'`.

**Files:** `stats/internal/store/pricing.go`, `stats/internal/store/stageruns_test.go`
**Tests:** `TestRepriceSinceCountsPricedAndSkipped`
**Regression:** `TestRepriceSinceCountsPricedAndSkipped` fails, because nothing re-prices an already-harvested run.
**Baseline:** before=327 after=328 <!-- measured: grep -rhE '^func Test' stats/internal/store | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 2
**Commit:** `feat(store): reprice every stage run since a date`

**Decision:** reprice-command

- [x] 5. API: POST /api/v1/pricing/reprice
  - [x] **Step 1: Write the failing test** `TestPricingRepriceAnswersCounts` in `stats/internal/api/pricing_test.go`: `POST /api/v1/pricing/reprice` with `{"since": "2026-09-01"}` calls the store with 2026-09-01T00:00:00Z and answers 200 `{"priced": N, "skipped": M}`; a malformed `since` answers 400; a store error answers 500. Run `cd stats && go test ./internal/api/ -run 'TestPricingReprice'` — RED.
  - [x] **Step 2: Implement** `stats/internal/api/pricing.go` (a `pricingHandler` on a one-method store interface, in the shape of `stats/internal/api/settings.go`) and register the route in `stats/internal/api/server.go` beside the others.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/api/ -run 'TestPricingReprice'`.

**Files:** `stats/internal/api/pricing.go`, `stats/internal/api/pricing_test.go`, `stats/internal/api/server.go`
**Allowed-collateral:** `stats/internal/api/*_test.go`, `stats/cmd/flowd/*.go`
**Tests:** `TestPricingRepriceAnswersCounts`
**Regression:** `TestPricingRepriceAnswersCounts` fails, because the route answers 404.
**Baseline:** before=181 after=182 <!-- measured: grep -rhE '^func Test' stats/internal/api | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 3, 4
**Commit:** `feat(api): reprice route`

**Decision:** reprice-command

- [x] 6. Client: Reprice
  - [x] **Step 1: Write the failing test** `TestClientRepricePostsSince` in `stats/internal/client/client_test.go`, on the httptest server pattern the file uses: `Reprice(ctx, "2026-09-01")` posts `{"since":"2026-09-01"}` to `/api/v1/pricing/reprice` and decodes `{priced, skipped}`; a non-200 returns an error carrying the body. Run `cd stats && go test ./internal/client/ -run 'TestClientReprice'` — RED.
  - [x] **Step 2: Implement** `stats/internal/client/pricing.go`: `type RepriceResult struct{ Priced, Skipped int }` and `func (c *Client) Reprice(ctx context.Context, since string) (RepriceResult, error)` via `sendJSON`.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/client/ -run 'TestClientReprice'`.

**Files:** `stats/internal/client/pricing.go`, `stats/internal/client/client_test.go`
**Tests:** `TestClientRepricePostsSince`
**Regression:** `TestClientRepricePostsSince` fails, because the client has no reprice call.
**Baseline:** before=71 after=72 <!-- measured: grep -rhE '^func Test' stats/internal/client | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 5
**Commit:** `feat(client): reprice call`

**Decision:** reprice-command

- [x] 7. CLI: flow pricing reprice
  - [x] **Step 1: Write the failing test** `TestPricingRepricePrintsCounts` in `stats/cmd/flow/pricing_test.go`, on the fake-daemon pattern `stats/cmd/flow/settings_test.go` uses: `flow pricing reprice -since 2026-09-01` exits 0 and prints `priced N, skipped M`; a missing or malformed `-since` exits 2 with usage; an unreachable store exits 1 with the client's error on stderr. Run `cd stats && go test ./cmd/flow/ -run 'TestPricingReprice'` — RED.
  - [x] **Step 2: Implement** `stats/cmd/flow/pricing.go` (`runPricing`, subcommand `reprice`, flag `-since` validated as `YYYY-MM-DD`) and the `case "pricing":` arm plus its usage line in `stats/cmd/flow/main.go`.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./cmd/flow/ -run 'TestPricingReprice|TestRun'`.

**Files:** `stats/cmd/flow/pricing.go`, `stats/cmd/flow/pricing_test.go`, `stats/cmd/flow/main.go`, `stats/README.md`
**Tests:** `TestPricingRepricePrintsCounts`
**Regression:** `TestPricingRepricePrintsCounts` fails, because `flow pricing` is an unknown command.
**Baseline:** before=348 after=349 <!-- measured: grep -rhE '^func Test' stats/cmd/flow | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 6
**Commit:** `feat(cli): flow pricing reprice`

**Decision:** reprice-command

- [x] 8. SPA: Visual verify line, model·effort chart, permanent cost labels
  - [x] **Step 1: Write the failing tests** in `stats/web/src/views/views.test.tsx`, time pinned as the existing weekly tests pin it, adding a `weekly-model-effort-costs` fixture (rows across weeks 2026-06-03, 06-10, 06-17 for `claude-opus-5-5/low`, `claude-opus-5-5/session`, `claude-sonnet-5-5/medium`) and `visualVerifyUsd` and `otherUsd` to the `weekly-costs` fixture rows:
    - `weekly costs draws a Visual verify series` — toggles `Visual verify` then `Other`, last, seven in all, and their points.
    - `weekly costs labels every dot with its cost` — Overall 06-03 (first week) label reads `$5.00` alone; Overall 06-10 reads `$110.00` and `▲2100%`; the week after a gap reads its cost alone.
    - `weekly costs draws a model·effort chart from the second view` — exactly two fetches (`weekly-costs`, `weekly-model-effort-costs`) over the same 8-week window; a second chart titled `Cost by model · effort` with toggles `opus-5-5 · low`, `opus-5-5 · session`, `sonnet-5-5 · medium`; toggling one removes its points from that chart only; its dots carry the same cost labels.
    Update `weekly costs labels each point with its change from the previous week` to the new label text. Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'weekly costs'` — RED.
  - [x] **Step 2: Implement** per the prototype: in `stats/web/src/api.ts` add `visualVerifyUsd` and `otherUsd` to `WeeklyCostPoint`, a `WeeklyModelEffortPoint` type and the `weekly-model-effort-costs` view name; in `stats/web/src/views/WeeklyCosts.tsx` add the `Visual verify` and `Other` categories, make `WeeklyCostChart` take its series (`{name, values: (number|null)[]}[]`) and an aria label so both charts reuse it, pivot the long rows into one series per model·effort (label per `model-label`, series ordered by total cost descending), extend `stackLabels` to week 0, and render each pill as `$cost` plus the delta `tspan` coloured by `deltaClass`, its rect width from the text (the prototype's 40/84 px); the second `Panel` titled `Cost by model · effort` with its own `useStatsView`. Colours: add `--series-6` and `--series-7` (whichever the palette lacks) to `stats/web/src/theme.css` in the dark default and both light blocks, from the prototype's `--c5` and `--c6`; pill styles in `stats/web/src/styles.css`. Update the file's header comment.
  - [x] **Step 3: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'weekly costs'`.

**Files:** `stats/web/src/api.ts`, `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/theme.css`, `stats/web/src/styles.css`
**Tests:** `weekly costs draws a Visual verify series`, `weekly costs labels every dot with its cost`, `weekly costs draws a model·effort chart from the second view`
**Regression:** `weekly costs draws a model·effort chart from the second view` fails, because the view fetches one view and draws one chart; `weekly costs labels every dot with its cost` fails, because pills carry only the delta.
**Baseline:** before=170 after=173 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 3
**Commit:** `feat(stats): Weekly costs model·effort chart, Visual verify and Other lines, permanent cost labels`

**Decision:** permanent-cost-label
**Decision:** model-label
**Decision:** visual-verify-category
**Decision:** other-category

- [x] 9. Seed and re-baseline the Weekly costs visual spec
  - [x] **Step 1:** in `stats/cmd/uitest-seed/seed.go`, give the seeded stage runs a `flow.visual-verify` run, a run outside every phase (`flow.verify`) and priced `dispatches` buckets (served model and effort, `cost_usd`) in two consecutive weeks, so both charts draw lines and labels against the UI-test stack.
  - [x] **Step 2:** in `stats/web/tests/visual/weekly-costs.spec.ts`, keep the existing tests (their assertions updated to seven phase toggles) and add `draws the model·effort chart with cost labels` (second chart visible, its toggles, at least one pill reading `$`, screenshot `weekly-costs-model-effort.png`). Capture per `.flow/project.md` `## visual verification` (`--update-snapshots`), then run the verify row clean; open each PNG and compare with the prototype; fix overlap or clipping before committing.

**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/weekly-costs.spec.ts`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-light-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-empty-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-toggled-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-tooltip-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-model-effort-darwin.png`
**Allowed-collateral:** `stats/web/tests/visual/*.spec.ts-snapshots/*.png`
**Tests:** `draws the model·effort chart with cost labels`
**Regression:** `draws the model·effort chart with cost labels` fails, because no second chart renders.
**Baseline:** before=52 after=53 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/tests/visual | wc -l @ a9f37e89 -->
**Build:** green
**After:** Task 8
**Commit:** `test(stats): re-baseline Weekly costs visual spec with the model·effort chart`

- [x] 10. In-run fix 1 — flow.verify: reprice fails on an unattached plan session
  - [x] **Step 1: Write the failing tests:** in `stats/internal/store/stageruns_test.go`, `TestRepriceSinceCountsPricedAndSkipped` gains an unattached `plan.session` run (`JiraKey`, no change) after `since` with priceable tokens, and expects it priced; in `stats/internal/store/aggregate_test.go`, `TestWeeklyModelEffortCostsSplitsDispatchAndSession` gains an unattached plan session with a priced model bucket in the period and expects no row from it. Run `cd stats && go test ./internal/store/ -run 'TestRepriceSince|TestWeeklyModelEffort'` — RED.
  - [x] **Step 2: Implement:** `GetStageRun` scans `COALESCE(change_id, 0)`, as `StageRun.ChangeID`'s doc and the other two scans already do; `WeeklyModelEffortCosts` counts only runs joined to a change, as `WeeklyCosts` does, so its lines add up to Overall.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestRepriceSince|TestWeekly|TestPrice|TestBeginStage'`.

**Files:** `stats/internal/store/stageruns.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/stageruns_test.go`, `stats/internal/store/aggregate_test.go`
**Tests:** `TestRepriceSinceCountsPricedAndSkipped`, `TestWeeklyModelEffortCostsSplitsDispatchAndSession`
**Regression:** `TestRepriceSinceCountsPricedAndSkipped` fails, because `GetStageRun` cannot scan a NULL `change_id` and RepriceSince aborts with a store error; `TestWeeklyModelEffortCostsSplitsDispatchAndSession` fails, because an unattached run's cost is charted though Overall excludes it.
**Build:** green
**After:** Task 4
**Commit:** `fix(store): price and chart unattached plan sessions consistently`

**Decision:** reprice-command


- [x] 11. In-run fix 2 — flow.verify: per-week model·effort sums exceed Overall on runs with a stale total
  - [x] **Step 1: Write the failing tests:** in `stats/internal/store/stageruns_test.go`, `TestPriceOneUnpriceableBucketOmitsTopLevelTotal` first prices the run while only the `claude-opus-5` bucket exists (top-level `cost_usd` written), then merges the `no-such-model` bucket and prices again, and still expects no top-level `cost_usd`; in `stats/internal/store/aggregate_test.go`, `TestWeeklyModelEffortCostsSplitsDispatchAndSession` gains an attached run in the period whose model bucket carries `cost_usd` but whose metrics carry no top-level `cost_usd`, and expects no row from it. Run `cd stats && go test ./internal/store/ -run 'TestPriceOneUnpriceable|TestWeeklyModelEffort'` — RED.
  - [x] **Step 2: Implement:** `Price` removes the top-level `cost_usd` when any model bucket could not be priced, as its unmeasured branch already does, so a total written before a later unpriceable bucket arrived never stays reading as complete; `WeeklyModelEffortCosts` counts only runs carrying a top-level `cost_usd`, the runs Overall sums, so its lines add up to Overall.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test ./internal/store/ -run 'TestPrice|TestRepriceSince|TestWeekly'`.

**Files:** `stats/internal/store/pricing.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/stageruns_test.go`, `stats/internal/store/aggregate_test.go`
**Tests:** `TestPriceOneUnpriceableBucketOmitsTopLevelTotal`, `TestWeeklyModelEffortCostsSplitsDispatchAndSession`
**Regression:** `TestPriceOneUnpriceableBucketOmitsTopLevelTotal` fails, because the first pass's top-level `cost_usd` survives the second pass that found an unpriceable bucket; `TestWeeklyModelEffortCostsSplitsDispatchAndSession` fails, because a run Overall excludes is charted.
**Build:** green
**After:** Task 10
**Commit:** `fix(store): drop a stale run total and chart only the runs Overall sums`

- [x] 12. In-run fix 3 — flow.visual-verify: header copy and pill width depart from the prototype
  - [x] **Step 1: Write the failing tests:** in `stats/web/src/views/views.test.tsx`, a new test `weekly costs matches the prototype header and pill widths` asserts the header description reads exactly `Wednesday-to-Wednesday cost, last 8 weeks.`, and that a first-week pill (cost alone) is 40 wide and a pill with a delta is 84 wide when its text fits, as the prototype's `wd=o.p==null?40:84` draws them. Run it — RED.
  - [x] **Step 2: Implement:** in `stats/web/src/views/WeeklyCosts.tsx`, the `ViewFrame` description becomes the prototype's copy, and a pill's width is the prototype's fixed 40 (cost alone) or 84 (with a delta), widened to its text only when the text is longer, so a real figure such as `$2253.34 ▲12%` never overflows its pill.
  - [x] **Step 3: Verify:** SPA lint (`npx tsc -b`), the task's **Tests:**, then re-capture `stats/web/tests/visual/weekly-costs.spec.ts` and `full-app-suite.spec.ts` with the project's `capture` command, read every changed PNG, and rebuild `stats/web/tests/visual/full-app-suite.zip`.

**Files:** `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-empty-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-light-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-model-effort-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-toggled-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-tooltip-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-weekly-costs-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`
**Tests:** `weekly costs matches the prototype header and pill widths`
**Regression:** `weekly costs matches the prototype header and pill widths` fails, because the description carries `, every project and model` and every pill is sized `12 + chars × 6` instead of the prototype's 40/84.
**Build:** green
**After:** Task 8,9,11
**Commit:** `fix(stats): match the Weekly costs header and pill widths to the prototype`

- [x] 13. In-run fix 4 — flow.verify: guard tests report a data race in runGuard's merged buffer
  - [x] **Step 1: Write the failing test:** in `stats/internal/guard/check_base_moved_test.go`, `TestRunGuardMergesAChildsTwoStreams` registers a guard whose child process writes a burst of lines to each stream, and runs it through `runGuard`. Run `cd stats && go test -race ./internal/guard/ -run TestRunGuardMergesAChildsTwoStreams` — RED under the race detector.
  - [x] **Step 2: Implement:** `runGuard`'s merged buffer, which both `io.MultiWriter`s write into, takes a mutex, since a guard that hands both writers to a child process (`workspace`'s `docker exec`) gets one copying goroutine per stream.
  - [x] **Step 3: Verify:** Go lint, then `cd stats && go test -race ./internal/guard/ -run 'TestRunGuardMergesAChildsTwoStreams|TestWorkspace'`.

**Files:** `stats/internal/guard/check_base_moved_test.go`
**Tests:** `TestRunGuardMergesAChildsTwoStreams`
**Regression:** `TestRunGuardMergesAChildsTwoStreams` fails under `-race`, because two exec copy goroutines write the shared merged buffer unlocked.
**Build:** green
**After:** none
**Commit:** `fix(guard): lock runGuard's merged test buffer against a child's two streams`
