#!/usr/bin/env bash
# check-archive-scope.sh — refuse an archive commit whose staged diff reaches
# outside the path prefixes that commit is allowed to touch, in a guard
# instead of trusting `git add -A` to have only picked up the archive move.
#
# Usage: check-archive-scope.sh <worktree> <allowed-prefix> [<allowed-prefix> ...]
#
# Prints one OUT-OF-SCOPE line per offending staged path, then ONE verdict
# line:
#   SCOPE-OK: <worktree>                    every staged path matches a prefix
#   SCOPE-VIOLATION: <worktree> — <n>       <n> staged path(s) do not
#
# Exit 0 on SCOPE-OK (including nothing staged at all — an empty diff is
# never a violation), 1 on SCOPE-VIOLATION, 2 with NOTHING on stdout when
# <worktree> is not a readable directory, is not a git worktree, or no
# <allowed-prefix> was given — an inability to answer is never reported as a
# verdict.
#
# A PREFIX MATCH IS A PATH-COMPONENT MATCH, NOT A BARE STRING PREFIX. Each
# <allowed-prefix> is normalized to end in exactly one `/` before matching,
# so `spectre/changes/` never lets `spectre/changes-backup/x` through — the
# same distinction check-worktree-location.sh's own header draws for
# worktree roots.
#
# WHY THIS GUARD EXISTS (KAN-472 follow-up, self-review finding). The finish
# contract's own step 4 stages the archive commit with `git -C
# <landing-worktree> add -A` — every path in the worktree, not only the
# archived change's own move. A landing worktree left stale by a skipped or
# failed fast-forward in step 2 stages and commits whatever else is sitting
# in that stale tree right alongside the archive move, silently. This is
# exactly what happened to kan-474's archive commit (`a574bf4`): it reverted
# skill-file content another change had shipped minutes earlier, because
# `add -A` swept it up along with the intended archive move. This guard
# makes that class of accident a refusal instead of a silent commit.
set -euo pipefail

SELF="check-archive-scope"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -ge 2 ] || die "usage: check-archive-scope.sh <worktree> <allowed-prefix> [<allowed-prefix> ...]"

WORKTREE_ARG="$1"
shift
PREFIXES=("$@")

ROOT="$(cd "$WORKTREE_ARG" 2>/dev/null && pwd -P)" || die "$WORKTREE_ARG is not a directory"
git -C "$ROOT" rev-parse --is-inside-work-tree >/dev/null 2>&1 \
  || die "$ROOT is not a git worktree"

STAGED="$(git -C "$ROOT" diff --cached --name-only 2>/dev/null)" \
  || die "cannot read the staged diff of $ROOT"

OFFENDERS=()
while IFS= read -r path; do
  [ -n "$path" ] || continue
  matched=0
  for prefix in "${PREFIXES[@]}"; do
    prefix="${prefix%/}/"
    case "$path" in
      "$prefix"*) matched=1; break ;;
    esac
  done
  [ "$matched" -eq 1 ] || OFFENDERS+=("$path")
done <<< "$STAGED"

if [ "${#OFFENDERS[@]}" -gt 0 ]; then
  for path in "${OFFENDERS[@]}"; do
    printf 'OUT-OF-SCOPE: %s\n' "$path"
  done
  printf 'SCOPE-VIOLATION: %s — %s\n' "$ROOT" "${#OFFENDERS[@]}"
  exit 1
fi

printf 'SCOPE-OK: %s\n' "$ROOT"
exit 0
