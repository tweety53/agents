---
description: Do-fix manual review — confirm review of the fix is in progress and advance to fix-review-started
---

Use the **myflow-state-advance** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

**TARGET_STAGE:** `fix-review-started`
**ACCEPTED_STAGES:** `awaiting-fix-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** review the fix, then `/myflow-do-fix-done`.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
