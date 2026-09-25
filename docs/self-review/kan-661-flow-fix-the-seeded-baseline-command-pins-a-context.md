# Self-review context bundle for kan-661-flow-fix-the-seeded-baseline-command-pins-a

found: 1 of 7 sources; skipped: 6 of 7 sources
note: RECORDS LOSS — a flow.review-panel stage run completed for kan-661-flow-fix-the-seeded-baseline-command-pins-a, but the store holds no dispatch rows for it: the run's dispatch and finding records never reached this store, most plausibly written to a per-workspace database later removed at cleanup. The ledger and panel sources below are absent or degraded for that reason, not because no panel ran.
skipped: change summary (absent)
skipped: .superpowers/sdd/ledgers/kan-661-flow-fix-the-seeded-baseline-command-pins-a.md (absent)
skipped: .superpowers/sdd/reviews/kan-661-flow-fix-the-seeded-baseline-command-pins-a-panel.md (absent)
skipped: spectre/changes/archive/kan-661-flow-fix-the-seeded-baseline-command-pins-a/tasks.md (absent)
skipped: spectre/changes/archive/kan-661-flow-fix-the-seeded-baseline-command-pins-a/design.md (absent)
skipped: spectre/changes/archive/kan-661-flow-fix-the-seeded-baseline-command-pins-a/narrative.md (absent)

## git log --stat

commit 795475b4f2f2d9f17da62ba8f4603093d956a39b
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 21:31:19 2026 +0300

    docs(planner): the Baseline measured command names no ref

 skills/flow/brainstorm-planner.md | 6 +++++-
 1 file changed, 5 insertions(+), 1 deletion(-)

## Session narrative

This run fixed the seeding template for baseline `measured:` commands: `skills/flow/brainstorm-planner.md`'s `**Baseline:**` bullet now states that the recorded command names no ref of its own — no tree, sha, branch or revision argument — because `check-task-commit-fields.sh`'s `_run_measured_at` checks each re-measurement point out itself and runs the command in that working tree; a command naming a ref measures that one tree at both points and always reports identical counts. The comment's `@ <ref>` is stated to be provenance annotation, never part of the command. The run struggled most with locating the defect site: the issue describes a template but no literal command template exists anywhere in the repository — the shape a planning session follows is the bullet itself plus `skills/flow-contracts/plan-provenance.md`'s `measured:<command> @ <ref>` grammar, and the guard, its parser, and its test fixtures were all verified to already use the correct ref-free shape before the bullet was concluded to be the only defect site. Deliberately left out: any edit to `plan-provenance.md` (its grammar is correct for plan-time provenance; the bullet owns the re-measurement contract), any guard or fixture change, and every `-rationale.md` appendix. Decided as `micro` (one task, one file): execution inline in this session, no review panel, no groups.
