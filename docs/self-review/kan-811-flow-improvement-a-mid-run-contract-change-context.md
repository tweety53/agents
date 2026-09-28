# Self-review context bundle for kan-811-flow-improvement-a-mid-run-contract-change

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-811-flow-improvement-a-mid-run-contract-change, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-811-flow-improvement-a-mid-run-contract-change.md (absent)
skipped: .superpowers/sdd/reviews/kan-811-flow-improvement-a-mid-run-contract-change-panel.md (absent)
skipped: spectre/changes/archive/kan-811-flow-improvement-a-mid-run-contract-change/tasks.md (absent)
skipped: spectre/changes/archive/kan-811-flow-improvement-a-mid-run-contract-change/design.md (absent)
skipped: spectre/changes/archive/kan-811-flow-improvement-a-mid-run-contract-change/narrative.md (absent)

## git log --stat

commit cddd38bd5fbf3cdfaba69cd03b8d0da873b2afa5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 22:30:34 2026 +0300

    docs(review-panel): record the governing rule beside each round, apply mid-run rule changes forward-only

 skills/flow/review-panel.md | 9 +++++++++
 1 file changed, 9 insertions(+)

## Session narrative

This /flow-fast run implemented KAN-811 as a one-paragraph contract edit to skills/flow/review-panel.md (commit cddd38bd): a rule that changes mid-run governs from the round it lands in, and a round decided under wording that changed records that wording beside its pass-log rows, so the rendered record shows which rule governed what and rounds the old rule closed are never re-decided. It struggled nowhere structurally; the one friction was environmental — check-normative-inventory.sh printed 7 sentences from 89 files both before and after the edit (byte-identical diff), far from the roughly-a-thousand-lines figure .flow/project.md's lint section quotes, which reads as drift in that prose rather than a guard problem and was left alone as out of scope. The bundle's panel-records note above is expected: the decide step classified the plan micro, so flow.review-panel was marked as an empty pair and dispatched nothing — the store holds no dispatch rows because no panel ran, not because records were lost.
