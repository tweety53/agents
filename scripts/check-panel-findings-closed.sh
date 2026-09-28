#!/usr/bin/env bash
# check-panel-findings-closed.sh <worktree> <change-name>
#
# THE CHANGE NAME IS A SECOND ARGUMENT for the same reason it is one on
# check-panel-reproducers.sh: a change's findings are rows in the store,
# keyed by change name, not a file at a fixed path. `flow record findings
# -change <name> -C <worktree>` answers with the decoded JSON array of that
# change's findings, and the guard reads that array plus the change's
# dispatch rows the same way (`flow record dispatches -change <name> -C
# <worktree>`), the second read answering only the fixed-without-clean-rerun
# class below.
#
# This is the gate design.md's `gate-is-a-guard` decision chose: nothing
# checked, before this guard existed, that a review panel actually closed
# every finding it verified — `check-unfinished-work.sh` only ever caught
# the omission at /flow's integrate gate, after the work was done. This
# guard runs at the panel's own close, immediately before
# `flow stage end -command '/flow' -stage flow.review-panel`.
#
# THE OPEN-FINDING PREDICATE IS DUPLICATED, on purpose, from
# check-unfinished-work.sh's own copy — design.md's `duplicate-the-predicate`
# decision, following that guard's own precedent for why a one-line
# predicate gains nothing from being centralized and loses the property that
# both harnesses assert the same shape.
#
# THE GUARD NEVER CONSULTS THE JOURNAL — design.md's `no-journal-excuse`
# decision. `flow record status` never blocks, so a store outage journals a
# close instead of landing it, and a finding whose close only reached the
# journal still reads `open` here and reports exit 1. That is the honest
# verdict; the existing handback is where the operator resolves it.
#
# Exit codes:
#   0  no finding's status is open, no Minor is wrongly deferred, and every
#      finding recorded `fixed` has a clean re-run dispatch of its slot in
#      a later round
#   1  one or more findings are open, or a Minor is recorded `deferred`
#      in a round that raised a Critical or Important not recorded
#      `withdrawn` (review-panel.md's **Panel re-runs** sends such a
#      Minor to that round's fix), or a finding is recorded `fixed`
#      with no clean re-run dispatch of its slot in any later round —
#      the ordering review-panel.md's **Recording findings** requires,
#      violated, the store's only witness of that re-run being the
#      slot's own dispatch row (KAN-770); each still-open ref, each
#      wrongly deferred ref with its round, and each unverified ref
#      with its slot is named on stderr
#   2  cannot answer at all — no worktree, no change name, a change name
#      outside the allowlist, a worktree that is not a directory, the store
#      unreachable, or its answer unreadable (empty output included)
# Ported to Go: the body's reasoning for every branch is in
# stats/internal/guard/panelfindingsclosed.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-panel-findings-closed: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-panel-findings-closed 2 "check-panel-findings-closed:" "$@"
