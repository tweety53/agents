#!/usr/bin/env bash
# check-review-gate.sh — judge the per-task review gate of skills/flow/implement.md
# (**The review gate.**) for one task commit check-task-commit-fields.sh passed.
#
# Usage: check-review-gate.sh <worktree> <task-id> <commit-sha|worktree=sha[,worktree=sha…]>
#          <canonical-worktree> <change-name> [refused-path…]
#
# The first five arguments are check-task-commit-fields.sh's, less its empty
# parent placeholder, and the plan resolves the same way (one shared helper,
# so the two cannot drift). The commit map sums changed lines and unions paths
# across its repositories. Each trailing refused path is one the fields guard
# named undeclared before a Files: widening was transcribed: it stays
# undeclared, so the widening cannot disarm the gate.
#
# The gate fires when the commit changes more than 40 lines (inserted plus
# deleted, `git diff --no-renames --numstat`, a binary file's `-` counted as
# 0) or touches any path outside the task's declared set — its Files: plus
# its Allowed-collateral: globs, widened by a squash fold. The 40 lives in
# stats/internal/guard/reviewgate.go, the one site to re-tune it.
#
# Prints ONE verdict line to stdout:
#   FIRE: task <id> — <n> changed lines (more than 40)[; undeclared paths: <p>, …]
#   FIRE: task <id> — undeclared paths: <p>, …
#   QUIET: task <id> — <n> changed lines, every path declared
#
# Exit 0 a verdict was reached; exit 2 it could not judge (an unreadable
# plan, a task it cannot find, a commit git cannot resolve, a usage error),
# with "check-review-gate: COULD NOT JUDGE —" on stderr and nothing on stdout.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-review-gate: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-review-gate 2 "check-review-gate:" "$@"
