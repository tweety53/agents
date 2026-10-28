# self-review-kan-1046-review-dynamics-dashboard

## Why

The self-review pass of `kan-1046-review-dynamics-dashboard` found nine fixable defects:

- 1a: `flow record dispatch end` on a closed row overwrites its `ended_at` with now; rows 8415 and
  8419 read the moment of a re-point, not the end of the work.
- 1b: the inline back-fill floor ignores the `visual-verify*` verifier rows, so row 8415 starts
  before the verdict its task fixes.
- 1c: `panelFixFloor` has no floor while the round's reviewer rows are still open, so row 8419
  starts before round 3's reviewers began.
- 1d: `flow record mutation -round <r>` accepts rows after round `<r>`'s re-review began; three
  round-1 rows claim mutations nobody ran.
- 1e: the planning session left `explore-1` open and the run closed it with a guessed end; row 8371
  reads 22m33s against a real 1m36s.
- 1f: an appended in-run fix task skipped the plan guards — `close-task` never runs
  `check-plan-shape` — and the visual record it had to commit was still untracked.
- 4a: the planner's live-check paragraph never points at a project-declared stand-in tool
  (`scripts/live-store-copy.sh`).
- 8c: the planner's measure-before-commit rule for capture tasks covers only a declared mockups
  sidecar, not an approved prototype.
- 8d: the closing full verifier's `--update-snapshots=all` rewrote four unrelated baselines with
  run-to-run rasterisation noise at native select corners.

## What changes

- A re-end of a closed dispatch row keeps its `ended_at`; rows 8415 and 8419 get their real ends.
- An inline back-fill floors at the change's `visual-verify*` verifier ends; 8415 starts at its
  verdict.
- An inline panel-fix back-fill is refused while a reviewer row of the round it fixes is open;
  8419 starts at round 3's last report.
- `flow record mutation -round <r>` is refused once a reviewer row of panel round `<r>` has begun;
  a new `flowd drop-mutation` removes the three invented round-1 rows.
- `flow stage end` of `/flow`'s `flow.brainstorm` is refused while a dispatch row of the change is
  open; 8371 ends at the agent's hand-back.
- `close-task` runs `check-plan-shape` when it closes an appended or `In-run fix` task, and
  `verify-fix-loop.md` says the visual record is untracked until that task's commit.
- The planner's live-check paragraph names a project-declared stand-in tool by name.
- The planner's capture-task measure rule covers an approved prototype rendered as the frame.
- The stats SPA's Playwright config pins Chromium's rasterisation, and every baseline is recaptured
  once under the pin.
