---
model: sonnet
description: Do-fix manual review — confirm review of the fix is in progress and advance to fix-review-started
---

Use the **myflow-state-advance** skill (`.claude/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** `fix-review-started`
**ACCEPTED_STAGES:** `awaiting-fix-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** review the fix, then `/myflow-do-fix-done`.
