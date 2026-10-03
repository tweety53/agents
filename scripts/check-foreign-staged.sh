#!/usr/bin/env bash
# check-foreign-staged.sh — list the foreign staged work a main checkout
# carries, before a resumed /flow run's preflight reaches it (KAN-546,
# observed in kan-437's run 2).
#
# Usage: check-foreign-staged.sh <main-checkout>
#
# Prints one FOREIGN-STAGED line per staged or unmerged entry, then ONE
# verdict line:
#   FOREIGN-STAGED: <porcelain line>       per staged or unmerged entry
#   STAGED-CLEAN: <main-checkout>          nothing is staged
#   STAGED-FOREIGN: <main-checkout> — <n>  <n> staged/unmerged entries
#
# Exit 0 on either verdict; exit 2 with NOTHING on stdout when the argument
# is missing, is not a readable directory, is not a git repository, or
# `git status` fails there — an inability is never reported as a verdict.
#
# THE VERDICT NAMES THE PHYSICAL PATH, resolved with `cd … && pwd -P`, so a
# caller that passed a symlinked path still sees the real checkout named —
# the same resolution `git worktree list` applies when it prints paths.
#
# HOW TO HAND-VERIFY A STAGED-FOREIGN VERDICT. Run the same read by hand:
# `git -C <main-checkout> status --porcelain --untracked-files=no` and keep
# the lines whose first column is not a space, together with any
# intent-to-add entry's ` A <path>` line — those, and only those, are
# the entries the FOREIGN-STAGED lines name.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-foreign-staged: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-foreign-staged 2 "check-foreign-staged:" "$@"
