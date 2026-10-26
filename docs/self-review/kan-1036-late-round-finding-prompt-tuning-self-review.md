# Self-review report for kan-1036-late-round-finding-prompt-tuning

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

**Rating:** skipped — paused by the operator

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The new late-round-miss rule read "round 2 or later" while panel rounds count from 0, excluding the first re-run; now "round 1 or later" and "a first-round reviewer instruction" — fixed: b51efadd
- **[flow-fix]** The slot-facing REPRODUCER paragraph named reproducers "by round, finding id and sequence number", contradicting the canonical round-id-index shape; two slots' scripts collided on a case-insensitive filesystem and one was overwritten; the paragraph now states the shape and check-panel-reproducers refuses an open finding's path off it — fixed: 9c093b46
- **[flow-fix]** design.md's Verification wording led to a bulk run of every check script instead of the lint list; one-off planner phrasing, the templates carry no such wording — noted
- **[flow-fix]** The ledger names task and fix commits the reshape rewrote; refs/flow/pre-reshape keeps them reachable, so no defect — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** An inline verifier window during which the parent wrote no transcript record read "Tokens: not measured"; the harvest now records a zero when one batch proves nothing fell inside the window — fixed: 857996db
- **[flow-cost]** The docs-only reduction kept the panel to primary plus the two always-on extras with no re-run round — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Inline micro execution and closing the Minor-only round inline landed the change with one panel round, and integrate chained with no operator stop — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** The rendered panel findings table carried no round, so a late-round-miss trace could not tell which round raised a finding; it now renders a Round column from the store — fixed: cf8608f3

## What can be sped up — `flow-speed`

- **[flow-speed]** Verify took about six minutes of an eighteen-minute run on a prose diff; the bundle does not show which steps dominated, and the Go suites are the list's own — noted

## What would make a good flow-eval case — `flow-eval`

_none — this angle produced no findings._

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

_none — this angle produced no findings._
