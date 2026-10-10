#!/usr/bin/env bash
# check-task-reviewer-single-dispatch.sh <worktree> <change-name> <session-token>
#
# THE CHANGE NAME IS A SECOND ARGUMENT and the session token a third, for
# the same reason the name is one on check-panel-fix-single-dispatch.sh: a
# change's dispatches are rows in the store, keyed by change name, not a
# file at a fixed path, and the token scopes the count to THIS run --
# `flow record dispatches -change <name> -C <worktree>` answers with the
# decoded JSON array of that change's dispatch rows, and this guard reads
# only that array plus <worktree>/spectre/changes/<name>/tasks.md (the
# canonical worktree's copy -- the only one that exists on a cross-repo
# change, per skills/flow-contracts/worktree-resolution.md; its
# changes/archive/<name>/ copy once integrate's run 1 archived the change).
#
# This is the gate KAN-527 asked for, the gated-per-task-reviewer sibling
# of check-panel-fix-single-dispatch.sh: on that run the conductor launched
# nine one-task reviewer dispatches against nine gate-fired tasks that
# belonged to three implementer groups, before the operator stopped the
# run and implement.md's "The gated per-task reviewer" paragraph was
# rewritten to require ONE bundled dispatch per implementer group whose
# gate fired (on `big`; KAN-934 later dropped the gated reviewer on
# `micro`/`small`/`regular`). Prose
# alone already failed once for the panel's own fix step (KAN-482); this
# guard exists so the same failure mode is caught here too, at the stage's
# own close, immediately before `flow stage end -command '/flow' -stage
# flow.sdd-tdd`, the same position check-panel-fix-single-dispatch.sh holds
# at `flow.review-panel`'s close.
#
# THE COUNT IS SCOPED to the session token named on the command line --
# a change's rows span every run that ever touched it, and only this run's
# dispatches are this run's business. Within that token, every row whose
# role is `reviewer` and whose key matches the gated-per-task-reviewer
# family (`^task-[0-9]+(\+[0-9]+)*-reviewer(-fix-[0-9]+)?(-retry)?$`) must
# satisfy:
#
#   - its key is in that canonical shape -- a key naming one task
#     (`task-5-reviewer`) or several `+`-joined in plan order
#     (`task-13+14+15-reviewer`), optionally a fix round
#     (`task-13+14-reviewer-fix-1`) and optionally the handshake's one
#     retry (`task-13+14-reviewer-fix-1-retry`); a conductor that invents
#     one key per task (`task-5-reviewer`, `task-6-reviewer`, ... for tasks
#     that share an implementer group) is exactly the KAN-527 shape;
#   - its base (the key with a trailing `-retry` stripped) carries at most
#     one non-retry dispatch, at most two rows total, and two only when
#     exactly one of them is the `-retry` variant -- the handshake's one
#     allowed second dispatch, per implement.md's **The handshake**;
#   - a `-retry` row never stands alone: a retry with no original has no
#     bundle it could be a retry of.
#
# BEYOND the per-key shape (which is everything check-panel-fix-single-
# dispatch.sh's own key family checks), this guard also checks the
# BUNDLING ITSELF, since a KAN-527-shaped run can produce well-formed keys
# that are still one-per-task: no two ORIGINAL (non-fix, non-retry) keys
# may name tasks belonging to the same implementer group. Group membership
# is read from `<worktree>/spectre/changes/<name>/tasks.md` via
# `plan-dispatch-bundles.sh` (task -> bundle) composed with
# `plan-dispatch-groups.sh` (bundle -> group) -- the same source of truth
# `skills/flow/brainstorm-planner.md`'s Decide step and implement.md's own
# wave-grouping already use, never re-derived by this guard from the raw
# markdown. On the decision's `class` `micro`, `small` or `regular` (read via `flow
# record decisions -change <name> -C <worktree>`, the newest entry's -- the verb lists newest first --
# `.decision.class`; a failed read, output that is not JSON, or no class
# is treated as `big`, and JSON that is not an array of decision rows is a
# cannot-answer, exit 2 -- `big` tolerates gated reviewer rows, so reading
# it as `big` would loosen the check),
# implement.md dispatches no gated reviewer at all: every gate-fired task is
# the whole-branch panel's review focus instead (KAN-934) -- so on those
# classes, any gated-reviewer row of this token is itself a violation.
#
# Rows of other roles and rows of other session tokens never count. A task
# id that never appears in any `reviewer`-role key of this token was never
# gated, or its gate hasn't fired yet, or its bundle is still in flight --
# this guard says nothing about coverage, only about the shape of what was
# already dispatched.
#
# THE GUARD NEVER READS THE JOURNAL and never consults anything but the
# store's answer for dispatch rows, matching check-panel-fix-single-
# dispatch.sh: a read that failed is a question this guard cannot answer,
# not a clean verdict. The tasks.md read is a plain file read, not a store
# read; a missing or unparsable tasks.md is likewise "cannot answer" (exit
# 2), never a clean verdict, since without it group membership cannot be
# established at all.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-fix-single-dispatch.sh's own copy at d71a2327 (now the
# CONTAINMENT check in stats/internal/guard/panelfixsingledispatch.go; itself duplicated from
# check-panel-findings-closed.sh) -- the change name arrives from a
# pull-request-editable state file and is passed to `flow record
# dispatches -change`, so `../../../planted` and a glob metacharacter are
# hazards here exactly as they are there. The worktree is canonicalised
# before it is ever passed to `flow` as `-C`, and refused rather than
# proceeded with if it vanished between the `-d` check and here, for the
# same reasons that guard's header names.
#
# Exit codes:
#   0  every gated-per-task reviewer row of this session token is
#      shape-clean, retry-clean, and bundled per implementer group (and,
#      on `micro`/`small`/`regular`, there is no such row)
#   1  at least one violation -- each offending key or group named on
#      stderr
#   2  cannot answer at all -- missing arguments, a non-directory
#      worktree, a change name outside the allowlist, the store
#      unreachable, tasks.md missing or unreadable, plan-dispatch-bundles.sh
#      or plan-dispatch-groups.sh failing, dispatch rows that are not
#      readable JSON, or decision output that is JSON but not an array of
#      decision rows
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
# FLOW_GUARD_SELF (the path this script was invoked by) is exported so the Go
# guard execs $SCRIPT_DIR/plan-dispatch-bundles.sh and
# $SCRIPT_DIR/plan-dispatch-groups.sh from beside this script, as the bash did.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-task-reviewer-single-dispatch: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec check-task-reviewer-single-dispatch 2 "check-task-reviewer-single-dispatch:" "$@"
