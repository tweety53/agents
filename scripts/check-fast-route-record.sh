#!/usr/bin/env bash
# check-fast-route-record.sh — the /flow-fast record guard (KAN-838).
#
# A fast route writes no plan, so the change branch's commit series is its
# whole record and this guard is what reads it back before the route lands
# anything — skills/flow-fast/SKILL.md's verify stage calls it with the
# worktree and the base resolve-base-branch.sh printed. The contract — what
# each commit since the merge base must carry, the finding lines, the exit
# codes — is the Go file's own header comment, canonical for itself:
# stats/internal/guard/fastrouterecord.go.
#
#   check-fast-route-record.sh <worktree> <base>
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-fast-route-record: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-fast-route-record 2 "check-fast-route-record:" "$@"
