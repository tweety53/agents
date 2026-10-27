## Context

- Run cost lives in `stage_runs.metrics->>'cost_usd'`. Each stage run carries a `stage_key`.
- The self-review planning and fix loops are recorded under their own changes, named `self-review-<name>` and `pipeline-fix-<name>-<k>`.
- `Store.TrendOverTime` already sums cost per day over the same table. This change copies its query with a week bucket and per-category filters.
- One change: one store query, one API view, and one SPA view. There is no migration and no new dependency.

## Decisions

### Stages outside the four categories count in overall only

**ID:** other-stages-overall-only
**Status:** active
**Chosen:** Costs from verify, landing, kickoff and the like count in `overall` only, so the category columns may sum to less than overall. Operator's answer, 2026-10-27.
**Considered:** A sixth always-shown "other" line, which lost because the operator asked for exactly five lines.

### Table fixed to 8 weeks, graph on the picker

**ID:** table-fixed-graph-on-picker
**Status:** active
**Chosen:** The table always covers the last 8 Wed→Wed weeks, whatever the picker says. The graph follows the dashboard's `PeriodPicker` from–to, in weekly buckets. Operator's answer, 2026-10-27.
**Considered:** Both following the picker, defaulted to 8 weeks. It lost because the table would stop being the fixed "last 8 weeks" view the operator asked for.

### Current partial week included

**ID:** partial-current-week
**Status:** active
**Chosen:** The 8th row is the current week, from its Wednesday until now. Operator's answer, 2026-10-27.
**Considered:** 8 complete weeks only, which lost because this week's spend would be hidden until next Wednesday.

### Classify by stage, not by role

**ID:** classify-by-stage
**Status:** active
**Chosen:** Cost is classified by the run's `stage_key`, with a loop change name (`self-review-%`, `pipeline-fix-%`) checked first. Everything under `flow.review-panel`, its fix rounds included, is review. Operator's answer, 2026-10-27.
**Considered:** Per-dispatch role. It lost because dispatches pricing could not price carry no cost, so the totals would undercount.

Category map, applied in this order:
- **self-review:** the change name LIKE `self-review-%` or `pipeline-fix-%`, or `stage_key = 'flow.self-review'`
- **planning:** `plan.session`, `flow.brainstorm`, `flow.design-approval`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide`
- **implementation:** `flow.sdd-tdd`
- **review:** `flow.review-panel`

### Weeks in UTC, Wednesday-anchored

**ID:** wednesday-utc-bucket
**Status:** active
**Chosen:** `date_trunc('week', started_at - interval '2 days') + interval '2 days'`, which gives Wednesday 00:00 UTC. The store already holds UTC. The SPA computes the table's `from` the same way: the current week's UTC Wednesday minus 7 weeks.
**Considered:** The browser's local timezone, which lost because SQL would then need a timezone parameter that nobody asked for.

### Multi-series on the existing chart primitive

**ID:** multi-series-timeseriespanel
**Status:** active
**Chosen:** `TimeSeriesPanel` gets an optional `series` prop: named point lists, one stroke each, with a legend. The single-series callers are unchanged.
**Considered:** A charting library, which lost because the app draws all of its charts in inline SVG and adds no dependency for that. A separate chart component, which lost because it would duplicate the axes, scaling and null-gap logic.

### Absent category is null, not zero

**ID:** absent-category-null
**Status:** active
**Chosen:** A week with no costed run in a category returns `null` for it. The table renders `Unavailable` and the chart breaks the line. This follows the app's absence-is-never-zero rule.
**Considered:** `0`, which lost because it would say "measured, free" when nothing was measured.

## Open questions

## Recommended approach

Add `Store.WeeklyCosts(ctx, Period, project, model)` beside `TrendOverTime` in `stats/internal/store/aggregate.go`. It returns one row per week, with `overall` and four `SUM(...) FILTER (WHERE …)` columns per **classify-by-stage**. Wire it as the `weekly-costs` view in `stats/internal/api/stats.go`. Add `stats/web/src/views/WeeklyCosts.tsx`, which makes two `useStatsView` calls: the fixed 8-week table (**table-fixed-graph-on-picker**) and the picker-driven chart (**multi-series-timeseriespanel**). The alternatives were rejected under the decision IDs above.

## Critical files

- `stats/internal/store/aggregate.go` — the new `WeeklyCostPoint` and `WeeklyCosts`
- `stats/internal/store/aggregate_test.go` — the bucket and category test
- `stats/internal/api/stats.go` — the store interface method, `viewWeeklyCosts`, the handler case, and `weeklyCostPointDTO`
- `stats/internal/api/stats_test.go` — the fake stores' `WeeklyCosts` and the every-view fixtures
- `stats/web/src/api.ts` — `ViewName`, `VIEW_NAMES`, and the `WeeklyCostPoint` type
- `stats/web/src/App.tsx` — `VIEW_LABELS` and `VIEW_COMPONENTS`
- `stats/web/src/components/TimeSeriesPanel.tsx` — the `series` prop
- `stats/web/src/views/WeeklyCosts.tsx` — the new view
- `stats/web/src/views/views.test.tsx` — the fixture and test

## Existing code to reuse

- `Store.TrendOverTime` (`stats/internal/store/aggregate.go`): its query shape, project and model filters, and scan loop
- `viewTrend` and `toTrendDTOs` (`stats/internal/api/stats.go`): the handler case and DTO pattern
- `useStatsView`, `Panel`, `DataTable`, `Unavailable` and `formatUsd` (`stats/web/src/`)
- `PeriodPicker` and `DashboardBar` (`stats/web/src/components/`): the existing from–to picker, which reaches every view through `ViewProps.period`
- `costRun`, `baseBeginInput` and `seedChange` (`stats/internal/store/aggregate_test.go`): the test seeding

## Verification

- `cd stats && go test ./internal/store -run TestWeeklyCosts` and `go test ./internal/api`. The every-view tests pick up the new view.
- `cd stats/web && npx vitest run src/views/views.test.tsx`
- Lint: `cd stats && make -s web-dist && go vet ./... && gofmt -l .`, then `cd stats/web && npx tsc -b`.
- Open the Weekly costs view on the workspace stack: the table shows 8 Wednesday-keyed rows, and the chart redraws when the picker changes.

## Live check

Run against the change's own workspace `flowd` instance, which reads the dev store:
- `curl '<workspace url>/api/v1/stats/weekly-costs?from=<UTC Wednesday 7 weeks before this week's>&to=<now>'` returns 8 or fewer rows, each `week` a Wednesday 00:00Z.
- For one week, `overallUsd` equals `select sum((metrics->>'cost_usd')::numeric) from stage_runs where started_at >= '<wed>' and started_at < '<wed>'::timestamptz + interval '7 days'`, run in `psql` against the dev store, read-only.
- **Not working looks like:** a week keyed on a Monday, `overallUsd` differing from the psql sum, or every category column null while `overallUsd` is non-null.
