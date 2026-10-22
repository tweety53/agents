# Self-review report for kan-919-flow-improvement-port-the-dev-stack-and-project

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** Data race in the new `TestCheckDevStackFresh` case 3: the httptest handler read `served` while the test goroutine wrote it; the race detector flagged parallel tests too, and verify waived it as environmental — fixed: 74492d54
- **[flow-fix]** `run-verify-list.sh` deleted a failing command's full output and kept 20 lines, losing the race stack; the output now stays in the git dir and its path is printed — fixed: 5e9f6de7
- **[flow-fix]** Verify's one inline re-run waived a data race that passed on re-run; a `WARNING: DATA RACE` is now never environmental and goes to The loop — fixed: 47c4ae9a
- **[flow-fix]** The plan named `npx` and two unrelated Files paths; panel F5 and `check-task-commit-fields` caught both, each fixed by a one-line Correction — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The verify dispatch spent 11m13s and 2.4M cache tokens, mostly the race re-run; removed by the race fixes above — noted
- **[flow-cost]** Reviewer rows reading "output not measured" are the deliberate `output_unmeasured` state, not a defect — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Naming a precedent sha (c20a7ab2) with its shim and test shapes kept every implementer at 1–2.5 minutes with no rework — noted
- **[flow-improvement]** The test-sanity slot caught a vacuous harness case carried over from bash (F8) — noted
- **[flow-improvement]** The parity tests read the deleted bash libraries from history (`visualLibAt` at b1b0d601), keeping the parity record after deletion — noted
- **[flow-improvement]** `live-verification.md`'s before/after table at the merge base gave an explicit parity verdict — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `check-plan-shape` passed an unbackticked multi-path `**Files:**` that `check-task-commit-fields` read as one path; it is now a plan-shape violation — fixed: 02d5dcea
- **[flow-automation]** `run-reproducer` silently read a near-miss `# mutation-reproducer: …` marker as generic; it now refuses with exit 2 — fixed: 22ed70c5

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** About 5 minutes of verify wall-clock re-ran the race-failed suite; removed by the race fixes — noted
- **[flow-speed]** Panel round 0 took 7m32s with both slots in parallel, as designed — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case F8, vacuous pipe test carried over from bash — not chosen for mining — noted
  - severity: Important (test-sanity); location: `stats/internal/guard/check_dev_stack_fresh_test.go:84`; fix: ec64f5fa (squashed into 8f7c42da)
- **[flow-eval]** review case F1, unreadable Makefile reported clean — not chosen for mining — noted
  - severity: Important (primary); location: `stats/internal/guard/uitestoverrides.go:60`; fix: ec64f5fa (squashed into 8f7c42da)
- **[flow-eval]** impl case Task 3, port check-uitest-overrides to Go — not chosen for mining — noted
  - code: `stats/internal/guard/uitestoverrides.go:34`; commit 8f7c42da
