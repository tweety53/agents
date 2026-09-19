#!/usr/bin/env bash
# plan-class.sh — the planner's mechanical classifier and rolls.
#
# Usage: plan-class.sh <tasks.md> <repos> [<worktree> <merge-base>]
#
# Prints exactly three lines to stdout:
#   inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no
#   class: micro|small|regular|big
#   rolls: compact N · experimental N · bundle N
#
# Exit 0 on a printed answer, exit 2 on a missing <tasks.md>, a
# non-integer <repos>, an argument count that is neither 2 nor 4, or — with
# the optional arguments — a <worktree>/<merge-base> that cannot be answered
# (not a directory, not a git worktree, an unresolving merge base, or a
# failed git invocation). Rules from design.md's "Inputs and the class" and
# "The rolls" (kan-472-flow-dynamic-review-panel-roster-repo-scoped); the
# small/big thresholds were raised by kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop
# so more changes classify small/regular and roll compact, then lowered by
# roughly 25% so a plan is classed up one step sooner:
#
#   tasks     = count of column-0 `- [ ] <n>.` / `- [x] <n>.` lines
#   files     = size of the union of every task's `**Files:**` backticked
#               paths
#   repos     = the <repos> argument, verbatim — the planner resolves the
#               worktree set, this script only classifies it
#   migration = any **Files:** path under stats/internal/store/migrations/
#               or ending .sql
#   spec      = any **Files:** path under spectre/specs/
#   red       = any task tagged `**Build:** red`
#   unverified= any `unverified:` provenance tag anywhere in the plan
#
#   small:   tasks<=8 and files<=19 and repos=1 and not migration and not spec
#   big:     tasks>=22 or files>=60 or (repos>1 and tasks>=11)
#            or (migration and tasks>=11)
#   regular: everything else
#
#   micro:   the small thresholds met with tasks<=2, every **Files:** path
#            documentation (.md/.mdc), no `**Build:** red` tag, and — when
#            <worktree> <merge-base> are passed — this change's own touched
#            paths (lib/panel-touched-paths.sh's union of
#            committed-since-merge-base, staged and unstaged) entirely
#            documentation and at most MICRO_LINE_CAP changed lines; an empty
#            touched surface passes vacuously. Without the optional arguments
#            micro never fires, so the two-argument form classifies exactly as
#            before it existed. Added by
#            kan-617-flow-cost-the-decide-record-ceremony-runs-full so a
#            two-line prose change stops paying the full decide ceremony
#            (brainstorm-planner.md's Decide collapses to a recorded default
#            decision on this class).
#
# The planner may raise this class one step with a recorded override — that
# judgment call lives in brainstorm-planner.md, never in this script, so
# `override` never appears in this script's own output (test-plan-class.sh
# asserts exactly that).
#
# Rolls are reproducible per change name (basename of the directory holding
# <tasks.md>): compact_roll = sha256("<name>") mod 100, experimental_roll =
# sha256("<name>exp") mod 100, bundle_roll = sha256("<name>bundle") mod 100,
# each over the first 8 hex digits of the digest read as an integer —
# scripts/lib/sha256-hex.sh's sha256_hex is reused rather than a second
# shasum wrapper (it already tries shasum, sha256sum, openssl in turn).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/sha256-hex.sh
source "$SCRIPT_DIR/lib/sha256-hex.sh"
# shellcheck source=lib/panel-touched-paths.sh
source "$SCRIPT_DIR/lib/panel-touched-paths.sh"

# The micro class's changed-line cap on the change's own touched surface.
MICRO_LINE_CAP=20

if [ "$#" -ne 2 ] && [ "$#" -ne 4 ]; then
  echo "usage: plan-class.sh <tasks.md> <repos> [<worktree> <merge-base>]" >&2
  exit 2
fi

TASKS_FILE="$1"
REPOS="$2"

if [ ! -f "$TASKS_FILE" ]; then
  echo "plan-class.sh: no such file: $TASKS_FILE" >&2
  exit 2
fi

case "$REPOS" in
  ''|*[!0-9]*)
    echo "plan-class.sh: <repos> must be a non-negative integer, got: $REPOS" >&2
    exit 2
    ;;
esac

TASKS_COUNT="$(grep -cE '^- \[[ xX]\] [0-9]+\.' "$TASKS_FILE" || true)"

# Every backticked path on a **Files:** line, deduplicated across the plan.
FILES_LIST="$(grep -E '^\*\*Files:\*\*' "$TASKS_FILE" \
  | grep -oE '`[^`]+`' \
  | tr -d '`' \
  | sort -u || true)"
if [ -n "$FILES_LIST" ]; then
  FILES_COUNT="$(printf '%s\n' "$FILES_LIST" | wc -l | tr -d ' ')"
else
  FILES_COUNT=0
fi

