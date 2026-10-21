#!/usr/bin/env bash
# guard-autosquash.sh — mechanical ancestry guard for the fixup-to-autosquash
# sequence. A fixup against a sha that resolves but is no ancestor of the
# branch — a stale duplicate object, for instance — hands
# `git rebase --autosquash` a spurious merge base and replays unrelated
# history into the branch; a rebase can likewise strand shas a plan file
# still records. This guard asserts the ancestry the sequence silently
# assumes, before and after.
#
# Subcommands:
#   targets <worktree> <sha> [<sha>...]
#       Run before any autosquash: every fixup target must resolve as a
#       commit object and be an ancestor of the worktree's HEAD.
#   after <worktree> <base-ref> <tasks-md>
#       Run after an autosquash or any rebase: <base-ref> must still be an
#       ancestor of HEAD, and every run of 7 or more lowercase hex characters
#       in <tasks-md> must resolve as a commit object AND be reachable from
#       HEAD. Existence alone passes a dangling sha — the stale
#       recorded-sha shape — so the two are separate checks. There is no
#       length ceiling: a sha glued into a longer hex run fails resolution
#       and is reported, never silently skipped. Lowercase only: git object
#       names are lowercase, so uppercase hex words in prose (DEADBEEF and
#       its like) never trip the sweep.
#
# Exit codes:
#   0  every check passed
#   1  one or more violations, one line each on stderr, each naming the sha
#   2  cannot answer — bad arguments, <worktree> not a git repository,
#      HEAD unresolvable, or <tasks-md> missing, not a regular file, or
#      unreadable

# The guard itself is Go: stats/internal/guard/guardautosquash.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "guard-autosquash: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec guard-autosquash 2 "guard-autosquash:" "$@"
