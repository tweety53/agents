#!/usr/bin/env bash
# commit-split.sh — the guarded two-commit chain, in one place.
#
# Usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>
#
# Wraps the exact guarded chain specified in skills/myflow-contracts/pipeline.md's
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
# `<name>` is accepted and currently unused by the chain itself; it is taken
# as a parameter because both call sites already have a change name in hand
# and passing it keeps the call shape stable if a future caller needs it in
# a message or a path.
#
# A PLANNING PATH THAT IS A TRACKED SYMLINK IS NEVER WORKED AROUND. If
# `spectre/` or `docs/superpowers/` is a tracked symlink, `git add -A -- .
# ':(exclude)spectre/' ':(exclude)docs/superpowers/'` exits 128 with a
# message naming the path, and stages nothing. `set -euo pipefail` is what
# propagates that exit code and message as-is — there is no trap or `||`
# around that call to catch and reinterpret it. The only way past the stop is
# to fix the repository so the path is a real directory; working around it
# here would let planning content land in the implementation commit, which is
# the one outcome this split exists to prevent.
set -euo pipefail

worktree="$1" name="$2" impl_msg="$3" plan_msg="$4"

git -C "$worktree" reset -q -- spectre/ docs/superpowers/
git -C "$worktree" add -A -- . ':(exclude)spectre/' ':(exclude)docs/superpowers/'
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$impl_msg"
git -C "$worktree" add -A
git -C "$worktree" diff --cached --quiet \
  || git -C "$worktree" commit -m "$plan_msg"
