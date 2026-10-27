# Visual verification — kan-1037-stats-weekly-costs-dashboard

Unscoped full run closing in-run fix 1. No `mockups` declared: no frame composed; bands, seams and matrix are n/a — no frame. Captures from `stats/web/tests/visual/weekly-costs.spec.ts` (clock pinned to 2026-08-16T12:00Z, viewport 1280x1024, UI-test fixture). Earlier rounds' findings (one-row table, lineless one-week chart, 2 of 5 lines seedable) were fixed in in-run fix 1 and are superseded by this entry.

## weekly-costs (dark) — `visual-verification/evidence/weekly-costs-darwin.png`

Nav shows "Weekly costs" active between "Trend over time" and "Reviewers". Heading "Weekly costs", description "Wednesday-to-Wednesday cost by pipeline phase." Panel "LAST 8 WEEKS", caption "Overall includes stages outside the four phases.", search field, header Week/Overall/Planning/Implementation/Review/Self-review and eight rows 2026-06-24, 07-01, 07-08, 07-15, 07-22, 07-29, 08-05, 08-12. 2026-07-22: $2.07 / $0.62 / $1.10 / unavailable / $0.35 (0.62+1.10+0.35 = 2.07, re-derived). 2026-08-12: $3.37 / unavailable / unavailable / $0.41 / unavailable. Every other cell reads `unavailable` in the muted italic. Panel "WEEKLY COST" draws the pinned period's one week as two points at x≈281, the centre of the 73..488 plot (fix 1's lone-point centring), Overall at the top, Review at the floor; no point crosses the plot or the panel. Legend Overall/Planning/Implementation/Review/Self-review in series-1..5 (blue, green, yellow, orange, purple). Every text run is a display value; no raw float, NaN or null. Page height 1091 > 1024 viewport: the full-page capture reaches the last element, the legend and the chart panel's bottom edge.

## weekly-costs (light) — `visual-verification/evidence/weekly-costs-light-darwin.png`

Same layout and values on the light surface; legend in the light series tokens, all legible.

## weekly-costs (empty) — `visual-verification/evidence/weekly-costs-empty-darwin.png`

Clock 2020-01-01, period 2020-01-01..02: both panels read "No data was recorded for this period."

Finding — departure, this change's own (the view is new, nothing at the merge base): the "Last 8 weeks" table draws its empty state instead of eight `unavailable` weeks. The operator's answer recorded in tasks.md (2026-10-27) is "always 8 rows", and the capture spec's own header states the empty path shows "the table eight Unavailable weeks"; `WeeklyCosts.tsx` builds the eight rows with `everyWeek` only inside `Panel`'s children, and `Panel` renders its empty state first when the response carries no rows, so `everyWeek` never runs. The spec asserts only that the panel is visible, so the baseline carries the defect.

Resolution (814a6333): not a defect. The empty path is a window with no recorded run, and `Panel` shows its not-recorded banner there in both panels by design — the not-recorded / unmeasured / empty distinction every dashboard keeps. "Always 8 rows" governs a recorded window, which `everyWeek` fills. The spec header claiming eight Unavailable weeks was wrong; it now says both panels show the banner, and the test asserts both banners.

## Derived

Period changed to EMPTY_QUERY at the pinned clock: the chart empties while the table still shows $3.37 (spec test 4) — table ignores the picker, chart follows it.

## Trend over time (shared TimeSeriesPanel) — committed baseline `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-trend-darwin.png`

Single-series line between 2026-08-14 and 08-15, unchanged. Re-capture with `=all` differed from the committed baseline in 4 pixels at (612..614, 139..141), the first point's anti-aliasing, under the comparison threshold; the committed baseline was kept.

## Full app suite zip

The committed `full-app-suite.zip` did not match the committed PNGs (weekly-costs, trend, state-board, stage-leaderboard, run-detail entries stale); rebuilt from the committed PNGs.

## Motion

None named.
