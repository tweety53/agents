#!/usr/bin/env bash
# commit-split.sh — the guarded two-commit chain, in one place.
#
# Usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>
#
# Wraps the exact guarded chain specified in skills/flow-contracts/pipeline.md's
# "Git boundaries" section — that section is the canonical spec of this
# behavior and is not rewritten here; this script only exists to give the two
# call sites (`myflow-do`'s PR-exception path and `myflow-finish`'s run 1) one
# place to call instead of each spelling the chain out inline.
#
# Clears the two planning paths from the index, stages everything else, and
# commits it as the implementation commit — skipped, not failed, if nothing
# is staged. Then stages the planning paths (and anything else left) and
# commits that as the planning commit, skipped the same way if empty.
#
# THE PLANNING PATH INSIDE THE SPECTRE TREE IS `spectre/changes/`, NOT
# `spectre/`. A capability spec under `spectre/specs/` is implementation —
# the implementer writes and commits it on the change branch, in the task
# commit that implements the requirement — so it must fall on the
# implementation side of this split, and widening either pathspec back to
# `spectre/` would classify it as planning again. See
# skills/flow-contracts/git-boundaries.md, canonical for that boundary.
#
# A CHANGE DIRECTORY'S `link.md` IS ALSO IMPLEMENTATION, CARVED OUT BY FILE
# RATHER THAN BY DIRECTORY. It lives at `<plan_dir><id>/link.md`, inside the
# planning exclusion, so it needs its own `git add` after the exclude-add —
# a `:(exclude)` pathspec always wins over a positive pathspec alongside it
# in the same call, so a combined `add -A -- . ":(exclude)$plan_dir"
# "$plan_dir*/link.md"` would still exclude it (verified by running git, not
# by reasoning about it). The re-add runs after the reset above, not before,
# so a later reset never strips it back out. `git add -A -- <pathspec>`
# itself exits 128 when the pathspec matches nothing, so the glob is checked
# with the shell first — most commits touch no link.md at all, and that must
# stay a no-op rather than a failure.
#
# `<name>` is accepted and currently unused by the chain itself; it is taken
# as a parameter because both call sites already have a change name in hand
# and passing it keeps the call shape stable if a future caller needs it in
# a message or a path.
#
# A PLANNING PATH THAT IS A TRACKED SYMLINK IS NEVER WORKED AROUND. If
# `spectre/changes/` or `docs/superpowers/` is a tracked symlink — or
# `spectre/` is, putting `spectre/changes/` behind one — `git add -A -- .
# ':(exclude)spectre/changes/' ':(exclude)docs/superpowers/'` exits 128 with
# a message naming the path, and stages nothing. `set -euo pipefail` is what
# propagates that exit code and message as-is — there is no trap or `||`
# around that call to catch and reinterpret it. The only way past the stop is
# to fix the repository so the path is a real directory; working around it
# here would let planning content land in the implementation commit, which is
# the one outcome this split exists to prevent.
set -euo pipefail

worktree="$1" name="$2" impl_msg="$3" plan_msg="$4"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/spec-root.sh"
plan_dir="$(spec_root_leaf "$worktree")/changes/"

git -C "$worktree" reset -q -- "$plan_dir" docs/superpowers/
git -C "$worktree" add -A -- . ":(exclude)$plan_dir" ':(exclude)docs/superpowers/'
shopt -s nullglob
link_md_files=("$worktree/$plan_dir"*/link.md)
shopt -u nullglob
if [ "${#link_md_files[@]}" -gt 0 ]; then
  git -C "$worktree" add -A -- "${plan_dir}*/link.md"
fi
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$impl_msg"
git -C "$worktree" add -A
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$plan_msg"
