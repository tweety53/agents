#!/usr/bin/env bash
# check-untasked-commits.sh — list the commits of a change branch that sit
# outside every task, for /flow's run-1 unfinished-work gate.
#
# Usage: check-untasked-commits.sh <worktree> <base>
#
# WHY THIS GUARD EXISTS. A commit the run makes outside every task — a stray
# implementer edit, a parent's unrecorded fix — was caught only by a review
# panel reading the log (kan-982 panel F3: "Two non-chore commits sit outside
# every task"). Run 1's reshape then squashes it into the implementation
# commit, and the evidence is gone.
#
# Every non-merge commit on the first-parent chain of <base>..HEAD is listed
# — a merged-in branch's commits are that branch's own, never the change's
# (kan-941: a develop merge listed develop's commits) — unless it carries a
# non-empty `Task-Id:` trailer (a task's own commit), a non-empty `Flow-Key:`
# trailer (a commit a run makes outside every task on a sanctioned route —
# **Commits outside every task**, skills/flow-contracts/git-boundaries.md),
# has a `chore` subject (the planning, link and archive commits), or touches
# only paths under the worktree's `<spec-root>/changes/` (a plan amendment).
#
# Prints one verdict line to stdout:
#   UNTASKED-OK: <n> commit(s) checked                      nothing listed — exit 0
#   UNTASKED: <k> commit(s) carry no Task-Id: or Flow-Key: trailer
#     <sha12> <subject>                                      one line per commit — exit 1
# Exit 2 with no verdict line when it cannot answer: a usage error, a
# worktree that is not a directory, or a range git cannot read.
#
# Go, in flow-guard (stats/internal/guard/untaskedcommits.go). flow-guard is
# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 (this guard's cannot-answer code) with the cause
# when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-untasked-commits: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-untasked-commits 2 "check-untasked-commits:" "$@"
