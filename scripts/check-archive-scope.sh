#!/usr/bin/env bash
# check-archive-scope.sh — refuse an archive commit whose staged diff reaches
# outside the path prefixes that commit is allowed to touch, in a guard
# instead of trusting `git add -A` to have only picked up the archive move.
#
# Usage: check-archive-scope.sh <worktree> <allowed-prefix> [<allowed-prefix> ...]
#
# Prints one OUT-OF-SCOPE line per offending staged path, then ONE verdict
# line:
#   SCOPE-OK: <worktree>                    every staged path matches a prefix
#   SCOPE-VIOLATION: <worktree> — <n>       <n> staged path(s) do not
#
# Exit 0 on SCOPE-OK (including nothing staged at all — an empty diff is
# never a violation), 1 on SCOPE-VIOLATION, 2 with NOTHING on stdout when
# <worktree> is not a readable directory, is not a git worktree, or no
# <allowed-prefix> was given — an inability to answer is never reported as a
# verdict.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-archive-scope: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-archive-scope 2 "check-archive-scope:" "$@"
