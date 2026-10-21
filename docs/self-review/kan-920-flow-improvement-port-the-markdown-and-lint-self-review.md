# Self-review report for kan-920-flow-improvement-port-the-markdown-and-lint

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The inline panel-fix round left no `panel-fix` ledger row; disproved by the first review — `review-panel.md:802-809` fixes a Minor-only round inline with no dispatch record and round 0 was four Minors, so the guard change was dropped and its back-filled row deleted — noted
- **[flow-fix]** Task 5's inline `dispatch begin` ran after its work (7 s row); an inline begin now refuses while tracked files other than the change's `tasks.md` are modified, and the task-5 row's start was repaired — fixed: 69c9c7cc
- **[flow-fix]** `decision.json` recorded `parent.effort` as `unknown`, which `flow record dispatch` refuses; it now records `default`, and migration 0041 rewrites the 165 stored `unknown` rows — fixed: e233883d
- **[flow-fix]** Subagent `output_tokens` are harvested from stream-start usage (`stop_reason: null`), undercounting output; that output is now marked not measured, stored rows marked by migration 0042, and any in-flight `cost_usd` cleared — fixed: 4352139d
- **[flow-fix]** `check-task-records` exited 1 because lint ran between task 4's commit and its `close-task` tick; the guard was right and implement.md already pairs the two in one call — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The parent ran each `## lint` guard bare by hand (4 false exit-2s) before switching to `run-verify-list.sh`, which the pipeline already prescribes — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Test expectations captured from the base bash, and a byte-for-byte live check against base `75261910` through the symlinked layout — noted
- **[flow-improvement]** F1's reproducer re-run green after its fix, and the green-sha skip of `npm test` worked — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The `## test` list ran its commands serially (~7 min, twice); `run-verify-list.sh` now runs them concurrently (307 s vs 385–418 s), the SPA build lock bounded at 600 s — fixed: 6c6ece7c
- **[flow-speed]** The read-only review panel waited for the full suite; pass 1 now reviews while the suite runs, a suite fix applied only after the round's wait — fixed: f849b430

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 1, port check-normative-inventory and share the owned-corpus walk; operator chose not to mine — noted
  - Task: `/Users/tweety53/Projects/agents-worktrees/kan-920-flow-improvement-port-the-markdown-and-lint/spectre/changes/archive/kan-920-flow-improvement-port-the-markdown-and-lint/tasks.md:96`
  - Code: `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-920-flow-improvement-port-the-markdown-and-lint/normativeinventory.go:111`
  - Why: an awk sentence splitter ported to Go with byte parity on the real 119-file corpus
- **[flow-eval]** impl case — Task 2, port check-mutation-reproducer-pin; operator chose not to mine — noted
  - Task: `/Users/tweety53/Projects/agents-worktrees/kan-920-flow-improvement-port-the-markdown-and-lint/spectre/changes/archive/kan-920-flow-improvement-port-the-markdown-and-lint/tasks.md:141`
  - Code: `/Users/tweety53/Agents/flow/state/agents-a740d89c/flow-eval/kan-920-flow-improvement-port-the-markdown-and-lint/mutationreproducerpin.go:187`
  - Why: the joined two-line scan reports only boundary-spanning matches, awk's RSTART reproduced; an off-by-one brings back KAN-624 F5's phantom rows
