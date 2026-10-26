# kan-1025-flow-improvement-require-a-suite-regression-test

## Why

KAN-1019's review panel fixed three Important findings (F1 caller cwd, F2/F3 whitespace-trimming
reads) whose only proof was reproducers under `.superpowers/sdd/reproducers` — untracked, deleted
with the worktree; its self-review had to add the suite tests afterwards. Neither
`skills/flow/review-panel-fix-round.md` nor `skills/flow/review-panel.md` requires a fixed Critical
or Important finding to leave a tracked suite test behind (reachability check on `97a55ef9`: no
such wording in either file), so a regression of a reviewed-and-fixed defect goes undetected once
the worktree is gone.

## What changes

- Every fixed Critical or Important panel finding carries one `regression-test:` line in the
  round's fix report — a git-tracked test that fails at the pre-fix code, or `none — <reason>`.
- `check-panel-findings-closed` refuses the panel close (exit 1) when a fixed Critical/Important
  finding has no such line, or its path is untracked or under `.superpowers/`. Fixed Minors are
  exempt.
