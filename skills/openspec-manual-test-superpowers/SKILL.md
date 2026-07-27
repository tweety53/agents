---
name: openspec-manual-test-superpowers
description: >-
  Use for /myflow-manual-test after Gate B manual review and before /myflow-review
  when a per-change manual test guide (run apps + functionality checklist) is needed.
  Always asks whether to skip Gate C (default No); if skipped, the same guide is
  generated but marked SKIPPED instead of checking any boxes.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI.
metadata:
    author: gymie
    version: "1.5"
---

Generate (or refresh) a **manual test guide** markdown file for an OpenSpec change, then **give the user a link to that file only** and stop for Gate C (manual testing). No commits, push, merge, or archive in this stage.

**Announce at start:** "Using openspec-manual-test-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Stage gate

Accepts **two** stages per **Stage transitions** in `rules/myflow-manual-review.mdc`:

| Incoming stage | Mode |
|----------------|------|
| `do-done` | **advance** — Gate B passed; write `gates.reviewed: true` and advance to `awaiting-manual-test` |
| `awaiting-manual-test` | **refresh** — re-emit the guide after a fix round; stage stays `awaiting-manual-test` |

Any other stage is a mismatch: stop with the standard mismatch handoff and AskUserQuestion override (default: **No**).

### Refresh mode (`awaiting-manual-test`)

Entered **automatically** — no prompt, no override. A guide refresh after a `/myflow-do-fix` round is a routine operation and must not be routed through the stage-mismatch escape hatch (that same "run anyway" is what would let `/myflow-do` wipe a populated worktree).

In refresh mode:

- Do **not** re-ask the skip question — the decision was made on the advance run. This is what makes `gates.tested: "skipped"` unclobberable.
- Do **not** rewrite `gates.reviewed` — it is already `true` (or `false` if Gate B was explicitly skipped); carry it forward as found.
- Preserve every already-checked box, exactly as the refresh rules in step 4 require.
- Re-emit `stage: awaiting-manual-test` unchanged and carry **all** gates forward untouched.
- Announce: "refreshed after fix round N".

### Advance mode (`do-done`)

Entering this command from `do-done` means Gate B passed, so set `gates.reviewed: true` — **except** under `/myflow-full skip-review`, where Gate B was explicitly skipped: write `gates.reviewed: false` and note the skip in the summary, so the state never claims a review that never happened.

## Skip prompt (advance mode only — always ask there)

Before generating the guide **on an advance run**, always ask via AskUserQuestion (in refresh mode this prompt is not asked at all):

> **Skip manual testing for this change?**
> - **No — write the checklist and test it** *(default, recommended)* — normal guide; you run the apps and check boxes.
> - **Yes — skip Gate C** — same guide, marked `SKIPPED`, every box left unchecked.

- **No** → normal mode; `gates.tested: false`
- **Yes** → skip mode; `gates.tested: "skipped"`

Both answers advance the stage to `awaiting-manual-test`. Never default to skip; never infer skip from context. The only exception is `/myflow-full` with the `skip-manual-test` flag, which pre-answers **Yes** — in that case announce that the flag pre-answered the prompt rather than asking again.

`/myflow-manual-test-skip` no longer exists; this prompt replaces it.

## Pipeline position

See **Pipeline stages** in `rules/myflow-manual-review.mdc` for the full twelve-stage sequence. This stage is **Gate C**, entered from `do-done` and sitting between Gate B (manual review, with optional `/myflow-do-fix` rounds) and `/myflow-review`.

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

### 2. Load the project configuration

Read `<main checkout>/.myflow/project.md` — the project's own apps, run/test/lint commands,
local URLs, and credentials. Its format and the auto-detect fallback are defined once, under
**Project configuration** in `rules/myflow-manual-review.mdc`; that section is canonical and is
not restated here.

- **File present** → take the app list from its `## apps`, the start/stop commands from `## run`,
  and any sign-in credentials from `## credentials`. Use them verbatim; do not "improve" a
  command the project wrote down.
- **File absent, or a key absent** → **auto-detect from this repository**: its build files,
  `package.json` scripts, compose files, and existing run docs. Announce that no project file
  was found and what you detected from.
