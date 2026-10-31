#!/usr/bin/env bash
# check-plan-shape.sh — thin shim over `flow-guard check-plan-shape`
# (stats/internal/guard/planshape.go), whose header states the findings,
# arguments and exit codes.
#
# This file exists so .flow/project.md's declared lint/test commands and an
# operator's muscle memory invoking this exact filename keep working. Scope
# is one file per call:
#
#   - no arguments: scan every live change's tasks.md under
#     <spec-root>/changes/*/tasks.md (never the archive/ entry: run 1
#     drops a change's directory, so nothing in flight sits there),
#     aggregating exit codes — non-zero if ANY file has a violation,
#     printing each file's own violations;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
#
# The no-argument scan's root is CHECK_PLAN_SHAPE_ROOT when set (a test
# harness's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT — derived by
# flow_guard_root (scripts/lib/flow-guard.sh) — as the checkout whose changes are
# scanned.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-plan-shape: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-plan-shape 2 "check-plan-shape:" "$@"
