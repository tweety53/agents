# Visual verification — kan-1038-stats-weekly-costs-one-chart

`mockups` is not declared for this project, so no frame is composed; band, seam and matrix
readings are n/a — no frame. The approved prototype (`.superpowers/sdd/weekly-costs-preview.html`)
was read for the drawn element list and the tooltip's declared CSS only. Every evidence PNG below
is a copy of the baseline committed at HEAD.

## weekly-costs (dark) — `visual-verification/evidence/weekly-costs-darwin.png`
One "LAST 8 WEEKS" panel with caption "Overall includes stages outside the four phases.", five
legend chips (Overall, Planning, Implementation, Review, Self-review). The view opens with Overall
alone pressed (the spec asserts `aria-pressed` true on Overall, false on the four phases), then
clicks the four phases on and parks the pointer before the capture, so all five chips show pressed.
Y axis $0.00–$5.00 in $1.00 steps titled "Cost (USD)", x axis 2026-06-24 … 2026-08-12 (8 Wednesday
ticks) titled "Week starting (Wed, UTC)". No filter bar. Isolated points at 2026-07-22; Overall line
1.50 → 3.37 with pill "▲124%" (red, delta-up; (3.37 − 1.50)/1.50 = +124.5%), Review 0.60 → 0.41
with "▼32%" (green, delta-down; −31.7% → 32%). Pills sit inside the card's right edge. No raw
float, NaN or placeholder.

## weekly-costs-light — `visual-verification/evidence/weekly-costs-light-darwin.png`
Same state (phases clicked on) and values in the light palette; series and delta colours swap to
the light tokens.

## weekly-costs-empty — `visual-verification/evidence/weekly-costs-empty-darwin.png`
Panel heading and caption, then "No data was recorded for this period."; no chips, no chart, no
filter bar.

## weekly-costs-toggled — `visual-verification/evidence/weekly-costs-toggled-darwin.png`
Phases clicked on, then Overall clicked off: Overall chip dimmed (aria-pressed false), Overall
points gone, y axis rescaled to $0.00–$1.25 in $0.25 steps. The pointer is parked
(`page.mouse.move(0, 0)`) before the capture.

## weekly-costs-tooltip — `visual-verification/evidence/weekly-costs-tooltip-darwin.png`
The default state: Overall chip pressed, the four phase chips dimmed and their series hidden, the
y axis $0.00–$5.00. Hovering Overall's 2026-08-12 point enlarges it and opens the tooltip:
"Overall", "Week of 2026-08-12", "$3.37", "▲ 124.5% vs prev week" in delta-up red, inside the plot,
beside the "▲124%" pill. `.weekly-tooltip` now declares `border-radius: 6px`
(`stats/web/src/styles.css:452`), matching the prototype's `#tip` `border-radius:6px`
(`.superpowers/sdd/weekly-costs-preview.html` line 24); with the earlier `width: 150px`, copy and
`box-shadow` fixes, no tooltip departure remains open.

## full-app-suite weekly-costs — `visual-verification/evidence/full-weekly-costs-darwin.png`
The default state on the stack's own data: Overall alone pressed, two isolated Overall points
(2026-07-22, 2026-08-12), y axis $0.00–$5.00, both axis titles, no filter bar. The committed
`full-app-suite.zip` still held the pre-fix copy of this screen and was rebuilt from the suite.

## Sweeps (capture scope)
- text: done — every label and value above transcribed; display precision 2dp currency, integer %
  on pills, 1dp on the tooltip.
- order: n/a — no frame; containment: pills, tooltip and points stay inside the card.
- reach: done — page fits the 1024px viewport, nothing below the card.
- derived: done — deltas re-derived by hand (above); toggling re-derives the y scale.
- rows, ink: n/a — no frame.

## Motions
None named.
