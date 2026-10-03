#!/usr/bin/env bash
# check-panel-fix-single-dispatch.sh <worktree> <change-name> <session-token>
#
# THE CHANGE NAME IS A SECOND ARGUMENT and the session token a third, for
# the same reason the name is one on check-panel-findings-closed.sh: a
# change's dispatches are rows in the store, keyed by change name, not a
# file at a fixed path, and the token scopes the count to THIS run --
# `flow record dispatches -change <name> -C <worktree>` answers with the
# decoded JSON array of that change's dispatch rows (every role, every
# session token, seq order), and this guard reads only that array.
#
# This is the gate KAN-482 asked for: on the kan-449 run the conductor
# launched four background fix subagents, one per reviewer's findings,
# against the contract review-panel.md already stated -- give the surviving
# findings to ONE fix subagent as the combined list. Prose alone failed
# once, so the obligation is checked here, at the panel's own close,
# immediately before `flow stage end -command '/flow' -stage
# flow.review-panel`, beside check-panel-findings-closed.sh. Dispatch rows
# only exist once dispatches have happened, so a pre-dispatch check has
# nothing to read; the close gate is the last point where the run can still
# be stopped before handoff.
#
# THE COUNT IS SCOPED to the session token named on the command line --
# a change's rows span every run that ever touched it, and only this run's
# dispatches are this round's business. Within that token, every row whose
# role is `panel-fix` must satisfy:
#
#   - its key matches `^panel-fix-[0-9]+(-[0-9]+)?(-retry)?$` -- the canonical
#     shape review-panel.md's fix step declares since kan-499: the bare
#     `panel-fix-<round>` carries the round's first chunk, `panel-fix-<round>-<n>`
#     continues it for each further chunk of at most 10 findings, and `-retry`
#     is the handshake's one allowed second dispatch of that same key; a
#     conductor that invents keys (panel-fix-f1, one per finding) would
#     otherwise read as four clean single-dispatch rounds, which is exactly
#     the KAN-482 shape;
#   - its base (the key with a trailing `-retry` stripped) carries at most
#     one non-retry dispatch, at most two rows total, and two only when
#     exactly one of them is the `-retry` variant -- the handshake's one
#     allowed second dispatch, per implement.md's **The handshake**. The rule
#     holds per chunk key, not merely per round;
#   - a `-retry` row never stands alone: a retry with no original has no
#     round it could be a retry of.
#
# and each ROUND that carries chunked dispatches must also satisfy (kan-499):
#
#   - the round's own bare key `panel-fix-<round>` carries at least one
#     original dispatch -- chunk numbering starts there, at `-2` for the
#     second chunk;
#   - its chunks are contiguous -- `-2`, `-3`, `-4`, ... with no gap: a gap
#     means the "numbering" is decoration, not a chunking of a finding list;
#   - its chunk count is at most `ceil(findings raised in earlier rounds /
#     10)`, counted from `flow record findings` -- a BOUND, not a target. A
#     well-formed per-finding key sequence (`panel-fix-1-2` through
#     `panel-fix-1-25` against 24 findings) must stay caught: the KAN-482
#     abuse in this contract's own clothing. The guard reads only the
#     store's answer, never a rendered document; a findings read that fails
#     is "cannot answer" (exit 2), never a clean verdict, matching the
#     dispatches read below.
#
# Rows of other roles and rows of other session tokens never count.
#
# THE GUARD NEVER READS THE JOURNAL and never consults anything but the
# store's answer, matching check-panel-findings-closed.sh: a read that
# failed is a question this guard cannot answer, not a clean verdict.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-findings-closed.sh's own copy -- the change name arrives from
# a pull-request-editable state file and is passed to `flow record
# dispatches -change`, so `../../../planted` and a glob metacharacter are
# hazards here exactly as they are there. The worktree is canonicalised
# before it is ever passed to `flow` as `-C`, and refused rather than
# proceeded with if it vanished between the `-d` check and here, for the
# same reasons that guard's header names.
#
# Exit codes:
#   0  every panel-fix row of this session token is shape- and count-clean
#   1  at least one violation -- each offending round or key named on stderr
#   2  cannot answer at all -- missing arguments, a non-directory worktree,
#      a change name outside the allowlist, the store unreachable (the flow
#      call exits non-zero), or rows that are not readable JSON
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-panel-fix-single-dispatch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-fix-single-dispatch 2 "check-panel-fix-single-dispatch:" "$@"