MIGRATION=no
if [ -n "$FILES_LIST" ] && printf '%s\n' "$FILES_LIST" | grep -qE '^stats/internal/store/migrations/|\.sql$'; then
  MIGRATION=yes
fi

SPEC=no
if [ -n "$FILES_LIST" ] && printf '%s\n' "$FILES_LIST" | grep -qE '^spectre/specs/'; then
  SPEC=yes
fi

RED=no
if grep -qE '^\*\*Build:\*\* red\b' "$TASKS_FILE"; then
  RED=yes
fi

UNVERIFIED=no
if grep -q 'unverified:' "$TASKS_FILE"; then
  UNVERIFIED=yes
fi

CLASS=regular
if [ "$TASKS_COUNT" -le 8 ] && [ "$FILES_COUNT" -le 19 ] && [ "$REPOS" -eq 1 ] \
  && [ "$MIGRATION" = no ] && [ "$SPEC" = no ]; then
  CLASS=small
elif [ "$TASKS_COUNT" -ge 22 ] || [ "$FILES_COUNT" -ge 60 ] \
  || { [ "$REPOS" -gt 1 ] && [ "$TASKS_COUNT" -ge 11 ]; } \
  || { [ "$MIGRATION" = yes ] && [ "$TASKS_COUNT" -ge 11 ]; }; then
  CLASS=big
fi

# The change's own diff side: with the optional arguments it is answered on
# every plan, micro-eligible or not — an unanswerable <worktree>/<merge-base>
# is exit 2 unconditionally, exactly as the header states. One numstat of the
# merge base against the working tree covers committed, staged and unstaged
# together, so churn on one file counts once. A numstat entry git cannot count
# (binary) is fail-closed: never micro.
SURFACE_PATHS=""
SURFACE_LINES=""
if [ "$#" -eq 4 ]; then
  WORKTREE_ARG="$3"
  MERGEBASE_ARG="$4"
  GIT_BIN="$(panel_resolve_git "plan-class")" || exit 2
  panel_validate_worktree "plan-class" "$WORKTREE_ARG" "$MERGEBASE_ARG" "$GIT_BIN" || exit 2
  SURFACE_PATHS="$(panel_touched_paths "plan-class" "$WORKTREE_ARG" "$MERGEBASE_ARG" "$GIT_BIN")" || exit 2
  SURFACE_N="$("$GIT_BIN" -C "$WORKTREE_ARG" diff --no-renames --numstat --end-of-options "${MERGEBASE_ARG}")" || exit 2
  SURFACE_LINES="$(printf '%s\n' "$SURFACE_N" \
    | awk '$1 == "-" || $2 == "-" { bad = 1 } { s += $1 + $2 } END { if (bad) print 999999; else print s + 0 }')"
fi

# The micro class: the plan side is decidable from the tasks.md alone; the
# diff side above must have been answered too. An empty touched surface (a
# creating run, nothing implemented yet) passes vacuously; a non-empty one
# must itself be docs-only and within the cap.
if [ -n "$SURFACE_LINES" ] && [ "$CLASS" = small ] && [ "$TASKS_COUNT" -le 2 ] \
  && [ "$RED" = no ] && [ -n "$FILES_LIST" ] \
  && ! printf '%s\n' "$FILES_LIST" | grep -qvE '\.mdc?$'; then
  SURFACE_NONDOC="$(printf '%s\n' "$SURFACE_PATHS" | awk '$0 !~ /\.mdc?$/ { print; exit }')"
  if [ -z "$SURFACE_NONDOC" ] && [ "$SURFACE_LINES" -le "$MICRO_LINE_CAP" ]; then
    CLASS=micro
  fi
fi

CHANGE_NAME="$(basename "$(dirname "$TASKS_FILE")")"
COMPACT_HEX="$(sha256_hex "$CHANGE_NAME" | cut -c1-8)"
EXPERIMENTAL_HEX="$(sha256_hex "${CHANGE_NAME}exp" | cut -c1-8)"
BUNDLE_HEX="$(sha256_hex "${CHANGE_NAME}bundle" | cut -c1-8)"
COMPACT_ROLL=$(( 16#$COMPACT_HEX % 100 ))
EXPERIMENTAL_ROLL=$(( 16#$EXPERIMENTAL_HEX % 100 ))
BUNDLE_ROLL=$(( 16#$BUNDLE_HEX % 100 ))

printf 'inputs: tasks=%s files=%s repos=%s migration=%s spec=%s red=%s unverified=%s\n' \
  "$TASKS_COUNT" "$FILES_COUNT" "$REPOS" "$MIGRATION" "$SPEC" "$RED" "$UNVERIFIED"
printf 'class: %s\n' "$CLASS"
printf 'rolls: compact %s · experimental %s · bundle %s\n' "$COMPACT_ROLL" "$EXPERIMENTAL_ROLL" "$BUNDLE_ROLL"
