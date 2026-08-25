---
model: sonnet
description: View and change the harness-wide flow defaults — default model and reviewer slots
---

Use the **flow-settings** skill — installed globally, so let your harness resolve it by name
rather than assuming a project-local path.

Follow that skill exactly. **Standalone, not a pipeline stage** — it takes no change name, reads
and writes no per-change state file, and marks no `myflow stage` call. It reads and writes the
harness-wide settings record (`myflow settings get`/`set`): the default model and the reviewer
slots `/flow` runs default to.

**Input:** none — this command takes no arguments and no flags. Any argument given is reported
rather than ignored.

**When done:** nothing further to run — the settings are in effect for the next `/flow` run.
