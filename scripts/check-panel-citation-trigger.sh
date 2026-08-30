#!/usr/bin/env bash
# check-panel-citation-trigger.sh — decide whether the review panel's
# citation pre-check should run, in a guard instead of prose.
#
# Usage: check-panel-citation-trigger.sh <worktree> <merge-base>
#
# Prints nothing; the exit code is the whole answer.
#   0  the change's own paths include at least one path ending .md or .mdc
#   1  none do
#   2  usage error: a missing argument, <worktree> not a directory or not a
#      git worktree, <merge-base> not resolving, or a git invocation failed
#
# THE CHANGE'S OWN PATHS INCLUDE THE INDEX AND THE WORKING TREE (design.md:
# touched-paths-include-index-and-worktree, the same rule check-base-moved.sh
# already applies): the union of what HEAD carries since the merge base,
# what is staged, and what is unstaged.
set -euo pipefail

# GIT_BIN (KAN-304 review round 6, F10-F13): resolved to an absolute path
# with `type -P`, which — unlike a bare `git` word — ignores any shell
# function or alias named `git` and performs a real PATH search. Every
# call below invokes "$GIT_BIN", never bare `git`: invoking a command by
# absolute path never triggers bash's function/alias lookup, so a `git()`
# function or `alias git=` defined ANYWHERE in this file (any spacing, any
# position on its line) cannot intercept these calls, closing F8/F11/F13's
# whole bypass class structurally rather than by pattern-matching the
# shapes that class can take. This line is also the reason F12's scoped
# PATH swap around one call no longer matters: GIT_BIN is captured here,
# once, before any later line in this file runs, so a PATH change further
# down is resolved against nothing — "$GIT_BIN" is already an absolute
# path by the time any later line could execute.
GIT_BIN="$(type -P git)" || {
  echo "check-panel-citation-trigger: no git binary found on PATH" >&2
  exit 2
}

WORKTREE="${1:-}"
MERGEBASE="${2:-}"

if [ -z "$WORKTREE" ] || [ -z "$MERGEBASE" ]; then
  echo "check-panel-citation-trigger: usage: check-panel-citation-trigger.sh <worktree> <merge-base>" >&2
  exit 2
fi

if [ ! -d "$WORKTREE" ]; then
  echo "check-panel-citation-trigger: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-panel-citation-trigger: $WORKTREE is not a git worktree — cannot determine anything" >&2
  exit 2
fi

if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --verify --end-of-options "${MERGEBASE}^{commit}" >/dev/null 2>&1; then
  echo "check-panel-citation-trigger: merge base '$MERGEBASE' does not resolve in $WORKTREE" >&2
  exit 2
fi

COMMITTED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --end-of-options "${MERGEBASE}..HEAD" 2>/dev/null)" || {
  echo "check-panel-citation-trigger: cannot list this change's committed paths in $WORKTREE" >&2
  exit 2
}
STAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --cached 2>/dev/null)" || {
  echo "check-panel-citation-trigger: cannot list this change's staged paths in $WORKTREE" >&2
  exit 2
}
UNSTAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only 2>/dev/null)" || {
  echo "check-panel-citation-trigger: cannot list this change's unstaged paths in $WORKTREE" >&2
  exit 2
}

MATCH="$(printf '%s\n%s\n%s\n' "$COMMITTED" "$STAGED" "$UNSTAGED" | awk 'NF && ($0 ~ /\.mdc?$/) { print; exit }')"

if [ -n "$MATCH" ]; then
  exit 0
fi
exit 1
