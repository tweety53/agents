#!/usr/bin/env bash
# check-tree-markers.sh snapshot|verify <worktree> <markers-file> <snapshot-file>
#
# The guard KAN-643 asked for: a subagent's self-reported clean state is
# never evidence for anything — every dispatch that can touch the worktree
# is bracketed by CONTENT MARKERS recorded before the dispatch and
# re-grepped after it, and a mismatch is read as the subagent having
# mutated the tree. On gymie KAN-579 a reviewer destroyed never-committed
# planning artifacts and self-reported the tree restored; nothing
# independent checked the claim until the deferred self-review grepped the
# tree by hand against strings the artifacts were known to contain. This
# guard is that grep, recorded.
#
#   snapshot  run immediately BEFORE the dispatch's Agent call — greps each
#             marker and writes one <path><TAB><marker><TAB><status> line
#             per spec line to <snapshot-file>
#   verify    run when the dispatch's report file exists, BEFORE any report
#             or verdict is read or acted on — recomputes the same lines
#             and compares byte-for-byte
#
# <markers-file> carries one <path><TAB><ERE> line per artifact the
# dispatch must not damage, each ERE a string the artifact is known to
# contain — a decision ID, a task title, a section heading. Paths are
# worktree-relative and are refused before grep ever runs when they are
# absolute or carry a `..` component: the markers file names what grep
# reads, so it is a containment boundary exactly like the change-name case
# block in check-plan-unchanged.sh, and it is validated the same way —
# refused, not sanitised.
#
# MARKERS SIT BESIDE check-plan-unchanged.sh, never in its place: that
# guard asserts git's view of the plan tree only, while markers pin known
# content anywhere in the tree, independent of git entirely — content that
# was never committed (the KAN-579 artifacts) has no git answer at all.
# Both guards' verify exit 1 stops the run the same way, and for the same
# reason: the dispatch's own clean-or-restored prose is not evidence, so
# recovery is by hand from before the dispatch, never by re-running it.
#
# Exit codes:
#   0  snapshot written, or every recorded marker matches its snapshot
#   1  verify found the tree's markers changed since the snapshot — each
#      difference named on stderr
#   2  cannot answer at all — missing arguments, a bad mode, a non-directory
#      worktree, an unreadable markers file, a malformed spec line (no tab,
#      an empty path or marker), a path that is absolute or carries a `..`
#      component, a markers file naming no artifact at all, the snapshot
#      path being an existing directory, a grep failure on an existing
#      file, an unwritable snapshot path (on snapshot), or a missing or
#      unreadable snapshot file (on verify)
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-tree-markers: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-tree-markers 2 "check-tree-markers:" "$@"
