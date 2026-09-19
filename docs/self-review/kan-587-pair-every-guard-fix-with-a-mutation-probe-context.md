# Self-review context bundle for kan-587-pair-every-guard-fix-with-a-mutation-probe

found: 0 of 6 sources; skipped: 6 of 6 sources
skipped: .superpowers/sdd/ledgers/kan-587-pair-every-guard-fix-with-a-mutation-probe.md (absent)
skipped: .superpowers/sdd/reviews/kan-587-pair-every-guard-fix-with-a-mutation-probe-panel.md (absent)
skipped: spectre/changes/archive/kan-587-pair-every-guard-fix-with-a-mutation-probe/tasks.md (absent)
skipped: spectre/changes/archive/kan-587-pair-every-guard-fix-with-a-mutation-probe/design.md (absent)
skipped: spectre/changes/archive/kan-587-pair-every-guard-fix-with-a-mutation-probe/narrative.md (absent)
skipped: git log --stat (absent)


## Branch log

commit fc652b1450b0feaf24f8a9eb6913da4df73f2513
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 20:52:30 2026 +0300

    flow(review-panel): prove a guard fix by its own pre/post observable

 skills/flow/review-panel.md | 15 ++++++++++++++-
 1 file changed, 14 insertions(+), 1 deletion(-)

## Session narrative

This /flow-fast run resolved KAN-587 ("pair every guard fix with a mutation probe that flips the failure it fixed") to a single-file prose change in `skills/flow/review-panel.md`: the fix-round mutation-proof section gained a paragraph making a guard script'\''s fix provable by the guard'\''s own observable, measured pre-fix (probe fails, defect reported) and post-fix (probe passes), recorded in the third field of the `fix-mutation:` ledger line as `<pre>→<post> <what the observable counts>`, and the MUTATION PROOF dispatch paragraph gained the matching sentence so the fix subagent runs both sides itself. The struggle worth naming: the decide step and the plan-shape guard consumed most of the run'\''s care — an early tasks.md draft indented its field lines past column 0, which the shape guard rejected (fields invisible to the grammar), and classifying the change as micro (1 task, 1 file) is a judgment a reviewer should double-check, since it is what collapsed the panel roster to `default` and skipped every dispatch; the harness task list, decision.json and the two contract edits were then verified through the project'\''s full lint list (25 guards plus gofmt/go vet/tsc after a fresh `make web-build`), which is green.
