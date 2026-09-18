# kan-538-flow-fix-build-green-gate-reports-malformed — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-538-flow-fix-build-green-gate-reports-malformed-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `flow record dispatch` refused `-role panel` before `reviewer` fit, a trial-and-error cycle, because the role vocabulary lives in review-panel.md while the CLI enforces its own set; serve the accepted roles or name them in the refusal (the served-vocabulary pattern KAN-533 established for stage keys) — filed: KAN-595
- **[flow-fix]** the plan's first `**After:**` value did not gate (the field wants `Task <ids>` or `none`); the plan gate caught it before anything recorded wrong, so no change filed — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the dispatch row reads `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
