# Self-review report for kan-1064-flow-self-review-run-the-coordination-around-the

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the self-review skill's new "no no-progress escalation" sentence also covered the pipeline-fix loop, whose flow-high escalation row stays; two cross-references went stale — fixed: 676a6c4c
- **[flow-fix]** the planner tagged a diff `verified:` after a scratch apply without running the lint list there, so task 2 failed two guards at implementation — fixed: a1d10614
- **[flow-fix]** a flowd restart by another session failed a `flow tasks count` call because the client never retries a refused dial — fixed: ffe24012
- **[flow-fix]** `pipeline-fix.sh end <key>-pass` reported nothing to close while `check-loop-dispatch.sh` refused the plan dispatch on the still-open pass row — fixed: d992596f
- **[flow-fix]** the inline panel fix began as `panel-fix-0` and after its edits; both guards refused with self-explaining messages, no contract gap — noted
- **[flow-fix]** task 1's baseline counts went stale after the pre-task rebase and were re-measured at task close, the designed correction path — noted
- **[flow-fix]** the plan's absence sweep missed a paraphrase of the moved fact; the round-0 primary slot caught it in one round, as designed — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** planning dominated the run's spend (plan-2 read about 4M cache tokens); nothing in the bundle shows waste to cut — noted
- **[flow-cost]** the inline verifier row records zero tokens because run-verify-list.sh opens and closes it within one call; accurate attribution — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** a small class run inline closed with one panel fix round and a clean re-review — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** none — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** none — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** the round-0 mutation slot launched about six minutes after the reading slots because it waits for the full suite (kan-1039), by design — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review candidate: a stale contradicting sentence left behind when a fact moved, not mined while mining is paused — noted
  - severity: Important (panel round 0, primary)
  - location: skills/flow-self-review/SKILL.md step 7
  - reviewer's sentence: "Step 7 still says \"a pass inside a run has no bundle file and commits the report alone\", but this change writes the in-run bundle to `<sr-dir>/bundle.md`"
  - fixed code: commit 972a4129 (extracted under the state dir's flow-eval/kan-1064-flow-self-review-run-the-coordination-around-the/)
  - why: a classic moved-fact contradiction a first-round reviewer must catch

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** none — noted
