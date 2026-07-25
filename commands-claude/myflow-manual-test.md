---
model: sonnet
description: Manual test — generate run guide + functionality checklist MD (Gate C)
---

Use the **openspec-manual-test-superpowers** skill (`.claude/skills/openspec-manual-test-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage `awaiting-review`; on mismatch, stop with the standard mismatch handoff. Entering this command means Gate B passed — set `gates.reviewed: true`.

**Always ask** (AskUserQuestion, default and recommended **No**) whether to skip manual testing for this change. Either answer generates/refreshes `docs/manual-test/<change-name>.md` with:

1. How to run all **involved** apps for this change — **from each app's current apply worktree** (absolute paths in every command; `-PfrontendRoot=…` when KMP is in scope)
2. A **functionality checklist** derived from specs/design/tasks

- **No** → normal guide, unchecked checklist, `gates.tested: false`
- **Yes** → same guide marked `SKIPPED`, every box left unchecked, `gates.tested: "skipped"`

Both answers write `stage: awaiting-test`. The only exception to always-asking is `/myflow-full` with the `skip-manual-test` flag, which pre-answers **Yes** — announce that instead of asking again.

**Create/update the file, stage it and the state file (`git add`, no commit), reply with a link to the file only** (do not paste the guide body), then **stop** for the user to test.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is apply-ready for this stage, use it automatically; if multiple, ask which.

**When done testing:** `/myflow-code-review <name>`. Fixes: `/myflow-do-fix <name>` (loop as many times as needed, then re-test).
