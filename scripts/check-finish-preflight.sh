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
# THE MAIN-CHECKOUT ASSERTION (KAN-462 §4, design.md:
# main-checkout-is-asserted-not-moved). A would-be RUN2 verdict is REFUSEd
# instead when the main checkout — resolved from <worktree> via `git
# rev-parse --git-common-dir`, never taken as an argument — is not fit to
# host the landing worktree that run 2's own prepare-archive-branch.sh derives
# from it afterwards: `REFUSE: main checkout <path> is on <branch>, not
# <base>` when its current branch differs from <base-ref> with any `origin/`
# prefix stripped; `REFUSE: main checkout <path> has tracked changes` when
# `status --porcelain --untracked-files=no` is non-empty; `REFUSE: stray
# worktree(s): …` when `check-worktree-location.sh <main-checkout>` exits 1,
# relaying its STRAY lines. This check runs last, immediately before the
# RUN2 line it can still turn into a REFUSE — it never affects a RUN1 or an
# earlier REFUSE, since only run 2 ever touches the main checkout at all.
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
# Base-branch resolution is owned by resolve-base-branch.sh, not by this
# script — the caller resolves it and passes it in. That guard already
# carries a hard-won assertion against HEAD@{upstream}; a second copy here
# could drift from it.
#
# What THIS script owns (KAN-88, design.md:
# preflight-resolves-remote-tracking): the base ref it was HANDED may name a
# local branch that has fallen behind its remote-tracking counterpart — the
# caller composes `origin/$BASE`, but a caller that regresses to a bare name
# must not silently feed a stale local branch into the RUN1/RUN2 decision.
# EFFECTIVE_REF below substitutes `refs/remotes/origin/<base-ref>` for
# BASE_REF when it resolves (design.md:
# remote-lookup-is-a-preference-not-a-rewrite — a preference, not a rewrite:
# `origin/main` looks up `origin/origin/main`, finds nothing, and passes
# through unchanged); every verdict line names EFFECTIVE_REF rather than the
# raw argument, so each names the ref the test actually ran against.
#
# HOW TO HAND-VERIFY A REFUSE VERDICT (KAN-446). Re-derive the refused signal
# by hand: which of the branch states, the main checkout's branch, its
# tracked changes, or stray worktrees fired. A main checkout mid-rebase or
# carrying an unrelated staged file is a typical structural cause — fix the
# cause, never the verdict.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/resolve-remote-base.sh
. "$SCRIPT_DIR/lib/resolve-remote-base.sh"

WORKTREE="${1:-}"
BASE_REF="${2:-}"
RECORDED="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$BASE_REF" ] || [ -z "$RECORDED" ]; then
  cat >&2 <<'EOF'
usage: check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->
  <base-ref>  the base branch name, bare (main) or remote-tracking
              (origin/main). The guard prefers refs/remotes/origin/<base-ref>
              when it resolves, so a bare name is never tested against a
              stale local branch.
EOF
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

# Resolve the effective base ref (KAN-88) before testing whether it resolves,
# so the substitution — not the raw argument — is what gets tested and named.
EFFECTIVE_REF="$(resolve_remote_base "$WORKTREE" "$BASE_REF")"

# A base ref that does not resolve would make the ancestor test below fail for
# an environmental reason and read as "not merged" — a RUN1 verdict reached by
# accident. Resolve it first so the failure is named instead.
if ! git -C "$WORKTREE" rev-parse --verify --end-of-options "${EFFECTIVE_REF}^{commit}" >/dev/null 2>&1; then
  echo "REFUSE: base ref '$EFFECTIVE_REF' does not resolve in $WORKTREE — cannot test whether HEAD reached it"
  exit 0
fi

# (c) The ancestor test. Only exit 1 means "not an ancestor"; anything else is
#     git failing to answer, which must not be read as a RUN1 verdict. stderr is
#     redirected so no git chatter can prefix the verdict line for a caller that
#     merges the two streams.
set +e
git -C "$WORKTREE" merge-base --is-ancestor --end-of-options "$HEAD_SHA" "$EFFECTIVE_REF" 2>/dev/null
ANCESTOR_RC=$?
set -e
if [ "$ANCESTOR_RC" -eq 1 ]; then
  echo "RUN1: HEAD is not an ancestor of $EFFECTIVE_REF — not merged"
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
  echo "REFUSE: $EFFECTIVE_REF contains HEAD, but $WORKTREE has $DIRTY uncommitted entries — a merged change should have nothing left to commit"
  exit 0
