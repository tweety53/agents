# kan-312-myflow-per-task-review-and-the-panel-duplicate

## Why

On a documentation-only branch the whole-branch review panel re-reads text every per-task
reviewer has already read, against the same plan and the same delta specs, and finds nothing.
KAN-302 (4 tasks, two Markdown files, 114 changed lines) ran 9 reviewer dispatches — 4 per-task
reviewers, 2 per-task re-reviews, 3 panel slots — and every finding came from the per-task layer;
all three panel slots returned clean. The panel earns its cost where commit seams exist that no
per-task reviewer can reach by construction (KAN-287: six Major defects at commit seams in Go
store and harvester code, 24 tasks / 17 commits). The signal separating those two branches is diff
composition: KAN-302 touched only `.md` paths; KAN-287 did not.

`skills/flow/review-panel.md` today dispatches the settings-store roster in full on every pass 1,
and its own text names a "documentation-, prompt-, or test-only diff" as a case that runs that
list alone — paying a full second read exactly where there is one commit's worth of subject matter
per reviewer.

## What changes

- A shipped guard, `check-panel-docs-only.sh <worktree> <merge-base>`, answers one question by
  exit code: is every path this branch touched — committed since the merge base, staged or
  unstaged — a `.md` or `.mdc` file? Exit 0 yes; exit 1 no, printing the first path that is not;
  exit 2 cannot answer.
- On exit 0 the panel's pass 1 dispatches `primary` alone plus any slot the operator's existing
  per-run instruction names; every other resolved slot is recorded as not dispatched with the
  reason. On exit 1 or 2 the resolved roster runs in full, unchanged.
- The guard re-runs beside the cap check on every fix round: a branch that stops being docs-only
  dispatches every resolved slot not yet run this run, on the whole `final-review.diff`.
- The verdict and the roster actually dispatched are recorded in `final-review-panel.md` on every
  run, and the `Panel:` handoff line gains a `reduced:` field.
- Unchanged: findings, reproducers, fix rounds, mutation-proof, `check-panel-findings-closed.sh`,
  and the zero-open-findings handoff bar. No slot is ever added automatically; this is the one
  automatic reduction, and it only removes.
