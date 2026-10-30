# Live verification — kan-1060

Run 2026-10-30 against a clone of the change branch at `f6d7e935`, on a copy of the dev store
(`scripts/live-store-copy.sh up kan-1060`, port 4175) and the UI-test stack on 4174 — the dev
`flowd` and `flow` database were never touched. Every step of design.md's `## Live check` was run.

| Measure | Before | After |
|---------|--------|-------|
| `flow.visual-verify` Main share (2026-09-30 → 2026-10-30) | 34% (decision `done-when-baseline`) | 33.4% |
| `flow.visual-verify` Main cost | $332 (decision `done-when-baseline`) | $311.90 of $934.13 total, 211 runs |

The Main share lies inside 25–45% and Main is not equal to total, so the columns read the split the
plan defines.

## Steps

1. Seeded the change's state in the store copy; green verify row `verify-livecheck-2`. Deviation:
   the design's token `mf-k1060-livecheck` was refused (tokens must end in a UTC stamp); the run used
   `mf-k1060-livecheck-20261030T193618Z`. With the unstamped token `prepare` printed
   `NEXT: restart-stack` and a stage-begin warning.
2. `visual-verify-pass.sh prepare` printed `DISPATCH: visual-verify-livecheck …` and
   `NEXT: dispatch 1`; the row ran opus/low with diffBase HEAD, and the bundle carried the resolved
   `capture` row and `verify: verify-livecheck-2@f6d7e93`.
3. `measure --prototype` wrote the manifest: two pairs, every exit 0, four over rows per frame
   (#3 unresolved, #4 ink.width, #11 unresolved, #12 colour+unresolved).
4. The first `close` refused an incomplete report (`VERIFY-REPORT-INCOMPLETE … trend-darwin.png is on
   no "- <view>: <path>" line`) and chained `NEXT: rest visual-verify-livecheck-rest`. With the
   report completed, `close` printed `STAGE: completed`, `NEXT: done` and committed in the clone
   (`ROUND-BASE … 41a6c238`). Dropping row #12 from one frame made `check-verify-report.sh` exit 1
   with `stage-leaderboard-prototype-sorted over-tolerance: row #12 unnamed`. Run with `--design`
   on gymie's kan-938 design, it exited 1 naming lines 100, 169 and 228 (and 471, 632–635).
5. `GET /api/v1/stats/stage-leaderboard` returned the figures in the table above.
6. `make -C stats ui-test-down`, `scripts/live-store-copy.sh down kan-1060`, and the clone removed.
