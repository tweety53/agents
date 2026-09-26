# kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made — self-review

**Deferred:** reasoning pass run on account:zai-individual-coding-plan/GLM-5.3-Flash from docs/self-review/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** a background SPA build silently ran from the main checkout because background tool calls do not inherit the session's `cd`, producing false vet failures until re-run with explicit `-C`; fail-visible and self-corrected, and the git-write half of this cwd class is now hook-guarded, so the residual stays below the bar — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._
