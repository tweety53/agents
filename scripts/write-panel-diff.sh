#!/usr/bin/env bash
# write-panel-diff.sh — write one review-panel diff and its touched list
# (skills/flow/review-panel.md).
#
# Usage:
#   write-panel-diff.sh final      <canonical-wt> <wt> <merge-base> [<wt> <merge-base>...]
#   write-panel-diff.sh late-fix   <canonical-wt> <wt> <since-close-sha> [...]
#   write-panel-diff.sh fix-round <N> <canonical-wt> <wt> <fix-base> [...]
#   write-panel-diff.sh slot-delta <round> <slot> <canonical-wt> <wt> <merge-base> <held-sha|-> [...]
#
# Writes two files into <canonical-wt>/.superpowers/sdd/, each atomically
# (a temp file renamed over the target), and prints both paths, one per line:
#   final-review.diff | late-fix.diff | fix-round-<N>.diff |
#   slot-delta-<round>-<slot>.diff
#   <that file>.touched — the [TOUCHED_FILES] list
#
# Both are sectioned per worktree, in argument order, each section opened by
#   # worktree: <wt> — merge base <sha>
# naming the worktree's own second argument (its merge base, since-close sha
# or fix base). final-review.diff alone opens with the semantics line
#   # final-review.diff — working tree vs merge-base, unstaged changes included
# Each diff section is that worktree's `git diff` over the kind's range, and
# each .touched section `git diff --name-status` over the same range:
#   final       git diff <merge-base>          (the working tree: staged and unstaged)
#   late-fix    git diff <since-close-sha>     (the working tree)
#   fix-round   git diff <fix-base>..HEAD
#   slot-delta  git diff <held-sha> HEAD; a `-` held sha — a worktree in
#               which the slot holds no sha — reads git diff <merge-base>
#
# Exit codes:
#   0  both files written
#   2  cannot answer — a usage error, a worktree/sha pair that fails
#      panel_validate_worktree (scripts/lib/panel-touched-paths.sh), or a git
#      failure; nothing is written
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "write-panel-diff: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec write-panel-diff 2 "write-panel-diff:" "$@"
