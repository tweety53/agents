# kan-541-flow-cost-run-the-systematic-per-frame-mockup — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-541-flow-cost-run-the-systematic-per-frame-mockup-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the ticket asked for the per-frame sweep before the first fix round, but `flow.visual-verify` had already landed it two days before the issue was filed, so a whole pipeline run (brainstorm, panel, verdict change) went on an ask main already delivered; kan-548's ticket shows the same stale-finding class; a run on a filed finding should verify the defect still exists on the resolved base before planning — filed: KAN-597
- **[flow-fix]** the store was unreachable for two stage marks (journalled with one warning line each) and three `flow record pass` notes were refused as re-stamps; the degraded-but-loud path worked as designed — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the run rejected inventing machinery for the one remaining hypothetical gap (a project declaring `## visual verification` but no `mockups` row despite mockup files) as speculative; the simplest-thing rule already governs, so no filing — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the slot's original reproducers demonstrated standing corpus facts that could never exit 0 and had to be re-authored defect-shaped; duplicate of KAN-554, the reproducer exit-code contract change — declined

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the store's change record carries no summary/verdict field, so the bundle's central content ("the change summary carries the verdict") had no locatable referent for anyone but the run's inline session; persist the verdict and a one-paragraph summary on the change record — filed: KAN-598
- **[flow-stats-app]** all four dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
