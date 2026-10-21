#!/usr/bin/env bash
# check-panel-diff-size.sh — report whether a branch's diff is within the
# review panel's reading cap.
#
# Usage: check-panel-diff-size.sh <worktree> <merge-base> [cap]
#
# Prints two lines to stdout: the changed-line count, then the verdict
# ("under cap" or "over cap") — so a caller can quote either line rather
# than re-derive the number.
#
# Exit codes: 0 at or under the cap; 1 over the cap; 2 cannot answer at all.
#
# WHY THE CAP IS 5000. KAN-109's own measurement found a seven-slot review
# panel costing 948k tokens to read a 9877-line diff — about 96 tokens per
# diff line across that whole panel, or roughly 14 per line per slot. The
# cap was 2000 on that basis, chosen as roughly a fifth of the measured
# diff.
#
# The operator raised it to 5000 on 2026-08-28, during KAN-341, whose own
# cross-repo diff ran past 2000 and whose panel they chose to run anyway.
# Raising the default records that choice once rather than re-answering the
# same over-cap prompt on every change of that size. On the rate above, a
# four-slot panel at 5000 lines projects to roughly 275k tokens — the same
# order the 2000 cap was set to protect — while a seven-slot panel at that
# length would be about 480k, so a large roster against a full cap is still
# worth a second thought even though the guard now permits it.
#
# That projection is arithmetic from KAN-109's one measurement, not a second
# measurement: nothing has re-measured a real panel at 5000 lines.
#
# WHY BOTH COMMITTED WORK AND THE WORKING TREE COUNT. `final-review.diff`,
# the text the panel actually reads, carries everything that changed since
# the merge base: commits already made plus whatever is still staged or
# unstaged in the working tree. Measuring only `merge-base..HEAD` would
# report zero for a run whose work sits entirely uncommitted, understating
# the exact diff the panel is about to be handed.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-panel-diff-size: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-diff-size 2 "check-panel-diff-size:" "$@"
