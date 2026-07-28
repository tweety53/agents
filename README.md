# agents-data

Portable, project-agnostic agent configuration — the **myflow** pipeline (OpenSpec +
Superpowers), its skills, slash commands, and always-on rules.
Contains the rule set, an index of the skills, and installation instructions for every
supported AI harness.

**This repository is the source of truth.** Edit the files here, then run
`./setup.sh global` to publish them to every harness. Nothing is ever synced *into*
this repo from a project checkout.

---

## What's in here

```
agents-data/
├── README.md                          ← this file
├── CLAUDE.md                          ← drop into project root for Claude Code
├── AGENTS.md                          ← drop into project root for Codex / OpenAI CLI
├── setup.sh                           ← installer: `global` (recommended) or per-project harness installs
├── rules/
│   ├── myflow-manual-review.mdc       ← myflow trigger + contract pointers (always-on stub, installed globally)
│   ├── lint-fix-priority.mdc          ← never suppress/bypass linters (always-on, installed globally)
│   └── kotlin-backend-development-standard.mdc  ← opt-in: named in a project's `.myflow/project.md`, rendered into that project's CLAUDE.md + AGENTS.md
├── scripts/
│   ├── check-vocabulary.sh            ← guards the pipeline vocabulary used across these files
│   └── test-setup.sh                  ← regression harness for setup.sh (sandboxed HOME under /tmp)
├── commands/                          ← Cursor slash commands (myflow + opsx)
├── commands-claude/                   ← Claude Code slash commands (myflow only)
└── skills/                            ← OpenSpec / /myflow skills
    ├── README.md                      ← myflow command map
    ├── openspec-propose/
    ├── openspec-propose-superpowers/
    ├── openspec-apply-change/
    ├── openspec-apply-superpowers/   ← includes the review-panel reviewer prompts + engineering-principles.md
    ├── openspec-apply-fix-superpowers/  ← fixes for Gate B/C/D findings, reuses the panel prompts above
    ├── openspec-fast-path-superpowers/  ← shortened single-session flow for small features
    ├── openspec-manual-test-superpowers/
    ├── openspec-review-superpowers/  ← coverage check, tests/linters, commit + push + open PR (never merges)
    ├── openspec-archive-change/
    ├── openspec-archive-superpowers/  ← the "finish" stage: verify PR merged, sync specs, archive
    ├── openspec-full-cycle-superpowers/
    ├── openspec-explore/
    ├── openspec-sync-specs/
    ├── openspec-update-change/
    ├── openspec-propose-fix-superpowers/  ← revise proposal after Gate A review, republish artifact to same URL
    ├── myflow-status/                 ← read-only stage report for open changes
    ├── myflow-info/                   ← reads skills/myflow-contracts/pipeline.md and explains the pipeline
    ├── myflow-state-advance/          ← pure state write used by every `*-done`/`*-manual-review` command
    └── myflow-contracts/              ← on-demand contract definitions (state file, self-heal, project config, Jira)
```

**Rules** — whether a rule is always-on is a property of the rule itself, declared once in
its own frontmatter. `setup.sh global` installs whichever rules declare `alwaysApply: true`,
and only those. Every other rule is **opt-in**: it is never installed globally, and a project
adopts it by naming it in the `## standards` section of its own `.myflow/project.md`. A
per-project install (`cursor`, `claude-code`, `codex`, `all`) then renders every opt-in rule
that project named into a managed block in **that project's own `CLAUDE.md` and `AGENTS.md`**
— the same delimiters and the same inlining as the global block, so the rule text actually
loads for the harness instead of merely being referenced. The tree above is an illustrative
snapshot of today's set, not the definition; read the frontmatter to be sure.

