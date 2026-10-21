#!/usr/bin/env bash
# check-plan-unchanged.sh snapshot|verify <worktree> <change-name> <snapshot-file>
#
# The guard KAN-635 asked for: the change's plan tree survives every
# reviewer dispatch, asserted by git — never by the reviewer's own prose.
# On gymie KAN-579 the gated per-task reviewer ran `git checkout <sha> -- .`
# mid-review, silently destroyed the never-committed design.md/tasks.md
# working-tree edits (task commits never carry plan paths, per
# check-task-commit-planning-paths.sh, so the working tree held the only
# current copy), and then self-reported the tree restored — a claim nothing
# checked, and a false one. implement.md's **The plan tree survives every
# reviewer dispatch** and review-panel.md's round-dispatch paragraph wire
# the two modes around every review dispatch:
#
#   snapshot  run immediately BEFORE the dispatch's Agent call, after the
#             parent has committed any uncommitted plan edits — writes the
#             plan tree's git status to <snapshot-file>
#   verify    run when the dispatch's report file(s) exist, BEFORE any
#             verdict is read or acted on — recomputes the same status and
#             compares byte-for-byte
#
# THE GUARD IS SCOPED TO THE PLAN TREE (`spectre/changes/<name>/`, and
# `spectre/changes/archive/<name>/`, where a fix run writes once integrate's
# run 1 archived the change), deliberately, and never asserts HEAD or the rest of the worktree: a
# gated reviewer bundle flies BESIDE the next implementer group (implement.md's
# bundled dispatch), whose source edits and commits legitimately move
# everything else while the reviewer is in flight — a whole-worktree assert
# would fire on every normal run, and the incident's own loss was the plan
# tree ("source code was never at risk", the KAN-635 issue records). The
# status includes untracked files, so a file planted under the plan tree is
# a violation too; the plan directory being absent at snapshot and still
# absent at verify is clean (nothing was there to lose).
#
# THE SNAPSHOT IS THE EVIDENCE, and verify is the only restore claim this
# pipeline accepts: the dispatching parent never reads the reviewer's
# "restored"/"clean" prose as an answer to what happened to the tree. Verify
# exit 1 therefore stops the run — the reviewer reported about a tree it
# also changed — and recovery is by hand from the pre-dispatch commit (which
# holds the plan tree precisely because the parent commits before it
# snapshots), never by re-running the reviewer.
#
# THE CHANGE NAME IS A PATHSPEC COMPONENT here, so the package's one
# change-name allowlist (stats/internal/guard/changename.go's
# plainChangeName) refuses anything but a plain change name before git ever
# sees it.
#
# Exit codes:
#   0  snapshot written, or the plan tree matches its snapshot
#   1  verify found the plan tree changed since the snapshot — each
#      difference named on stderr
#   2  cannot answer at all — missing arguments, a bad mode, a non-directory
#      worktree, a change name outside the allowlist, the worktree not being
#      a git repository, a git failure, or (on verify) a missing or
#      unreadable snapshot file or an unwritable snapshot path (on snapshot)
#
# The guard itself is Go: stats/internal/guard/planunchanged.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-plan-unchanged: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-plan-unchanged 2 "check-plan-unchanged:" "$@"
