#!/usr/bin/env bash
# commit-archive.sh — make /flow run 2's archive commit (skills/flow/archive.md
# step 4) on the landing worktree, in one call.
#
# Usage: commit-archive.sh <landing-worktree> <canonical-worktree> <name>
#
# In order: asserts <landing-worktree> is on chore/archive-<name>; copies
# <canonical-worktree>/.superpowers/sdd/ledgers/<name>.md and
# .../reviews/<name>-panel.md into spectre/changes/archive/<name>/ as
# ledger.md and panel.md, each when present (an absent file copies nothing);
# stages with `git add -A`; runs check-archive-scope.sh <landing-worktree>
# "spectre/changes/"; then commits with the fixed subject
# `chore(spectre): archive <name>` unless nothing is staged.
#
# Prints ONE verdict line to stdout:
#   ARCHIVE-COMMITTED: <sha>        the archive commit was made
#   ARCHIVE-NOTHING-STAGED          nothing to commit — already committed
#   ARCHIVE-WRONG-BRANCH: <found>   the landing worktree is on another branch
#                                   (or `(detached HEAD)`); nothing written
#   (or check-archive-scope.sh's OUT-OF-SCOPE and SCOPE-VIOLATION lines)
#
# Exit 0 committed or nothing staged; 1 wrong branch or a scope violation —
# nothing committed, the change stays at IN_PROGRESS; 2 cannot answer, with
# NOTHING on stdout: a usage error, a name that is not a plain change name, a
# landing path that is not a git worktree, a present record that could not be
# copied, a git step that failed, or check-archive-scope.sh unable to answer.
#
# The branch is asserted before anything is copied, so a refusal leaves the
# landing worktree as it was found. The subject is a fixed literal: `flow
# self-review bundle` resolves this commit by matching it whole.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
# FLOW_GUARD_SELF (the path this script was invoked by) is exported so the Go
# guard execs $SCRIPT_DIR/check-archive-scope.sh from beside this script.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "commit-archive: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
[ -x "$SCRIPT_DIR/check-archive-scope.sh" ] || {
  echo "commit-archive: no executable check-archive-scope.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec commit-archive 2 "commit-archive:" "$@"
