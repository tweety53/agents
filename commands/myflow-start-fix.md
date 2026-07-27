---
name: /myflow-start-fix
id: myflow-start-fix
category: myflow
description: Start fix — revise the proposal after review, republish the artifact, stay at the proposal gate
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-propose-fix-superpowers** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Requires stage `awaiting-proposal-review`. Revises proposal/design/specs/tasks per your feedback, republishes the artifact to the **same URL**, and stays at `awaiting-proposal-review`. Never writes code.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation, plus what you want changed.

**When done:** `/myflow-start-fix <name>` again for another round, or `/myflow-start-done <name>` to accept the plan.
