#!/usr/bin/env bash
# check-worktree-location.sh — refuse any registered worktree that lives
# outside <project>/.worktrees/, the strict layout every /flow worktree is
# created under (design.md's strict-worktrees-guard decision, KAN-462 §11).
#
# Usage: check-worktree-location.sh <project>
#
# Prints one STRAY line per offender, then ONE verdict line:
#   STRAY: <path> (<branch>|detached)      per worktree outside .worktrees/
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
# `_landing-<name>` NEEDS NO RULE OF ITS OWN. It is created at
# <project>/.worktrees/_landing-<name>, already at or under the root this
# guard checks, so it is reported like any other in-tree worktree: not at all.
#
# PATHS ARE COMPARED IN PHYSICAL FORM. The project root is resolved with
# `cd … && pwd -P` before the comparison, matching what `git worktree list`
# itself already reports — git resolves a worktree's path (through /tmp's
# macOS symlink to /private/tmp, for instance) at `add` time, so the entries
# read from porcelain need no separate resolution of their own.
#
# "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals
# <project>/.worktrees or begins with it plus a slash — a bare string-prefix
# test would report <project>-worktrees/x as in-tree, the retired sibling
# layout this guard exists to catch (KAN-462's proposal.md).
set -euo pipefail

SELF="check-worktree-location"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -eq 1 ] || die "usage: check-worktree-location.sh <project>"

ROOT="$(cd "$1" 2>/dev/null && pwd -P)" || die "$1 is not a directory"
PORCELAIN="$(git -C "$ROOT" worktree list --porcelain 2>/dev/null)" \
  || die "cannot list the worktrees of $ROOT"

# One pass over the porcelain records. n counts every `worktree` line seen so
# far; the main checkout is always first (n == 1 at its record) and is never
# flagged. count accumulates the number of STRAY lines printed, carried out
# via the trailing COUNT: line rather than a second command substitution that
# `set -e` could trip over on a zero result.
#
# `printf '%s\n\n'`, not `'%s\n'`: `$(...)` strips every trailing newline from
# `$PORCELAIN`, including the blank line that terminates the LAST record, so a
# single appended newline reconstructs a porcelain stream with no closing
# blank line and the final worktree's record never reaches the `/^$/` block
# below. The second newline restores that terminator regardless of how many
# `git worktree list --porcelain` actually emitted.
RESULT="$(printf '%s\n\n' "$PORCELAIN" | awk -v root="$ROOT/.worktrees" '
  /^worktree / { n++; w = substr($0, 10); b = "detached" }
  /^branch /   { b = $2 }
  /^$/         {
    if (n > 1 && w != root && index(w, root "/") != 1) {
      print "STRAY: " w " (" b ")"
      count++
    }
    w = ""
  }
  END { print "COUNT:" count + 0 }
')"

STRAY_LINES="$(printf '%s\n' "$RESULT" | sed '$d')"
COUNT="$(printf '%s\n' "$RESULT" | tail -n 1)"
COUNT="${COUNT#COUNT:}"

[ -n "$STRAY_LINES" ] && printf '%s\n' "$STRAY_LINES"

if [ "$COUNT" -gt 0 ]; then
  printf 'LOCATION-STRAY: %s — %s\n' "$ROOT" "$COUNT"
  exit 1
fi
printf 'LOCATION-OK: %s\n' "$ROOT"
exit 0
