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
#   2  usage
#
# The reset is `git reset --hard`, which never touches untracked files.
# The proof it is safe: the index tree is byte-identical to a commit the
# branch has already moved past, and the worktree is identical to the index,
# so there is no edit anywhere in the checkout that the reset could lose.
set -euo pipefail

usage() { echo "usage: refresh-main-checkout.sh <main-checkout> <base>" >&2; }
[ "$#" -eq 2 ] || { usage; exit 2; }
REPO="$1"
BASE="$2"
[ -d "$REPO" ] || { usage; exit 2; }

g() { git -C "$REPO" "$@"; }
refuse() { echo "REFRESH-REFUSED: $REPO $1 — nothing touched"; exit 1; }

CUR="$(g symbolic-ref -q --short HEAD 2>/dev/null || true)"
[ -n "$CUR" ] || refuse "is detached"
[ "$CUR" = "$BASE" ] || refuse "is on $CUR, not $BASE"
g diff --quiet || refuse "has unstaged changes"

INDEX_TREE="$(g write-tree)"
TIP="$(g rev-parse --short HEAD)"
if [ "$INDEX_TREE" = "$(g rev-parse 'HEAD^{tree}')" ]; then
  echo "REFRESH-CURRENT: $REPO already at $TIP"
  exit 0
fi

MATCH=""
for c in $(g rev-list --max-count=300 --skip=1 HEAD); do
  if [ "$(g rev-parse "$c^{tree}")" = "$INDEX_TREE" ]; then MATCH="$c"; break; fi
done
[ -n "$MATCH" ] || refuse "has staged changes that match no recent $BASE tip"

g reset -q --hard
echo "REFRESH-DONE: $REPO index was the tree of $(g rev-parse --short "$MATCH"); now at $TIP"
