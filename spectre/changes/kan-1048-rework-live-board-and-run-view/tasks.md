# kan-1048-rework-live-board-and-run-view

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** the Live state board and the opened-run view match the approved prototype (`proposal.md`).

**Spec:** `design.md`; prototype `.superpowers/sdd/live-board-prototype.html` (https://claude.ai/artifact/64Fdyq1iuCfNuwvhzra86R, v8).

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1048-rework-live-board-and-run-view`; never the main checkout.
- `cd stats && gofmt -l .` prints nothing; `cd stats && make -s web-dist && go vet ./...` and `cd stats/web && npx tsc -b` exit 0.
- Colours come from `stats/web/src/theme.css` tokens only; new tokens (`--st-started`, `--st-progress`, `--st-finished`) are declared there in all three theme blocks with the prototype's values.
- The UI-test stack is started once by the project's `## run` UI-test row (`make -C stats ui-test-up`); no task step starts, stops or rebuilds it.

- [x] 1. Board fields and session activity in the store and API
  - [x] **Step 1: Test first.** `TestLiveStateBoardCarriesOpenStageDoneTasksAndActivity` seeds a change with a closed and an open stage run (open run's `session_id` `s-1`), two implementer dispatches with a `commit_sha` for two distinct `task_id`s and one without, and a `session_activity` row for `s-1` (`question`, `Proceed?`); asserts `CurrentStage`, `CurrentStageStartedAt`, `DoneTasks == 2`, `Activity == "question"`, `ActivityDetail`, `ActivitySince`; a second change with no open run has all six nil. Run it red.
  - [x] **Step 2: Migration.** `stats/internal/store/migrations/0052_session_activity.sql`: `session_activity (session_id TEXT PRIMARY KEY, state TEXT NOT NULL CHECK (state IN ('working','question','permission','idle')), detail TEXT NOT NULL DEFAULT '', updated_at TIMESTAMPTZ NOT NULL DEFAULT now())`.
  - [x] **Step 3: Store.** `stats/internal/store/activity.go`: `SetSessionActivity(ctx, sessionID, state, detail string) error` — upsert, `updated_at = now()` only when `state` changes (so `ActivitySince` measures the wait). `LiveStateRow` gains `CurrentStage *string`, `CurrentStageStartedAt *time.Time`, `DoneTasks *int`, `Activity *string`, `ActivityDetail *string`, `ActivitySince *time.Time`; `LiveStateBoard` adds the `LEFT JOIN LATERAL` on the latest `stage_runs` row with `ended_at IS NULL` (stage, started_at, session_id), a `LEFT JOIN session_activity sa ON sa.session_id = open_run.session_id`, and `(SELECT COUNT(DISTINCT d.task_id) FROM dispatches d WHERE d.change_id = c.id AND d.role = 'implementer' AND d.commit_sha IS NOT NULL AND d.task_id IS NOT NULL)`, NULL when zero and no implementer dispatch exists.
  - [x] **Step 4: API.** `stats/internal/api/activity.go`: `PUT /api/v1/sessions/{session}/activity` with `{"state","detail"}` — 400 on an unknown state or a detail over 500 bytes (truncated by the hook, refused here), 204 on success; registered in `server.go`. `stateBoardRowDTO` gains `currentStage`, `currentStageStartedAt`, `doneTasks`, `activity`, `activityDetail`, `activitySince`, all `omitempty`; `toStateBoardDTOs` maps them. `TestSessionActivityPut` covers 204, 400 on a bad state, and the board reading the value back. `stats/web/src/api.ts` `StateBoardRow` gains the six optional fields.
  - [x] **Step 5: Verify.** `cd stats && go test ./internal/store ./internal/api -run 'TestLiveStateBoard|TestSessionActivityPut' -count=1`; the global-constraint lint lines. Commit.

**Build:** green
**Files:** `stats/internal/store/migrations/0052_session_activity.sql`, `stats/internal/store/activity.go`, `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/activity.go`, `stats/internal/api/activity_test.go`, `stats/internal/api/server.go`, `stats/internal/api/stats.go`, `stats/web/src/api.ts`, `stats/internal/web/embed_test.go`
**Allowed-collateral:** `stats/internal/api/stats_test.go`, `stats/internal/client/client_test.go`
**Tests:** `TestLiveStateBoardCarriesOpenStageDoneTasksAndActivity`, `TestSessionActivityPut`
**Regression:** `TestLiveStateBoardCarriesOpenStageDoneTasksAndActivity` fails — the board loses the open stage, the done count or the activity; `TestSessionActivityPut` fails — the route stops storing or validating activity.
**Baseline:** before=30 after=31 <!-- measured: grep -c -E '^func Test' stats/internal/store/aggregate_test.go @ origin/main -->
**Commit:** feat(stats): open stage, done tasks and session activity on the state board
**After:** none

Correction (2026-10-28): `SetSessionActivity` joined the `api.StatsStore` interface, so `stats/internal/web/embed_test.go`'s fake gained the method too; `**Files:**` carries it.

**Decision:** board-open-stage-done-tasks
**Decision:** session-activity-hooks

- [x] 2. The session-activity hook
  - [x] **Step 1: Test first.** `TestFlowSessionActivityHook` (repotest, the shape of `flow_active_change_hook_test.go`) runs the hook against an `httptest` server for each event: PreToolUse `AskUserQuestion` → `PUT .../sessions/<id>/activity` `{"state":"question","detail":"<first question>"}`; Notification whose message names a permission → `permission`; UserPromptSubmit and PostToolUse → `working`; Stop → `idle`; an unreachable store, a payload without `session_id` and malformed JSON each exit 0 with no output. Run it red.
  - [x] **Step 2: Hook.** `hooks/flow-session-activity.py`, Python 3 stdlib only: reads the payload from stdin, maps `hook_event_name` as above, truncates `detail` to 500 bytes, sends the PUT with a 2 s timeout to `FLOW_STORE_ADDR` or `http://127.0.0.1:4173` (the resolution `hooks/flow-active-change.py` uses), never prints, always exits 0.
  - [x] **Step 3: Register.** `setup.sh`'s Claude Code block registers it on `PreToolUse` (matcher `AskUserQuestion`), `PostToolUse`, `Notification`, `UserPromptSubmit` and `Stop`, with a `warn_unregistered_hook` line like `flow-active-change`'s.
  - [x] **Step 4: Verify.** `cd stats && go test ./internal/repotest -run TestFlowSessionActivityHook -count=1`; `scripts/check-references.sh`. Commit.

**Build:** green
**Files:** `hooks/flow-session-activity.py`, `setup.sh`, `stats/internal/repotest/flow_session_activity_hook_test.go`
**Tests:** `TestFlowSessionActivityHook`
**Regression:** `TestFlowSessionActivityHook` fails — an event stops reporting its state, or the hook blocks on an unreachable store.
**Commit:** feat(hooks): report session activity to flowd
**After:** Task 1

Correction (2026-10-28): Step 2 named `FLOW_STORE_ADDR`; the hook reads `FLOW_ADDR`, the variable `hooks/flow-active-change.py` actually resolves the store from. `setup.sh` warns with the per-event registration rather than registering, as every other Claude Code hook there does.

**Decision:** session-activity-hooks

- [x] 3. Live state board
  - [x] **Step 1: Test first.** vitest in `views.test.tsx`: `state board filters by chips and state tiles` (Project chip narrows rows and tile counts; Updated `24h` hides an older change; With errors only; search; a state tile toggles and un-collapses its group; Clear filters resets), `state board running now cards` (one card per row with `currentStage`; a `question` card shows "Awaiting your answer", its detail and an amber `needs` class, sorted before a `working` card), `state board groups by state` (In progress, Started, Finished order, Finished collapsed). Remove assertions on the old column filters, the Changes stat panel and the period/project bar. Run red.
  - [x] **Step 2: DataTable.** `controls?: boolean` (default true); `false` renders no `FilterBar`. Unit test `DataTable renders no filter bar when controls is false` in `DataTable.test.tsx`.
  - [x] **Step 3: Board.** `StateBoard.tsx` per the prototype's board view: fetch from the epoch to now; chip bar (Project with counts, Updated 24h/7d/30d/All default 7d, With errors only, Clear filters, search); five tiles (three state tiles as toggles, Running now, Cost in period); Running now cards (name, state pill, stage and elapsed, activity badge with wait and detail, progress `doneTasks/currentTasks`, project, tasks and cost); state-grouped table (Change with live dot, Project, State pill, Current stage plus non-working badge, Tasks with growth, Updated relative with ISO title, Runs, Cost with bar, Errors red when non-zero); `pipeline-fix-*` still hidden. Relative-time and model-label formatters go in `format.ts` with `format.test.ts` cases `formatRelative`.
  - [x] **Step 4: App.** `App.tsx` hides `DashboardBar` on `state-board`; the `state-board` entry in `MODEL_DISABLED_REASONS` goes; `App.test.tsx` asserts no dashboard bar on `#/state-board`.
  - [x] **Step 5: Styles.** `styles.css` and `theme.css` carry the prototype's pill, tile, chip, card, activity-badge, group-row and cost-bar rules on theme tokens.
  - [x] **Step 6: Verify.** `cd stats/web && npx vitest run -t 'state board|DataTable|formatRelative|dashboard bar'`; `npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/App.tsx`, `stats/web/src/App.test.tsx`, `stats/web/src/format.ts`, `stats/web/src/format.test.ts`, `stats/web/src/styles.css`, `stats/web/src/theme.css`
**Tests:** `state board filters by chips and state tiles`, `state board running now cards`, `state board groups by state`, `formatRelative`
**Regression:** the board vitests fail — a chip stops filtering, cards lose the activity badge or ordering, or groups lose their order.
**Commit:** feat(stats): rework the live state board
**After:** Task 1

Correction (2026-10-28): the plan declared Step 2's `DataTable` `controls` prop and its unit test; the board shipped its own grouped `<table>` instead and `DataTable` is untouched, because the prototype's single-header, state-grouped table cannot be drawn by `DataTable` and the prop would have had no caller (design.md's grouped-tables-own-markup supersedes datatable-controls-prop). `App.test.tsx`'s period-from-URL cases moved from `#/cost-per-change`, an unknown route that fell back to the board, to `#/trend`, since the board no longer reads the dashboard period.

**Decision:** board-chip-filters-run-none
**Decision:** board-client-side-filters
**Decision:** hide-dashboard-bar
**Decision:** grouped-tables-own-markup

- [ ] 4. Opened run view
  - [ ] **Step 1: Test first.** `RunDetail.test.tsx`: `run detail groups stage runs by session and phase` (two sessions → `Session 1 · planning` and `Session 2 · implementation` with span, duration, cost; missing `sessionId` falls back to command), `run detail shows model version and effort` (`claude-opus-4-7` + effort `high` → `opus-4.7 · high`; no effort → model alone; dispatch row uses its bucket model and most-served effort), `run detail has no filter controls` (no select, no search box, no dashboard bar), `run detail header and kpis` (state pill, Jira key from the name prefix, Now running with activity, Total cost, Wall clock, Input tokens with cached share, Failed runs), `stage timeline draws a lane per stage coloured by outcome`. Delete the model-variable cases (`:391-450`). Run red.
  - [ ] **Step 2: Data.** `useRunDetail.ts` drops the `model` parameter; the run view also reads its row from the state board (`project`, whole history) for state, activity and `updatedAt`.
  - [ ] **Step 3: View.** `RunDetail.tsx` per the prototype's `#run` view: back link, header, KPI strip in one panel, timeline panel with legend, then stage runs and the Cost by stage rail side by side (one column under 900 px). `App.tsx` hides `DashboardBar` on the run route and stops passing `model`.
  - [ ] **Step 4: Timeline.** `StageTimeline.tsx`: lane per stage, fill by outcome, hatched open bar, hour ticks, now marker, full width (`viewBox` with `width="100%"`), hover tooltip.
  - [ ] **Step 5: Table.** `StageRunTable.tsx`: its own `<table>` (grouped-tables-own-markup), no filter controls; session group rows; columns Stage, Outcome, Started, Duration, Tokens in, Model · effort, Cost; dispatch rows under the existing toggle.
  - [ ] **Step 6: Verify.** `cd stats/web && npx vitest run -t 'run detail|stage timeline'`; `npx tsc -b`. Commit.

**Build:** green
**Files:** `stats/web/src/views/RunDetail.tsx`, `stats/web/src/views/RunDetail.test.tsx`, `stats/web/src/hooks/useRunDetail.ts`, `stats/web/src/components/StageTimeline.tsx`, `stats/web/src/components/StageRunTable.tsx`, `stats/web/src/metrics.ts`, `stats/web/src/App.tsx`, `stats/web/src/styles.css`
**Allowed-collateral:** `stats/web/src/App.test.tsx`, `stats/web/src/format.ts`
**Tests:** `run detail groups stage runs by session and phase`, `run detail shows model version and effort`, `run detail has no filter controls`, `run detail header and kpis`, `stage timeline draws a lane per stage coloured by outcome`
**Regression:** the run vitests fail — groups fall back to command text, the model loses version or effort, a filter returns, or the timeline loses its stage lanes.
**Commit:** feat(stats): rework the opened run view
**After:** Task 3

**Decision:** board-chip-filters-run-none
**Decision:** hide-dashboard-bar
**Decision:** model-version-effort
**Decision:** session-phase-groups
**Decision:** timeline-lane-per-stage

- [ ] 5. UI-test seed and visual captures
  - [ ] **Step 1: Seed.** `stats/cmd/uitest-seed/seed.go`: one change with an open stage run whose session has `question` activity, one with `working` activity, task-count observations and committed implementer dispatches, two sessions on `kan-201-refactor-thing` with model and effort in their bags.
  - [ ] **Step 2: Specs.** `baseline.spec.ts` captures `dashboard.png` and `run-detail.png` and hovers a timeline bar (`run-detail-hover.png`); `combinations.spec.ts` drops the model-disabled case and captures `combination-chips-plus-tile` (project chip + state tile) instead of filter-plus-search and period-plus-sort; `controls.spec.ts` drives the board's chips, tiles, search and Clear filters by role and name instead of PeriodPicker/ProjectFilter/FilterBar; `run-detail.spec.ts` drops the filter cases and asserts session groups and no filter controls. Locators by role, label or test id only.
  - [ ] **Step 3: Capture.** The project's `capture` row for those specs and the full app suite with `--update-snapshots=all`; read every changed PNG against the prototype; run `compose-mockup-frames.sh` and `measure-visual-properties.sh` against the prototype per the project's visual-verification row and fix until clean or each leftover is named; `rebuild-full-app-suite-zip.sh`. `npm run test:visual` exits 0. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/baseline.spec.ts`, `stats/web/tests/visual/combinations.spec.ts`, `stats/web/tests/visual/controls.spec.ts`, `stats/web/tests/visual/run-detail.spec.ts`, `stats/web/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png`, `stats/web/tests/visual/baseline.spec.ts-snapshots/run-detail-darwin.png`, `stats/web/tests/visual/baseline.spec.ts-snapshots/run-detail-hover-darwin.png`, `stats/web/tests/visual/combinations.spec.ts-snapshots/combination-chips-plus-tile-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-run-detail-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`
**Allowed-collateral:** `stats/web/tests/visual/combinations.spec.ts-snapshots/*.png`, `stats/web/tests/visual/**/*-darwin.png`
**Tests:** `run-detail-hover`, `combination-chips-plus-tile`
**Regression:** the capture specs fail — the board or run page no longer matches its baseline, or a chip/tile stops filtering.
**Commit:** test(stats): capture the reworked board and run view
**After:** Task 4
