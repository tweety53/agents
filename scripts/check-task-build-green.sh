#!/usr/bin/env bash
# check-task-build-green.sh — thin shim over `flow-guard check-task-build-green`.
#
# All classification logic lives in stats/internal/guard/taskbuildgreen.go,
# whose header carries the rule's pointer, the parsing model and the exit
# codes. This file exists only so .flow/project.md's declared lint/test
# commands and an operator's muscle memory invoking this exact filename keep
# working.
#
#   - no arguments: scan every live change's tasks.md under
#     spectre/changes/*/tasks.md, and that of every archived change under
#     spectre/changes/archive/ the base does not carry yet (unlanded-archives,
#     run in-process), aggregating exit codes — non-zero if ANY file has a
#     violation, printing each file's own violations;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
#
# The no-argument scan's root is CHECK_TASK_BUILD_GREEN_ROOT when set (a test
# harness's sandboxed fixture tree), else FLOW_GUARD_REPO_ROOT — the parent
# of this script's real directory — so the scan works from any cwd.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-task-build-green: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
# The root is this script's REAL parent's parent, resolved through
# lib/resolve-file.sh rather than `dirname` of the invoked path: run through
# a symlink under skills/<skill>/scripts/, the invoked path's parent's parent
# is skills/<skill> — a scan of a tree with no changes/ there exits 0 with no
# output.
[ -r "$SCRIPT_DIR/lib/resolve-file.sh" ] && . "$SCRIPT_DIR/lib/resolve-file.sh" || {
  echo "check-task-build-green: cannot load lib/resolve-file.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
SELF_REAL="$(resolve_file "${BASH_SOURCE[0]}")" || {
  echo "check-task-build-green: cannot resolve this script's own location" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "$SELF_REAL")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-task-build-green 2 "check-task-build-green:" "$@"
