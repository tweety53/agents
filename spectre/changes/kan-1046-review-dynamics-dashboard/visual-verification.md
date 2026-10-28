# Visual verification — kan-1046 review dynamics

Frame: the approved prototype `.superpowers/sdd/review-dynamics-prototype.html` (no `mockups` root), rendered per capture by `visual-verification/prototype/render.mjs` at 1280x1024, dpr 1, the capture's colour scheme; map `visual-verification/prototype.mockups`. Composites carry `diff=n/a`: the capture (2486px, with the app's nav header) and the frame (2459px, no header) differ in height, so no difference panel is drawn. Matrices: `visual-verification/matrix/<frame>.json` (bands/seams in `<frame>.bands.json`). The view is new: no merge-base rendering, so every departure below is this change's unless stated.

## review-dynamics (dark, full page)
- Capture: `visual-verification/evidence/review-dynamics-darwin.png`; composite `visual-verification/review-dynamics.png`.
- Seen: nav header with "Review dynamics" current, disabled Model filter with its reason, h1, intro, Range row (no preset pressed — range pinned by query), Family pressed, six tiles, 10 charts in four sections with markers, heatmap of 3 slots x 5 model·effort columns, footnote. All numbers are display values; no NaN/null/undefined.
- Bands 1/1 paired (whole page one band in both — dark page has no full-width rule the scanner splits on); seams paired at page edges. Matrix: 5 paired, 27 missing, 35 extra elements — layout offset by the app header (extra, expected app chrome) and different seed data unpairs the rest.
- Departures (this change's):
  1. Intro paragraph uses the app's `.view-description` (~14px, primary-ish colour); prototype `.sub{font-size:12px;color:var(--text-secondary)}`.
  2. Heatmap table stretches the panel's full width with a wide row-label column and hairlines between rows; prototype `.heat{border-collapse:separate;border-spacing:3px}` with `td{width:92px}` and no row rules, table left-packed. Global table styles cascade into `.dynamics-heat`.
  3. Heatmap columns and rows sorted alphabetically (`ReviewDynamics.tsx:115-116`): efforts read high, low, medium; prototype orders low, medium, high and slots primary, principles, failure-modes, test-sanity, mutation, call-efficiency.
  4. "Findings per dispatch by reviewer effort" legend/series order high, low, medium (`ReviewDynamics.tsx:304` sorts by name); prototype low, medium, high.
  5. No "Last 8 weeks" pressed state captured: the prototype draws the default preset pressed (`.bar button[aria-pressed="true"]{border-color:var(--s1)}`); the spec pins the range by query so no preset is ever pressed in any capture.
- Not data-caused: seed data differs from prototype sample data (values, tick labels, marker positions) — named cause for every text/value difference.

## review-dynamics-hover (dark, severity panel)
- Capture: `visual-verification/evidence/review-dynamics-hover-darwin.png`; composite `visual-verification/review-dynamics-hover.png`.
- Seen: crosshair on week 05-18, tooltip "Week of 05-18", Critical 1.0, Important 0.5, Minor 4.0, "based on 2 changes", marker line "◆ Prompt edit · primary: primary reviewer re-reads every writer of state the diff reuses (2e053aa)".
- Bands 1/1, seams paired at the panel border. Matrix: 5 paired, 3 missing, 1 extra.
- Departures:
  1. Tooltip background `var(--surface-2)` (#181b1f dark, same as the panel); prototype `.tip{background:var(--surface-tip)}` = #22252b.
  2. The capture is an element clip (`severity.toHaveScreenshot`), not full page/viewport — step 8 forbids a clipped capture.
- Frame artefact: the prototype render's tooltip sits below the panel (its `svg.offsetTop+8` positioning on the frame's data), not a capture defect.

## review-dynamics-light (light, full page)
- Capture: `visual-verification/evidence/review-dynamics-light-darwin.png`; composite `visual-verification/review-dynamics-light.png`.
- Seen: same content as dark in the light palette; all values display values.
- Bands: frame 1 band (missing), capture 24 extra; seams 1 missing, 3 extra. Cause: the app's light palette (`--surface-1:#fcfcfb`, `--border:#d8d6cf`, warm greys) versus the prototype's (`--surface-1:#f4f5f7`, `--border:#d0d7de`) — present at the merge base (theme.css light block, merge-base baseline: c4202f60:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-light-darwin.png). Matrix: 4 paired, 14 missing, 20 extra.
- Departures 1-5 of the dark frame repeat here.

## Missing captures (prototype states no test drives)
- Heatmap cell hover (design.md `## Context` names it; prototype draws it via `title`).
- Version toggle pressed (chart titles/heatmap header switch to "model version").
- Default "Last 8 weeks" pressed preset (see departure 5).

## Sweeps (per frame)
- text: done (every run transcribed; display values). order: done (prototype order h1, sub, ranges, models, tiles, review panel, implementer, visual verify, heatmap, note — capture matches, plus app header extra). reach: done (footnote, last element, reached in full-page capture). derived: done — tile deltas follow range; not re-driven by input (no editable input other than presets/toggle, which no capture drives). rows: done — Range/Models bars gap 12/16px per `.dynamics-bar` = prototype `.bar`. ink: done — heatmap row hairlines extra (departure 2), all panel borders present.

## In-run fix 1 (task 11)
- Dark frame departures 1–5 and hover departures 1–2 are fixed:
  - the intro is `.dynamics-sub` (12px);
  - `.dynamics-heat` is `width: auto`, with 92px cells, no row rules and `padding: 0`;
  - rows run busiest slot first, and columns and effort series run low → high;
  - the tooltip is on `--surface-tip` (#22252b dark);
  - the hover capture is full page;
  - new captures: `review-dynamics-version`, `review-dynamics-default-preset` (Last 8 weeks pressed, clock pinned 2026-06-29) and `review-dynamics-heat-hover`.
- The heatmap cell hover is the native `title` tooltip, which no screenshot draws, so its text is asserted by the spec.
- The light palette's departure predates the change and stands: the app's theme tokens are the design-system source for every view.
- The next verifier run rewrites this record's frame sections.