- **Neither** → leave that part of the guide an explicit `TBD` and say so in the handoff. Never
  substitute app names, task names, ports, URLs, or credentials from any other project.

### 2a. Detect involved apps

From proposal Impact, design, tasks, and any staged/uncommitted paths, mark which of the
configured (or detected) apps are in scope. `## apps` rows carry a "when to include" note — use
it as the detection rule, together with any app name, module, or path the change's artifacts
mention.

Omit run sections for apps clearly out of scope (e.g. proposal says "Out of scope: admin panel").
When exactly one app exists, it is in scope.

### 2b. Resolve absolute roots for every involved app (required)

**All "How to run" commands must target the apply worktree (or main checkout if no worktree) for that app — never a different branch's checkout.**

For each involved app, resolve an **absolute** root path and put it in the guide header + every `cd` and every command:

| Source (prefer in order) | What to read |
|--------------------------|--------------|
| `.superpowers/sdd/progress-<name>.md` or `progress.md` | the recorded worktree line for that app |
| `git worktree list` in that app's repo | Path whose branch is `openspec/<name>` (or the apply branch) |
| Fallback | that app's main-checkout absolute path from `## apps`, only if no apply worktree exists |

Record one absolute root per involved app, keyed by the app's name from `## apps` (use the real
absolute paths in the MD, never a variable name). The root of the repo holding the change is also
`repoRoot`, where the guide is written.

Hard rules for the guide:

1. **Every** shell block starts with `cd <absolute-root>` (or passes absolute paths as parameters). Never a bare relative sibling `cd ../<other-app>`.
2. When a `## run` command takes another app's root as a parameter, pass that app's **resolved absolute worktree root** — the command's own default is a sibling path and is wrong from inside a worktree.
3. Optional per-app sections must `cd` into that app's absolute worktree root.
4. Prerequisites must say: do **not** run from main-branch checkouts when apply worktrees exist.
5. Header **Branch / worktree** line lists absolute paths for every involved app.

### 3. Build the functionality checklist

Derive checkbox items from, in order of preference:

1. Delta spec **Requirements / Scenarios** (user-visible behaviors)
2. `design.md` acceptance criteria / UX flows
3. User-facing `tasks.md` items (skip pure infra like "add a linter", "wire the build" unless they have a verifiable runtime effect)

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
   - The full-stack sequence from `## run`, rooted at the app that owns it, with any other-app root passed as an absolute path
   - Single-app blocks from `## run` for the involved apps only, each with its own absolute `cd`
   - Local URLs from `## apps`, and sign-in credentials from `## credentials` only when the checklist actually needs a signed-in session
3. **Functionality checklist** — unchecked `- [ ]` items from step 3
4. **Sign-off** — ready for `/myflow-review <name>`

If `docs/manual-test/<change-name>.md` already exists:

- **Refresh** run instructions and merge new checklist items from current specs
- **Preserve** any boxes the user already checked (`[x]`) unless the requirement was removed

Legacy: if only `<changeRoot>/manual-test.md` exists, migrate content into `docs/manual-test/<change-name>.md` (preserving checked boxes), then prefer the new path going forward.

**Skip mode only (user answered Yes to the skip prompt):** add a status line directly under **Next after sign-off:** in the header —

```
**Manual test status:** SKIPPED — YYYY-MM-DD (Gate C intentionally bypassed)
```

In skip mode, leave **every** functionality-checklist and sign-off checkbox unchecked (`- [ ]`) — nothing was actually run or verified, so nothing may read as verified. The status line is the only thing that marks the gate bypassed; `/myflow-review` looks for it. If a guide already has checked boxes from a prior normal run and skip mode is invoked afterward (e.g. re-run after a change), preserve those existing checks rather than un-checking them — skip mode never removes evidence, it only adds the bypass marker when a fresh guide has none.

### 5. Stage the guide (no commit)

