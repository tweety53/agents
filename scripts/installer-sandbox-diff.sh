#!/usr/bin/env bash
# installer-sandbox-diff.sh — thin shim over `flow-guard installer-sandbox-diff`
# (stats/internal/guard/installersandboxdiff.go), whose header states the
# normalized manifest, the sandbox it keeps and the exit codes: run
# `setup.sh global` from two trees under two fresh HOME sandboxes and diff
# the installed result.
#
# Usage: installer-sandbox-diff.sh <old-tree> <new-tree>
#
#   scripts/installer-sandbox-diff.sh ../agents.main ../agents-worktrees/kan-536-…   # refactored installer vs main
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this command's
# usage-or-environment code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "installer-sandbox-diff: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec installer-sandbox-diff 2 "installer-sandbox-diff:" "$@"
