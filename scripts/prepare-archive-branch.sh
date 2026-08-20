#!/usr/bin/env bash
# prepare-archive-branch.sh — positions the main checkout on the archive
# branch before /myflow-finish run 2 archives a change.
#
# Usage: prepare-archive-branch.sh <main-checkout> <base> <archive-branch>
#
# Puts <main-checkout> on <archive-branch>, cut from <base> fast-forwarded
# to origin/<base>. <base> is whatever resolve-base-branch.sh printed for
# the apply worktree — never guessed here, and never derived from whatever
# branch happens to be checked out in <main-checkout>.
#
# stdout carries ONE line on success and NOTHING ELSE: the branch the
# checkout started from, then the branch it is now on — so a caller composes
# it directly and a refusal can never be misread as a report.
#
#   Exit 0   Positioned. <main-checkout> is on <archive-branch>, cut from a
#            fast-forwarded <base>. stdout: "<started-from> -> <archive-branch>".
#   Exit 1   A named refusal: a dirty working tree (on <base> or off it), a
#            detached HEAD, an existing <archive-branch> that is not descended
#            from origin/<base>, or a <base> or <archive-branch> argument whose
#            name fails the shape check validate_branch_name() applies (see
#            step 2 below) — a name is refused before any git call could read
#            it as an option.
#   Exit 2   Cannot answer — an argument is missing, <main-checkout> is
#            absent, unreadable, or not a git worktree, HEAD's own ref cannot
#            be read, or a `git checkout` this script performs fails: moving
#            to <base>, reusing an existing <archive-branch>, or creating it.
#            Those three are exit 2 rather than exit 1 because a checkout that
#            fails after its preconditions were all checked is a state this
#            script cannot account for, not a refusal it decided on.
#   Exit 3   <base> cannot be fast-forwarded to origin/<base> — the local
#            branch has diverged — INCLUDING when <main-checkout> has no
#            'origin' remote at all, and when 'origin' exists but has no
#            <base> branch on it, so origin/<base> does not resolve. All
#            three are the same verdict because each leaves <base> unable to
#            be reconciled with origin. This mirrors
#            resolve-base-branch.sh's own verdict for a missing origin
#            remote: also exit 3 there, for the same reason — "no origin"
#            and "diverged from origin" are both "the base cannot be
#            reconciled with origin", not a worktree-readability problem.
#
# THE STATE MACHINE, IN ORDER (this header is the authority the by-hand
# fallback in the finish contract cites rather than restates):
#   1. Validate the three arguments and that <main-checkout> is a worktree.
#   2. Validate <base> and <archive-branch> against the same branch-name
#      shape resolve-base-branch.sh applies: the first character is one of
#      [A-Za-z0-9._] and every character is one of [A-Za-z0-9._/-] — so
#      neither argument ever reaches a `git` call as something that could be
#      read as an option.
#   3. Confirm 'origin' exists, then run the bounded, credential-free fetch
#      (WHY THE FETCH IS WRAPPED, below).
#   4. Read HEAD. Detached -> exit 1.
#   5. HEAD is <base> and the tree is dirty -> exit 1: a dirty tree is
#      refused wherever it is, because the changes would otherwise ride onto
#      the archive branch unremarked.
#      HEAD is anything else and the tree is dirty -> exit 1, naming both
#      the branch found and <base>.
#      Otherwise (clean, on <base> or not) -> check out <base> if not
#      already on it.
#   6. Fast-forward <base> to origin/<base> with `git merge --ff-only`,
#      which is a no-op when local already contains origin's tip and fails
#      only on a genuine divergence -> exit 3.
#   7. <archive-branch> does not exist -> create it from the fast-forwarded
#      <base> and check it out.
#      <archive-branch> exists and origin/<base> is an ancestor of it (i.e.
#      it is descended from origin/<base>) -> check it out and reuse it,
#      never recreate or reset it.
#      <archive-branch> exists and is NOT descended from origin/<base> ->
#      exit 1.
#   8. Print the one success line.
#
# Cleanliness is `git -C <main-checkout> status --porcelain --untracked-files=normal`
# being empty — the same test check-finish-preflight.sh uses for its signal
# 3, so run 2's two cleanliness judgments cannot disagree.
#
# WHY THE FETCH IS WRAPPED. Carried from resolve-base-branch.sh verbatim, for
# the same reason: `git remote show origin` (or an ordinary fetch) against an
# unreachable host can block for the better part of a minute on the default
# TCP timeout, which would turn a correct refusal into a long hang.
# `-c core.askpass=true` stops it prompting for credentials, `2>/dev/null`
# swallows the chatter, and `|| true` means a failed fetch is not this
# script's failure — a stale origin/<base> is still usable, just possibly
# behind, which the next invocation's fetch corrects.
set -euo pipefail

# Same defect resolve-base-branch.sh guards against: `A-Za-z0-9` inside a
# `case` bracket expression is a COLLATING range, not a byte range, under a
# UTF-8 locale on this repository's bash 3.2 floor. Pinning the locale here
# costs nothing outside this file — everything below is git plumbing and
# shell builtins.
export LC_ALL=C

