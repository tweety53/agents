#!/usr/bin/env bash
# check-task-records.sh — thin wrapper.
#
# All parsing and checking logic lives in check-task-records.py (Python 3,
# standard library only), following the same split check-plan-provenance.sh
# and check-plan-shape.sh use and for the same reason: this file exists only
# so .flow/project.md's declared commands and an operator's muscle memory
# invoking this exact filename keep working, while the record checking
# underneath is a real parser — lib/plan_grammar.py's task grammar and
# check-task-commit-fields.py's field parser, imported rather than
# restated. The Python module's own docstring is canonical for the three
# checks (tick state vs commit existence, declared Files vs the commit's
# real files in both directions, Build tag format) and the exit-code
# contract.
#
#   check-task-records.sh [repo-root] [base-ref]
#
# With no arguments the repo root is the CWD and the base ref is the branch
# refs/remotes/origin/HEAD points at; zero plans in flight under
# spectre/changes/ exits 0, which is what makes the guard a legitimate
# bare-tree lint step.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/check-task-records.py"
# Named and checked for the two reasons every wrapper here names its
# siblings: a Python import is invisible to check-guard-symlinks.sh's rule
# 2, which derives a guard's required siblings by grepping its source for
# $SCRIPT_DIR/<name>, and a missing module should say which one rather than
# surface as a traceback.
GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"
COMMIT_FIELDS_MODULE="$SCRIPT_DIR/check-task-commit-fields.py"

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-task-records.sh: required sibling module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

if [ ! -f "$COMMIT_FIELDS_MODULE" ]; then
  echo "check-task-records.sh: required sibling module not found: $COMMIT_FIELDS_MODULE" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "check-task-records.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

# `command -v` proves the file exists, not that it runs (see
# check-plan-provenance.sh's own comment for the macOS-stub-python3 case
# this probe exists to catch).
if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-task-records.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -gt 2 ]; then
  echo "usage: check-task-records.sh [repo-root] [base-ref]" >&2
  exit 2
fi

exec python3 "$PYTHON_GUARD" "$@"
