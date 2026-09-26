#!/usr/bin/env bash
# guard-autosquash.sh — mechanical ancestry guard for the fixup-to-autosquash
# sequence. A fixup against a sha that resolves but is no ancestor of the
# branch — a stale duplicate object, for instance — hands
# `git rebase --autosquash` a spurious merge base and replays unrelated
# history into the branch; a rebase can likewise strand shas a plan file
# still records. This guard asserts the ancestry the sequence silently
# assumes, before and after.
#
# Subcommands:
#   targets <worktree> <sha> [<sha>...]
#       Run before any autosquash: every fixup target must resolve as a
#       commit object and be an ancestor of the worktree's HEAD.
#   after <worktree> <base-ref> <tasks-md>
#       Run after an autosquash or any rebase: <base-ref> must still be an
#       ancestor of HEAD, and every run of 7 or more lowercase hex characters
#       in <tasks-md> must resolve as a commit object AND be reachable from
#       HEAD. Existence alone passes a dangling sha — the stale
#       recorded-sha shape — so the two are separate checks. There is no
#       length ceiling: a sha glued into a longer hex run fails resolution
#       and is reported, never silently skipped. Lowercase only: git object
#       names are lowercase, so uppercase hex words in prose (DEADBEEF and
#       its like) never trip the sweep.
#
# Exit codes:
#   0  every check passed
#   1  one or more violations, one line each on stderr, each naming the sha
#   2  cannot answer — bad arguments, <worktree> not a git repository,
#      HEAD unresolvable, or <tasks-md> missing, not a regular file, or
#      unreadable

set -uo pipefail

PROG=guard-autosquash

usage() {
  printf 'usage: %s targets <worktree> <sha> [<sha>...]\n' "$PROG" >&2
  printf '       %s after <worktree> <base-ref> <tasks-md>\n' "$PROG" >&2
}

cannot_answer() { printf '%s: %s\n' "$PROG" "$1" >&2; exit 2; }
violation() { printf '%s: %s\n' "$PROG" "$1" >&2; VIOLATIONS=$((VIOLATIONS + 1)); }

[ $# -ge 1 ] || { usage; exit 2; }
SUB=$1
shift

case $SUB in
  targets) [ $# -ge 2 ] || { usage; exit 2; } ;;
  after) [ $# -eq 3 ] || { usage; exit 2; } ;;
  *) usage; exit 2 ;;
esac

WORKTREE=$1
shift

git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1 ||
  cannot_answer "$WORKTREE is not a git repository"
git -C "$WORKTREE" rev-parse --verify -q HEAD >/dev/null ||
  cannot_answer "$WORKTREE has no resolvable HEAD"

VIOLATIONS=0

if [ "$SUB" = targets ]; then
  for sha in "$@"; do
    if ! git -C "$WORKTREE" cat-file -e "$sha^{commit}" 2>/dev/null; then
      violation "fixup target $sha does not resolve as a commit object — refusing the autosquash"
      continue
    fi
    git -C "$WORKTREE" merge-base --is-ancestor "$sha" HEAD >/dev/null 2>&1 ||
      violation "fixup target $sha is not an ancestor of HEAD — refusing the autosquash"
  done
else
  BASE=$1
  TASKS_MD=$2
  [ -f "$TASKS_MD" ] || cannot_answer "tasks file $TASKS_MD is missing, not a regular file, or unreadable"
  git -C "$WORKTREE" merge-base --is-ancestor "$BASE" HEAD >/dev/null 2>&1 ||
    violation "base $BASE is no longer an ancestor of HEAD"
  # Maximal runs of lowercase hex characters, 7 or more long — the span an
  # abbreviated-to-full git sha occupies, with no upper bound: a sha glued
  # into a longer hex run fails resolution and is reported, never silently
  # skipped. Splitting on the complement is portable where word-boundary
  # regexes differ between BSD and GNU grep.
  TOKENS="$(tr -c '0-9a-f' '\n' <"$TASKS_MD" | awk 'length($0) >= 7' | sort -u)"
  for token in $TOKENS; do
    if ! git -C "$WORKTREE" cat-file -e "$token^{commit}" 2>/dev/null; then
      violation "$token does not resolve as a commit object"
      continue
    fi
    git -C "$WORKTREE" merge-base --is-ancestor "$token" HEAD >/dev/null 2>&1 ||
      violation "recorded sha $token is not reachable from HEAD — stale after a rewrite?"
  done
fi

if [ "$VIOLATIONS" -gt 0 ]; then
  printf '%s: %d violation(s)\n' "$PROG" "$VIOLATIONS" >&2
  exit 1
fi
printf '%s: %s ok\n' "$PROG" "$SUB"
