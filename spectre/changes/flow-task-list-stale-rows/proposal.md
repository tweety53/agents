# flow-task-list-stale-rows

## Why

Operator, 2026-10-10, during gymie KAN-924: "flow-task-list seems broken … I don't understand what is the current task". With one subagent running, the band showed four rows of finished work: closing status lines carrying text after the state word (`— blocked. …`) never matched, the same unit written two ways never closed its earlier row, a `blocked` row stayed for the session, a review dispatch named "Review in-run fix 6" showed as `fix`, and `Task 26/26 review` lost its prefix once the plan grew past 26 tasks.

## What changes

- `mods/flow-task-list/hooks/register.tsx`: a status line may carry text after its state word; a unit's row is keyed by its name before any `(`, `:` or ` —`, a trailing ` review` dropped; a `blocked` row is kept only until the next main turn starts; a running row's kind is the one whose word appears earliest in its description (`panel-fix` stays a fix); a numbered unit not from the running plan keeps its full text. Each with a test that fails without it.
- `skills/flow-contracts/pipeline.md`: the band's status-line rows state the key and the `blocked` rule.
