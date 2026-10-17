# kan-989-stats-ui-dashboard-review

## Context

- Operator walk-through of the nine dashboards, 2026-10-16; per-dashboard verdicts below are the operator's.
- The `state-board` endpoint also feeds `flow state list/resolve` (`stats/internal/client/client.go` `StateBoardRow`, `stats/cmd/flow/state.go` `listStateBoard`) and `/flow-status`'s `candidates[].updatedBy` — its existing fields and row set are a contract the CLI depends on.
- `changes` has no creation timestamp (`stats/internal/store/migrations/0001_init.sql`); only `updated_at`.
- Reviewers endpoint measured 0.27–0.35 s against 0.01–0.06 s for the other views on the dev stack. <!-- measured: curl -w '%{time_total}' http://127.0.0.1:4173/api/v1/stats/<view>?from=…&to=… @ 2026-10-16 dev stack -->
- One change: every edit is the stats UI or the endpoints only it consumes, and they share `stats.go`, `api.ts`, `App.tsx` and `views.test.tsx`.

## Decisions

### Retain all stored data

**ID:** retain-stored-data
**Status:** active
**Chosen:** removing a view or column deletes only its UI, endpoint, store query and their tests — never a table, column, migration or stored row — operator instruction: nothing is deleted unless completely useless for future decisions, and nothing here qualifies.
**Considered:** dropping now-unread tables/columns (e.g. cache metrics) — ruled out by the operator; the data still feeds future decisions.

### Board filtering and column removal in the SPA

**ID:** board-filter-in-spa
**Status:** active
**Chosen:** the SPA's StateBoard filters out rows whose name starts `pipeline-fix-` and drops the `By`/`Next command` columns; the `state-board` endpoint keeps every row and its `updatedBy`/`nextCommand` fields.
**Considered:** filtering in `LiveStateBoard` — narrows `flow state list/resolve` and `/flow-status`; an opt-in query param — extra server surface for a pure display choice.

### Runs merged into the board as whole-life totals

**ID:** board-run-totals
**Status:** active
**Chosen:** `LiveStateBoard` gains `RunCount`, `CostUsd`, `WallClockMs`, `Errors` per change from one `stage_runs` aggregate LEFT JOINed on `change_id` — a run is the stage rows sharing a `session_token` (a token-less row is its own run), wall clock is `max(ended_at)-min(started_at)` per run summed, cost `metrics->>'cost_usd'`, errors `metrics.signals.{main,sidechain}.{tool_errors,api_errors}` summed — the same definitions `store/runs.go` used; whole life of the change, not bounded by the period. The DTO adds fields only (additive for the CLI client). Then the Runs view, its endpoint, `store/runs.go` and their tests are deleted; RunDetail (`#/run/…`, `useRunDetail`) stays.
**Considered:** expandable per-change rows reproducing the Runs table — operator chose totals plus the existing change link; period-bounded totals — operator chose whole life.

### Stage leaderboard ranks by total cost, hard cap 10

**ID:** leaderboard-top10-total
**Status:** active
**Chosen:** `StageLeaderboard` adds `SUM(cost)` as `TotalCostUsd`, orders `SUM(cost) DESC` with `LIMIT 10` in SQL; the UI shows a `Total cost` column, default sort on it; the tiles count those 10.
**Considered:** mean/median/P90 as the rank — operator chose total; a "show all" toggle — operator chose a hard cap.

### Cache efficiency removed

**ID:** remove-cache-efficiency
**Status:** active
**Chosen:** delete the view, nav entry, `cache-efficiency` endpoint, `Store.CacheEfficiency` and their tests and fakes; data untouched (retain-stored-data).
**Considered:** keeping the endpoint for later — no caller; it can be rebuilt from retained data.

### Reviewers: fix the query, swap columns for Efficiency

