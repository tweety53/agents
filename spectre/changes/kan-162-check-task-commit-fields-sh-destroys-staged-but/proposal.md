# kan-162-check-task-commit-fields-sh-destroys-staged-but

## Why

`check-task-commit-fields.py`'s `_commit_reverted()` context manager verifies `Regression:` and
`Baseline:` by reverting a task's commit and running its test in place, then unconditionally
restoring with `git reset --hard commit_sha`. `reset --hard` discards **every** tracked-file
modification in the whole worktree, not only what the revert touched.

This is not a rare edge case. `/flow`'s implementation guardrail forbids the implementer from
committing `spectre/changes/` or `docs/superpowers/`, so a change's planning artifacts sit
staged-but-uncommitted in the worktree for the entire implementation phase, by design. Every time
this guard runs — right after each task's commit, before that task is dispatched for review — it
wipes them. It destroyed KAN-144's planning artifacts twice; the second loss was first
misattributed to the implementer subagent before the guard's own `reset --hard` was found to be the
cause. Found by KAN-144's self-review (KAN-162).

## What changes

- `check_regression` and `check_baseline` (both call `_commit_reverted`) are wrapped, at the
  `check_task_commit` call site, in a single `git stash push --include-untracked` /
  `git stash pop` cycle spanning both checks — one stash/pop per guard invocation, not one per
  check.
- If `git stash push` reports nothing to stash (clean tree), no pop is needed and behavior is
  unchanged from today.
- If `git stash pop` itself fails (a conflict against the restored tree), the guard **never** falls
  back to `git reset --hard` to paper over it. It exits 2 ("cannot determine anything"), leaves the
  stash entry in place for manual recovery, and reports `git stash list` as the way to see it. This
  is the "refuse rather than destroy" case the ticket asks for, scoped to the one situation stashing
  cannot transparently recover from.
- Applied identically to both copies of the guard: `skills/flow/scripts/check-task-commit-fields.py`
  and `scripts/check-task-commit-fields.py` (confirmed byte-identical before this change).
