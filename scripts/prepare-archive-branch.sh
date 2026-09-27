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
#            off it — the refusal then names every dirty entry on stderr and,
#            when the landing path's leaf is `_landing-<name>` beside an
#            existing apply worktree whose branch `<name>` exists, says for
#            each entry whether it looks like this change's output, meaning
#            the entry's path is changed on `<name>` relative to its
#            merge-base with <base>; when that worktree or branch is missing
#            the entries are still named, under an unavailable-classification
#            line, and nothing is guessed), a detached HEAD, an existing
#            <archive-branch> that is
#            not descended from origin/<base>, or a <base> or <archive-branch>
#            argument whose name fails the branch-name shape check (see
#            step 2 below) — a name is refused before any git call could
#            read it as an option.
#   Exit 2   Cannot answer — an argument is missing, <landing-worktree> could
#            not be created when absent, or — once positioned as a worktree —
#            is unreadable or not a git worktree, is not ITSELF a git
#            worktree (its directory carries no .git entry of its own, so
#            the resolution walked up out of it — step 2c), is a worktree of
#            a different repository than the main checkout above its
#            .worktrees parent (step 2c), the working tree state cannot be
#            read — before the branch moves, or on the post-run recompute —,
#            HEAD's own ref cannot be
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
#   2c. Assert the landing directory is itself a git worktree of the right
#      repository, before any further `git -C` run. `git -C <landing>`
#      resolves by walking up: a `_landing-<name>` a daemon recreated as a
#      plain directory inside the main checkout's tree resolves there, and
#      every later step would act on the main checkout, whose stale tree the
#      chain then found only through the dirty-tree refusal (KAN-823's
#      incident). So the landing root must carry its own .git entry —
#      anything the resolution reached by walking up does not — and, when
#      its parent is `.worktrees` (the construction this script creates
#      under), its common directory must be the main checkout's, resolved
#      as in step 2b. Off that construction no repository is compared
#      against and none is guessed.
#   3. Confirm 'origin' exists, then run the bounded, credential-free fetch
#      (WHY THE FETCH IS WRAPPED, below).
#   4. Read HEAD. Detached -> exit 1.
#   5. HEAD is <base> and the tree is dirty -> exit 1: a dirty tree is
#      refused wherever it is, because the changes would otherwise ride onto
#      the archive branch unremarked; the refusal names every dirty entry
#      and classifies it against the change's own branch (see exit 1).
#      HEAD is anything else and the tree is dirty -> exit 1, naming both
#      the branch found and <base>, with the same per-entry report.
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
# `-c core.askpass=true` stops it prompting for credentials, its stderr is
# discarded, and its exit status is ignored: a failed fetch is not this
# script's failure — a stale origin/<base> is still usable, just possibly
# behind, which the next invocation's fetch corrects.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this script's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "prepare-archive-branch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec prepare-archive-branch 2 "prepare-archive-branch:" "$@"
