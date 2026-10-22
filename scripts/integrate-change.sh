#!/usr/bin/env bash
# integrate-change.sh — the mechanical half of a bare /flow at IN_PROGRESS:
# run 1 (integrate) and run 2 (cleanup), sequenced through the existing
# guards.
#
# Usage:
#   integrate-change.sh prepare <main-checkout> <name> [--accept-foreign] [--accept-outstanding]
#   integrate-change.sh commit  <main-checkout> <name> [<worktree>=<impl subject>]... [--outstanding <text>]
#   integrate-change.sh land    <main-checkout> <name> [--route <route>] [--pr-url <worktree>=<url|none>]...
#   integrate-change.sh cleanup <main-checkout> <name> [--proceed <main-checkout>]...
#
# Go, in flow-guard: integrateChange's doc comment in
# stats/internal/guard/integratechange.go is the contract — every
# subcommand, its flags, the NEXT:/STOP:/JIRA: closing line and exit 0/1/2.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 with the cause when it
# cannot. FLOW_GUARD_SCRIPT_DIR is exported so the Go execs every guard it
# sequences from beside this script, as the bash did.
set -uo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "integrate-change: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
missing=""
for g in check-base-moved check-cleanup-complete check-finish-preflight check-foreign-staged \
  check-main-checkout-drift check-unfinished-work check-untasked-commits check-visual-verify-dispatched commit-archive \
  commit-split land-self-review-report migrate-worktrees project-get refresh-main-checkout \
  remove-change-worktrees reshape-branch resolve-base-branch run-verify-list sync-onto-base; do
  [ -x "$SCRIPT_DIR/$g.sh" ] || missing="$missing $g.sh"
done
[ -z "$missing" ] || { printf 'integrate-change: %s\n' "missing beside this script:$missing — run the procedure by hand" >&2; exit 2; }
: "$SCRIPT_DIR/check-base-moved.sh" "$SCRIPT_DIR/check-cleanup-complete.sh" "$SCRIPT_DIR/check-finish-preflight.sh"
: "$SCRIPT_DIR/check-foreign-staged.sh" "$SCRIPT_DIR/check-main-checkout-drift.sh" "$SCRIPT_DIR/check-unfinished-work.sh"
: "$SCRIPT_DIR/check-untasked-commits.sh" "$SCRIPT_DIR/check-visual-verify-dispatched.sh" "$SCRIPT_DIR/commit-archive.sh"
: "$SCRIPT_DIR/commit-split.sh" "$SCRIPT_DIR/land-self-review-report.sh" "$SCRIPT_DIR/migrate-worktrees.sh"
: "$SCRIPT_DIR/project-get.sh" "$SCRIPT_DIR/refresh-main-checkout.sh" "$SCRIPT_DIR/remove-change-worktrees.sh"
: "$SCRIPT_DIR/reshape-branch.sh" "$SCRIPT_DIR/resolve-base-branch.sh" "$SCRIPT_DIR/run-verify-list.sh"
: "$SCRIPT_DIR/sync-onto-base.sh" "$SCRIPT_DIR/prepare-workspace.sh" "$SCRIPT_DIR/check-loop-landing.sh"
: "$SCRIPT_DIR/check-narrative-runs.sh"
FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR" FLOW_GUARD_CALLER_DIR="$PWD"
export FLOW_GUARD_SCRIPT_DIR FLOW_GUARD_CALLER_DIR
# Outside every worktree: remove-change-worktrees' check 6 counts this
# process's own cwd. Relative arguments resolve against FLOW_GUARD_CALLER_DIR.
cd / || { echo "integrate-change: cannot cd /" >&2; exit 2; }
flow_guard_exec integrate-change 2 "integrate-change:" "$@"
