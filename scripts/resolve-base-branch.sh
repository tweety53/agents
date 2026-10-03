#!/usr/bin/env bash
# resolve-base-branch.sh — the single owner of base-branch resolution.
#
# Usage: resolve-base-branch.sh <dir>
#
# stdout carries the bare branch name and NOTHING ELSE on success, so a
# caller composes it directly: BASE="$(resolve-base-branch.sh "$WT")". Every
# refusal writes its reason to stderr and leaves stdout empty — an empty
# stdout must never be readable as a resolved base.
#
#   Exit 0   Resolved; the name is on stdout.
#   Exit 1   A named refusal — detached HEAD, no base resolved, base equal
#            to the current branch, or a base name that fails validation.
#            Validation requires the first character to be one of
#            [A-Za-z0-9._] (so a leading '-' or a leading '/' is refused)
#            and every character in the name to be one of [A-Za-z0-9._/-].
#   Exit 2   Cannot answer — the argument is missing, <dir> is absent,
#            unreadable, or not a git worktree, or HEAD's own ref cannot be
#            read (corrupt or permission-denied).
#   Exit 3   The repository has no 'origin' remote at all.
#
# A RECORDED BASE WINS. When `git config branch.<current>.flowBase` is set —
# kickoff-worktree.sh writes it when a change is created with `--base
# <branch>` — that name is the base, ahead of origin/HEAD, and faces the same
# equal-to-current and name-validation refusals.
#
# WHY THE FETCH IS WRAPPED. `git remote show origin` against an unreachable
# host blocks for roughly 75s on the default TCP timeout, which would turn a
# correct refusal into a two-minute hang. `-c core.askpass=true` stops it
# prompting for credentials, `2>/dev/null` swallows the chatter, and `|| true`
# means a failed fetch is not this script's failure — a stale origin/HEAD is
# still resolved, just possibly out of date, which the fetch on the next
# invocation corrects.
#
# WHY HEAD@{upstream} IS NEVER CONSULTED. Inside an apply worktree, HEAD *is*
# the change's own branch, spectre/<name>, by construction. Its upstream is
# origin/spectre/<name>, so HEAD@{upstream} would compare the branch with
# itself the moment it is pushed — always "true", never useful. The base
# branch is resolved from origin/HEAD (or the `git remote show origin`
# fallback), never from HEAD's own upstream.
#
# WHY EXIT 3 IS SEPARATE FROM EXIT 2. The finish contract requires a distinct
# no-remote message. A caller that could only tell the two apart by matching
# stderr text would stop being able to the first time the wording changes.
#
# The unreachable-remote hang is deliberately NOT covered by this script's
# tests (TestResolveBaseBranch): a wall-clock assertion on a TCP timeout is flaky by
# construction. The wrapping above is carried forward verbatim from the
# fenced block this script replaces, and this comment is where its reasoning
# now lives.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "resolve-base-branch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec resolve-base-branch 2 "resolve-base-branch:" "$@"
