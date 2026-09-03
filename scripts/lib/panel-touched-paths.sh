# scripts/lib/panel-touched-paths.sh — this change's own touched paths,
# defined once.
#
# check-panel-docs-only.sh and check-panel-citation-trigger.sh both need the
# same answer to "which paths did this change touch?": the union of
# committed-since-merge-base, staged and unstaged paths, sorted and
# de-duplicated. Before this file existed the two guards carried a
# line-for-line copy of that preamble — GIT_BIN resolution through the
# COMMITTED/STAGED/UNSTAGED collection — differing only in the program-name
# string embedded in each error message (KAN-312 review round 0, F1
# Principles). One definition, sourced by both, is what stops that copy
# from drifting the way resolve_file's five copies did (see
# scripts/lib/resolve-file.sh) or check-unfinished-work.sh's and
# check-panel-reproducers.sh's shared total-count pattern did before it was
# folded into a library of its own.
#
# Every function here takes the CALLING GUARD'S OWN NAME as its first
# argument and uses it as the error-message prefix, so each guard's
# stderr wording stays byte-for-byte what it was before extraction —
# only the logic moved, never the user-facing text.
#
# Not meant to be executed directly — a caller sources it and calls its
# functions; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.

# panel_resolve_git <prog> -> prints the resolved git binary's absolute path
# to stdout. `type -P git`, not bare `git`: it ignores any shell function or
# alias named `git` and performs a real PATH search (KAN-304 review round 6,
# F8/F11/F13) — invoking by absolute path afterward means a `git()` function
# or `alias git=` defined anywhere cannot intercept the calls this preamble
# makes. Returns 2, with a message on stderr, when no git binary is on PATH.
panel_resolve_git() {
  local prog="$1" GIT_BIN
  GIT_BIN="$(type -P git)" || {
    echo "$prog: no git binary found on PATH" >&2
    return 2
  }
  printf '%s\n' "$GIT_BIN"
}

# panel_validate_worktree <prog> <worktree> <mergebase> <git_bin> -> exit 0
# when <worktree> is a directory, a git worktree, and <mergebase> resolves
# to a commit inside it. Returns 2, with a message on stderr naming which
# check failed, otherwise — including a missing <worktree> or <mergebase>
# argument.
#
# Parameters are bound to the SAME uppercase local names
# (WORKTREE/MERGEBASE/GIT_BIN) the two guards used before extraction —
# not a style choice, a call-site contract: test-check-panel-citation-trigger.sh's
# F8/F11/F13 and F2/F3/F9/F10 checks assert, by source-text grep and by
# `bash -x` line-anchored trace, that every real git invocation literally
# reads "$GIT_BIN" -C "$WORKTREE" and that the rev-parse/diff calls carry
# this run's own $MERGEBASE. Renaming these to lowercase would silently
# defeat those checks without breaking any functional case.
panel_validate_worktree() {
  local prog="$1" WORKTREE="$2" MERGEBASE="$3" GIT_BIN="$4"
  if [ -z "$WORKTREE" ] || [ -z "$MERGEBASE" ]; then
    echo "$prog: usage: $prog.sh <worktree> <merge-base>" >&2
    return 2
  fi
  if [ ! -d "$WORKTREE" ]; then
    echo "$prog: $WORKTREE is not a directory — cannot determine anything" >&2
    return 2
  fi
  if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
    echo "$prog: $WORKTREE is not a git worktree — cannot determine anything" >&2
    return 2
  fi
  if ! "$GIT_BIN" -C "$WORKTREE" rev-parse --verify --end-of-options "${MERGEBASE}^{commit}" >/dev/null 2>&1; then
    echo "$prog: merge base '$MERGEBASE' does not resolve in $WORKTREE" >&2
    return 2
  fi
}

# panel_touched_paths <prog> <worktree> <mergebase> <git_bin> -> prints this
# change's own touched paths to stdout, one per line, sorted and
# de-duplicated: the union of what HEAD carries since the merge base, what
# is staged, and what is unstaged (design.md:
# touched-paths-include-index-and-worktree, the same rule check-base-moved.sh
# already applies). Prints nothing when the union is empty. Returns 2, with
# a message on stderr naming which collection failed, when a `git diff`
# invocation fails. Same uppercase-local-name contract as
# panel_validate_worktree above — COMMITTED's assignment line and its
# "$GIT_BIN" -C "$WORKTREE" call site are what the citation-trigger
# harness's line-anchored trace locates.
panel_touched_paths() {
  local prog="$1" WORKTREE="$2" MERGEBASE="$3" GIT_BIN="$4"
  local COMMITTED STAGED UNSTAGED
  COMMITTED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --end-of-options "${MERGEBASE}..HEAD" 2>/dev/null)" || {
    echo "$prog: cannot list this change's committed paths in $WORKTREE" >&2
    return 2
  }
  STAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only --cached 2>/dev/null)" || {
    echo "$prog: cannot list this change's staged paths in $WORKTREE" >&2
    return 2
  }
  UNSTAGED="$("$GIT_BIN" -C "$WORKTREE" diff --name-only 2>/dev/null)" || {
    echo "$prog: cannot list this change's unstaged paths in $WORKTREE" >&2
    return 2
  }
  printf '%s\n%s\n%s\n' "$COMMITTED" "$STAGED" "$UNSTAGED" | awk 'NF' | sort -u
}
