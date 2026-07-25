---
model: sonnet
description: Manual test — generate run guide + functionality checklist MD (Gate C)
---

Use the **openspec-manual-test-superpowers** skill (`.claude/skills/openspec-manual-test-superpowers/SKILL.md`).

Follow that skill exactly. Accepts stage `awaiting-review` (**advance** — Gate B passed, set `gates.reviewed: true`) and stage `awaiting-test` (**refresh** after a fix round — entered automatically, no prompt: the skip question is not re-asked, `gates.reviewed` is untouched, checked boxes are preserved, and the stage stays `awaiting-test`). Any other stage: stop with the standard mismatch handoff. Under `/myflow-full skip-review`, record `gates.reviewed: false` — Gate B was explicitly skipped.

**On an advance run, always ask** (AskUserQuestion, default and recommended **No**) whether to skip manual testing for this change; in refresh mode the question is not asked at all, so a recorded `"skipped"` can never be clobbered. Either answer generates/refreshes `docs/manual-test/<change-name>.md` with:

1. How to run all **involved** apps for this change — **from each app's current apply worktree** (absolute paths in every command; `-PfrontendRoot=…` when KMP is in scope)
2. A **functionality checklist** derived from specs/design/tasks

- **No** → normal guide, unchecked checklist, `gates.tested: false`
- **Yes** → same guide marked `SKIPPED`, every box left unchecked, `gates.tested: "skipped"`

Both answers write `stage: awaiting-test`. The only exception to always-asking is `/myflow-full` with the `skip-manual-test` flag, which pre-answers **Yes** — announce that instead of asking again.

**Create/update the file, stage it and the state file (`git add`, no commit), reply with a link to the file only** (do not paste the guide body), then **stop** for the user to test.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is apply-ready for this stage, use it automatically; if multiple, ask which.

**When done testing:** `/myflow-code-review <name>`. Fixes: `/myflow-do-fix <name>` (loop as many times as needed, then re-test).
