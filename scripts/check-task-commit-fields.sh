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
# in, among the non-archived ones under
# <worktree>/spectre/changes/*/tasks.md. Zero is an invocation error
# (exit 2), and so is more than one ROOT change — two changes neither of
# which is the other's sub-change, which nothing here may guess between.
#
# SPECTRE'S LAYOUT GUARANTEES NOTHING LIKE "exactly one active change", and
# this header used to say it did. That assumption is what broke. A
# <name>-fix-N sub-change is a FLAT SIBLING of its parent under
# spectre/changes/ — `spectre new` refuses an id that is not a single flat
# directory name — so the glob matches the parent AND the sibling the moment
# a fix round opens one, and a wrapper refusing on "more than one" took this
# guard out of service on exactly the runs it was added for. A fix sibling is
# therefore not ambiguity: the highest-numbered one resolves, or the root when
# there is none. The mechanism, the digit test that keeps a merely
# similarly-named change out of it, and two caveats on the numbering are in
# the body below.
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
NAMES=()
if [ -d "$CHANGES_DIR" ]; then
  for tasks_file in "$CHANGES_DIR"/*/tasks.md; do
    [ -e "$tasks_file" ] || continue
    MATCHES+=("$tasks_file")
    change_dir="${tasks_file%/tasks.md}"
    NAMES+=("${change_dir##*/}")
  done
fi

if [ "${#MATCHES[@]}" -eq 0 ]; then
  echo "check-task-commit-fields.sh: no tasks.md found under $CHANGES_DIR" >&2
  exit 2
fi

# A <name>-fix-N SUB-CHANGE IS NOT AMBIGUITY. Under spectre a sub-change is a
# FLAT SIBLING of its parent under spectre/changes/ -- `spectre new` refuses an
# id that is not a single flat directory name -- so the glob above matches the
# parent AND every fix sibling as soon as one exists. Under OpenSpec a
# sub-change was nested and never matched, so "more than one tasks.md" meant
# two unrelated changes and refusing was right. Refusing now would take this
# guard out of service on exactly the runs it was added for: every fix round
# after the first sub-change is created.
#
# is_fix_sibling_of_set <name> -- true when <name> is "<stem>-fix-<digits>" AND
# <stem> is itself one of the matched change names. The digit test is the same
# one check-cleanup-complete.sh's sub-change row uses, and for the same reason:
# a change merely named like a neighbour (`demo-fix-the-parser`) is a change of
# its own, not this change's sub-change, and must still count as ambiguity.
is_fix_sibling_of_set() {
  local candidate="$1" suffix stem other
  case "$candidate" in *-fix-*) ;; *) return 1 ;; esac
  suffix="${candidate##*-fix-}"
  case "$suffix" in '' | *[!0-9]*) return 1 ;; esac
  stem="${candidate%-fix-$suffix}"
  for other in "${NAMES[@]}"; do
    [ "$other" = "$stem" ] && return 0
  done
  return 1
}

ROOTS=()
for change_name in "${NAMES[@]}"; do
  is_fix_sibling_of_set "$change_name" || ROOTS+=("$change_name")
done

# MORE THAN ONE ROOT IS STILL A REFUSAL, unchanged: two changes neither of
# which is the other's fix sibling is the genuine ambiguity this check has
# always existed to catch, and nothing here may guess between them.
if [ "${#ROOTS[@]}" -ne 1 ]; then
  echo "check-task-commit-fields.sh: more than one tasks.md found under $CHANGES_DIR, cannot resolve which change: ${MATCHES[*]}" >&2
  exit 2
fi

ROOT="${ROOTS[0]}"

# THE HIGHEST-NUMBERED FIX SIBLING WINS, and the root wins when there is none.
# A fix round creates <name>-fix-N and implements THAT plan; an earlier
# sub-change is finished, and the parent's own tasks were done before any fix
# round opened. So the newest sub-change is the plan whose tasks are being
# dispatched, which is the plan this guard has to read.
#
# TWO CAVEATS, both accepted. (1) N INCREMENTING IS CONVENTION, NOT CONTRACT:
# nothing in skills/myflow-do/SKILL.md specifies that a fix round numbers its
# sub-change one higher than the last, so "highest-numbered" reads an ordering
# nobody promised. (2) Reading the wrong plan is normally LOUD rather than
# silent — the commit's files are undeclared there and the guard exits 1
# naming one — but it can pass silently when the chosen plan's task N declares
# a SUPERSET of the intended plan's files. Passing the change name as an
# argument would remove both; that changes a call signature
# skills/myflow-do/SKILL.md documents, and was judged not worth it.
CHOSEN="$ROOT"
CHOSEN_N=-1
for change_name in "${NAMES[@]}"; do
  case "$change_name" in "$ROOT"-fix-*) ;; *) continue ;; esac
  suffix="${change_name##*-fix-}"
  case "$suffix" in '' | *[!0-9]*) continue ;; esac
  if [ "$((10#$suffix))" -gt "$CHOSEN_N" ]; then
    CHOSEN_N="$((10#$suffix))"
    CHOSEN="$change_name"
  fi
done

TASKS_MD="$CHANGES_DIR/$CHOSEN/tasks.md"

if [ -n "$PARENT_SHA" ]; then
  exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA" "$PARENT_SHA"
fi

exec python3 "$PYTHON_GUARD" "$TASKS_MD" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA"
