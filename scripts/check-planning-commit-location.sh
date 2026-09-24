#!/usr/bin/env bash
# check-planning-commit-location.sh — refuse a planning commit (a link commit
# after `spectre link`, any other `chore(spectre): …` commit over the change
# folder, or the two-commit chain's planning half) anywhere but the change's
# own worktree on its own `spectre/<name>` branch.
#
# Usage: check-planning-commit-location.sh <worktree> <name>
#
# `spectre link` writes `link.md` on both sides of a cross-repo change, and
# it is run with the working directory at a repository's PRIMARY checkout so
# its peers file resolves — which is exactly where a link commit made from
# the wrong directory lands. A planning commit in the main checkout, or on
# any branch but `spectre/<name>`, puts change-folder content on the landing
# target (or on another change's branch) where no reshape, review or archive
# step of this change ever sees it. Every planning commit therefore runs
# behind this guard; git-boundaries.md's **Planning commits** is canonical
# for where it is called.
#
# Verdict lines, on stdout:
#
#   PLANNING-COMMIT-MAIN-CHECKOUT: <worktree>          the path is a
#                                                     repository's main
#                                                     checkout, not a
#                                                     linked worktree
#   PLANNING-COMMIT-WRONG-BRANCH: <worktree> on <branch|detached> — expected spectre/<name>
#   PLANNING-COMMIT-LOCATION-OK: <worktree> on spectre/<name>
#
# Both violation lines print when both hold. Exit 0 on the OK verdict; exit 1
# on any violation; exit 2 with NOTHING on stdout when the arguments are
# missing or empty, or <worktree> is not a readable git work tree — an
# inability is never reported as a verdict.
#
# "MAIN CHECKOUT" IS THE WORKTREE WHOSE GIT DIR IS THE COMMON DIR — the same
# test hooks/protect-main-checkout.py makes. A linked worktree's `--git-dir`
# is `<common>/worktrees/<id>`, so it differs from `--git-common-dir`; both
# are asked for with `--path-format=absolute` so a relative `.git` from a
# main checkout never compares unequal to its own absolute form.
set -euo pipefail

SELF="check-planning-commit-location.sh"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -eq 2 ] && [ -n "$1" ] && [ -n "$2" ] \
  || die "usage: check-planning-commit-location.sh <worktree> <name>"
WT="$1"
NAME="$2"

[ -d "$WT" ] || die "not a readable directory: $WT"
[ "$(git -C "$WT" rev-parse --is-inside-work-tree 2>/dev/null)" = "true" ] \
  || die "not a git work tree: $WT"
GIT_DIR_ABS="$(git -C "$WT" rev-parse --path-format=absolute --git-dir 2>/dev/null)" \
  || die "cannot resolve the git dir of: $WT"
COMMON_ABS="$(git -C "$WT" rev-parse --path-format=absolute --git-common-dir 2>/dev/null)" \
  || die "cannot resolve the common git dir of: $WT"

# symbolic-ref exits 1 on a detached HEAD, which is an answer here, not an
# inability.
BRANCH="$(git -C "$WT" symbolic-ref --short -q HEAD)" || BRANCH="detached"

violations=0
if [ "$GIT_DIR_ABS" = "$COMMON_ABS" ]; then
  printf 'PLANNING-COMMIT-MAIN-CHECKOUT: %s\n' "$WT"
  violations=1
fi
if [ "$BRANCH" != "spectre/$NAME" ]; then
  printf 'PLANNING-COMMIT-WRONG-BRANCH: %s on %s — expected spectre/%s\n' "$WT" "$BRANCH" "$NAME"
  violations=1
fi
[ "$violations" -eq 0 ] || exit 1
printf 'PLANNING-COMMIT-LOCATION-OK: %s on spectre/%s\n' "$WT" "$NAME"
exit 0
