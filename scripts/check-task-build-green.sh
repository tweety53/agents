#!/usr/bin/env bash
# check-task-build-green.sh — thin wrapper.
#
# All classification logic lives in check-task-build-green.py (Python 3,
# standard library only), following the same split check-plan-provenance.sh
# uses and for the same reason: this file exists only so
# .myflow/project.md's declared lint/test commands and an operator's muscle
# memory invoking this exact filename keep working, while the block-parsing
# logic underneath gets a real language rather than a hand-rolled Bash ERE
# allowlist.
#
# Unlike check-plan-provenance.py (which scans the whole repository tree in
# one call), check-task-build-green.py's scope is ONE file per invocation.
# This wrapper is what resolves WHICH files that means:
#
#   - no arguments: scan every non-archived change's tasks.md under
#     openspec/changes/*/tasks.md (excluding openspec/changes/archive/),
#     calling the Python script once per file and aggregating exit codes —
#     non-zero if ANY file has a violation, printing each file's own
#     violations as the Python script emits them;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/check-task-build-green.py"
# The Python guard imports the tasks.md grammar it shares with
# check-task-commit-fields.py from lib/plan_grammar.py, resolving it through
# its own real path. It is named here as well, and checked, for two
# reasons: a Python `import` is invisible to check-guard-symlinks.sh's rule
# 2, which derives a guard's required siblings by grepping its source for
# $SCRIPT_DIR/<name> — so without this line the shipped guard would carry a
# sibling dependency no guard can see — and a module that is missing should
# say so rather than surface as a traceback.
GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-task-build-green.sh: shared grammar module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "check-task-build-green.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# check-plan-provenance.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch). Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-task-build-green.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -eq 1 ]; then
  exec python3 "$PYTHON_GUARD" "$1"
fi

if [ "$#" -gt 1 ]; then
  echo "usage: check-task-build-green.sh [path-to-tasks.md]" >&2
  exit 2
fi

# No arguments: scan every non-archived change's tasks.md. REPO_ROOT is
# derived from this script's own location so the scan works from any cwd,
# the same convention check-plan-provenance.sh uses for its own root —
# unless CHECK_TASK_BUILD_GREEN_ROOT is set, in which case it names the root
# explicitly (the same opt-in override check-plan-provenance.py accepts as
# CHECK_PLAN_PROVENANCE_ROOT), so a test harness can point this wrapper at a
# sandboxed fixture tree instead of this repository's own openspec/changes/.
REPO_ROOT="${CHECK_TASK_BUILD_GREEN_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
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
