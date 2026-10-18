# self-review-backlog-flow-8

## Why

The next five live tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier
self-review passes and not already fixed on `origin/main`, are planned here as one self-review fix
change (`design.md`, **Context**):

- **KAN-821** — sibling tickets sharing a mockup frame family have no rule to scope their parts
  per ticket or to declare the not-yet-built differences, so a verifier reads a sibling's unbuilt
  parts as departures; gymie's kan-743 did it by hand and kept KAN-744's scope intact.
- **KAN-831** — `run-reproducer.sh`'s 20-second bound times out a reproducer that builds or runs
  tests unless every caller remembers `RUN_REPRODUCER_BOUND_SECONDS`; the reproducer cannot carry
  its own bound.
- **KAN-837** — **The loop**'s shared-source recapture runs `--update-snapshots=all` over the
  whole suite, so failures the fix did not cause are written over their baselines too; gymie's
  kan-748 re-baselined per capture by blob-classifying each diff image instead.
- **KAN-883** — `check-worktree-location.sh` fails on another session's harness agent worktree
  under `<project>/.claude/worktrees/`, a stray no flow run created and none can remove.
- **KAN-885** — `plan-class.sh` measures the change's diff from the merge base it is given, so a
  base ref resolved after the base moved counts the base's own new commits as the change's.

## What changes

- `skills/flow/brainstorm-planner.md` plans a frame a sibling ticket finishes as this change's
  parts only, its sidecar line under a comment naming the expected differences;
  `skills/flow/visual-verify-verifier.md` reads those parts as expected, not as departures.
- `run-reproducer.sh` reads a `# reproducer-bound: <seconds>` header line and bounds every run of
  that reproducer by it; **review-panel.md** asks for the line and **review-panel-fix-round.md**
  drops the per-call knob.
- **The loop** (`skills/flow/verify-fix-loop.md`) re-baselines per capture, by blob
  classification against the fix's footprint, when the suite also fails for another reason.
- `check-worktree-location.sh` names a harness agent worktree on a non-failing `FOREIGN:` line.
- `plan-class.sh` measures the diff side from `git merge-base HEAD <merge-base>`.
- `scripts/test-check-task-records.sh`'s two pipes into `grep -q`, which fail
  `TestNoPipeIntoEarlyExitGrep` on `origin/main`, become here-strings.
