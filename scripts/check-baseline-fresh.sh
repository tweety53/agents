#!/usr/bin/env bash
# check-baseline-fresh.sh — thin wrapper.
#
# All parsing and freshness logic lives in check-baseline-fresh.py (Python
# 3, standard library only), following the same split check-plan-shape.sh
# uses and for the same reason: this file exists only so .flow/project.md's
# declared lint commands and an operator's muscle memory invoking this
# exact filename keep working, while the freshness logic underneath IMPORTS
# the real shared parsers rather than reimplementing them.
#
# Usage: check-baseline-fresh.sh <changeRoot>
#   <changeRoot> is the change directory holding the tasks.md to judge —
#   the brainstorm-planner.md section D call site passes <changeRoot>.
#
# Exit codes (check-baseline-fresh.py's docstring is canonical):
#   0  fresh, or nothing to check
#   1  stale hit — a declared directory is absent or predates the worktree's
#      newest commit
#   2  cannot answer — usage, unreadable tasks.md, ambiguous or empty key,
#      escaping path, or git cannot answer
#
# Sibling declarations for check-guard-symlinks.sh rule 2 (a Python import
# is invisible to its grep): the Python guard loads the shared grammar and
# the field parser beside itself, and both are named and checked here so a
# missing module says which one rather than surfacing as a traceback.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/check-baseline-fresh.py"
COMMIT_FIELDS_MODULE="$SCRIPT_DIR/check-task-commit-fields.py"
GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"

if [ ! -f "$PYTHON_GUARD" ]; then
  echo "check-baseline-fresh.sh: required sibling module not found: $PYTHON_GUARD" >&2
  exit 2
fi

if [ ! -f "$COMMIT_FIELDS_MODULE" ]; then
  echo "check-baseline-fresh.sh: required sibling module not found: $COMMIT_FIELDS_MODULE" >&2
  exit 2
fi

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-baseline-fresh.sh: shared grammar module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "check-baseline-fresh.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# check-plan-provenance.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch). Probing with a trivial program is what tells
# "python3 is a name on PATH" apart from "python3 is a working interpreter".
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-baseline-fresh.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -ne 1 ]; then
  echo "usage: check-baseline-fresh.sh <changeRoot>" >&2
  exit 2
fi

exec python3 "$PYTHON_GUARD" "$1"
