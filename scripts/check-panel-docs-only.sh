#!/usr/bin/env bash
# check-panel-docs-only.sh — decide whether this change's own paths are
# entirely documentation, in a guard instead of prose.
#
# Usage: check-panel-docs-only.sh <worktree> <merge-base>
#
# Exit codes:
#   0  every touched path ends .md or .mdc
#   1  at least one touched path does not — that first path is printed to
#      stdout — including an empty touched-path set
#   2  cannot answer: a missing argument, <worktree> not a directory or not
#      a git worktree, <merge-base> not resolving, or a git invocation
#      failed
#
# THE CHANGE'S OWN PATHS are the union of committed-since-merge-base,
# staged and unstaged paths, exactly as check-panel-citation-trigger.sh
# collects them.
#
# WHY THIS GUARD EXISTS (KAN-312): a docs-only branch has no code seam for a
# second slot to find; the implementer's self-review and the prose guards
# cover it.
#
# The logic is the Go port in stats/internal/guard/paneldocsonly.go; the
# touched-path collection is stats/internal/guard/paneltouchedpaths.go's.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-panel-docs-only: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-docs-only 2 "check-panel-docs-only:" "$@"
