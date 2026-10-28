# self-review-kan-1047-review-dynamics-heatmap-by-project

## Why

kan-1047's self-review pass found five pipeline defects, none `big`:

- The whole-suite visual `verify` that must end the parent's full-suite run on a UI change is
  prose only: `run-verify-list.sh` has no visual step and nothing checks it ran, so a base spec
  asserting the old default range preset surfaced only at a visual-verify dispatch.
- A plan whose design cites an approved prototype can touch `ui paths` with no task capturing the
  states the prototype draws; kan-1047's first three tasks named no file under `stats/web/tests/visual`.
- The prototype file was a caption and a JPEG of the live page, so `flow.visual-verify`, which
  renders the whole file as the frame, never paired it with a capture.
- Mutation reproducers were written broken: an EXIT trap restoring a relative path after `cd
  stats`, nested single quotes in `sed`, a reversed exit convention.
- A mutation reproducer left a live mutant (`2 * s * 1.12` → `2 * s` in
  `stats/web/src/components/DynamicsChart.tsx`) on the canonical worktree: `run-reproducer.sh`
  restores only the paths that were dirty before the run.

## What changes

- `run-reproducer.sh` puts back every tracked path a run changed, clean before it or not.
- The mutation-testing brief forbids `cd` and nested shell quoting in a reproducer.
- The prototype rule says the file draws the page alone.
- `check-plan-shape.sh` gains F22: a design citing a prototype, a task touching `ui paths`, and no
  task naming a file under `screenshots` is refused.
- `run-verify-list.sh <worktree> test` ends with the `## visual verification` `verify` run when its
  worktree's diff touches `ui paths` and the decision did not skip visual verification.