MAIN_CHECKOUT="${1:-}"
BASE="${2:-}"
ARCHIVE_BRANCH="${3:-}"

if [ -z "$MAIN_CHECKOUT" ] || [ -z "$BASE" ] || [ -z "$ARCHIVE_BRANCH" ]; then
  echo "usage: prepare-archive-branch.sh <main-checkout> <base> <archive-branch>" >&2
  exit 2
fi

if [ ! -d "$MAIN_CHECKOUT" ]; then
  echo "prepare-archive-branch: $MAIN_CHECKOUT is not a directory" >&2
  exit 2
fi

if ! git -C "$MAIN_CHECKOUT" rev-parse --git-dir >/dev/null 2>&1; then
  echo "prepare-archive-branch: $MAIN_CHECKOUT is not a git worktree" >&2
  exit 2
fi

# validate_branch_name <name> <label> — refuses (exit 1) a name whose first
# character is not one of [A-Za-z0-9._], or that carries any character
# outside [A-Za-z0-9._/-] — the same shape resolve-base-branch.sh enforces
# on the value it resolves, applied here to both arguments before either
# reaches a `git` call.
validate_branch_name() {
  case "$1" in
    [A-Za-z0-9._]*) ;;
    *)
      echo "prepare-archive-branch: $2 '$1' is not a valid branch name" >&2
      exit 1
      ;;
  esac
  case "$1" in
    *[!A-Za-z0-9._/-]*)
      echo "prepare-archive-branch: $2 '$1' is not a valid branch name" >&2
      exit 1
      ;;
  esac
}

validate_branch_name "$BASE" "base branch"
validate_branch_name "$ARCHIVE_BRANCH" "archive branch"

if ! git -C "$MAIN_CHECKOUT" remote get-url origin >/dev/null 2>&1; then
  echo "prepare-archive-branch: no 'origin' remote configured in $MAIN_CHECKOUT — cannot resolve origin/$BASE" >&2
  exit 3
fi

git -C "$MAIN_CHECKOUT" -c core.askpass=true fetch --quiet origin 2>/dev/null || true

# See resolve-base-branch.sh's own comment: `branch --show-current` failing
# (non-zero exit) is a different fact from detached HEAD, where the same
# command succeeds and prints nothing. Told apart the same way here.
if ! CUR="$(git -C "$MAIN_CHECKOUT" branch --show-current 2>/dev/null)"; then
  echo "prepare-archive-branch: could not read the current branch in $MAIN_CHECKOUT" >&2
  exit 2
fi

if [ -z "$CUR" ]; then
  echo "prepare-archive-branch: HEAD is detached in $MAIN_CHECKOUT" >&2
  exit 1
fi

DIRTY=""
if [ -n "$(git -C "$MAIN_CHECKOUT" status --porcelain --untracked-files=normal 2>/dev/null)" ]; then
  DIRTY=1
fi

if [ "$CUR" = "$BASE" ]; then
  if [ -n "$DIRTY" ]; then
    echo "prepare-archive-branch: $MAIN_CHECKOUT has a dirty working tree on '$BASE' — refusing" >&2
    exit 1
  fi
else
  if [ -n "$DIRTY" ]; then
    echo "prepare-archive-branch: $MAIN_CHECKOUT is on '$CUR' with uncommitted changes, not '$BASE' — refusing" >&2
    exit 1
  fi
  git -C "$MAIN_CHECKOUT" checkout -q "$BASE" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not check out '$BASE' in $MAIN_CHECKOUT" >&2
    exit 2
  }
fi

if ! git -C "$MAIN_CHECKOUT" rev-parse -q --verify "refs/remotes/origin/$BASE" >/dev/null 2>&1; then
  echo "prepare-archive-branch: origin/$BASE does not exist — cannot fast-forward '$BASE'" >&2
  exit 3
fi

if ! git -C "$MAIN_CHECKOUT" merge --ff-only -q "origin/$BASE" >/dev/null 2>/dev/null; then
  echo "prepare-archive-branch: '$BASE' cannot be fast-forwarded to origin/$BASE — it has diverged" >&2
  exit 3
fi

if git -C "$MAIN_CHECKOUT" show-ref --verify --quiet "refs/heads/$ARCHIVE_BRANCH"; then
  if ! git -C "$MAIN_CHECKOUT" merge-base --is-ancestor "origin/$BASE" "$ARCHIVE_BRANCH" 2>/dev/null; then
    echo "prepare-archive-branch: '$ARCHIVE_BRANCH' already exists and is not descended from origin/$BASE — refusing" >&2
    exit 1
  fi
  git -C "$MAIN_CHECKOUT" checkout -q "$ARCHIVE_BRANCH" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not check out existing '$ARCHIVE_BRANCH' in $MAIN_CHECKOUT" >&2
    exit 2
  }
else
  git -C "$MAIN_CHECKOUT" checkout -q -b "$ARCHIVE_BRANCH" "$BASE" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not create '$ARCHIVE_BRANCH' from '$BASE' in $MAIN_CHECKOUT" >&2
    exit 2
  }
fi

printf '%s -> %s\n' "$CUR" "$ARCHIVE_BRANCH"
