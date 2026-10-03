#!/usr/bin/env bash
# commit-split.sh — the guarded two-commit chain, in one place.
#
# Usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>
#
# Wraps the exact guarded chain specified in skills/flow-contracts/pipeline.md's
# "Git boundaries" section — that section is the canonical spec of this
# behavior and is not rewritten here; this script only exists to give the two
# call sites (`/flow`'s PR-exception path in verify-and-handoff and its integrate run) one
# place to call instead of each spelling the chain out inline.
#
# Clears the planning path from the index, stages everything else, and
# commits it as the implementation commit — skipped, not failed, if nothing
# is staged. Then stages the planning paths (and anything else left) and
# commits that as the planning commit, skipped the same way if empty.
#
# THE PLANNING PATH INSIDE THE SPECTRE TREE IS `spectre/changes/`, NOT
# `spectre/`. A capability spec under `spectre/specs/` is implementation —
# the implementer writes and commits it on the change branch, in the task
# commit that implements the requirement — so it must fall on the
# implementation side of this split, and widening either pathspec back to
# `spectre/` would classify it as planning again. See
# skills/flow-contracts/git-boundaries.md, canonical for that boundary.
#
# A CHANGE DIRECTORY'S `link.md` IS ALSO IMPLEMENTATION, CARVED OUT BY FILE
# RATHER THAN BY DIRECTORY. It lives at `<plan_dir><id>/link.md`, inside the
# planning exclusion, so it gets its own `git add` after the exclude-add, and
# a commit that touches no link.md stays a no-op rather than a failure —
# stats/internal/guard/commitsplit.go carries the pathspec reasoning beside
# the code.
#
# `<name>` is the change: before anything is staged,
# check-planning-commit-location.sh refuses a <worktree> that is a main
# checkout or is not on `spectre/<name>`, and its verdict and exit code stop
# the chain — the planning half of this split is a planning commit, and a
# planning commit lands only in the change's own worktree.
#
# A PLANNING PATH THAT IS A TRACKED SYMLINK IS NEVER WORKED AROUND. If
# `spectre/changes/` is a tracked symlink — or `spectre/` is, putting
# `spectre/changes/` behind one — `git add -A -- .
# ':(exclude)spectre/changes/'` exits 128 with a message naming the path, and
# stages nothing; that exit code and message are this guard's, as-is. The
# only way past the stop is to fix the repository so the path is a real
# directory; working around it here would let planning content land in the
# implementation commit, the one outcome this split exists to prevent.
#
# Exit codes: 0 when the chain ran (either commit possibly skipped); 2 with a
# usage line on stderr when fewer than four arguments are given; the location
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
