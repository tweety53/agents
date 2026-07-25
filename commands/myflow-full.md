---
name: /myflow-full
id: myflow-full
category: myflow
description: Full cycle — start → do → manual review → manual test → review → Gate D (PR open, stop)
---

**Model:** Sonnet (or your default) is fine for most of this pipeline. If Phase A ends up running brainstorming (a fresh proposal, not `skip-propose`), switch to Opus for that phase manually — Cursor can't auto-switch mid-command. For brainstorming-heavy work, prefer running `/myflow-start` standalone on Opus first, then `/myflow-full <name> skip-propose` on Sonnet for the rest.

Use the **openspec-full-cycle-superpowers** skill (`.cursor/skills/openspec-full-cycle-superpowers/SKILL.md`).

Follow that skill exactly. Pipeline: **start → do (#2–#6, no commits) → Gate B manual review (optional `/myflow-do-fix` loop) → Gate C manual test (skip prompt; optional `/myflow-do-fix` loop) → review (coverage check, commit + push + open PR) → Gate D — STOP**.

`/myflow-full` ends at Gate D with an open, unmerged PR **unless `automerge` was passed**. Merging is normally a human action on the forge; `automerge` is the one explicit exception — passed through to `/myflow-review`, it merges immediately, so there is no PR to stop at and the cycle ends at `review-done`. `/myflow-finish <name>` (verify merged, sync specs, archive) is always a separate command you run yourself after merging — this command never invokes it.

Also follow `.cursor/rules/myflow-manual-review.mdc`.

**Input:** Change name + what to build from `$ARGUMENTS` or conversation. If the name is omitted and a description implies a new change, propose with that description; if both are omitted: run `openspec list --json`, use the sole active change automatically, or ask if there are multiple.

**Flags:** `skip-propose`, `propose-only`, `skip-review`, `skip-manual-test`, `automerge`, `commit-during-apply` (legacy) — honor if present in the user message. `skip-manual-test` pre-answers the Gate C skip prompt with **Yes** (announce that it did so); without it, the prompt is asked normally and defaults to **No**. `automerge` is passed through to `/myflow-review`, which merges into the base branch instead of opening a PR; the cycle then ends at `review-done`. It is never inferred and never defaulted on. `no-archive` has been **removed** — the cycle no longer reaches archiving, so it had no effect to preserve.

**Stages (individual):** `/myflow-start` (#1+#3), `/myflow-do` (#2–#6), manual review (Gate B), `/myflow-do-fix` (Gate B/C/D fixes), `/myflow-manual-test` (Gate C, always asks the skip prompt), `/myflow-review` (coverage+commit+push+PR, Gate D), `/myflow-finish` (separate, human-initiated: verify merged+sync+archive)
