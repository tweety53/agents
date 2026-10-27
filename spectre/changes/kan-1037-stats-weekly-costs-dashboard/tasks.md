# kan-1037-stats-weekly-costs-dashboard

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** a Weekly costs dashboard: Wed→Wed totals for the last 8 weeks, split by overall, planning, implementation, review and self-review, plus a picker-driven chart of the same totals.

**Spec:** `design.md` beside this file.

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1037-stats-weekly-costs-dashboard`. Every `**Files:**` path is relative to its root. Never touch the main checkout.
- Never stop the dev workspace's `flowd` or `flow-postgres`.

- [x] 1. Store and API: the `weekly-costs` view
  - [x] **Step 1: Write the failing store test** `TestWeeklyCostsBucketsByWednesdayAndCategory` in `stats/internal/store/aggregate_test.go`, after `TestTrendOverTimeBucketsByDay`. Seed with `seedChange`, `baseBeginInput` and `costRun`, as that test does, setting each input's stage key to the category's key (read `store.BeginStageInput` for the field name). Seed:
    - change `kan-1`:
      - Tue 2026-06-09 23:00Z, `flow.sdd-tdd`, $5. Lands in week Wed 06-03.
      - Wed 2026-06-10 00:00Z, `flow.brainstorm`, $10.
      - Wed 06-10 01:00Z, `flow.sdd-tdd`, $20.
      - Thu 06-11, `flow.review-panel`, $30.
      - Fri 06-12, `flow.verify`, $7. Counts in overall only.
    - change `self-review-kan-1`: Sat 06-13, `flow.sdd-tdd`, $40. Self-review, because the loop name wins.
    - change `kan-1`: Sun 06-14, `flow.self-review`, $3.

    Query `WeeklyCosts(ctx, Period{2026-06-01, 2026-07-01}, &projectKey, nil)` and assert 2 points:
    - point 0: `Week` = 2026-06-03 00:00Z, overall 5, implementation 5, the other three `nil`.
    - point 1: `Week` = 2026-06-10 00:00Z, overall 110, planning 10, implementation 20, review 30, self-review 43.

    Run `cd stats && go test ./internal/store -run TestWeeklyCosts`. It fails to compile, which is the expected RED.
  - [x] **Step 2: Implement** in `stats/internal/store/aggregate.go`, after `TrendOverTime`:

```go unverified:compiles-against-store
// WeeklyCostPoint is one Wednesday-to-Wednesday week's cost totals: overall,
// and the four pipeline categories (design.md, classify-by-stage). A
// category with no costed run that week is nil, never 0.
type WeeklyCostPoint struct {
	Week           time.Time
	OverallUSD     *float64
	PlanningUSD    *float64
	ImplementUSD   *float64
	ReviewUSD      *float64
	SelfReviewUSD  *float64
}

