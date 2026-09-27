#!/usr/bin/env bash
# check-task-commit-planning-paths.sh — fail any task commit that swept the
# planning paths (KAN-553, observed twice on kan-468: staged planning
# artifacts reached the task commit through the implementer's own `git
# commit`, and the second `git reset --hard` recovery destroyed uncommitted
# planning work alongside them).
#
# Usage:
#
#   check-task-commit-planning-paths.sh <worktree> <base>
#
# A "task commit" is a commit in `<base>..HEAD` carrying the pipeline's
# `Task-Id:` trailer — the commit an implementer makes for a plan task.
# The planning paths are the spec tree's changes directory — the leaf
# resolved per project through spec-root.sh's Go twin
# (stats/internal/guard/specroot.go), `spectre` or `openspec`, never
# hardcoded — and `docs/superpowers/`. Anything else
# under the spec tree is implementation: a capability spec under
# `<leaf>/specs/` is exactly what a task commit SHOULD carry, and
# `docs/superpowers/` is named in full rather than as `docs/` for the same
# reason. The trailerless planning commits — git-boundaries.md's **Planning
# commits** and the one bare /flow makes at integrate — carry exactly these
# paths and stay outside this contract by having no trailer; so does
# everything at or before <base>. A merge task commit is
# judged by its combined diff (`-c`): a planning path its result carries
# that no parent had is what the merge itself introduces and sweeps; content
# carried over unchanged from a parent stays that parent commit's, judged
# by its own walk entry.
#
# Verdict lines, on stdout:
#
#   TASK-COMMIT-SWEEP: <short sha> <task id> <path>   per planning path a
#                                                     task commit touched
#   PLANNING-PATHS-CLEAN: <worktree> — <n> task commit(s) checked
#   PLANNING-PATHS-SWEPT: <worktree> — <n> task commit(s)
#
# Exit 0 on the clean verdict; exit 1 on the swept verdict; exit 2 with
# NOTHING on stdout when the arguments are missing, the worktree is not a
# readable git repository, HEAD or <base> does not resolve, or git refuses
# the walk — an inability is never reported as a verdict.
#
# Like the other guards that need a change in flight and a real worktree
# passed in (check-foreign-staged.sh and its siblings), this one answers a
# question about one change, not about the state of the repository's text,
# and is deliberately not a `## lint` step; its tests in
# stats/internal/guard/check_task_commit_planning_paths_test.go are what
# `## test` runs.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays.
#
# The guard runs as the Go port in
# stats/internal/guard/taskcommitplanningpaths.go. flow-guard is built from
# this checkout, never taken from PATH: scripts/lib/flow-guard.sh derives it,
# and exits 2 (this guard's cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-task-commit-planning-paths: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-task-commit-planning-paths 2 "check-task-commit-planning-paths:" "$@"
