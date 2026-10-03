#!/usr/bin/env bash
# check-done-when-paths.sh — refuse a tree whose `## Done when` sections name
# paths the index does not track (KAN-817, kan-743's self-review): a done
# criterion is only as good as the committed state it names, and one
# satisfied by uncommitted, gitignored files is silently erased by cleanup.
# commit-archive runs this check before it stages anything; this shim is its
# standalone form.
#
# Usage: check-done-when-paths.sh <worktree>
#
# Scans every tracked `.md` file's `## Done when` sections (heading variants
# `Done when`, `Done-When`, any level, optional trailing colon; fenced code
# blocks are invisible to the scanner — a fence line neither opens nor closes
# a section nor is judged) and extracts path-like tokens — backticked, bare,
# markdown-link targets and autolinks, each carrying a `/` whose last segment
# names a file: not empty and not purely numeric, so a ticket's
# `Snapshots 27/28/29` never reads as a path while a dotless `src/Makefile`
# is judged like any other name. A URL, a glob (`*?[`), a directory (no final
# segment) and a
# peer-qualified name (`peer:shots/27.png`) are judged as named — the last
# fails unless that exact path is tracked in THIS index.
#
# Prints one DONE-WHEN-PATH line per missing path with the file that named
# it, then ONE verdict line:
#   DONE-WHEN-OK: <worktree>                 every named path is tracked
#   DONE-WHEN-VIOLATION: <worktree> — <n>    <n> named path(s) missing
#
# Exit 0 on DONE-WHEN-OK (a tree with no `## Done when` section included), 1
# on DONE-WHEN-VIOLATION, 2 with NOTHING on stdout when <worktree> is not a
# readable directory, is not a git worktree, the index cannot be read, or a
# tracked `.md` file is missing from the worktree — an inability to answer is
# never reported as a verdict.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-done-when-paths: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-done-when-paths 2 "check-done-when-paths:" "$@"
