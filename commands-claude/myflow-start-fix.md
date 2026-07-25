---
model: sonnet
description: Start fix — revise the proposal after review, republish the artifact, stay at the proposal gate
---

Use the **openspec-propose-fix-superpowers** skill (`.claude/skills/openspec-propose-fix-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage `awaiting-proposal-review`. Revises proposal/design/specs/tasks per your feedback, republishes the artifact to the **same URL**, and stays at `awaiting-proposal-review`. Never writes code.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation, plus what you want changed.

**When done:** `/myflow-start-fix <name>` again for another round, or `/myflow-start-done <name>` to accept the plan.
