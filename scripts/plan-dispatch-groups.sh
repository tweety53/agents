#!/usr/bin/env bash
# plan-dispatch-groups.sh — thin shim over `flow-guard plan-dispatch-groups`
# (stats/internal/guard/plandispatchgroups.go), whose header states the
# grouping rule, arguments and exit codes.
#
# Unlike plan-dispatch-bundles.sh, this carries only the one-argument form:
# the Decide step (skills/flow/brainstorm-planner.md, step 5) always names
# one explicit tasks.md path, and there is no "every open change" scan to
# resolve.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "plan-dispatch-groups: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec plan-dispatch-groups 2 "plan-dispatch-groups:" "$@"
