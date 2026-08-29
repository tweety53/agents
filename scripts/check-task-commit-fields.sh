#!/usr/bin/env bash
# check-task-commit-fields.sh — thin wrapper.
#
# All field-parsing and diff-checking logic lives in
# check-task-commit-fields.py (Python 3, standard library only), following
# the same split check-task-build-green.sh uses and for the same reason:
# this file exists only so an operator's muscle memory invoking this exact
# filename, and .flow/project.md's declared commands, keep working, while
# the field grammar underneath gets a real parser.
#
# Unlike check-task-build-green.sh (which resolves WHICH tasks.md files to
# scan, zero or more of them, with no per-task identity), this guard checks
# ONE task's fields against ONE real commit, so its calling convention names
# the task and commit explicitly rather than scanning:
#
#   check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha] [canonical-worktree] [change-name]
#
# This wrapper's own job is resolving WHICH tasks.md the named task lives
# in, among the non-archived ones under
# <worktree>/spectre/changes/*/tasks.md. Zero matches there is not
# automatically a refusal any more (KAN-363 task 9) — see WHERE A LINK IS
# FOLLOWED below — and more than one ROOT change is still a refusal: two
# changes neither of which is the other's sub-change, which nothing here may
# guess between. The optional sixth argument, <change-name>, skips that
# ambiguity scan entirely (KAN-367): when given, this wrapper resolves
# directly against <worktree>/spectre/changes/<change-name> — still honoring
# its own -fix-N sibling and its own link.md, but never looking at any other
# directory under spectre/changes/ — and takes priority over the glob below.
#
# WHERE A LINK IS FOLLOWED. A satellite change directory (spectre task 1's
# `link.md`, carrying `## Part of`) has no `tasks.md` of its own by design,
# so the glob above finds nothing at all under a satellite worktree — the
# exact shape that used to be an unconditional "no tasks.md found" refusal.
# When that glob comes back empty, and exactly one LINK-ONLY change
# directory exists under `<worktree>/<spec-root>/changes/` — one whose own
# `link.md` exists and whose own `tasks.md` does not — this wrapper resolves
# that satellite's plan through `scripts/lib/change-plan.sh` (task 7),
# passing the optional fifth argument straight through as its
# canonical-worktree, before giving up. More than one link-only directory,
# or a resolution that fails, still ends in the same "no tasks.md found"
# refusal as before.
#
# A LINK-ONLY DIRECTORY IS NEVER COUNTED TOWARD THE ROOT-CHANGE AMBIGUITY
# TEST below either, for the same reason a `<name>-fix-N` sibling is not: a
# satellite carries no `tasks.md`, so the `*/tasks.md` glob that test is
# built from never sees it — a satellite sitting beside a genuine root
# change changes nothing about which root that test resolves to.
#
# SYMLINKS. `-f` and `-d`, used throughout this file and in
# scripts/lib/change-plan.sh, follow symlinks — so a symlink at
# `<worktree>/<spec-root>/changes/<allowlisted-name>` is followed and its
# content read as that change's plan. check-unfinished-work.sh's own header
# accepts the identical tradeoff for the identical reason; this is not a
# hole this change opens.
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
# Named and checked for the same two reasons as the grammar module above:
# a `source` is as invisible to check-guard-symlinks.sh rule 2 as an
# import is unless the path appears as $SCRIPT_DIR/<name>, and a missing
# sibling should say which one rather than surface as a bash error.
SPEC_ROOT_LIB="$SCRIPT_DIR/lib/spec-root.sh"
# change_plan_path (KAN-363 task 7) is what WHERE A LINK IS FOLLOWED above
# resolves a satellite's plan through. Named and checked for the same two
# reasons as the grammar module and spec-root module above.
CHANGE_PLAN_LIB="$SCRIPT_DIR/lib/change-plan.sh"

if [ ! -f "$GRAMMAR_MODULE" ]; then
  echo "check-task-commit-fields.sh: shared grammar module not found: $GRAMMAR_MODULE" >&2
  exit 2
fi

if [ ! -f "$SPEC_ROOT_LIB" ]; then
  echo "check-task-commit-fields.sh: shared spec-root module not found: $SPEC_ROOT_LIB" >&2
  exit 2
fi
source "$SPEC_ROOT_LIB"

if [ ! -f "$CHANGE_PLAN_LIB" ]; then
  echo "check-task-commit-fields.sh: shared change-plan module not found: $CHANGE_PLAN_LIB" >&2
  exit 2
fi
source "$CHANGE_PLAN_LIB"

command -v python3 >/dev/null 2>&1 || {
  echo "check-task-commit-fields.sh: python3 not found on PATH — cannot run the guard" >&2
  exit 2
}

if ! python3 -c 'import sys; sys.exit(0)'; then
  echo "check-task-commit-fields.sh: python3 is present but failed to run a trivial program (see above) — cannot run the guard" >&2
  exit 2
fi

if [ "$#" -lt 3 ] || [ "$#" -gt 6 ]; then
  echo "usage: check-task-commit-fields.sh <worktree> <task-id> <commit-sha> [parent-sha] [canonical-worktree] [change-name]" >&2
  exit 2
fi

WORKTREE="$1"
TASK_ID="$2"
COMMIT_SHA="$3"
PARENT_SHA="${4:-}"
CANONICAL_WORKTREE="${5:-}"
CHANGE_NAME="${6:-}"

if [ ! -d "$WORKTREE" ]; then
  echo "check-task-commit-fields.sh: worktree not found: $WORKTREE" >&2
  exit 2
