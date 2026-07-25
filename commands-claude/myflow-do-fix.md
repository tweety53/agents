---
model: sonnet
description: Fix a Gate B (manual review), Gate C (manual test), or Gate D (PR review) finding — documents it in the proposal first, then Superpowers #4–#6. Stage-only (no commit) at Gate B/C; commit + push to the PR at Gate D only.
---

Use the **openspec-apply-fix-superpowers** skill (`.claude/skills/openspec-apply-fix-superpowers/SKILL.md`).

Follow that skill exactly. For fixes to an **already-applied** change found during manual review (Gate B), manual test (Gate C), or PR review (Gate D). Resumes the **existing** apply worktree/branch — never creates a new one. Documents the fix in the change's proposal first (append a "Manual Review Fixes"/"Manual Test Fixes"/"PR Review Fixes" section, or a linked nested sub-change — asks which), then runs Superpowers Basic Workflow **#4–#6** (SDD + TDD + full strict review panel re-run).

Accepts three incoming stages, and the stage — never asked, always derived — selects the git behavior at the end:

- `awaiting-review` (Gate B) or `awaiting-test` (Gate C) → **stage only**: `git add -A`, **no commit**, no push, no PR.
- `awaiting-pr-review` (Gate D) → **PR-fix**: `git add -A`, **commit, and push to the existing PR branch** — the only place in the myflow pipeline where `/myflow-do-fix` commits, and only after the full strict review panel passes. Never merge, force-push, or amend.

In every case the change returns to the stage it started from — this command never advances or rewinds a stage.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change already has an apply worktree, use it automatically; if multiple, ask which.

**When done:** Manual review (Gate B) again on the updated diff, or `/myflow-manual-test <name>` first if this was a Gate C fix — loop through `/myflow-do-fix` again as many times as needed. Once both gates are satisfied: `/myflow-code-review <name>`, then `/myflow-finish <name>` (also archives any nested `<name>-fix-N` sub-changes together). If this was a Gate D fix, ask the user to re-review the pushed commit on the PR and merge it (never merged by this command), then `/myflow-finish <name>` once merged.
