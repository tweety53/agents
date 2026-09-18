# kan-567-flow-fix-research-note-seeding-should-emit-guard — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-567-flow-fix-research-note-seeding-should-emit-guard-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel F2 (deferred): the kan-485 and kan-468 precedent citations are unverifiable in-tree; the deferral's recorded reason stands (cross-project citations by bare key are the corpus's existing convention) and the naming fix is KAN-599 from the kan-542 pass — declined
- **[flow-fix]** panel F1's root cause (a backticked path-shaped example tripping the citation guard) is the false-positive class of KAN-619 from the kan-561 pass — declined
- **[flow-fix]** the fix round's first reproducer re-run answered 2 because `--pre-fix-exit` got the raw exit instead of the dispatch-time verdict, the fourth recorded occurrence of the misfeed filed as KAN-614 from the kan-556 pass — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all three dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
