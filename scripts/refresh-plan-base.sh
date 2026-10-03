#!/usr/bin/env bash
# refresh-plan-base.sh — report what the base branch moved under a plan
# before its first task runs (skills/flow/implement.md, "The plan is
# refreshed against the base before task 1 runs").
#
# Usage: refresh-plan-base.sh <worktree> <merge-base> <tasks.md> [<spec-path>…]
#
# Fetches through resolve-base-branch, then compares <merge-base> with
# origin/<base>. Prints to stdout:
#   UNMOVED: <worktree> — origin/<base> has not moved since <merge-base>
# or, on a moved base:
#   MOVED: <worktree> — <n> commits on origin/<base> since <merge-base>
#   CHANGED: <path>          one per path `git diff --name-only <merge-base>
#                            origin/<base>` names that a **Files:** field of
#                            <tasks.md> declares, exactly or as a leading
#                            directory; `CHANGED: none of the plan's
#                            **Files:** paths` when there is none
#   ----- BEGIN <spec-path> @ origin/<base> -----
#   <the spec's text at origin/<base>>
#   ----- END <spec-path> -----
#                            per <spec-path>, or instead
#   SPEC-ABSENT: <spec-path> — not at origin/<base>
#
# The intersection is done in Go, not as a git pathspec: a multi-repo plan's
# **Files:** names paths outside this repository, which git refuses.
#
# Exit 0 whenever it answered, UNMOVED or MOVED; exit 2 when it cannot
# answer — usage, a merge base that is not a commit, an unreadable
# <tasks.md>, or a base branch resolve-base-branch cannot resolve (its own
# reason precedes this guard's on stderr). It names, it never rebases: the
# change branch is synced onto the base at integrate.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "refresh-plan-base: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec refresh-plan-base 2 "refresh-plan-base:" "$@"
