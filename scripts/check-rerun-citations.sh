#!/usr/bin/env bash
# check-rerun-citations.sh — judge whether a fix-round re-review read the fix
# it reports on, from its report file rather than from the parent's reading.
#
# Usage: check-rerun-citations.sh <report> <diff> F<n> [F<n>…]
#
# <report> is the re-running slot's panel-report-<round>-<id>.md; <diff> is
# the diff its dispatch named ([DIFF_PATH]: the round's fix-round-N.diff or
# its slot-delta-<round>-<id>.diff); each F<n> is one finding the slot was
# sent to re-review. The FIX-ROUND SCOPE paragraph
# (skills/flow/review-panel-fix-round.md) requires one verdict line per
# finding:
#
#   verdict: F<n> — fixed — <path>:<line>
#   verdict: F<n> — not fixed — <path>:<line>|none
#
# A leading list marker (`- ` or `* `) and backticks are ignored. A `fixed`
# verdict's <path>:<line> must land inside a hunk of <diff> — <path> one of
# the hunk's file headers (`+++ b/<path>` or `--- a/<path>`), <line> inside
# that hunk's new-side or old-side range. A worktree prefix
# (`gymie-frontend:src/Foo.tsx:42`, the WORKTREES paragraph's form) and a
# leading `./` are stripped before the match. A `--- `/`+++ ` line inside an
# open hunk is a removed `-- x` or added `++ x` body line, never a header. A `not fixed` verdict's citation is not
# checked: it keeps the finding open, the safe direction.
#
# WHY THIS EXISTS (kan-964, fix round 1): a sonnet/low re-review reported
# "all fixed" without reading a hunk; only the parent's judgement caught it,
# and sent back once it raised F17. A citation that must resolve in the diff
# makes that skip detectable mechanically.
#
# Prints one line per named finding to stdout:
#   CITED: F<n> — <verdict> — <citation>
#   UNCITED: F<n> — <why>
#
# The logic is the Go port in stats/internal/guard/reruncitations.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
#
# Exit codes: 0 every named finding carries a verdict line, every `fixed`
# one citing a hunk of <diff>; 1 at least one UNCITED line; 2 cannot judge
# (a missing argument, an unreadable report or diff, an argument not shaped
# F<n>), with the cause on stderr.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-rerun-citations: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-rerun-citations 2 "check-rerun-citations:" "$@"
