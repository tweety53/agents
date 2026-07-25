---
description: Do manual review — confirm review of the implementation diff is in progress and advance to do-review-started
---

Use the **myflow-state-advance** skill (`.cursor/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** `do-review-started`
**ACCEPTED_STAGES:** `awaiting-do-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** review the staged diff, then `/myflow-do-done` (or `/myflow-do-fix` if changes needed).

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
