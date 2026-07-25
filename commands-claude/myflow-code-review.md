---
model: sonnet
description: Code review — verify Gate C, check coverage, verify tests/linters, commit, push, open PR (never merges)
---

Use the **openspec-code-review-superpowers** skill (`.claude/skills/openspec-code-review-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage **`awaiting-test`** — on mismatch, stop with the standard mismatch handoff. Runs **after Gate B (code review) and Gate C (manual test)**, before Gate D (human PR review): verify Gate C completion (reads `gates.tested`, promotes `false`→`true` when every checkbox is ticked, never overwrites `"skipped"`) → check test coverage against delta specs (if thin, suggests `/myflow-do-fix <name>` to add tests) → run tests/linters → **commit** the apply work → push and open a PR (Basic Workflow **#7**, constrained to the PR path — **never merges**) → write `stage: awaiting-pr-review`. Ends at Gate D with a PR URL; the human reviews and merges the PR themselves.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is at stage `awaiting-test`, use it automatically; if multiple, ask which.

**If `docs/manual-test/<name>.md` is missing:** prefer `/myflow-manual-test <name>` first — the stage gate makes this unlikely since `awaiting-test` is only reached once that guide exists.

**When done:** the human reviews and merges the PR at Gate D, then runs `/myflow-finish <name>` — verify merged, sync specs, archive.
