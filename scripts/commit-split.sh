#!/usr/bin/env bash
# commit-split.sh — the guarded commit chain, in one place.
#
# Usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg> [<trailer>…]
#
# Wraps the exact guarded chain specified in skills/flow-contracts/pipeline.md's
# "Git boundaries" section — that section is the canonical spec of this
# behavior and is not rewritten here; this script only exists to give the two
# call sites (`/flow`'s PR-exception path in verify-and-handoff and its integrate run) one
# place to call instead of each spelling the chain out inline.
#
# Refuses (exit 2, naming the path) when `<spec root>/changes/` is not
# gitignored, since `git add -A` would then commit the change directory.
# Stages everything (`git add -A -- .`) and commits it as the
# implementation commit — skipped, not failed, if nothing is staged — with
# `Flow-Key: flow.commit-two` and each <trailer> (a squashed commit's
# `Task-Id: <n>`) in its trailer block. Then
# stages anything left and commits that under <plan-msg>, skipped the same
# way if empty. `<spec root>/changes/` is gitignored, and
# `<spec root>/changes/<name>/` is first untracked (`git rm -r --cached`, the
# files kept on disk), so the change directory is in neither commit even where
# an older tooling committed it.
#
# `<name>` is the change: before anything is staged, the in-process
# change-worktree location check refuses a <worktree> that is a main
# checkout or is not on `spectre/<name>`, and its verdict and exit code stop
# the chain.
#
# Exit codes: 0 when the chain ran (either commit possibly skipped); 2 with a
# usage line on stderr when fewer than four arguments are given, or when the
# change directory is not gitignored; the location
# guard's own code on its refusal; git's own code on a git failure.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "commit-split: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec commit-split 2 "commit-split:" "$@"
