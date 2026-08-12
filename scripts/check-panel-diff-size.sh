#!/usr/bin/env bash
# check-panel-diff-size.sh — report whether a branch's diff is within the
# review panel's reading cap.
#
# Usage: check-panel-diff-size.sh <worktree> <merge-base> [cap]
#
# Prints two lines to stdout: the changed-line count, then the verdict
# ("under cap" or "over cap") — so a caller can quote either line rather
# than re-derive the number.
#
# Exit codes: 0 at or under the cap; 1 over the cap; 2 cannot answer at all.
#
# WHY THE CAP IS 2000. KAN-109's own measurement found a seven-slot review
# panel costing 948k tokens to read a 9877-line diff. A cap at roughly a
# fifth of that line count keeps a full panel pass in the low hundreds of
# thousands of tokens while leaving ordinary myflow changes — almost all of
# which fall well under it — untouched.
#
# WHY BOTH COMMITTED WORK AND THE WORKING TREE COUNT. `final-review.diff`,
# the text the panel actually reads, carries everything that changed since
# the merge base: commits already made plus whatever is still staged or
# unstaged in the working tree. Measuring only `merge-base..HEAD` would
# report zero for a run whose work sits entirely uncommitted, understating
# the exact diff the panel is about to be handed.
set -euo pipefail

WORKTREE="${1:-}"
MERGE_POINT="${2:-}"
CAP="${3:-2000}"

if [ -z "$WORKTREE" ]; then
  echo "check-panel-diff-size: usage: check-panel-diff-size.sh <worktree> <merge-base> [cap]" >&2
  exit 2
fi

if [ ! -d "$WORKTREE" ]; then
  echo "check-panel-diff-size: $WORKTREE is not a directory" >&2
  exit 2
fi

if ! git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-panel-diff-size: $WORKTREE is not a directory containing a git repository" >&2
  exit 2
fi

if [ -z "$MERGE_POINT" ]; then
  echo "check-panel-diff-size: merge base argument is required" >&2
  exit 2
fi

# Every ref this script did not choose itself arrives from a caller, so a
# value beginning with `-` is read as a ref and rejected rather than parsed
# as a git option.
if ! git -C "$WORKTREE" rev-parse --verify --end-of-options "$MERGE_POINT" >/dev/null 2>&1; then
  echo "check-panel-diff-size: merge base does not resolve in $WORKTREE: '$MERGE_POINT'" >&2
  exit 2
fi

case "$CAP" in
  '' | *[!0-9]*)
    echo "check-panel-diff-size: cap is not a non-negative decimal integer: '$CAP'" >&2
    exit 2
    ;;
esac

# `git diff <commit>` (no second ref) compares the named commit against the
# current working tree, which already folds in staged and unstaged changes
# alongside every commit made since — one shortstat covers committed work
# and the working tree together.
SHORTSTAT="$(git -C "$WORKTREE" diff --shortstat --end-of-options "$MERGE_POINT")"

INSERTIONS=0
DELETIONS=0
if [[ "$SHORTSTAT" =~ ([0-9]+)\ insertion ]]; then
  INSERTIONS="${BASH_REMATCH[1]}"
fi
if [[ "$SHORTSTAT" =~ ([0-9]+)\ deletion ]]; then
  DELETIONS="${BASH_REMATCH[1]}"
fi

TOTAL=$((INSERTIONS + DELETIONS))

echo "$TOTAL"

if [ "$TOTAL" -le "$CAP" ]; then
  echo "under cap"
  exit 0
else
  echo "over cap"
  exit 1
fi
