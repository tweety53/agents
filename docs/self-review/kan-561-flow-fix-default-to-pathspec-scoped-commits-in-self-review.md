# kan-561-flow-fix-default-to-pathspec-scoped-commits-in — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-561-flow-fix-default-to-pathspec-scoped-commits-in-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the citation guard read the backticked placeholder list as a rootless path and the lint list caught it, the second recorded false positive of this class after kan-548's backticked `n/a`; a declared exemption for non-path tokens stops the per-token rewording commits — filed: KAN-619
- **[flow-fix]** the base-movement rebase conflict stopped the stage until the operator ruled; the pipeline stopped correctly and the operator's resolution carried both intents, no defect — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** both dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
