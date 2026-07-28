---
description: Do-fix done — confirm the fix was reviewed and advance back to the stage the fix was raised at
---

**Run the script first.** It handles the mechanical write; the skill is only for what needs
judgment.

```bash
~/.cursor/skills/myflow-state-advance/state-advance.sh \
  --name <name> --target originStage \
  --accepted awaiting-fix-review,fix-review-started --by /myflow-do-fix-done
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

**TARGET_STAGE:** `originStage` — the dynamic form. The skill reads the state file's `originStage` and targets it, including the retarget and clearing rules; see **Target forms** in `myflow-state-advance/SKILL.md`.
**ACCEPTED_STAGES:** `awaiting-fix-review`, `fix-review-started`

Pure state write — no verification, no git operations. Follow that skill exactly.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted, resolve per the skill's step 1.

**When done:** resume at the origin stage — see the resolved `TARGET_STAGE` above and **Fix re-entry** in the rule file for what comes next.

**Model:** run on Sonnet (Cursor cannot enforce this per-command — switch manually if needed).
