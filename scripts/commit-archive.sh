#!/usr/bin/env bash
# commit-archive.sh — make /flow run 1's drop commit on the change
# branch, in the apply worktree (skills/flow/integrate.md), in one call. The
# name and the ARCHIVE-* tokens predate the drop and stay.
#
# Usage: commit-archive.sh <worktree> <name>
#
# In order: asserts <worktree> is on spectre/<name>; runs
# check-done-when-paths (in-process, the same flow-guard), which refuses a
# `## Done when` naming a path the index does not track; removes
# spectre/changes/<name>/ and every spectre/changes/<name>-fix-<N>/ from the
# index and the disk (`git rm -r`, then the untracked rest); runs
# check-archive-scope <worktree> <each removed directory> (in-process, the
# same flow-guard), which admits only staged deletions under them; then
# commits with the fixed subject `chore(spectre): drop <name>` unless
# nothing is staged. No ledger or panel record is copied anywhere.
#
# Prints ONE verdict line to stdout:
#   ARCHIVE-COMMITTED: <sha>        the drop commit was made
#   ARCHIVE-NOTHING-STAGED          nothing to commit — already committed
#   ARCHIVE-WRONG-BRANCH: <found>   the worktree is on another branch
#                                   (or `(detached HEAD)`); nothing written
#   (or check-archive-scope.sh's OUT-OF-SCOPE and SCOPE-VIOLATION lines, or
#   check-done-when-paths's DONE-WHEN-PATH and DONE-WHEN-VIOLATION lines)
#
# Exit 0 committed or nothing staged; 1 wrong branch, a scope violation, or
# a Done-when violation — nothing committed, the change stays at IN_PROGRESS;
# 2 cannot answer, with NOTHING on stdout: a usage error, a name that is not a
# plain change name, a path that is not a git worktree, a directory that
# could not be removed, a git step that failed, or
# check-archive-scope.sh / check-done-when-paths unable to answer.
#
# The branch is asserted before anything is copied, so a refusal leaves the
# worktree as it was found. The subject is a fixed literal: `flow
# self-review bundle` resolves this commit by matching it whole.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "commit-archive: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec commit-archive 2 "commit-archive:" "$@"
