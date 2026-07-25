---
name: /myflow-full
id: myflow-full
category: myflow
description: Full cycle — start → do → manual review → manual test → code review → finish
---

**Model:** Sonnet (or your default) is fine for most of this pipeline. If Phase A ends up running brainstorming (a fresh proposal, not `skip-propose`), switch to Opus for that phase manually — Cursor can't auto-switch mid-command. For brainstorming-heavy work, prefer running `/myflow-start` standalone on Opus first, then `/myflow-full <name> skip-propose` on Sonnet for the rest.

Use the **openspec-full-cycle-superpowers** skill (`.cursor/skills/openspec-full-cycle-superpowers/SKILL.md`).

Follow that skill exactly. Pipeline: **start → do (#2–#6, no commits) → Gate B manual review (optional `/myflow-do-fix` loop) → Gate C manual test (optional `/myflow-do-fix` loop) → code review (coverage check, commit + #7) → finish (verify merged, sync specs, archive)**.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name + what to build from `$ARGUMENTS` or conversation. If the name is omitted and a description implies a new change, propose with that description; if both are omitted: run `openspec list --json`, use the sole active change automatically, or ask if there are multiple.

**Flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test`, `no-archive`, `commit-during-apply` (legacy) — honor if present in the user message.

**Stages (individual):** `/myflow-start` (#1+#3), `/myflow-do` (#2–#6), manual review, `/myflow-do-fix` (Gate B/C fixes), `/myflow-manual-test` / `/myflow-manual-test-skip` (Gate C), `/myflow-code-review` (coverage+commit+#7), `/myflow-finish` (verify merged+sync+archive)
