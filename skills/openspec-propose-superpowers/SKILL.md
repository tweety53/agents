---
name: openspec-propose-superpowers
description: Propose an OpenSpec change with Superpowers Basic Workflow #1 brainstorming and #3 writing-plans woven into OpenSpec artifacts. Use for /myflow-start.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "2.4"
---

Propose an OpenSpec change with Superpowers Basic Workflow steps **#1** and **#3** fully intertwined with OpenSpec artifact creation.

**Announce at start:** "Using openspec-propose-superpowers for change `<name>`."

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| **1** | **brainstorming** | Before any OpenSpec artifacts — full checklist, design approval gate |
| **3** | **writing-plans** | After OpenSpec draft artifacts — enrich `tasks.md` to plan quality |

Steps **2, 4–7** run in **openspec-apply-superpowers**. Do not skip or substitute them here.

## Required sub-skills

1. **superpowers:brainstorming** — Basic Workflow **#1** (complete checklist; no shortcuts).
2. **openspec-propose** — CLI-driven artifact creation (steps 2–5).
3. **superpowers:writing-plans** — Basic Workflow **#3** (mandatory; output feeds OpenSpec `tasks.md`).

## Workflow

### A. Understand the change

- If a change name or description was given, use it (derive kebab-case name from description if only a description was given).
- **If both are omitted:** run `openspec list --json`, filter to non-archived changes with incomplete planning artifacts. Exactly one match → resume it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → ask what to build.
- If a change with that name exists: ask continue vs new name.
- **Resuming a change whose artifacts already exist** (e.g. recovering from a failure between
  publishing the artifact and writing state): do not rebuild the proposal page from scratch.
  The artifact's source path is deterministic (see **Publish the proposal artifact**), so
  republish from the existing source file at
  `/Users/tweety53/Agents/myflow/state/<project-key>/<name>-proposal-artifact.html` — this
  recovers the same URL — then write state normally.

### B. Basic Workflow #1 — Brainstorming (before OpenSpec)

Invoke **superpowers:brainstorming** in full:

- Complete checklist items **1–8** (explore context → visual companion JIT if needed → clarifying questions → approaches → design sections → design doc → self-review → user reviews spec).
- **Do not** skip the design-doc step. Save to `docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and commit when the brainstorming skill requires it.
- **HARD-GATE:** Do not run `openspec new change` until the user approves the design (brainstorming item 5 + 8).
- For multi-subsystem work: decompose before proposing.

**OpenSpec bridge:** Treat the approved brainstorming design as the source for OpenSpec `design.md` content (adapt format to OpenSpec template; do not duplicate conflicting designs).

### Architect pass (always runs)

Dispatch an **architect subagent** before writing any artifact. It always runs — do not gate it
behind a heuristic; a skipped architect on a change that needed one is the expensive error.

Give the subagent: the brainstormed intent, the affected modules, and the project's layout rules
from `CLAUDE.md`. Ask it to identify the architectural decisions this change genuinely forces —
module placement, persistence shape, API surface, boundary ownership, migration strategy — and for
each to produce: the decision, 2–3 concrete options, the tradeoffs of each, and a recommendation.

Then put each decision to the user with **AskUserQuestion**, one question per decision,
recommendation first and labelled `(Recommended)`.

**If the architect reports no significant choices, say so in one line and continue** — do not
manufacture questions to justify the pass.

Record the outcome in `design.md` under a `## Decisions` heading, one entry per decision:

```markdown
### <the decision>

**ID:** <kebab-case-slug>
**Status:** active
**Chosen:** <option> — <one-line rationale>
**Considered:** <other options, each with the tradeoff that ruled it out>
```

**ID** is a short kebab-case slug derived from the decision (e.g. `persistence-shape`,
`api-surface`). It is assigned once, at creation, and is **immutable** — the heading prose may be
reworded across rounds, but the ID never changes. This ID is the match key `/myflow-start-fix`
(Task 6) uses to **supersede** a decision when feedback reopens it: it keeps the old entry, sets
its `**Status:**` to `superseded by <new-id>`, and appends the new entry with a fresh ID rather
than duplicating or silently rewriting the old one. Matching on the free-text heading alone would
break on any rewording, so the ID — not the heading — is authoritative.

This section is what keeps the reasoning alive during implementation, when the alternatives are
no longer visible.

### C. OpenSpec change + draft artifacts

After design approval, follow **openspec-propose** steps 2–5:

