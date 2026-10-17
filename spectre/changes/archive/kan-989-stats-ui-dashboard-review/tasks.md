# kan-989-stats-ui-dashboard-review

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Tasks appended:** 3
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

- [x] 1. Reviewers: fast query, Efficiency column, two columns dropped
  - [x] **Step 1: Reproduce before touching anything; do not guess.** With `make -C stats web-build` done, time the reviewers query: add a temporary `testing.B`-free timing via `go test ./internal/store -run TestReviewers -v` and read the SQL in `stats/internal/store/aggregate.go` `Reviewers` (the `roster_slots` CTE and the correlated `description` subquery using `jsonb_path_exists` over `decisions`). Confirm against the dev endpoint: `curl -s -o /dev/null -w '%{time_total}\n' "http://127.0.0.1:4173/api/v1/stats/reviewers?from=2000-01-01T00:00:00Z&to=2030-01-01T00:00:00Z"` (read-only GET; ~0.3 s expected).
  - [x] **Step 2: Rewrite the SQL.** Replace the per-slot correlated `description` subquery with one CTE `slot_desc` — `SELECT DISTINCT ON (r->>'slot') r->>'slot' AS slot, r->>'description' AS description FROM decisions d, jsonb_array_elements(d.decision->'panel'->'roster') r ORDER BY r->>'slot', d.recorded_at DESC` — LEFT JOINed on slot; derive `roster_slots` from the same CTE so `decisions` is read once. Keep every output value identical (existing `Reviewers` tests in `stats/internal/store/aggregate_test.go` stay green unchanged except for the two dropped fields).
  - [x] **Step 3: Index.** Add `stats/internal/store/migrations/0039_findings_dispatch_id.sql`: `CREATE INDEX IF NOT EXISTS findings_dispatch_id ON findings (dispatch_id);`.
  - [x] **Step 4: Drop `FindingsPerDispatch` and `WithdrawnShare`** from the store row, the DTO in `stats/internal/api/stats.go`, `ReviewerRow` in `stats/web/src/api.ts`, the two columns in `stats/web/src/views/Reviewers.tsx`, and their assertions in `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats_test.go` (`TestReviewersWireCarriesNoDeferredShare` keeps asserting the remaining wire fields).
  - [x] **Step 5: Efficiency column.** In `Reviewers.tsx` add a sortable `Efficiency` column after `Minor`: `(10*critical + 4*important + minor) / dispatches`, rendered with 2 decimals, `—` when `dispatches` is 0 (decision `reviewers-efficiency`). Add vitest `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch` in `stats/web/src/views/views.test.tsx` asserting `1.53` for critical 4 / important 36 / minor 70 / dispatches 166 and `—` for 0 dispatches.
  - [x] **Step 6: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store ./internal/api -run 'Reviewers' -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'reviewers'`; rebuild/restart is not this task's — the live timing is `design.md`'s Live check. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/migrations/0039_findings_dispatch_id.sql`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/Reviewers.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch`
**Regression:** `reviewers efficiency weighs critical 10, important 4, minor 1 per dispatch` — reverted, the Efficiency column is absent and the `1.53` cell is not found.
**Baseline:** before=162 after=163 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** perf(stats): load reviewers in one decisions pass and score reviewer efficiency
**After:** none

**Decision:** reviewers-efficiency
**Decision:** retain-stored-data

- [x] 2. Remove the Cache efficiency dashboard
  - [x] **Step 1: Delete** `stats/web/src/views/CacheEfficiency.tsx`; remove `cache-efficiency` from `stats/web/src/api.ts` (`ViewName`, `VIEW_NAMES`, `CacheEfficiencyRow`), `stats/web/src/App.tsx`, and its fixture and test (`cache efficiency: a real zero ratio …`) in `stats/web/src/views/views.test.tsx`.
  - [x] **Step 2: Server.** Remove `viewCacheEfficiency`, its `knownViews` entry, `rowsFor` case and DTO from `stats/internal/api/stats.go`; `CacheEfficiency` from the `StatsStore` interface and the three fakes; `CacheEfficiencyRow`/`CacheEfficiency` from `stats/internal/store/aggregate.go`; `TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero` and the cache-efficiency cases in `stats/internal/store/aggregate_test.go` (`TestAggregateOtherViewsRunWithoutError`), `stats/internal/store/harvestshape_test.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/changes_test.go`. No migration (decision `retain-stored-data`).
  - [x] **Step 3: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'TestAggregateOtherViewsRunWithoutError' -count=1 && go test ./internal/api -run 'TestEveryViewAcceptsAPeriod' -count=1`; `cd stats/web && npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/CacheEfficiency.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`, `stats/web/src/views/views.test.tsx`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/api/changes_test.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/store/harvestshape_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/web/src/views/RunDetail.test.tsx`, `README.md`
**Tests:** none — repairs existing tests: `TestAggregateOtherViewsRunWithoutError`, `TestEveryViewAcceptsAPeriod`; removes `TestCacheEfficiencyRatioStaysNilWhenCacheCreationIsExactlyZero`.
**Regression:** none — the task declares no tests
**Baseline:** before=163 after=162 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): remove the cache efficiency dashboard
**After:** Task 1

**Decision:** remove-cache-efficiency

Correction (2026-10-17): the plan's `**Files:**` omitted two paths the removal reaches — `stats/web/src/views/RunDetail.test.tsx` pins `VIEW_NAMES`' length (9 → 8), and the root `README.md` lists the dashboards; both were updated in this task's commit and the field widened.

- [x] 3. Live state board rows carry whole-life run totals
  - [x] **Step 1: Store.** In `stats/internal/store/aggregate.go` add `RunCount int`, `CostUsd float64`, `WallClockMs int64`, `Errors int` to `LiveStateRow`, filled in `LiveStateBoard` from a LEFT JOIN on a `stage_runs` aggregate: inner `GROUP BY change_id, COALESCE(session_token, 'stage-' || id)` computing `MIN(started_at)`, `MAX(ended_at)`, `SUM((metrics->>'cost_usd')::numeric)`, and the sum of `metrics->'signals'->{'main','sidechain'}->>{'tool_errors','api_errors'}` (COALESCE to 0); outer per `change_id`: `COUNT(*)`, `SUM(cost)`, `SUM(EXTRACT(EPOCH FROM max-min)*1000)` over runs whose `MAX(ended_at)` is not null, `SUM(errors)` — unbounded by the period (decision `board-run-totals`). Definitions match `stats/internal/store/runs.go` `groupRuns`/`addCost`/`addSignals`; read them before writing the SQL.
  - [x] **Step 2: Test.** Add `TestLiveStateBoardCarriesWholeLifeRunTotals` in `stats/internal/store/aggregate_test.go`: one change, two stage rows sharing a token plus one token-less row, one row started before the period — asserts RunCount 2, summed cost, wall clock, errors.
  - [x] **Step 3: API.** Add `runCount`, `costUsd`, `wallClockMs`, `errors` to `stateBoardRowDTO` in `stats/internal/api/stats.go`; keep `updatedBy` and `nextCommand` (decision `board-filter-in-spa`); extend `TestLiveStateBoardMatchesStatusOutput` in `stats/internal/api/stats_test.go` to assert the four fields. `stats/internal/client/client.go` is untouched (additive fields).
  - [x] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'LiveStateBoard' -count=1 && go test ./internal/api -run 'LiveStateBoard' -count=1 && go test ./internal/client ./cmd/flow -count=1`. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`
**Tests:** `TestLiveStateBoardCarriesWholeLifeRunTotals`
**Regression:** `TestLiveStateBoardCarriesWholeLifeRunTotals` — reverted, the row has no run totals and the assertions fail.
**Baseline:** before=286 after=287 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): give each live state board row its whole-life run totals
**After:** Task 1, 2

**Decision:** board-run-totals

- [x] 4. Live state board UI: hide pipeline-fix, swap columns
  - [x] **Step 1: UI.** In `stats/web/src/views/StateBoard.tsx` drop the `By` and `Next command` columns and the "next command" wording in the view description; filter out rows whose `name` starts with `pipeline-fix-` before the table and the stat panels; add `Runs` (`formatInt`), `Cost` (`formatUsd`), `Wall clock` (`formatMs`), `Errors` (`formatInt`) sortable columns after `Updated`. Add the four fields to `StateBoardRow` in `stats/web/src/api.ts` (keep `updatedBy`/`nextCommand` on the type — the wire still carries them).
  - [x] **Step 2: Tests.** In `stats/web/src/views/views.test.tsx` rewrite `state board shows the change's state, updater and next command` to `state board shows the change's state and run totals` (no By/Next command headers; Runs/Cost/Wall clock/Errors cells), extend its fixture with the four fields, and add `state board hides pipeline-fix changes` (a `pipeline-fix-x-14` row in the fixture is absent from the table).
  - [x] **Step 3: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'state board'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/StateBoard.tsx`, `stats/web/src/api.ts`, `stats/web/src/views/views.test.tsx`