fi

CHANGES_DIR="$WORKTREE/$(spec_root_leaf "$WORKTREE")/changes"

# highest_fix_sibling <changes-dir> <root> <candidate...> — selects the
# highest-numbered "<root>-fix-N" candidate whose own tasks.md exists,
# echoing <root> unchanged when none qualify. Shared by the named-change
# resolution block below and the ambiguity-scan path at the bottom of this
# file, so the fix-sibling-selection rule (same digit validation, same
# $((10#$suffix)) numeric comparison, same tasks.md existence requirement,
# same root fallback) lives in exactly one place.
highest_fix_sibling() {
  local changes_dir="$1" root="$2"
  shift 2
  local chosen="$root" chosen_n=-1 candidate suffix
  for candidate in "$@"; do
    case "$candidate" in "$root"-fix-*) ;; *) continue ;; esac
    suffix="${candidate##*-fix-}"
    case "$suffix" in '' | *[!0-9]*) continue ;; esac
    [ -f "$changes_dir/$candidate/tasks.md" ] || continue
    if [ "$((10#$suffix))" -gt "$chosen_n" ]; then
      chosen_n="$((10#$suffix))"
      chosen="$candidate"
    fi
  done
  printf '%s\n' "$chosen"
}

# dispatch_python_guard <tasks-md> — execs the Python guard against
# <tasks-md>, forwarding $PARENT_SHA when the caller set one. Shared by all
# three resolution paths (named change, satellite link, glob-path) so the
# exec dispatch lives in exactly one place; the function itself execs, so
# there is no return to the caller either way — same early-return shape the
# three separate copies had.
dispatch_python_guard() {
  local tasks_md="$1"
  if [ -n "$PARENT_SHA" ]; then
    exec python3 "$PYTHON_GUARD" "$tasks_md" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA" "$PARENT_SHA"
  fi
  exec python3 "$PYTHON_GUARD" "$tasks_md" "$TASK_ID" "$WORKTREE" "$COMMIT_SHA"
}

if [ -n "$CHANGE_NAME" ]; then
  case "$CHANGE_NAME" in
    *[!A-Za-z0-9._-]* | . | .. | */*)
      echo "check-task-commit-fields.sh: invalid change name: $CHANGE_NAME" >&2
      exit 2
      ;;
  esac

  ROOT_DIR="$CHANGES_DIR/$CHANGE_NAME"
  TASKS_MD=""

  if [ -f "$ROOT_DIR/tasks.md" ]; then
    # Scoped version of the existing highest-numbered-fix-sibling rule (see
    # the unchanged glob path below): look only at $CHANGE_NAME's own
    # -fix-N family, never at any other directory under $CHANGES_DIR.
    FIX_CANDIDATES=()
    for change_dir in "$CHANGES_DIR/$CHANGE_NAME"-fix-*/; do
      [ -d "$change_dir" ] || continue
      cname="${change_dir%/}"
      cname="${cname##*/}"
      FIX_CANDIDATES+=("$cname")
    done
    CHOSEN="$(highest_fix_sibling "$CHANGES_DIR" "$CHANGE_NAME" "${FIX_CANDIDATES[@]+"${FIX_CANDIDATES[@]}"}")"
    TASKS_MD="$CHANGES_DIR/$CHOSEN/tasks.md"
  elif [ -f "$ROOT_DIR/link.md" ]; then
    TASKS_MD="$(change_plan_path "$WORKTREE" "$CHANGE_NAME" "$CANONICAL_WORKTREE" 2>/dev/null || true)"
  fi

  if [ -z "$TASKS_MD" ] || [ ! -f "$TASKS_MD" ]; then
    echo "check-task-commit-fields.sh: no tasks.md found for change '$CHANGE_NAME' under $CHANGES_DIR" >&2
    exit 2
  fi

  dispatch_python_guard "$TASKS_MD"
fi

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
  # WHERE A LINK IS FOLLOWED (see header): find the link-only change
  # directories — link.md present, tasks.md absent — under CHANGES_DIR.
  # Exactly one is a satellite worktree; anything else (none, or more than
  # one) cannot be resolved without guessing and falls through to the same
  # refusal a plain "no tasks.md at all" worktree always got.
  SATELLITES=()
  if [ -d "$CHANGES_DIR" ]; then
    for change_dir in "$CHANGES_DIR"/*/; do
      [ -d "$change_dir" ] || continue
      cname="${change_dir%/}"
      cname="${cname##*/}"
      [ -f "$CHANGES_DIR/$cname/link.md" ] || continue
      [ -f "$CHANGES_DIR/$cname/tasks.md" ] && continue
      SATELLITES+=("$cname")
    done
  fi

  TASKS_MD=""
  if [ "${#SATELLITES[@]}" -eq 1 ]; then
    TASKS_MD="$(change_plan_path "$WORKTREE" "${SATELLITES[0]}" "$CANONICAL_WORKTREE" 2>/dev/null || true)"
  fi

  if [ -z "$TASKS_MD" ] || [ ! -f "$TASKS_MD" ]; then
    echo "check-task-commit-fields.sh: no tasks.md found under $CHANGES_DIR" >&2
    exit 2
  fi

  dispatch_python_guard "$TASKS_MD"
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
CHOSEN="$(highest_fix_sibling "$CHANGES_DIR" "$ROOT" "${NAMES[@]+"${NAMES[@]}"}")"

TASKS_MD="$CHANGES_DIR/$CHOSEN/tasks.md"

dispatch_python_guard "$TASKS_MD"
