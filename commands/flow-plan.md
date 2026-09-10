---
name: /flow-plan
id: flow-plan
category: flow
description: Research — a thinking partner for exploring ideas and clarifying requirements
---

Use the **flow-plan** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. A stance, not a workflow: investigate, ask, and visualize — never write application code, never advance pipeline state, never commit. Investigation and questions run deeper than a single-pass answer, per that skill's stopping rule, and a captured note always includes the step-by-step breakdown by default. The thinking itself runs in the current session, on its own model — no subagent is dispatched, per that skill's **The session does the thinking itself** section.

**Input:** the argument after `/flow-plan` is whatever the user wants to think about — a vague idea, a specific problem, a change id (to explore in context of that change), a comparison, or nothing at all. Pass it through as the topic.

**When done:** there's no required next step. If the thinking is ready to become a change, run `/flow`.
