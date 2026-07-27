---
model: sonnet
description: Do-fix done — confirm the fix was reviewed and advance back to the stage the fix was raised at
---

Use the **myflow-state-advance** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

**TARGET_STAGE:** `originStage` — the dynamic form. The skill reads the state file's `originStage` and targets it, including the retarget and clearing rules; see **Target forms** in `myflow-state-advance/SKILL.md`.
**ACCEPTED_STAGES:** `awaiting-fix-review`, `fix-review-started`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** resume at the origin stage — see the resolved `TARGET_STAGE` above and **Fix re-entry** in the rule file for what comes next.
