---
description: Do done — mark implementation review complete and advance to do-done
---

Use the **myflow-state-advance** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

**TARGET_STAGE:** `do-done`
**ACCEPTED_STAGES:** `awaiting-do-review`, `do-review-started`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** `/myflow-manual-test <name>` — write the manual test guide (asks whether to skip Gate C).

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
