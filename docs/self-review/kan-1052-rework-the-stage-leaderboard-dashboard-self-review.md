# Self-review report for kan-1052-rework-the-stage-leaderboard-dashboard

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `ui paths` misses `skills/flow-contracts/model-policy.md`, which flowd embeds and the Experiments page renders, so bd8a8112 landed stale Experiments baselines — fixed: c153822c
- **[flow-fix]** `/flow-interactive` and `/flow-fast` land a `ui paths` diff without the `verify` row, so 1552c25b/9bf31b17 left seven state-board baselines stale and a swallowed hover border unseen — fixed: a78c130f
- **[flow-fix]** the Dispatch-sites implementer key shape `task-<n>-implementer` omits the `+`-joined group form the hook requires, costing one refused group launch — fixed: 73c17540
- **[flow-fix]** `test-run-capped.sh` case (a) asserts a wall-clock `elapsed<=6` bound that failed at 12 s under load though status 137 already proves the kill — fixed: a3e8e41f
- **[flow-fix]** run-verify-list's concurrent test commands oversubscribe the CPU, pushing a 235 ms vitest test past the 5 s timeout — fixed: cf44fbaf
- **[flow-fix]** the mutation slot's prompt lacks the never-symlink-`node_modules` rule, so the slot linked the shared directory into its copy — fixed: a21cc933
- **[flow-fix]** the round-1 mutation re-run wrote its verdict lines only to its message; check-rerun-citations caught it and the one designed send-back closed it — noted
- **[flow-fix]** two zsh slips (an unquoted multi-flag variable, a `$PWD:` modifier) in the parent's own shell calls, no pipeline text carries either pattern — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** planning dominated the spend (six planners, about 9.5M cache tokens) across operator-driven prototype rounds v1→v3; nothing in the bundle shows waste — noted
- **[flow-cost]** the inline verifier rows record zero tokens because run-verify-list opens and closes them inside one call; accurate attribution — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the always-on mutation slot raised seven Important test gaps in round 0, closed in one fix round with a clean round 1 — noted
- **[flow-improvement]** the live check against a dev-store copy matched on the first try against design.md's "not working looks like" list — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** the post-rebase re-verification cost about 21 minutes (verify-2, full-suite-fix-2, verify-3, visual-verify-2), caused by upstream stale baselines the F1 and F2 fixes remove — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review candidate: a sort test whose fixture also sorts the same way by the wrong key — noted
  - severity: Important (panel round 0, mutation)
  - location: stats/internal/store/aggregate.go:617; test stats/internal/store/aggregate_test.go:336
  - reviewer's sentence: "Spend-descending order of projects (`ORDER BY 2 DESC, 1` → `ORDER BY 1, 2`) survives: the test's fixture has A (spend 4) sorting before B (spend 2) alphabetically too."
  - fixed code: commit 24cc0c72 (extracted under the state dir's flow-eval/kan-1052-rework-the-stage-leaderboard-dashboard/review-F9/)
- **[flow-eval]** impl candidate: Task 1, return every costed stage and each project's spend from the stage leaderboard — noted
  - commit: 54e5d8c4; code impl-task-1/aggregate.go:609 under the state dir's flow-eval/kan-1052-rework-the-stage-leaderboard-dashboard/
  - tests: TestStageLeaderboardReturnsEveryCostedStage, TestStageLeaderboardProjects, TestStatsStageLeaderboardProjects
  - why: a precisely specified store + API task rippling through four fakes, with break-and-prove steps

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** plan-specified tests whose fixture or step order cannot tell the asserted rule from its nearest wrong alternative (panel F8, F9) cost a fix round; the plan-writing rules carry no such rule — fixed: f792237d
