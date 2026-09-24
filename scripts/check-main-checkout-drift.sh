#!/usr/bin/env bash
# check-main-checkout-drift.sh — name a main checkout's drift from its
# expected post-merge state, before a /flow run's preflight reaches it
# (KAN-647, observed in kan-574's archive run).
#
# Usage: check-main-checkout-drift.sh <main-checkout>
#
# Prints its findings, then at most one verdict line, all to stdout:
#   DRIFT-BRANCH: <path> — on <branch>, not <default>   not on the default branch
#   DRIFT-DIRTY: <path> — <n> tracked entries            tracked changes present
#   DRIFT-CLEAN: <path> — <default>                      neither finding
#
# DRIFT-BRANCH and DRIFT-DIRTY are findings: either or both may print, in
# that order. DRIFT-CLEAN prints exactly when neither does. A detached HEAD
# is named `on (detached HEAD), not <default>` rather than left blank.
#
# Exit 0 on any verdict; exit 2 with NOTHING on stdout when the argument
# is missing, is not a readable directory, is not a git repository, the
# repository has no resolvable `refs/remotes/origin/HEAD`, or `git status`
# fails there — an inability is never reported as a verdict.
#
# WHAT DRIFT MEANS. Every pipeline run works in a worktree, and the main
# checkout's expected state is the post-merge one: on the repository's
# default branch with nothing tracked modified, staged or unmerged — the
# only state from which a landing or an archive can proceed and in which
# the checkout reflects what actually landed. kan-574's archive run found
# two "reverse image of a landing" incidents — tracked content silently
# reverted to pre-merge state — and a main checkout left on an unrelated
# feature branch; both were discovered ad hoc, mid-run. This guard names
# both shapes before any work starts.
#
# WHY THE STATUS SHAPE IS THE CONTENT MARKER. `git status --porcelain
# --untracked-files=no` is the cheap read: one invocation, no object
# database walk. A checkout that is clean but BEHIND its default branch is
# ordinary not-pulled state, indistinguishable from benign by any local
# marker and never this guard's finding — the archive's own refresh step
# owns bringing the checkout forward. Untracked files are hidden from the
# read entirely, exactly as the KAN-546 guard's identical read hides them:
# the pipeline never creates them in a main checkout, and the operator's
# own residue is not drift this gate stops for.
#
# THE DEFAULT BRANCH is what `refs/remotes/origin/HEAD` points at — the
# repository's own answer, not a name invented here. A repository without
# one cannot answer, so it exits 2 rather than guessing `main`.
#
# THE VERDICT NAMES THE PHYSICAL PATH, resolved with `cd … && pwd -P`, so a
# caller that passed a symlinked path still sees the real checkout named —
# the same resolution `git worktree list` applies when it prints paths.
#
# HOW TO HAND-VERIFY A DRIFT VERDICT. Run the same reads by hand:
# `git -C <main-checkout> branch --show-current` against the branch
# `git -C <main-checkout> symbolic-ref refs/remotes/origin/HEAD` names with
# its `refs/remotes/origin/` prefix stripped, and
# `git -C <main-checkout> status --porcelain --untracked-files=no` — those,
# and only those, are the inputs the findings report. A main checkout
# mid-rebase or carrying an unrelated staged file is a typical structural
# cause — fix the cause, never the verdict.
set -euo pipefail

SELF="check-main-checkout-drift"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -eq 1 ] || {
  printf 'usage: check-main-checkout-drift.sh <main-checkout>\n' >&2
  exit 2
}

ROOT="$(cd "$1" 2>/dev/null && pwd -P)" || die "$1 is not a directory"
git -C "$ROOT" rev-parse --git-dir >/dev/null 2>&1 ||
  die "$ROOT is not a git repository"

DEFAULT="$(
  git -C "$ROOT" symbolic-ref --quiet refs/remotes/origin/HEAD 2>/dev/null
)" || die "$ROOT has no resolvable refs/remotes/origin/HEAD — the default branch cannot be named"
DEFAULT="${DEFAULT#refs/remotes/origin/}"
# A symbolic ref can dangle — `git update-ref -d` on the branch it points at
# leaves the symref behind — and a name that resolves to nothing is not an
# answer. Verify the target before trusting it.
git -C "$ROOT" rev-parse --verify --quiet "refs/remotes/origin/$DEFAULT^{commit}" >/dev/null 2>&1 ||
  die "$ROOT's refs/remotes/origin/HEAD dangles — $DEFAULT does not resolve"

BRANCH="$(git -C "$ROOT" branch --show-current 2>/dev/null)" ||
  die "cannot read the current branch of $ROOT"
[ -n "$BRANCH" ] || BRANCH="(detached HEAD)"

STATUS_OUT="$(git -C "$ROOT" status --porcelain --untracked-files=no 2>/dev/null)" ||
  die "cannot read the status of $ROOT"

ENTRIES=0
if [ -n "$STATUS_OUT" ]; then
  ENTRIES="$(printf '%s\n' "$STATUS_OUT" | wc -l | tr -d ' ')"
fi

if [ "$BRANCH" != "$DEFAULT" ]; then
  printf 'DRIFT-BRANCH: %s — on %s, not %s\n' "$ROOT" "$BRANCH" "$DEFAULT"
fi

if [ "$ENTRIES" -ne 0 ]; then
  printf 'DRIFT-DIRTY: %s — %d tracked entries\n' "$ROOT" "$ENTRIES"
fi

if [ "$BRANCH" = "$DEFAULT" ] && [ "$ENTRIES" -eq 0 ]; then
  printf 'DRIFT-CLEAN: %s — %s\n' "$ROOT" "$DEFAULT"
fi
