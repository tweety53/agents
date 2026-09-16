---
name: /flow-plan
id: flow-plan
category: flow
description: Research — a thinking partner for exploring ideas and clarifying requirements
---

Use the **flow-plan** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. A stance, not a workflow: investigate, ask, and visualize — never write application code, never advance past `STARTED`, commit only what its **Capturing a new change** section names. Investigation and questions run deeper than a single-pass answer, per its stopping rule; a captured design always includes the step-by-step breakdown. The thinking runs in the current session, on its own model — no subagent, per its **The session does the thinking itself** section.

**Input:** the argument after `/flow-plan` is whatever the user wants to think about — a Jira key, an idea, a problem, a change id, a comparison, or nothing. Pass it through as the topic. A key names the change; without one the skill creates a Jira Task first.

**When done:** a captured session has created the change at `STARTED`, planned and pushed on `spectre/<name>`. Ready to implement? Run `/flow <name>`.
