---
description: Review done — confirm the PR was reviewed (and merged) and advance to review-done
---

Use the **myflow-state-advance** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

**TARGET_STAGE:** `review-done`
**ACCEPTED_STAGES:** `awaiting-pr-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** `/myflow-finish <name>` — verify the PR merged, sync specs, and archive.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
