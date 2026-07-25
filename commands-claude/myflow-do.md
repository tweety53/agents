---
model: sonnet
description: Do — Superpowers Basic Workflow #2–#6 with OpenSpec tasks (no commits)
---

Use the **openspec-apply-superpowers** skill (`.claude/skills/openspec-apply-superpowers/SKILL.md`).

Follow that skill exactly. Requires stage `proposal-done` — the stage `/myflow-start-done` writes once the human has read the proposal artifact. On mismatch, stop with the standard mismatch handoff. Ends at `awaiting-do-review`. Runs Superpowers Basic Workflow **#2–#6** only (#2 worktree, #3 plan validate, #4 SDD, #5 TDD, #6 review). **No git commits, push, merge, or PR.** **#7 deferred to `/myflow-review`.** Do **not** use the lightweight openspec-apply-change task loop.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: run `openspec list --json`; if exactly one active proposal is apply-ready, use it automatically; if multiple, ask which. Optional legacy flag: `commit-during-apply`.

**If this change already looks applied** (manual-test guide exists, or a clean final-review-panel record exists, or every original task is checked): ask first whether the user meant `/myflow-do-fix` instead — default/recommended answer is **No, use `/myflow-do-fix`**. Only proceed with a fresh/expanded run if they explicitly say yes.

**When done:** Manual review (Gate B) — optionally `/myflow-do-manual-review <name>`, then `/myflow-do-done <name>` to confirm (writes `do-done`) — then `/myflow-manual-test <name>` (Gate C), `/myflow-manual-test-done <name>`, and `/myflow-review <name>`. Request fixes via `/myflow-do-fix <name>`.
