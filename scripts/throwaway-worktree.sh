#!/usr/bin/env bash
# throwaway-worktree.sh — create or remove the detached throwaway copy of a
# worktree that a wave group's implementer (skills/flow/sdd-dispatch.md) or a
# review-panel slot (skills/flow/review-panel-optional-slots.md) works in.
#
# Usage: throwaway-worktree.sh create <worktree> <copy> [--sdd]
#        throwaway-worktree.sh remove <worktree> <copy> [--fold-back]
#
# create: `git worktree add --detach <copy> HEAD`, then the worktree's
#   uncommitted diff against HEAD (`diff HEAD --binary`) applied in the copy,
#   then every untracked (`??`) entry of `status --porcelain -z` copied with
#   `cp -a`. --sdd also copies `<worktree>/.superpowers/sdd`, when present,
#   to the same path in the copy.
# remove: `git worktree remove --force <copy>`. --fold-back first copies each
#   non-empty `.superpowers/sdd/panel-report-*.md` and
#   `.superpowers/sdd/reproducers/*.sh` from the copy back into the worktree,
#   unless the worktree already holds a non-empty file the copy's is not
#   newer than.
#
# Prints nothing on success.
# Exit 0 done; exit 2 usage, or a step failed — one stderr line names it, and
# no later step ran.
#
# The logic lives in stats/internal/guard/throwawayworktree.go.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "throwaway-worktree: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec throwaway-worktree 2 "throwaway-worktree:" "$@"
