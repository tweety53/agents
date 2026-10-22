# Self-review report for kan-921-flow-improvement-port-the-guard-test-runner-run

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 5 — excellent

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** Plan step 2.3 wrongly said the two fixture-copy lists copy `coverage.sh` from the tree (both copy it from `d71a2327`); a measured test failure caught it and dated Corrections recorded it — noted
- **[flow-fix]** The durations case's `(0s)` → `\d+s` assertion matches decision `elapsed-unix-seconds` (whole Unix seconds); not a widened tolerance — noted
- **[flow-fix]** Reviewer ledger rows read `output not measured` per the harvester's documented contract — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Inline task 1 (3.26M cache read) and the opus panel slot (2.1M) are proportionate to a 3394-line diff — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Mutation proofs found a surviving mutant in case 7a, which gained a pre-existing untracked file — noted
- **[flow-improvement]** Review moved signal registration above `MkdirTemp`, closing a temp-dir leak window — noted
- **[flow-improvement]** The live check compared bash vs Go on the real tree: 19→16 harnesses, 202s→155s, exit 0 both — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** Verify's 11m18s was the live check's serial before/after suite runs; running them concurrently would not cut wall time with `JOBS`=`NumCPU` — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case: task 1, Port the guard test runner and its parallel library to Go (the operator chose not to mine it) — noted
  - Task: tasks.md task 1
  - Code: `stats/internal/guard/runguardtests.go:317` (`rgtRun`), commit d37153ac
  - Why: byte-exact bash oracle; traps in pipe vs file capture, the signal/tempdir race, Unix-second elapsed and list-order replay
