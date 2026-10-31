# Self-review report for kan-1072-audit-whether-archived-spectre-change-artifacts

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** commit-split scoped the plan commit to all of `spectre/changes/` rather than the change folder, folding the tip cleanup's archive deletion into `chore(spectre): plan` commits across six repositories — fixed: eace1362
- **[flow-fix]** implement.md named a group implementer's key `task-<n>-implementer` where the hook requires `task-<ids>-implementer`, aborting the first group-1 row — fixed: 73c17540
- **[flow-fix]** verify-and-handoff.md:83's archive exclude and check-task-commit-fields.sh:50's "non-archived" sit outside the design's grep allow-list; both are still accurate, and the exclude still covers this change's own leftover archive — noted
- **[flow-fix]** Task 6's planned specroot.go archive-skip removal was not made; the skip is still needed for the archive this change's own landing leaves — noted
- **[flow-fix]** the gated reviewer was first begun at the wrong effort; model-policy states the pair plainly and the hook refused the slip at once — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** panel F18 re-raised a deferral the design accepted; its correct fix cost panel round 2 (about 2.5 min); no defect, as the decision did not forbid closing the gap — noted
- **[flow-cost]** planning (plan-3, 6.1M cache read) and implementer group 1 (34.0M) dominated, both following scope the operator widened twice — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** gated reviewers reproduced F1, F2 and F4 with overlay tests or a worktree-built flow-guard before raising them, and panel fixes carried fix-mutation proofs; keep the reproducer requirement in both briefs — noted
- **[flow-improvement]** the live check measured all three behaviours before and after against stand-ins (scratch clones, a store copy on 4175), never the dev store — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** check-plan-shape does not compare the paths a task's steps edit with its Files, so every implementer group needed Files corrections at close — fixed: 601e90cf

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** none — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** `flow self-review bundle` ran on the generic 2 s store timeout while its git-history assembly takes about 2-3 s, so the first call timed out — fixed: af81d2ac
- **[flow-speed]** implementer group 1 (seven Go tasks, 44 min) was the critical path; the grouping is recorded as deliberate for one Go package, and no safe faster split is visible — noted
- **[flow-speed]** the round-0 mutation slot launched 3m39s after the reading slots, the designed wait for the full suite — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review candidate: a polled board read fell back to whole-history `git log --all` per row (not mined — mining ask suspended) — noted
  - severity: Critical (panel round 0, call-efficiency, F21)
  - location: stats/internal/api/stats.go:852
  - reviewer's sentence: "readPlanFile now falls back to `landed.File` for every row whose live tasks.md is missing. `toStateBoardDTOs` (stats.go:866 planTasks, stats.go:895 selfReviewTasks) runs it for every non-finished row on every StateBoard/RunDetail poll, and both screens poll every 5s"
  - fixed code: commits b0714512 and 10664f28 (extracted under the state dir's flow-eval/kan-1072-audit-whether-archived-spectre-change-artifacts/)
  - why: the gated reviewer's own fix suggestion introduced it; it needs tracing a fallback onto a polled render path
- **[flow-eval]** review candidate: a removed exemption left a fix run's restored plan failing check-task-records (not mined — mining ask suspended) — noted
  - severity: Important (gated per-task reviewer, task 6, F4)
  - location: stats/internal/guard/taskrecords.go:274
  - reviewer's sentence: "Removing the archived-plan exemption with no drop-keyed replacement: a fix run's restored tasks.md fails check-task-records (a lint step) for every task ticked before the drop, whose commits run 1 squashed; reproduced with the worktree-built flow-guard (exit 0 after drop, exit 1 after restore)."
  - fixed code: commit 8601f1b2 (extracted under the same directory)
  - why: a defect of absence, visible only by tracing the squash, drop and restore lifecycle

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** none — noted
