---
name: openspec-manual-test-superpowers
description: >-
  Use for /myflow-manual-test after Gate B code review and before /myflow-code-review
  when a per-change manual test guide (run apps + functionality checklist) is needed.
  Always asks whether to skip Gate C (default No); if skipped, the same guide is
  generated but marked SKIPPED instead of checking any boxes.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI.
metadata:
    author: gymie
    version: "1.4"
---

Generate (or refresh) a **manual test guide** markdown file for an OpenSpec change, then **give the user a link to that file only** and stop for Gate C (manual testing). No commits, push, merge, or archive in this stage.

**Announce at start:** "Using openspec-manual-test-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Stage gate

Requires stage **`awaiting-review`** per **Stage transitions** in `rules/myflow-manual-review.mdc`. On mismatch, stop with the standard mismatch handoff and AskUserQuestion override (default: **No**).

Entering this command means Gate B passed, so set `gates.reviewed: true`.

## Skip prompt (always ask)

Before generating the guide, **always** ask via AskUserQuestion:

> **Skip manual testing for this change?**
> - **No — write the checklist and test it** *(default, recommended)* — normal guide; you run the apps and check boxes.
> - **Yes — skip Gate C** — same guide, marked `SKIPPED`, every box left unchecked.

- **No** → normal mode; `gates.tested: false`
- **Yes** → skip mode; `gates.tested: "skipped"`

Both answers advance the stage to `awaiting-test`. Never default to skip; never infer skip from context. The only exception is `/myflow-full` with the `skip-manual-test` flag, which pre-answers **Yes** — in that case announce that the flag pre-answered the prompt rather than asking again.

`/myflow-manual-test-skip` no longer exists; this prompt replaces it.

## Pipeline position

```text
start → do → manual review (Gate B) → manual test (Gate C) → code review → finish
```

This stage is **Gate C**. It sits after the user has reviewed staged code (Gate B, with optional `/myflow-do-fix` rounds) and before `/myflow-code-review`.

## Required inputs

1. Change name from `$ARGUMENTS` / conversation, **or auto-resolved when omitted** (see step 1).
2. OpenSpec artifacts for that change (`proposal.md`, `design.md`, `tasks.md`, delta `specs/**`).

## Workflow

### 1. Resolve change

```bash
openspec list --json
```

If a name was given, use it directly. **If no name was given:**

- Filter the list to changes that are apply-ready for this stage (have an apply worktree / applied tasks, not yet archived).
- Exactly one match → use it automatically; announce which one ("No change named — using the only active change: `<name>`.").
- Multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess.
- Zero matches → **stop**; suggest `/myflow-do <name>` first.

```bash
openspec status --change "<name>" --json
```

- Announce: "Manual test guide for change: `<name>`."
- Resolve `changeRoot` from CLI JSON.
- If change missing or not apply-ready: **stop** — suggest `/myflow-do <name>`.
- Prefer the apply worktree if present (from `.superpowers/sdd/progress-*.md` / `progress.md`, or `openspec/<name>`); otherwise use the main checkout. Treat that checkout/worktree root as `repoRoot`.

### 2. Detect involved apps

From proposal Impact, design, tasks, and any staged/uncommitted paths, mark which surfaces are in scope:

| App | Detect when | Local URL |
|-----|-------------|-----------|
| **Backend + gateway** (`gymie`) | Always if any backend/API/Flyway/module work; default on for most changes | Gateway http://localhost:8080 · App http://localhost:8081 |
| **KMP frontend** (`gymie-frontend`) | tasks/design mention KMP, Compose, web/mobile, `gymie-frontend`, `:webApp` | http://localhost:3000 |
| **Admin frontend** (`gymie-admin-frontend`) | tasks/design mention admin panel, Next.js admin, `gymie-admin-frontend` | http://localhost:3001 |

Omit run sections for apps clearly out of scope (e.g. proposal says "Out of scope: admin panel").

