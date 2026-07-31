#!/usr/bin/env bash
# check-cleanup-complete.sh — verify that everything the cleanup registry says
# should be gone after /myflow-finish run 2 actually is.
#
# Usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>
#
# Prints ONE verdict line to stdout:
#   COMPLETE: <reason>    every registry row whose lifetime ends at run 2 is gone
#   LEFTOVER: <breakdown> one or more are still present
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable repository or state directory, or a change name outside the
# allowlist. The VERDICT carries the answer, not the exit status
# — see check-finish-preflight.sh's header for why this repository separates
# them. Finish runs this once per repository, so each verdict names the
# repository it judged; a bare COMPLETE would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-unfinished-work.sh's. That guard says
# CLEAR/OUTSTANDING, this one says COMPLETE/LEFTOVER, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence, as
# the paragraph below spells out: there a missing file counts AGAINST the
# change, here a missing artifact IS the answer. A reader who carried the
# meaning of CLEAR across to COMPLETE would carry that inversion with it.
# check-finish-preflight.sh's third vocabulary (RUN1/RUN2/REFUSE) differs for a
# further reason again: it selects a procedure rather than reporting a state.
#
# Reporting a leftover is the whole point: run 2 previously assumed its own
# removals succeeded.
#
# THE FIVE ROWS, AND WHY THEY ARE THESE FIVE. The registry in
# skills/myflow-contracts/pipeline.md owns the list; the rows whose lifetime
# ends at run 2 are the worktree, the local branch, the remote branch, the
# change directory and the proposal artifact source. The per-task diffs, the
# panel record and the session ledger are not checked separately because the
# registry removes them WITH the worktree — a surviving one of those is a
# surviving worktree, already reported. The state file is not checked at all:
# its row says it is never removed, so its presence is correct and reporting it
# would train the operator to ignore this guard's output.
#
# THAT DERIVATION IS DECLARED BELOW RATHER THAN LEFT IN PROSE, because a prose
# restatement of a table in another file is exactly what goes stale: a registry
# row added later whose lifetime ends at run 2 would not be checked here, and
# this guard would then report COMPLETE over a real leftover — the failure it
# exists to prevent, one layer up. Every registry row is named in exactly one of
# the two lists below, and test-check-cleanup-complete.sh reads BOTH the
# registry and these markers and fails when they disagree in either direction:
# a registry row with no line here, or a line here naming a row the registry no
# longer has. Adding a registry row therefore fails this guard's suite until
# someone decides, in writing, which list it belongs in.
#
# registry-row-checked: Worktree
# registry-row-checked: Local branch
# registry-row-checked: Remote branch
# registry-row-checked: Change directory
# registry-row-checked: Proposal artifact source
# registry-row-not-checked: Per-task and review diffs — removed with the worktree
# registry-row-not-checked: Panel record — removed with the worktree
# registry-row-not-checked: SDD ledger — removed with the worktree
# registry-row-not-checked: State file — never removed; it is the terminal record
#
# ABSENCE IS THE ANSWER HERE, not a gap in the evidence. The run-1 gate treats
# a file it cannot find as outstanding, because a missing record proves nothing
# about the work. This guard asks the opposite question — "is it gone?" — so a
# row it cannot find is that row answered. What must never be inferred is
# absence from a path that was never readable, which is why an unreadable
# repository and an unreadable state directory both refuse to answer instead.
#
# THE ARCHIVED CHANGE DIRECTORY IS NOT A LEFTOVER. The registry says the change
# directory is MOVED into openspec/changes/archive/ and never deleted, so only
# one still sitting at openspec/changes/<name>/ counts.
set -euo pipefail

REPO="${1:-}"
NAME="${2:-}"
STATE_DIR="${3:-}"

if [ -z "$REPO" ] || [ -z "$NAME" ] || [ -z "$STATE_DIR" ]; then
  echo "usage: check-cleanup-complete.sh <repo> <change-name> <state-dir>" >&2
  exit 2
fi

