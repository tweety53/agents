---
description: Manual test done — confirm manual testing is complete and advance to manual-test-done
---

Use the **myflow-state-advance** skill (`.cursor/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** `manual-test-done`
**ACCEPTED_STAGES:** `awaiting-manual-test`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** `/myflow-review <name>` — commit, push, and open the PR.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
