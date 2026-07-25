---
name: /myflow-review
id: myflow-review
category: myflow
description: Review — verify Gate C, check coverage, verify tests/linters, commit, push, open PR (never merges)
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-review-superpowers** skill (`.cursor/skills/openspec-review-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage **`manual-test-done`** — on mismatch, stop with the standard mismatch handoff. Runs **after Gate B (manual review) and Gate C (manual test)**, before Gate D (human PR review): verify Gate C completion (reads `gates.tested`, promotes `false`→`true` when every checkbox is ticked, never overwrites `"skipped"`) → check test coverage against delta specs (if thin, suggests `/myflow-do-fix <name>` to add tests) → run tests/linters → **commit** the apply work → push and open a PR (Basic Workflow **#7**, constrained to the PR path — **never merges**) → write `stage: awaiting-pr-review`. Ends at Gate D with a PR URL; the human reviews and merges the PR themselves.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is at stage `manual-test-done`, use it automatically; if multiple, ask which.

**If the change is still at `awaiting-manual-test`:** testing has not been confirmed complete — run `/myflow-manual-test-done <name>` first; that is what writes `manual-test-done`.

**If `docs/manual-test/<name>.md` is missing:** prefer `/myflow-manual-test <name>` first — the stage gate makes this unlikely since `manual-test-done` is only reached once that guide exists and testing was confirmed.

**When done:** the human reviews and merges the PR at Gate D, then runs `/myflow-review-done <name>` (writes `review-done`), then `/myflow-finish <name>` — verify merged, sync specs, archive. With `automerge`, this command writes `review-done` itself and `/myflow-finish <name>` follows directly.
