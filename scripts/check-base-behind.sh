#!/usr/bin/env bash
# check-base-behind.sh — compare each worktree's recorded base with the branch
# origin tracks, before a stale base turns into failures (kan-1007: a feature
# base behind develop was papered over by merging develop into the change
# branches, and integrate's reshape folded KAN-1010 into the squash).
#
# Usage: check-base-behind.sh <worktree>...
#
# Per worktree: the recorded base is `branch.<current branch>.flowBase`; the
# tracked branch is the branch `refs/remotes/origin/HEAD` names. Once a base is
# recorded, the guard runs `git fetch -q origin`; a failing fetch is ignored and
# the refs already on hand are compared.
#
# Prints to stdout, one block per worktree in argument order:
#   BASE-CURRENT: <wt> — no recorded base
#   BASE-CURRENT: <wt> — <base> is the tracked branch
#   BASE-CURRENT: <wt> — origin/<base> holds every commit of origin/<tracked>
#   BASE-BEHIND: <wt> — origin/<base> lacks <n> commits of origin/<tracked>; sync the base, never merge <tracked> into the change branch (kan-1007):
#     T=$(git -C <wt> merge-tree --write-tree origin/<base> origin/<tracked>) && C=$(git -C <wt> commit-tree "$T" -p origin/<base> -p origin/<tracked> -m 'Merge <tracked> into <base>') && git -C <wt> push origin "${C}:refs/heads/<base>"
#     git -C <wt> fetch -q origin && sync-onto-base.sh <wt> origin/<base> <merge-base>
# where <merge-base> is the sha `git merge-base HEAD origin/<base>` prints. The
# first command line syncs the base with no checkout (merge-tree exits 1 on a
# conflict and stops the chain); its push runs unasked (operator-prompts-auto-resolution.md).
#
# Exit 0 no BASE-BEHIND, 1 at least one, 2 cannot answer (no argument, a path
# that is not a git worktree, no origin/HEAD, a failing rev-list or
# merge-base) — stdout empty, the cause on stderr.
#
# Go, in flow-guard (stats/internal/guard/basebehind.go).
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-base-behind: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-base-behind 2 "check-base-behind:" "$@"
