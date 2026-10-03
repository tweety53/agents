# Self-review context bundle for kan-777-parent-shell-discipline-in-contracts

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-777-parent-shell-discipline-in-contracts, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-777-parent-shell-discipline-in-contracts.md (absent)
skipped: .superpowers/sdd/reviews/kan-777-parent-shell-discipline-in-contracts-panel.md (absent)
skipped: spectre/changes/archive/kan-777-parent-shell-discipline-in-contracts/tasks.md (absent)
skipped: spectre/changes/archive/kan-777-parent-shell-discipline-in-contracts/design.md (absent)
skipped: spectre/changes/archive/kan-777-parent-shell-discipline-in-contracts/narrative.md (absent)

## git log --stat

commit 80a23e6e6738e5070a30ca9c33a601f4ed524b34
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 00:25:58 2026 +0300

    feat(flow): state the parent session's shell discipline in the implement contract

 skills/flow/implement.md | 10 ++++++++++
 1 file changed, 10 insertions(+)

## Session narrative

The run placed the ticket's one-paragraph shell-discipline statement in `skills/flow/implement.md`, in **The parent orchestrates directly** — the parent-facing implement contract the ticket named — after `flow lesson resolve` confirmed no brief or narrative in the lessons home already states it. The run's real friction was `check-verbatim-moves.sh`: a flow-fast run never writes `<project>/spectre/`, so the guard's acknowledgement route (`spectre/changes/<name>/verbatim-moves.txt`, which every prose-adding `/flow` run here uses) was unavailable, and the guard's only other pass for a new run-loaded sentence is a `(`skills/….md`)` citation inside the sentence itself. The paragraph was therefore authored with one provenance citation per sentence — each naming a real file whose heading matches the bold token beside it, so `check-references.sh` stays green — and a first draft failed the guard once more because the bold label's trailing period split the label off as its own uncited sentence; the corpus's colon-label form fixed it. `check-normative-inventory.sh`'s output was captured against the base and diffed byte-identical after the edit, so no normative sentence moved. A tension is left on the record for whoever edits these contracts next: flow-fast changes that must add deliberate rule prose have exactly one guard-clean shape — citation-shaped sentences — and if a change ever needs prose a citation cannot honestly carry, it has no acknowledgement mechanism at all without breaking the flow-fast no-spectre rule.
