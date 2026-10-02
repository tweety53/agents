# status-line-emojis

## Why

Operator, 2026-10-02: each status line should lead with its own emoji so the state reads at a glance.

## What changes

- `rules/be-brief.mdc` — a status line is `<emoji> <unit> (<a few words>) — <state>`, one distinct emoji per state; it also does not say why the unit is in that state.
- `rules/agent-baseline.md` — the be-brief one-liner carries the same shape.
