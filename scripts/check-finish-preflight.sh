#!/usr/bin/env bash
# check-finish-preflight.sh — decide whether /myflow-finish should integrate
# (run 1) or archive (run 2), or refuse because it cannot tell.
#
# Usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->
#
# Prints ONE verdict line to stdout:
#   RUN1: <reason>    integrate — the branch has not reached the base branch
#   RUN2: <reason>    archive   — the branch is merged and nothing is outstanding
#   REFUSE: <reason>  stop and ask the operator
#
# Exit 0 whenever a verdict was reached; exit 2 when the tree cannot be read.
# The VERDICT carries the answer, not the exit status: an exit-code-only
# protocol makes "cannot determine" indistinguishable from "violation", which
# is the exact confusion this script exists to remove. Exit 2 keeps the
# meaning this repository's other guards give it.
#
# WHY SIGNAL ORDER IS THE WHOLE FIX. A branch with no commits of its own is an
# ancestor of EVERY branch, so `merge-base --is-ancestor` answers "merged" on a
# branch whose work is staged and never committed — the normal IN_PROGRESS
# state, since /myflow-do may not commit before a PR exists. Run 2 would then
# archive the change and `git worktree remove --force` the worktree holding all
# of it. Comparing HEAD with the merge base RECORDED IN THE STATE FILE catches
# it; counting commits ahead of the base branch does NOT, because a genuinely
# merged branch is also zero ahead once its commit joins the base branch.
#
# Base-branch resolution deliberately stays in pipeline.md's Finish contract
# and is passed in. That resolution already carries a hard-won guard against
# HEAD@{upstream}; a second copy here could drift from it.
set -euo pipefail

WORKTREE="${1:-}"
BASE_REF="${2:-}"
RECORDED="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$BASE_REF" ] || [ -z "$RECORDED" ]; then
  echo "usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->" >&2
  exit 2
fi

if [ ! -d "$WORKTREE" ]; then
  echo "check-finish-preflight: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

if ! git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-finish-preflight: $WORKTREE is not a git worktree — cannot determine anything" >&2
  exit 2
fi

# (a) No recorded merge base: an honest unknown, never an inferred verdict.
if [ "$RECORDED" = "-" ]; then
  echo "REFUSE: no merge base recorded for $WORKTREE — cannot tell an unmerged branch from a merged one"
  exit 0
fi

# Every ref this script did not choose itself arrives from a state file and is
# passed after --end-of-options, so a value beginning with `-` is read as a ref
# and rejected rather than parsed as a git option.
HEAD_SHA="$(git -C "$WORKTREE" rev-parse --verify --end-of-options "HEAD^{commit}" 2>/dev/null)" || {
  echo "check-finish-preflight: cannot resolve HEAD in $WORKTREE" >&2
  exit 2
}

RECORDED_SHA="$(git -C "$WORKTREE" rev-parse --verify --end-of-options "${RECORDED}^{commit}" 2>/dev/null)" || {
  echo "REFUSE: recorded merge base '$RECORDED' does not resolve in $WORKTREE"
  exit 0
}

# (b) HEAD is still the merge base: the branch has no commits of its own.
#     This MUST be tested before the ancestor test. It reads only HEAD and the
#     recorded merge base, so it is deliberately answered BEFORE the base ref is
#     resolved below — an unresolvable base ref must not hide this shape, which
#     is the one this script exists to catch.
if [ "$HEAD_SHA" = "$RECORDED_SHA" ]; then
  echo "RUN1: HEAD is still the recorded merge base — the branch has no commits of its own"
  exit 0
fi

# A base ref that does not resolve would make the ancestor test below fail for
# an environmental reason and read as "not merged" — a RUN1 verdict reached by
# accident. Resolve it first so the failure is named instead.
if ! git -C "$WORKTREE" rev-parse --verify --end-of-options "${BASE_REF}^{commit}" >/dev/null 2>&1; then
  echo "REFUSE: base ref '$BASE_REF' does not resolve in $WORKTREE — cannot test whether HEAD reached it"
  exit 0
fi

# (c) The ancestor test. Only exit 1 means "not an ancestor"; anything else is
#     git failing to answer, which must not be read as a RUN1 verdict. stderr is
#     redirected so no git chatter can prefix the verdict line for a caller that
#     merges the two streams.
set +e
git -C "$WORKTREE" merge-base --is-ancestor --end-of-options "$HEAD_SHA" "$BASE_REF" 2>/dev/null
ANCESTOR_RC=$?
set -e
if [ "$ANCESTOR_RC" -eq 1 ]; then
  echo "RUN1: HEAD is not an ancestor of $BASE_REF — not merged"
  exit 0
fi
if [ "$ANCESTOR_RC" -ne 0 ]; then
  echo "check-finish-preflight: merge-base failed in $WORKTREE (exit $ANCESTOR_RC) — cannot determine anything" >&2
  exit 2
fi

# (d) Merged by ancestry, so nothing should be outstanding. Tracked changes
#     and untracked-unignored files both count; ignored files do not, because
#     they are disclosed separately at removal time rather than gating.
#     The status output is captured on its own line rather than piped straight
#     into `wc`: in a pipeline, `wc` and `tr` succeeding make the pipeline's
#     status theirs, so a git that failed (a held index lock, a transient I/O
#     error, a worktree that became unreadable mid-run) would abort the script
#     under `set -o pipefail` with no verdict, no message and an exit code
#     outside this script's contract. A silent `0` would be worse still: it
#     falls through to RUN2, the destructive verdict. An unreadable status is
#     the same class of failure as an unreadable worktree — exit 2, named.
STATUS_OUT="$(git -C "$WORKTREE" status --porcelain --untracked-files=normal 2>/dev/null)" || {
  echo "check-finish-preflight: cannot read the worktree status in $WORKTREE — cannot determine anything" >&2
  exit 2
}
if [ -z "$STATUS_OUT" ]; then
  DIRTY=0
else
  DIRTY="$(printf '%s\n' "$STATUS_OUT" | wc -l | tr -d ' ')"
fi
if [ "$DIRTY" != "0" ]; then
  echo "REFUSE: $BASE_REF contains HEAD, but $WORKTREE has $DIRTY uncommitted entries — a merged change should have nothing left to commit"
  exit 0
fi

echo "RUN2: HEAD is an ancestor of $BASE_REF, differs from the recorded merge base, and the worktree is clean"
exit 0
