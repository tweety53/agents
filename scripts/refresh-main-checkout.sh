#!/usr/bin/env bash
# refresh-main-checkout.sh — fast-forward a main checkout onto origin/<base>
# once a change has landed there, and only when nothing in it can be lost.
#
# WHY. /flow's last step brings the operator's main checkout forward so it
# shows the landed change. Nothing in the pipeline moves the local <base>
# ref any more — every change lands from its own apply worktree, as a push
# of its branch or a forge merge — so there is no stale index to repair,
# only a checkout behind its remote; the old reset path for a ref moved
# under the checkout is gone with the landing worktree that moved it.
#
# Usage:
#   refresh-main-checkout.sh <main-checkout> <base>
#
# In order: the bounded, credential-free fetch resolve-base-branch.sh runs
# (`-c core.askpass=true fetch --quiet origin`, its failure ignored — a
# stale origin/<base> only means a smaller step); then every refusal below,
# checked before anything moves; then `git merge --ff-only origin/<base>`.
#
# Exit codes:
#   0  REFRESH-DONE: <path> fast-forwarded <old> -> <new>
#      REFRESH-CURRENT: <path> already at <tip> (nothing to do)
#   1  REFRESH-REFUSED: <path> <reason> — nothing touched. Reasons: is
#      detached; is on <branch>, not <base>; has tracked changes (any
#      modified, staged or unmerged entry); has no origin/<base>; has <base>
#      commits origin/<base> lacks (local <base> is not an ancestor of
#      origin/<base>)
#   2  usage; or cannot answer — a git read or the merge fails outright
#      (an unborn HEAD), git's own message on stderr
#
# Untracked files are never touched; a fast-forward that would overwrite
# one fails in git itself, exit 2.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "refresh-main-checkout: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec refresh-main-checkout 2 "refresh-main-checkout:" "$@"
