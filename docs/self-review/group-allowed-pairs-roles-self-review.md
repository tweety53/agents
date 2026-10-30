# Self-review report for group-allowed-pairs-roles

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The mutation slot's own mutant runs carried no scope or time budget — the whole guard package per mutant breached the 15-minute ceiling; the rule now sizes them by planned mutant count and filters by test-family prefix past half the ceiling — fixed: 69d379df
- **[flow-fix]** The plan's live check expected a before-window figure from a just-opened experiment, whose before window is empty by design (`experimentstats.go`); worked around on the store copy — noted
- **[flow-fix]** A test red on the base had already been fixed identically on main; the rebase kept main's side — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The planner's dispatch was the run's largest cost (15m32s, 6.96M cache read) with no defect behind it — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The markdown table stays the one source of the roles' groups and descriptions, parsed and served, never stored — noted
- **[flow-improvement]** A pre-committed `verbatim-moves.txt` let the prose tasks pass the lint guard on the first try — noted
- **[flow-improvement]** The verifier-on-rebased-HEAD gate found a real defect, fixed in-run with a test that fails without it — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** The full-app-suite zip rebuild was a manual step after `capture` (task 6's first commit missed it); the project's `capture` row now chains it with absolute arguments — fixed: fbeb034b

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The live check moved an experiment's start with raw SQL on the store copy; a one-off, no recurring need — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** The mutation-slot breach (~16 min) and the mismatched-width prototype render (~3.5 min) are fixed under `flow-fix` and `flow-prompt-tuning` — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 3, the experiment-stats role selectors at `stats/internal/store/experimentstats.go:99`, tested by `TestExperimentStatsExtraBundleRoles`; mining not asked this run — noted
  - Why: regex selectors that split one ledger role into three by key prefix, with a test that tells old from new selectors apart

## Instruction gaps that cost fix rounds, and the generic prompt update that would have avoided them — `flow-prompt-tuning`

- **[flow-prompt-tuning]** Rendering an approved prototype as a region capture's frame named no region width, did not state `frame none`'s `diff=n/a`, and contradicted the no-clipped-region rule; the verifier and planner text now state all three — fixed: ded38c96
- **[flow-prompt-tuning]** The first verifier left the change's other moved captures unnamed and unswept, missing a clipped chart label; `check-verify-report` now refuses an unnamed capture or an unswept frameless view — fixed: 14f7e300
