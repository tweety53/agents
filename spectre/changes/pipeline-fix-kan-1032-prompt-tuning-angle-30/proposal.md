# pipeline-fix-kan-1032-prompt-tuning-angle-30

## Why

The operator asked: "update flow to always sync before implementation". Today `/flow`'s
implementation run only names base movement before task 1 — `check-base-behind.sh` relays a
recorded `--base` that lags what it tracks, and `refresh-plan-base.sh` names what moved under the
plan — and leaves the change branch where it was cut: "The step names, it never rebases: the change
branch is synced onto the base at integrate". On kan-1032 the base moved 13 commits during
planning, overlapping a planned file; the review panel's `sync-panel-base.sh` then met the overlap
(exit 3) only after every task and stopped for an operator answer, and its rebase pulled in a new
self-review report that broke a guard. A rebase before task 1 would have surfaced both up front.
`/flow-fast` has the same gap.

## What changes

- `/flow`'s implementation run, on every run — creating, resumed or fix — rebases the change
  branch onto its base before task 1, in every worktree of the run's resolved set, with
  `sync-panel-base.sh --rebase <worktree> <merge-base>`, right after the plan refresh names what
  moved.
- After a clean rebase the worktree's working-notes merge base becomes the sha the last
  `REBASED:` line names; the run records it with `flow state add-worktree` and pushes the
  rewritten branch `--force-with-lease`.
- A rebase conflict (exit 1) is handled as the review panel's **Conflict** bullet: never resolved,
  the worktree left mid-rebase, and the run stops for the operator. Exit 2 or 3 stops the run.
- `/flow-fast` runs the same step at the start of its implement section, without the state-file
  call, pushing `<name>`; its `<merge-base>` substitution follows the rebase.
- **Branch backup** (`skills/flow-contracts/git-boundaries.md`) names the base-sync rebase as a
  `--force-with-lease` push, and `refresh-plan-base.sh`'s header points at the rebase that now
  follows it.
- A repotest test pins the step in both skills.
