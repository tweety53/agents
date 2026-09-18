# kan-547-flow-fix-known-failures-baseline-so-flow-verify — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-547-flow-fix-known-failures-baseline-so-flow-verify-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the round-1 reviewer observed that `check-task-commit-fields.sh` does not resolve a flow-fast layout, so the guard policing task commits in every `/flow` run cannot see a flow-fast change's commits; teach it the layout or declare its spectre-only scope where flow-fast is defined — filed: KAN-602
- **[flow-fix]** the frozen plan artifact kept quoting superseded wording in its Task 1 step after Task 2's step was corrected; the plan is an untracked planning artifact, corrected in place, no contract change filed — declined
- **[flow-fix]** decision.json's roster predates the docs-only reduction; the decision record doing its job while the pass log carries the reduction — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** both dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
