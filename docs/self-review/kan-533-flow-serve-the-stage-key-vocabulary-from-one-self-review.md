# kan-533-flow-serve-the-stage-key-vocabulary-from-one — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-533-flow-serve-the-stage-key-vocabulary-from-one-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `flow stage keys` silently ignores unexpected arguments (exit 0) while sibling `state list` rejects them (exit 2); the panel deferred it as its own change with its own test — filed: KAN-581
- **[flow-fix]** the bundled panel reply carried no MODEL HANDSHAKE line and the run recorded the breach instead of re-dispatching, because zcode's Agent tool carries no model parameter so a re-dispatch cannot change the variable the handshake guards; the breach line is structural on this harness, not a reviewer lapse, and the contract should treat a recorded harness mapping as satisfying it — filed: KAN-582
- **[flow-fix]** the flow-fast bundle assembled with 4 of 6 sources absent (tasks.md, design.md, narrative.md, git log --stat), so the deferred pass ran on half a bundle; the assembler should derive git log from the repo it resolves and the flow-fast contract should state which sources it can never have — filed: KAN-583

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the one dispatch row reads `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the plan's README-rename harness case was replaced with two sandbox cases because the rename would have raced concurrently-run guard harnesses on the real README; reproduce by giving any harness case that mutates a real shared file a sandbox fixture — filed: KAN-584
- **[flow-improvement]** the stage-mark-calls guard consumes its stage-key vocabulary from `go run ./cmd/flow stage keys` in the checkout, never the installed binary, so the checked vocabulary can never lag the served source; reproduce by pointing every vocabulary guard at its serving source in-tree — filed: KAN-585

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the dispatch row's unmeasured tokens; duplicate of KAN-525 from the kan-512 pass — declined
