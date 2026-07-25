---
model: sonnet
description: Fix a Gate B (manual review) or Gate C (manual test) finding — documents it in the proposal first, then Superpowers #4–#6 (no commits)
---

Use the **openspec-apply-fix-superpowers** skill (`.claude/skills/openspec-apply-fix-superpowers/SKILL.md`).

Follow that skill exactly. For fixes to an **already-applied** change found during manual review (Gate B) or manual test (Gate C). Resumes the **existing** apply worktree/branch — never creates a new one. Documents the fix in the change's proposal first (append a "Manual Review Fixes"/"Manual Test Fixes" section, or a linked nested sub-change — asks which), then runs Superpowers Basic Workflow **#4–#6** (SDD + TDD + full strict review panel re-run). **No git commits, push, merge, or PR** — commits happen later, in `/myflow-code-review`.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change already has an apply worktree, use it automatically; if multiple, ask which.

**When done:** Manual review (Gate B) again on the updated diff, or `/myflow-manual-test <name>` first if this was a Gate C fix — loop through `/myflow-do-fix` again as many times as needed. Once both gates are satisfied: `/myflow-code-review <name>`, then `/myflow-finish <name>` (also archives any nested `<name>-fix-N` sub-changes together).
