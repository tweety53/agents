#!/usr/bin/env bash
# stage-task-files.sh — stage one /flow task's files for its pathspec-scoped
# commit (skills/flow/implementer-brief.md, FLOW — COMMIT-PER-TASK), and refuse when
# git leaves any of them unstaged.
#
# Usage: stage-task-files.sh <abs-worktree> <file>...
#
# In this order, every git call anchored by `git -C <abs-worktree>`:
#   1. git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/
#      — the clearing pass, before any add, because an exclude governs what
#      an add adds and cannot retract what an earlier step already staged.
#   2. git add -- ':(exclude)spectre/changes/' ':(exclude)openspec/changes/'
#      ':(exclude)docs/superpowers/' <file>... — the excludes lead the
#      paths: git 2.50.1 given the paths first and an exclude after them
#      staged the tracked paths alone, left every new file untracked and
#      exited 0 (kan-1012 self-review, 15 paths).
#   3. git status --porcelain --untracked-files=all -- <file>... — any line
#      whose worktree column is not blank (an unstaged edit or `??`), outside
#      the three planning prefixes, is a task path the commit would miss.
#
# Exit codes:
#   0  every named path git reports a change for is staged
#   1  git add exited 0 and left a named path unstaged — each such status
#      line is printed; commit nothing and report it
#   2  usage, a worktree that is not a directory, no git, or a git call
#      failed — git's own output is printed
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "stage-task-files: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec stage-task-files 2 "stage-task-files:" "$@"