**Skills** (loaded on demand):
- `/myflow-start` (+ `/myflow-start-fix`, `/myflow-start-done`), `/myflow-do` (+ `/myflow-do-manual-review`, `/myflow-do-done`), manual review (Gate B), `/myflow-do-fix` (Gate B/C/D fixes, + `/myflow-do-fix-manual-review`, `/myflow-do-fix-done`), `/myflow-manual-test` (Gate C, always asks whether to skip, + `/myflow-manual-test-done`), `/myflow-review` (+ `/myflow-review-done`), `/myflow-finish`, `/myflow-full`, `/myflow-status`, `/myflow-info` — OpenSpec + Superpowers with **manual review + manual test before commit**. The `*-done`/`*-manual-review` commands are pure state writes (via `myflow-state-advance`) — no verification, no git.
- `/opsx:*` — lighter OpenSpec-only variants
- `/opsx:explore` — thinking-partner mode (no code)

**myflow pipeline (default, twelve stages):**

```text
awaiting-proposal-review (Gate A) → proposal-done → awaiting-do-review (Gate B) → do-review-started → do-done →
[awaiting-fix-review → fix-review-started] → awaiting-manual-test (Gate C) → manual-test-done →
awaiting-pr-review (Gate D) → review-done → finished
```

`automerge` on `/myflow-review`/`/myflow-full` skips Gate D entirely (commits, pushes, and merges) and ends at `review-done` instead of `awaiting-pr-review`.

See `skills/myflow-contracts/pipeline.md` (the pipeline itself), `rules/myflow-manual-review.mdc` (the always-on stub that points at it), and `skills/README.md`.

---

## Installation

### Global install (recommended)

One install, every project. Run it once from this repo:

```bash
cd /path/to/agents-data
./setup.sh global
```

It symlinks straight out of this checkout, so editing a file here takes effect
immediately — no re-run needed except when a file is **added** or **removed**.

**One exception, and it is the highest-stakes one:** the always-on rules
(`rules/*.mdc` with `alwaysApply: true`) are *not* symlinked into Claude Code or Codex —
their text is **inlined** into the managed blocks in `~/.claude/CLAUDE.md` and
`~/.codex/AGENTS.md`. Editing `rules/lint-fix-priority.mdc` or
`rules/myflow-manual-review.mdc` has no effect on either harness until you re-run
`./setup.sh global`. Only the `~/.cursor/rules/` copy is a live symlink.

| Target | What lands there |
|--------|------------------|
| `~/.claude/skills/` | every directory in `skills/` (one symlink per skill) |
| `~/.cursor/skills/` | every directory in `skills/`; Cursor resolves `/myflow-*` commands through these |
| `~/.claude/commands/` | every file in `commands-claude/` — the `/myflow-*` Claude Code commands |
| `~/.cursor/commands/` | every file in `commands/` — the `/myflow-*` and `/opsx:*` Cursor commands |
| `~/.cursor/rules/` | whichever rules declare `alwaysApply: true` in their frontmatter, and only those |
| `~/.claude/CLAUDE.md` | a managed block, delimited by `<!-- myflow:begin -->` / `<!-- myflow:end -->`, containing the always-on rule text inlined — Claude Code's global rule layer |
| `~/.codex/AGENTS.md` | the **same** managed block, same delimiters, same inlined rule text — Codex's global rule layer. A global install writes this file even if you never ran a Codex-specific install |

Two things are deliberate:

