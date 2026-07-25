---
description: Start — Superpowers #1 brainstorming + #3 writing-plans woven into OpenSpec artifacts
model: opus
---

Use the **openspec-propose-superpowers** skill (`.claude/skills/openspec-propose-superpowers/SKILL.md`).

Follow that skill exactly. Runs Superpowers Basic Workflow **#1** (brainstorming) and **#3** (writing-plans) intertwined with OpenSpec artifact creation. Do not skip either.

**Input:** Change name (kebab-case) and/or description of what to build — from `$ARGUMENTS` or conversation. If both are omitted: run `openspec list --json`; if exactly one active (non-archived) change has incomplete planning artifacts, resume it automatically; if multiple, ask which; if none, ask what to build.

**Next:** `/myflow-do <name>` (#2–#6, no commits), or `/myflow-full <name>` for the gated end-to-end flow (manual review before code review/finish).
