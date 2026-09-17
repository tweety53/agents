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
# WHAT "FOREIGN" MEANS. Every pipeline run works in a worktree; nothing in
# `/flow` or `/flow-fast` ever stages into a main checkout. So ANY staged
# entry in a main checkout is foreign to the change in flight — residue of
# some other session — and is listed as such, whatever it contains. The
# operator decides what it is and what becomes of it; this guard classifies
# nothing beyond "staged".
#
# WHY ONLY STAGED ENTRIES. Staged work is what a resumed run is tempted to
# `git reset` or `git stash` when the preflight's main-checkout assertion
# refuses — the high-judgment surgery kan-437's run 2 performed inline
# across three repos. Unstaged worktree modifications also refuse that
# assertion, but they are not the reset/stash target class and naming them
# "foreign staged work" would be false; they stay the preflight's own
# finding to report. Untracked files are hidden from the status read
# entirely, exactly as the preflight's assertion hides them.
#
# An intent-to-add entry (`git add -N`) is index work like any other stage,
# but porcelain prints its index code blank — ` A <path>` — so the filter
# also lists a line whose SECOND column is `A`: that second-column A is the
# only index-work state a blank first column can hide, and the preflight's
# identical status read refuses on it, so hiding it would leave the surface
# silent on a state the very next gate stops for.
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
set -euo pipefail

SELF="check-foreign-staged"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -eq 1 ] || {
  printf 'usage: check-foreign-staged.sh <main-checkout>\n' >&2
  exit 2
}

ROOT="$(cd "$1" 2>/dev/null && pwd -P)" || die "$1 is not a directory"
git -C "$ROOT" rev-parse --git-dir >/dev/null 2>&1 ||
  die "$ROOT is not a git repository"

STATUS_OUT="$(git -C "$ROOT" status --porcelain --untracked-files=no 2>/dev/null)" ||
  die "cannot read the status of $ROOT"

COUNT=0
if [ -n "$STATUS_OUT" ]; then
  while IFS= read -r line; do
    FIRST="${line:0:1}"
    SECOND="${line:1:1}"
    if [ "$FIRST" != " " ] || [ "$SECOND" = "A" ]; then
      printf 'FOREIGN-STAGED: %s\n' "$line"
      COUNT=$((COUNT + 1))
    fi
  done <<EOF
$STATUS_OUT
EOF
fi

if [ "$COUNT" -eq 0 ]; then
  printf 'STAGED-CLEAN: %s\n' "$ROOT"
else
  printf 'STAGED-FOREIGN: %s — %d\n' "$ROOT" "$COUNT"
fi
