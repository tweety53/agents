#!/usr/bin/env bash
# prepare-archive-branch.sh — positions a throwaway landing worktree on the
# archive branch before /flow's archive run archives a change (KAN-462 §3).
#
# Usage: prepare-archive-branch.sh <landing-worktree> <base> <archive-branch>
#
# Puts <landing-worktree> on <archive-branch>, cut from <base> fast-forwarded
# to origin/<base>. <base> is whatever resolve-base-branch.sh printed for
# the apply worktree — never guessed here, and never derived from whatever
# branch happens to be checked out in <landing-worktree>. <landing-worktree>
# is never the main checkout: it is created under it, at
# <project>/.worktrees/_landing-<name>, and the main checkout itself is never
# read, checked out, or modified by this script — see step 2b below.
#
# stdout carries ONE line on success and NOTHING ELSE: the branch the
# checkout started from, then the branch it is now on — so a caller composes
# it directly and a refusal can never be misread as a report.
#
#   Exit 0   Positioned. <landing-worktree> is on <archive-branch>, cut from a
#            fast-forwarded <base>. stdout: "<started-from> -> <archive-branch>".
#   Exit 1   A named refusal: <landing-worktree>'s parent directory is not
#            named `.worktrees` (step 2b), a dirty working tree (on <base> or
#            off it), a detached HEAD, an existing <archive-branch> that is
#            not descended from origin/<base>, or a <base> or <archive-branch>
#            argument whose name fails the shape check validate_branch_name()
#            applies (see step 2 below) — a name is refused before any git
#            call could read it as an option.
#   Exit 2   Cannot answer — an argument is missing, <landing-worktree> could
#            not be created when absent, or — once positioned as a worktree —
#            is unreadable or not a git worktree, HEAD's own ref cannot be
#            read, or a `git checkout` this script performs fails: moving to
#            <base>, reusing an existing <archive-branch>, or creating it.
#            These are exit 2 rather than exit 1 because a checkout or
#            creation that fails after its preconditions were all checked is
#            a state this script cannot account for, not a refusal it
#            decided on. Exit 2 has one further meaning — post-run drift:
#            the checkout and fast-forward succeeded, but the tree no longer
#            matches the snapshot taken immediately before the first branch
#            move — a new stash entry or an unexpected status line, i.e.
#            residue the run left behind (KAN-423's incident). stderr then
#            carries "post-run drift detected:" followed by every finding,
#            one per line; a failed checkout prints its own named failure
#            the same way. Either way stdout stays empty.
#   Exit 3   <base> cannot be fast-forwarded to origin/<base> — the local
#            branch has diverged — INCLUDING when <landing-worktree> has no
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
#   1. Validate the three arguments.
#   2. Validate <base> and <archive-branch> against the same branch-name
#      shape resolve-base-branch.sh applies: the first character is one of
#      [A-Za-z0-9._] and every character is one of [A-Za-z0-9._/-] — so
#      neither argument ever reaches a `git` call as something that could be
#      read as an option.
#   2b. When <landing-worktree> does not already exist: refuse (exit 1,
#      naming the path) unless its parent directory is named `.worktrees` —
#      by construction it is always <project>/.worktrees/_landing-<name>.
#      Resolve <main-checkout> as the git toplevel of that parent's own
#      parent directory, then run
#      `git -C <main-checkout> worktree add --force <landing-worktree> <base>`.
#      `--force` is required, and is the one deliberate deviation from a bare
#      `worktree add`: the main checkout is, by design, ordinarily already on
#      <base> at this point (`check-finish-preflight.sh`'s own main-checkout
#      assertion, KAN-462 §4), and git refuses to check out a branch that is
#      already checked out in another worktree unless told to. The main
#      checkout's own HEAD is untouched by this — `worktree add` never moves
#      the branch pointer of the worktree already holding it, only permits a
#      second one. When <landing-worktree> already exists, this step is
#      skipped entirely and every check below runs against it unchanged,
#      exactly as it always has against what used to be called
#      <main-checkout>.
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
# <archive-branch> equal to <base> itself means "position on <base> itself",
# and is accepted — the merge-and-push route's own use of this script
# (`skills/flow-contracts/finish-contract-run1.md`'s route table).
#
# Cleanliness is `git -C <landing-worktree> status --porcelain --untracked-files=normal`
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

LANDING="${1:-}"
BASE="${2:-}"
ARCHIVE_BRANCH="${3:-}"

if [ -z "$LANDING" ] || [ -z "$BASE" ] || [ -z "$ARCHIVE_BRANCH" ]; then
  echo "usage: prepare-archive-branch.sh <landing-worktree> <base> <archive-branch>" >&2
  exit 2
fi

# Source the post-mutation self-check library beside the argument
# validation, before the first `git` call — $0-relative resolution holds
# regardless of the caller's cwd, because the kernel resolved $0 itself
# against that same cwd.
. "$(dirname "$0")/lib/post-mutation-check.sh"

# validate_branch_name <name> <label> — refuses (exit 1) a name whose first
# character is not one of [A-Za-z0-9._], or that carries any character
# outside [A-Za-z0-9._/-] — the same shape resolve-base-branch.sh enforces
# on the value it resolves, applied here to both arguments before either
# reaches a `git` call. Defined before step 2b below, which needs $BASE
# already validated before handing it to `git worktree add`.
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

