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
│   ├── never-touch-production.mdc     ← no route to a production system, ever (always-on)
│   ├── no-direct-pushes-to-main.mdc   ← land on the integration branch, promote by PR (always-on)
│   ├── be-brief.mdc                   ← answer at the length the question needs; prose only (always-on)
│   ├── build-the-simplest-thing.mdc   ← complexity is opt-in (always-on)
│   ├── dependency-versions.mdc        ← look up the current stable version before adding one (always-on)
│   ├── design-mockups-are-specs.mdc   ← build a mockup exactly as drawn (always-on)
│   ├── context7.mdc                   ← fetch library docs through Context7, not from memory (always-on)
│   ├── dispatch-carries-the-baseline.mdc  ← every subagent dispatch carries the agent-baseline pointer (always-on)
│   ├── agent-baseline.md              ← NOT a rule: the file a dispatched subagent is told to read, listing every rule above and pointing at its installed full text
│   └── kotlin-backend-development-standard.mdc  ← opt-in: named in a project's `.myflow/project.md`, rendered into that project's CLAUDE.md + AGENTS.md
├── hooks/
│   └── enforce-agent-baseline.py      ← PreToolUse hook: denies a subagent dispatch whose prompt omits the agent-baseline pointer
├── scripts/
│   ├── check-vocabulary.sh            ← guards the pipeline vocabulary used across these files
│   └── test-setup.sh                  ← regression harness for setup.sh (sandboxed HOME under /tmp)
├── commands/                          ← Cursor slash commands (myflow + opsx:explore)
├── commands-claude/                   ← Claude Code slash commands (myflow only)
└── skills/                            ← OpenSpec / /myflow skills
    ├── README.md                      ← myflow command map
    ├── myflow-start/                  ← /myflow-start
    ├── myflow-do/                     ← /myflow-do; carries the review-panel prompts + engineering-principles.md
    ├── myflow-finish/                 ← /myflow-finish (integrate, then archive + clean up)
    ├── myflow-fast/                   ← /myflow-fast (composite: brainstorm+implement, then integrate+archive, chained)
    ├── myflow-status/                 ← read-only state report for open changes
    ├── myflow-contracts/              ← on-demand contracts; pipeline.md is canonical for the state machine
    └── openspec-explore/              ← /opsx:explore — thinking-partner mode, touches no state
```

**Rules** — whether a rule is always-on is a property of the rule itself, declared once in
its own frontmatter. The tree above is an illustrative
snapshot of today's set, not the definition; read the frontmatter to be sure.

**Skills** (loaded on demand): `/myflow-start`, `/myflow-do`, `/myflow-finish`, `/myflow-fast`
(the composite that chains the other three), plus the read-only `/myflow-status`, and
`/opsx:explore` for thinking-partner mode.

**myflow pipeline — three states, three commands.**

Each command ends in the state named after it, and **the human gate is a property of the state** —
which is why no command exists whose only job is to record that a review happened. `/myflow-do`
emits both the staged diff and the run instructions, so reviewing and testing are one sitting.
Every command is re-entrant, and a fix never moves the state.

`/myflow-finish` runs **twice**: once to integrate the branch (open a PR by default, merge and
push, or leave it to you), and again once the branch is merged, to sync delta specs, archive and
commit onto a `chore/archive-<name>` branch, remove the worktrees, and — after self-review — push
that branch and open its pull request. **Run 2 never pushes the base branch**; run 1's merge-and-push
route still does, when you choose it. It runs **no** tests, linters or coverage check — that
happened during `/myflow-do`.

See **How the pipeline works** (`README.md`) below for the state diagram and the per-command stage table, plus `rules/myflow-manual-review.mdc` (the always-on stub that points at the pipeline) and `skills/README.md`.

---

## How the pipeline works

```mermaid
stateDiagram-v2
    [*] --> STARTED: /myflow-start
    STARTED --> STARTED: /myflow-start (revise the proposal)
    STARTED --> IN_PROGRESS: /myflow-do
    IN_PROGRESS --> IN_PROGRESS: /myflow-do (fix — never moves the state)
    IN_PROGRESS --> IN_PROGRESS: /myflow-finish (run 1 — integrate)
    IN_PROGRESS --> FINISHED: /myflow-finish (run 2 — after the merge)
    [*] --> IN_PROGRESS: /myflow-fast (brainstorm + implement)
    IN_PROGRESS --> IN_PROGRESS: /myflow-fast (fix — argument present)
    IN_PROGRESS --> FINISHED: /myflow-fast (merge+push route)
    FINISHED --> [*]
