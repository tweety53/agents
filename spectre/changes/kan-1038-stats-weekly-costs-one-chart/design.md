## Context

- Spec: the approved prototype, `.superpowers/sdd/weekly-costs-preview.html` in this worktree (published https://claude.ai/artifact/QQtmw9r3qLqpMtnrCKHGbP, v5). It is the mockup for layout and behaviour; colours come from the app's `theme.css` tokens, not the prototype's.
- Current view: `stats/web/src/views/WeeklyCosts.tsx` renders a `DataTable` over `lastEightWeeks` and a `TimeSeriesPanel` over the picker period. `TimeSeriesPanel` is shared with Trend and run detail.
- The filter bar is `DashboardBar`, rendered once by `App.tsx` above every view.
- The API (`weekly-costs`, `WeeklyCostPoint`) is unchanged.
- The prototype request came mid-planning (operator, 2026-10-27: "update flow to always prototype the things visually … as a part of this task"), so the pipeline change rides this change.

## Decisions

### Own chart component, not TimeSeriesPanel

**ID:** weekly-chart-own-component
**Status:** active
**Chosen:** a `WeeklyCostChart` inside `WeeklyCosts.tsx` (inline SVG, the app's established chart pattern) — toggles, delta labels, label stacking and tooltip are specific to this view.
**Considered:** extending `TimeSeriesPanel` — it would grow five options only one caller uses and risk Trend/run-detail rendering; adding a chart library — no second caller, and inline SVG already covers it.

### Fixed last 8 weeks, no filters

**ID:** fixed-eight-weeks-no-filters
**Status:** active
**Chosen:** one fetch of `lastEightWeeks(now)` with no project and no model; `App.tsx` does not render `DashboardBar` on the `weekly-costs` route.
**Considered:** keeping the period picker for the chart — the operator asked the filter block removed completely; keeping the project filter alone — same reason.

### Delta semantics

**ID:** week-over-week-delta
**Status:** active
**Chosen:** a point's label is `(this − previous) / previous`, previous = the same series' immediately preceding week; no label on the first week, after an Unavailable week, or when previous is 0. Rounded to a whole %, `▲` red (`--delta-up`) for > 0, `▼` green (`--delta-down`) for < 0, `0%` neutral for exactly 0. The tooltip shows one decimal.
**Considered:** comparing with the last measured week across a gap — it mislabels a two-week change as weekly.

### Label placement

**ID:** delta-label-stacking
**Status:** active
**Chosen:** labels sit right of their dot as outlined pills; per week column they are sorted by y and pushed apart to ≥17px, then shifted up together if the stack overruns the plot; a moved label gets a leader line to its dot (prototype v5).
**Considered:** labels above the dot (prototype v4) — overlapped when series were close, rejected by the operator.

### Planning prototypes for visible changes

**ID:** plan-visual-prototype
**Status:** active
**Chosen:** a checklist item in `skills/flow/brainstorm-planner.md` section B: when the change can alter what a user sees and no mockup frame already specifies it, the planner builds a clickable prototype with sample data, publishes it as an Artifact, iterates it on the operator's feedback, and on approval saves it under `.superpowers/sdd/` and cites it from `design.md` as the spec. Flow-only, so not in the vendored `skills/spectre-planning/SKILL.md`.
**Considered:** putting it in `spectre-planning` — that file is vendored generic and the Artifact tool is a harness feature; asking before each prototype — the operator asked for "always … when applicable".

## Open questions

## Recommended approach

Rewrite `WeeklyCosts.tsx` around one fetch and a view-local SVG chart (`weekly-chart-own-component`); hide the bar on this route in `App.tsx` (`fixed-eight-weeks-no-filters`); re-baseline the visual spec; add the planning-prototype checklist item (`plan-visual-prototype`).

## Critical files

- `stats/web/src/views/WeeklyCosts.tsx` — the dashboard; rewritten.
- `stats/web/src/views/views.test.tsx` — weekly costs unit tests; replaced.
- `stats/web/src/theme.css` — adds `--delta-up` / `--delta-down` tokens for both palettes.
- `stats/web/src/styles.css` — chart, pill and tooltip styles.
- `stats/web/src/App.tsx`, `stats/web/src/App.test.tsx` — no `DashboardBar` on `weekly-costs`.
- `stats/web/tests/visual/weekly-costs.spec.ts` and its snapshots — re-baselined.
- `skills/flow/brainstorm-planner.md` — the prototype checklist item.

## Existing code to reuse

- `lastEightWeeks`, `everyWeek`, `categories` in `WeeklyCosts.tsx`.
- `useStatsView`, `Panel`, `ViewFrame`, `formatUsd`.
- `--series-1..5` tokens in `theme.css`; `segments()` gap logic pattern from `TimeSeriesPanel.tsx`.

## Verification

- `cd stats/web && npx tsc -b && npx vitest run`.
- `flow.visual-verify` against the UI-test stack: Weekly costs shows one chart, labelled axes, 5 pill toggles, red/green delta pills, no filter bar; the capture compared with the prototype.

## Live check

None — the change is SPA rendering and a skill text; no service, store or runtime state changes, and the API is untouched.
