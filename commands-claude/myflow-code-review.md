---
model: sonnet
description: Code review — verify Gate C, check test coverage, verify tests/linters, commit apply work, Basic Workflow #7 (finishing-a-development-branch)
---

Use the **openspec-code-review-superpowers** skill (`.claude/skills/openspec-code-review-superpowers/SKILL.md`).

Follow that skill exactly. Runs **after Gate B (code review) and Gate C (manual test)**, and before `/myflow-finish`: verify Gate C completion (or its `SKIPPED` marker) → check test coverage against delta specs (if thin, suggests `/myflow-do-fix <name>` to add tests) → run tests/linters → **commit** the apply work → Basic Workflow **#7** (`finishing-a-development-branch`) — merge / PR / push, always.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change has finished Gate B/C, use it automatically; if multiple, ask which.

**If `docs/manual-test/<name>.md` is missing:** prefer `/myflow-manual-test <name>` first unless the user explicitly skips Gate C (via `/myflow-manual-test-skip <name>`).

**When done:** `/myflow-finish <name>` — verify merged, sync specs, archive.
