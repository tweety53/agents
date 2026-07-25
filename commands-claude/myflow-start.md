---
description: Start — Superpowers #1 brainstorming + #3 writing-plans woven into OpenSpec artifacts
model: opus
---

Use the **openspec-propose-superpowers** skill (`.claude/skills/openspec-propose-superpowers/SKILL.md`).

Follow that skill exactly. Runs Superpowers Basic Workflow **#1** (brainstorming) and **#3** (writing-plans) intertwined with OpenSpec artifact creation, plus an always-on **architect pass** (before any artifact is written — it may find no significant decisions and say so) and a **published proposal artifact** (self-contained page, review surface for the new proposal gate) before handoff. Do not skip any of these.

**Input:** Change name (kebab-case) and/or description of what to build — from `$ARGUMENTS` or conversation. If both are omitted: run `openspec list --json`; if exactly one active (non-archived) change has incomplete planning artifacts, resume it automatically; if multiple, ask which; if none, ask what to build.

**Ends at:** `awaiting-proposal-review` (not `start`) — a new human gate. The handoff prints the artifact URL and the `open -na "IntelliJ IDEA"` command for the **main checkout** (no worktree exists yet).

**Next:** `/myflow-start-fix <name>` to revise the plan, or `/myflow-start-done <name>` once reviewed, then `/myflow-do <name>` (#2–#6, no commits). `/myflow-full <name>` runs the gated end-to-end flow.
