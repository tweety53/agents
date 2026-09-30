#!/usr/bin/env bash
# sync-panel-base.sh — bring a review panel's worktree onto its base's tip, in
# a guard instead of the hand-typed recipe skills/flow/review-panel.md carried
# under "Check base movement first" (KAN-860).
#
# Usage: sync-panel-base.sh [--rebase] <worktree> <working-notes-merge-base>
#
# Fetches through resolve-base-branch and prints `BASE: <base>`, then runs
# check-base-moved against origin/<base> and <working-notes-merge-base>,
# echoing every check-base-moved line it reaches on stdout:
#   CLEAR  -> nothing more, exit 0.
#   MOVED  with no overlap, or with any overlap under --rebase (the
#          operator's **Rebase**) -> a bare `git rebase` onto origin/<base>'s
#          tip, resolved to a sha first; nothing is set aside. Each clean
#          rebase prints
#            REBASED: <worktree> — merge base <sha>
#          — the worktree's new working-notes merge base; the last such line
#          wins — and check-base-moved re-runs against that sha. A re-check
#          that answers MOVED with no overlap rebases again, at most 3
#          rebases in all; one that answers MOVED with an overlap stops.
#   MOVED  with an overlap it did not rebase over -> exit 3: offer the
#          operator the panel's base-movement prompt; its **Rebase** re-runs
#          this with --rebase and the newest merge base.
# A rebase that stops on a conflict prints
#   CONFLICT: <worktree> — onto <sha>; unmerged: <paths>
# and exits 1, the worktree left mid-rebase exactly as git left it — never
# resolve it; hand it off.
#
# Exit 0 settled (CLEAR, or REBASED then CLEAR); 1 CONFLICT; 3 an overlap
# awaiting the operator; 2 everything else — usage, a base branch
# resolve-base-branch cannot resolve (its reason precedes this guard's on
# stderr), a REFUSE verdict, a rebase git refused to start (nothing changed),
# or a base still moving after 3 rebases. Exit 2 means stop and ask.
#
# The guard is Go: stats/internal/guard/syncpanelbase.go, over the shared
# baseMoved verdict (basemoved.go) and rebaseOntoTip core (baserebase.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "sync-panel-base: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec sync-panel-base 2 "sync-panel-base:" "$@"
