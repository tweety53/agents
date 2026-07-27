---
model: sonnet
description: Info — explain the myflow pipeline stages, gates, commands, and flags
---

Use the **myflow-info** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. **Read-only reference** — reads the installed `myflow-manual-review.mdc` at invocation time rather than answering from memory.

**Input:** Optional stage or command name from `$ARGUMENTS` (e.g. `awaiting-manual-test`, `do-fix`). Without one, explain the whole pipeline.
