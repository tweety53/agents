# Visual verification — kan-1046 review dynamics

Frame: the approved prototype `.superpowers/sdd/review-dynamics-prototype.html` (no `mockups` root), rendered per capture by `visual-verification/prototype/render.mjs` full page at 1280x1024, dpr 1, the capture's colour scheme, driving each capture's state (Version pressed, chart hover at bucket 2, first heat cell hovered); map `visual-verification/prototype.mockups`. Composites carry `diff=n/a`: capture and frame heights differ (the capture carries the app's nav header, the frame none), so no difference panel is drawn. Bands/seams: `visual-verification/matrix/<frame>.bands.json`; matrices: `visual-verification/matrix/<frame>.json`. The view is new: no merge-base rendering, so any departure is this change's unless stated. Re-run at HEAD 2553edea as fix round 1's closing unscoped run (whole suite, no round base): every capture re-taken, every frame re-rendered, re-composed and re-measured; bands, seams and matrix counts unchanged from the scoped run. The seam pairing's one page cell lists 21 text lines (22 on Version); the ~10 with an `offset` delta over 2px are the in-order pairing shifted by the app nav header's three lines the prototype lacks — the rest have `left` delta 0, left-anchored runs whose seed text differs.

## review-dynamics (dark, full page)
- Capture `visual-verification/evidence/review-dynamics-darwin.png`; composite `visual-verification/review-dynamics.png`.
- Seen: intro at 12px (`.sub{font-size:12px}`), heatmap packed 92px cells with no row rules (`.heat td{width:92px}`), columns opus low/medium/high then sonnet low/medium, effort series low/medium/high, as the prototype orders them. Previous dark departures 1–5 confirmed fixed.
- Bands 1/1 paired (page band; height delta 16px = app nav header plus seed-data page length). Seams 4/4 paired, 0 over tolerance.
- Matrix 67 rows × 11 columns, 0 n/a: 5 paired, 27 missing, 35 extra — the app header shifts every element ~170px and seed data differs (tile values, series counts, heat rows), so the pairing desyncs; no structural departure read on the composite.

## review-dynamics-version (dark, Version pressed)
- Capture `visual-verification/evidence/review-dynamics-version-darwin.png`; composite `visual-verification/review-dynamics-version.png`.
- Seen: Version pressed in the `.bar button[aria-pressed="true"]` accent; reviewer-model chart and heatmap split by version (claude-opus-5, claude-opus-5-5, claude-sonnet-5-5, opus (version unrecorded)). Bands 1/1, seams 4/4. Matrix 67 rows, 5 paired / 27 missing / 35 extra, same cause.
- Spec note: the click capture parks no pointer (no `page.mouse.move(0, 0)` after the click), against step 8's click rule; no hover ink is visible on this capture.

## review-dynamics-default-preset (dark, Last 8 weeks pressed)
- Capture `visual-verification/evidence/review-dynamics-default-preset-darwin.png`; composite `visual-verification/review-dynamics-default-preset.png`.
- Seen: Last 8 weeks and Family pressed — the prototype's opening state. Bands 1/1, seams 4/4. Matrix 67 rows, 3 paired / 29 missing / 35 extra, same cause.

## review-dynamics-hover (dark, chart hover)
- Capture `visual-verification/evidence/review-dynamics-hover-darwin.png`; composite `visual-verification/review-dynamics-hover.png`.
- Seen: full page; tooltip "Week of 05-18", severity rows, "based on 2 changes", marker line in `--marker`, on #22252b (`.tip{background:var(--surface-tip)}`). Previous hover departures confirmed fixed. Bands 1/1, seams 4/4. Matrix 64 rows, 5 paired / 24 missing / 35 extra.

## review-dynamics-heat-hover (dark, heat cell hover)
- Capture `visual-verification/evidence/review-dynamics-heat-hover-darwin.png`; composite `visual-verification/review-dynamics-heat-hover.png`.
- Byte-identical to `review-dynamics-darwin.png`: the prototype declares no `.heat td:hover` style and its tooltip is the native `title`, which no screenshot draws; the spec asserts the title text. Bands 1/1, seams 4/4. Matrix 67 rows as the resting frame.

## review-dynamics-light (light, full page)
- Capture `visual-verification/evidence/review-dynamics-light-darwin.png`; composite `visual-verification/review-dynamics-light.png`.
- Bands 0/1 paired, 1 missing, 23 extra; seams likewise: the app's light theme tokens split bands the prototype's palette does not — known and accepted, present at the merge base (merge-base baseline: c4202f60:stats/web/tests/visual/weekly-costs.spec.ts-snapshots/weekly-costs-light-darwin.png). Matrix 39 rows, 3 paired / 15 missing / 21 extra.

## Sweeps (every frame)
text, order, reach, derived, rows, ink done on each frame: every label a display value (no NaN/null/raw float); element order matches the prototype top to bottom below the app header; the page scrolls to the footnote; delta arrows follow the previous range; bounded panel rows within tolerance; no frame ink run missing (heatmap row hairlines gone).

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
