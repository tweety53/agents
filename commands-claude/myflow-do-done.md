---
model: sonnet
description: Do done — mark implementation review complete and advance to do-done
---

Use the **myflow-state-advance** skill (`.claude/skills/myflow-state-advance/SKILL.md`).

**TARGET_STAGE:** `do-done`
**ACCEPTED_STAGES:** `awaiting-do-review`, `do-review-started`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** `/myflow-manual-test <name>` — write the manual test guide (asks whether to skip Gate C).
