# Self-review report for self-review-backlog-flow-5-oct18

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The loop-dispatch hook denied every dispatch naming a /flow change called `self-review-…`; it now skips a name whose `spectre/<name>` branch exists (in-run pipeline fix) — fixed: eebc4f8c
- **[flow-fix]** The plan's global constraints carried the loop-only "no `Task-Id:` trailer" rule, which close-task.sh refuses under /flow; check-plan-shape now refuses that rule in a plan whose change has a `spectre/<name>` branch — fixed: 6609137c
- **[flow-fix]** `flow record dispatch` refused `-effort unknown`; model-policy.md already names `default` and the refusal message says so — noted
- **[flow-fix]** `spectre/peers` lacked `gymie-playwright` and stopped the merge-order write; the stop is the intended check and this change declared the peer — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The `verify-gymie-playwright` ledger row carries 2.39M cache-read tokens against ~72k for its siblings; the parent's turns are attributed to the longest-open inline window, as designed — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The plan carried diffs verified at plan time, so the five tasks took under two minutes and the panel was clean on its first pass — noted
- **[flow-improvement]** Each backlog ticket was reproduced before it was taken; seven stale tickets were reported with evidence — noted
- **[flow-improvement]** The live check ran a before/after pair on a throwaway `flow_<id>` database — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The twelve-minute gap between the last task and the panel was the hook denial and its in-run fix (eebc4f8c) — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — task 1, `replaysSharedJournal` (KAN-913); operator chose not to mine — noted
  - Commit: `7b8e8c07`; code `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/self-review-backlog-flow-5-oct18/main.go:421`
  - Tests: `TestReplaysSharedJournalOnlyIntoTheFlowDatabase`
  - Why: small exact fix with a near-miss input (`flowd`) and an observable before/after effect
- **[flow-eval]** impl case — task 2, setup.sh linked-worktree refusal (KAN-773); operator chose not to mine — noted
  - Commit: `7b8e8c07`; code `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/self-review-backlog-flow-5-oct18/setup.sh:1223`
  - Tests: `TestGlobalInstallRefusedFromLinkedWorktree`
  - Why: the refusal must tell the operator's real home from a sandboxed `$HOME`
