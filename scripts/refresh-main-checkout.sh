#!/usr/bin/env bash
# refresh-main-checkout.sh — bring a main checkout's index and worktree back
# onto the branch pointer a landing moved under it, and only then.
#
# WHY. Run 2 positions the landing worktree on <base> with `worktree add
# --force` (prepare-archive-branch.sh step 2b), fast-forwards it and merges
# the archive branch there. `refs/heads/<base>` moves; the main checkout's
# HEAD names that ref, but its index and worktree still hold the tree of
# the old tip. `git status` there then shows the landing in reverse as
# staged changes — the "reversal artifact" every stash labelled since
# kan-487 — and a session that trusts it commits, stashes or resets the
# ghost. This script closes that gap right after the landing, and refuses
# whenever the checkout holds anything that is NOT that ghost.
#
# Usage:
#   refresh-main-checkout.sh <main-checkout> <base>
#
# Exit codes:
#   0  REFRESH-DONE: <path> index was the tree of <old-tip>; now at <tip>
#      REFRESH-CURRENT: <path> already at <tip> (nothing to do)
#   1  REFRESH-REFUSED: <reason> — nothing touched. Reasons: detached HEAD;
#      on a branch other than <base>; unstaged changes (staleness never
#      produces any — the worktree and index go stale together); an index
#      tree that is not the tree of any of the last 300 commits reachable
#      from HEAD (real staged work, not staleness)
#   2  usage; or cannot answer — a git read or the reset fails outright
#      (an unmerged index, an unborn HEAD), git's own message on stderr
#
# The reset is `git reset --hard`, which never touches untracked files.
# The proof it is safe: the index tree is byte-identical to a commit the
# branch has already moved past, and the worktree is identical to the index,
# so there is no edit anywhere in the checkout that the reset could lose.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "refresh-main-checkout: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec refresh-main-checkout 2 "refresh-main-checkout:" "$@"
