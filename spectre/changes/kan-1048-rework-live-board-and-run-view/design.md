# kan-1048-rework-live-board-and-run-view — design

## Context

The approved prototype is the specification: `.superpowers/sdd/live-board-prototype.html` in the
change's worktree, published at https://claude.ai/artifact/64Fdyq1iuCfNuwvhzra86R (version 6,
approved 2026-10-28). The board is its default view, the opened run its `#run` view. Any departure
found later is fixed to match it.

Today the board (`stats/web/src/views/StateBoard.tsx`) is one stat panel over a `DataTable`; the
run view (`stats/web/src/views/RunDetail.tsx`) is five stat panels, `StageTimeline` (lane per
command) and `StageRunTable` (filterable by command, stage and model). `App.tsx` renders the
`DashboardBar` (period, project, model) on every route but weekly-costs.

## Decisions

### Filters on the board, none on the run

**ID:** board-chip-filters-run-none
**Status:** active
**Chosen:** the board carries the prototype's chip bar and clickable state tiles; the opened run carries no filter of any kind — the operator's answers in brainstorming ("I especially don't like filters" of the old controls, "I need them" on the board, "in opened run remove them completely").
**Considered:** keeping the shared `DashboardBar` on both routes — the operator named it as one of the disliked filters; chip filters on the run (prototype v4) — rejected by the operator.

### The board filters client side over every change

**ID:** board-client-side-filters
**Status:** active
**Chosen:** the board fetches the state board for the whole history (`from` the epoch, `to` now) once and applies Project, Updated, With errors, search and state client side — the board holds tens of rows, and every chip count must reflect the other filters without a request per click.
**Considered:** server-side period/project parameters per chip — a request per click for a list this size, and counts on the Project chips would need a second query.

### Hide the dashboard bar on the board and run routes

**ID:** hide-dashboard-bar
**Status:** active
**Chosen:** `App.tsx` skips `DashboardBar` on `state-board` and on the run route, as it already does on weekly-costs; `RunDetail` drops its `model` prop and `useRunDetail` its model restriction.
**Considered:** keeping the bar with only the model control on the run — contradicts board-chip-filters-run-none.

### Table filters switch off through one DataTable prop

**ID:** datatable-controls-prop
**Status:** active
**Chosen:** `DataTable` gains `controls?: boolean` (default `true`); `false` renders no `FilterBar` and no search box. The board and the stage-run table pass `false`, and drop every `filterable` flag.
**Considered:** an empty `searchableKeys` — still renders the search box.

### Open stage and done tasks come from the store

**ID:** board-open-stage-done-tasks
**Status:** active
**Chosen:** `LiveStateBoard` adds a `LEFT JOIN LATERAL` picking the change's latest stage run with `ended_at IS NULL` (stage, started_at) and a count of distinct implementer `task_id`s whose dispatch carries a `commit_sha`; the row and DTO gain `currentStage`, `currentStageStartedAt`, `doneTasks`, each omitted when absent. The task total stays `currentTasks`.
**Considered:** inferring the open stage in the SPA from stage runs — a request per change; filtering dispatches on `outcome` — free text with no check constraint.

### Model, version and effort per row

**ID:** model-version-effort
**Status:** active
**Chosen:** a stage-run row shows `<family>-<version> · <effort>` (e.g. `opus-4.7 · high`), the model from the bag's `models` keys shortened by one formatter (`claude-opus-4-7` → `opus-4.7`, several models joined by `, `), the effort from the bag's top-level `effort` (`readEffort`); a dispatch row uses its bucket's `model` and its `signals.sidechain.served_efforts` most-served value. A missing effort shows the model alone.
**Considered:** fetching `/api/v1/records/{project}/{change}` for the recorded effort — a second request and a join for the same figure the bag already observes.

### Group stage runs by session and phase

**ID:** session-phase-groups
**Status:** active
**Chosen:** stage runs group by `sessionId` (falling back to `command` when absent), groups in start order, each labelled `Session N · <phases>` with `HH:MM–HH:MM|now · duration · cost`; a stage's phase is planning (kickoff, brainstorm, design-approval, create-artifacts, writing-plans, decide), implementation (implement, review-panel, verify, visual-verify), landing (integrate, cleanup), other otherwise — the operator could not tell `/flow` from `/flow kan-1048`.
**Considered:** the raw command text (today) — rejected by the operator.

### Timeline lanes per stage

**ID:** timeline-lane-per-stage
**Status:** active
**Chosen:** one lane per stage in first-seen order, bars filled by outcome (completed `--status-ok`, failed `--status-critical`, open hatched `--st-progress`), an hour-tick axis, a dashed now marker, full panel width, a hover tooltip (stage, session, outcome, duration, cost).
**Considered:** lane per command (today) — two lanes say nothing about which stage took the time.

### Session activity from Claude Code hooks

**ID:** session-activity-hooks
**Status:** active
**Chosen:** a new hook `hooks/flow-session-activity.py` (Python 3, stdlib) on PreToolUse matching AskUserQuestion (`question` + the first question text), Notification with a permission prompt (`permission` + the notification message), UserPromptSubmit and PostToolUse (`working`) and Stop (`idle`) sends `PUT /api/v1/sessions/{session_id}/activity` `{state, detail}` to flowd, failing open like `hooks/flow-active-change.py`. flowd upserts one `session_activity` row per session (migration `0052_session_activity.sql`); `LiveStateBoard` joins it on the open stage run's `session_id` and returns `activity`, `activityDetail`, `activitySince`. The board shows the badge on Running now cards (amber border, sorted first, when not working) and beside the current stage; the run view shows it under Now running. Operator-approved prototype v8.
**Considered:** keying activity by change name — the hook would need the store round trip `flow-active-change.py` makes on every event; the session id is in every hook payload and the open stage run already carries it. Parsing the transcript in the harvester — lags a harvest cycle and has to infer a pending prompt from an unanswered tool_use.

