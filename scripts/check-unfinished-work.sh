#!/usr/bin/env bash
# check-unfinished-work.sh — report whether a change carries unfinished work,
# for /flow's integrate-run gate.
#
# Usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]
#
# Prints ONE verdict line to stdout:
#   CLEAR: <reason>          nothing outstanding
#   OUTSTANDING: <breakdown> one or more signals fired
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable worktree, a change name outside the allowlist, or a file that
# exists but cannot be read. The VERDICT carries the answer, not the exit
# status — see check-finish-preflight.sh's header for why this repository
# separates them. Finish runs this once per worktree, so each verdict names the
# worktree it judged; a bare CLEAR would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-cleanup-complete.sh's. That guard says
# COMPLETE/LEFTOVER, this one says CLEAR/OUTSTANDING, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence:
# here a file that is missing counts AGAINST the change, because a missing
# record proves nothing about the work; there a missing artifact IS the answer,
# because the question is "is it gone?". A reader who carried the meaning of
# CLEAR across to COMPLETE would carry that inversion with it. Distinct words
# are what stop the two verdicts reading as one. check-finish-preflight.sh's
# third vocabulary (RUN1/RUN2/REFUSE) differs for a further reason again: it
# selects a procedure rather than reporting a state.
#
# A MISSING FILE COUNTS AS OUTSTANDING, NOT CLEAR. Treating an absent plan or
# an absent finding record as clearance is the failure this guard exists to
# prevent: a branch merged over unfinished work because nothing was written
# down.
#
# SIGNAL TWO READS THE STORE, NOT A RENDERED FILE. A change's findings are
# rows in the store, keyed by change name, and `flow record findings -change
# <name> -C <worktree>` answers with the decoded JSON array — the same verb
# check-panel-reproducers.sh reads. There is no longer a rendered panel record
# for this guard to resolve a path to, and no missing-file case for it either:
# a change the store has never heard of, or one that genuinely raised no
# findings, answers `[]` at exit 0, which is CLEAR on this signal.
#
# A BARE `withdrawn` IS OPEN, not closed (KAN-791), the same line
# check-panel-findings-closed's duplicate of this predicate draws: the word
# with no reason after it is the reasonless drop the finding-status contract
# forbids, so only `withdrawn <reason>` counts as closed on this signal.
#
# Both signals are counted independently and reported together on the one
# line, so the operator sees the whole picture in one prompt rather than being
# sent back around the loop one signal at a time.
#
# THE STORE CALLS ANCHOR AT THE PLAN'S PROJECT (KAN-260). The change's
# records — findings and verdicts alike — live under the project of the repo
# that ran /flow, which is the repo where the plan lives, and the project key
# derives from -C's git common dir. So the findings query and both verdict
# calls pass -C anchored at STORE_ANCHOR: the resolved plan's directory when
# the plan is not local to the judged worktree, the worktree itself when it
# is or when nothing resolved. A satellite run anchored at its own worktree
# would read the SECOND repo's project — an empty findings array read as
# CLEAR on this very verdict line — and record the verdict where the
# canonical project's tools never look. -worktree keeps naming the worktree
# the guard was asked to judge, never the anchor.
#
# WHY THE CHECKLIST PATTERN IS ANCHORED TO COLUMN 0. `- [ ]` is matched only at
# the very start of a line — no leading whitespace at all. A task line sits at
# column 0 by this repository's task/step grammar (`build-green.md`); a step's
# checkbox is indented two columns beneath its task, and `implement.md` is
# explicit that "a step's checkbox tracks the step and gates nothing" — so a
# ticked or unticked step must never move this guard's count. Column-0
# anchoring also keeps the original reason for anchoring at all: unanchored,
# `- [ ]` also matches a plan that merely QUOTES a checklist inside a fenced
# example, which this repository's own plans do. A guard that fires on every
# plan that documents a checkbox is a guard the operator learns to click past.
# Fence tracking was rejected as the fuller fix: it is real complexity, and
# what it would still catch — a fenced example whose line BEGINS with `- [ ]`
# at column 0 — errs toward OUTSTANDING, which prompts the operator rather
# than clearing the gate silently.
#
# A SATELLITE'S MISSING LOCAL PLAN IS NOT THE MISSING-RECORD CASE ABOVE
# REJECTS. "A missing file counts as OUTSTANDING, not CLEAR" is about a
# change whose plan was simply never written — the absence itself is the
# only evidence there is, and treating it as clearance is the exact failure
# this guard exists to prevent. A satellite (KAN-363) carries no local
# `tasks.md` by design: `link.md`'s `## Part of` names a peer tree and a
# canonical change id, and the plan the operator actually needs to check is
# there, not here. Reading that absence the same way as a genuinely missing
# plan would report OUTSTANDING with nothing an operator can act on — the
# plan is not missing, it is elsewhere, and this guard's job is to reach it,
# per `scripts/lib/change-plan.sh` (task 7) and design.md's
# `guards-take-the-canonical-worktree-path`. So a satellite takes one of two
# outcomes and never the third: its canonical plan resolves and is counted
# exactly like a local one would be, or it does not resolve at all and the
# guard refuses outright (exit 2, "cannot determine anything") rather than
# reporting OUTSTANDING over a plan it never actually read. A change
# directory that carries a `link.md` with no `## Part of` — the canonical
# side's own copy, which names its `## Parts` instead — is not a satellite by
# this definition and falls through to the ordinary missing-plan case: this
# guard is never asked to judge a canonical change directory as if it were
# one of its own parts.
#
# HOW TO HAND-VERIFY AN OUTSTANDING VERDICT (KAN-446). A verdict this guard
# prints can be a structural false positive — the KAN-423 shape: a cross-repo
# change whose plan resolves only in the canonical worktree, so the worktree
# this verdict names holds neither signal. Verify from the primary records
# before acting on the verdict: resolve the plan the way this guard does
# (scripts/lib/change-plan.sh) and count column-0 `^- \[ \]` lines by hand;
# run `flow record findings -change <name> -C <worktree>` and expect `[]`.
# Both clean means the verdict was structural — record it (unfinished-work-gate.md's
# false-positive course) rather than trusting it or silently overriding it.
# Anything else means the verdict was right, and the courses it offers stand.

# Header note for the Go port (KAN-778): scripts/lib/change-plan.sh's
# resolution now runs as its Go port, stats/internal/guard/changeplan.go, and
# the body's reasoning for every branch is in
# stats/internal/guard/unfinishedwork.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-unfinished-work: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-unfinished-work 2 "check-unfinished-work:" "$@"
