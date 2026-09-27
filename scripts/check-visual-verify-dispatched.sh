#!/usr/bin/env bash
# check-visual-verify-dispatched.sh — report whether a change whose diff
# touched a declared `## visual verification` UI path actually dispatched
# `flow.visual-verify`'s own verifier, for /flow's run-1 finish gate.
#
# Usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>
#
# WHY THIS GUARD EXISTS. `flow.visual-verify` (skills/flow/visual-verify.md)
# is the one stage built to catch a defect that is obvious the moment a page
# is opened but invisible to a diff, a review panel and both test suites —
# see that file's own **Blocking** paragraph. Nothing stopped an agent from
# writing `flow stage begin`/`flow stage end -outcome completed` around
# `flow.visual-verify` with NOTHING inside — no verifier dispatched, no screenshot
# read, no mockup composed — and reaching this handoff anyway. kan-30's own
# fix round 4 did exactly that: the stage's marks were written, the operator
# was told "no open finding", and a five-minute manual pass immediately
# after found the shipped screens badly wrong. A stage mark is not evidence
# the stage ran; only a recorded dispatch is. This guard makes the gap loud
# rather than silent, the same way check-panel-fix-single-dispatch.sh did
# for a different flavour of "the mark says one thing, the dispatches say
# another."
#
# Prints ONE verdict line to stdout:
#   VISUAL-VERIFY-OK: <reason>          nothing outstanding
#   VISUAL-VERIFY-MISSING: <reason>     UI paths touched, no verifier dispatch found
#
# Exit 0 on a VISUAL-VERIFY-OK verdict, exit 1 on VISUAL-VERIFY-MISSING —
# UI paths were touched and no qualifying verifier dispatch is recorded;
# the verdict line carries the answer — and exit 2 when it cannot answer at
# all — a non-directory worktree, a change name outside the allowlist, an
# empty merge-base, dispatch rows that are not one readable JSON array, or a
# store call that failed outright (never
# read as "no dispatches", exactly as check-panel-fix-single-dispatch.sh's
# own header explains for the identical hazard: an outage that read as zero
# rows would pass every round it was blind to).
#
# THE TRIGGER QUESTION IS DELEGATED, NOT RE-ASKED. Whether this change's
# diff touched a declared `ui paths` glob is exactly
# check-visual-trigger.sh's own job (flow.visual-verify step 2's own guard);
# this script pipes the same `git diff --name-only <merge-base>..HEAD`
# through it and reads its three exit codes as-is:
#   2  the project declares no `## visual verification` section at all —
#      VISUAL-VERIFY-OK: not configured. Nothing here to require.
#   1  the section is declared but this diff touched none of its `ui
#      paths` — VISUAL-VERIFY-OK: no UI paths touched.
#   0  at least one touched path matched — continue to the dispatch check
#      below. check-visual-trigger.sh reads `.flow/project.md` from the
#      worktree itself, matching flow.visual-verify's own per-worktree
#      resolution; this guard is called once per worktree in the run's
#      resolved set, exactly as that stage is.
#
# THE DISPATCH CHECK READS THE STORE, NOT A RENDERED FILE — the same
# posture check-panel-fix-single-dispatch.sh and check-panel-reproducers.sh
# already take for the identical reason: `flow record dispatches -change
# <name> -C <worktree>` answers with the decoded JSON array of every
# dispatch row this change has ever recorded, across every run. A row
# counts when its `role` is `verifier`, its `key` starts with
# `visual-verify` (`visual-verify`, `visual-verify-<worktree-basename>`,
# `visual-verify-2`, `visual-verify-retry`, and combinations — this guard
# does not validate the key's exact shape, only that the stage's own
# canonical prefix was used; a malformed key is a different guard's job,
# the same division check-panel-fix-single-dispatch.sh draws between "did
# it happen" here and "did it happen the right number of times" there),
# and it carries ONE OF TWO EVIDENCE SHAPES. The first is the ordinary
# close: its `outcome` is exactly `completed`. The second is the
# session-continuation signature (KAN-809): the row carries `endedAt` and
# no `outcome` at all — the verifier demonstrably ran to an end, but the
# closing `flow record dispatch end -outcome completed` call never landed
# because the session exhausted its context before it could, so the
# stage's completion is store-recoverable from the row's own end instant.
# A row explicitly closed `aborted`, `fallback` or `blocked` is not
# evidence in either shape — the end instant makes it no more evidence
# than case 4's aborted row, because those outcomes say the verifier
# never delivered a report that was read — and a still-open row (no
# `endedAt`) is not evidence either: the guard cannot tell a running
# verifier from a dead one. See **The return** (skills/flow/implement.md)
# for why only `completed` closes a dispatch's own story.
#
# THE VERDICT IS RECORDED, AND PRIOR FALSE POSITIVES ARE ADVISORY — the
# habit stats/internal/guard/unfinishedwork.go already carries, mirrored
# in stats/internal/guard/visualverifydispatched.go. Once the dispatch
# check has reached one of its two
# store-evidence verdicts (the OK above, or the MISSING below), the guard
# records the line verbatim — `flow record verdict -change <name> -guard
# check-visual-verify-dispatched -worktree <worktree> -verdict "<line>"
# -C <worktree>` — with the write's stdout, stderr and exit code all
# discarded: a store outage can neither move the verdict nor change the
# exit code, and a store that refuses the row leaves the gate exactly as
# it printed it. The recorded rows are what make an operator's
# hand-verified override land at all: `flow record verdict false-positive
# -guard check-visual-verify-dispatched` refuses a (change, guard) pair
# the store holds no verdict row for (KAN-809's second break — the
# override path the gate's refusal names refused the very verdict the
# operator had just hand-verified), and on MISSING the guard reads those
# prior false positives back (`flow record verdicts -guard ...
# -false-positive`) and prints the same advisory stderr hint
# unfinished-work prints — skipped on any failure or an empty array,
# never moving the verdict — so the next run reads the override where it
# was written. The two trigger-side verdicts (`not configured`, `no UI
# paths touched`) fire before the store read in the guard's own code
# order, record nothing, and stay store-free.
#
# NO PROJECT NEEDS A ROW PER WORKTREE. A cross-repo change resolves more
# than one worktree, and flow.visual-verify dispatches once per worktree in
# its own resolved set (**Visual verification**, verify-and-handoff.md, steps 1–2) —
# so this guard is called once per worktree too, by its own caller, exactly
# as check-base-moved.sh and check-unfinished-work.sh already are; it never
# tries to resolve the whole set itself.
#
# THE CHANGE-NAME CONTAINMENT CASE IS DUPLICATED, on purpose, from
# check-panel-fix-single-dispatch.sh's own copy — see that guard's header
# for why: the name arrives from a pull-request-editable state file and is
# passed to `flow record dispatches -change`, so `../../../planted` and a
# glob metacharacter are hazards here exactly as they are there.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
# FLOW_GUARD_SELF (the path this script was invoked by) is exported so the Go
# guard execs $SCRIPT_DIR/check-visual-trigger.sh from beside this script, as
# the bash did.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-visual-verify-dispatched: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_SELF="${BASH_SOURCE[0]}"
export FLOW_GUARD_SELF
flow_guard_exec check-visual-verify-dispatched 2 "check-visual-verify-dispatched:" "$@"