### 2b. Resolve absolute roots for every involved app (required)

**All "How to run" commands must target the apply worktree (or main checkout if no worktree) for that app — never a different branch's checkout.**

For each involved app, resolve an **absolute** root path and put it in the guide header + every `cd` / Gradle / npm command:

| Source (prefer in order) | What to read |
|--------------------------|--------------|
| `.superpowers/sdd/progress-<name>.md` or `progress.md` | `Backend worktree`, `Frontend worktree`, admin worktree lines |
| `git worktree list` in each repo | Path whose branch is `openspec/<name>` (or the apply branch) |
| Fallback | Main checkout absolute path only if no apply worktree exists |

Record variables (use real absolute paths in the MD, not these names):

- `BACKEND_ROOT` — `gymie` apply worktree (this is also `repoRoot` for writing the guide)
- `FRONTEND_ROOT` — `gymie-frontend` apply worktree when KMP is in scope
- `ADMIN_ROOT` — `gymie-admin-frontend` apply worktree when admin is in scope

Hard rules for the guide:

1. **Every** shell block starts with `cd <absolute-root>` (or uses absolute `-P…=` paths). No bare relative `cd ../gymie-frontend` / `cd ../gymie-admin-frontend`.
2. When KMP is in scope, `devStart` **must** pass `-PfrontendRoot=<FRONTEND_ROOT>` (default sibling `../gymie-frontend` is wrong from `.worktrees/`).
3. Optional per-app sections (`:webApp:jsBrowserDevelopmentRun`, `npm run dev`, backend-only bootRun) must `cd` into that app's absolute worktree root.
4. Prerequisites must say: do **not** run from main `develop` checkouts when apply worktrees exist.
5. Header **Branch / worktree** line lists absolute paths for every involved app.

### 3. Build the functionality checklist

Derive checkbox items from, in order of preference:

1. Delta spec **Requirements / Scenarios** (user-visible behaviors)
2. `design.md` acceptance criteria / UX flows
3. User-facing `tasks.md` items (skip pure infra like "add detekt", "wire Gradle" unless they have a verifiable runtime effect)

Each item must be:

- A concrete **manual** action or observation ("Create a workout for tomorrow on Calendar", not "WorkoutService works")
- One behavior per checkbox
- Grouped by capability or screen when helpful

Include negative/edge cases from specs when present (validation errors, empty states, auth redirects).

### 4. Write the guide file

**Path (required):** `<repoRoot>/docs/manual-test/<change-name>.md`

Create `docs/manual-test/` if missing. Use the structure in [manual-test-template.md](manual-test-template.md). Fill every placeholder; do not leave `TBD` for involved-app run steps.

**Always include:**

1. **Change** name + one-line purpose (from proposal Why/What)
2. **How to run** — only involved apps, **all commands rooted at resolved absolute worktree paths** (step 2b):
   - Preferred full stack from `BACKEND_ROOT`: `docker compose up -d`, `./gradlew dbSeed` when seeded users/catalog needed, then `./gradlew devStart` with `-PfrontendRoot=<FRONTEND_ROOT>` when KMP is in scope
   - Admin-only / KMP-only optional blocks use `ADMIN_ROOT` / `FRONTEND_ROOT` absolute `cd`s
   - URLs, default local credentials when relevant (`superadmin@gymie.dev` / `GymieDev1` for admin; note user auth separately)
3. **Functionality checklist** — unchecked `- [ ]` items from step 3
4. **Sign-off** — ready for `/myflow-code-review <name>`

If `docs/manual-test/<change-name>.md` already exists:

- **Refresh** run instructions and merge new checklist items from current specs
- **Preserve** any boxes the user already checked (`[x]`) unless the requirement was removed

Legacy: if only `<changeRoot>/manual-test.md` exists, migrate content into `docs/manual-test/<change-name>.md` (preserving checked boxes), then prefer the new path going forward.