# Step 2b: <landing-worktree> is created, from the main checkout, when it
# does not already exist. See the header for why --force is required.
if [ ! -e "$LANDING" ]; then
  LANDING_PARENT="$(dirname "$LANDING")"
  if [ "$(basename "$LANDING_PARENT")" != ".worktrees" ]; then
    echo "prepare-archive-branch: $LANDING's parent is not a .worktrees directory — refusing to create it" >&2
    exit 1
  fi
  MAIN_CHECKOUT="$(git -C "$(dirname "$LANDING_PARENT")" rev-parse --show-toplevel 2>/dev/null)" || {
    echo "prepare-archive-branch: cannot resolve the main checkout above $LANDING" >&2
    exit 2
  }
  git -C "$MAIN_CHECKOUT" worktree add --force --quiet -- "$LANDING" "$BASE" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not create the landing worktree $LANDING from '$BASE' in $MAIN_CHECKOUT" >&2
    exit 2
  }
fi

if [ ! -d "$LANDING" ]; then
  echo "prepare-archive-branch: $LANDING is not a directory" >&2
  exit 2
fi

if ! git -C "$LANDING" rev-parse --git-dir >/dev/null 2>&1; then
  echo "prepare-archive-branch: $LANDING is not a git worktree" >&2
  exit 2
fi

if ! git -C "$LANDING" remote get-url origin >/dev/null 2>&1; then
  echo "prepare-archive-branch: no 'origin' remote configured in $LANDING — cannot resolve origin/$BASE" >&2
  exit 3
fi

git -C "$LANDING" -c core.askpass=true fetch --quiet origin 2>/dev/null || true

# See resolve-base-branch.sh's own comment: `branch --show-current` failing
# (non-zero exit) is a different fact from detached HEAD, where the same
# command succeeds and prints nothing. Told apart the same way here.
if ! CUR="$(git -C "$LANDING" branch --show-current 2>/dev/null)"; then
  echo "prepare-archive-branch: could not read the current branch in $LANDING" >&2
  exit 2
fi

if [ -z "$CUR" ]; then
  echo "prepare-archive-branch: HEAD is detached in $LANDING" >&2
  exit 1
fi

DIRTY=""
if [ -n "$(git -C "$LANDING" status --porcelain --untracked-files=normal 2>/dev/null)" ]; then
  DIRTY=1
fi

if [ "$CUR" = "$BASE" ]; then
  if [ -n "$DIRTY" ]; then
    echo "prepare-archive-branch: $LANDING has a dirty working tree on '$BASE' — refusing" >&2
    exit 1
  fi
else
  if [ -n "$DIRTY" ]; then
    echo "prepare-archive-branch: $LANDING is on '$CUR' with uncommitted changes, not '$BASE' — refusing" >&2
    exit 1
  fi
fi

# Snapshot the tree the branch moves start from — status and stash list —
# so the post-run check can tell residue from the state the run left.
TREE_SNAPSHOT="$(snapshot_tree_state "$LANDING")"

if [ "$CUR" != "$BASE" ]; then
  git -C "$LANDING" checkout -q "$BASE" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not check out '$BASE' in $LANDING" >&2
    exit 2
  }
fi

if ! git -C "$LANDING" rev-parse -q --verify "refs/remotes/origin/$BASE" >/dev/null 2>&1; then
  echo "prepare-archive-branch: origin/$BASE does not exist — cannot fast-forward '$BASE'" >&2
  exit 3
fi

if ! git -C "$LANDING" merge --ff-only -q "origin/$BASE" >/dev/null 2>/dev/null; then
  echo "prepare-archive-branch: '$BASE' cannot be fast-forwarded to origin/$BASE — it has diverged" >&2
  exit 3
fi

if git -C "$LANDING" show-ref --verify --quiet "refs/heads/$ARCHIVE_BRANCH"; then
  if ! git -C "$LANDING" merge-base --is-ancestor "origin/$BASE" "$ARCHIVE_BRANCH" 2>/dev/null; then
    echo "prepare-archive-branch: '$ARCHIVE_BRANCH' already exists and is not descended from origin/$BASE — refusing" >&2
    exit 1
  fi
  git -C "$LANDING" checkout -q "$ARCHIVE_BRANCH" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not check out existing '$ARCHIVE_BRANCH' in $LANDING" >&2
    exit 2
  }
else
  git -C "$LANDING" checkout -q -b "$ARCHIVE_BRANCH" "$BASE" >/dev/null 2>&1 || {
    echo "prepare-archive-branch: could not create '$ARCHIVE_BRANCH' from '$BASE' in $LANDING" >&2
    exit 2
  }
fi

# Post-run self-check: the tree must still match the snapshot taken before
# the first branch move — any new stash entry or unexpected status line is
# residue this run left behind, named here rather than left for a conductor
# to retry blind (KAN-423's incident).
drift="$(check_tree_restored "$LANDING" "$TREE_SNAPSHOT" || true)"
if [ -n "$drift" ]; then
  echo "prepare-archive-branch: post-run drift detected:" >&2
  printf '%s\n' "$drift" >&2
  exit 2
fi

printf '%s -> %s\n' "$CUR" "$ARCHIVE_BRANCH"
