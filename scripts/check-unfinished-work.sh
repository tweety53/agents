#!/usr/bin/env bash
# check-unfinished-work.sh — report whether a change carries unfinished work,
# for /flow's integrate-run gate.
#
# Usage: check-unfinished-work.sh <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...
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
# SIGNAL TWO ALSO COUNTS THE PANEL REPORTS (kan-1015), only in the worktree
# holding the plan: <n> panel report(s) count findings the store does not
# hold (<list>) — record each with `flow record finding -change <name> -C
# <anchor>` before landing. Each `.superpowers/sdd/panel-report-<round>-<slot>.md`
# whose `Critical:` plus `Important:` exceeds the store's rows of that round
# and slot is listed as `<file> (report <n>, recorded <m>)`, the count
# check-panel-findings-closed.sh makes.
#
# SIGNAL THREE: A DISPATCH ROW WITH NO OUTCOME IS OPEN (kan-929). `flow record
# dispatches -change <name> -C <anchor>` is read the way signal two reads
# findings — a failed or unreadable read is exit 2 — and every row whose
# `outcome` is empty (never ended, or ended without -outcome) is named by key
# in the breakdown, so the run closes it with `flow record dispatch end`
# before the change lands with "Outcome: not recorded" in its ledger. The rows
# are the change's, so only the worktree holding the plan reads them — a
# multi-repo change lists each open row once, not once per worktree.
#
# SIGNAL FOUR: A TICKED APPENDED TASK NO DISPATCH ROW NAMES IS UNFINISHED
# (kan-979). Only the last <n> tasks of the plan are judged, <n> from its
# `**Tasks appended:** <n>` header line (absent: 0), and every task titled
# `In-run fix <k> — …` wherever it sits — verify-fix-loop.md's loop appends
# those without raising the count (kan-941: tasks 12-16 went unjudged);
# a row names a task by its `taskId` or by the `+`-joined ids of a
# `task-<ids>[-…]` key. Plan-time tasks run grouped or inline and carry no row
# of their own, so they are never judged.
# A named appended task also needs a panel reviewer row — role `reviewer`, key
# `panel-<round>-<slot>` — whose seq is greater than the earliest non-reviewer
# row naming it: the fix run's panel reviews it under keys that name no task,
# so any panel reviewer row of the change counts, and a reviewer row under any
# other key does not.
# A ticked appended task also needs a commit carrying `Task-Id: <n>` in one of
# the `<worktree>=<merge-base>` ranges passed — one per worktree of the
# change's resolved set, since a cross-repo fix task may commit in a peer
# repository: a dispatch row back-filled after the fact is not a commit. A
# plan with a ticked appended task and no range passed is exit 2, and a range
# git cannot read is exit 2.
#
# SIGNAL FIVE: A PANEL ROUND DISPATCHED WITH NO PASS ROW IS UNFINISHED
# (kan-939 self-review). Every round a reviewer row keyed
# `panel-<round>-<slot>` ran in needs a pass row of that round, read with
# `flow record passes -change <name> -C <anchor>` — every panel run and
# fix-round pass records one (skills/flow/review-panel.md). The pass rows are
# read only when such a round exists; a failed or unreadable read is exit 2,
# like the dispatch read. Like signal three, only the worktree holding the
# plan reads them.
#
# SIGNAL SIX: A CHANGE WITH NO RECORDED SUMMARY IS UNFINISHED (kan-957
# self-review: 35 of 36 gymie changes landed with no summary row). The
# summary is read through `flow self-review bundle -change <name> -C
# <anchor>`, whose header — the lines above its first `## ` section — names
# it `skipped: change summary (absent)` when the store holds none. A failed
# read, or a header with no `found:` line, is exit 2, like the dispatch read.
# Like signal three, only the worktree holding the plan reads it.
#
# A CALL PASSING NO RANGE GETS ITS OWN WORKTREE'S RANGE DERIVED (kan-989
# self-review): the base resolve-base-branch.sh names, then `git -C <worktree>
# merge-base HEAD origin/<base>`. A range that cannot be derived — no `origin`,
# no git — leaves the call exactly as one passing none.
#
# SIGNAL SEVEN: A CODE COMMIT NO COMPLETED INLINE VERIFY ROW COVERS IS
# UNFINISHED (kan-989 self-review). The newest commit in any range touching a
# path outside `<spec root>/changes/`, authored after the newest `startedAt`
# of the `verifier` rows with outcome `completed` and key `verify` or
# `verify-…` — or with no such row at all — names its short sha in the
# breakdown. A re-run after run 1's drop is judged from the newest drop
# commit on, so only a fix run's own code asks for a newer row. Like signal three, only
# the worktree holding the plan reads it.
#
# SIGNAL EIGHT: A CODE COMMIT WITH NO REVIEW-PANEL PASS IS UNFINISHED (kan-970
# self-review: four implementer groups and the panel ran unrecorded, and the
# gate cleared the change). When signal seven's walk finds a commit outside
# `<spec root>/changes/` and no `panel-<round>-…` reviewer row exists — signal
# five judges a change that has one — `flow record passes -change <name> -C
# <anchor>` must hold at least one pass row; a failed or unreadable read is
# exit 2, like signal five's. Asked exactly where signal seven is.
#
# SIGNAL NINE: A VERIFY-REPORT FAILURE NO LATER COMMIT ADDRESSED IS UNFINISHED
# (kan-994 self-review, decision verify-failure-signal). Read in the worktree
# holding the plan whenever a range exists: in each
# `.superpowers/sdd/verify-report-*.md`, a `- verify: exit <n>` item with <n>
# not 0 and its whitespace-indented continuation lines, up to the next `- `
# item, name failures — every `<spec>.spec.ts:<line>`, its capture the first
# `<name>.png` after it before the next spec. A failure is addressed when a
# range commit committed at or after the report's mtime touches
# `<name>.png` or `<name>-<anything>.png` (the spec's base name for a failure
# naming no capture), or when the change's narrative.md holds
# `Verify failure addressed: <spec>:<line> — <sha>` with <sha> a range
# commit, or when an earlier false-positive verdict of this change names
# `<spec>:<line>` in its reason, no path character before it and no digit
# after (kan-1010's self-review) — a store that cannot answer clears
# nothing — or when the report next after it by mtime opens its `- verify:`
# item `exit 0 — rerun of`, a rerun of exactly those failed specs that passed
# (kan-1022's self-review). Each other failure is one reason, `verify failure not addressed:
# <spec>:<line>[ <name>.png] (<report>)`; an unreadable report is exit 2.
#
# SIGNAL TEN: A MULTI-REPOSITORY CHANGE WHOSE MERGE ORDER MISSES A REPOSITORY
# IS UNFINISHED (kan-995 self-review, decision merge-order-signal). Read in
# the worktree holding the plan when more than one range is passed: every
# range whose worktree's main checkout differs from the plan worktree's must
# be named by a numbered line of the plan directory's link.md under
# `## Merge order` (up to the next `## `), its backticked name a peer the
# spectre/peers file resolves to that main checkout. Each repository not
# covered is one reason, `merge order does not name <main checkout>:
# <planDir>/link.md`; an absent link.md holds no entries, an unreadable one
# is exit 2. A `## Part of` section is never the merge-order record.
#
# THE VISUAL VERDICT IS FOLDED IN (kan-989 self-review): over this worktree's
# own range, check-visual-verify-dispatched.sh's VISUAL-VERIFY-MISSING line
# is one more reason, and its exit 2 is this guard's exit 2 — unless an
# operator flagged that line a false positive (a check-visual-verify-dispatched
# verdict row of this change on this worktree ending with its "HEAD <sha>"),
# the flag integrate-change.sh's visual gate already honours.
#
# All signals are counted independently and reported together on the one
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
# per stats/internal/guard/changeplan.go (task 7) and design.md's
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
# (stats/internal/guard/changeplan.go) and count column-0 `^- \[ \]` lines by hand;
# run `flow record findings -change <name> -C <worktree>` and expect `[]`.
# Both clean means the verdict was structural — record it (unfinished-work-gate.md's
# false-positive course) rather than trusting it or silently overriding it.
# Anything else means the verdict was right, and the courses it offers stand.

# Header note for the Go port (KAN-778): the plan resolution runs in
# stats/internal/guard/changeplan.go, and the body's reasoning for every branch is in
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