**Tests:** `state board hides pipeline-fix changes`
**Regression:** `state board hides pipeline-fix changes` — reverted, the `pipeline-fix-x-14` row renders.
**Baseline:** before=162 after=163 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** feat(stats): hide pipeline-fix changes and show run totals on the live board
**After:** Task 1, 2, 3

**Decision:** board-filter-in-spa
**Decision:** board-run-totals

- [x] 5. Remove the Runs dashboard
  - [x] **Step 1: Check the seed.** Read `stats/web/tests/visual/run-detail.spec.ts` and `stats/cmd/uitest-seed/seed.go` `seedRunsFixture`; keep any seed data run-detail reads, remove only what the runs view alone used.
  - [x] **Step 2: SPA.** Delete `stats/web/src/views/Runs.tsx` and `stats/web/src/views/Runs.test.tsx`; remove `runs` from `stats/web/src/api.ts` (`ViewName`, `VIEW_NAMES`, the `RunTotals`/`RunDispatchRow`/`RunStageSpan`/`RunRow`/`ChangeRuns` types where nothing else imports them) and `stats/web/src/App.tsx`. The `#/run/…` RunDetail route stays.
  - [x] **Step 3: Server.** Remove `viewRuns`, its `knownViews` entry, `rowsFor` case and run DTOs from `stats/internal/api/stats.go`; `ListRuns` from `StatsStore` and the three fakes; `TestRunsViewIncludesStageRows` from `stats/internal/api/stats_test.go`; delete `stats/internal/store/runs.go` and `stats/internal/store/runs_test.go` after moving any helper another file still calls (grep `humanGateStages`, `addSignals`, `addCost` first).
  - [x] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/api -run 'TestEveryViewAcceptsAPeriod' -count=1`; `cd stats/web && npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/Runs.tsx`, `stats/web/src/views/Runs.test.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/store/runs.go`, `stats/internal/store/runs_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/cmd/uitest-seed/seed.go`, `stats/cmd/uitest-seed/seed_test.go`, `stats/web/src/views/RunDetail.test.tsx`, `stats/web/src/views/views.test.tsx`, `README.md`
**Allowed-collateral:** `stats/internal/store/*.go`
**Tests:** none — repairs existing tests: `TestEveryViewAcceptsAPeriod`; removes `stats/internal/store/runs_test.go`, `stats/web/src/views/Runs.test.tsx` and `TestRunsViewIncludesStageRows`.
**Regression:** none — the task declares no tests
**Baseline:** before=163 after=158 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): remove the runs dashboard now the live board carries its totals
**After:** Task 1, 2, 3, 4

**Decision:** board-run-totals
**Decision:** retain-stored-data

Correction (2026-10-17): `seedRunsFixture` fed only the runs view, so it was removed with `seed_test.go`'s count for it; `RunDetail.test.tsx` pins `VIEW_NAMES`' length (8 → 7), `views.test.tsx` carried the `runs` fixture, and the root `README.md` listed the view — all four were widened into `**Files:**`. `store/export_test.go` existed only to expose `fanOutMax` to `runs_test.go` and went with it.

- [x] 6. Stage leaderboard: total cost, top 10
  - [x] **Step 1: Store.** In `stats/internal/store/aggregate.go` `StageLeaderboard` add `SUM(cost)` as `TotalCostUsd` to the row and replace `ORDER BY command, stage` with `ORDER BY SUM(cost) DESC, command, stage LIMIT 10`. Update `TestStageLeaderboardComputesMeanMedianP90` for the new order and add `TestStageLeaderboardKeepsTopTenByTotalCost` (eleven stages; the cheapest by total is absent, rows ordered by total desc).
  - [x] **Step 2: API + SPA.** Add `totalCostUsd` to `stageLeaderboardRowDTO` (`stats/internal/api/stats.go`, and `TestEveryViewCarriesItsRealNumbersThrough` in `stats/internal/api/stats_test.go`), to `StageLeaderboardRow` in `stats/web/src/api.ts`, and a sortable `Total cost` column (`formatUsd`) in `stats/web/src/views/StageLeaderboard.tsx`, default sort Total cost descending; update the leaderboard test in `stats/web/src/views/views.test.tsx` to assert the header.
  - [x] **Step 3: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'StageLeaderboard|ModelFilter' -count=1 && go test ./internal/api -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'stage leaderboard'`. Commit.

**Build:** green
**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/StageLeaderboard.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `TestStageLeaderboardKeepsTopTenByTotalCost`
**Regression:** `TestStageLeaderboardKeepsTopTenByTotalCost` — reverted, eleven rows return in command/stage order.
**Baseline:** before=277 after=278 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): rank the stage leaderboard by total cost, top ten
**After:** Task 1, 2, 3, 4, 5

**Decision:** leaderboard-top10-total

Correction (2026-10-17): step 2's "default sort Total cost descending" shipped as the server's order — `StageLeaderboard` returns rows ranked by total cost and `DataTable` renders them in that order until a reader sorts another column; no client-side initial sort or ▼ indicator was added.

- [x] 7. Decisions: By implementer model replaces Every decision
  - [x] **Step 1: UI.** In `stats/web/src/views/Decisions.tsx` remove the `Runs` stat panel and the `Every decision` panel and its `columns`; add a `By implementer model` panel below `By class and execution`: client-side group of `data.rows` on `implementerModel` with Runs, Mean cost (`formatUsd`), Mean wall-clock, Mean findings (critical+important+minor), Mean fix rounds, sorted by Runs desc (decision `decisions-by-implementer-model`).
  - [x] **Step 2: Tests.** In `stats/web/src/views/views.test.tsx` update `decisions shows a run's class, execution and cost, and a class × execution summary` (no Every decision table, no Runs tile) and add `decisions groups runs by implementer model` (two rows on `opus`, one on `skipped — inline` → two groups, opus first with Runs 2 and their mean cost).
  - [x] **Step 3: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'decisions'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/Decisions.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `decisions groups runs by implementer model`
**Regression:** `decisions groups runs by implementer model` — reverted, no `By implementer model` panel renders.
**Baseline:** before=158 after=159 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** feat(stats): summarise decisions by implementer model
**After:** Task 1, 2, 3, 4, 5, 6

**Decision:** decisions-by-implementer-model

- [x] 8. Flow health becomes eight totals tiles
  - [x] **Step 1: Store.** Add `FlowSummary(ctx, from, to, project)` in `stats/internal/store/health.go` returning one `FlowSummaryRow{ChangesStarted, ChangesFinished int; TotalCostUsd, MeanCostPerFinished float64; MedianLeadTimeMs *int64; MeanPanelRounds *float64; Critical, Important, Minor, GuardFires int}` per decision `flow-health-tiles` (started = first `stage_runs.started_at` per change in period; finished = `state='FINISHED' AND NOT withdrawn AND updated_at` in period; lead time = finish `updated_at` − first stage-run start; panel rounds and C/I/Mi from the existing `PanelRounds`; guard fires from `guard_runs` with outcome `fired` in period). Add `TestFlowSummaryTotals` in `stats/internal/store/health_test.go`. Delete `StageRedo`, `GuardActivity` and their tests (`TestStageRedoCountsReentries`, `TestGuardActivityMergesRunsAndVerdicts`) once nothing calls them; keep `RecordGuardRun` and every table (decision `retain-stored-data`).
  - [x] **Step 2: API.** In `stats/internal/api/health.go` and `stats/internal/api/stats.go` replace the `guards`, `stage-redo`, `panel-rounds` health views with one `flow-summary`; update `stats/internal/api/health_test.go` and the view tables in `stats/internal/api/stats_test.go`; update the three fakes.
  - [x] **Step 3: SPA.** Rewrite `stats/web/src/views/FlowHealth.tsx` to fetch `flow-summary` and render eight `StatPanel` tiles: Changes started · Changes finished · Total cost · Mean cost per finished change · Median lead time · Mean panel rounds · Findings (`C/I/Mi`) · Guard fires; nothing else. Update `HealthViewSlug` and types in `stats/web/src/api.ts`. In `stats/web/src/views/views.test.tsx` replace the two flow-health tests with `flow health shows eight totals tiles`.
  - [x] **Step 4: Verify.** `make -C stats web-build && cd stats && gofmt -l . && go vet ./... && go test ./internal/store -run 'FlowSummary|PanelRounds|GuardRun' -count=1 && go test ./internal/api ./internal/client ./internal/web -count=1`; `cd stats/web && npx tsc -b && npx vitest run -t 'flow health'`. Commit.

**Build:** green
**Files:** `stats/internal/store/health.go`, `stats/internal/store/health_test.go`, `stats/internal/api/health.go`, `stats/internal/api/health_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`, `stats/internal/web/embed_test.go`, `stats/internal/client/client_test.go`, `stats/web/src/api.ts`, `stats/web/src/views/FlowHealth.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/App.tsx`
**Tests:** `TestFlowSummaryTotals`
**Regression:** `TestFlowSummaryTotals` — reverted, `FlowSummary` does not exist and the package fails to compile.
**Baseline:** before=278 after=277 <!-- measured: grep -hE '^func Test' stats/internal/store/*_test.go | wc -l -->
**Commit:** feat(stats): replace flow health tables with period totals
**After:** Task 1, 2, 3, 4, 5, 6, 7

**Decision:** flow-health-tiles
**Decision:** retain-stored-data

Correction (2026-10-17): `MeanCostPerFinished` shipped as `*float64`, not `float64` — nil when nothing finished, the same absence-never-zero rule as the other two means; `FlowSummary` takes a `Period` like every sibling store query rather than `from, to`. The view's row is one object, not a list. `App.tsx`'s model-disabled reason named the removed tables and was widened into `**Files:**`; `rfc3339Ptr` lost its last caller with the guards DTO and was deleted from `stats.go`.

- [x] 9. Self-review fixes: drop the Per change block
  - [x] **Step 1:** In `stats/web/src/views/SelfReview.tsx` delete the `Per change` panel and `perChange()`; update `renders self-review fixes` in `stats/web/src/views/views.test.tsx` to assert no `Per change` heading.
  - [x] **Step 2: Verify.** `cd stats/web && npx tsc -b && npx vitest run -t 'self-review'`. Commit.

**Build:** green
**Files:** `stats/web/src/views/SelfReview.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** none — repairs existing tests: `renders self-review fixes`.
**Regression:** none — the task declares no tests
**Baseline:** before=159 after=159 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** refactor(stats): drop the per-change block from self-review fixes
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8

**Decision:** self-review-drop-per-change

- [x] 10. Visual specs and baselines follow the new dashboards
  - [x] **Step 1:** Start the disposable UI-test stack on `127.0.0.1:4174` per `.flow/project.md` `## run` (never `flowd` on 4173).
  - [x] **Step 2:** Delete `stats/web/tests/visual/runs.spec.ts` and the cache-efficiency describe in `stats/web/tests/visual/views.spec.ts`; drop `runs` and `cache-efficiency` from `stats/web/tests/visual/full-app-suite.spec.ts` and `RUNS_QUERY` from `stats/web/tests/visual/support.ts` if unused; delete their PNGs; adjust `controls.spec.ts`, `combinations.spec.ts`, `reviewers-decisions.spec.ts`, `self-review.spec.ts`, `baseline.spec.ts` assertions that name removed columns or blocks.
  - [x] **Step 3:** Re-capture changed baselines with the project's `capture` command (`--update-snapshots`) for the affected specs, then run the `verify` command for all visual specs: exit 0.
  - [x] **Step 4: Verify.** `cd stats/web && npx tsc -b`; `scripts/resolve-visual-screenshots.sh . baseline.spec.ts`; `scripts/check-visual-verification.sh .`. Commit.

**Build:** green
**Files:** `stats/web/tests/visual/runs.spec.ts`, `stats/web/tests/visual/views.spec.ts`, `stats/web/tests/visual/full-app-suite.spec.ts`, `stats/web/tests/visual/support.ts`
**Allowed-collateral:** `stats/web/tests/visual/**`
**Tests:** none — repairs existing tests: `views.spec.ts`, `full-app-suite.spec.ts`, `controls.spec.ts`, `combinations.spec.ts`, `reviewers-decisions.spec.ts`, `self-review.spec.ts`, `baseline.spec.ts`.
**Regression:** none — the task declares no tests
**Baseline:** before=159 after=159 <!-- measured: cd stats/web && grep -rhE '^[[:space:]]*(it|test)\(' src --include='*.test.ts' --include='*.test.tsx' | wc -l -->
**Commit:** test(stats): update visual specs and baselines for the reworked dashboards
**After:** Task 1, 2, 3, 4, 5, 6, 7, 8, 9

## Fix round 1 (2026-10-17)

Operator answers to the visual verifier's questions: extend the UI-test seed so every reworked view renders populated; make dollar amounts consistent; a change with no stage runs reads `—`, not a zero duration and a zero dollar amount.

- [x] 11. Consistent cost precision and `—` for a change with no runs
  - [x] **Step 1: Cost precision.** In `stats/web/src/format.ts` `formatUsd` renders exactly 2 fraction digits (`minimumFractionDigits: 2, maximumFractionDigits: 2`) everywhere it is used — decision `cost-two-decimals`. Add vitest `formatUsd always renders two decimals` in a new `stats/web/src/format.test.ts`: `2.957` → `$2.96`, `2.75` → `$2.75`, `0.028` → `$0.03`, `3` → `$3.00`.
  - [x] **Step 2: No-runs row.** In `stats/web/src/views/StateBoard.tsx` the `Cost` and `Wall clock` cells render `—` when `runCount` is 0 (Runs and Errors keep `0`) — decision `no-runs-dash`. Add vitest `state board shows a dash for cost and wall clock of a change with no runs` in `stats/web/src/views/views.test.tsx`.
  - [x] **Step 3: Verify.** Run the task's **Tests:** selectors; fix every existing test whose expected dollar string changes. Commit.

**Build:** green
**Files:** `stats/web/src/format.ts`, `stats/web/src/format.test.ts`, `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx`
**Allowed-collateral:** `stats/web/src/**/*.test.ts`, `stats/web/src/**/*.test.tsx`
**Tests:** `formatUsd always renders two decimals`, `state board shows a dash for cost and wall clock of a change with no runs`
**Regression:** `formatUsd always renders two decimals` — reverted, `$2.957` is rendered and the `$2.96` assertion fails.
**Baseline:** unverified: before=159 after=161
**Commit:** fix(stats): render costs at two decimals and a dash for changes with no runs
**After:** Task 1, 2, 4, 5, 6, 7, 8, 9, 10

**Decision:** cost-two-decimals
**Decision:** no-runs-dash

- [x] 12. UI-test seed covers every reworked view; baselines re-captured
  - [x] **Step 1: Seed.** Extend `stats/cmd/uitest-seed/seed.go` through the store's exported types only (never raw SQL) so the disposable stack renders populated: at least two recorded decisions with different implementer models (Decisions' `By class and execution` and `By implementer model` tables), review-panel dispatches with critical/important/minor findings (Reviewers' Efficiency and Flow health's `C/I/Mi`), one errored stage run (State board `Errors`), and a guard fire (Flow health `Guard fires`). Keep `fixtureNow`-anchored timestamps so the fixture stays deterministic.
  - [x] **Step 2: Baselines.** Bring up the UI-test stack on `127.0.0.1:4174` (never `flowd` on 4173), re-capture every changed baseline with the project's `capture` command, then run `verify` for all visual specs: exit 0.
  - [x] **Step 3: Verify.** `make -C stats web-build`, then `go test ./internal/store -run TestRecordDecisionKeepsSuppliedRecordedAt -count=1` and `scripts/check-visual-verification.sh .`. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Allowed-collateral:** `stats/web/tests/visual/**`, `stats/cmd/uitest-seed/**`
**Tests:** none — repairs existing tests: the visual specs under `stats/web/tests/visual/`.
**Regression:** none — the task declares no tests
**Baseline:** unverified: before=159 after=159
**Commit:** test(stats): seed decisions, findings, errors and guard fires for the UI-test stack
**After:** Task 5, 11

Correction (2026-10-17): `RecordDecision` always stamped `now()`, so seeded decisions fell outside the pinned visual period; it now keeps a non-zero caller-supplied `RecordedAt` (zero still means `now()`), covered by `TestRecordDecisionKeepsSuppliedRecordedAt` — `records.go`/`records_test.go` widened into `**Files:**`. Baselines were re-captured with `--update-snapshots=all`, since the 2% tolerance let the stale empty Reviewers baseline pass. Measured baselines: task 11 before=166 after=168, task 12 before=168 after=168, counted from the SPA test run.

## In-run fix 1 (2026-10-17)

`flow.verify` of the integrate run reported three failing commands: `scripts/check-plan-provenance.sh` (`design.md`, `tasks.md` quoted rendered zero strings with no provenance), `scripts/check-plan-shape.sh` (tasks 11 and 12 named shared files with no **After:** ordering and ran suites in their verify steps), and `go test ./... -race -count=1` (`TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` fails whenever the caller exports `FLOW_RECORDS_ADDR`). The two plan-text defects are corrected in this round's planning commit; task 13 fixes the test.

- [x] 13. In-run fix 1 — flow.verify: plan provenance, plan shape, and a records-address leak in the render-to-guard test
  - [x] **Step 1: Reproduce.** `cd stats && FLOW_RECORDS_ADDR=http://127.0.0.1:4173 go test ./cmd/flow -run TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard -count=1` fails with `OUTSTANDING: … no change summary recorded`: the test pins `FLOW_ADDR` to its fake daemon, but the guard's `flow record` reads resolve `FLOW_RECORDS_ADDR` first, so an exported value sends them to the dev daemon. <!-- measured: same command, 2026-10-17, exit 1; without FLOW_RECORDS_ADDR exit 0 -->
  - [x] **Step 2: Fix.** In `stats/cmd/flow/record_test.go` pin `FLOW_RECORDS_ADDR` to `srv.URL` beside the existing `t.Setenv("FLOW_ADDR", srv.URL)`, and extend the comment above it to name both variables.
  - [x] **Step 3: Verify.** Re-run the Step 1 command: exit 0. Commit.

**Build:** green
**Files:** `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard`
**Regression:** `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard` — with the new `t.Setenv("FLOW_RECORDS_ADDR", srv.URL)` removed and `FLOW_RECORDS_ADDR` exported to the dev daemon, the test fails with an `OUTSTANDING` verdict.
**Baseline:** unverified: before=0 after=0
**Commit:** fix(flow): pin the records address in the render-to-guard test
**After:** Task 12
