#!/usr/bin/env bash
# migrate-worktrees.sh — move every registered worktree under
# <main-checkout>/.worktrees/, the retired in-repo layout, of every
# repository named, to the sibling layout
# <dirname main>/<basename main>-worktrees/<name> (kan-916 design.md's
# migrate-all-now and sibling-worktrees-layout decisions), and rewrite every
# flow state record that named a moved path.
#
# Usage: migrate-worktrees.sh <main-checkout> [<main-checkout>...]
#
# Name every repository at once: a cross-repo change's record lives in one
# project and names another repository's worktree, which only a run that
# moved both and reads every project's records can rewrite. A repository
# named twice is migrated once.
#
# Every state record of every repository named is read first —
# `flow state list -C <main>` per repository, then `flow state get` per
# change; a synthetic record (a stage mark's bootstrap)
# is dropped, never written back; a get answered from the local fallback
# ("store unreachable") is refused. Then, per worktree under .worktrees/:
#   one holding another registered worktree is refused (FAILED), since the
#   move would carry the inner one away from its registration;
#   check-worktree-processes.sh says HELD → its HELD line is printed and the
#   worktree is left in place; otherwise `git worktree move <old> <new>`.
# The renames are then read from every repository's git — every worktree
# registered under a sibling root answers for its old .worktrees/ path,
# whichever run moved it —
# and each record whose `worktrees` map names a renamed path (keys compared
# physically, so a key through a symlinked parent matches) is re-read and
# written back with `flow state set -C <main> <name>`: every other field and
# every other key unchanged, the moved key renamed with its merge-base value
# kept, <main> being the repository whose project holds the record. Finally
# each <main>/.worktrees is removed when it is empty. Re-running
# after a HELD worktree is released moves it; re-running after a failed
# record rewrite repairs the record.
#
# Prints, per worktree or record:
#   MIGRATED: <old> -> <new>
#   HELD: <old> — <pid> <cwd>[; ...]        check-worktree-processes.sh's line
#   FAILED: <old> — <reason>                the worktree holds another one, the
#                                           process check failed, the destination
#                                           already exists, or the move failed
#   FAILED: <change> — flow state set: <output>
#   FAILED: <change> — <reason>             the re-read before the write failed
#
# Exit 0 when every worktree was moved or held (nothing to move included),
# 1 when a move or a record rewrite failed, 2 with NOTHING on stdout when it
# cannot answer — no argument, a <main-checkout> that is not a git
# repository, no `flow` on PATH, a `flow state list` that fails or is partial
# (source fallback), or a `flow state get` that fails or reads the local
# fallback. Nothing is moved on
# exit 2.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "migrate-worktrees: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
[ -x "$SCRIPT_DIR/check-worktree-processes.sh" ] || {
  echo "migrate-worktrees: no executable check-worktree-processes.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec migrate-worktrees 2 "migrate-worktrees:" "$@"
