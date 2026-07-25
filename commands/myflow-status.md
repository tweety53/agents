---
description: Status — show every open myflow change with its stage, gates, next command, and worktree
---

Use the **myflow-status** skill (`.claude/skills/myflow-status/SKILL.md`).

Follow that skill exactly. **Read-only** — reports stage and gate state for open changes; never commits, merges, or advances a stage.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Optional change name from `$ARGUMENTS`. With a name, show the detail view for that change; without one, show the table of all open changes.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
