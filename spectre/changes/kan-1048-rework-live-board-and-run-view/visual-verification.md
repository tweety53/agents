# Visual verification — kan-1048-rework-live-board-and-run-view

Frames: no `mockups` root is declared, so each capture is composed against the approved prototype
`.superpowers/sdd/live-board-prototype.html` (design.md `## Context`), rendered by
`visual-verification/prototype/render.cjs` at 1280x1024, dsf 1, dark, and mapped by
`visual-verification/prototype.mockups` (`frame none`: native size, no difference panel).

The prototype file has no `<!DOCTYPE>`, so it renders in quirks mode, where tables do not inherit
the body's `font: 13px` — its table text draws at ~16px. Its CSS declares 13px; the app's 13px
table text follows the declaration, not the quirks-mode pixels.

## Views

### Live state board — `dashboard` / `full-state-board`
- Capture: `visual-verification/evidence/dashboard-darwin.png`, `visual-verification/evidence/full-state-board-darwin.png` (byte-identical content to each other's view, separate specs).
- Composite: `visual-verification/dashboard.png`, `visual-verification/full-state-board.png`; diff=n/a (frame none).
- Seen: chip bar, five tiles, Running now with two cards (awaiting-answer card first, yellow border), state-grouped Changes table, Finished collapsed. Text runs are display values: `$5.44`, `$0.59`, `$2.07`, `5 +1`, `just now`, `20m`, `12m`, `tasks 3/5`, `tasks —`; no NaN/raw float/placeholder.
- Bands: one whole-page band each side (paired, height 739 vs 725; cause: the app's `flow stats` nav header, +98px, and the fixture's 4 rows against the prototype's 7 — data). Seams: 0 paired, 2 missing, 4 extra, all on that one whole-page band; cause: the band covers the whole page, so seams are read per control below.
- Per control (measure-visual-properties, regions offset by the 98px header): state tile 238x69, radius 2, border 1 #2c3235, fill #181b1f, padding 14/15 — every delta 0 (content.height -1). Running card 395x122, border #f2cc0c, padding 13/11/13/14, gaps 12/10/12/12 — every delta 0. Filter bar 1232x43, border/fill/padding — every delta 0.
- Matrix: `visual-verification/matrix/dashboard.json`, `visual-verification/matrix/full-state-board.json` (21 paired elements each).

### Board, Project chip + Started tile — `combination-chips-plus-tile`
- Capture: `visual-verification/evidence/combination-chips-plus-tile-darwin.png`; composite `visual-verification/combination-chips-plus-tile.png`, diff=n/a.
- Seen: uitest-alpha chip pressed, Started tile pressed (white border), Clear filters shown, one Started group with kan-101.
- **Fixed (in-run fix 1) — Cost in period follows the filters.** The tile now sums every board row, as the prototype's does; recaptured `dashboard`, `combination-chips-plus-tile` and `full-state-board` change only that tile's value. Was: The app's tile reads `$2.66` under the uitest-alpha chip (`$5.44` unfiltered); the prototype sums every change regardless of filter (`changes.reduce((a, c) => a + c.cost, 0)` in its `board()`), and reads `$160.76` both unfiltered and under the `agents` chip. App source: `StateBoard.tsx` `base.reduce(...)`. Absent at the merge base — merge-base baseline: 6a53dd04f89c7eefe9712dde4987b445a6f44bef:stats/web/tests/visual/baseline.spec.ts-snapshots/dashboard-darwin.png draws the old one-panel board with no Cost in period tile; this change's.
- Matrix: `visual-verification/matrix/combination-chips-plus-tile.json` (13 paired elements).

### Run view — `run-detail` / `full-run-detail` / `run-detail-hover`
- Captures: `visual-verification/evidence/run-detail-darwin.png`, `visual-verification/evidence/full-run-detail-darwin.png`, `visual-verification/evidence/run-detail-hover-darwin.png`; composites `visual-verification/run-detail.png`, `visual-verification/full-run-detail.png`, `visual-verification/run-detail-hover.png`, diff=n/a.
- Seen: crumb, title + FINISHED pill, five KPIs (`—`, `$2.78`, `25h 05m`, `0.32M`, `0`), stage-lane timeline (two lanes, green bars, no now line — both runs ended), session-grouped Stage runs (`Session 1 · other 07:00–07:05 · 5m · $2.75`, `Session 2 · landing …`), Cost by stage (`SDD + TDD per t…` truncated by the 120px name column, as the prototype's `.hbar .n` ellipsis does). Hover: tooltip `SDD + TDD per task / Session 1 / completed · 5m · $2.75`, bar outlined.
- Bands: 0 paired, 1 missing (frame 0–1002), 1 extra (capture 0–672); cause: one whole-page band each side, heights differ by the prototype's 9 runs/8 lanes and open-stage KPI against the fixture's 2 runs — data. KPI panel 1232x107 vs 1232x91 (-16): cause the prototype's Now running KPI carries an open stage's hint and activity badge; the fixture's run is FINISHED. Border, fill, radius, padding deltas 0.
- Prototype tooltip's second line is the command (`/flow`); the app's is the session (`Session 1`) — the dispatch's motion text names session, so not counted.
- Matrix: `visual-verification/matrix/run-detail.json`, `visual-verification/matrix/full-run-detail.json`, `visual-verification/matrix/run-detail-hover.json` (17 paired elements each).

## Motions
- live-dot-pulse — `visual-verification/motion/live-dot-pulse/`, 182 frames — clean: `lb-pulse 1.6s infinite`, the ring grows and fades over the loop; under `prefers-reduced-motion: reduce` the computed `animation-name` is `none`.
- group-caret — `visual-verification/motion/group-caret/`, 66 frames — clean: transform `rotate(-90deg)` → `none` over `transform 0.15s`, the group's rows appear in the click's frame.
- timeline-tooltip — `visual-verification/motion/timeline-tooltip/`, 5 frames — **fixed (in-run fix 1)** — `onMouseLeave` now hides the tip under `flushSync`, in the same paint; pinned by `run-detail.spec.ts` `the tooltip is gone in the paint the pointer leaves the bar` (5/5 under `--repeat-each 5 --workers 1`). Was: **frame 003 (424ms)**: the pointer has left the bar (its hover outline gone) and the tooltip is still drawn; it is gone at frame 004 (435ms). Cause: `StageTimeline.tsx` hides the tip through React state on `onMouseLeave`, one paint after the CSS `:hover` outline drops; the prototype hides it synchronously in its `mousemove` handler. This change's (the tooltip is new).
