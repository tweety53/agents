---
model: opus
description: Single-command pipeline — brainstorm, implement behind the review panel resolved from the settings store, and integrate, pausing only at the human gates
---

Use the **flow** skill — installed globally, so let your harness resolve it by name rather than
assuming a project-local path.

Follow that skill exactly. Accepts **no state** (creates a change), **`STARTED`** (resumes a
creating run that stopped before implementation), or **`IN_PROGRESS`**.

**Input:** the change name or a description/Jira key to seed a new change, from `$ARGUMENTS` or the
conversation — and nothing else, save `--base <branch>` on a creating run.
Report any argument that is not a change name, description, or fix instruction rather than ignoring
it.
