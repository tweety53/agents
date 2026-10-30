# Visual verification — kan-1067 (unscoped full run)

Frames: the approved prototype `.superpowers/sdd/experiments-prototype.html`, rendered to `.superpowers/sdd/experiments-frames/`, mapped by `stats/web/tests/visual/experiments.spec.ts.mockups`. No `mockup frame` geometry is declared, so compose wrote side-by-side composites with `diff=n/a` (capture 1280x1873/1864 vs frame 1280x1366/1386 — sizes differ, no difference panel).

## experiments-prototype-open
- capture: visual-verification/evidence/experiments-open-darwin.png; composite: visual-verification/experiments-open.png, diff=n/a.
- Seen: roles table (28 rows, every row of model-policy.md; frame draws an abbreviated 15), experiment list, override table, five stat panels, five per-metric charts, summary.
- Bands: 0/1 paired, 1 missing, 1 extra — the whole page is one band in both images (frame top 0 h1366, capture top 0 h1873); height delta 507px. Cause: roles table row count (data, model-policy.md real rows) plus the two-line app header — header present at the merge base (merge-base baseline: 52c67c31:stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-state-board-darwin.png).
- Seams: 0 paired; same single band — nothing to pair.
- Departures (view created by this change, so each is this change's):
  1. Experiment list order — fixed (Task 15): open experiments first, then newest start first; capture lists "planning on xhigh" (open) above "implementer sonnet/flow-low" (closed), as the frame lists open first.
  2. Stat panel delta wrap — fixed (Task 15): the delta line is `white-space: nowrap`; every delta renders one line, 16.8px tall in both the app and the prototype ("▼ 100% vs 1.9 before").
  3. Tokens / change and Wall-clock — fixed (Task 15) by seed data: explorer dispatches on both sides of the start; the panels read 60 (▲ 23% vs 48 before) and 8.5 (▼ 31% vs 12 before), each chart draws 5 weekly points with both mean lines. Plan corrections still draws one point — the fixture records one plan-correction count; frame draws a series. Data, not view.
  4. Chart x-axis last label clipped ("08-1", "07-2") at the plot's right edge — the frame draws the same clipping (prototype's own defect, faithfully reproduced).
- Matrix: matrix/experiments-prototype-open.json — 16 rows (1 paired, 7 missing, 8 extra) by the elements seed; unpaired rows follow from the 523px vertical shift, not from element absence (every frame element is seen in the capture by eye).
- Sweeps: text done (all values display precision: 0.0, 0.5, 50%, 1.0, "—"); order done (open first, matches frame); reach done (fullPage capture reaches Summary); derived n/a — no operator-editable input; rows done via bands (single band); ink done via bands/seams.

## experiments-prototype-closed
- capture: visual-verification/evidence/experiments-closed-darwin.png; composite: visual-verification/experiments-closed.png, diff=n/a. Same structure as open with the closed experiment selected; list order as open (departure 1 fixed); departure 4 as open; its Tokens/Wall-clock panels read 0.0 and 24 with no before figure — the closed implementer experiment's before window holds no implementer dispatches with tokens (fixture data); band/seam result identical in shape (frame h1386, capture h1864). Matrix: matrix/experiments-prototype-closed.json.

## experiments-prototype-hover / experiments-prototype-focus
- captures: visual-verification/evidence/experiments-hover-darwin.png, visual-verification/evidence/experiments-focus-darwin.png; composites visual-verification/experiments-hover.png, visual-verification/experiments-focus.png.
- The two captures are byte-identical, and so are the two prototype frames: the prototype styles `.pt:hover,.pt:focus{r:5;outline:none}` identically and draws no tooltip, so both states render the same pixels. Matrices: matrix/experiments-prototype-hover.json, matrix/experiments-prototype-focus.json.

Motions: none named.

## In-run fix 2 re-verify (round base 4a48720c)
- Captures re-taken with `--update-snapshots=all` came back byte-identical to the committed baselines; departures 1 (list order) and 2 (delta wrap) are confirmed fixed and 3 (Tokens/Wall-clock seeded) confirmed on the open experiment. Matrices re-saved over the previous ones.
- New departure, this change's (a view this change creates, no merge-base rendering): on the closed experiment, Review rounds / change and Verify reds read `▼ vs 0.0 before` in the green "better" colour while before and after are both 0.0. `StatCell` in `stats/web/src/views/Experiments.tsx` picks `d > 0 ? "▲" : "▼"`, so an unchanged value (d = 0) draws as an improvement. The prototype draws no equal case; expected value read from the data, not the pixels: no arrow claim of improvement when after equals before. The chart's mean-after line has the mirror problem: `after < before` is green, so equal draws red ("worse").

## Departure A — fixed (in-run fix 3)
- `StatCell` now draws an unchanged value (d = 0) as `= vs <before> before` with neither the `up` nor the `down` class, and the chart's mean-after line takes `var(--text-muted)` when after equals before. The prototype draws no equal case; the neutral reading is the recommended one, recorded on the handoff.
- Recaptured `experiments-closed-darwin.png`: Review rounds / change and Verify reds read `= vs 0.0 before` in the neutral text colour; their mean-after lines are muted. Full app suite recaptured with no changed PNG; `full-app-suite.zip` rebuilt. Probe: `views.test.tsx` "Experiments draws an unchanged metric as neutral, with no arrow and no delta colour".
