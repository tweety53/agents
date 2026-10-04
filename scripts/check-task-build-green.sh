#!/usr/bin/env bash
# check-task-build-green.sh — thin wrapper.
#
# All classification logic lives in check-task-build-green.py (Python 3,
# standard library only), following the same split check-plan-provenance.sh
# uses and for the same reason: this file exists only so
# .flow/project.md's declared lint/test commands and an operator's muscle
# memory invoking this exact filename keep working, while the block-parsing
# logic underneath gets a real language rather than a hand-rolled Bash ERE
# allowlist.
#
# Unlike check-plan-provenance (which scans the whole repository tree in
# one call), check-task-build-green.py's scope is ONE file per invocation.
# This wrapper is what resolves WHICH files that means:
#
#   - no arguments: scan every live change's tasks.md under
#     spectre/changes/*/tasks.md, and that of every archived change under
#     spectre/changes/archive/ the base does not carry yet (`flow-guard
#     unlanded-archives`), calling the Python script once per file and
#     aggregating exit codes — non-zero if ANY file has a violation,
#     printing each file's own violations as the Python script emits them;
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

# `command -v` proves the file exists, not that it runs: on macOS without
# the Command Line Tools, /usr/bin/python3 is a stub that exits 1 with
# "xcrun: error: invalid active developer path", which a caller would read
# as "violations found". Probing with a trivial program is what tells
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

# No arguments: scan every in-flight change's tasks.md. REPO_ROOT is
# derived from this script's own location so the scan works from any cwd,
# the same convention check-plan-provenance.sh uses for its own root —
# unless CHECK_TASK_BUILD_GREEN_ROOT is set, in which case it names the root
# explicitly (the same opt-in override check-plan-provenance accepts as
# CHECK_PLAN_PROVENANCE_ROOT), so a test harness can point this wrapper at a
# sandboxed fixture tree instead of this repository's own spectre/changes/.
REPO_ROOT="${CHECK_TASK_BUILD_GREEN_ROOT:-$(cd "$SCRIPT_DIR/.." && pwd)}"
source "$SCRIPT_DIR/lib/spec-root.sh"
# flow_guard_exec runs `flow-guard unlanded-archives` below; named as
# $SCRIPT_DIR/<name> and checked before sourcing, like every sibling here.
FLOW_GUARD_LIB="$SCRIPT_DIR/lib/flow-guard.sh"
if [ ! -f "$FLOW_GUARD_LIB" ]; then
  echo "check-task-build-green.sh: shared flow-guard module not found: $FLOW_GUARD_LIB" >&2
  exit 2
fi
source "$FLOW_GUARD_LIB"
CHANGES_DIR="$REPO_ROOT/$(spec_root_leaf "$REPO_ROOT")/changes"

STATUS=0

if [ -d "$CHANGES_DIR" ]; then
  # The glob (one directory level under CHANGES_DIR) never descends into an
  # archived change's archive/<name>/tasks.md, a level deeper. An archived
  # change joins the scan only while the base does not carry it — archived
  # on this branch by integrate's run 1, then written into by a fix run — as
  # `flow-guard unlanded-archives` names it
  # (stats/internal/guard/unlandedarchives.go, the one statement of that
  # rule); every archive the base carries stays out.
  ARCHIVED="$(flow_guard_exec unlanded-archives 2 "check-task-build-green.sh:" "$REPO_ROOT")" || exit 2
  TASKS_FILES=("$CHANGES_DIR"/*/tasks.md)
  while IFS= read -r archived_name; do
    if [ -n "$archived_name" ]; then
      TASKS_FILES+=("$CHANGES_DIR/archive/$archived_name/tasks.md")
    fi
  done <<<"$ARCHIVED"
  for tasks_file in "${TASKS_FILES[@]}"; do
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
