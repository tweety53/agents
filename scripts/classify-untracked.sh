#!/usr/bin/env bash
# classify-untracked.sh — the archive pre-flight that sorts a landing
# worktree's untracked entries into classes instead of leaving one flat
# refusal count behind (KAN-411; the implementation half of KAN-397).
#
# Usage: classify-untracked.sh <landing-worktree>
#
# Run immediately before prepare-archive-branch.sh, against the same
# <landing-worktree> path. The guard downstream still refuses a dirty tree —
# this script's job is to make the classes it can settle never reach that
# refusal, and to name the rest for the conductor and the operator:
#
#   capture   stray screenshot images from fix rounds (*.png, *.jpg, *.jpeg)
#             — moved to <project>/.worktrees/_scratchpad/, outside the
#             landing worktree, so the throwaway worktree's forced removal
#             cannot destroy them. A same-named file already in the
#             scratchpad is never clobbered; the incoming copy gains a
#             timestamp prefix, plus a counter when that name is taken too.
#   config    local configuration — a `.claude` entry at the worktree root
#             — appended to the checkout's LOCAL exclude file
#             (<common git dir>/info/exclude), never to a committed
#             .gitignore: one project's local residue must not become every
#             clone's ignore rule. That exclude file is uncommitted but
#             shared by every worktree of this checkout — a linked
#             worktree's own <git-dir>/info/exclude is NOT read by status
#             on this machine's git (measured, 2.50.1) — so `.claude/` at
#             any worktree root, the main checkout's included, stops
#             showing in status. That scope is the point: `.claude/` is
#             exactly the residue no checkout of this repository wants
#             tracked. The files stay on disk, now invisible to status.
#   asset     everything else — reported, never touched. The conductor
#             prompts the operator: commit it to the archive branch
#             deliberately, or delete it. Until then it stays untracked and
#             prepare-archive-branch.sh keeps refusing, which is the gate
#             working as designed.
#
# Classification reads `git ls-files --others --exclude-standard --directory`:
# untracked-and-unignored entries exactly, directories collapsed to one
# `<dir>/` line. A directory entry is therefore classified whole — by its own
# name for `.claude`, otherwise as an asset, never by a file inside it; no
# whole directory is moved or deleted on the strength of one file's
# extension. Tracked modifications and staged entries are not this script's
# concern and still refuse downstream, as they always have.
#
# stdout carries one line per action taken or finding reported, and nothing
# else; a clean worktree prints exactly `CLEAN`. An absent <landing-worktree>
# prints nothing — there is nothing to classify, and the guard creates it
# fresh and clean.
#
#   Exit 0   Could answer: clean, classified, or an absent worktree.
#   Exit 2   Cannot answer — an argument is missing, the path is not a
#            directory or not a git worktree, the repository's common
#            directory cannot be resolved, the untracked entries cannot be
#            listed, or a capture or exclude write fails (the cause on
#            stderr; entries already settled stay settled). This script has
#            no exit 1: it never refuses anything, it only settles and
#            reports; refusing stays prepare-archive-branch.sh's job.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "classify-untracked: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec classify-untracked 2 "classify-untracked:" "$@"
