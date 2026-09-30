#!/usr/bin/env bash
# remove-change-worktrees.sh — run /flow run 2's Worktree cleanup
# (skills/flow-contracts/finish-contract-run2.md) for one repository: every
# check on every worktree before anything is removed, then the removals, the
# local branch and the remote branch.
#
# Usage: remove-change-worktrees.sh <repo> <name> <merge-base|-> [--proceed]
#
# The worktrees are the ones `git -C <repo> worktree list --porcelain` lists on
# refs/heads/spectre/<name>, plus every detached wave-group copy of one
# (`<worktree>-wave-group-<g>`). A listed worktree whose directory is gone is
# already removed: success, its registration pruned. <merge-base> scopes the
# wave-group copies' disclosed log; `-` discloses their status alone.
#
# In order, never removing anything before every check has passed:
#   checks 1-4 on each apply worktree — no uncommitted tracked change; no
#     untracked, unignored file; no commit that exists only here (the base
#     resolved fresh for THIS worktree, in-process as resolve-base-branch.sh
#     does); and the ignored files `--force` will destroy, split by path into
#     regeneratable (counted) and unclassified (listed).
#   the disclosure stop — without --proceed, a non-empty unclassified bucket
#     or any wave-group copy stops here, before check 5 runs anything.
#   checks 5-6 on each copy, then each apply worktree — the fenced command
#     lines of <repo>/.flow/project.md's `## stop` body (the key's shape is
#     project-configuration.md's), run as one script with the
#     worktree as its directory under `bash -o pipefail -c`, bounded at 60
#     seconds (own process group, SIGTERM, a 2-second grace, SIGKILL); a file,
#     key or fence that is absent means no command is declared, the check is
#     skipped, and nothing outside the fence is ever run. Then the sibling
#     check-worktree-processes.sh.
#   `worktree remove --force` for each copy, then each apply worktree;
#   `worktree prune`; `branch -d spectre/<name>` (never -D) when it exists;
#   and `push origin --delete spectre/<name>`, not gated on the local steps —
#   a "remote ref does not exist" refusal is already gone, and prunes the
#   stale tracking ref with `fetch --prune`.
#
# Prints to stdout, one line each:
#   REFUSED: <worktree> — check <n>: <reason>   a check failed (gate)
#   HELD: <worktree> — <pid> <cwd>[; ...]        check 6 found a live process
#   UNCLASSIFIED: <worktree> — <path>            check 4, one per entry
#   REGENERATABLE: <worktree> — <count>          check 4
#   SKIPPED: check 5 — ## stop declares no fenced command   a declared `## stop`
#     whose body carries no fence; check 5 is skipped, never silently
#   DISCLOSE: <copy> — status: <line> / commit: <line>, and a final
#   DISCLOSE: stopped before removal — …          the disclosure stop
#   REMOVED: <worktree> | <worktree> — already gone | spectre/<name>
#   REFUSED: <worktree> — git worktree remove: <git's message>
#   REFUSED: spectre/<name> — git branch -d: <git's message>
#   REMOTE-DELETED: origin/spectre/<name>
#   REMOTE-GONE: origin/spectre/<name>
#   REMOTE-REFUSED: origin/spectre/<name> — <git's message>
#
# Exit 0 every check passed and every removal succeeded (whatever the REMOTE-*
# line says — it is reported, and check-cleanup-complete.sh verifies it);
# 1 a check failed — nothing removed — or a removal failed — its entry stays
# in the state file's `worktrees`; 2 cannot answer, with NOTHING on stdout: a
# usage error, a name that is not a plain change name, a <repo> that is not a
# git repository, or its worktree list unreadable; 3 the disclosure stop —
# nothing removed, nothing run: relay, judge each unclassified entry, ask the
# one ask, then call again with --proceed.
#
# Run it from outside every worktree it removes: check 6 counts the caller's
# own shell.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
# FLOW_GUARD_SELF (the path this script was invoked by) is exported so the Go
# guard execs $SCRIPT_DIR/check-worktree-processes.sh from beside this script.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "remove-change-worktrees: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
[ -x "$SCRIPT_DIR/check-worktree-processes.sh" ] || {
  echo "remove-change-worktrees: no executable check-worktree-processes.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec remove-change-worktrees 2 "remove-change-worktrees:" "$@"
