---
model: sonnet
description: Full cycle — start → do → manual review → manual test → code review → finish
---

Use the **openspec-full-cycle-superpowers** skill (`.claude/skills/openspec-full-cycle-superpowers/SKILL.md`).

Follow that skill exactly. Pipeline: **start → do (#2–#6, no commits) → Gate B manual review (optional `/myflow-do-fix` loop) → Gate C manual test (optional `/myflow-do-fix` loop) → code review (coverage check, commit + #7) → finish (verify merged, sync specs, archive)**.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name + what to build from `$ARGUMENTS` or conversation. If the name is omitted and a description implies a new change, propose with that description; if both are omitted: run `openspec list --json`, use the sole active change automatically, or ask if there are multiple.

**Flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test`, `no-archive`, `commit-during-apply` (legacy) — honor if present in the user message.

**Model note:** this command runs on Sonnet throughout, including Phase A brainstorming — the `model:` frontmatter only applies to the outer command, not per-phase sub-skill calls. For brainstorming-heavy new work, prefer running `/myflow-start` standalone (Opus) first, then `/myflow-full <name> skip-propose` for the rest.

**Stages (individual):** `/myflow-start` (#1+#3), `/myflow-do` (#2–#6), manual review, `/myflow-do-fix` (Gate B/C fixes), `/myflow-manual-test` / `/myflow-manual-test-skip` (Gate C), `/myflow-code-review` (coverage+commit+#7), `/myflow-finish` (verify merged+sync+archive)
