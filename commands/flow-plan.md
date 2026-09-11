---
name: /flow-plan
id: flow-plan
category: flow
description: Research — a thinking partner for exploring ideas and clarifying requirements
---

Use the **flow-plan** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. A stance, not a workflow: investigate, ask, and visualize — never write application code, never advance pipeline state, commit only what its **Landing the note** section names. Investigation and questions run deeper than a single-pass answer, per its stopping rule; a captured note always includes the step-by-step breakdown. The thinking runs in the current session, on its own model — no subagent, per its **The session does the thinking itself** section.

**Input:** the argument after `/flow-plan` is whatever the user wants to think about — a Jira key, an idea, a problem, a change id, a comparison, or nothing. Pass it through as the topic. A key names the note's file; without one the skill creates a Jira Task and names it after that.

**When done:** the note, plan and decision are committed and pushed to the default branch. Ready to become a change? Run `/flow <key>`.
