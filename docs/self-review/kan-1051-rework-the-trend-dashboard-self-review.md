# Self-review report for kan-1051-rework-the-trend-dashboard

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** stage-task-files.sh did not stage paths deleted with git rm — fixed: 1372f365
- **[flow-fix]** close-task judged every suite zip of a close rather than only the last one its commits write — fixed: 56afc108
- **[flow-fix]** check-unfinished-work read a dropped re-run's plan at the drop commit and flagged pre-drop appended tasks as uncommitted — fixed: 4a26d5ef
- **[flow-fix]** after a fix run restores the change directory, the contract's since-close sha (the drop commit) counts the restored directory as 12,397 lines, so rounds 4 and 5 picked another commit by hand — fixed: 647ce423
- **[flow-fix]** panel reproducers write their Go test file into the shared worktree; one was left behind and failed a guard-test run mid-suite — fixed: e5951a3c
- **[flow-fix]** step 2 tells the self-review pass to write findings.md and report-draft.md, which the harness refuses for a subagent — fixed: 5f69f430
- **[flow-fix]** the inline verify-5 row stayed unmeasured although assistant records fall inside its window — fixed: ed043b7f
- **[flow-fix]** a write-in-progress was refused over an unrecorded run-instructions begin; the guard caught the skipped mark as designed — noted
- **[flow-fix]** a fix-run append was committed around the planning-commit chain, whose check-plan-shape step would have refused it; corrected before implement — noted
- **[flow-fix]** Task 9's first commit carried evidence PNGs; close caught it and a 41 s recommit dropped them — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** planning took four planner dispatches (29 min, about 9.5M cache read), following scope widened in design rounds — noted
- **[flow-cost]** three self-review passes of about 8M cache read each, one per product defect met and fixed — noted
- **[flow-cost]** a 30 s citation re-dispatch of mutation+failure-modes in round 1 — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the live check showed the failure look, which led to the finish-instant fix, then measured exact in-span attribution and an idempotent restart on a store copy — noted
- **[flow-improvement]** motion capture plus a requestAnimationFrame DOM read proved the hover-paint fix frame by frame — noted
- **[flow-improvement]** a two-worktree implementer wave ran tasks 1–2 and 3–6 in parallel in about 12 min — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** live-store-copy.sh has no restart, so the live check's idempotence step restarted the copy's flowd by hand — fixed: 02bbb8c2

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** a pipeline-fix loop runs verify and then land, which runs the same verify again at the same HEAD — fixed: 163570d8

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review candidate: the back-fill re-added spend a window swept back had already counted (not mined — mining ask suspended) — noted
  - severity: Important (panel round 0, primary, F1)
  - location: stats/internal/harvest/attribute.go:528
  - reviewer's sentence: "the back-fill double-counts spend on runs that were open at harvest time and later swept/clipped"
  - fixed code: commit f5f5b2de (extracted under the state dir's flow-eval/kan-1051-rework-the-trend-dashboard/review-F1/)
  - why: the defect is visible only by tracing an end mark moved after the record was counted
- **[flow-eval]** impl candidate: Task 1, gap spend to the stage that began last (not mined — mining ask suspended) — noted
  - task: 1. Attribute spend between two stage windows to the stage that began last
  - code: stats/internal/harvest/attribute.go:717 (extracted under the state dir's flow-eval/kan-1051-rework-the-trend-dashboard/impl-task-1/)
  - why: four worked boundaries, one built to fail a nearest-window rule

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
