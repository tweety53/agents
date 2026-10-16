# kan-989-stats-ui-dashboard-review

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

## Global constraints

- `design.md` decision `retain-stored-data` binds every task: no migration drops a table or column, no task deletes stored rows. Only UI, endpoints, store queries and their tests are removed.
- `internal/web/dist` is gitignored and must exist for `go test` to compile: run `make -C stats web-build` before any Go test step.
- Removing a `StatsStore` method also removes it from the three fakes: `stats/internal/web/embed_test.go` `fakeStore`, `stats/internal/client/client_test.go` `stubStageStore`, `stats/internal/api/stats_test.go` `statsFake`.
- Removing a view name also removes it from `stats/internal/api/stats.go` (`viewName` consts, `knownViews`, `rowsFor`), `stats/web/src/api.ts` (`ViewName`, `VIEW_NAMES`), `stats/web/src/App.tsx` (`VIEW_LABELS`, `VIEW_COMPONENTS`, imports, `MODEL_DISABLED_REASONS`) and every view table in `stats/internal/api/stats_test.go`.
- Visual specs and their PNG baselines are updated once, in the last task, against the disposable UI-test stack on `127.0.0.1:4174` — never `flowd` on `127.0.0.1:4173`.
- Go: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`; SPA: `cd stats/web && npx tsc -b`.

- [ ] 1. Reviewers: fast query, Efficiency column, two columns dropped
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With `make -C stats web-build` done, time the reviewers query: add a temporary `testing.B`-free timing via `go test ./internal/store -run TestReviewers -v` and read the SQL in `stats/internal/store/aggregate.go` `Reviewers` (the `roster_slots` CTE and the correlated `description` subquery using `jsonb_path_exists` over `decisions`). Confirm against the dev endpoint: `curl -s -o /dev/null -w '%{time_total}\n' "http://127.0.0.1:4173/api/v1/stats/reviewers?from=2000-01-01T00:00:00Z&to=2030-01-01T00:00:00Z"` (read-only GET; ~0.3 s expected).
  - [ ] **Step 2: Rewrite the SQL.** Replace the per-slot correlated `description` subquery with one CTE `slot_desc` — `SELECT DISTINCT ON (r->>'slot') r->>'slot' AS slot, r->>'description' AS description FROM decisions d, jsonb_array_elements(d.decision->'panel'->'roster') r ORDER BY r->>'slot', d.recorded_at DESC` — LEFT JOINed on slot; derive `roster_slots` from the same CTE so `decisions` is read once. Keep every output value identical (existing `Reviewers` tests in `stats/internal/store/aggregate_test.go` stay green unchanged except for the two dropped fields).
  - [ ] **Step 3: Index.** Add `stats/internal/store/migrations/0039_findings_dispatch_id.sql`: `CREATE INDEX IF NOT EXISTS findings_dispatch_id ON findings (dispatch_id);`.
  - [ ] **Step 4: Drop `FindingsPerDispatch` and `WithdrawnShare`** from the store row, the DTO in `stats/internal/api/stats.go`, `ReviewerRow` in `stats/web/src/api.ts`, the two columns in `stats/web/src/views/Reviewers.tsx`, and their assertions in `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats_test.go` (`TestReviewersWireCarriesNoDeferredShare` keeps asserting the remaining wire fields).
  - [ ] **Step 5: Efficiency column.** In `Reviewers.tsx` add a sortable `Efficiency` column after `Minor`: `(10*critical + 4*important + minor) / dispatches`, rendered with 2 decimals, `—` when `dispatches` is 0 (decision `reviewers-efficiency`). Add vitest `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch` in `stats/web/src/views/views.test.tsx` asserting `1.53` for critical 4 / important 36 / minor 70 / dispatches 166 and `—` for 0 dispatches.
  - [ ] **Step 6: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store ./internal/api -run 'Reviewers' -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'reviewers'`; rebuild/restart is not this task's — the live timing is `design.md`'s Live check. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/migrations/0039_findings_dispatch_id.sql`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/Reviewers.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch`
