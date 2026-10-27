# Visual verification — kan-1037-stats-weekly-costs-dashboard

Unscoped full run after fix 2 (814a6333). No `mockups` declared: no frame composed; bands, seams and matrix are n/a — no frame. Captures from `stats/web/tests/visual/weekly-costs.spec.ts` (clock pinned to 2026-08-16T12:00Z, viewport 1280x1024, UI-test fixture). `verify` passed 57/57; the `=all` re-capture of the weekly-costs spec and the suite's weekly-costs screen rewrote every PNG byte-identically, and the rebuilt zip committed nothing. No departures.

## weekly-costs (dark) — `visual-verification/evidence/weekly-costs-darwin.png`

Nav shows "Weekly costs" active between "Trend over time" and "Reviewers". Heading "Weekly costs", description "Wednesday-to-Wednesday cost by pipeline phase." Panel "LAST 8 WEEKS", caption "Overall includes stages outside the four phases.", search field, header Week/Overall/Planning/Implementation/Review/Self-review and eight rows 2026-06-24 through 2026-08-12. 2026-07-22: $2.07 / $0.62 / $1.10 / unavailable / $0.35 (0.62+1.10+0.35 = 2.07, re-derived). 2026-08-12: $3.37 / unavailable / unavailable / $0.41 / unavailable. Every other cell reads `unavailable` in muted italic. Panel "WEEKLY COST" draws the pinned period's one week as two points at the plot's centre x, Overall at the top, Review at the floor; no ink crosses the plot or panel. Legend Overall/Planning/Implementation/Review/Self-review in series-1..5 (blue, green, yellow, orange, purple). Every text run is a display value; no raw float, NaN or null. The full-page capture (1091px tall) reaches the legend and the chart panel's bottom edge.

## weekly-costs (light) — `visual-verification/evidence/weekly-costs-light-darwin.png`

Same layout and values on the light surface; legend in the light series tokens, all legible.

## weekly-costs (empty) — `visual-verification/evidence/weekly-costs-empty-darwin.png`

Clock 2020-01-01, period 2020-01-01..02: both panels read "No data was recorded for this period." — Panel's designed not-recorded banner for a window with no recorded run; the spec asserts both banners (814a6333). The previous round's finding on this capture is resolved as designed behaviour.

## Derived

Period changed to EMPTY_QUERY at the pinned clock: the chart empties while the table still shows $3.37 (spec test 4) — the table ignores the picker, the chart follows it.

## Trend over time (shared TimeSeriesPanel) — committed baseline `stats/web/tests/visual/views.spec.ts-snapshots/trend-darwin.png`

Single-series line between 2026-08-14 and 08-15, same plot geometry as the weekly chart; its spec passed against the committed baseline.

## Full app suite zip

`stats/web/tests/visual/full-app-suite.zip`, 9 PNGs, rebuilt byte-identical.

## Motion

None named.
