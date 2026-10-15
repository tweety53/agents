# self-review-kan-982-prod-fixes-group-add-person-tap-set-value-bleed

## Why

The self-review pass over kan-982 found eight non-`big` defects in the pipeline; six hold against
the tree and two do not (`design.md`, **Context**):

- Integrate's `check-visual-verify-dispatched.sh` asks the visual trigger of each worktree alone, so
  a cross-repo change whose frontend worktree declares no `## visual verification` and whose
  canonical diff touches no `ui paths` passed integrate with every visual-verify row `blocked`.
- `close-task.sh` only warns on a ticked appended task no dispatch row names; kan-982 ticked Tasks
  19–21 past that warning and the gap surfaced only at integrate.
- Load flakes surface one per full-suite round, each round costing an implementer, a review, a
  re-review and a full verify.
- The verifier reported "Back draws in-between frames" from a sampling gap: motion frames carry no
  timestamps, and nothing asks for the sibling trigger at the same sampling.
- A determinism "fix" whose flake never reproduced (0 failures in 8 runs without it) was committed.
  <!-- measured: read from the kan-982 self-review findings file (machine-local, not re-runnable) @ 2026-10-15 -->
- Back-filled dispatch rows are stamped with the back-fill's own instant: all five kan-982
  back-fills read 17:47:20Z.

## What changes

- `flow record dispatch begin` takes `-started-at` and `end` takes `-ended-at`, RFC 3339, so a
  back-filled row carries the dispatch's real instants.
- `close-task.sh` refuses (exit 1, `close-task: RECORD —`, nothing ticked or pushed) to close an
  appended task no dispatch row names.
- `check-visual-verify-dispatched.sh` takes an optional canonical worktree: a satellite's diff is
  also matched against the canonical `ui paths` by its repository name, and the store is read under
  the canonical worktree; integrate passes it.
- The in-run fix loop runs the full test list twice more after a load-flake fix and folds every
  further failure into the same task, and drops a determinism fix whose flake does not reproduce.
- The verifier names motion frames with their millisecond offset and records a sibling trigger at
  the same sampling before reporting a difference between two triggers.
