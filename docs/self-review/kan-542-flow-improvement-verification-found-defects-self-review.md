# kan-542-flow-improvement-verification-found-defects — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-542-flow-improvement-verification-found-defects-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel F3 (deferred): the "(the KAN-29/KAN-30 precedent)" citation is not resolvable from this repository and the deferred status records KAN-442 and KAN-423 as cross-repo keys resolved the same way; cross-repo precedent citations should name their repository so deliberate cross-repo keys do not read as dangling — filed: KAN-599
- **[flow-fix]** the handshake line was absent from both round-1 primary replies and the run recorded the breach instead of re-dispatching; duplicate of KAN-582, the single-model-harness finding from the kan-533 pass — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

_none — this angle produced no findings._

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** panel F2 struck a plan task whose premise (the budget ratchet tripped) the reviewer disproved with the guard's own exit; a plan task asserting a guard's current behaviour should run that guard during brainstorm and cite its output — filed: KAN-600
- **[flow-improvement]** the unflippable F1 reproducer was honestly re-authored so it could actually flip; duplicate of KAN-554, the reproducer exit-code contract change — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the `--pre-fix-exit` flag's semantics (run-reproducer's dispatch-time verdict, not the underlying script's exit) cost a correction; the same contract work KAN-554 carries covers the class — declined

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the dispatch row reads `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