**ID:** reviewers-efficiency
**Status:** active
**Chosen:** rewrite the `Reviewers` SQL so the roster slot→description lookup is one CTE (`DISTINCT ON (slot) … ORDER BY recorded_at DESC`) instead of a per-slot correlated `jsonb_path_exists` scan, and add migration `0039` `CREATE INDEX IF NOT EXISTS findings_dispatch_id ON findings(dispatch_id)`; target under 0.1 s on the dev stack. Drop `findingsPerDispatch`/`withdrawnShare` from store row, DTO and UI. Add `Efficiency` = (10×Critical + 4×Important + 1×Minor) / Dispatches, computed in the SPA (derived, not stored), 2 decimals, `—` at 0 dispatches, sortable.
**Considered:** Important weighted above Critical — operator chose Critical 10 / Important 4 / Minor 1; computing Efficiency server-side — it is a display formula over fields the row already carries.

### Decisions: By implementer model replaces Every decision

**ID:** decisions-by-implementer-model
**Status:** active
**Chosen:** remove the `Runs` tile and `Every decision` table; add `By implementer model`, client-side grouped on `implementerModel` (inline/default runs group under their recorded string): Runs, Mean cost, Mean wall-clock, Mean findings (C+I+Mi), Mean fix rounds, sorted by Runs desc. `By class and execution` unchanged. Planner's choice at the operator's delegation.
**Considered:** top-10 most expensive decisions — repeats the leaderboard's angle; by panel roster — roster size already summarised by class.

### Flow health becomes eight totals tiles

**ID:** flow-health-tiles
**Status:** active
**Chosen:** one new `flow-summary` health view backed by `Store.FlowSummary(from,to,project)` returning: changes started (first `stage_runs.started_at` per change in period), changes finished (`state='FINISHED'` and `updated_at` in period, `withdrawn` excluded), total cost (stage-run cost started in period), mean cost per finished change (whole-life cost of those finished), median lead time (finish `updated_at` − first stage-run start, finished changes), mean panel rounds and C/I/Mi findings (from the existing `PanelRounds` rows), guard fires (`guard_runs` fired in period). The view renders eight `StatPanel` tiles only; the `guards`, `stage-redo`, `panel-rounds` endpoints and `StageRedo`/`GuardActivity` store queries lose their last caller and are deleted; `PanelRounds` stays as a Go helper `FlowSummary` reuses.
**Considered:** a `changes.created_at` migration — operator chose the first-stage-run proxy; keeping the guards table — operator chose tiles only.

### Self-review: Per change block removed

**ID:** self-review-drop-per-change
**Status:** active
**Chosen:** delete the `Per change` panel and its client-side `perChange()`; endpoint unchanged.
**Considered:** none — UI-only removal.

### Costs render at two decimals

**ID:** cost-two-decimals
**Status:** active
**Chosen:** `formatUsd` always renders exactly two fraction digits, on every view (operator, fix round 1: "make them consistent").
**Considered:** four fixed decimals — noisy for whole-dollar totals; variable up-to-4 digits — the inconsistency the operator flagged.

### A change with no runs reads a dash

**ID:** no-runs-dash
**Status:** active
**Chosen:** on the live board, `Cost` and `Wall clock` render `—` when the change has 0 stage runs; `Runs` and `Errors` stay `0`, since a count of zero is a measurement.
**Considered:** `0 ms`/`$0.00` — reads as a measured zero for something never measured.

## Open questions

## Live check

On the rebuilt dev stack: `curl -w '%{time_total}'` the `reviewers` endpoint — before ~0.30 s, after under 0.1 s; a figure still near 0.3 s means the fix did not land. Open `#/state-board`: no `pipeline-fix-*` row, Runs/Cost/Wall clock/Errors filled for a finished change with stage runs (non-zero cost); `flow state resolve -C <repo>` still lists `pipeline-fix-*` candidates with `updatedBy` (CLI untouched). `#/runs` and `#/cache-efficiency` fall back to the default view; `#/flow-health` shows eight tiles with non-zero changes finished for a 30-day period.
