---
model: sonnet
description: Fix a Gate B (manual review), Gate C (manual test), or Gate D (PR review) finding — documents it in the proposal first, then Superpowers #4–#6. Stage-only (no commit) at Gate B/C; commit + push to the PR at Gate D only.
---

Use the **openspec-apply-fix-superpowers** skill (`.claude/skills/openspec-apply-fix-superpowers/SKILL.md`).

Follow that skill exactly. For fixes to an **already-applied** change found during manual review (Gate B), manual test (Gate C), or PR review (Gate D). Resumes the **existing** apply worktree/branch — never creates a new one. Documents the fix in the change's proposal first (append a "Manual Review Fixes"/"Manual Test Fixes"/"PR Review Fixes" section, or a linked nested sub-change — asks which), then runs Superpowers Basic Workflow **#4–#6** (SDD + TDD + full strict review panel re-run).

Accepts **six** incoming stages. The incoming stage is recorded as `originStage` and — never asked, always derived — selects the git behavior at the end:

- `awaiting-do-review`, `do-review-started`, or `do-done` (Gate B) → **stage only**: `git add -A`, **no commit**, no push, no PR.
- `awaiting-manual-test` or `manual-test-done` (Gate C) → **stage only**: `git add -A`, **no commit**, no push, no PR.
- `awaiting-pr-review` (Gate D) → **PR-fix**: `git add -A`, **commit, and push to the existing PR branch** — the only place in the myflow pipeline where `/myflow-do-fix` commits, and only after the full strict review panel passes. Never merge, force-push, or amend.

`do-done` and `manual-test-done` are accepted because both mean "the work is complete and confirmed, but something was found before the next command ran". Accepting `manual-test-done` is what lets `/myflow-review`'s coverage check recommend `/myflow-do-fix` without a stage-mismatch override.

**This command always ends at `stage: awaiting-fix-review`**, in every mode — it never returns the change to the stage it started from. Returning to `originStage` is `/myflow-do-fix-done`'s job (`do-review-started` returns to `awaiting-do-review`; every other origin returns to itself). Do not start another fix round from `awaiting-fix-review` or `fix-review-started` — those are not accepted origins; run `/myflow-do-fix-done <name>` first.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change already has an apply worktree, use it automatically; if multiple, ask which.

**When done:** the change sits at `awaiting-fix-review`. Review the fix, then `/myflow-do-fix-manual-review <name>` (optional, marks review started) and `/myflow-do-fix-done <name>` — that returns it to `originStage`. Only then may another `/myflow-do-fix <name>` round start. If this was a Gate C fix, refresh the guide with `/myflow-manual-test <name>` after returning to `awaiting-manual-test`. Once both gates are satisfied: `/myflow-manual-test-done <name>` → `/myflow-review <name>` → the human merges the PR (Gate D) → `/myflow-review-done <name>` → `/myflow-finish <name>` (also archives any nested `<name>-fix-N` sub-changes together). If this was a Gate D fix, ask the user to re-review the pushed commit on the PR and merge it (never merged by this command).
