# kan-546-flow-fix-surface-foreign-staged-work-in-main — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-546-flow-fix-surface-foreign-staged-work-in-main-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the first three files were written into the main checkout before being caught and moved into the worktree; `check-worktree-location.sh` already guards this in lint and the run corrected inside the round, so no new guard filed — declined
- **[flow-fix]** the reproducer re-runs initially reported refusals because the parent passed the raw pre-fix exit rather than the verdict run-reproducer.sh printed; the same exit-code-contract class KAN-554 carries — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** panel F3 (important): the surfacing step's exit-2 cannot-answer outcome was unhandled in both wiring files, so an unreadable main checkout would have read as clean; every guard wiring should name the cannot-answer exit alongside the violation exit, and the wiring corpus is worth an audit for implicit exit-2 paths — filed: KAN-601

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all four dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
