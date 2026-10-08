#!/usr/bin/env bash
# check-late-fix-trigger.sh — decide whether a fix run's pass 1 takes the
# late-fix reduction (skills/flow/review-panel-late-fix.md).
#
# Usage: check-late-fix-trigger.sh <change> <tasks.md> <worktree> <since-close-sha|-> <base-verdict> \
#          [<worktree> <since-close-sha|-> <base-verdict>...]
#
# One triple per worktree in the resolved set, the canonical worktree's
# first. <since-close-sha> is the sha the stage's last clean close reviewed
# in that worktree, `-` where there was no earlier clean close;
# <base-verdict> is that worktree's check-base-moved.sh line this round.
#
# Every condition must hold to reduce:
#   1  already clean, already verified — check-panel-findings-closed.sh
#      <canonical-worktree> <change> exits 0 (run in process), and every
#      worktree names a since-close sha;
#   2  the base has not moved since that close — every base verdict is CLEAR;
#   3  the delta is small — `git diff --numstat <since-close-sha>` per
#      worktree, insertions and deletions summed across the set, is at most
#      40 changed lines; a binary entry (`-`) cannot be counted and fails;
#   4  no scope growth — `git diff <since-close-sha> -- <tasks.md>` in the
#      canonical worktree adds no line matching `^- \[[ xX]\] [0-9]+\.`;
#      an untracked <tasks.md> fails;
#   5  the panel's own machinery is untouched — no path in
#      `git diff --name-only <since-close-sha>` matches
#      skills/flow/review-panel*.md, scripts/check-panel-*.sh,
#      skills/flow/*-reviewer-prompt.md or skills/flow/engineering-principles.md.
#
# Prints on stdout:
#   exit 0  late-fix reduction: <n> changed lines since <canonical since-close-sha>
#   exit 1  full path: condition <k> — <reason>   (one line per failed condition)
#   exit 3  append scope: <n> changed lines since <canonical since-close sha>
#
# Exit codes:
#   0  reduce
#   1  full path
#   2  cannot answer — usage, a worktree/sha pair that fails
#      panel_validate_worktree, a git failure, or check-panel-findings-closed
#      exit 2 (its stderr passed through); the caller reads it as full path
#   3  append scope — conditions 1, 2 and 5 hold and only 3 and/or 4 failed:
#      pass 1 runs the decided roster on the since-close delta
#      (skills/flow/review-panel-late-fix.md, The append scope)
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-late-fix-trigger: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-late-fix-trigger 2 "check-late-fix-trigger:" "$@"
