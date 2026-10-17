# Self-review report for kan-989-stats-ui-dashboard-review

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the stage_runs, guard_runs and pass rows, commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The human-gate fix run closed flow.visual-verify and flow.verify without running them, and check-unfinished-work exited 2 (no merge-base range) while IN_PROGRESS was still written; the guard now derives the range, folds in the visual verdict (honouring a false-positive flag) and flags a code commit no completed verify row covers — fixed: 0397c8f7
- **[flow-fix]** A fix run's appended tasks bypassed check-plan-shape and check-plan-provenance; document-fix now runs both before its planning commit — fixed: 89100553
- **[flow-fix]** A panel key with a non-numeric round (`panel-fix1-primary`) was silently not counted as a round; `flow record dispatch begin` now refuses it — fixed: 800020d4
- **[flow-fix]** Guard and cmd/flow package tests and the shell guard harnesses posted guard telemetry into the dev store (6,065 rows); telemetry is now off in both TestMains, three hand-built envs and run-guard-tests.sh — fixed: 87c3b79b
- **[flow-fix]** The 2% visual tolerance let a stale empty Reviewers baseline pass; tolerance is now zero, 15 runs showed no flake, so no pin — fixed: 1f74527c
- **[flow-fix]** Off-convention implementer keys (`task-11-12`) recorded the wrong commit for task 11; such keys are now refused, `-retry` accepted — fixed: ac96854b
- **[flow-fix]** The failed verify row stayed open while the in-run fix loop opened `verify-fix-1`, stopping prepare; the loop now closes it first — fixed: 6fb15fde
- **[flow-fix]** The visual verifier invented exit 126 for a permission-denied capture; a denied command is now reported blocked with the denial text verbatim — fixed: b10774ea
- **[flow-fix]** main's 6bd7b020 left `integrate-change.sh:362` piping into an early-exit grep, failing TestNoPipeIntoEarlyExitGrep; now a here-string — fixed: fe3bae00
- **[flow-fix]** Task 13's FLOW_RECORDS_ADDR pin is on the branch and no other cmd/flow test pins FLOW_ADDR alone (state_test.go, suite_test.go checked) — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Round-0 primary+principles review read 13.7M cache tokens on a 5,020-line over-cap diff; inherent to the change — noted
- **[flow-cost]** The integrate inline verify read 6.6M cache tokens mostly polling a ~10 min backgrounded verify; addressed by the flow-speed fix — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Task 1's reproduce-before-touching step with measured figures took the reviewers query from 325 ms to 8 ms — noted
- **[flow-improvement]** Integrate's gates caught every fix-run omission before landing — noted
- **[flow-improvement]** The visual verifier cross-checked Total cost across three views — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The scoped re-capture's `-g` screen list was written by hand; check-visual-trigger already holds the changed UI paths it could derive from — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** guard_runs has no change or session column, so per-change attribution needs a LIKE on the worktree path — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** The 33-command inline verify ran serially past the 600 s tool timeout, twice; lint and test lists now run as two concurrent groups — fixed: 7b49ab4d

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case, not chosen for mining — noted
  - task: 13. In-run fix 1 — flow.verify: plan provenance, plan shape, and a records-address leak in the render-to-guard test
  - commit: fix(flow): pin the records address in the render-to-guard test
  - code: /Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-989-stats-ui-dashboard-review/record_test.go:1349
  - tests: `TestRecordRenderPanelWithNoFindingsReadsClearToTheRealGuard`, failing with FLOW_RECORDS_ADDR exported before the fix, passing after
  - why: an environment-precedence leak with a one-command reproducer and a mutation-provable one-line fix
- **[flow-eval]** review case, not chosen for mining — noted
  - severity: Important
  - location: stats/internal/store/health.go:113
  - reviewer: "Mean cost per finished change averages in every finished change with no stage run as $0."
  - bad: /Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-989-stats-ui-dashboard-review/health.go.bad
  - fixed: /Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-989-stats-ui-dashboard-review/health.go.fixed (e8606b58)
  - why: a subtle absent-read-as-zero bias in a SQL aggregate, invisible to tests seeding only costed changes
