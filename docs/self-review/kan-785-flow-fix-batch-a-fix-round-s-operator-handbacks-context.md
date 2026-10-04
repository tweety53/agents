# Self-review context bundle for kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks.md (absent)
skipped: .superpowers/sdd/reviews/kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks-panel.md (absent)
skipped: spectre/changes/archive/kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks/tasks.md (absent)
skipped: spectre/changes/archive/kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks/design.md (absent)
skipped: spectre/changes/archive/kan-785-flow-fix-batch-a-fix-round-s-operator-handbacks/narrative.md (absent)

## git log --stat

commit 96a1b78b60c3d680e8218febd6ce6f2eaa9ad01e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:19:33 2026 +0300

    feat(flow): a fix round batches its operator handbacks into one ask

 skills/flow/review-panel-fix-round.md | 31 +++++++++++++++++++------------
 1 file changed, 19 insertions(+), 12 deletions(-)

commit f58d2850cef19e6b423710792442e1eb915aa570
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sun Oct 4 23:17:24 2026 +0300

    feat(flow): the operator-prompts contract states the batched-ask shape

 skills/flow-contracts/operator-prompts.md | 11 +++++++++++
 1 file changed, 11 insertions(+)

## Session narrative

A `/flow-fast` creating run for KAN-785, which asks that a fix round batch its outstanding
operator handbacks into one ask instead of one ask per finding. The lessons home held nothing on
operator handbacks, so the ticket's own text was the specification; a survey located the
judgment-call handback in `skills/flow/review-panel-fix-round.md`'s non-convergence loop and four
further per-finding "put to the operator" sites in the same file's reproducer guards, with the
ask's shape owned by `skills/flow-contracts/operator-prompts.md`. The batched-ask shape went into
the shape contract as **Batched asks** (stated once, per its doctrine), and the fix-round loop
now requires the walk to run to its end before one batched ask collects everything. Where it
struggled: `check-verbatim-moves.sh` is the real reviewer for prose changes here — the first run
acknowledged three of four new sentences but refused the heading until it was re-listed with the
guard's leading-`\` form, and after Task 2's edits fifteen lines needed acknowledging, most of
them the same sentence printed twice (once as a deletion, once as an addition). The `sort -u`
collapse of `verbatim-moves.txt` mixed the comment header into the entries; harmless to the
guard, slightly untidy to read.
