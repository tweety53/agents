# Visual verification — kan-1038-stats-weekly-costs-one-chart

`mockups` is not declared for this project, so no frame is composed; band, seam and matrix
readings are n/a — no frame. The approved prototype (`.superpowers/sdd/weekly-costs-preview.html`)
was read for the drawn element list only.

## weekly-costs (dark) — `visual-verification/evidence/weekly-costs-darwin.png`
One "LAST 8 WEEKS" panel with caption "Overall includes stages outside the four phases.", five
pressed legend chips (Overall, Planning, Implementation, Review, Self-review), y axis $0.00–$5.00 in
$1.00 steps titled "Cost (USD)", x axis 2026-06-24 … 2026-08-12 (8 Wednesday ticks) titled "Week
starting (Wed, UTC)". No filter bar. Isolated points at 2026-07-22; Overall line 1.50 → 3.37 with
pill "▲124%" (red, delta-up; (3.37 − 1.50)/1.50 = +124.5%), Review 0.60 → 0.41 with "▼32%" (green,
delta-down; −31.7% → 32%). Pills sit inside the card's right edge. No raw float, NaN or placeholder.

## weekly-costs-light — `visual-verification/evidence/weekly-costs-light-darwin.png`
Same layout and values in the light palette; series and delta colours swap to the light tokens.

## weekly-costs-empty — `visual-verification/evidence/weekly-costs-empty-darwin.png`
Panel heading and caption, then "No data was recorded for this period."; no chips, no chart.

## weekly-costs-toggled — `visual-verification/evidence/weekly-costs-toggled-darwin.png`
Overall chip dimmed (aria-pressed false), Overall points gone, y axis rescaled to $0.00–$1.25 in
$0.25 steps. The spec clicks the chip and parks the pointer (`page.mouse.move(0, 0)`) before the
capture (in-run fix 2); the re-captured PNG is byte-identical.

## weekly-costs-tooltip — `visual-verification/evidence/weekly-costs-tooltip-darwin.png`
Hovering Overall's 2026-08-12 point enlarges it and opens the tooltip: "Overall", "Week of
2026-08-12", "$3.37", "▲ 124.5% vs previous week" in delta-up red, inside the plot (in-run fix 2).

## full-app-suite weekly-costs — `visual-verification/evidence/full-weekly-costs-darwin.png`
Recaptured: the committed baseline (1280×1091, the pre-change two-chart view) was stale and failed
`verify`; the new capture shows the one-chart view on the stack's own data (isolated points only).

## Sweeps (capture scope)
- text: done — every label and value above transcribed; display precision 2dp currency, integer %.
- order: n/a — no frame; containment: pills and points stay inside the card.
- reach: done — page fits the 1024px viewport, nothing below the card.
- derived: done — deltas re-derived by hand (above); toggle re-derives the y scale.
- rows, ink: n/a — no frame.
- The hover tooltip is driven by `hovering a point shows its tooltip` (in-run fix 2).

## Motions
None named.
