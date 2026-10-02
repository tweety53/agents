# status-line-format

## Why

Operator, 2026-10-02: progress lines carried how a unit got done ("closed, ticked and pushed") and what was still running in prose; they want `Task 22 (very short description) - done, Task 21 (very short desc) - in progress`.

## What changes

- `rules/be-brief.mdc` — a status line is `<unit> (<a few words>) — <state>`, one per unit, nothing else.
- `rules/agent-baseline.md` — the be-brief one-liner carries the same shape.
