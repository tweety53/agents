---
model: sonnet
description: Status — show every open myflow change with its stage, gates, next command, and worktree
---

Use the **myflow-status** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. **Read-only** — reports stage and gate state for open changes; never commits, merges, or advances a stage.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Optional change name from `$ARGUMENTS`. With a name, show the detail view for that change; without one, show the table of all open changes.
