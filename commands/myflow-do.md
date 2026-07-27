---
name: /myflow-do
id: myflow-do
category: myflow
description: Do — Superpowers Basic Workflow #2–#6 with OpenSpec tasks (no commits)
---

**Model:** Sonnet (or your default) is fine here — Opus is reserved for `/myflow-start`'s brainstorming stage. Cursor doesn't yet support a per-command model frontmatter field, so this is a recommendation, not an enforced switch.

Use the **openspec-apply-superpowers** skill — installed globally, so let your harness resolve it by name rather than assuming a project-local path.

Follow that skill exactly. Requires stage `proposal-done` — the stage `/myflow-start-done` writes once the human has read the proposal artifact. On mismatch, stop with the standard mismatch handoff. Ends at `awaiting-do-review`. Runs Superpowers Basic Workflow **#2–#6** only (#2 worktree, #3 plan validate, #4 SDD, #5 TDD, #6 review). **No git commits, push, merge, or PR.** **#7 deferred to `/myflow-review`.** Do **not** use the lightweight openspec-apply-change task loop.

Also follow the myflow manual-review rule (`myflow-manual-review.mdc`) — installed globally, so let your harness resolve it rather than assuming a project-local path.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active proposal is apply-ready, use it automatically; if multiple, ask which. Optional flags: `full-panel` (every review-panel re-run uses every roster slot, including both extra principle lenses, over the whole-branch diff instead of the default targeted re-run), `commit-during-apply` (legacy).

**If this change already looks applied** (manual-test guide exists, or a clean final-review-panel record exists, or every original task is checked): ask first whether the user meant `/myflow-do-fix` instead — default/recommended answer is **No, use `/myflow-do-fix`**. Only proceed with a fresh/expanded run if they explicitly say yes.

**When done:** Manual review (Gate B) — optionally `/myflow-do-manual-review <name>`, then `/myflow-do-done <name>` to confirm (writes `do-done`) — then `/myflow-manual-test <name>` (Gate C), `/myflow-manual-test-done <name>`, and `/myflow-review <name>`. Request fixes via `/myflow-do-fix <name>`.
