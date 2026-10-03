#!/usr/bin/env bash
# check-finish-preflight.sh — decide whether /flow should integrate
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
# state, since /flow's implement phase may not commit before a PR exists. Run 2 would then
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
# EFFECTIVE_REF (resolveRemoteBase in stats/internal/guard/resolveremotebase.go)
# substitutes `refs/remotes/origin/<base-ref>` for BASE_REF when it resolves
# (design.md: remote-lookup-is-a-preference-not-a-rewrite — a preference, not
# a rewrite: `origin/main` looks up `origin/origin/main`, finds nothing, and
# passes through unchanged); every verdict line names EFFECTIVE_REF rather
# than the raw argument, so each names the ref the test actually ran against.
#
# HOW TO HAND-VERIFY A REFUSE VERDICT (KAN-446). Re-derive the refused signal
# by hand: which of the branch states, the main checkout's branch, its
# tracked changes, or stray worktrees fired. A main checkout mid-rebase or
# carrying an unrelated staged file is a typical structural cause — fix the
# cause, never the verdict.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
# FLOW_GUARD_WORKTREE_LOCATION is exported so the Go guard execs
# check-worktree-location.sh from beside this script, as the bash did.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-finish-preflight: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_WORKTREE_LOCATION="$SCRIPT_DIR/check-worktree-location.sh"
export FLOW_GUARD_WORKTREE_LOCATION
flow_guard_exec check-finish-preflight 2 "check-finish-preflight:" "$@"
