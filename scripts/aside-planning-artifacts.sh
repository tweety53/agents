#!/usr/bin/env bash
# aside-planning-artifacts.sh — set the planning paths aside before a
# pipeline rebase, restore them after (KAN-628, observed on kan-361 — a fix
# round's autosquash had to cross uncommitted design.md/tasks.md edits — and
# again on kan-579, where uncommitted planning edits from
# flow.write-in-progress blocked the pre-landing rebase. Both runs
# improvised the same temporary-WIP-commit dance; this helper is what the
# runs cite instead).
#
# Usage:
#
#   aside-planning-artifacts.sh aside <worktree>
#   aside-planning-artifacts.sh restore <worktree>
#
# The planning paths are the spec tree's changes directory — the leaf
# resolved per project through scripts/lib/spec-root.sh, `spectre` or
# `openspec`, never hardcoded — and `docs/superpowers/`, the same two paths
# check-task-commit-planning-paths.sh sweeps for. Each enters the pathspec
# only when the directory exists in <worktree>: a pathspec naming an absent
# directory would have git refuse the stash outright. `aside` stashes the
# paths' uncommitted state — tracked modifications and untracked files
# alike, `git stash push --include-untracked` — under a message carrying the
# literal marker `aside-planning-artifacts`, and nothing else: dirty
# implementation paths are left exactly where they are, for
# check-unfinished-work.sh and the guards that own them, and a worktree
# whose planning paths are already clean is reported and left untouched.
#
# `restore` refuses outright while a rebase, merge, cherry-pick or am is
# still in progress — a stash popped over an unresolved sequencer state
# would fold four kinds of confusion into one worktree. Once the rebase has
# finished or aborted, it pops the top stash only when its message carries
# the marker: an operator's own stash is never popped, and one pushed after
# the aside shadows the helper's entry, so restore reports NONE and leaves
# every entry in place — `git stash list` is always the recovery path. A
# conflicted apply (the rebase moved a planning file the aside also changed)
# is reported with exit 1 and the stash kept, which git itself does on a
# conflicted pop.
#
# Verdict lines, on stdout:
#
#   PLANNING-ARTIFACTS-CLEAN: <worktree> — nothing to set aside
#   PLANNING-ARTIFACTS-ASIDE: <worktree> — <short sha>
#   PLANNING-ARTIFACTS-RESTORED: <worktree> — <short sha>
#   PLANNING-ARTIFACTS-CONFLICT: <worktree> — <short sha> kept
#   PLANNING-ARTIFACTS-NONE: <worktree> — no aside stash on top
#
# Exit 0 on every answered verdict; exit 1 on PLANNING-ARTIFACTS-CONFLICT;
# exit 2 with NOTHING on stdout when the arguments are missing or unknown,
# <worktree> is not a git repository, git refuses an operation, or restore
# is called with a rebase/merge/cherry-pick/am still in progress — an
# inability is never reported as a verdict.
#
# The guard is Go: stats/internal/guard/asideplanningartifacts.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "aside-planning-artifacts.sh: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec aside-planning-artifacts 2 "aside-planning-artifacts.sh:" "$@"
