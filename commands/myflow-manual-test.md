---
name: /myflow-manual-test
id: myflow-manual-test
category: myflow
description: Manual test — generate run guide + functionality checklist MD (Gate C)
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-manual-test-superpowers** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Accepts stage `do-done` (**advance** — Gate B passed and was confirmed by `/myflow-do-done`, set `gates.reviewed: true`) and stage `awaiting-manual-test` (**refresh** after a fix round — entered automatically, no prompt: the skip question is not re-asked, `gates.reviewed` is untouched, checked boxes are preserved, and the stage stays `awaiting-manual-test`). Any other stage: stop with the standard mismatch handoff. Under `/myflow-full skip-review`, record `gates.reviewed: false` — Gate B was explicitly skipped.

**On an advance run, always ask** (AskUserQuestion, default and recommended **No**) whether to skip manual testing for this change; in refresh mode the question is not asked at all, so a recorded `"skipped"` can never be clobbered. Either answer generates/refreshes `docs/manual-test/<change-name>.md` with:

1. How to run all **involved** apps for this change — **from each app's current apply worktree** (absolute paths in every command; any build parameters the run needs come from `.myflow/project.md`'s `## run` section, not from memory)
2. A **functionality checklist** derived from specs/design/tasks

- **No** → normal guide, unchecked checklist, `gates.tested: false`
- **Yes** → same guide marked `SKIPPED`, every box left unchecked, `gates.tested: "skipped"`

Both answers write `stage: awaiting-manual-test`. The only exception to always-asking is `/myflow-full` with the `skip-manual-test` flag, which pre-answers **Yes** — announce that instead of asking again.

**Create/update the file and stage it (`git add`, no commit) — the state file lives outside the repo and is never staged. Reply with a link to the file only** (do not paste the guide body), then **stop** for the user to test.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active change is apply-ready for this stage, use it automatically; if multiple, ask which.

**When done testing:** `/myflow-manual-test-done <name>` to confirm (writes `manual-test-done`), then `/myflow-review <name>`. Fixes: `/myflow-do-fix <name>` (loop as many times as needed, then re-test).