// WeeklyCosts buckets cost by week starting Wednesday 00:00 UTC, split by
// category. A self-review loop change is classified first, so its stages
// never count as planning or implementation; a stage outside every
// category counts in overall only. project and model filter as
// TrendOverTime's do.
func (s *Store) WeeklyCosts(ctx context.Context, period Period, project, model *string) ([]WeeklyCostPoint, error) {
	rows, err := s.pool.Query(ctx, `
		WITH r AS (
			SELECT
				date_trunc('week', sr.started_at - interval '2 days') + interval '2 days' AS week,
				CASE WHEN $4::text IS NULL THEN (sr.metrics->>'cost_usd')::numeric
				     ELSE (sr.metrics->'models'->$4::text->>'cost_usd')::numeric END AS cost,
				CASE
					WHEN c.name LIKE 'self-review-%' OR c.name LIKE 'pipeline-fix-%' OR sr.stage_key = 'flow.self-review' THEN 'self-review'
					WHEN sr.stage_key IN ('plan.session', 'flow.brainstorm', 'flow.design-approval', 'flow.create-artifacts', 'flow.writing-plans', 'flow.decide') THEN 'planning'
					WHEN sr.stage_key = 'flow.sdd-tdd' THEN 'implementation'
					WHEN sr.stage_key = 'flow.review-panel' THEN 'review'
				END AS category
			FROM stage_runs sr
			JOIN changes c ON c.id = sr.change_id
			WHERE sr.started_at >= $1 AND sr.started_at < $2
			  AND ($3::text IS NULL OR c.project_key = $3)
			  AND ($4::text IS NULL OR sr.metrics->'models' ? $4)
		)
		SELECT week,
			SUM(cost),
			SUM(cost) FILTER (WHERE category = 'planning'),
			SUM(cost) FILTER (WHERE category = 'implementation'),
			SUM(cost) FILTER (WHERE category = 'review'),
			SUM(cost) FILTER (WHERE category = 'self-review')
		FROM r
		GROUP BY week
		ORDER BY week
	`, period.From, period.To, project, model)
	if err != nil {
		return nil, fmt.Errorf("store: weekly costs: %w", err)
	}
	defer rows.Close()

	var out []WeeklyCostPoint
	for rows.Next() {
		var p WeeklyCostPoint
		if err := rows.Scan(&p.Week, &p.OverallUSD, &p.PlanningUSD, &p.ImplementUSD, &p.ReviewUSD, &p.SelfReviewUSD); err != nil {
			return nil, fmt.Errorf("store: weekly costs: scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("store: weekly costs: %w", err)
	}
	return out, nil
}
```

    Match the column names (`c.name`, `sr.stage_key`) to the schema; read `stats/internal/store/migrations/` if either differs. `date_trunc` on a `timestamptz` uses the session timezone, so confirm the pool's session runs in UTC, or write `date_trunc('week', (sr.started_at AT TIME ZONE 'UTC') - interval '2 days') AT TIME ZONE 'UTC' + interval '2 days'`. The store test's boundary rows decide which is right.
  - [x] **Step 3: Wire the API** in `stats/internal/api/stats.go`, following `viewTrend` exactly:
    - Add `WeeklyCosts(ctx, period, project, model *string) ([]store.WeeklyCostPoint, error)` to the store interface (near line 49).
    - Add `viewWeeklyCosts viewName = "weekly-costs"` and its place in the view list (near lines 109–121).
    - Add a `case viewWeeklyCosts:` returning `toWeeklyCostDTOs(points)`.
    - Add `weeklyCostPointDTO{Week string \`json:"week"\`, OverallUsd, PlanningUsd, ImplementationUsd, ReviewUsd, SelfReviewUsd *float64}`, with JSON keys `overallUsd`, `planningUsd`, `implementationUsd`, `reviewUsd` and `selfReviewUsd`. `week` is formatted as `trendPointDTO` formats `day`.
    - Add `WeeklyCosts` to both fakes in `stats/internal/api/stats_test.go` (`fakeStore`, `statsFake`), mirroring their `TrendOverTime`. Extend any every-view fixture table that lists views per name (`TestEveryViewCarriesItsRealNumbersThrough` and its siblings) with a `weekly-costs` entry carrying a real number.
  - [x] **Step 4: Verify:** `cd stats && gofmt -w . && go vet ./... && go test ./internal/store -run 'TestWeeklyCosts' && go test ./internal/api`.

**Files:** `stats/internal/store/aggregate.go`, `stats/internal/store/aggregate_test.go`, `stats/internal/api/stats.go`, `stats/internal/api/stats_test.go`
**Tests:** `TestWeeklyCostsBucketsByWednesdayAndCategory`
**Regression:** `TestWeeklyCostsBucketsByWednesdayAndCategory` fails, because `WeeklyCosts` is undefined.
**Baseline:** before=498 after=499 <!-- measured: grep -hE '^func Test[A-Z]' stats/internal/store/*_test.go stats/internal/api/*_test.go | wc -l @ 0e970879 -->
**Build:** green
**After:** none
**Commit:** `feat(stats): weekly Wed-to-Wed cost totals by category`

Correction (2026-10-27): the plan classified on `sr.stage_key`; that column is never written (NULL on every dev-store row), and the stage key lives in `sr.stage`, so the query and the test seed classify on `stage`. The week bucket uses `date_trunc('week', …, 'UTC')`, the time-zone argument, rather than relying on the session timezone. `fakeStore`'s `WeeklyCosts` returns no rows and carries no field, so `stats/internal/api/changes_test.go`, where `fakeStore`'s fields live, is untouched.

**Decision:** classify-by-stage
**Decision:** other-stages-overall-only
**Decision:** wednesday-utc-bucket
**Decision:** absent-category-null

- [x] 2. SPA: Weekly costs view, multi-series chart
  - [x] **Step 1: Write the failing tests** in `stats/web/src/views/views.test.tsx`:
    - Add a `"weekly-costs"` fixture to the fixtures keyed by view slug, alongside the `trend` one: two rows with `week` set to `2026-06-03` and `2026-06-10`. The first row is `overallUsd` 5, `implementationUsd` 5 and the other three `null`; the second is 110, 10, 20, 30, 43.
    - Add `it("weekly costs shows all five totals per week and a five-line chart")`. Render `WeeklyCosts` as the `trend` test renders `Trend`, then assert:
      - the five column headers `Overall`, `Planning`, `Implementation`, `Review` and `Self-review`;
      - `$110.00` appears;
      - the chart legend lists the five series.

    Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'weekly costs'`. It fails, which is the expected RED.
  - [x] **Step 2: Extend `stats/web/src/components/TimeSeriesPanel.tsx`** with an optional `series?: { name: string; points: TimeSeriesPoint[] }[]` prop. When `series` is given:
    - draw one `<polyline>` per series over a shared x axis (the union of days, in order) and a shared y range;
    - use the same null-breaks-the-line rule;
    - give each series a distinct stroke from CSS variables already defined in `stats/web/src/theme.css`, adding `--series-1`…`--series-5` there if none fit;
    - render a legend `<ul data-testid="time-series-legend">` of the names.

    With only `points` it renders exactly as before. Keep the empty state when every series is empty.
  - [x] **Step 3: Add `stats/web/src/views/WeeklyCosts.tsx`**, using `Trend.tsx`'s structure:
    - `lastEightWeeks(now)`: `from` is the UTC Wednesday 00:00 of the current week (`(now.getUTCDay() + 4) % 7` days back) minus 49 days, and `to` is `now`.
    - The table makes `useStatsView<WeeklyCostPoint[]>("weekly-costs", { ...lastEightWeeks(new Date()), project, model })`. It renders a `DataTable` with columns Week, Overall, Planning, Implementation, Review and Self-review, and each cost cell renders `<Unavailable value=… format={formatUsd} />`. It ignores `period`.
    - The chart makes `useStatsView<WeeklyCostPoint[]>("weekly-costs", { from: period.from, to: period.to, project, model })` and renders `<TimeSeriesPanel series={…five series named as the columns…} format={formatUsd} />`.
    - The `ViewFrame` title is "Weekly costs" and the description is "Wednesday-to-Wednesday cost by pipeline phase."
    - Compute the 8-week period once per mount (`useMemo`), so a re-render does not refetch.
  - [x] **Step 4: Register** `"weekly-costs"` in `stats/web/src/api.ts`: add it to `ViewName` and `VIEW_NAMES`, and add an exported `WeeklyCostPoint` interface mirroring `weeklyCostPointDTO`. In `stats/web/src/App.tsx`, add `VIEW_LABELS["weekly-costs"] = "Weekly costs"` and `VIEW_COMPONENTS["weekly-costs"] = WeeklyCosts`.
  - [x] **Step 5: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'weekly costs'`.

**Files:** `stats/web/src/components/TimeSeriesPanel.tsx`, `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/api.ts`, `stats/web/src/App.tsx`
**Allowed-collateral:** `stats/web/src/theme.css`
**Tests:** `weekly costs shows all five totals per week and a five-line chart`
**Regression:** `weekly costs shows all five totals per week and a five-line chart` fails, because no `weekly-costs` view is registered.
**Baseline:** before=162 after=163 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ 0e970879 -->
**Build:** green
**After:** Task 1
**Commit:** `feat(stats): Weekly costs dashboard with five-line chart`

Correction (2026-10-27): the test also asserts the table's own fetch starts on a Wednesday 00:00 UTC, 7 to 8 weeks back, so `lastEightWeeks` carries a check that fails when the anchor is wrong. The 8-week table is given no `period`/`onPeriodChange`, since it ignores the picker and its empty state must not blame the picker. Series strokes and the legend layout are inline styles, so `stats/web/src/styles.css` is untouched.

**Decision:** table-fixed-graph-on-picker
**Decision:** partial-current-week
**Decision:** multi-series-timeseriespanel
**Decision:** absent-category-null

- [x] 3. In-run fix 1 — flow.visual-verify: 8-week table omits empty weeks; one-week chart draws no line; seed shows 2 of 5 lines <!-- measured: points in the weekly-costs capture at 57f0eeca -->
  - [x] **Step 1: Write the failing tests:** in `stats/web/src/views/views.test.tsx`, `weekly costs table always lists the last 8 weeks` — the table renders 8 Wednesday rows, a week with no fixture row reading Unavailable; in `stats/web/src/components/Panel.test.tsx`, `centres a single day on the x axis` — one point at x = 240.
  - [x] **Step 2: Implement:** `WeeklyCosts.tsx` builds the 8 Wednesday keys from `lastEightWeeks` and fills each week the response lacks with nulls; `TimeSeriesPanel.tsx` places a lone day at the plot's centre; `stats/cmd/uitest-seed/seed.go` gives `kan-201-refactor-thing` a `flow.brainstorm` run, a `flow.self-review` run, and keys its SDD run `flow.sdd-tdd`, so the capture shows all five lines.
  - [x] **Step 3: Verify:** run the task's **Tests:** selector, then re-baseline the visual suite.

Operator's answers, 2026-10-27: always 8 rows; centre a single point; extend the seed.

**Files:** `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/components/TimeSeriesPanel.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/components/Panel.test.tsx`, `stats/cmd/uitest-seed/seed.go`
**Tests:** `weekly costs table always lists the last 8 weeks`, `centres a single day on the x axis`
**Regression:** `weekly costs table always lists the last 8 weeks` fails, because the table renders only the weeks the response carries.
**Baseline:** before=164 after=166 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ 9e31d93b -->
**Build:** green
**After:** Task 2
**Commit:** `fix(stats): Weekly costs table lists every week, lone chart point centred`
