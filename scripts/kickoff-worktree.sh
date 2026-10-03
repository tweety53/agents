#!/usr/bin/env bash
# kickoff-worktree.sh — /flow's kickoff steps 1–5 (skills/flow/brainstorm.md
# **A**), run in order in one call so no session re-types them.
#
# Usage: kickoff-worktree.sh <project> <name> [<base>]
#
# <base>, when given, names the branch the change is cut from and lands on
# instead of origin/HEAD's: origin/<base> must exist (exit 1 otherwise,
# nothing added), a fresh branch is created from it, and either form records
# it as `git config branch.spectre/<name>.flowBase <base>`, which
# resolve-base-branch.sh reads ahead of origin/HEAD. A name failing
# resolve-base-branch's validation is exit 2.
#
#   0. `flow` must be on PATH, checked first: a worktree step 3 could not
#      persist is never created.
#   1. check-worktree-location.sh <project> — its exit 1 or 2 stops the run
#      with its own lines, as this script's exit 1 or 2.
#   2. `git check-ignore -q .worktrees` from <project>; where it exits
#      non-zero, `.worktrees/` is appended to the repository's info/exclude
#      (once) — never a commit on any branch.
#   3. `git fetch origin`; when origin/spectre/<name> exists, `git worktree
#      add <project>/.worktrees/<name> spectre/<name>`, a local branch
#      tracking it, otherwise `-b spectre/<name> origin/<default-branch>`,
#      the default branch read from refs/remotes/origin/HEAD, never HEAD —
#      or origin/<base> when <base> is given.
#      The merge base is `git -C <worktree> rev-parse HEAD` right after the
#      add, persisted before anything else runs by `flow state add-worktree
#      -C <project> <name> <worktree> <merge-base>`.
#   4. project-get.sh <worktree> "worktree setup": exit 0 runs every line
#      inside the body's ``` fences, in order, from the worktree root, in the
#      foreground — never the fence markers or the prose around them; exit 1
#      (no such section) is one stderr line and the run continues; exit 2
#      stops the run.
#   5. `git -C <worktree> push -u origin spectre/<name>`.
#
# On success stdout carries exactly:
#   worktree: <project>/.worktrees/<name>
#   merge-base: <sha>
#
# Exit codes: 0 all five steps done; 1 a step stopped the run — the location
# guard's refusal, a failed git call, `flow state add-worktree` refusing (no
# STARTED record), a setup command's non-zero exit (named, with its output)
# or a failed push — its lines relayed on stderr, and the worktree already
# persisted when step 3 got that far; 2 usage, `flow` absent from PATH, no
# resolvable default branch, the location guard's or project-get's own
# cannot-answer exit, or a sibling this shim sets unset.
#
# Go, in flow-guard (stats/internal/guard/kickoffworktree.go). flow-guard is
# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "kickoff-worktree: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_WORKTREE_LOCATION="$SCRIPT_DIR/check-worktree-location.sh"
FLOW_GUARD_PROJECT_GET="$SCRIPT_DIR/project-get.sh"
export FLOW_GUARD_WORKTREE_LOCATION FLOW_GUARD_PROJECT_GET
flow_guard_exec kickoff-worktree 2 "kickoff-worktree:" "$@"