fi

# (e) The main-checkout assertion — see "THE MAIN-CHECKOUT ASSERTION" in the
#     header above (KAN-462 §4, design.md main-checkout-is-asserted-not-moved)
#     for the REFUSE grammar this block emits.
COMMON_DIR="$(git -C "$WORKTREE" rev-parse --git-common-dir 2>/dev/null)" || {
  echo "check-finish-preflight: cannot resolve the main checkout from $WORKTREE" >&2
  exit 2
}
case "$COMMON_DIR" in
  /*) : ;;
  *) COMMON_DIR="$WORKTREE/$COMMON_DIR" ;;
esac
MAIN_CHECKOUT="$(cd "$(dirname "$COMMON_DIR")" && pwd -P)" || {
  echo "check-finish-preflight: cannot resolve the main checkout from $WORKTREE" >&2
  exit 2
}

STRIPPED_BASE="${BASE_REF#origin/}"
MC_BRANCH="$(git -C "$MAIN_CHECKOUT" branch --show-current 2>/dev/null)" || {
  echo "check-finish-preflight: cannot read the current branch of the main checkout $MAIN_CHECKOUT" >&2
  exit 2
}
if [ "$MC_BRANCH" != "$STRIPPED_BASE" ]; then
  if [ -z "$MC_BRANCH" ]; then
    MC_BRANCH="(detached HEAD)"
  fi
  echo "REFUSE: main checkout $MAIN_CHECKOUT is on $MC_BRANCH, not $STRIPPED_BASE"
  exit 0
fi

MC_STATUS_OUT="$(git -C "$MAIN_CHECKOUT" status --porcelain --untracked-files=no 2>/dev/null)" || {
  echo "check-finish-preflight: cannot read the main checkout status in $MAIN_CHECKOUT" >&2
  exit 2
}
if [ -n "$MC_STATUS_OUT" ]; then
  echo "REFUSE: main checkout $MAIN_CHECKOUT has tracked changes"
  exit 0
fi

SCRIPT_DIR_LOCATION="$SCRIPT_DIR/check-worktree-location.sh"
if [ ! -x "$SCRIPT_DIR_LOCATION" ]; then
  echo "check-finish-preflight: $SCRIPT_DIR_LOCATION is missing or not executable — cannot determine anything" >&2
  exit 2
fi
LOCATION_ERR="$(mktemp "${TMPDIR:-/tmp}/check-finish-preflight-location-err.XXXXXX")"
set +e
LOCATION_OUT="$("$SCRIPT_DIR_LOCATION" "$MAIN_CHECKOUT" 2>"$LOCATION_ERR")"
LOCATION_RC=$?
set -e
LOCATION_ERR_TEXT="$(cat "$LOCATION_ERR")"
rm -f "$LOCATION_ERR"
if [ "$LOCATION_RC" -eq 1 ]; then
  STRAY_LIST="$(printf '%s\n' "$LOCATION_OUT" | grep '^STRAY: ' | sed 's/^STRAY: //' | paste -sd ';' - | sed 's/;/; /g')"
  echo "REFUSE: stray worktree(s): $STRAY_LIST"
  exit 0
fi
if [ "$LOCATION_RC" -eq 2 ]; then
  [ -n "$LOCATION_ERR_TEXT" ] && printf '%s\n' "$LOCATION_ERR_TEXT" >&2
  exit 2
fi
if [ "$LOCATION_RC" -ne 0 ]; then
  echo "check-finish-preflight: check-worktree-location.sh exited $LOCATION_RC against $MAIN_CHECKOUT — cannot determine anything" >&2
  exit 2
fi

echo "RUN2: HEAD is an ancestor of $EFFECTIVE_REF, differs from the recorded merge base, and the worktree is clean"
exit 0
