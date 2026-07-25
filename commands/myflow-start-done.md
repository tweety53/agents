---
description: Start done — confirm the proposal was reviewed and advance to proposal-done
---

Use the **myflow-state-advance** skill (`.claude/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** `proposal-done`
**ACCEPTED_STAGES:** `awaiting-proposal-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** `/myflow-do <name>` — start implementation.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
