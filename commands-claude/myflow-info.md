---
model: sonnet
description: Info — explain the myflow pipeline stages, gates, commands, and flags
---

Use the **myflow-info** skill (`.claude/skills/myflow-info/SKILL.md`).

Follow that skill exactly. **Read-only reference** — reads `.cursor/rules/myflow-manual-review.mdc` at invocation time rather than answering from memory.

**Input:** Optional stage or command name from `$ARGUMENTS` (e.g. `awaiting-manual-test`, `do-fix`). Without one, explain the whole pipeline.
