# Self-review report for kan-1025-flow-improvement-require-a-suite-regression-test

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `flow stage end` for `flow.review-panel` without `-C` ran its findings-closed gate against the cwd (the main checkout) and falsely refused (in-run pipeline fix) — fixed: 21127f9f
- **[flow-fix]** The mutation brief demanded a whole-package test run with no rule for a package whose run eats the slot ceiling (guard package 332 tests, 165 s idle / ~455 s contended; the slot ran 55 min), and stated the `# mutation-reproducer` exit convention only indirectly; the brief now allows a test-family prefix filter past a third of the ceiling, sizes the bound from a measured run and states the convention — fixed: 6a058a5f
- **[flow-fix]** `hooks/protect-main-checkout.py` scanned a non-`cat`/`tee` heredoc body as shell, denying `flow record summary -file - <<'EOF'` with prose containing `->`, `rm` or `>` from the main checkout; a body is now data only for a `flow` opener whose command runs nothing (fail-closed allowlist after review found the first denylist design bypassable) — fixed: bf971fa8
- **[flow-fix]** Task 2's planned prose tripped `check-installed-citations` and `check-verbatim-moves`; the guards caught it before commit, as designed — noted
- **[flow-fix]** Round-0 primary reproducers' premise lines were off by one with verbatim content; the exit-contract audit caught it and the parent corrected the line number, as the contract allows — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The timed-out failure-modes+mutation dispatch (55 min, ~480k tokens) is the cost of the mutation-brief and slot-ceiling findings, fixed under their angles — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Six break-and-prove flips held for the panel fix round, and the new guard closed this change's own panel from its own `regression-test:` lines — noted
- **[flow-improvement]** The in-run pipeline fix landed in the background without stopping the run, and round 1 closed clean on one sonnet/low re-run — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The slot-ceiling wake-up, now named as the bounded wait loop's elapsed check before every re-wait — fixed: 2163b94c

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The guard package's 165 s idle run was profiled: `TestChangePlan`'s 63 s is parallel-scheduling wait, its subtests sum under 2 s; no cheap speed-up exists — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1 (Important, primary, round 0), not mined (mining ask suspended by the operator) — noted
  - location: `stats/internal/guard/panelfindingsclosed.go:336`
  - reviewer: "git ls-files --error-unmatch reads the path as a pathspec, so a directory, `.` or a glob is accepted whenever anything under it is tracked."
  - bad/fixed code and fix diff extracted to the change's state dir `flow-eval/kan-1025-flow-improvement-require-a-suite-regression-test/` (pre-reshape d2265180 / cfb828f8)
  - why: a subtle tool-semantics membership bug, checkable mechanically with a directory or `.` input
- **[flow-eval]** impl case — Task 1 (regression-test close check), not mined (mining ask suspended by the operator) — noted
  - code: `stats/internal/guard/panelfindingsclosed.go`, test `TestCheckPanelFindingsClosedFixedNeedsARegressionTest`
  - why: a well-specified guard extension whose round 0 still found four Important gaps — a strong grading signal

## Instruction gaps that cost fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