**Regression:** `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch` — reverted, the Efficiency column is absent and the `1.53` cell is not found.
**Baseline:** before=162 after=163 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** perf(stats): load reviewers in one decisions pass and score reviewer efficiency
**After:** none

**Decision:** reviewers-efficiency
**Decision:** retain-stored-data

- [ ] 2. Remove the Cache efficiency dashboard
  - [ ] **Step 1: Delete** `stats/web/src/views/CacheEfficiency.tsx`; remove `cache-efficiency` from `stats/web/src/api.ts` (`ViewName`, `VIEW_NAMES`, `CacheEfficiencyRow`), `stats/web/src/App.tsx`, and its fixture and test (`cache efficiency: a real zero ratio …`) in `stats/web/src/views/views.test.tsx`.
  - [ ] **Step 2: Server.** Remove `viewCacheEfficiency`, its `knownViews` entry, `rowsFor` case and DTO from `stats/internal/api/stats.go`; `CacheEfficiency` from the `StatsStore` interface and the three fakes; `CacheEfficiencyRow`/`CacheEfficiency` from `stats/internal/store/aggregate.go`; `TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero` and the cache-efficiency cases in `stats/internal/store/aggregate_test.go` (`TestAggregateOtherViewsRunWithoutError`), `stats/internal/store/harvestshape_test.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/changes_test.go`. No migration (decision `retain-stored-data`).
  - [ ] **Step 3: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'TestAggregateOtherViewsRunWithoutError' -count=1 && go test ./internal/api -run 'TestEveryViewAcceptsAPeriod' -count=1`; `cd stats/web && npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/CacheEfficiency.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`, `stats/web/src/views/views.test.tsx`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/changes_test.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/harvestshape_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`
**Tests:** none — repairs existing tests: `TestAggregateOtherViewsRunWithoutError`, `TestEveryViewAcceptsAPeriod`; removes `TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero`.
**Regression:** none — the task declares no tests
**Baseline:** before=163 after=162 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): remove the cache efficiency dashboard
**After:** Task 1

**Decision:** remove-cache-efficiency

- [ ] 3. Live state board rows carry whole-life run totals
  - [ ] **Step 1: Store.** In `stats/internal/store/aggregate.go` add `RunCount int`, `CostUsd float64`, `WallClockMs int64`, `Errors int` to `LiveStateRow`, filled in `LiveStateBoard` from a LEFT JOIN on a `stage_runs` aggregate: inner `GROUP BY change_id, COALESCE(session_token, 'stage-' || id)` computing `MIN(started_at)`, `MAX(ended_at)`, `SUM((metrics->>'cost_usd')::numeric)`, and the sum of `metrics->'signals'->{'main','sidechain'}->>{'tool_errors','api_errors'}` (COALESCE to 0); outer per `change_id`: `COUNT(*)`, `SUM(cost)`, `SUM(EXTRACT(EPOCH FROM max-min)*1000)` over runs whose `MAX(ended_at)` is not null, `SUM(errors)` — unbounded by the period (decision `board-run-totals`). Definitions match `stats/internal/store/runs.go` `groupRuns`/`addCost`/`addSignals`; read them before writing the SQL.
  - [ ] **Step 2: Test.** Add `TestLiveStateBoardCarriesWholeLifeRunTotals` in `stats/internal/store/aggregate_test.go`: one change, two stage rows sharing a token plus one token-less row, one row started before the period — asserts RunCount 2, summed cost, wall clock, errors.
  - [ ] **Step 3: API.** Add `runCount`, `costUsd`, `wallClockMs`, `errors` to `stateBoardRowDTO` in `stats/internal/api/stats.go`; keep `updatedBy` and `nextCommand` (decision `board-filter-in-spa`); extend `TestLiveStateBoardMatchesStatusOutput` in `stats/internal/api/stats_test.go` to assert the four fields. `stats/internal/client/client.go` is untouched (additive fields).
  - [ ] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'LiveStateBoard' -count=1 && go test ./internal/api -run 'LiveStateBoard' -count=1 && go test ./internal/client ./cmd/flow -count=1`. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`
