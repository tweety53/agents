#!/usr/bin/env bash
# panel-round-begins.sh <worktree> <change> <round> <session-token> <slot>:<model>:<effort>[:<diff-base>]...
# — records every slot's begin of one review-panel round, each under the
# literal key panel-<round>-<slot>, in the one Bash call after the round's
# launch message (skills/flow/review-panel.md, "Every slot's dispatch is
# recorded"). <round> is the number itself, never a substitution.
#
# stdout: one `<key>: <flow's answer>` line per slot, in slot order.
#
# Exit codes: 0 every begin recorded; 1 a begin failed (its output relayed,
# the later slots still recorded); 2 usage, or no flow binary on PATH.
#
# Go, in flow-guard (stats/internal/guard/panelroundbegins.go). flow-guard
# is built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "panel-round-begins: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec panel-round-begins 2 "panel-round-begins:" "$@"
