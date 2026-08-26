# scripts/lib/spec-root.sh — spec_root_leaf, defined once.
#
# Sourced by every guard that has to locate a project's change directory:
# check-cleanup-complete.sh, check-task-commit-fields.sh,
# check-unfinished-work.sh, commit-split.sh, gather-dispatch-context.sh,
# gather-self-review-context.sh and plan-dispatch-bundles.sh. Each of those
# used to hardcode the literal `spectre/changes`, which is correct for a
# project whose tree is named `spectre/` and wrong — silently — for one whose
# tree is still named `openspec/`.
#
# THIS IS A LOCATE, NOT A PREFERENCE. The rename from `openspec/` to
# `spectre/` happened in this repository's own skills and contracts, not in
# the projects those skills are installed into. A project initialised before
# the rename still carries `openspec/changes/`, and nothing migrates it: the
# directory name is the project's, and `spectre new` is not going to move a
# tree it did not create. So a guard that hardcodes one name does not fail
# loudly against the other — it reports "no tasks.md found", which reads as a
# plan that was never written rather than as a directory it never looked in.
# That is the failure this file exists to stop, measured on gymie, whose tree
# is `openspec/` and whose every task-1 commit-fields check had to be done by
# hand because check-task-commit-fields.sh could not see the plan.
#
# WHY A LEAF NAME AND NOT A FULL PATH. Three of the seven callers need the
# name as a git pathspec prefix (`:(exclude)<leaf>/changes/`) or inside a
# message, not as an absolute path — so returning `<dir>/<leaf>/changes`
# would have each of them tear the answer back apart. The leaf composes into
# whichever shape a caller needs, and composing is what every caller already
# does with the literal it replaces.
#
# "SAFELY REACH IT" — the criterion scripts/lib/resolve-file.sh's header
# states, and this file follows it too: every one of the seven callers ships
# through the skills/*/scripts/ symlink farm, where a sibling `lib/` travels
# with them, so each can source this rather than carry a copy.

# spec_root_leaf <dir> — print the leaf name of <dir>'s OpenSpec/spectre tree.
#
# Prints `spectre` when <dir>/spectre/changes/ exists, `openspec` when
# <dir>/openspec/changes/ exists, and `spectre` when neither does.
#
# `changes/` rather than the tree root itself is what is probed, deliberately.
# A bare `spectre/` or `openspec/` directory can exist for reasons that have
# nothing to do with a project having been initialised — an empty directory a
# tool created, a stray checkout — whereas `changes/` is what `spectre init`
# actually makes and what every caller here goes on to glob. Probing the
# directory the caller needs means a hit is evidence the caller can use.
#
# WHEN NEITHER EXISTS, `spectre` IS RETURNED RATHER THAN AN ERROR. Every
# caller already handles an absent change directory — an uninitialised
# project, or one with no open change — and each has its own message for it.
# Returning a name keeps that handling exactly where it is, and keeps those
# messages naming the canonical directory a project is expected to gain.
# Failing here instead would move seven scripts' worth of error handling into
# this one, and would turn "you have no open changes" into "the guard is
# broken".
#
# WHEN BOTH EXIST, `spectre` WINS AND THE OTHER IS REPORTED ON STDERR. Both
# present means a half-finished migration, which is a real condition worth
# saying out loud: whichever this function picked, the other tree's changes
# are invisible to the caller. It is a note rather than a failure, because a
# migration in progress is a legitimate state to be in for as long as it takes
# to move the directories, and refusing to run every guard until it finishes
# would make the migration harder rather than safer.
spec_root_leaf() {
  local dir="$1"
  local has_spectre=0 has_openspec=0

  [ -d "$dir/spectre/changes" ] && has_spectre=1
  [ -d "$dir/openspec/changes" ] && has_openspec=1

  if [ "$has_spectre" -eq 1 ] && [ "$has_openspec" -eq 1 ]; then
    echo "spec-root: $dir carries both spectre/changes/ and openspec/changes/ — using spectre/; every change under openspec/changes/ is invisible to this guard until that tree is moved" >&2
    printf 'spectre\n'
    return 0
  fi

  if [ "$has_openspec" -eq 1 ]; then
    printf 'openspec\n'
    return 0
  fi

  printf 'spectre\n'
}
