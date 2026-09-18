# kan-554-flow-fix-validate-reproducer-scripts-exit-code — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-554-flow-fix-validate-reproducer-scripts-exit-code-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the harness needed three rounds of its own fixes before green (an `:-` default swallowing the empty change-name argument, a `VAR=`-prefixed function argument read as a command name, sandbox copies missing the runner's sourced metachar file); all fixed in-run — declined
- **[flow-fix]** the first lint run caught the missing skills/flow/scripts/ symlink, a shipping convention the implementation missed; the guard enforced the convention exactly as designed — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the pass-1 bundle dispatch breached the wall-clock ceiling (~16m against 15m) because the blocking Agent call completed before a stop could land; the ceiling records overruns after the cost is paid, so the contract should state that or make it enforceable — filed: KAN-612

## What went well, and how to reproduce it — `flow-improvement`

_none — this angle produced no findings._

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all three dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