```

### Level 1 — the stages of each command

Two tables. The first is the stage vocabulary itself — every documented stage, across the five
commands this pipeline has (three pipeline commands, one composite command, one read-only, exactly
as **Command surface** (`skills/myflow-contracts/pipeline.md`) names them). The second is the human
gate that *follows* each command's run — a property of the state the command ends in, never a stage
of its own, so it is kept out of the first table rather than repeated per stage.

**Stages, in order.** One row per stage: a stable **key** a mark carries and the store groups by, a
human-readable **name** that may be reworded without splitting recorded history, and every command
that runs it. A name marked ▸ hides substructure and is expanded at level 2 below.

A key is namespaced by the command that *defines* the stage, never by the command that merely runs
it. `/myflow-fast` chains `/myflow-start`, `/myflow-do` and `/myflow-finish` without minting a
single stage name of its own: its allowed stage set is the **union** of the three, expressed here by
listing `/myflow-fast` in the Commands column of every row those three define, so a fast run is
directly comparable, stage for stage, against the equivalent start→do→finish sequence. See design.md
under kan-172 for the rejected alternatives. `/myflow-status` marks no stages at all and contributes
no rows.

| Key | Name | Commands |
|-----|------|----------|
| `start.resolve-change` | Resolve the change | `/myflow-start`, `/myflow-fast` |
| `start.ask-options` | Ask planning effort, models & panel roster (creating run only) | `/myflow-start`, `/myflow-fast` |
| `start.brainstorm` | Brainstorm ▸ | `/myflow-start`, `/myflow-fast` |
| `start.design-approval` | Design approval | `/myflow-start`, `/myflow-fast` |
| `start.create-artifacts` | Create the OpenSpec artifacts | `/myflow-start`, `/myflow-fast` |
| `start.writing-plans` | Writing-plans ▸ | `/myflow-start`, `/myflow-fast` |
| `start.publish-proposal` | Publish the proposal artifact | `/myflow-start`, `/myflow-fast` |
| `start.write-started` | Write `STARTED` | `/myflow-start`, `/myflow-fast` |
| `do.state-gate` | State gate | `/myflow-do`, `/myflow-fast` |
| `do.load-context` | Load context and validate the plan | `/myflow-do`, `/myflow-fast` |
| `do.isolate-workspace` | Isolate the workspace (first run only) | `/myflow-do`, `/myflow-fast` |
| `do.document-fix` | Document the fix (re-runs only) | `/myflow-do`, `/myflow-fast` |
| `do.sdd-tdd` | SDD + TDD per task ▸ | `/myflow-do`, `/myflow-fast` |
| `do.review-panel` | The review panel ▸ | `/myflow-do`, `/myflow-fast` |
| `do.run-instructions` | Resolve the run instructions | `/myflow-do`, `/myflow-fast` |
| `do.workspace-export` | Validate and export workspace isolation | `/myflow-do`, `/myflow-fast` |
| `do.lint-and-test` | Run the project's lint and test commands | `/myflow-do`, `/myflow-fast` |
| `do.stage-diff` | Stage, excluding the planning paths | `/myflow-do`, `/myflow-fast` |
| `do.write-in-progress` | Write `IN_PROGRESS` | `/myflow-do`, `/myflow-fast` |
| `finish.preflight` | Preflight verdict (decides run 1 vs run 2) ▸ | `/myflow-finish`, `/myflow-fast` |
| `finish.unfinished-work-gate` | Unfinished-work gate (run 1) ▸ | `/myflow-finish`, `/myflow-fast` |
| `finish.landing-question` | The landing question (run 1) | `/myflow-finish`, `/myflow-fast` |
| `finish.preserve-sessions` | Preserve the session records (run 1) | `/myflow-finish`, `/myflow-fast` |
| `finish.commit-two` | Two commits, implementation first (run 1) | `/myflow-finish`, `/myflow-fast` |
| `finish.landing-routes` | The landing routes (run 1) ▸ | `/myflow-finish`, `/myflow-fast` |
| `finish.write-in-progress` | Write `IN_PROGRESS` (run 1) | `/myflow-finish`, `/myflow-fast` |
| `finish.move-in-review` | Move the issue to In Review (run 1) | `/myflow-finish`, `/myflow-fast` |
| `finish.verify-merge` | Verify the merge (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.sync-archive` | Position the checkout, sync delta specs and archive (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.commit-archive` | Commit the archive (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.cleanup` | Cleanup (run 2) ▸ | `/myflow-finish`, `/myflow-fast` |
| `finish.verify-cleanup` | Verify the cleanup (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.write-finished` | Write `FINISHED` (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.self-review` | Self-review (run 2) | `/myflow-finish`, `/myflow-fast` |
| `finish.push-archive` | Push the archive branch and open its PR (run 2) | `/myflow-finish`, `/myflow-fast` |

**Gate after it.** `/myflow-finish` is one command with two runs, so its row carries both, labelled:
the preflight verdict picks which one a given invocation performs, and the run is never a command of
its own.

| Command | Gate after it |
|---------|---------------|
| `/myflow-start` | you read the proposal artifact |
| `/myflow-do` | you review the staged diff **and** run the apps |
| `/myflow-finish` | after run 1, you wait for the branch to merge; after run 2, nothing — the state is terminal |
| `/myflow-fast` | creating run or fix: you review the staged diff **and** run the apps; open PR or manual: you wait for the branch to merge (or finish your manual steps); merge-and-push: nothing — the state is terminal |
| `/myflow-status` | — |

`/myflow-finish` run 2's sequence ends with `push-archive`; the row before it, `self-review`,
carries no ▸ either: its procedure is not expanded at level 2 below because it is canonical under
**Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`), step 9. The
requirement to change first when that procedure changes is
**Requirement: Self-review runs only after FINISHED is written**
(`openspec/specs/myflow-self-review/spec.md`).

### Level 2 — the stages that hide substructure

Each expansion below states the **structure** — the shape that changes only when the pipeline
changes — and cites the file that owns any tuned threshold rather than restating it: a threshold
copied here is a copy that can go wrong silently the next time the owning file changes.

#### Brainstorm — `/myflow-start`

superpowers:brainstorming runs its checklist in full and ends with the operator approving the
design, which is a hard gate: nothing is created under `openspec/changes/` until that approval
lands. The approved design is saved under `docs/superpowers/specs/` and becomes the source for the
change's OpenSpec design artifact — adapted, never duplicated into a conflicting second design.

The stage iterates rather than passing once. After every planning-stage exchange — a round of
clarifying questions, the approval of a design section, the operator's review of the written spec —
one convergence test asks whether the command now holds a question its inputs do not answer, and
while it does, another round opens or is offered. The stage closes only the way any pipeline stage
does — **Stage exit — never the command's own judgment** (`skills/myflow-contracts/pipeline.md`).

The threshold, the two prompts, the bounded exception, and why their opposite recommendations are
both honest are **Convergence** (`skills/myflow-start/SKILL.md`).

The planning level recorded on the creating run sizes the thinking *inside* this gate and never the
gate itself. The three levels and which of them is the default are owned by **Planning effort**
(`skills/myflow-contracts/state-file.md`).

#### Writing-plans — `/myflow-start`

superpowers:writing-plans enriches `tasks.md` from a checkbox scaffold into a plan whose every item
carries exact paths, verification commands and no placeholders — the unit `/myflow-do` dispatches
one implementer against. Its self-review — spec coverage, placeholder scan, type consistency — runs
before the stage finishes.

Every fenced block and every numeric claim in a planning artifact carries a provenance tag,
`verified:<how>` or `unverified:`, which `scripts/check-plan-provenance.sh` makes mechanical rather
than a habit. An unverifiable snippet is tagged and **kept**: a plan without the snippet is worse
than a plan carrying a labelled guess.

A revision round re-enters at this stage and republishes the proposal artifact to the **same** URL
rather than minting a second one.

#### SDD + TDD per task — `/myflow-do`

One implementer dispatch per checkbox in `tasks.md`, or per tightly coupled group, in plan order.
Every dispatch carries the same four required blocks — the no-commits boundary,
superpowers:test-driven-development as a required sub-skill, `engineering-principles.md` as required
reading, and the plan-provenance rule above — and names its model explicitly rather than inheriting
the parent's.

The task's diff is written to a file and the reviewer is given that path, never a commit range,
because nothing is committed at this stage. A checkbox is marked `[x]` only after its task passes
spec **and** quality review; a blocked task pauses and reports rather than guessing.

Which model a dispatch runs on, and the rule that every dispatch records it, are **Model policy**
(`skills/myflow-contracts/model-policy.md`).

#### The review panel — `/myflow-do`

A change records one of three review panel rosters — `light` *(default)*, `standard` or `full` — and
every preset dispatches exactly three required slots, whichever is recorded; `full` reproduces the
roster in force before presets existed. What each preset means, and the roster table itself, are
canonical under **5. The review panel** (`skills/myflow-do/SKILL.md`). Four
further slots stay conditional under every preset — Security, Adversarial and the two extra principle
lenses, B for simplicity and state, C for robustness and ops — selected from what the diff touches. Each
selected slot is a **separate** subagent with its own prompt, in every affected worktree; two slots are
never merged into one.

Every slot runs on the panel's model — Sonnet by default — except the two dispatched by
`subagent_type`. Bugbot and Security Review carry their own agent definitions and take no model
override. There is no parent-model inheritance and no economy tier: the panel's cost does not
depend on the model the operator happens to be running.

No handoff happens while any finding is open, at any severity — a minor finding blocks exactly as a
critical one does. Re-runs are targeted by default and escalate to the full roster automatically,
without asking; escalation widens the panel's **breadth** — more lenses — and never its model. When
a finding survives its last fix round the run hands back to the operator, one finding at a time,
with named options.

The tuned values are cited rather than copied: which diff sizes and which touched areas select a
conditional slot is **Optional slot selection** in `skills/myflow-do/SKILL.md`; the conditions that
force a full re-run in place of a targeted one are **Panel re-runs** in the same file.

#### The preflight verdict — `/myflow-finish`

`scripts/check-finish-preflight.sh` decides which run happens, from three signals in a fixed order,
taken once per worktree in the resolved set — never a raw read of the state file's `worktrees` map,
per **Resolving a change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`). It prints exactly
one verdict line and exits 0 whenever it reached a verdict; a missing verdict line is not a verdict,
and neither is a worktree it cannot read. `RUN1` integrates, `RUN2` archives, and `REFUSE` stops the
run and asks the operator rather than guessing. Run 2 proceeds only when every worktree in the
resolved set returns `RUN2` — and a resolved set that comes back empty is never read as that,
per the same section.

The three signals and why their order is load-bearing are **Finish contract**
(`skills/myflow-contracts/finish-contract.md`).

#### The unfinished-work gate — `/myflow-finish` run 1

Runs **before** the landing question and before any git action, once per worktree in the resolved
set — see **Resolving a change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`).
`scripts/check-unfinished-work.sh` returns `CLEAR` — go straight to the question, with no extra
prompt — or `OUTSTANDING`, which shows the breakdown and offers **exactly three** courses, with
**Stop** marked as the recommendation. There is no fourth course, and none that hands back to
`/myflow-do` inline.

The ordering is the point: an operator asked how to land a branch, and only then told it carries
unfinished work, has already answered a question about a branch they believed was complete. What
was integrated over is written into the planning commit's message and into the handoff, so the
record outlives the session.

Each course and what run 1 then does are **Run 1 — the branch is not merged**
(`skills/myflow-contracts/finish-contract.md`).

#### The landing routes — `/myflow-finish` run 1

The operator is asked once, before any git action, how the branch should land: open a pull request
*(default)*, merge and push, or handle it manually. The run then completes without asking again,
and the answer is never remembered between runs.

All three routes do the same two things first, in this order: preserve the session records out of
the gitignored worktree into the repository, then commit in **two** commits — implementation first,
planning artifacts second. The linked issue moves to In Review on every route, including the manual
one.

The route table is **Run 1 — the branch is not merged**
(`skills/myflow-contracts/finish-contract.md`); the guarded two-commit chain every route uses is
**Git boundaries** (`skills/myflow-contracts/git-boundaries.md`).

#### Cleanup — `/myflow-finish` run 2

Every removal is *remove-or-move if present*, which is what makes run 2 re-entrant: a step whose
artifact is already gone is a success rather than an error, so a re-run after the operator clears a
leftover repeats the verification and nothing else.

The removals are verified rather than assumed. `scripts/check-cleanup-complete.sh` runs once per
repository, **after** all of them: `COMPLETE:` allows the `FINISHED` write, `LEFTOVER:` names what
remains and leaves the change at `IN_PROGRESS`, and a non-zero exit carrying no verdict line is
treated exactly as `LEFTOVER`.

What is removed, when, and on what condition is **Temporary artifacts registry**
(`skills/myflow-contracts/artifacts-registry.md`) — the one place a cleanup rule is stated. The procedure for
the rows it removes is **Worktree cleanup** (`skills/myflow-contracts/finish-contract.md`).

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

**One exception, and it is the highest-stakes one:** what the managed blocks in
`~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md` carry is **rendered text, not a symlink**.
Editing an always-on rule has no effect on either harness's injected prompt until you
re-run `./setup.sh global`. The `~/.cursor/rules/` and `~/.claude/rules/` copies are live
symlinks and need no re-run.

| Target | What lands there |
|--------|------------------|
| `~/.claude/skills/` | every directory in `skills/` (one symlink per skill) |
| `~/.cursor/skills/` | every directory in `skills/`; Cursor resolves `/myflow-*` commands through these |
| `~/.claude/commands/` | every file in `commands-claude/` — the `/myflow-*` Claude Code commands |
| `~/.cursor/commands/` | every file in `commands/` — the `/myflow-*` and `/opsx:*` Cursor commands |
| `~/.cursor/rules/` | whichever rules declare `alwaysApply: true` in their frontmatter, and only those |
| `~/.claude/rules/` | the same always-on rules, symlinked as `<name>.md` — their **full text**, which the managed block's `Full rule:` pointers name. Plus `agent-baseline.md`, the file a dispatched subagent is told to read |
| `~/.claude/hooks/` | every file in `hooks/`. Installed, never registered: `settings.json` is yours, so the installer prints the snippet and leaves the paste to you |
| `~/.claude/CLAUDE.md` | a managed block, delimited by `<!-- myflow:begin -->` / `<!-- myflow:end -->`, containing each always-on rule's **core** and a pointer to its full text — Claude Code's global rule layer |
| `~/.codex/AGENTS.md` | the **same** managed block, same delimiters, same rendered text — Codex's global rule layer. A global install writes this file even if you never ran a Codex-specific install |

### Core and full text are one source

A rule may wrap part of its body in `<!-- core -->` / `<!-- /core -->`. Globally, the
managed block gets that core plus `Full rule: ~/.claude/rules/<name>.md`, and the pointer
resolves to a symlink back to the very `.mdc` the core was rendered from — one file, no
second copy to go stale. A rule with no markers is inlined whole, exactly as before.

Core extraction is global-only. A project's opted-in standards render into that project's
own `CLAUDE.md` with the full body and no pointer, because `~/.claude/rules/` never holds
an opt-in rule. The markers themselves are stripped from every render.

Subagents inherit none of this. A dispatched agent reads no `CLAUDE.md`, so every dispatch
has to carry a pointer to `~/.claude/rules/agent-baseline.md` — which lists every rule in a
line, points at these same paths, and instructs the agent to pass the same pointer on, so
the rules survive to any dispatch depth. `hooks/enforce-agent-baseline.py` denies a
dispatch whose prompt omits it.

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

This symlinks `agents-data/skills/` into `.cursor/skills/`, the always-on rules from `agents-data/rules/` into `.cursor/rules/`, and `agents-data/commands/` into `.cursor/commands/`.

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
Claude Code only auto-discovers skills by their `SKILL.md` `name:` (e.g. `myflow-do`),
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
Codex-specific install; your own content outside the delimiters is left alone.

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

**Model note:** Codex has no per-skill/per-command model override mechanism — model is a session or profile-level setting (`~/.codex/config.toml`). Switch to a stronger model manually before invoking `myflow-start` (the `/myflow-start` equivalent); the rest of the pipeline is fine on your default.

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
Read file: .claude/skills/myflow-start/SKILL.md
(follow the instructions in that file)
```

The `/myflow-*` skills internally reference Superpowers skills (brainstorming, TDD, etc.).
Without Superpowers those general skills won't auto-trigger, so the overall workflow is
degraded but the OpenSpec-specific steps still work.

---

## /myflow commands reference

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name; anything else is reported rather than ignored.

**Model:** See "Model policy" in `skills/myflow-contracts/model-policy.md`.

| Command | Skill | What it does |
|---------|-------|-------------|
| `/myflow-start <name>` | `myflow-start` | Turns an idea into an approved plan: brainstorming behind a design-approval gate, the OpenSpec artifacts, and a published proposal artifact. Ends at `STARTED`; re-run to revise, republishing to the **same** URL. |
| *(gate)* | You | Read the proposal artifact |
| `/myflow-do <name>` | `myflow-do` | Implements that plan under SDD + TDD behind the **review panel**, sized by the recorded `reviewPanelRoster` — `light` *(default)*, `standard` or `full`, each dispatching exactly three required slots, with Security, Adversarial and extra lenses staying conditional under every preset — which hands off only at **zero open findings at any severity**. Ends at `IN_PROGRESS`; re-run to fix, which never moves the state. |
| *(gate)* | You | Review the staged diff **and** run the apps |
| `/myflow-finish <name>` | `myflow-finish` | Integrates the branch on its first run — after checking each worktree for unfinished work, it asks how to land it: open a PR (default), merge and push, or handle it manually — and, on its second run once the branch has merged, archives the change and removes what the pipeline created. Runs no tests, linters or coverage check. |
| `/myflow-fast <name>` | `myflow-fast` | Composite command: chains `/myflow-start`'s brainstorming (fully interactive, unchanged) directly into `/myflow-do`'s implementation and review panel with no gate in between, and chains `/myflow-finish`'s run 1 into run 2 when the chosen landing route is merge and push. Accepts no state (creates the change) or `IN_PROGRESS` — an argument at `IN_PROGRESS` is fix instructions, a bare invocation asks how to land the branch. Publishes no proposal artifact. A creating or fix run ends at `IN_PROGRESS` — a fix leaves the state unchanged; a bare invocation ends at `IN_PROGRESS` or `FINISHED`, depending on the route chosen. Re-run to fix or to integrate. |
| `/myflow-status [name]` | `myflow-status` | Read-only state report for open changes |
| `/opsx:explore` | `openspec-explore` | Thinking-partner mode — no implementation, no state |

Each row above says what a command is *for*. Its stages, in order — and the human gate that follows
each — are stated once under
**Level 1 — the stages of each command** (`README.md`) above and are deliberately not repeated
here — one ordered list rather than two competing ones in the same file.

The branch's merge status alone decides which `/myflow-finish` run happens, so a PR you merged on
the forge and a merge it performed itself are indistinguishable to it — which is correct.

All skills require the `openspec` CLI (`npm install -g openspec` or check project README).

---

## Making a change

**Edit the files here — this repo is the source of truth — then run `./setup.sh global`.**

```bash
cd /path/to/agents-data
$EDITOR skills/myflow-do/SKILL.md   # or any rule / command / skill
./setup.sh global
```

There is no importer, no sync hook, and no rsync from a project checkout. A project's
`.cursor/` or `.claude/` tree is an *install target* fed from here; never edit an
installed copy expecting it to travel back.

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
