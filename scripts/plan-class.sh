#!/usr/bin/env bash
# plan-class.sh — the planner's mechanical classifier and rolls.
#
# Usage: plan-class.sh <tasks.md> <repos>
#
# Prints exactly three lines to stdout:
#   inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no
#   class: small|regular|big
#   rolls: compact N · experimental N · bundle N
#
# Exit 0 on a printed answer, exit 2 on a missing <tasks.md> or a
# non-integer <repos>. Rules from design.md's "Inputs and the class" and
# "The rolls" (kan-472-flow-dynamic-review-panel-roster-repo-scoped):
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
#   small:   tasks<=5 and files<=12 and repos=1 and not migration and not spec
#   big:     tasks>=15 or files>=40 or (repos>1 and tasks>=8)
#            or (migration and tasks>=8)
#   regular: everything else
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

if [ "$#" -ne 2 ]; then
  echo "usage: plan-class.sh <tasks.md> <repos>" >&2
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
if [ "$TASKS_COUNT" -le 5 ] && [ "$FILES_COUNT" -le 12 ] && [ "$REPOS" -eq 1 ] \
  && [ "$MIGRATION" = no ] && [ "$SPEC" = no ]; then
  CLASS=small
elif [ "$TASKS_COUNT" -ge 15 ] || [ "$FILES_COUNT" -ge 40 ] \
  || { [ "$REPOS" -gt 1 ] && [ "$TASKS_COUNT" -ge 8 ]; } \
  || { [ "$MIGRATION" = yes ] && [ "$TASKS_COUNT" -ge 8 ]; }; then
  CLASS=big
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
