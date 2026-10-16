# pipeline-fix-kan-957-body-weight-and-calorie-reads-reselect-rows-15

## Why

**Explain before asking** (formerly **The explanation is the turn's final message**,
`skills/flow-contracts/operator-prompts.md`) made every operator ask end the turn after its
explanation and fire `AskUserQuestion` only in the next turn. The operator meant that split for
flow-eval candidates alone; applied to the convergence confirm and the plan gate it dropped
`AskUserQuestion` from ordinary flow questions (gymie kan-957, 2026-10-16).

## What changes

- The turn split applies only to flow-eval candidates; every other ask explains and asks in one turn.
- Self-review step 4 splits only when the call offers an angle-7 candidate.
- The unfinished-work gate's filing ask explains and asks in the same turn.
- The section is renamed **Explain before asking**; its three citers follow.
