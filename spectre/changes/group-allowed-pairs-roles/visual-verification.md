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
