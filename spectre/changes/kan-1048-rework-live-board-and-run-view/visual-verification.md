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
- Seen: chip bar, five tiles, Running now with two cards (awaiting-answer card first, yellow border), state-grouped Changes table, Finished collapsed. Text runs are display values: `$6.94`, `$0.59`, `$2.07`, `5 +1`, `just now`, `20m`, `12m`, `tasks 3/5`, `tasks —`; no NaN/raw float/placeholder.
- Bands: one whole-page band each side (paired, height 739 vs 725; cause: the app's `flow stats` nav header, +98px, and the fixture's 4 rows against the prototype's 7 — data). Seams: 0 paired, 2 missing, 4 extra, all on that one whole-page band; cause: the band covers the whole page, so seams are read per control below.
- Per control (measure-visual-properties, regions offset by the 98px header): state tile 238x69, radius 2, border 1 #2c3235, fill #181b1f, padding 14/15 — every delta 0 (content.height -1). Running card 395x122, border #f2cc0c, padding 13/11/13/14, gaps 12/10/12/12 — every delta 0. Filter bar 1232x43, border/fill/padding — every delta 0.
- Matrix: `visual-verification/matrix/dashboard.json`, `visual-verification/matrix/full-state-board.json` (21 paired elements each).

### Board, Project chip + Started tile — `combination-chips-plus-tile`
- Capture: `visual-verification/evidence/combination-chips-plus-tile-darwin.png`; composite `visual-verification/combination-chips-plus-tile.png`, diff=n/a.
- Seen: uitest-alpha chip pressed, Started tile pressed (white border), Clear filters shown, one Started group with kan-101.
- Cost in period reads `$6.94` under the uitest-alpha chip + Started tile, the same as unfiltered (`dashboard`): the tile sums every board row (`StateBoard.tsx` `rows.reduce(...)`), as the prototype's `board()` sums every change (`changes.reduce((a, c) => a + c.cost, 0)`). Round 1's departure (filtered `$2.66`) is fixed; clean.
- Matrix: `visual-verification/matrix/combination-chips-plus-tile.json` (13 paired elements).

### Run view — `run-detail` / `full-run-detail` / `run-detail-hover`
- Captures: `visual-verification/evidence/run-detail-darwin.png`, `visual-verification/evidence/full-run-detail-darwin.png`, `visual-verification/evidence/run-detail-hover-darwin.png`; composites `visual-verification/run-detail.png`, `visual-verification/full-run-detail.png`, `visual-verification/run-detail-hover.png`, diff=n/a.
- Seen: crumb, title + FINISHED pill, five KPIs (`—`, `$2.78`, `25h 05m`, `0.32M`, `0`), stage-lane timeline (two lanes, green bars, no now line — both runs ended), session-grouped Stage runs (`Session 1 · other 07:00–07:05 · 5m · $2.75`, `Session 2 · landing …`), Cost by stage (`SDD + TDD per t…` truncated by the 120px name column, as the prototype's `.hbar .n` ellipsis does). Hover: tooltip `SDD + TDD per task / Session 1 / completed · 5m · $2.75`, bar outlined.
- Bands: 0 paired, 1 missing (frame 0–1002), 1 extra (capture 0–672); cause: one whole-page band each side, heights differ by the prototype's 9 runs/8 lanes and open-stage KPI against the fixture's 2 runs — data. KPI panel 1232x107 vs 1232x91 (-16): cause the prototype's Now running KPI carries an open stage's hint and activity badge; the fixture's run is FINISHED. Border, fill, radius, padding deltas 0.
- Prototype tooltip's second line is the command (`/flow`); the app's is the session (`Session 1`) — the dispatch's motion text names session, so not counted.
- Matrix: `visual-verification/matrix/run-detail.json`, `visual-verification/matrix/full-run-detail.json`, `visual-verification/matrix/run-detail-hover.json` (17 paired elements each).

## Motions
Recorded with a CDP screencast at 1280x1024, dark, `reducedMotion: no-preference`, `Date` pinned to `FIXTURE_NOW` (2026-08-15T09:00:00Z) as the specs pin it; each strip replaces round 1's.
- live-dot-pulse — `visual-verification/motion/live-dot-pulse/`, 317 frames — clean: computed `animation: lb-pulse 1.6s`; under `prefers-reduced-motion: reduce` the computed `animation-name` is `none`.
- group-caret — `visual-verification/motion/group-caret/`, 111 frames — clean: In progress collapses at the click's frame (013–016), the caret turns over frames 016–032 (47–207ms, `transform .15s`), steady to 067; the expand click turns it back over 068–079 and the rows return; no empty or intermediate table rendering.
- timeline-tooltip — `visual-verification/motion/timeline-tooltip/`, 6 frames — clean: tooltip appears at 003–004 (18–28ms) beside the outlined bar (`SDD + TDD per task / Session 1 / completed · 5m · $2.75`); at frame 005 (422ms), the first frame after the pointer leaves, outline and tooltip are both gone — round 1's one-paint-late tooltip (its frame 003) is fixed by `flushSync` in `StageTimeline.tsx`'s `onMouseLeave`.
