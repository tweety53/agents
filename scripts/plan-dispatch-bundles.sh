#!/usr/bin/env bash
# plan-dispatch-bundles.sh — thin wrapper.
#
# All classification logic lives in plan-dispatch-bundles.py (Python 3,
# standard library only), following the same split check-task-build-green.sh
# uses and for the same reason: this file exists only so an operator's
# muscle memory invoking this exact filename, and `myflow-do` §4's own call
# site, keep working, while the block-parsing logic underneath gets a real
# language rather than a hand-rolled Bash ERE allowlist.
#
# Unlike check-plan-provenance.py (which scans the whole repository tree in
# one call), plan-dispatch-bundles.py's scope is ONE file per invocation.
# This wrapper is what resolves WHICH files that means:
#
#   - no arguments: scan every non-archived change's tasks.md under
#     openspec/changes/*/tasks.md (excluding openspec/changes/archive/),
#     calling the Python script once per file and aggregating exit codes —
#     non-zero if ANY file reports a violation or cannot be read, printing
#     each file's own bundles or violations as the Python script emits
#     them;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/plan-dispatch-bundles.py"

command -v python3 >/dev/null 2>&1 || {
  echo "plan-dispatch-bundles.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# check-task-build-green.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch). Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "plan-dispatch-bundles.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -eq 1 ]; then
  exec python3 "$PYTHON_GUARD" "$1"
fi

if [ "$#" -gt 1 ]; then
  echo "usage: plan-dispatch-bundles.sh [path-to-tasks.md]" >&2
  exit 2
fi

# No arguments: scan every non-archived change's tasks.md. REPO_ROOT is
# derived from this script's own location so the scan works from any cwd,
# the same convention check-task-build-green.sh uses for its own root —
# unless PLAN_DISPATCH_BUNDLES_ROOT is set, in which case it names the root
# explicitly (the same opt-in override check-task-build-green.sh accepts as
# CHECK_TASK_BUILD_GREEN_ROOT), so a test harness can point this wrapper at
# a sandboxed fixture tree instead of this repository's own
# openspec/changes/.
REPO_ROOT="${PLAN_DISPATCH_BUNDLES_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
CHANGES_DIR="$REPO_ROOT/openspec/changes"

STATUS=0

if [ -d "$CHANGES_DIR" ]; then
  # The glob below (one directory level under CHANGES_DIR) never descends
  # into openspec/changes/archive/*/tasks.md in the first place — an
  # archived change's tasks.md is nested a level deeper, at
  # archive/<date-name>/tasks.md, so the glob's own depth already excludes
  # every archived file without any extra filtering here.
  for tasks_file in "$CHANGES_DIR"/*/tasks.md; do
    [ -e "$tasks_file" ] || continue
    rc=0
    python3 "$PYTHON_GUARD" "$tasks_file" || rc=$?
    if [ "$rc" -eq 2 ] || [ "$STATUS" -eq 2 ]; then
      STATUS=2
    elif [ "$rc" -ne 0 ]; then
      STATUS=1
    fi
  done
fi

exit "$STATUS"
