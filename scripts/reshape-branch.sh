#!/usr/bin/env bash
# reshape-branch.sh — integrate's reshape: collapse a change branch's task
# and fixup commits back into the working tree while keeping every planning
# commit as the separate commit it was made as.
#
# Usage: reshape-branch.sh <worktree> <name> <merge-base>
#
# Replaces the bare `git reset --soft <merge-base>` integrate once ran, which
# folded every planning commit into the single planning commit the two-commit
# chain then made. skills/flow-contracts/finish-contract-run1.md is canonical
# for the reshape; this script is what runs it.
#
# WHAT IT DOES. A planning commit is any non-merge commit in
# <merge-base>..HEAD touching `<leaf>/changes/` — the pathspec-scoped
# commits of git-boundaries.md's **Planning commits**, link commits included,
# and /flow-plan's capture commit. Task and fixup commits never touch that
# directory (check-task-commit-planning-paths.sh), so the directory's state
# at each planning commit is exactly the merge base's plus the planning
# commits up to it. Each is rebuilt, in order, on top of <merge-base> as a
# commit whose tree is its parent's with `<leaf>/changes/` replaced by that
# planning commit's own — message and author kept, committer now. Then HEAD
# is moved there with `reset --soft`, so the index and working tree keep
# everything they held: the task and fixup work, operator edits at the human
# gate and any uncommitted planning delta, all uncommitted, for
# commit-split.sh to commit on top as the implementation commit and the last
# planning commit.
#
# THE WORKING TREE AND THE REAL INDEX ARE NEVER TOUCHED until the final
# `reset --soft`, so a failure anywhere before that leaves the branch exactly
# as it was.
#
# Planning commits run behind check-planning-commit-location.sh, and this
# script rewrites them, so it calls that guard first and stops on its
# verdict, printing the guard's own lines.
#
# Prints `RESHAPED: <worktree> — <n> planning commit(s) kept on <short sha>`
# and exits 0. Exits 2 with a reason on stderr on bad arguments or a base
# that does not resolve, the guard's own code on a guard refusal, and git's
# own code on a git failure.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "reshape-branch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec reshape-branch 2 "reshape-branch:" "$@"
