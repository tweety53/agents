#!/usr/bin/env bash
# sync-onto-base.sh — sync a finishing worktree onto its base's tip, in a
# guard instead of the hand-typed recipe skills/flow/sync-onto-base.md
# carried (KAN-860).
#
# Usage:
#
#   sync-onto-base.sh <worktree> <base-ref> <merge-base>
#   sync-onto-base.sh --resume <worktree> <base-ref> <onto>
#
# <base-ref> is the base branch, bare (main) or remote-tracking
# (origin/main), resolved exactly as check-base-moved.sh resolves it; the
# caller has already fetched (resolve-base-branch.sh), and this guard fetches
# nothing. Every check-base-moved.sh line it reaches is echoed on stdout.
#
# The first form runs check-base-moved.sh against <merge-base>:
#   CLEAR  -> `CLEAN: <worktree> — nothing to rebase`, exit 0.
#   MOVED  -> the planning paths are set aside (aside-planning-artifacts.sh),
#             the worktree is rebased onto the base's tip — resolved to a sha
#             first, so a fetch landing mid-run cannot move it — and the
#             aside is restored. check-base-moved.sh then re-runs against
#             that sha; while it answers MOVED (the base moved during the
#             rebase) the rebase runs again, at most 3 rebases in all. Then
#             one line per path any MOVED verdict named under `overlaps:`,
#             sorted:
#               GUARD-TEST: <path> — <agents repo>/scripts/test-<stem>.sh
#               NO-GUARD-TEST: <path>
#             (<stem> is the path's basename without its extension — the
#             guard test scripts/run-guard-tests.sh discovers by glob), and
#             `REBASED: <worktree> — onto <sha>`, exit 0. The caller runs each
#             named test; this guard runs none.
#   A rebase that stops on a conflict prints
#             `CONFLICT: <worktree> — onto <sha>; unmerged: <paths>` and exits
#             1, the worktree left mid-rebase exactly as git left it and the
#             aside still set aside — resolve in place, then --resume. The
#             aside's stash sha (or none) is recorded in the worktree's own
#             git dir (`git rev-parse --git-path flow-sync-aside`).
#
# --resume runs once an in-place resolution has finished the rebase: it
# refuses while a rebase is still in progress and unless <onto> — the sha the
# CONFLICT line named — is an ancestor of HEAD, restores the recorded aside
# only while that stash is on top (the stash list is shared by every
# worktree; another worktree's aside is never popped), and runs
# the same re-check loop. It prints no GUARD-TEST lines: a resolved rebase
# runs the project's whole lint and test lists instead.
#
# Exit 0 CLEAN or REBASED; 1 CONFLICT; 2 everything else — a REFUSE verdict,
# a question it cannot answer, a rebase git refused to start (nothing
# changed, the aside restored), a base still moving after 3 rebases, or an
# aside it cannot restore (git stash list holds it). Exit 2 means stop and
# ask.
#
# The guard is Go: stats/internal/guard/syncontobase.go, over the shared
# baseMoved verdict (basemoved.go) and rebaseOntoTip core (baserebase.go).
# The binary does not live in this checkout, so this shim exports
# FLOW_GUARD_REPO_ROOT — the agents repository, derived through lib/'s
# physical path so a skill's symlink to this shim still answers the
# repository and never the skill directory — for the GUARD-TEST lookup.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "sync-onto-base: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec sync-onto-base 2 "sync-onto-base:" "$@"
