#!/usr/bin/env bash
# check-planning-commit-location.sh — refuse a planning commit (a link commit
# after `spectre link`, any other `chore(spectre): …` commit over the change
# folder, or the two-commit chain's planning half) anywhere but the change's
# own worktree on its own `spectre/<name>` branch.
#
# Usage: check-planning-commit-location.sh <worktree> <name>
#
# Verdict lines, on stdout:
#
#   PLANNING-COMMIT-MAIN-CHECKOUT: <worktree>          the path is a
#                                                     repository's main
#                                                     checkout, not a
#                                                     linked worktree
#   PLANNING-COMMIT-WRONG-BRANCH: <worktree> on <branch|detached> — expected spectre/<name>
#   PLANNING-COMMIT-LOCATION-OK: <worktree> on spectre/<name>
#   PLANNING-COMMIT-TICKED-APPEND: <tasks.md path> task <n> — appended already ticked; …
#
# The last line follows the location verdict, once per task the working
# tree's `<name>/tasks.md` adds over HEAD's already ticked (`- [x] <n>.`)
# whose number HEAD's file did not carry: a task is appended unticked and
# committed before its work runs (kan-1010's self-review). A tasks.md HEAD
# lacks is a new plan and is not read. reshape-branch.sh and commit-split.sh
# run the location check alone — their planning half carries every tick.
#
# A tasks.md the working tree appends a task to — any task number HEAD's file did not carry — is also run through check-plan-shape.sh and check-task-build-green.sh, their lines printed as they print them; either one's exit 1 is this guard's exit 1 and its exit 2 this guard's exit 2 (kan-1015: tasks 15–24 were appended with `**Build:** pending` and no declared tests).
#
# Both location violation lines print when both hold. Exit 0 on the OK
# verdict with no ticked append; exit 1 on any violation; exit 2 with NOTHING
# on stdout when the arguments are missing or empty, or <worktree> is not a readable git work tree — an
# inability is never reported as a verdict.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-planning-commit-location.sh: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-planning-commit-location 2 "check-planning-commit-location.sh:" "$@"
