# Self-review report for kan-1028-mutation-failure-modes-extra-dispatch

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** A decided panel's `panel.dispatches` was dispatched literally, so a pre-kan-1028 decision listing `failure-modes`/`mutation` would run them twice; a dropped role now leaves its bundle too — fixed: db6cdfd9
- **[flow-fix]** `handoff-blocks.md` still carried the old `## Learning — <topic>` section in the `IN_PROGRESS` template; it now prints the `**Learning:**` path line — fixed: 0b417323
- **[flow-fix]** `/flow-settings` still offered `failure-modes` and `mutation` though every resolved roster drops them; both retired from `ValidReviewers` with migration 0044 — fixed: 29db9a0d
- **[flow-fix]** Session tokens derived from the change key plus a letter collided across sessions (`mf-k1028b`); **Stage marks** now requires a UTC-second timestamp suffix — fixed: 557f69d6
- **[flow-fix]** A task appended after Decide left `decision.json` inputs stale (`tasks=5 files=20` vs `6/23`); `implement.md` now re-decides, and a corrected decision row was recorded — fixed: f741896f
- **[flow-fix]** Dev-postgres fast shutdowns at 16:27:40Z and 16:54:30Z were a manual external restart; no pipeline tool call caused them — noted
- **[flow-fix]** F6's withdrawal on the simplest-build rule holds: the plan prescribed the struct and no second caller exists — noted
- **[flow-fix]** "mutation is always the last pass" vs the bundle name `mutation+failure-modes` is no contradiction: `renderslotprompt.go` renders no mutation PASS — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The dead session's aborted round-0 dispatches cost ~267k cache-read tokens before re-dispatch; a consequence of the session death — noted
- **[flow-cost]** Pass 1 `primary+principles` on opus/medium (2.38M cache-read) raised both Importants; the cost bought them — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Inline execution finished 6 tasks in ~8 minutes, every task commit guard-green — noted
- **[flow-improvement]** The post-fix re-review on sonnet/low ran in 27s, clean, ~98k cache-read — noted
- **[flow-improvement]** The panel caught the planner's default-only drop rule (F1) before landing — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** A BSD `sed` range-quoting failure extracting a contract section cost one retry; `Read` with an offset already covers it — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The verifier's 7m35s was mostly the full suite plus the `internal/store` re-run forced by the external postgres restart — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, decided-panel double dispatch; not chosen for mining — noted
  - severity: Important; location: `skills/flow/review-panel.md:159`
  - reviewer: "a decided panel recorded before kan-1028 runs failure-modes and mutation twice"
  - fixed in: 3b07580a, squashed into 3ee5df68
  - why: cross-file reasoning catch, legacy decision plus no-key fallback
- **[flow-eval]** impl case — Task 1, `plan-class.sh` second extra line; not chosen for mining — noted
  - code: `/Users/tweety53/Projects/agents-worktrees/kan-1028-mutation-failure-modes-extra-dispatch/stats/internal/guard/planclass.go:74`
  - why: small, fully specified Go change with exact stdout and red-then-green test
