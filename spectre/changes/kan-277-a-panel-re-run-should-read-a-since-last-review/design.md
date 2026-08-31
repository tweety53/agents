## Context

`skills/flow/review-panel.md`'s Panel re-runs section already tracks, per slot, the HEAD
sha it last reviewed, and already writes `slot-delta-<round>-<slot>.diff` from
`git diff <that sha> HEAD` on a Full re-run — that mechanism landed in `72f4f0e`
(2026-08-26), after KAN-77 (2026-08-21) hit the cost this change addresses. What is still
missing is a size check that reflects it: `check-panel-diff-size.sh` is called once, under
"## The roster" (line 118 of `review-panel.md` as of this change), against
`<worktree> <merge-base>` — the full branch diff — with no distinction between pass 1 and
a later round, and Targeted re-runs call it not at all.

This is one change: one guard's call sites, in one file, gated the same way throughout.

## Decisions

### full-diff-is-correct-for-pass-1-and-full-slots

**ID:** full-diff-is-correct-for-pass-1-and-full-slots
**Status:** active
**Chosen:** Keep measuring the full branch diff for pass 1 and for any Full-re-run slot
that has no held last-reviewed sha (Primary, or a slot added mid-run) — one-line rationale:
that number is exactly what those slots are actually handed, so it is not overstating
anything for them.
**Considered:** Always measuring only scoped deltas — rejected because Primary's read is
genuinely the full diff every round by design (the table in `review-panel.md`'s Panel
re-runs section), and a check that never reflects that would silently pass panels whose
real bottleneck slot is over cap.

### gate-on-per-slot-max-not-aggregate

**ID:** gate-on-per-slot-max-not-aggregate
**Status:** active
**Chosen:** On a Full re-run, gate the over-cap prompt on
`max(full-branch count when a full-diff slot runs this round, largest scoped-delta count
among returning slots)` — the largest single read any one dispatched agent actually faces
this round — rather than the aggregate branch-wide diff. Report the full-branch count
alongside in the prompt and in `final-review-panel.md`, so the operator still sees branch
growth even when it isn't what gated.
**Considered:** Reporting the scoped number as extra context but leaving the aggregate
full-branch diff as the sole gate — rejected per the operator's own answer: it does not
stop the repeated late-round prompting the ticket describes, since the full-branch count
only grows.

### targeted-gets-its-own-check

**ID:** targeted-gets-its-own-check
**Status:** active
**Chosen:** Add a cap check to Targeted re-runs too, gated on `fix-round-N.diff`'s own
size — the one diff every Targeted-re-run slot reads. No per-slot max is needed there since
there is only one diff in play.
**Considered:** Leaving Targeted uncovered, since KAN-77's measured cost was entirely in
Full mode — rejected per the operator's own answer: Targeted diffs are small by
construction today, so the addition is low-cost, and consistency between the two modes
avoids a reader having to remember that only one of them is checked.

### no-interface-change-to-the-guard

**ID:** no-interface-change-to-the-guard
**Status:** active
**Chosen:** `check-panel-diff-size.sh` keeps its exact current signature
(`<worktree> <point> [cap]`, measuring `point..working-tree`). A scoped delta's size is
obtained by calling it again with the slot's last-reviewed sha (or `FIX_BASE` for
Targeted) as `<point>` — correct once the round's fixup is committed and the working tree
is clean, which the panel's own fix-round procedure already guarantees before this check
would run.
**Considered:** Extending the script to accept two explicit refs (`<point> <point>`)
instead of point-vs-working-tree — rejected as unneeded complexity: every call site this
change adds already has a clean working tree at the point it calls the script, so the
existing single-point form already gives the right number, and `test-check-panel-diff-size.sh`'s
existing contract keeps holding unmodified.

## Open questions

<!-- none -->
