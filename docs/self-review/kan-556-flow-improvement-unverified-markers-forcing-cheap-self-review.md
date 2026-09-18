# kan-556-flow-improvement-unverified-markers-forcing-cheap — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-556-flow-improvement-unverified-markers-forcing-cheap-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the first `--pre-fix-exit` value was the raw reproducer exit (1) rather than the dispatch-time verdict the script printed (0), the third recorded occurrence of this misfeed across runs (kan-542, kan-546, kan-556); the flag should validate its value against the printed verdict — filed: KAN-614

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the reviewer's first fix advice (trim, not raise) was infeasible because the budget headroom was nearly zero, a fact only the re-run reviewer proved; the budget guard's failure line should print current size, budget row and headroom so trim-vs-raise is decidable from the guard itself — filed: KAN-615

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all three dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
