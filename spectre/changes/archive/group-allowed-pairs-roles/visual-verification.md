# Visual verification — group-allowed-pairs-roles

Frame: the approved prototype `.superpowers/sdd/roles-prototype.html`, rendered by
`visual-verification/prototype/render.mjs` at the capture's 1280×1024 viewport, dark, DPR 1, as
an element screenshot of the `Current roles` region — resting, `plan-designer` row hovered, its
`#<id>` link focused (`visual-verification/prototype/experiments-roles{,-hover,-focus}.png`).

## experiments-roles / -hover / -focus

- Captures: `visual-verification/evidence/experiments-roles-darwin.png`,
  `visual-verification/evidence/experiments-roles-hover-darwin.png`,
  `visual-verification/evidence/experiments-roles-focus-darwin.png` (1232×1503 each).
- Seen: title, header row Role/Does/Allowed pairs, seven uppercase muted group rows (Session,
  Planning, Implementation, Review, Visual verification, Self-review, Pipeline fix) each with a
  bottom hairline, 26 role rows in the prototype's order and text; struck-through current pairs
  followed by orange experiment pairs and a blue `#2` link on explorer/plan-designer/planner;
  muted second line `<pair> with <marker>` on implementer, panel-fix, panel-reviewer,
  self-review-fix, pipeline-fix-fix. Hover: plan-designer row filled (surface-tip). Focus: `#2`
  underlined, no outline — as the prototype's `a:focus` rule.
- Composite: `compose-mockup-frames.sh` printed `diff=n/a` for all three pairs and wrote no
  `.frame.png`: the frame is 1100×2559, the capture 1232×1503. Cause: the prototype's
  `.panel { max-width: 1100px }` caps the panel at 1100px in a 1280px viewport while the app's
  panel spans 1232px; at 1100px the prototype's `td.pairs { white-space: nowrap }` planner row
  overflows the panel's right edge and the Does column collapses to ~60px, wrapping every row to
  3–9 lines. The prototype cannot be rendered at the capture's size without editing it, so no
  band, seam or matrix reading is possible.
- Expected departures (design.md): `#2` vs prototype `#4` (decision `seed-matches-experiment-4`);
  panel box padding/radius (`roles-panel-chrome-unchanged`).
- Spec departure: the three roles captures are element (clipped) screenshots of the region, not
  full page, and the hover capture asserts no hover state (no positive control on the row fill).
- Identical: `experiments-hover-darwin.png` and `experiments-focus-darwin.png` (chart point) are
  byte-identical, and equally so at the merge base (merge-base baseline:
  35f477f8d6f697deb4913c2409acddef40979ab7:stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-hover-darwin.png)
  — not this change's views.

No motions named.

## Step 10 (rerun at matched width)

Prototype re-rendered at 1280x1024, DPR 1, dark scheme, with `.panel{max-width:none}` injected so the `Current roles` region is 1232px wide, the capture's width. Composed under `frame none` (native size, so `diff=n/a` by design). Per frame: bands 1 paired / 0 missing / 0 extra; both outer seams paired with zero width and colour delta; matrix 82 paired / 9 missing / 8 extra. The residual height (1524 vs 1503) and every unpaired matrix element trace to the panel box padding (decision `roles-panel-chrome-unchanged`): every line's left inset is 4px smaller, the Does column is wider, implementer's description wraps one line shorter, and every later row and group rule sits 17px higher, which desyncs the in-order pairing. The `#<id>` text differs per `seed-matches-experiment-4`. No unexplained departure.

## Re-verify on the rebased build (c47730cb)

- Captures `../../../stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-roles{,-hover,-focus}-darwin.png` recaptured with `--update-snapshots=all`: byte-identical to the committed baselines. Prototype frames re-rendered (1280x1024, DPR 1, dark, `.panel{max-width:none}`): byte-identical to `visual-verification/prototype/*.png`.
- Composites `visual-verification/experiments-roles{,-hover,-focus}.png`, diff=n/a (frame none native-size mode). Bands 1/1 paired, height -21 (panel padding, `roles-panel-chrome-unchanged`, and the one-line-shorter wrap). Seams 2/2 paired, zero delta; lines 82/83 (79/80 hover) paired, one unpaired line from the same wrap.
- Matrix per frame: `visual-verification/matrix/experiments-roles{,-hover,-focus}.json` — 99 elements, 82 paired, 9 missing, 8 extra; every missing/extra is the group-rule/text shift of the single wrap; `#<id>` text per `seed-matches-experiment-4`. No unexplained departure.
- Rest of the Experiments view (`experiments-open`, `-closed`, `-hover`, `-focus`): renders without new defects. The weekly-trend charts clip their last x-axis tick label at the plot's right edge (`08-1`, `07-2`); present at the merge base (merge-base baseline: f1163fad3e0fbcab24b07fd1dacccd64bdc3dea4:stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-open-darwin.png), not this change's.
- `experiments-focus` and `experiments-hover` are byte-identical: chart point hover and focus draw the same r=5 point, identical at the merge base too.

## Round: weekly-trend end labels (visual-verify-3)

Round base `c47730cb`; fix `20699081` anchors each weekly-trend chart's first x-axis label `start` and last `end` (`stats/web/src/views/Experiments.tsx`). `verify` exit 0 (75 passed); `capture` of `experiments.spec.ts` and the full app suite's experiments test with `--update-snapshots=all` left every PNG byte-identical to the committed baselines.

- experiments-open — `../../../stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-open-darwin.png`: every end label now renders whole — Review rounds `03-09`…`08-10`, Round-0 Crit+Imp `03-09`…`08-10`, Plan corrections single point `08-10` centred (n=1 keeps `middle`), Tokens `03-09`…`07-20`, Wall-clock `03-09`…`07-20`. Pixel diff against `c47730cb` is confined to x 388–948, y 2221–2394: the two axis-label rows; nothing else on the view moved. Same for experiments-hover and experiments-focus (still identical to each other: point hover and focus draw the same r=5 point, identical at the merge base too).
- experiments-closed: diff confined to y 2128–2135, the axis-label row; end labels whole.
- experiments-roles, -roles-hover, -roles-focus: byte-identical to `c47730cb`, so the visual-verify-2 prototype comparison (frame none, native size, `diff=n/a`), bands, seams, matrices and sweeps stand unchanged; expected departures `seed-matches-experiment-4` and `roles-panel-chrome-unchanged` only.
