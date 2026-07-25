---
model: sonnet
description: Manual test SKIP — generate run guide + checklist MD but mark Gate C explicitly bypassed (no boxes checked)
---

Use the **openspec-manual-test-superpowers** skill (`.claude/skills/openspec-manual-test-superpowers/SKILL.md`) in **skip mode**.

Follow that skill exactly, in skip mode. Generate or refresh `docs/manual-test/<change-name>.md` exactly as `/myflow-manual-test` would (run instructions + functionality checklist), but:

1. Add a `**Manual test status:** SKIPPED — YYYY-MM-DD (Gate C intentionally bypassed)` line to the header.
2. Leave **every** functionality-checklist and sign-off checkbox unchecked — nothing was actually run or verified.
3. Preserve any boxes already checked from a prior normal `/myflow-manual-test` run (skip mode never un-checks evidence, it only adds the bypass marker to a fresh guide).

**Create/update the file, stage it (`git add`, no commit), reply with a link to the file only** (do not paste the guide body), then **stop**.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is apply-ready for this stage, use it automatically; if multiple, ask which.

**When done:** `/myflow-code-review <name>` — it will detect the `SKIPPED` status and surface it before proceeding rather than treating the change as untested-but-silent.
