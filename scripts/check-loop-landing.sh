#!/usr/bin/env bash
# check-loop-landing.sh <repo> <branch> — may a pipeline-fix-* or
# self-review-* branch land on the default branch?
#
# Only once flowd holds, under the loop's change (its branch name), a
# completed plan dispatch, a completed first fix (`fix` or `fix-g<g>`), at
# least one review, every row closed, the highest-numbered review closed
# `clean` — no Critical, no Important — and no `fix-v<v>` after it. A
# loop branch whose repository's project holds no rows for it — a
# satellite, the loop's branch in another repository it carried a task into —
# is read under the project of the agents checkout this script sits in, where
# a loop run from that checkout files them; a FINISHED loop's rows still
# admit its satellites, which land after its own branch. A self-review-*
# branch whose diff from origin/HEAD touches only docs/prompt-tuning/ lands
# without rows (LOOP-LANDING-OK … (prompt-tuning cases only)): it carries the
# self-review's prompt-tuning cases and no code, so no review had anything to
# judge.
# hooks/protect-main-checkout.py runs it on a landing push typed by hand;
# pipeline-fix.sh land runs the same check before its own push.
#
# Exit codes: 0 LOOP-LANDING-OK, or LOOP-LANDING-NA for any other branch;
# 1 one LOOP-LANDING-REFUSED line per reason; 2 cannot answer (flowd
# unreadable, not a git checkout, usage).
#
# Go, in flow-guard (stats/internal/guard/looplanding.go). flow-guard is
# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-loop-landing: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-loop-landing 2 "check-loop-landing:" "$@"
