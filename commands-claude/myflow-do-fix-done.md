---
model: sonnet
description: Do-fix done — confirm the fix was reviewed and advance back to the stage the fix was raised at
---

Use the **myflow-state-advance** skill (`.claude/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** the value of `originStage` in the state file (see **Fix re-entry** in the rule file). If `originStage` is `do-review-started`, target `awaiting-do-review` instead — the diff changed, so review restarts. Clear `originStage` to `null` after writing.
**ACCEPTED_STAGES:** `awaiting-fix-review`, `fix-review-started`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** resume at the origin stage — see the resolved `TARGET_STAGE` above and **Fix re-entry** in the rule file for what comes next.