```bash
openspec new change "<name>"
openspec status --change "<name>" --json
# For each ready artifact:
openspec instructions <artifact-id> --change "<name>" --json
```

Create all artifacts required by `applyRequires`:

- **proposal.md** — what & why (from brainstorm)
- **specs/** — delta specs (from brainstorm + design)
- **design.md** — how (from approved brainstorming design; aligned with spec file above)
- **tasks.md** — initial checkbox scaffold (sections + high-level tasks only; **writing-plans enriches this next**)

Do not copy `<context>` / `<rules>` blocks from CLI instructions into artifact files.

The state file is not written yet — it is written once, at the end (see **Write final state and handoff**), once the artifact URL and decision count are known. Do not create it here.

### D. Basic Workflow #3 — Writing plans (mandatory)

Invoke **superpowers:writing-plans** with inputs:

- Approved design: OpenSpec `design.md` + delta specs + `docs/superpowers/specs/…` if present
- Target: enrich `<changeRoot>/tasks.md` to writing-plans quality (exact paths, verification commands, bite-sized steps)

**OpenSpec bridge:**

- **Canonical apply plan:** `<changeRoot>/tasks.md` — every checkbox must meet writing-plans task structure where the schema allows.
- **Optional mirror:** also save full plan to `docs/superpowers/plans/YYYY-MM-DD-<name>.md`; if saved, add a one-line pointer at the top of `tasks.md`.
- Run writing-plans self-review (spec coverage, placeholder scan, type consistency) before finishing.

Add this header to `tasks.md`:

```markdown
> **Execution:** `/myflow-do` runs Basic Workflow #2–#6 via `openspec-apply-superpowers` (#7 runs later, in `/myflow-review`). Mark each checkbox when its task passes spec + quality review (SDD #6).
```

### Publish the proposal artifact

Load the `artifact-design` skill first, then build a single self-contained page containing:
the proposal's why and what; the design including the `## Decisions` section; the delta specs;
and the task list. Publish it with the Artifact tool.

Write the page's source file to the deterministic path
`/Users/tweety53/Agents/myflow/state/<project-key>/<name>-proposal-artifact.html`, using the same
`<project-key>` resolution (`--git-common-dir`) as **State file** in
`rules/myflow-manual-review.mdc`. This keeps the file outside the repo, beside the state file —
never `git add` or commit it. `/myflow-start-fix` republishes to this **same** file path, which is
what keeps the artifact URL stable across revision rounds.

Record the returned URL in the state file as `artifactUrl`. The page is the review surface for the
proposal gate — it is what the user reads before running `/myflow-start-done`.

### E. Write final state and handoff

Write the state file per **State file** in `rules/myflow-manual-review.mdc`
(`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`;
`mkdir -p` its directory first):

```json
{
  "stage": "awaiting-proposal-review",
  "gates": { "reviewed": null, "tested": null, "prOpened": null, "prMerged": null },
  "worktree": null,
  "branch": null,
  "originStage": null,
  "artifactUrl": "<published URL>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

The state file lives **outside** the repo — do **not** `git add` it, commit it, or archive it. Only the planning artifacts under `<changeRoot>` are staged.

```bash
openspec status --change "<name>"
```

Hand off:

```
## Proposal Ready — Review Required

**Change:** <name>
**Artifact:** <artifactUrl>
**Decisions recorded:** <N> | none

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute main checkout path>"

**Next:**
- Changes to the plan → `/myflow-start-fix <name>`
- Plan looks right → `/myflow-start-done <name>`, then `/myflow-do <name>`
```

The IntelliJ path is the **main checkout**, not a worktree — no worktree exists at this stage; the
proposal artifacts live in the main checkout. Resolve it the same way as `<project-key>`
(`--git-common-dir`).

## Guardrails

- **Never skip** brainstorming (#1) or writing-plans (#3) for net-new features or behavior changes.
- **Never skip** the architect pass — it always runs and self-assesses; it may conclude "no significant choices" and exit, but it must run.
- **Never** skip the design approval gate before `openspec new change`.
- **Never** leave `tasks.md` as a thin scaffold — writing-plans enrichment is mandatory.
- **Never** finish this skill without publishing the proposal artifact and recording `artifactUrl` in the state file — it is the review surface for the new proposal gate.
- Do not start implementation (#2+) in this skill — propose only.
- Prefer reasonable decisions over blocking; pause only on genuine ambiguity.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Propose with Superpowers | `/myflow-start <name>` |
| Then implement (#2–#6) | `/myflow-do <name>` |
