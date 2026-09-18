# kan-560-flow-fix-the-store-should-reject-placeholder — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-560-flow-fix-the-store-should-reject-placeholder-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel F3 (deferred): the CLI mirrors none of the agent-id shape rule before contacting the store, so an offline placeholder exits 0 and is retired at reconcile; the deferral's recorded reason stands (the store is the rule's single validator and reconcile retires the row before it lands) — declined
- **[flow-fix]** F1's reproducer hardcoded the stale plan names and could never flip until the parent repaired it; duplicate of KAN-606, the reproducer-instrument class from the kan-552 pass — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the session opened flow.verify while flow.review-panel was still open and the store recorded the superseded attempt-1 rows that read as gaps; flowd could warn when a stage opens while another stage's run is still open, catching the slip at write time — filed: KAN-618
- **[flow-stats-app]** all three dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
