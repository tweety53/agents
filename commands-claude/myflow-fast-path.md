---
model: sonnet
description: Shortened single-session flow for a small, well-understood feature — minimal OpenSpec artifacts, inline TDD, two-agent review, ends at a PR. Five human gates collapse to one. Escalates to the standard pipeline on any size trigger.
---

Use the **openspec-fast-path-superpowers** skill (`.claude/skills/openspec-fast-path-superpowers/SKILL.md`).

Follow that skill exactly. For **small, well-understood features** where the design is already
settled. Writes a terse `proposal.md` + `tasks.md` (no `design.md`, no proposal artifact, no Gate A),
creates the standard worktree, implements **inline with TDD** (no implementer subagents), reviews
with **primary + Bugbot**, runs web-scoped tests and full lint, then commits, pushes, and opens a PR.

**Ends at `stage: awaiting-pr-review`** with `gates.reviewed: false`, `gates.tested: "skipped"`, and
`fastPath: true` — honest values; it never claims a gate that nobody ran.

**Flags:** `checkpoint` (stop on the staged diff before pushing; resumable by re-invoking),
`full-panel` (all six reviewers). **`automerge` is not accepted** — this command never merges.

**Escalation:** more than 3 tasks, or touching `core/ports/`, a migration, or a delta spec, or a new
Critical from the panel → it stops and asks whether to continue or hand off to `/myflow-do`.

`model: sonnet` — no brainstorming happens here; the premise is that the design is settled. If it
isn't, use `/myflow-start` (Opus) instead.

**Input:** Change name from `$ARGUMENTS` or conversation. If omitted: `openspec list --json`; one
active candidate → use it; multiple → ask; none → treat as a new change and ask for a name.

**When done:** review and merge the PR (Gate D), then `/myflow-review-done <name>` →
`/myflow-finish <name>`. Findings on the PR → `/myflow-do-fix <name>`.
