#!/usr/bin/env bash
# check-worktree-location.sh — refuse any registered worktree that lives
# outside <parent>/<project>-worktrees/, the sibling of the main checkout every
# /flow worktree is created under (kan-916 design.md's sibling-worktrees-layout
# decision; the strict guard itself is KAN-462 §11). Outside the repository
# `git check-ignore` exits 128, so Claude Code's LSP filter drops nothing.
#
# Usage: check-worktree-location.sh <project>
#
# Prints one STRAY line per offender, then ONE verdict line:
#   STRAY: <path> (<branch>|detached)      per worktree outside <project>-worktrees/
#   LOCATION-OK: <project>                 every worktree is at or under it
#   LOCATION-STRAY: <project> — <n>        <n> worktree(s) are not
#
# Exit 0 on LOCATION-OK, 1 on LOCATION-STRAY, 2 with NOTHING on stdout when
# <project> is not a readable directory or `git worktree list` fails — an
# inability is never reported as a verdict.
#
# THE FIRST `worktree` ENTRY IS ALWAYS THE MAIN CHECKOUT. `git worktree list
# --porcelain` prints it first, unconditionally, so this guard skips it by
# position rather than by comparing it against <project> — comparing paths
# would need the same physical-form resolution the strays already require,
# for a fact the porcelain format already guarantees.
#
# PATHS ARE COMPARED IN PHYSICAL FORM. The project root is resolved with
# `cd … && pwd -P` before the comparison, matching what `git worktree list`
# itself already reports — git resolves a worktree's path (through /tmp's
# macOS symlink to /private/tmp, for instance) at `add` time, so the entries
# read from porcelain need no separate resolution of their own.
#
# "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals
# <project>-worktrees or begins with it plus a slash — a bare string-prefix
# test would report <project>-worktrees-old/x as in-tree. <project>/.worktrees/,
# the retired in-repo layout, is STRAY like any other path.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-worktree-location: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-worktree-location 2 "check-worktree-location:" "$@"
