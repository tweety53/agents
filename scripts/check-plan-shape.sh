#!/usr/bin/env bash
# check-plan-shape.sh — thin wrapper.
#
# All parsing and finding logic lives in check-plan-shape.py (Python 3,
# standard library only), following the same split check-task-build-
# green.sh uses and for the same reason: this file exists only so
# .flow/project.md's declared lint/test commands and an operator's muscle
# memory invoking this exact filename keep working, while the shape
# checking underneath IMPORTS the real check-task-commit-fields.py parser
# rather than reimplementing it.
#
# Unlike check-task-commit-fields.sh (which checks ONE task's fields
# against ONE real commit), check-plan-shape.py's scope is ONE FILE per
# invocation — exactly check-task-build-green.py's convention. This
# wrapper is what resolves WHICH files that means:
#
#   - no arguments: scan every live change's tasks.md under
#     <spec-root>/changes/*/tasks.md, and that of every archived change
#     under <spec-root>/changes/archive/ the base does not carry yet
#     (`flow-guard unlanded-archives`), calling the Python script once per
#     file and aggregating exit codes — non-zero if ANY file has a
#     violation, printing each file's own violations as the Python script
#     emits them;
#   - one argument: treat it as an explicit tasks.md path and scan only
#     that file.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/check-plan-shape.py"
# The Python guard imports the real check-task-commit-fields.py parser and
# the tasks.md grammar it shares with check-task-build-green.py from
# lib/plan_grammar.py, resolving both through its own real path. Both are
# named here as well, and checked, for two reasons: a Python `import` is
# invisible to check-guard-symlinks.sh's rule 2, which derives a guard's
# required siblings by grepping its source for $SCRIPT_DIR/<name> — so
# without these lines the shipped guard would carry sibling dependencies no
# guard can see — and a missing module should say which one rather than
# surface as a traceback.
COMMIT_FIELDS_MODULE="$SCRIPT_DIR/check-task-commit-fields.py"
GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"

if [ ! -f "$COMMIT_FIELDS_MODULE" ]; then
  echo "check-plan-shape.sh: required sibling module not found: $COMMIT_FIELDS_MODULE" >&2
  exit 2
fi

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-plan-shape.sh: shared grammar module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "check-plan-shape.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs: on macOS without
# the Command Line Tools, /usr/bin/python3 is a stub that exits 1 with
# "xcrun: error: invalid active developer path", which a caller would read
# as "violations found". Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-plan-shape.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -eq 1 ]; then
  exec python3 "$PYTHON_GUARD" "$1"
fi

if [ "$#" -gt 1 ]; then
  echo "usage: check-plan-shape.sh [path-to-tasks.md]" >&2
  exit 2
fi

# No arguments: scan every in-flight change's tasks.md. Unless
# CHECK_PLAN_SHAPE_ROOT is set — in which case it names the root explicitly
# (the same opt-in override check-task-build-green.sh accepts as
# CHECK_TASK_BUILD_GREEN_ROOT), so a test harness can point this wrapper at
# a sandboxed fixture tree instead of this repository's own
# spectre/changes/ — REPO_ROOT is derived from this script's own REAL
# location, resolved through lib/resolve-file.sh's resolve_file rather than
# a fixed number of directory levels above wherever it was invoked from:
# unlike check-task-build-green.sh, this guard is ALSO shipped as a symlink
# under skills/flow/scripts/ (per check-guard-symlinks.sh rule 4), so a
# fixed-depth walk would silently answer skills/flow instead of the
# repository root when run from there.
if [ -n "${CHECK_PLAN_SHAPE_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_PLAN_SHAPE_ROOT"
else
  source "$SCRIPT_DIR/lib/resolve-file.sh"
  SELF_REAL="$(resolve_file "${BASH_SOURCE[0]}")" || {
    echo "check-plan-shape.sh: cannot resolve this script's own location" >&2
    exit 2
  }
  REPO_ROOT="$(cd "$(dirname "$SELF_REAL")/.." && pwd)"
fi
source "$SCRIPT_DIR/lib/spec-root.sh"
# flow_guard_exec runs `flow-guard unlanded-archives` below; named as
# $SCRIPT_DIR/<name> and checked before sourcing, like every sibling here.
FLOW_GUARD_LIB="$SCRIPT_DIR/lib/flow-guard.sh"
if [ ! -f "$FLOW_GUARD_LIB" ]; then
  echo "check-plan-shape.sh: shared flow-guard module not found: $FLOW_GUARD_LIB" >&2
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
  ARCHIVED="$(flow_guard_exec unlanded-archives 2 "check-plan-shape.sh:" "$REPO_ROOT")" || exit 2
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