In `repoRoot` (the change's repo, or its apply worktree):

```bash
mkdir -p docs/manual-test
git add "docs/manual-test/<change-name>.md"
```

Do **not** `git commit`. Leave for archive with the rest of the apply work.

### 5b. Write state

Resolve the state file path per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`). It lives outside the repo: **never `git add` it, never commit it.**

**Read the existing file first and carry forward everything this command does not own.** Writes render the whole object, so anything not carried forward would be silently erased:

```json
{
  "stage": "awaiting-manual-test",
  "gates": {
    "reviewed": <true on advance | false under skip-review | unchanged in refresh mode>,
    "tested": <false | "skipped" on advance; unchanged in refresh mode>,
    "prOpened": "<carried forward from the file as read>",
    "prMerged": "<carried forward from the file as read>"
  },
  "worktree": "<unchanged>",
  "branch": "<unchanged>",
  "originStage": "<carried forward from the file as read>",
  "artifactUrl": "<carried forward from the file as read>",
  "jiraIssue": "<unchanged — carried forward from the file as read>",
  "fastPath": "<carried forward from the file as read>",
  "REVIEWED_TREE": "<carried forward from the file as read>",
  "MERGE_BASE": "<carried forward from the file as read>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-manual-test"
}
```

**`artifactUrl` and `jiraIssue` are carried forward, never dropped** — writes render the whole object, so omitting either would erase the published proposal link that `myflow-status` surfaces or silently unlink the change from its issue. This command makes **no** Jira call of its own — see **Jira integration** in `rules/myflow-manual-review.mdc`.

**Monotonic gates (mandatory).** Never reset `prOpened`/`prMerged` to `null` — carry the read values through verbatim. Never lower `gates.tested`: `true` and `"skipped"` are sticky and this command never demotes them (it only sets `false`/`"skipped"` on a fresh advance run where the prior value was `null`). Never write a `stage` earlier than the one found — from `awaiting-manual-test`, re-emit `awaiting-manual-test`.

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

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"
Guide: <absolute path to docs/manual-test/<change-name>.md>

**What to do:**
1. Open the guide link and follow "How to run"
2. Work through every checklist item; check boxes in the file as you go
3. Fixes needed → `/myflow-do-fix <name>`
4. When satisfied → `/myflow-manual-test-done <name>` to mark testing complete, then `/myflow-review <name>`
```

Skip mode:

```
## Manual Test Guide Generated — Gate C SKIPPED

**Change:** <name>
**Guide:** [docs/manual-test/<change-name>.md](docs/manual-test/<change-name>.md)
**Status:** SKIPPED — checklist left unchecked, marked bypassed
**Git:** docs/manual-test/<change-name>.md staged (uncommitted)

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"
Guide: <absolute path to docs/manual-test/<change-name>.md>

**Note:** `/myflow-review <name>` will detect this as intentionally skipped rather than incomplete, but will still surface it before proceeding. Run `/myflow-manual-test-done <name>` first to mark the gate confirmed.
```

## Guardrails

- **Never commit, push, merge, or open a PR** in this stage.
- **Never skip** writing/updating `docs/manual-test/<change-name>.md` — the point of the command is to produce that file.
- **Never dump the full guide markdown into the chat response** — link only.
- Do not invent features not in proposal/specs/design/tasks.
- Do not include apps marked out of scope.
- Do not emit run commands that `cd` into main checkouts when apply worktrees exist for those apps.
- Do not use relative sibling paths (`../<other-app>`) in guide commands — always absolute worktree (or main-checkout) roots.
- **Never emit an app name, task name, port, URL, or credential that did not come from this project's `.myflow/project.md` or from detection in this repository** — no value carried over from another project, ever. When something cannot be resolved, write `TBD` and say so.
- Do not run `/myflow-review` or `/myflow-finish` from this skill.
- Keep the guide concise and actionable; prefer checklists over essays.
- **Skip mode never checks a box it didn't verify** — only the `SKIPPED` status line marks the gate bypassed.
- **Always ask the skip prompt on an advance run** — never default to skip, never infer it from context. Only `/myflow-full skip-manual-test` may pre-answer it. **Never ask it in refresh mode** — that would risk clobbering a recorded `"skipped"`.
- **Never lower a gate value or rewind the stage** — carry `prOpened`/`prMerged` forward from the file as read; gates are monotonic.
- **Never `git add` or commit the state file** — it is user-scoped and outside the repo.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Generate / refresh manual test guide (asks whether to skip) | `/myflow-manual-test <name>` |
| After checklist done (or skip acknowledged) | `/myflow-review <name>` |
| Fixes from testing | `/myflow-do-fix <name>` |
