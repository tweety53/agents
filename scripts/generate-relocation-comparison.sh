#!/usr/bin/env bash
# generate-relocation-comparison.sh — thin shim over
# `flow-guard generate-relocation-comparison`
# (stats/internal/guard/relocationcomparison.go), whose header states the
# comparison it writes, its arguments and exit codes.
#
# Usage: generate-relocation-comparison.sh <worktree> <changeRoot> <merge-base>
#
# The plan read is `<changeRoot>/tasks.md`; the generator spells that path
# itself, so no caller has to know the plan-directory convention.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this generator's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "generate-relocation-comparison: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec generate-relocation-comparison 2 "generate-relocation-comparison:" "$@"
