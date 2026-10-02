# status-line-task-total

## Why

Operator, 2026-10-02: a task status line should show progress through the plan, not a bare number.

## What changes

- `rules/be-brief.mdc` — a task unit is printed `Task <x>/<n>`, x the task number, n the plan's total task count; examples updated.
- `rules/agent-baseline.md` — the be-brief one-liner carries the same shape.
