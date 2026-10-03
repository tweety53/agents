#!/usr/bin/env bash
# fold-fixup.sh — fold a fix into the unpushed task commit it targets
# (skills/flow/review-panel-fix-round.md, "Rewrite-based folding is for
# unpushed history only").
#
# Usage: fold-fixup.sh <worktree> <task-sha> <tasks-md> <path>...
#        fold-fixup.sh --finish <worktree> <task-sha> <tasks-md>
#
# The fold, in order: `git add -- <path>...`; `guard-autosquash.sh targets
# <worktree> <task-sha>`; `git commit --fixup=<task-sha> -- <path>...`;
# aside-planning-artifacts aside; `GIT_SEQUENCE_EDITOR=: git rebase -i
# --autosquash <task-sha>^` (git ignores a non-interactive --autosquash);
# the empty-fold drop; aside-planning-artifacts restore; `guard-autosquash.sh
# after <worktree> <task-sha>^ <tasks-md>`. The rebase's upstream is the task
# commit's own parent, never the base branch re-resolved: replaying only the
# branch's own commits after <task-sha>^ keeps commits that landed on the base
# mid-panel out of the fold. The `after` check's base is that same upstream.
#
# THE EMPTY-FOLD DROP. A fixup whose fold empties its target commit is
# dropped, never kept as a no-op commit: git stops the rebase at that fixup
# ("would make it empty"), and the guard drops the emptied commit with
# `git reset --soft HEAD^` and continues — tip or not, the commits after it
# replay onto its parent. The drop runs while the planning paths are still
# set aside, so no reset reaches them.
#
# A CONFLICT IS LEFT IN PROGRESS, never resolved here: it is between two of
# the branch's own commits, resolved by hand keeping both sides, then
# `git rebase --continue`, then `fold-fixup.sh --finish` — which settles any
# further empty-fold stop, restores the aside and runs the `after` check.
# The aside is restored by its own stash sha, and only while that stash is
# on top (the stash list is shared by every worktree): a conflicted fold
# records the sha, or none, at `git rev-parse --git-path flow-fold-aside`
# for --finish to read.
#
# Prints one line on stdout:
#   FOLDED:   <worktree> — HEAD <sha>
#   DROPPED:  <worktree> — the fold emptied <task-sha>; HEAD <sha>
#   CONFLICT: <worktree> — unmerged: <paths>; resolve keeping both sides, …
#
# Exit codes:
#   0  folded, or the emptied commit dropped
#   1  guard-autosquash.sh refused (its lines on stderr) — the round stops
#   2  cannot answer — usage, a git refusal, an aside that cannot be set or
#      restored, or a rebase stop that is neither a conflict nor an empty fold
#   3  a conflict is left in progress (CONFLICT line)
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "fold-fixup: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_AUTOSQUASH="$SCRIPT_DIR/guard-autosquash.sh"
export FLOW_GUARD_AUTOSQUASH
flow_guard_exec fold-fixup 2 "fold-fixup:" "$@"
