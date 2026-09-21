# Self-review context bundle for kan-648-flow-cost-mutation-panel-bounded-entry-context

found: 1 of 7 sources; skipped: 6 of 7 sources
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-648-flow-cost-mutation-panel-bounded-entry-context.md (absent)
skipped: .superpowers/sdd/reviews/kan-648-flow-cost-mutation-panel-bounded-entry-context-panel.md (absent)
skipped: spectre/changes/archive/kan-648-flow-cost-mutation-panel-bounded-entry-context/tasks.md (absent)
skipped: spectre/changes/archive/kan-648-flow-cost-mutation-panel-bounded-entry-context/design.md (absent)
skipped: spectre/changes/archive/kan-648-flow-cost-mutation-panel-bounded-entry-context/narrative.md (absent)

## git log --stat

commit c20fd20caccd5e372d55e24d2cba28b4c4bebbc9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 01:36:52 2026 +0300

    docs(flow): bound the mutation slot's entry context

 skills/flow/review-panel.md | 11 +++++++++++
 1 file changed, 11 insertions(+)

## Session narrative

This run carried KAN-648 from the ticket to a landed-ready branch: the measured defect was the
review panel's mutation slot reading ~16.9M cached tokens in kan-574's round-0 dispatch — about
four times the primary slot — because the mutation-testing brief's search for the tests a
behaviour's mutations target was named nowhere, so the dispatched slot swept the whole tree. The
fix adds one paragraph to `skills/flow/review-panel.md`, directly after the `[TOUCHED_FILES]`
resolution: the mutation slot's dispatch prompt now carries a MUTATION ENTRY CONTEXT paragraph —
`final-review.diff` at each worktree's merge-base sha as the diff base, the plan's `**Tests:**`
tests and the touched-files list's test files inlined beneath it, and the discipline that the
brief's covering-test search begins from those two, never a whole-tree sweep, with an escape hatch
(read outside the list only when a behaviour cannot otherwise be judged, named in the report). The
run made four judgment calls worth recording: the change belongs to the agents repo, not gymie
where the session opened — KAN-521's reviewer-side precedent (bd1395c) and the file itself live
here, so the run relocated and the gymie worktree and branch it had briefly created were removed;
Bugbot is deliberately untouched, because only the mutation slot was measured and KAN-521's
precedent scoped to the measured slot; the new paragraph is not pinned in
`check-dispatch-paragraphs.sh`'s table, following the same precedent's prose-only shape and
keeping the change micro; and "the tests its mutations target" resolves mechanically from the plan
and the touched-files list rather than any inferred coverage map. It struggled once, briefly: the
flow store reported unreachable for one stage-mark write and journaled it, recovering on the next
call. Verification ran the project's whole `## lint` list (all 29 commands, including the
stats Go/SPA checks after a fresh `make web-build`) plus the six guard harnesses that reference
`review-panel.md`, all green, and the normative inventory is byte-identical to base — no SHALL or
MUST sentence added.
