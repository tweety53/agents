---
description: Do manual review — confirm review of the implementation diff is in progress and advance to do-review-started
---

**Run the script first.** It handles the mechanical write; the skill is only for what needs
judgment.

```bash
~/.cursor/skills/myflow-state-advance/state-advance.sh \
  --name <name> --target do-review-started \
  --accepted awaiting-do-review --by /myflow-do-manual-review
```

- **Exit 0** — print its output and stop. The stage is written; nothing further is needed.
- **Exit 2** — a usage error in this command file, or `jq` is not installed. Report it; do not
  work around it.
- **Exit 3, 4, 5, or 6** — do **not** retry and do **not** hand-edit the state file. Load the
  **myflow-state-advance** skill and follow it from step 1; it owns name resolution, the
  stage-mismatch override prompt, and self-heal.
- **Any other non-zero exit** — 1, 126, 127, or anything else the list above does not name: the
  script is missing, not executable, or failed in a way it does not define (a partial install is
  the common cause). Treat it exactly as "the script is unavailable" — load the
  **myflow-state-advance** skill and follow it from step 1. Never retry the script, and never
  hand-edit the state file.
- **`<name>` omitted** — skip the script and use the skill, which resolves the name first.

Use the **myflow-state-advance** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

**TARGET_STAGE:** `do-review-started`
**ACCEPTED_STAGES:** `awaiting-do-review`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** review the staged diff, then `/myflow-do-done` (or `/myflow-do-fix` if changes needed).

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
