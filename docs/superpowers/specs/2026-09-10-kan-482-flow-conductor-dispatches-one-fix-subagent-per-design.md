# KAN-482 design — one fix dispatch per round, mechanically checked

Date: 2026-09-10
Change: `kan-482-flow-conductor-dispatches-one-fix-subagent-per`
Jira: KAN-482 — "Flow conductor dispatches one fix subagent per finding instead of one combined fixer"

## Problem

`skills/flow/review-panel.md` already stated the fix goes to **one** fix subagent as a combined
list. On the KAN-449 run the conductor nonetheless dispatched four background fix subagents, one
per reviewer. Instructions failed once, so the obligation also becomes a mechanical check — the
repo's own `gate-is-a-guard` decision.

## Design

Six surfaces, three commits:

1. **`flow record dispatches -change <name>`** (`stats/cmd/flow/record.go` + tests) — the findings
   verb's read contract over dispatch rows: JSON array, `[]` for a change with none, non-zero exit
   when the store cannot be reached (never a silent empty array). Amended during implementation:
   no store change was needed — `RunRecord` already loads every dispatch row, seq-ordered.
2. **`scripts/check-panel-fix-single-dispatch.sh <worktree> <change> <session-token>`** — counts
   this run's `panel-fix` dispatches per round key and checks key shape
   (`panel-fix-<round>` / `panel-fix-<round>-retry`); exit 0 clean, 1 violations named, 2 cannot
   answer. Bash 3.2-safe: no associative arrays.
3. **Wiring** — the guard runs at the panel's close beside `check-panel-findings-closed.sh`;
   exit 1 is a handback question, exit 2 stops the close. Registered in `skills/flow/SKILL.md`'s
   guard list.
4. **Prose** — the fix step's full contract (one dispatch per round on the combined list, awaited
   foreground, canonical keys, pre-dispatch self-check) in `review-panel.md`; the same constraint
   in `implement.md`'s conductor relay contract.

## Decisions

- **prose-plus-guard** — prose alone already failed once at line 905; the obligation becomes a
  check per `gate-is-a-guard`. (Operator chose this over prose-only and guard-only.)
- **guard-at-panel-close** — dispatch rows only exist post-dispatch; the close gate is the last
  stoppable point before handoff.
- **key-shape-check** — a count alone misses invented keys, the exact KAN-482 shape.
- **reads-may-fail** — an unreachable store exits non-zero; a fallback empty array would turn an
  outage into a green verdict.

See the change's own `design.md` for the full `## Decisions` records.