# CONTAINMENT: the change name arrives from a pull-request-editable state file
# and is concatenated into a path, a ref name and a `-d` test below. Without
# this check `../../nonexistent-decoy` makes every row answer "already gone" and
# the guard reports COMPLETE while the real change directory and artifact source
# are still sitting there — a confirmed cleanup that removed nothing.
#
# The rule is preserve-session-records.sh's Protection 1, character for
# character, and its comment there is canonical for why each hazard is in it.
# THE COPY IS DELIBERATE, for the reason recorded at the same place in
# check-unfinished-work.sh: these guards are single-file by design and are
# copied into projects one at a time, so a sourced helper would make a guard
# that is present but unrunnable. Both harnesses assert the same rejected
# shapes, which is what keeps the copies from drifting apart silently.
#
# It also closes the symlink question at these paths: `-f` and `-d` follow
# symlinks, so a name that cannot leave the repository is what makes the paths
# this guard tests the ones it was asked about.
case "$NAME" in
  [!A-Za-z0-9]* | *[!A-Za-z0-9._-]*)
    echo "check-cleanup-complete: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

if [ ! -d "$REPO" ] || ! git -C "$REPO" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-cleanup-complete: $REPO is not a git repository — cannot determine anything" >&2
  exit 2
fi

# A state directory that is not there cannot answer whether the artifact source
# is gone. A mistyped path would otherwise read as "the artifact is absent" and
# contribute to a COMPLETE verdict nobody checked.
if [ ! -d "$STATE_DIR" ]; then
  echo "check-cleanup-complete: $STATE_DIR is not a directory — cannot tell whether the proposal artifact source remains" >&2
  exit 2
fi

LEFT=""
add() { LEFT="${LEFT:+$LEFT; }$1"; }

# Row one — the worktree, found by BRANCH rather than by path, exactly as the
# Worktree cleanup contract finds worktrees when the state file's map is
# absent: a path is never guessed from a conventional layout.
#
# A command substitution, never a scratch file: a verifier that writes into the
# tree it is verifying can leave behind exactly the class of leftover it exists
# to report.
#
# git's failure is not read as an empty list. `|| true` here would turn an
# unreadable repository into a COMPLETE verdict — the reassuring one — which is
# the same silence this guard was added to break.
WT_PORCELAIN="$(git -C "$REPO" worktree list --porcelain 2>/dev/null)" || {
  echo "check-cleanup-complete: cannot list the worktrees of $REPO — cannot determine anything" >&2
  exit 2
}

# The path is taken with substr, not $2: `worktree list --porcelain` emits the
# path raw, so a worktree under a TMPDIR or a home directory containing a space
# is truncated at the first one by a field reference, and the operator is then
# sent to a path that does not exist. The branch, on the next line, is a ref
# name and cannot contain a space, so $2 is right for it — and the comparison
# is for EQUALITY, so a neighbouring change's openspec/<name>-something is not
# reported as this change's leftover.
#
# A worktree still listed after `git worktree prune` should have run is a
# leftover whether or not its directory survives: the registration is what run
# 2 is required to remove.
WT_PATHS="$(printf '%s\n' "$WT_PORCELAIN" | awk -v b="refs/heads/openspec/$NAME" '
  /^worktree / { w = substr($0, 10) }
  /^branch /   { if ($2 == b) { out = out (n++ ? ", " : "") w } }
  END          { if (n) print out }
')"
if [ -n "$WT_PATHS" ]; then
  add "worktree(s) still registered for openspec/$NAME at $WT_PATHS"
fi

# Rows two and three — the local branch and the remote branch. The remote one
# is read through its tracking ref, which is what this repository can see
# without a network call; run 2 deletes the remote branch and prunes the ref
# together, so a surviving ref is the observable half of that step failing.
# `show-ref --verify` matches the full ref name, never a prefix.
if git -C "$REPO" show-ref --verify --quiet "refs/heads/openspec/$NAME"; then
  add "the local branch openspec/$NAME still exists"
fi

if git -C "$REPO" show-ref --verify --quiet "refs/remotes/origin/openspec/$NAME"; then
  add "the remote-tracking ref origin/openspec/$NAME still exists"
fi

# Row four — the change directory, which run 2 moves into the archive.
if [ -d "$REPO/openspec/changes/$NAME" ]; then
  add "openspec/changes/$NAME was never moved into the archive"
fi

# Row five — the proposal artifact source, whose removal at run 2 is
# conditional on a preserved copy existing. This guard reports it as remaining;
# it does not decide whether keeping it was right, which is the run's judgment
# and not a fact about the tree.
if [ -f "$STATE_DIR/$NAME-proposal-artifact.html" ]; then
  add "the proposal artifact source is still in the state directory"
fi

if [ -z "$LEFT" ]; then
  echo "COMPLETE: $REPO — no worktree, local branch, remote-tracking ref, unarchived change directory or proposal artifact source remains for openspec/$NAME"
else
  echo "LEFTOVER: $REPO — $LEFT"
fi
exit 0
