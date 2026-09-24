#!/usr/bin/env bash
# reshape-branch.sh — integrate's reshape: collapse a change branch's task
# and fixup commits back into the working tree while keeping every planning
# commit as the separate commit it was made as.
#
# Usage: reshape-branch.sh <worktree> <name> <merge-base>
#
# Replaces the bare `git reset --soft <merge-base>` integrate once ran, which
# folded every planning commit into the single planning commit the two-commit
# chain then made. skills/flow-contracts/finish-contract-run1.md is canonical
# for the reshape; this script is what runs it.
#
# WHAT IT DOES. A planning commit is any non-merge commit in
# <merge-base>..HEAD touching `<leaf>/changes/` — the pathspec-scoped
# commits of git-boundaries.md's **Planning commits**, link commits included,
# and /flow-plan's capture commit. Task and fixup commits never touch that
# directory (check-task-commit-planning-paths.sh), so the directory's state
# at each planning commit is exactly the merge base's plus the planning
# commits up to it. Each is rebuilt, in order, on top of <merge-base> as a
# commit whose tree is its parent's with `<leaf>/changes/` replaced by that
# planning commit's own — message and author kept, committer now. Then HEAD
# is moved there with `reset --soft`, so the index and working tree keep
# everything they held: the task and fixup work, operator edits at the human
# gate and any uncommitted planning delta, all uncommitted, for
# commit-split.sh to commit on top as the implementation commit and the last
# planning commit.
#
# THE WORKING TREE AND THE REAL INDEX ARE NEVER TOUCHED until the final
# `reset --soft`: the rebuild writes trees through a scratch GIT_INDEX_FILE,
# so a failure anywhere before that leaves the branch exactly as it was.
#
# Planning commits run behind check-planning-commit-location.sh, and this
# script rewrites them, so it calls that guard first and stops on its
# verdict, printing the guard's own lines.
#
# Prints `RESHAPED: <worktree> — <n> planning commit(s) kept on <short sha>`
# and exits 0. Exits 2 with a reason on stderr on bad arguments or a base
# that does not resolve, the guard's own code on a guard refusal, and git's
# own code on a git failure.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/spec-root.sh"

[ "$#" -eq 3 ] || {
  echo "usage: reshape-branch.sh <worktree> <name> <merge-base>" >&2
  exit 2
}
WT="$1" NAME="$2" BASE="$3"

"$SCRIPT_DIR/check-planning-commit-location.sh" "$WT" "$NAME"

base_sha="$(git -C "$WT" rev-parse --verify --quiet "$BASE^{commit}")" || {
  echo "reshape-branch.sh: base does not resolve to a commit: $BASE" >&2
  exit 2
}
plan_dir="$(spec_root_leaf "$WT")/changes"

plans="$(git -C "$WT" rev-list --reverse --no-merges "$base_sha..HEAD" -- "$plan_dir/")"

scratch="$(mktemp -d "${TMPDIR:-/tmp}/reshape-branch.XXXXXX")"
trap 'rm -rf "$scratch"' EXIT
export GIT_INDEX_FILE="$scratch/index"

tip="$base_sha"
kept=0
while IFS= read -r sha; do
  [ -n "$sha" ] || continue
  git -C "$WT" read-tree "$tip"
  git -C "$WT" rm -r -q -f --cached --ignore-unmatch -- "$plan_dir/" >/dev/null
  if git -C "$WT" cat-file -e "$sha:$plan_dir" 2>/dev/null; then
    git -C "$WT" read-tree --prefix="$plan_dir/" "$sha:$plan_dir"
  fi
  tree="$(git -C "$WT" write-tree)"
  tip="$(
    GIT_AUTHOR_NAME="$(git -C "$WT" log -1 --format=%an "$sha")" \
    GIT_AUTHOR_EMAIL="$(git -C "$WT" log -1 --format=%ae "$sha")" \
    GIT_AUTHOR_DATE="$(git -C "$WT" log -1 --format=%aI "$sha")" \
      git -C "$WT" commit-tree "$tree" -p "$tip" -F <(git -C "$WT" log -1 --format=%B "$sha")
  )"
  kept=$((kept + 1))
done <<EOF_PLANS
$plans
EOF_PLANS

unset GIT_INDEX_FILE
git -C "$WT" reset -q --soft "$tip"
printf 'RESHAPED: %s — %d planning commit(s) kept on %s\n' "$WT" "$kept" "${base_sha:0:12}"
