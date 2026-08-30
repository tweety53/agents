# kan-162-check-task-commit-fields-sh-destroys-staged-but — design

## Context

`_commit_reverted(worktree, commit_sha)` (`check-task-commit-fields.py`) does `git revert
--no-commit --no-edit commit_sha`, yields for the caller's test run, then always
`git reset --hard commit_sha` in a `finally`. `check_regression` and `check_baseline` each open
this context manager once, so a single guard invocation on a task that declares both fields resets
hard twice.

`reset --hard` resets the whole index and working tree to `commit_sha`'s tree, discarding every
staged and unstaged tracked-file change repo-wide — not merely what the revert modified. `/flow`
worktrees carry staged-but-uncommitted planning artifacts (`spectre/changes/`,
`docs/superpowers/`) throughout implementation, so this guard destroys them on every task commit
it checks.

## Decisions

### Stash/restore, not a throwaway worktree

**ID:** stash-not-worktree
**Status:** active
**Chosen:** Wrap the revert-and-test cycle in `git stash push --include-untracked` /
`git stash pop`, in the same worktree.
**Considered:** Running the revert-and-test cycle in a separate `git worktree add` checkout, as the
ticket's second suggestion. Rejected: a fresh linked worktree carries no installed dependencies
(`node_modules`, a virtualenv, a build cache) since those are typically gitignored, and this guard's
project-supplied `## test` command is not guaranteed to be runnable without them. A project would
need to reinstall its whole toolchain per guard invocation, which is not merely slow but may not be
possible in a project without repeating a project-specific setup this repository has no path to
running here. Stashing keeps the check running in the environment that already has whatever the
project's test command needs.

### One stash/pop cycle per guard invocation, not one per check

**ID:** one-stash-cycle
**Status:** active
**Chosen:** `check_task_commit` stashes once, runs `check_regression` then `check_baseline`, then
pops once.
**Considered:** Stashing inside `_commit_reverted` itself, once per call. Rejected: a task
declaring both `Regression:` and `Baseline:` would stash and pop twice in one guard run for no
benefit, and a stash/pop that fails on the *first* call would leave the second call operating
against a still-uncertain tree state.

### A failed pop refuses rather than resets

**ID:** refuse-on-pop-conflict
**Status:** active
**Chosen:** If `git stash pop` exits non-zero (a conflict against the tree `git reset --hard` left
behind), the guard exits 2 immediately, names the stash by pointing at `git stash list`, and does
**not** attempt `git reset --hard` or any other repair. The stash entry is left exactly where `git
stash pop` left it.
**Considered:** Falling back to `git reset --hard` to force a clean state after a failed pop,
which is the exact operation this change exists to stop performing over unrecoverable state.
Rejected outright — it would silently discard the very data the stash was protecting, the same
failure this whole change fixes, just moved one step later.

### `git stash push` reporting nothing to stash is not an error

**ID:** clean-tree-is-a-noop
**Status:** active
**Chosen:** When `git stash push` reports no local changes were saved (the tree was already
clean — the common case for a guard run against an isolated checkout with no in-flight planning
artifacts), no pop is attempted and the checks proceed exactly as they do today.
**Considered:** Nothing — this is the only sound reading of a no-op stash.

## Open questions

None. Every question was resolved in the design's own presentation.