**Skip mode only (user answered Yes to the skip prompt):** add a status line directly under **Next after sign-off:** in the header —

```
**Manual test status:** SKIPPED — YYYY-MM-DD (Gate C intentionally bypassed)
```

In skip mode, leave **every** functionality-checklist and sign-off checkbox unchecked (`- [ ]`) — nothing was actually run or verified, so nothing may read as verified. The status line is the only thing that marks the gate bypassed; `/myflow-code-review` looks for it. If a guide already has checked boxes from a prior normal run and skip mode is invoked afterward (e.g. re-run after a change), preserve those existing checks rather than un-checking them — skip mode never removes evidence, it only adds the bypass marker when a fresh guide has none.

### 5. Stage the guide (no commit)

In `repoRoot` (usually `gymie` or its apply worktree):

```bash
mkdir -p docs/manual-test
git add "docs/manual-test/<change-name>.md"
```

Do **not** `git commit`. Leave for archive with the rest of the apply work.

### 5b. Write state

Per **State file** in `rules/myflow-manual-review.mdc`:

```json
{
  "stage": "awaiting-test",
  "gates": { "reviewed": true, "tested": <false | "skipped">, "prOpened": null, "prMerged": null },
  "worktree": "<unchanged>",
  "branch": "<unchanged>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-manual-test"
}
```

Stage it with the guide: `git add openspec/changes/<name>/.myflow-state.json`. Do not commit.

### 6. Present the link only and stop

**Do not paste the guide body into chat.** Reply with a short Gate C handoff that includes a **clickable markdown link** to the file (relative path is fine; absolute path optional).

Then **stop** — do not archive, do not start apps for them unless they ask.

Normal mode:

```
## Manual Test Guide Ready (Gate C)

**Change:** <name>
**Guide:** [docs/manual-test/<change-name>.md](docs/manual-test/<change-name>.md)
**Involved apps:** <list>
**Git:** docs/manual-test/<change-name>.md staged (uncommitted)

**What to do:**
1. Open the guide link and follow "How to run"
2. Work through every checklist item; check boxes in the file as you go
3. Fixes needed → `/myflow-do-fix <name>`
4. When satisfied → `/myflow-code-review <name>`
```

Skip mode:

```
## Manual Test Guide Generated — Gate C SKIPPED

**Change:** <name>
**Guide:** [docs/manual-test/<change-name>.md](docs/manual-test/<change-name>.md)
**Status:** SKIPPED — checklist left unchecked, marked bypassed
**Git:** docs/manual-test/<change-name>.md staged (uncommitted)

**Note:** `/myflow-code-review <name>` will detect this as intentionally skipped rather than incomplete, but will still surface it before proceeding.
```

## Guardrails

- **Never commit, push, merge, or open a PR** in this stage.
- **Never skip** writing/updating `docs/manual-test/<change-name>.md` — the point of the command is to produce that file.
- **Never dump the full guide markdown into the chat response** — link only.
- Do not invent features not in proposal/specs/design/tasks.
- Do not include apps marked out of scope.
- Do not emit run commands that `cd` into main checkouts when apply worktrees exist for those apps.
- Do not use relative sibling paths (`../gymie-frontend`, `../gymie-admin-frontend`) in guide commands — always absolute worktree (or main-checkout) roots.
- Do not run `/myflow-code-review` or `/myflow-finish` from this skill.
- Keep the guide concise and actionable; prefer checklists over essays.
- **Skip mode never checks a box it didn't verify** — only the `SKIPPED` status line marks the gate bypassed.
- **Always ask the skip prompt** — never default to skip, never infer it from context. Only `/myflow-full skip-manual-test` may pre-answer it.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Generate / refresh manual test guide (asks whether to skip) | `/myflow-manual-test <name>` |
| After checklist done (or skip acknowledged) | `/myflow-code-review <name>` |
| Fixes from testing | `/myflow-do-fix <name>` |