**Tests:** `TestLiveStateBoardCarriesWholeLifeRunTotals`
**Regression:** `TestLiveStateBoardCarriesWholeLifeRunTotals` — reverted, the row has no run totals and the assertions fail.
**Baseline:** before=286 after=287 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): give each live state board row its whole-life run totals
**After:** Task 1, 2

**Decision:** board-run-totals

- [ ] 4. Live state board UI: hide pipeline-fix, swap columns
  - [ ] **Step 1: UI.** In `stats/web/src/views/StateBoard.tsx` drop the `By` and `Next command` columns and the "next command" wording in the view description; filter out rows whose `name` starts with `pipeline-fix-` before the table and the stat panels; add `Runs` (`formatInt`), `Cost` (`formatUsd`), `Wall clock` (`formatMs`), `Errors` (`formatInt`) sortable columns after `Updated`. Add the four fields to `StateBoardRow` in `stats/web/src/api.ts` (keep `updatedBy`/`nextCommand` on the type — the wire still carries them).
  - [ ] **Step 2: Tests.** In `stats/web/src/views/views.test.tsx` rewrite `state board shows the change's state, updater and next command` to `state board shows the change's state and run totals` (no By/Next command headers; Runs/Cost/Wall clock/Errors cells), extend its fixture with the four fields, and add `state board hides pipeline-fix changes` (a `pipeline-fix-x-14` row in the fixture is absent from the table).
  - [ ] **Step 3: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'state board'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/StateBoard.tsx`, `stats/web/src/api.ts`, `stats/web/src/views/views.test.tsx`
**Tests:** `state board hides pipeline-fix changes`
**Regression:** `state board hides pipeline-fix changes` — reverted, the `pipeline-fix-x-14` row renders.
**Baseline:** before=162 after=163 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** feat(stats): hide pipeline-fix changes and show run totals on the live board
**After:** Task 1, 2, 3

**Decision:** board-filter-in-spa
**Decision:** board-run-totals

- [ ] 5. Remove the Runs dashboard
  - [ ] **Step 1: Check the seed.** Read `stats/web/tests/visual/run-detail.spec.ts` and `stats/cmd/uitest-seed/seed.go` `seedRunsFixture`; keep any seed data run-detail reads, remove only what the runs view alone used.
  - [ ] **Step 2: SPA.** Delete `stats/web/src/views/Runs.tsx` and `stats/web/src/views/Runs.test.tsx`; remove `runs` from `stats/web/src/api.ts` (`ViewName`, `VIEW_NAMES`, the `RunTotals`/`RunDispatchRow`/`RunStageSpan`/`RunRow`/`ChangeRuns` types where nothing else imports them) and `stats/web/src/App.tsx`. The `#/run/…` RunDetail route stays.
  - [ ] **Step 3: Server.** Remove `viewRuns`, its `knownViews` entry, `rowsFor` case and run DTOs from `stats/internal/api/stats.go`; `ListRuns` from `StatsStore` and the three fakes; `TestRunsViewIncludesStageRows` from `stats/internal/api/stats_test.go`; delete `stats/internal/store/runs.go` and `stats/internal/store/runs_test.go` after moving any helper another file still calls (grep `humanGateStages`, `addSignals`, `addCost` first).
  - [ ] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/api -run 'TestEveryViewAcceptsAPeriod' -count=1`; `cd stats/web && npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/Runs.tsx`, `stats/web/src/views/Runs.test.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/store/runs.go`, `stats/internal/store/runs_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/cmd/uitest-seed/seed.go`
**Allowed-collateral:** `stats/internal/store/*.go`
**Tests:** none — repairs existing tests: `TestEveryViewAcceptsAPeriod`; removes `stats/internal/store/runs_test.go`, `stats/web/src/views/Runs.test.tsx` and `TestRunsViewIncludesStageRows`.
**Regression:** none — the task declares no tests
**Baseline:** before=163 after=158 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): remove the runs dashboard now the live board carries its totals
**After:** Task 1, 2, 3, 4

**Decision:** board-run-totals
**Decision:** retain-stored-data

- [ ] 6. Stage leaderboard: total cost, top 10
  - [ ] **Step 1: Store.** In `stats/internal/store/aggregate.go` `StageLeaderboard` add `SUM(cost)` as `TotalCostUsd` to the row and replace `ORDER BY command, stage` with `ORDER BY SUM(cost) DESC, command, stage LIMIT 10`. Update `TestStageLeaderboardComputesMeanMedianP90` for the new order and add `TestStageLeaderboardKeepsTopTenByTotalCost` (eleven stages; the cheapest by total is absent, rows ordered by total desc).
  - [ ] **Step 2: API + SPA.** Add `totalCostUsd` to `stageLeaderboardRowDTO` (`stats/internal/api/stats.go`, and `TestEveryViewCarriesItsRealNumbersThrough` in `stats/internal/api/stats_test.go`), to `StageLeaderboardRow` in `stats/web/src/api.ts`, and a sortable `Total cost` column (`formatUsd`) in `stats/web/src/views/StageLeaderboard.tsx`, default sort Total cost descending; update the leaderboard test in `stats/web/src/views/views.test.tsx` to assert the header.
  - [ ] **Step 3: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'StageLeaderboard|ModelFilter' -count=1 && go test ./internal/api -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'stage leaderboard'`. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/StageLeaderboard.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `TestStageLeaderboardKeepsTopTenByTotalCost`
**Regression:** `TestStageLeaderboardKeepsTopTenByTotalCost` — reverted, eleven rows return in command/stage order.
**Baseline:** before=277 after=278 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): rank the stage leaderboard by total cost, top ten
**After:** Task 1, 2, 3, 4, 5

**Decision:** leaderboard-top10-total

- [ ] 7. Decisions: By implementer model replaces Every decision
  - [ ] **Step 1: UI.** In `stats/web/src/views/Decisions.tsx` remove the `Runs` stat panel and the `Every decision` panel and its `columns`; add a `By implementer model` panel below `By class and execution`: client-side group of `data.rows` on `implementerModel` with Runs, Mean cost (`formatUsd`), Mean wall-clock, Mean findings (critical+important+minor), Mean fix rounds, sorted by Runs desc (decision `decisions-by-implementer-model`).
  - [ ] **Step 2: Tests.** In `stats/web/src/views/views.test.tsx` update `decisions shows a run's class, execution and cost, and a class × execution summary` (no Every decision table, no Runs tile) and add `decisions groups runs by implementer model` (two rows on `opus`, one on `skipped — inline` → two groups, opus first with Runs 2 and their mean cost).
  - [ ] **Step 3: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'decisions'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/Decisions.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `decisions groups runs by implementer model`
**Regression:** `decisions groups runs by implementer model` — reverted, no `By implementer model` panel renders.
**Baseline:** before=158 after=159 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** feat(stats): summarise decisions by implementer model
**After:** Task 1, 2, 3, 4, 5, 6

**Decision:** decisions-by-implementer-model

- [ ] 8. Flow health becomes eight totals tiles
  - [ ] **Step 1: Store.** Add `FlowSummary(ctx, from, to, project)` in `stats/internal/store/health.go` returning one `FlowSummaryRow{ChangesStarted, ChangesFinished int; TotalCostUsd, MeanCostPerFinished float64; MedianLeadTimeMs *int64; MeanPanelRounds *float64; Critical, Important, Minor, GuardFires int}` per decision `flow-health-tiles` (started = first `stage_runs.started_at` per change in period; finished = `state='FINISHED' AND NOT withdrawn AND updated_at` in period; lead time = finish `updated_at` − first stage-run start; panel rounds and C/I/Mi from the existing `PanelRounds`; guard fires from `guard_runs` with outcome `fired` in period). Add `TestFlowSummaryTotals` in `stats/internal/store/health_test.go`. Delete `StageRedo`, `GuardActivity` and their tests (`TestStageRedoCountsReentries`, `TestGuardActivityMergesRunsAndVerdicts`) once nothing calls them; keep `RecordGuardRun` and every table (decision `retain-stored-data`).
  - [ ] **Step 2: API.** In `stats/internal/api/health.go` and `stats/internal/api/stats.go` replace the `guards`, `stage-redo`, `panel-rounds` health views with one `flow-summary`; update `stats/internal/api/health_test.go` and the view tables in `stats/internal/api/stats_test.go`; update the three fakes.
  - [ ] **Step 3: SPA.** Rewrite `stats/web/src/views/FlowHealth.tsx` to fetch `flow-summary` and render eight `StatPanel` tiles: Changes started · Changes finished · Total cost · Mean cost per finished change · Median lead time · Mean panel rounds · Findings (`C/I/Mi`) · Guard fires; nothing else. Update `HealthViewSlug` and types in `stats/web/src/api.ts`. In `stats/web/src/views/views.test.tsx` replace the two flow-health tests with `flow health shows eight totals tiles`.
  - [ ] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'FlowSummary|PanelRounds|GuardRun' -count=1 && go test ./internal/api ./internal/client ./internal/web -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'flow health'`. Commit.

**Build:** green
**Files:** `stats/internal/store/health.go`, `stats/internal/store/health_test.go`, `stats/internal/api/health.go`, `stats/internal/api/health_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/FlowHealth.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `TestFlowSummaryTotals`
**Regression:** `TestFlowSummaryTotals` — reverted, `FlowSummary` does not exist and the package fails to compile.
**Baseline:** before=278 after=277 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): replace flow health tables with period totals
**After:** Task 1, 2, 3, 4, 5, 6, 7

**Decision:** flow-health-tiles
**Decision:** retain-stored-data

- [ ] 9. Self-review fixes: drop the Per change block
  - [ ] **Step 1:** In `stats/web/src/views/SelfReview.tsx` delete the `Per change` panel and `perChange()`; update `renders self-review fixes` in `stats/web/src/views/views.test.tsx` to assert no `Per change` heading.
  - [ ] **Step 2: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'self-review'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/SelfReview.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** none — repairs existing tests: `renders self-review fixes`.
**Regression:** none — the task declares no tests
**Baseline:** before=158 after=158 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): drop the per-change block from self-review fixes
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8

**Decision:** self-review-drop-per-change

- [ ] 10. Visual specs and baselines follow the new dashboards
  - [ ] **Step 1:** Start the disposable UI-test stack on `127.0.0.1:4174` per `.flow/project.md` `## run` (never `flowd` on 4173).
  - [ ] **Step 2:** Delete `stats/web/tests/visual/runs.spec.ts` and the cache-efficiency describe in `stats/web/tests/visual/views.spec.ts`; drop `runs` and `cache-efficiency` from `stats/web/tests/visual/full-app-suite.spec.ts` and `RUNS_QUERY` from `stats/web/tests/visual/support.ts` if unused; delete their PNGs; adjust `controls.spec.ts`, `combinations.spec.ts`, `reviewers-decisions.spec.ts`, `self-review.spec.ts`, `baseline.spec.ts` assertions that name removed columns or blocks.
  - [ ] **Step 3:** Re-capture changed baselines with the project's `capture` command (`--update-snapshots`) for the affected specs, then run the `verify` command for all visual specs: exit 0.
  - [ ] **Step 4: Verify.** `cd stats/web && npx tsc -b`; `scripts/resolve-visual-screenshots.sh . baseline.spec.ts`; `scripts/check-visual-verification.sh .`. Commit.

**Build:** green
**Files:** `stats/web/tests/visual/runs.spec.ts`, `stats/web/tests/visual/views.spec.ts`, `stats/web/tests/visual/full-app-suite.spec.ts`, `stats/web/tests/visual/support.ts`
**Allowed-collateral:** `stats/web/tests/visual/**`
**Tests:** none — repairs existing tests: `views.spec.ts`, `full-app-suite.spec.ts`, `controls.spec.ts`, `combinations.spec.ts`, `reviewers-decisions.spec.ts`, `self-review.spec.ts`, `baseline.spec.ts`.
**Regression:** none — the task declares no tests
**Baseline:** before=158 after=158 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** test(stats): update visual specs and baselines for the reworked dashboards
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 9
