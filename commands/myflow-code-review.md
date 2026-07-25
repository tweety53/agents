---
name: /myflow-code-review
id: myflow-code-review
category: myflow
description: Code review — verify Gate C, check test coverage, verify tests/linters, commit apply work, Basic Workflow #7 (finishing-a-development-branch)
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-code-review-superpowers** skill (`.cursor/skills/openspec-code-review-superpowers/SKILL.md`).

Follow that skill exactly. Runs **after Gate B (code review) and Gate C (manual test)**, and before `/myflow-finish`: verify Gate C completion (or its `SKIPPED` marker) → check test coverage against delta specs (if thin, suggests `/myflow-do-fix <name>` to add tests) → run tests/linters → **commit** the apply work → Basic Workflow **#7** (`finishing-a-development-branch`) — merge / PR / push, always.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change has finished Gate B/C, use it automatically; if multiple, ask which.

**If `docs/manual-test/<name>.md` is missing:** prefer `/myflow-manual-test <name>` first unless the user explicitly skips Gate C (via `/myflow-manual-test-skip <name>`).

**When done:** `/myflow-finish <name>` — verify merged, sync specs, archive.
