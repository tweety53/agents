#!/usr/bin/env bash
# reshape-branch.sh — integrate's reshape: fold every commit a change branch
# made since <merge-base> — task, fixup and planning commits alike — back
# into the index with `git reset --soft <merge-base>`, for commit-split.sh to
# commit on top as the implementation commit and the one planning commit.
#
# Usage: reshape-branch.sh <worktree> <name> <merge-base>
#
# skills/flow-contracts/finish-contract-run1.md is canonical for the reshape;
# this script is what runs it. Only HEAD moves: the index and working tree
# keep everything they held, operator edits at the human gate included.
#
# A reset of a planning commit's branch is a planning-commit rewrite, so it
# calls check-planning-commit-location.sh first and stops on its verdict,
# printing the guard's own lines.
#
# Prints `RESHAPED: <worktree> — reset to <short sha>` and exits 0. Exits 2
# with a reason on stderr on bad arguments or a base that does not resolve,
# the guard's own code on a guard refusal, and git's own code on a git
# failure.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "reshape-branch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec reshape-branch 2 "reshape-branch:" "$@"