- **Opt-in rules are excluded from everything on this page.** A rule that does not declare
  `alwaysApply: true` is never installed globally — not into `~/.cursor/rules/`, not into
  either managed block. The Kotlin backend standard is one: it applies to Kotlin backends,
  not to every project on the machine. Such a rule reaches a project only through that
  project's own `.myflow/project.md`, and only a **per-project** install renders it — see
  [Opt-in rules land in the project](#opt-in-rules-land-in-the-project) below.
- **The managed blocks are inlined, not referenced — in `~/.claude/CLAUDE.md` *and*
  `~/.codex/AGENTS.md`.** Neither Claude Code nor Codex reads `~/.cursor/rules/`, so for
  both harnesses the managed block is the only global rule layer. The two blocks are
  written by the same installer function with identical content. In each file, only the
  text between the delimiters is rewritten on re-install; your own notes outside them are
  never touched. If both delimiters are absent, a fresh block is appended. If they are
  present but not exactly one begin above one end, the installer stops and reports the
  offending line numbers rather than risk deleting content.

Per-project installs (`cursor`, `claude-code`, `codex`, `all`) below remain available for
projects that need a checked-in, project-local copy — and are the **only** way an opt-in
rule reaches a project. Prefer `global` for everything else.

---

## Opt-in rules land in the project

An opt-in rule is installed nowhere by path, deliberately: the Kotlin backend standard's
globs (`src/**/*.kt`, `**/*.kts`) would otherwise match every Compose Multiplatform or
IDE-plugin repo on the machine. But a rule installed nowhere is a rule no session ever
reads, which is why projects used to keep a pasted copy of the standard in their own
`CLAUDE.md` *and* `AGENTS.md` — two copies with nothing keeping either in step with the
rule they came from.

Every **per-project** install (`cursor`, `claude-code`, `codex`, `all`) closes that gap:

1. It reads `<project>/.myflow/project.md` and takes the `## standards` section. No file, or
   no such section, and it does nothing at all — silently.
2. It keeps the entries that resolve to the shared rule library: a **bare filename ending in
   `.mdc`** (`kotlin-backend-development-standard.mdc` → `<agents repo>/rules/<name>`). An
   entry containing a `/` is a project path, and any other bare filename is the project's own
   file — neither is a shared rule. See the resolution table in
   `skills/myflow-contracts/project-configuration.md`, which is canonical.
3. It drops any rule that is already `alwaysApply: true` — that one arrives through the
   global block, and rendering it again is the duplication this exists to remove.
4. It renders what remains into a managed block — same `<!-- myflow:begin -->` /
   `<!-- myflow:end -->` delimiters, same frontmatter stripping, same `.myflow.bak` and
   delimiter guards as the global block — in **both** `<project>/CLAUDE.md` and
   `<project>/AGENTS.md`.
5. An entry naming a rule that does not exist is reported by name and skipped. The rest of
   the install completes; the exit status still reports the skip.

`global` never does this: it installs no project files, and must not start writing into
whatever directory it was run from.

Only the text between the delimiters is rewritten, so your own notes around it survive. The
rendered block is generated content — edit `rules/<name>.mdc` in this repo and re-run the
per-project install; a hand-edit inside the delimiters is overwritten on the next run.

---

## Per-project installation per harness

### Cursor

Cursor reads rules from `.cursor/rules/` and skills from `.cursor/skills/`.
A global install already covers commands and always-on rules; use a project-local
install only when the project needs its own checked-in copy. It carries **skills and
commands only** — plus the same always-on rules, since `alwaysApply` is decided by the
rule, not by the install path. It additionally renders the project's **opt-in** rules into
the managed block in its `CLAUDE.md` and `AGENTS.md`; no install path ever places an opt-in
rule as a file. See [Opt-in rules land in the project](#opt-in-rules-land-in-the-project).

To install into a Cursor project, run:

```bash
cd /path/to/other-project
./path/to/agents-data/setup.sh cursor
```

This symlinks `agents-data/skills/` into `.cursor/skills/`, the always-on rules from `agents-data/rules/` into `.cursor/rules/`, and `agents-data/commands/` into `.cursor/commands/`. If the project's `.myflow/project.md` names any opt-in rule, that rule's text is also rendered into the managed block in the project's `CLAUDE.md` and `AGENTS.md`.

---

### Claude Code

Claude Code reads `CLAUDE.md` from the project root and discovers skills from
`.claude/skills/` (when Superpowers is installed).

**Step 1 — Install Superpowers** (general workflow skills: brainstorming, TDD, etc.)

In a Claude Code session inside the project:
```
/plugin install prime-radiant-inc/superpowers
```

**Step 2 — Add project instructions**

```bash
cp /path/to/agents-data/CLAUDE.md /path/to/project/CLAUDE.md
```

**Step 3 — Add project-specific skills and slash commands**

```bash
cd /path/to/project
./path/to/agents-data/setup.sh claude-code
# or manually:
mkdir -p .claude/skills .claude/commands
for d in /path/to/agents-data/skills/*/; do
  ln -sf "$d" .claude/skills/
done
for f in /path/to/agents-data/commands-claude/*.md; do
  ln -sf "$f" .claude/commands/
done
```

Step 3 also renders any opt-in rule the project named in its `.myflow/project.md` into the
managed block in `CLAUDE.md` (and `AGENTS.md`), so run it **after** step 2 — the block goes
into the file step 2 put there.

Without `.claude/commands/`, `/myflow-*` typed in the CLI will fail with "Unknown command" —
Claude Code only auto-discovers skills by their `SKILL.md` `name:` (e.g. `openspec-apply-superpowers`),
not by the `/myflow-*` alias. The `commands-claude/*.md` files are thin wrappers that map the
`/myflow-*` name to the underlying skill.

**Verify**: In a new Claude Code session, ask: *"What project skills do you have?"*
The agent should be able to list and describe the `/myflow-*` skills, and typing `/myflow-do`
should resolve without an "Unknown command" error.

---

### Codex (OpenAI)

Codex reads `AGENTS.md` from the project root and `~/.codex/AGENTS.md` globally, and
discovers skills natively when the Superpowers Codex plugin is installed. It reads neither
`~/.claude/CLAUDE.md` nor `~/.cursor/rules/`.

**`setup.sh global` writes `~/.codex/AGENTS.md`.** It inserts the same managed block it
writes into `~/.claude/CLAUDE.md` — the always-on rule text, between
`<!-- myflow:begin -->` / `<!-- myflow:end -->` — because that block is Codex's only global
rule layer. This happens on every `global` install, whether or not you also run a
Codex-specific install; your own content outside the delimiters is left alone. Note the
global install gives Codex the **rules** but not skills or commands (nothing is installed
under `~/.codex/`) — see "Where a Codex session gets its rules" in `AGENTS.md` for the
workaround.

**Step 1 — Install Superpowers for Codex**

The Superpowers Codex plugin is distributed from a separate fork repo.
Install it per the Superpowers README (look for the Codex install section).

Enable multi-agent support:
```toml
# ~/.codex/config.toml
[features]
multi_agent = true
```

**Step 2 — Add project instructions**

```bash
cp /path/to/agents-data/AGENTS.md /path/to/project/AGENTS.md
```

**Step 3 — Add project-specific skills**

```bash
cd /path/to/project
./path/to/agents-data/setup.sh codex
# or manually:
mkdir -p .codex/skills
for d in /path/to/agents-data/skills/*/; do
  ln -sf "$d" .codex/skills/
done
```

The `setup.sh codex` form (unlike the manual loop) also renders any opt-in rule the project
named in its `.myflow/project.md` into the managed block in `AGENTS.md` and `CLAUDE.md`.
Run it **after** step 2.

**Model note:** Codex has no per-skill/per-command model override mechanism — model is a session or profile-level setting (`~/.codex/config.toml`). Switch to a stronger model manually before invoking `openspec-propose-superpowers` (the `/myflow-start` equivalent); the rest of the pipeline is fine on your default.

---

### Gemini CLI

Gemini reads `GEMINI.md` from the project root and discovers skills via its extension
manifest's `contextFileName`.

**Step 1 — Install Superpowers for Gemini**

```bash
gemini extensions install prime-radiant-inc/superpowers
```

**Step 2 — Create GEMINI.md** with the rules inlined (same content as CLAUDE.md).

**Step 3 — Skills**: Gemini's Superpowers extension auto-discovers skills in its bundle.
For project-specific skills, symlink into `.gemini/skills/` if that path is recognized,
or inline the skill content into `GEMINI.md`.

---

## How skills work (no Superpowers installed)

If Superpowers is **not** installed, the agent can still use the project skills
by reading the `SKILL.md` file directly:

```
Read file: .claude/skills/openspec-propose-superpowers/SKILL.md
(follow the instructions in that file)
```

The `/myflow-*` skills internally reference Superpowers skills (brainstorming, TDD, etc.).
Without Superpowers those general skills won't auto-trigger, so the overall workflow is
degraded but the OpenSpec-specific steps still work.

---

## /myflow commands reference

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that stage is used automatically; if there are multiple, you're asked which.

**Model:** `/myflow-start` → Opus (enforced via frontmatter in Claude Code; switch manually in Cursor/Codex, which don't support per-command model selection yet). Every other command → Sonnet. `/myflow-full` runs as one session on Sonnet throughout, including Phase A — for brainstorming-heavy new work, run `/myflow-start` standalone first.

| Command | Skill file | What it does |
|---------|-----------|-------------|
| `/myflow-start <name>` | `openspec-propose-superpowers` | Brainstorm → design gate → artifacts → plan → publishes proposal artifact → `awaiting-proposal-review` |
| *(Gate A)* | User | Read the proposal artifact |
| `/myflow-start-fix <name>` | `openspec-propose-fix-superpowers` | Revise the proposal after Gate A feedback, republish artifact to the **same** URL, stay at `awaiting-proposal-review` |
| `/myflow-start-done <name>` | `myflow-state-advance` | *Pure state write* — confirms the proposal was reviewed → `proposal-done` |
| `/myflow-do <name>` | `openspec-apply-superpowers` | Worktree → validate plan → SDD+TDD → **strict review panel** (primary+Bugbot+Principles required; Security/Adversarial/extra lenses conditional) → **`git add` (staged; no commits)** → `awaiting-do-review` |
| *(manual review)* | User (Gate B) | Inspect **staged** diff in worktree IDE |
| `/myflow-do-manual-review <name>` | `myflow-state-advance` | *Pure state write* — confirms review is in progress → `do-review-started` |
| `/myflow-do-done <name>` | `myflow-state-advance` | *Pure state write* — confirms the diff was reviewed → `do-done` |
| `/myflow-do-fix <name>` | `openspec-apply-fix-superpowers` | Fix a Gate B/C/D finding — document in `proposal.md`/`tasks.md` (append) or a linked nested `<name>-fix-N` sub-change (your choice) → resume the **same** worktree → SDD+TDD → strict review panel (targeted re-run by default; full for Gate D origins or with `full-panel`) → staged; no commits (except Gate D, which commits+pushes to the PR branch). Records `originStage`; sets `awaiting-fix-review`. Loop at any of the four origins as many rounds as needed. |
| `/myflow-do-fix-manual-review <name>` | `myflow-state-advance` | *Pure state write* — confirms review of the fix is in progress → `fix-review-started` |
| `/myflow-do-fix-done <name>` | `myflow-state-advance` | *Pure state write* — confirms the fix was reviewed → returns to `originStage`, clears it |
| `/myflow-manual-test <name>` | `openspec-manual-test-superpowers` | Gate C — write `docs/manual-test/<name>.md` (run apps + checklist); always asks whether to skip (default No); reply with link only → `awaiting-manual-test` |
| *(manual test)* | User (Gate C) | Run the apps and check off items in the guide |
| `/myflow-manual-test-done <name>` | `myflow-state-advance` | *Pure state write* — confirms testing is complete → `manual-test-done` |
| `/myflow-review <name>` | `openspec-review-superpowers` | Verify Gate C (or `SKIPPED`) → test coverage check (routes gaps to `/myflow-do-fix`) → tests/linters → **commit + push + open PR** → `awaiting-pr-review` (or, with `automerge`, commits+pushes+**merges** → `review-done`, no PR) |
| *(PR review)* | User (Gate D) | Review and merge the PR on the forge — skipped entirely when `automerge` was used |
| `/myflow-review-done <name>` | `myflow-state-advance` | *Pure state write* — confirms the PR was reviewed (and merged) → `review-done` |
| `/myflow-finish <name>` | `openspec-archive-superpowers` | Verify the PR merged → delta sync → archive (also archives nested `<name>-fix-N` sub-changes together) |
| `/myflow-full <name>` | `openspec-full-cycle-superpowers` | Full cycle: Gate A + Gate B + Gate C + review, ending at Gate D (PR open, stop) — or at `review-done` with `automerge`; `/myflow-finish` is always a separate human-initiated step. Never auto-invokes any `*-done`/`*-manual-review` command. |
| `/myflow-status <name>` | — (read-only) | Stage report for open changes |
| `/myflow-info` | — (read-only) | Reads `skills/myflow-contracts/pipeline.md` and explains the pipeline |
| `/opsx:propose <name>` | `openspec-propose` | Lightweight: artifacts only, no Superpowers steps |
| `/opsx:apply <name>` | `openspec-apply-change` | Lightweight: implement tasks only |
| `/opsx:archive <name>` | `openspec-archive-change` | Lightweight: archive only |
| `/opsx:explore` | `openspec-explore` | Thinking-partner mode — no implementation |
| `/opsx:sync-specs <name>` | `openspec-sync-specs` | Sync delta specs to main specs |
| `/opsx:update <name>` | `openspec-update-change` | Revise planning artifacts, keep coherent |

**Flags:** `skip-propose`, `propose-only`, `skip-review` (skips Gate B only; the flag is the human's explicit opt-out at invocation time, so the cycle writes `do-done` with `gates.reviewed: false` rather than self-certifying a review nobody did), `skip-manual-test` (pre-answers the Gate C skip prompt with Yes, writing `manual-test-done` with `gates.tested: "skipped"` for the same reason; review still runs and still checks coverage), `automerge` (opt-in only, on `/myflow-review`/`/myflow-full` — commits, pushes, and merges, skipping Gate D, ending at `review-done`; never implied by any other flag), `full-panel` (forces every roster slot, including both extra principle lenses, over the whole-branch diff on every panel re-run, disabling the default targeted re-run; opt-in, never inferred), `commit-during-apply` (legacy per-task commits during apply)

All skills require the `openspec` CLI (`npm install -g openspec` or check project README).

---

## Making a change

**Edit the files here — this repo is the source of truth — then run `./setup.sh global`.**

```bash
cd /path/to/agents-data
$EDITOR skills/openspec-apply-superpowers/SKILL.md   # or any rule / command / skill
./setup.sh global
```

There is no importer, no sync hook, and no rsync from a project checkout. A project's
`.cursor/` or `.claude/` tree is an *install target* fed from here; never edit an
installed copy expecting it to travel back.

Because `setup.sh` installs **symlinks**, editing an existing skill or command takes effect
in the next session with no re-run. Re-run `./setup.sh global` when you **add** or **remove**
a skill, command, or rule — that is when the set of links changes.

**Rules are the exception — every rule's text is copied somewhere, not linked.** Treat every
edit to `rules/*.mdc` as requiring a re-install, and note that the two kinds re-install with
different commands:

| Rule kind | Where its text is copied | Re-install with |
|-----------|--------------------------|-----------------|
| `alwaysApply: true` (`lint-fix-priority`, `myflow-manual-review`) | the managed blocks in `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md` (plus a live symlink in `~/.cursor/rules/`) | `./setup.sh global` |
| opt-in (`kotlin-backend-development-standard`) | the managed block in the `CLAUDE.md` and `AGENTS.md` of **each project that named it** in `.myflow/project.md` | `./setup.sh <harness> /path/to/that/project`, once per adopting project |

An opt-in rule edited here therefore changes nothing for any project until that project's
install is re-run — and the projects that adopted it are listed nowhere but in their own
`.myflow/project.md` files, so a change to a widely-adopted opt-in rule needs a sweep.

`AGENTS.md` / `CLAUDE.md` carry their own myflow summary tables — update those by hand
when the pipeline description changes, and keep every command file in `commands/` and
`commands-claude/` consistent with the skill it points at. A command that contradicts its
skill is a defect, not a shorthand.
