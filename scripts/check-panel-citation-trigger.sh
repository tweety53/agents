#!/usr/bin/env bash
# check-panel-citation-trigger.sh — decide whether the review panel's
# citation pre-check should run, in a guard instead of prose.
#
# Usage: check-panel-citation-trigger.sh <worktree> <merge-base>
#
# Prints nothing; the exit code is the whole answer.
#   0  the change's own paths include at least one path ending .md or .mdc
#   1  none do
#   2  usage error: a missing argument, <worktree> not a directory or not a
#      git worktree, <merge-base> not resolving, or a git invocation failed
#
# THE CHANGE'S OWN PATHS INCLUDE THE INDEX AND THE WORKING TREE (design.md:
# touched-paths-include-index-and-worktree, the same rule check-base-moved.sh
# already applies): the union of what HEAD carries since the merge base,
# what is staged, and what is unstaged.
#
# The logic is the Go port in stats/internal/guard/panelcitationtrigger.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-panel-citation-trigger: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-citation-trigger 2 "check-panel-citation-trigger:" "$@"
