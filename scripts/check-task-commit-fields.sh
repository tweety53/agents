#!/usr/bin/env bash
# check-task-commit-fields.sh — thin wrapper.
#
# All field-parsing and diff-checking logic lives in
# check-task-commit-fields.py (Python 3, standard library only), following
# the same split check-task-build-green.sh uses and for the same reason:
# this file exists only so an operator's muscle memory invoking this exact
# filename, and .myflow/project.md's declared commands, keep working, while
# the field grammar underneath gets a real parser.
#
# Unlike check-task-build-green.sh (which resolves WHICH tasks.md files to
# scan, zero or more of them, with no per-task identity), this guard checks
# ONE task's fields against ONE real commit, so its calling convention names
# the task and commit explicitly rather than scanning:
#
#   check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha]
#
# This wrapper's own job is resolving WHICH tasks.md the named task lives
# in: the single non-archived tasks.md under
# <worktree>/spectre/changes/*/tasks.md. Zero or more than one such file is
# an invocation error (exit 2) — this guard runs against one change's
# worktree, which spectre's own layout guarantees holds exactly one active
# change's tasks.md outside archive/.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_GUARD="$SCRIPT_DIR/check-task-commit-fields.py"
# The Python guard imports the tasks.md grammar it shares with
# check-task-build-green.py from lib/plan_grammar.py, resolving it through
# its own real path. It is named here as well, and checked, for two
# reasons: a Python `import` is invisible to check-guard-symlinks.sh's rule
# 2, which derives a guard's required siblings by grepping its source for
# $SCRIPT_DIR/<name> — so without this line the shipped guard would carry a
# sibling dependency no guard can see — and a module that is missing should
# say so rather than surface as a traceback.
GRAMMAR_MODULE="$SCRIPT_DIR/lib/plan_grammar.py"

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-task-commit-fields.sh: shared grammar module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

command -v python3 >/dev/null 2>&1 || {
  echo "check-task-commit-fields.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-task-commit-fields.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -lt 3 ] || [ "$#" -gt 4 ]; then
  echo "usage: check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha]" >&2
  exit 2
fi

WORKTREE="$1"
TASK_ID="$2"
COMMIT_SHA="$3"
PARENT_SHA="${4:-}"

if [ ! -d "$WORKTREE" ]; then
  echo "check-task-commit-fields.sh: worktree not found: $WORKTREE" >&2
  exit 2
fi

CHANGES_DIR="$WORKTREE/spectre/changes"
MATCHES=()
if [ -d "$CHANGES_DIR" ]; then
  for tasks_file in "$CHANGES_DIR"/*/tasks.md; do
    [ -e "$tasks_file" ] || continue
    MATCHES+=("$tasks_file")
  done
fi

if [ "${#MATCHES[@]}" -eq 0 ]; then
  echo "check-task-commit-fields.sh: no tasks.md found under $CHANGES_DIR" >&2
  exit 2
fi

if [ "${#MATCHES[@]}" -gt 1 ]; then
  echo "check-task-commit-fields.sh: more than one tasks.md found under $CHANGES_DIR, cannot resolve which change: ${MATCHES[*]}" >&2
  exit 2
fi

TASKS_MD="${MATCHES[0]}"

if [ -n "$PARENT_SHA" ]; then
  exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA" "$PARENT_SHA"
fi

exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA"
