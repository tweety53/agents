---
name: openspec-propose-superpowers
description: Propose an OpenSpec change with Superpowers Basic Workflow #1 brainstorming and #3 writing-plans woven into OpenSpec artifacts. Use for /myflow-start.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "2.2"
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

### B. Basic Workflow #1 — Brainstorming (before OpenSpec)

Invoke **superpowers:brainstorming** in full:

- Complete checklist items **1–8** (explore context → visual companion JIT if needed → clarifying questions → approaches → design sections → design doc → self-review → user reviews spec).
- **Do not** skip the design-doc step. Save to `docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and commit when the brainstorming skill requires it.
- **HARD-GATE:** Do not run `openspec new change` until the user approves the design (brainstorming item 5 + 8).
- For multi-subsystem work: decompose before proposing.

**OpenSpec bridge:** Treat the approved brainstorming design as the source for OpenSpec `design.md` content (adapt format to OpenSpec template; do not duplicate conflicting designs).

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

### Write initial state

Create the change's state file at the user-scoped path resolved per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`; `mkdir -p` its directory first):

```json
{
  "stage": "start",
  "gates": { "reviewed": null, "tested": null, "prOpened": null, "prMerged": null },
  "worktree": null,
  "branch": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-start"
}
```

The state file lives **outside** the repo — do **not** `git add` it, commit it, or archive it. Only the planning artifacts under `<changeRoot>` are staged.

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

### E. Finish

```bash
openspec status --change "<name>"
```

Summarize:

- Change name and `changeRoot` path
- Artifacts created
- Basic Workflow **#1** ✓ (brainstorm design path)
- Basic Workflow **#3** ✓ (`tasks.md` task count + section overview)
- **Next:** `/myflow-do <name>` (runs #2, #4–#6)
- **Full cycle:** `/myflow-full <name>`

## Guardrails

- **Never skip** brainstorming (#1) or writing-plans (#3) for net-new features or behavior changes.
- **Never** skip the design approval gate before `openspec new change`.
- **Never** leave `tasks.md` as a thin scaffold — writing-plans enrichment is mandatory.
- Do not start implementation (#2+) in this skill — propose only.
- Prefer reasonable decisions over blocking; pause only on genuine ambiguity.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Propose with Superpowers | `/myflow-start <name>` |
| Then implement (#2–#6) | `/myflow-do <name>` |