### Follow-up for the other dashboards

**ID:** other-dashboards-followup
**Status:** active
**Chosen:** the other eight dashboards are reworked under a separate Jira task, filed at planning; this change touches only the board and the run route — the operator's scope answer.
**Considered:** every dashboard in this change — the operator chose board + run first.

## Open questions

## Recommended approach

One server task adds the board fields and the session-activity endpoint; one task adds the activity hook and registers it; one SPA task rebuilds the board (chip bar, tiles,
running-now cards, grouped table) and hides the dashboard bar there; one SPA task rebuilds the run
view (header, KPI strip, stage-lane timeline, session-grouped table with model·effort, cost rail)
and hides the bar on the run route; one task updates the UI-test seed and the visual specs and
recaptures. Rejected alternatives: server-side chip filters (board-client-side-filters), records
endpoint for effort (model-version-effort), command-text groups (session-phase-groups).

## Critical files

- `stats/internal/store/aggregate.go` — `LiveStateRow` and the `LiveStateBoard` query.
- `stats/internal/store/aggregate_test.go` — store tests for the board.
- `stats/internal/api/stats.go` — `stateBoardRowDTO` and `toStateBoardDTOs`.
- `stats/web/src/api.ts` — `StateBoardRow`, `StageRunDTO` types.
- `stats/web/src/App.tsx` — `DashboardBar` visibility, run route props.
- `stats/web/src/views/StateBoard.tsx` — the board.
- `stats/web/src/views/RunDetail.tsx` — the run view.
- `stats/web/src/components/StageTimeline.tsx` — the timeline.
- `stats/web/src/components/StageRunTable.tsx` — the stage-run table.
- `stats/web/src/components/DataTable.tsx` — the `controls` prop.
- `stats/web/src/hooks/useRunDetail.ts` — run data, model restriction removed.
- `stats/web/src/format.ts` — model label and relative-time formatters.
- `stats/web/src/styles.css`, `stats/web/src/theme.css` — styles and the state colour tokens.
- `stats/cmd/uitest-seed/seed.go` — UI-test seed.
- `stats/internal/store/migrations/0052_session_activity.sql`, `stats/internal/store/activity.go` — activity table and upsert.
- `stats/internal/api/server.go`, `stats/internal/api/activity.go` — the PUT route.
- `hooks/flow-session-activity.py` — the activity hook; `setup.sh` — its registration.
- `stats/internal/repotest/flow_session_activity_hook_test.go` — the hook's test.
- `stats/web/tests/visual/*.spec.ts` and snapshots — captures.

## Existing code to reuse

- `DataTable` (`stats/web/src/components/DataTable.tsx`) — sorting, paging, row render.
- `StatPanel`, `Panel`, `ViewFrame` — panel chrome.
- `readModels`, `readEffort`, `readCacheReadTokens` (`stats/web/src/metrics.ts`) — bag readers.
- `formatUsd`, `formatMs`, `formatInt`, `formatDateTime` (`stats/web/src/format.ts`).
- `projectLabel` (`stats/web/src/lib/projectLabel.ts`) — project display names.
- `useStatsView` (`stats/web/src/hooks/useStatsView.ts`) — the board fetch.
- `hooks/flow-active-change.py` — store address resolution and fail-open shape for the new hook.
- `warn_unregistered_hook` and the hook registration block in `setup.sh`.
- Index `stage_runs_change_id` (`stats/internal/store/migrations/0004_query_indexes.sql`) for the lateral join.

## Verification

- `cd stats && make test` — store and API tests, including the new board-field test.
- `cd stats/web && npm test` and `npx tsc -b`.
- `cd stats/web && { ../../scripts/check-dev-stack-fresh.sh ../.. || make -C .. ui-test-up; } && npm run test:visual` — the recaptured board and run baselines.
- Open the dev stack at `http://127.0.0.1:4173/`, compare the board and a run to the prototype side by side: chips filter and counts move, state tiles toggle, Finished starts collapsed, a running change shows a card, answering an AskUserQuestion flips its badge from Awaiting your answer to Working within one refresh, the run has no filter controls and its groups read `Session N · <phase>`.

## Live check

`flowd` serves `flow` state for every project, and this change alters the board query it runs.
Using `scripts/live-store-copy.sh`, start a workspace flowd on a scratch copy of the dev store and
request `GET /api/v1/stats?view=state-board&from=1970-01-01T00:00:00Z&to=<now>`. Record the number
of rows, and how many rows carry `currentStage`, `doneTasks` and `activity`, before and after the change. After
the change, `kan-1048-rework-live-board-and-run-view` (or any change with an open stage run) shows
`currentStage` and an `activity` matching what that session is doing (send one `PUT /api/v1/sessions/<its session id>/activity` with `{"state":"question","detail":"probe"}` to the workspace flowd and read it back). Failure looks like this: no row carries `currentStage` while `flow status` shows an
open stage, or the row count differs from the count before.
