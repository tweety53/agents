# Visual verification — kan-1037-stats-weekly-costs-dashboard

No `mockups` declared: no frame composed; bands, seams and matrix are n/a — no frame. Captures from `stats/web/tests/visual/weekly-costs.spec.ts` (clock pinned to 2026-08-16T12:00Z, viewport 1280x1024, UI-test fixture).

## weekly-costs (dark) — `visual-verification/evidence/weekly-costs-darwin.png`

Nav shows "Weekly costs" active between "Trend over time" and "Reviewers". Heading "Weekly costs", description "Wednesday-to-Wednesday cost by pipeline phase." Panel "LAST 8 WEEKS" with caption "Overall includes stages outside the four phases.", a search field and one row: Week 2026-08-12, Overall $3.37, Planning unavailable, Implementation unavailable, Review $0.41, Self-review unavailable. Panel "WEEKLY COST" draws two points (Overall at the top of the plot, Review near its floor), both on the plot's left edge, and a legend Overall/Planning/Implementation/Review/Self-review in series-1..5 (blue, green, yellow, orange, purple). Every text run is a display value; no raw float, NaN or null.

Findings:
- The 8-week table draws one row, not eight: the store returns only weeks with a costed run, so the six weeks of 2026-06-24..2026-08-05 with no run are omitted, where design.md's live check expects "8 Wednesday-keyed rows". This view is new, so the departure is this change's.
- A one-week chart draws no line: with a single x value both points sit at x = PADDING on the y axis and the plot's width is empty. TimeSeriesPanel's x scale does the same for any single-day series (unchanged at the merge base); the weekly view reaches it whenever the period spans one week.
- The fixture holds no planning, sdd-tdd or self-review stage, so three of five lines cannot be drawn by any capture over it.

## weekly-costs (light) — `visual-verification/evidence/weekly-costs-light-darwin.png`

Same layout on the light surface; legend in the light series tokens (#2a78d6, #37872d, #b58a00, #e0752d, #8f3bb8), all legible on the light background.

## weekly-costs (empty) — `visual-verification/evidence/weekly-costs-empty-darwin.png`

Clock 2020-01-01, period 2020-01-01..02: both panels read "No data was recorded for this period."; no chart, no legend.

## Derived

Changing the period to EMPTY_QUERY at the pinned clock empties the chart while the table still shows $3.37 — the table ignores the picker, the chart follows it (spec test 4).

## Trend over time (shared TimeSeriesPanel)

`views.spec.ts`'s trend.png differed from its baseline only in the nav row (the new link); the single-series chart is pixel-identical, so the single-series path is unchanged. Every other stale baseline differed in the nav row alone and was re-captured.

## Motion

None named.
