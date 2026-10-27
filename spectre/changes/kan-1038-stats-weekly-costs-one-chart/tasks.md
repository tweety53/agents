# kan-1038-stats-weekly-costs-one-chart

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** Weekly costs becomes one labelled 8-week chart with toggleable series and week-over-week delta pills, no filter bar; `/flow` planning prototypes visible changes.

**Spec:** `design.md` beside this file; the approved prototype `/Users/tweety53/Projects/agents-worktrees/kan-1038-stats-weekly-costs-one-chart/.superpowers/sdd/weekly-costs-preview.html` (open it in a browser — it is the layout and behaviour reference).

## Global constraints

- Work in `/Users/tweety53/Projects/agents-worktrees/kan-1038-stats-weekly-costs-one-chart`. Every `**Files:**` path is relative to its root. Never touch the main checkout.
- Never stop the dev workspace's `flowd` or `flow-postgres`.
- No new dependency; the chart is inline SVG.

- [x] 1. Weekly costs: one chart with toggles, axes and delta pills
  - [x] **Step 1: Replace the tests** in `stats/web/src/views/views.test.tsx`: delete `weekly costs shows all five totals per week and a five-line chart` and `weekly costs table always lists the last 8 weeks`; add, with time pinned to `2026-06-19T12:00:00Z` as the old tests did and the existing `weekly-costs` fixture (weeks 2026-06-03, 06-10, 06-17):
    - `weekly costs draws one 8-week chart with five toggleable series` — exactly one fetch, from a Wednesday 00:00 UTC 7–8 weeks back, with `project` and `model` undefined even when props pass them; no `table`; 8 x-axis tick labels (`2026-04-29` … `2026-06-17`, or the chart's chosen short form) and the axis titles `Week starting (Wed, UTC)` and `Cost (USD)`; 5 toggle buttons with `aria-pressed="true"`; clicking `Review` sets it `false` and removes every `data-series="Review"` point; clicking again restores it.
    - `weekly costs labels each point with its change from the previous week` — Overall 06-10 ($110 vs $5) shows `▲2100%` with class `delta-up`; Overall 06-17 ($11 vs $110) shows `▼90%` with class `delta-down`; Implementation 06-17 (previous week Unavailable) has no label; no label on the first week. Hovering (`fireEvent.mouseEnter`) the Overall 06-17 point shows a tooltip containing `Overall`, `$11.00` and `▼ 90.0%`.
    Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'weekly costs'` — RED.
  - [x] **Step 2: Implement** in `stats/web/src/views/WeeklyCosts.tsx` per decisions below and the prototype: one `useStatsView("weekly-costs", { ...lastEightWeeks(now), project: undefined, model: undefined })`, rows padded by `everyWeek`; a view-local `WeeklyCostChart` (inline SVG, viewBox ~960×440, Y ticks at 5 even steps from 0 with `formatUsd`, X ticks per week, axis titles as in Step 1); a pill row of `<button aria-pressed>` toggles coloured `var(--series-n)`; lines break at null weeks; dots `data-testid="time-series-point"` `data-series=<name>` with a wide transparent hit stroke; delta pills stacked per `delta-label-stacking`; a positioned tooltip `role="tooltip"` on `mouseenter`/`focus`, hidden on leave/blur. Keep `lastEightWeeks` exported. Remove the `DataTable`/`TimeSeriesPanel` imports. Add `--delta-up` (red) and `--delta-down` (green) to `stats/web/src/theme.css` in the dark default and both light blocks, and the chart/pill/tooltip classes to `stats/web/src/styles.css`. Update the file's header comment.
  - [x] **Step 3: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'weekly costs'`.

**Files:** `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`, `stats/web/src/theme.css`, `stats/web/src/styles.css`
**Tests:** `weekly costs draws one 8-week chart with five toggleable series`, `weekly costs labels each point with its change from the previous week`
**Regression:** `weekly costs labels each point with its change from the previous week` fails, because the view renders no delta labels; `weekly costs draws one 8-week chart with five toggleable series` fails, because the table and picker fetch return.
**Baseline:** before=166 after=166 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ 1be0e995 -->
**Build:** green
**After:** none
**Commit:** `feat(stats): Weekly costs as one chart with toggles and week-over-week deltas`

**Decision:** weekly-chart-own-component
**Decision:** week-over-week-delta
**Decision:** delta-label-stacking

- [x] 2. No filter bar on Weekly costs
  - [x] **Step 1: Write the failing test** `weekly costs renders no filter bar` in `stats/web/src/App.test.tsx`: at `#/weekly-costs` the `Statistics views` nav renders and no period/project/model control does (query the DashboardBar's own labels); at another view the bar still renders. Run `cd stats/web && npx vitest run src/App.test.tsx -t 'filter bar'` — RED.
  - [x] **Step 2: Implement** in `stats/web/src/App.tsx`: do not render `<DashboardBar>` when `route.kind === "view" && route.view === "weekly-costs"`, with a one-line comment citing KAN-1038.
  - [x] **Step 3: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/App.test.tsx -t 'filter bar'`.

**Files:** `stats/web/src/App.tsx`, `stats/web/src/App.test.tsx`
**Tests:** `weekly costs renders no filter bar`
**Regression:** `weekly costs renders no filter bar` fails, because the bar renders on every route.
**Baseline:** before=166 after=167 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/src | wc -l @ 1be0e995 -->
**Build:** green
**After:** Task 1
**Commit:** `feat(stats): hide the filter bar on Weekly costs`

**Decision:** fixed-eight-weeks-no-filters

- [x] 3. Re-baseline the Weekly costs visual spec
  - [x] **Step 1:** in `stats/web/tests/visual/weekly-costs.spec.ts` replace the four tests with: `draws one chart with delta labels in the dark palette` (no filter bar, 5 pressed toggles, axis titles visible, at least one `.delta-up` or `.delta-down` pill, full-page screenshot `weekly-costs.png`); `draws the series and delta tokens in the light palette` (screenshot `weekly-costs-light.png`); `renders the empty state before any recorded run` (one `not-recorded`, screenshot `weekly-costs-empty.png`); `a toggled-off series leaves the chart` (click `Overall`, its points gone, screenshot `weekly-costs-toggled.png`). Keep the clock pinning and seed queries the file uses.
  - [x] **Step 2:** capture per `.flow/project.md` `## visual verification` (`--update-snapshots`), then run the verify row clean; open each PNG and compare with the prototype; fix overlap or clipping before committing.

**Files:** `stats/web/tests/visual/weekly-costs.spec.ts`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-light-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-empty-darwin.png`, `stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-toggled-darwin.png`, `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `draws one chart with delta labels in the dark palette`, `draws the series and delta tokens in the light palette`, `renders the empty state before any recorded run`, `a toggled-off series leaves the chart`
**Regression:** `draws one chart with delta labels in the dark palette` fails, because the old table and picker chart render.
**Baseline:** before=51 after=51 <!-- measured: grep -rhE '^\s*(it|test)\(' stats/web/tests/visual | wc -l @ 1be0e995 -->
**Build:** green
**After:** Task 1, 2
**Commit:** `test(stats): re-baseline Weekly costs visual spec`

Correction (2026-10-27): the plan's spec ran against the UI-test fixture alone, but the fixture's runs fall in 2026-07-22 and 2026-08-12, never two consecutive weeks, so it draws no delta pill; the dark and light tests add one week, 2026-08-05, to the stack's real response through `page.route` (the toggle and empty tests read the stack unaltered). The first capture showed the Task 1 Y scale (prototype's `max(10, …)` floor, tens-steps) flattening a $3 week to the axis; the fix — a 1/2/2.5/5×10ⁿ tick step — and its `cost-tick` unit assertion ride this commit, which widens `**Files:**` to `WeeklyCosts.tsx` and `views.test.tsx`.

- [x] 4. flow: prototype visible changes during planning
  - [x] **Step 1:** in `skills/flow/brainstorm-planner.md`, section **B**, `### The checklist`, add after the frame-specified-design bullet a bullet stating `plan-visual-prototype`: when the change can alter what a user sees — a page, a view, a chart, a dialog — and no mockup frame already specifies it, build a clickable prototype of the proposed result with sample data as a single HTML file, publish it with the Artifact tool, and iterate it on the operator's feedback each round until they approve; the approval is the convergence confirm's **approve the design and move on**. Save the approved file under `<project>-worktrees/<name>/.superpowers/sdd/` and cite it, with its Artifact URL, from `design.md`'s `## Context` as the spec the implementation and `flow.visual-verify` compare against. A change with nothing visible builds none. On a harness without the Artifact tool, write the file and give the operator its absolute path. Cite kan-1038.
  - [x] **Step 2: Verify:** `scripts/check-references.sh` and every other `scripts/check-*.sh` guard `.flow/project.md` `## lint` names, clean.

**Files:** `skills/flow/brainstorm-planner.md`
**Tests:** **none**
**Regression:** none — a skill text; the guards in Step 2 check its references.
**Build:** green
**After:** none
**Commit:** `feat(flow): prototype visible changes during planning`

**Decision:** plan-visual-prototype

- [ ] 5. In-run fix 1 — verify: Weekly costs chart crashes when a visible series has no measured week
  - [ ] **Step 1: Write the failing test** `weekly costs draws a series with no measured week as no line` in `stats/web/src/views/views.test.tsx`: a `weekly-costs` response whose Review is null every week renders the chart, the Review toggle and no Review line or point. Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'weekly costs'` — RED (`TypeError: Cannot read properties of undefined (reading 'series')`).
  - [ ] **Step 2: Fix** in `stats/web/src/views/WeeklyCosts.tsx`: a line with no measured point draws nothing, before its series is read.
  - [ ] **Step 3: Verify:** `cd stats/web && npx tsc -b && npx vitest run src/views/views.test.tsx -t 'weekly costs'`.

**Files:** `stats/web/src/views/WeeklyCosts.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `weekly costs draws a series with no measured week as no line`
**Regression:** `weekly costs draws a series with no measured week as no line` fails, because the round-1 fix read `line.find(...)!.series` on an all-null line and threw.
**Build:** green
**After:** Task 1, 3
**Commit:** `fix(stats): Weekly costs draws a series with no measured week as no line`
