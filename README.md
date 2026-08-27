# agents-data

Portable, project-agnostic agent configuration — the **flow** pipeline (spectre +
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
│   ├── flow-manual-review.mdc       ← flow trigger + contract pointers (always-on stub, installed globally)
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
│   └── kotlin-backend-development-standard.mdc  ← opt-in: named in a project's `.flow/project.md`, rendered into that project's CLAUDE.md + AGENTS.md
├── hooks/
│   └── enforce-agent-baseline.py      ← PreToolUse hook: denies a subagent dispatch whose prompt omits the agent-baseline pointer
├── scripts/
│   ├── check-vocabulary.sh            ← guards the pipeline vocabulary used across these files
│   └── test-setup.sh                  ← regression harness for setup.sh (sandboxed HOME under /tmp)
├── commands/                          ← Cursor slash commands (/flow, /flow-status, /flow-research, /flow-settings)
├── commands-claude/                   ← Claude Code slash commands (the same four)
├── skills/                            ← spectre / /flow skills
│   ├── README.md                      ← flow command map
│   ├── flow/                          ← /flow — brainstorm, implement behind the review panel, integrate and archive, one command
│   ├── flow-status/                   ← read-only state report for open changes
│   ├── flow-research/                 ← /flow-research — thinking-partner mode, stages research notes, touches no state
│   ├── flow-settings/                 ← /flow-settings — global model/reviewer defaults
│   └── flow-contracts/                ← on-demand contracts; pipeline.md is canonical for the state machine
├── spectre/                           ← this repository's own artifact tree: specs/ and changes/
└── openspec/                          ← frozen at the 2026-08-25 cutover: the record of how this repository got here, never written to again
```

**Rules** — whether a rule is always-on is a property of the rule itself, declared once in
its own frontmatter. The tree above is an illustrative
snapshot of today's set, not the definition; read the frontmatter to be sure.

**Skills** (loaded on demand): `/flow` — the single-command pipeline — plus the read-only
`/flow-status`, `/flow-research` for thinking-partner mode, and `/flow-settings` for global
model/reviewer defaults.

**flow pipeline — three states.**

`/flow` drives the full pipeline as one command. Each phase ends in the state named after it, and
**the human gate is a property of the state** — which is why no command exists whose only job is
to record that a review happened. Every invocation is re-entrant, and a fix never moves the state.

`/flow` ends a creating or fix run at `IN_PROGRESS` with both the staged diff and the run
instructions, so reviewing and testing are one sitting. Once `IN_PROGRESS`, a bare invocation asks
how the branch should land — open a PR by default, merge and push, or leave it to you — and stops
there unless merge-and-push was chosen, in which case the same invocation continues straight into
archive once the merge lands: sync and commit onto a `chore/archive-<name>` branch, remove the
worktrees, and — after self-review — push that branch and open its pull request. **Archive never
pushes the base branch**; the merge-and-push route still does, when you choose it. It runs **no**
tests, linters or coverage check — that happened during implementation.

See **How the pipeline works** (`README.md`) below for the state diagram and the per-command stage table, plus `rules/flow-manual-review.mdc` (the always-on stub that points at the pipeline) and `skills/README.md`.

---

## How the pipeline works

```mermaid
stateDiagram-v2
    [*] --> STARTED: /flow (kickoff)
    STARTED --> STARTED: /flow (resume before implementation)
    STARTED --> IN_PROGRESS: /flow (same invocation, into implementation)
    IN_PROGRESS --> IN_PROGRESS: /flow (fix — argument present, never moves the state)
    IN_PROGRESS --> IN_PROGRESS: /flow (bare — integrate, open PR or manual route)
    IN_PROGRESS --> FINISHED: /flow (bare — integrate, merge+push route, chained into archive)
    FINISHED --> [*]
```

### Level 1 — the stages of each command

Two tables. The first is the stage vocabulary itself — every documented stage, across the
commands this pipeline has (three pipeline commands, one composite command, and `/flow` — the
single command that replaces all four — plus two read-only/no-state commands, exactly
as **Command surface** (`skills/flow-contracts/pipeline.md`) names them). The second is the human
gate that *follows* each command's run — a property of the state the command ends in, never a stage
of its own, so it is kept out of the first table rather than repeated per stage.

**Stages, in order.** One row per stage: a stable **key** a mark carries and the store groups by, a
human-readable **name** that may be reworded without splitting recorded history, and every command
that runs it. A name marked ▸ hides substructure and is expanded at level 2 below.

A key is namespaced by the command that *defines* the stage, never by the command that merely runs
it — a distinction the previous, four-command surface needed and `/flow`, the single command left,
does not: every row below is both defined and run by `/flow`. See design.md under kan-172 for the
rejected alternatives to that older namespacing. `/flow-status` marks no stages at all and
contributes no rows.

`/flow` mints its own namespace, `flow.*`, rather than reusing `start.*`/`do.*`/`finish.*` — several
of its stages are not the old stage unchanged (`flow.verify` merges two, `flow.landing-routes`
absorbs `finish.move-in-review`, `flow.kickoff` is new), so a reused key would sometimes mean
something different than its history records (design.md's `flow-rename-content-split`, and the
stage-keys note in `skills/flow/SKILL.md`). Which phase file under `skills/flow/` marks each `flow.*`
key is **Stage keys** (`skills/flow/SKILL.md`), cited rather than repeated as a column here.

| Key | Name | Commands |
|-----|------|----------|
| `flow.kickoff` | Kickoff — write `STARTED` | `/flow` |
| `flow.resolve-change` | Resolve the change | `/flow` |
| `flow.brainstorm` | Brainstorm ▸ | `/flow` |
| `flow.design-approval` | Design approval | `/flow` |
| `flow.create-artifacts` | Create the spectre artifacts | `/flow` |
| `flow.writing-plans` | Writing-plans ▸ | `/flow` |
| `flow.load-context` | Load context and validate the plan | `/flow` |
| `flow.isolate-workspace` | Isolate the workspace (first run only) | `/flow` |
| `flow.document-fix` | Document the fix (re-runs only) | `/flow` |
| `flow.sdd-tdd` | SDD + TDD per task ▸ | `/flow` |
| `flow.review-panel` | The review panel ▸ | `/flow` |
| `flow.verify` | Verify: workspace isolation, lint and test | `/flow` |
| `flow.visual-verify` | Visual verification | `/flow` |
| `flow.stage-diff` | Stage, excluding the planning paths | `/flow` |
| `flow.run-instructions` | Resolve the run instructions | `/flow` |
| `flow.write-in-progress` | Write `IN_PROGRESS` | `/flow` |
| `flow.preflight` | Preflight verdict (decides run 1 vs run 2) ▸ | `/flow` |
| `flow.unfinished-work-gate` | Unfinished-work gate (run 1) ▸ | `/flow` |
| `flow.landing-question` | The landing question (run 1) | `/flow` |
| `flow.preserve-sessions` | Preserve the session records (run 1) | `/flow` |
| `flow.commit-two` | Two commits, implementation first (run 1) | `/flow` |
| `flow.landing-routes` | The landing routes, including moving the issue to In Review (run 1) ▸ | `/flow` |
| `flow.verify-merge` | Verify the merge (run 2) | `/flow` |
| `flow.sync-archive` | Position the checkout and archive (run 2) | `/flow` |
| `flow.commit-archive` | Commit the archive (run 2) | `/flow` |
| `flow.cleanup` | Cleanup (run 2) ▸ | `/flow` |
| `flow.verify-cleanup` | Verify the cleanup (run 2) | `/flow` |
| `flow.write-finished` | Write `FINISHED` (run 2) | `/flow` |
| `flow.self-review` | Self-review (run 2) | `/flow` |
| `flow.push-archive` | Push the archive branch and open its PR (run 2) | `/flow` |

`/flow` has no `start.ask-options`/`start.publish-proposal` equivalent (design.md's
`ask-options-removed`, `publish-proposal-removed`) and no `finish.write-in-progress` (run 1)
equivalent — resolved as folded away rather than kept, per design.md's open question
`write-in-progress-fold`, since that stage wrote `IN_PROGRESS` → `IN_PROGRESS` with no state change
of its own.

**Gate after it.** `/flow` is one command with several runs — creating, resuming, fix, integrate,
archive — so its row states the gate for each.

| Command | Gate after it |
|---------|---------------|
| `/flow` | creating run or fix: you review the staged diff **and** run the apps; integrate with open PR or manual: you wait for the branch to merge (or finish your manual steps); integrate with merge-and-push, chained into archive: nothing — the state is terminal |
| `/flow-status` | — |
| `/flow-research` | — |

`/flow`'s run-2 sequence ends with `flow.push-archive`. The row before it, `flow.self-review`,
carries no ▸ either: its procedure is not expanded at level 2 below because it is canonical under
**Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract.md`), step 9 — which is
also the file to change when that procedure changes, since the requirements layer that once sat
above it is frozen with the rest of the `openspec/` tree at the spectre cutover.

### Level 2 — the stages that hide substructure

Each expansion below states the **structure** — the shape that changes only when the pipeline
changes — and cites the file that owns any tuned threshold rather than restating it: a threshold
copied here is a copy that can go wrong silently the next time the owning file changes.

#### Brainstorm — `/myflow-start`

superpowers:brainstorming runs its checklist in full and ends with the operator approving the
design, which is a hard gate: nothing is created under `spectre/changes/` until that approval
lands. The approved design is saved under `docs/superpowers/specs/` and becomes the source for the
change's `design.md` artifact — adapted, never duplicated into a conflicting second design.

The stage iterates rather than passing once. After every planning-stage exchange — a round of
clarifying questions, the approval of a design section, the operator's review of the written spec —
one convergence test asks whether the command now holds a question its inputs do not answer, and
while it does, another round opens or is offered. The stage closes only the way any pipeline stage
does — **Stage exit — never the command's own judgment** (`skills/flow-contracts/pipeline.md`).

The threshold, the two prompts, the bounded exception, and why their opposite recommendations are
both honest are **Convergence** (`skills/flow/brainstorm.md`).

The planning level recorded on the creating run sizes the thinking *inside* this gate and never the
gate itself. The three levels and which of them is the default are owned by **Planning effort**
(`skills/flow-contracts/state-file.md`).

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
(`skills/flow-contracts/model-policy.md`).

#### The review panel — `/myflow-do`

A change records one of three review panel rosters — `light` *(default)*, `standard` or `full` — and
every preset dispatches exactly three required slots, whichever is recorded; `full` reproduces the
roster in force before presets existed. This preset/roster system was later retired in favor of a
fixed 3-slot panel — see **The review panel — `/flow`** below and design.md's
`review-panel-fixed-3`; **The three required slots** and **The two on-demand slots**
(`skills/flow/review-panel.md`) are canonical for the panel's current shape. Four
further slots stay conditional under every preset — Security, Adversarial and two extra principle
slots — selected from what the diff touches. Each
selected slot is a **separate** subagent with its own prompt, in every affected worktree; two slots are
never merged into one.

Every slot runs on the panel's model — Sonnet by default — except the two dispatched by
`subagent_type`. Bugbot and Security Review carry their own agent definitions and take no model
override. There is no parent-model inheritance and no economy tier: the panel's cost does not
depend on the model the operator happens to be running.

No handoff happens while any finding is open, at any severity — a minor finding blocks exactly as a
critical one does. Re-runs are targeted by default and escalate to the full roster automatically,
without asking; escalation widens the panel's **breadth** — more slots — and never its model. When
a finding survives its last fix round the run hands back to the operator, one finding at a time,
with named options.

The tuned values are cited rather than copied: which diff sizes and which touched areas select a
conditional slot was **Optional slot selection** in the retired `myflow-do/SKILL.md` — replaced by
**The two on-demand slots**, now explicit-only (`skills/flow/review-panel.md`); the conditions that
force a full re-run in place of a targeted one are **Panel re-runs** in the same file.

#### The preflight verdict — `/myflow-finish`

`scripts/check-finish-preflight.sh` decides which run happens, from three signals in a fixed order,
taken once per worktree in the resolved set — never a raw read of the state file's `worktrees` map,
per **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`). It prints exactly
one verdict line and exits 0 whenever it reached a verdict; a missing verdict line is not a verdict,
and neither is a worktree it cannot read. `RUN1` integrates, `RUN2` archives, and `REFUSE` stops the
run and asks the operator rather than guessing. Run 2 proceeds only when every worktree in the
resolved set returns `RUN2` — and a resolved set that comes back empty is never read as that,
per the same section.

The three signals and why their order is load-bearing are **Finish contract**
(`skills/flow-contracts/finish-contract.md`).

#### The unfinished-work gate — `/myflow-finish` run 1

Runs **before** the landing question and before any git action, once per worktree in the resolved
set — see **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`).
`scripts/check-unfinished-work.sh` returns `CLEAR` — go straight to the question, with no extra
prompt — or `OUTSTANDING`, which shows the breakdown and offers **exactly three** courses, with
**Stop** marked as the recommendation. There is no fourth course, and none that hands back to
`/myflow-do` inline.

The ordering is the point: an operator asked how to land a branch, and only then told it carries
unfinished work, has already answered a question about a branch they believed was complete. What
was integrated over is written into the planning commit's message and into the handoff, so the
record outlives the session.

Each course and what run 1 then does are **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract.md`).

#### The landing routes — `/myflow-finish` run 1

The operator is asked once, before any git action, how the branch should land: open a pull request
*(default)*, merge and push, or handle it manually. The run then completes without asking again,
and the answer is never remembered between runs.

All three routes do the same two things first, in this order: preserve the session records out of
the gitignored worktree into the repository, then commit in **two** commits — implementation first,
planning artifacts second. The linked issue moves to In Review on every route, including the manual
one.

The route table is **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract.md`); the guarded two-commit chain every route uses is
**Git boundaries** (`skills/flow-contracts/git-boundaries.md`).

#### Cleanup — `/myflow-finish` run 2

Every removal is *remove-or-move if present*, which is what makes run 2 re-entrant: a step whose
artifact is already gone is a success rather than an error, so a re-run after the operator clears a
leftover repeats the verification and nothing else.

The removals are verified rather than assumed. `scripts/check-cleanup-complete.sh` runs once per
repository, **after** all of them: `COMPLETE:` allows the `FINISHED` write, `LEFTOVER:` names what
remains and leaves the change at `IN_PROGRESS`, and a non-zero exit carrying no verdict line is
treated exactly as `LEFTOVER`.

What is removed, when, and on what condition is **Temporary artifacts registry**
(`skills/flow-contracts/artifacts-registry.md`) — the one place a cleanup rule is stated. The procedure for
the rows it removes is **Worktree cleanup** (`skills/flow-contracts/finish-contract.md`).

#### Brainstorm — `/flow`

superpowers:brainstorming runs its checklist in full and ends with the operator approving the
design, which is a hard gate: nothing is created under `spectre/changes/` until that approval
lands. The approved design is saved under `docs/superpowers/specs/` and becomes the source for the
change's `design.md` artifact — adapted, never duplicated into a conflicting second design. Before
the checklist opens, the stage checks `docs/superpowers/research/` for a staged note matching this
topic (per `/flow-research`'s staging behaviour below) and, if found, seeds the round from it
without ever skipping straight to artifact-writing (design.md's `flow-research-staging`).

The stage iterates rather than passing once, the same way `/myflow-start`'s does — see
**Convergence** (`skills/flow/brainstorm.md`) for the threshold, the two prompts, and the bounded
exception, which apply unchanged.

No planning-effort, model, or review-panel-roster question runs on a creating run — the three
questions design.md's `ask-options-removed` retired. The panel is fixed at 3 required slots on every
run instead; see **The review panel** below.

#### Writing-plans — `/flow`

superpowers:writing-plans enriches `tasks.md` the same way it does for `/myflow-start` — see
**Writing-plans — `/myflow-start`** above for the provenance-tag rule, which applies unchanged.
`/flow` publishes no proposal artifact (design.md's `publish-proposal-removed`), so a revision round
re-enters at this stage without republishing anything.

#### SDD + TDD per task — `/flow`

Same dispatch shape as `/myflow-do`'s — see **SDD + TDD per task — `/myflow-do`** above. Which model
a dispatch runs on is **Model resolution** in `skills/flow/SKILL.md` — one default for implementer,
panel, and panel-fix roles, read from the settings store rather than asked per change (design.md's
`model-default-sonnet`, `settings-scope`); the per-harness enforcement notes in **Model policy**
(`skills/flow-contracts/model-policy.md`) still apply, but that file's three-role table does not.

#### The review panel — `/flow`

Every run dispatches exactly **3 required slots**, unconditionally — Primary, Principles, and
Code review (low) — no roster, no preset, no diff-size or touched-area trigger
(design.md's `review-panel-fixed-3`). Two further slots, Bugbot and Security, are **on-demand only**:
included solely when the operator explicitly names one, checked at the start of the panel stage and
again at every fix round — never by an automatic trigger. Each included slot is a **separate**
subagent with its own prompt, in every affected worktree; two slots are never merged into one. The
slot table itself is canonical under **The three required slots** and **The two on-demand slots**
(`skills/flow/review-panel.md`).

Every slot runs on `DEFAULT_MODEL` — the settings-store default, or this run's session-instruction
override — except the two dispatched by `subagent_type`. Bugbot and Security Review carry their own
agent definitions and take no model override. There is no parent-model inheritance and no economy
tier: the panel's cost does not depend on the model the operator happens to be running.

No handoff happens while any finding is open, at any severity — a minor finding blocks exactly as a
critical one does. A fix round re-checks for an explicit Bugbot/Security instruction before it
dispatches; when a finding survives its last fix round the run hands back to the operator, one
finding at a time, with named options.

#### The preflight verdict — `/flow` integrate

Same `scripts/check-finish-preflight.sh` decision `/myflow-finish` uses — see **The preflight
verdict — `/myflow-finish`** above, which applies unchanged.

#### The unfinished-work gate — `/flow` integrate, run 1

Same `scripts/check-unfinished-work.sh` gate `/myflow-finish` run 1 uses — see **The unfinished-work
gate — `/myflow-finish` run 1** above, which applies unchanged except that "hands back to
`/myflow-do` inline" reads as "hands back to the implement phase inline."

#### The landing routes — `/flow` integrate, run 1

Same three routes `/myflow-finish` run 1 offers — see **The landing routes — `/myflow-finish` run
1** above — except that moving the linked issue to In Review is folded into this stage as an
unconditional sub-step rather than a row of its own (design.md's `move-in-review-fold`).

#### Cleanup — `/flow` archive, run 2

Same re-entrant remove-or-move-if-present cleanup `/myflow-finish` run 2 uses — see **Cleanup —
`/myflow-finish` run 2** above, which applies unchanged.

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
| `~/.cursor/skills/` | every directory in `skills/`; Cursor resolves `/flow*` commands through these |
| `~/.claude/commands/` | every file in `commands-claude/` — the `/flow*` Claude Code commands |
| `~/.cursor/commands/` | every file in `commands/` — the `/flow*` Cursor commands |
| `~/.cursor/rules/` | whichever rules declare `alwaysApply: true` in their frontmatter, and only those |
| `~/.claude/rules/` | the same always-on rules, symlinked as `<name>.md` — their **full text**, which the managed block's `Full rule:` pointers name. Plus `agent-baseline.md`, the file a dispatched subagent is told to read |
| `~/.claude/hooks/` | every file in `hooks/`. Installed, never registered: `settings.json` is yours, so the installer prints the snippet and leaves the paste to you |
| `~/.claude/CLAUDE.md` | a managed block, delimited by `<!-- myflow:begin -->` / `<!-- myflow:end -->`, containing each always-on rule's **core** and a pointer to its full text — Claude Code's global rule layer |
| `~/.codex/AGENTS.md` | the **same** managed block, same delimiters, same rendered text — Codex's global rule layer. A global install writes this file even if you never ran a Codex-specific install |

Both the `skills/` and `commands*/` install steps discover their targets by walking the tree —
`skills/*/` and `commands*/*.md` — rather than from a fixed list, so a new skill or command
directory (like `flow/`, `flow-status/`, `flow-research/`, `flow-settings/`) is installed the next
time you run `./setup.sh global` with no change to `setup.sh` itself.

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
  project's own `.flow/project.md`, and only a **per-project** install renders it — see
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

1. It reads `<project>/.flow/project.md` and takes the `## standards` section. No file, or
   no such section, and it does nothing at all — silently.
2. It keeps the entries that resolve to the shared rule library: a **bare filename ending in
   `.mdc`** (`kotlin-backend-development-standard.mdc` → `<agents repo>/rules/<name>`). An
   entry containing a `/` is a project path, and any other bare filename is the project's own
   file — neither is a shared rule. See the resolution table in
   `skills/flow-contracts/project-configuration.md`, which is canonical.
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

Step 3 also renders any opt-in rule the project named in its `.flow/project.md` into the
managed block in `CLAUDE.md` (and `AGENTS.md`), so run it **after** step 2 — the block goes
into the file step 2 put there.

Without `.claude/commands/`, `/flow` typed in the CLI will fail with "Unknown
command" — Claude Code only auto-discovers skills by their `SKILL.md` `name:` (e.g. `flow`),
not by the slash-command alias. The `commands-claude/*.md` files are thin wrappers that map the
command name to the underlying skill.

**Verify**: In a new Claude Code session, ask: *"What project skills do you have?"*
The agent should be able to list and describe the `flow`/`flow-status`/`flow-research`/
`flow-settings` skills, and typing `/flow` should resolve without an "Unknown command" error.

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
named in its `.flow/project.md` into the managed block in `AGENTS.md` and `CLAUDE.md`.
Run it **after** step 2.

**Model note:** Codex has no per-skill/per-command model override mechanism — model is a session or profile-level setting (`~/.codex/config.toml`). Switch to a stronger model manually before invoking `flow` (the `/flow` equivalent) on a creating run; the rest of the pipeline is fine on your default.

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
Read file: .claude/skills/flow/SKILL.md
(follow the instructions in that file)
```

The `flow` skill (and the retiring `flow-*` skills) internally reference Superpowers skills
(brainstorming, TDD, etc.). Without Superpowers those general skills won't auto-trigger, so the
overall workflow is degraded but the spectre-specific steps still work.

---

## /flow commands reference

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name; anything else is reported rather than ignored.

**Model:** `/flow` reads its model from `skills/flow/SKILL.md`'s own **Model resolution**; see "Model policy" in `skills/flow-contracts/model-policy.md` for the per-harness enforcement notes that still apply.

| Command | Skill | What it does |
|---------|-------|-------------|
| `/flow <name>` | `flow` | Single-command pipeline. No state: creates the change, writes `STARTED`, and — same invocation — runs brainstorming (unchanged, fully interactive) then implementation behind the fixed 3-slot review panel, ending at `IN_PROGRESS`. Asks no planning-effort, model, or review-panel-roster question and publishes no proposal artifact. `IN_PROGRESS` with an argument: fix run, state unchanged. `IN_PROGRESS` bare: asks how to land the branch — open PR (default), merge and push, or manual — then, on merge-and-push, continues in the same invocation through archive to `FINISHED`; open PR and manual stop and hand off. Runs no tests, linters or coverage check outside implementation's own verify stage. |
| *(gate)* | You | Creating run or fix: review the staged diff **and** run the apps. Integrate with open PR or manual: wait for the branch to merge (or finish your manual steps). Merge-and-push: nothing — the state is terminal. |
| `/flow-status [name]` | `flow-status` | Read-only state report for open changes |
| `/flow-research` | `flow-research` | Thinking-partner mode — no implementation, no state; stages research notes under `docs/superpowers/research/` for `/flow`'s brainstorming to seed from |
| `/flow-settings` | `flow-settings` | Reads/writes the global model and reviewer defaults every `/flow` run reads from |

Each row above says what a command is *for*. Its stages, in order — and the human gate that follows
each — are stated once under
**Level 1 — the stages of each command** (`README.md`) above and are deliberately not repeated
here — one ordered list rather than two competing ones in the same file.

The branch's merge status alone decides which run `/flow`'s integrate/archive phase performs, so a
PR you merged on the forge and a merge it performed itself are indistinguishable to it — which is
correct.

Every skill above but `flow-research` requires the `spectre` CLI
(`go install github.com/tweety53/spectre/cmd/spectre@latest`, with `$(go env GOPATH)/bin` on your
`PATH`). `flow-research` needs none — reading a spectre tree is reading markdown.

---

## Cutting over from `myflow` to `flow`

The rename from `myflow` to `flow` changed names this repository owns **and** names that live on the
machine outside it — an installed binary, a running daemon, a Docker container, a database, a
per-project configuration directory. Merging the branch changes only the first kind. The five steps
below change the rest.

**They run in this order. None may be reordered, and steps 2 and 4 must not be separated.**

### 1. Merge

Nothing on the machine has changed yet. The installed skills are symlinks into the main checkout, so
they flip to the renamed text the moment the merge lands — which is why every step below is now
overdue rather than optional.

### 2 and 4 are one operation — read this before starting

The daemon stamps every response with a trust header so a look-alike server on its port cannot be
mistaken for the real store. That header is renamed: `Myflow-Daemon` becomes `Flow-Daemon`. **A
renamed CLI therefore cannot trust the old daemon, and the old CLI cannot trust a renamed one.**

This degrades safely — an untrusted response is a store failure, and every CLI path falls back to the
on-disk journal and exits 0, so nothing breaks loudly. What it does *not* do is read back what it
just wrote. **Do not run `/flow` between step 2 and step 4.**

### 2. Build and install the renamed CLI

```bash
cd stats && make build
cp bin/flow "$HOME/.local/bin/flow"
```

Leave the old `~/.local/bin/myflow` in place for now; step 5 removes it.

### 3. Rename the project configuration directory, in every consuming project

```bash
git -C <project> mv .myflow .flow
```

Three projects on this machine carry one today: this repository, `~/Projects/gymie`, and
`~/Projects/spectre-e2e`. Until a project is renamed, anything resolving its configuration refuses
with an actionable error naming the directory and the exact `git mv` — it never falls back to the old
path and never reports "no configuration found".

### 4. Rename the container and the database

**This is the only step that stops the dev workspace's service and storage, and it is yours to run.**
`ALTER DATABASE` requires zero live connections, so the daemon must be down first.

**Read the three traps below before running anything.** Each one was hit during the first real
cutover, and two of them can lose data.

**Trap 1 — `docker compose down` will NOT stop the old container.** The compose file's service key
changed from `myflow-postgres` to `flow-postgres`, so compose no longer recognises the running  <!-- vocab-guard:allow -->
container as one of its services: it reports it as an *orphan*, leaves it running, and fails to
remove the network. Stop it by name, or pass `--remove-orphans`.

**Trap 2 — the volume rename orphans your data.** The compose file now declares `flow-pgdata`, so a
plain `up` creates a **new, empty** volume and leaves every row behind in `stats_myflow-pgdata`. The
data must be copied across first — and copied from a **stopped** source, because `cp -a` over a live
PostgreSQL data directory can catch it mid-write. Combined with trap 1, the obvious sequence copies a
running database, which is how the first attempt produced an untrustworthy volume.

**Trap 3 — the role cannot rename itself.** `ALTER ROLE myflow RENAME TO flow` fails with
`session user cannot be renamed`, and this cluster has exactly one role, so there is no second
superuser to run it from. One must be created for the purpose, and the rename **invalidates the
role's password**, which has to be reset.

Take a backup first. It is the only copy that does not live in a Docker volume:

```bash
docker exec myflow-postgres pg_dumpall -U myflow > ~/flow-pre-cutover.sql  # vocab-guard:allow
```

Stop the daemon and the container — by name, not through compose:

```bash
kill $(lsof -tiTCP:4173 -sTCP:LISTEN)     # whatever holds the daemon's port
docker stop myflow-postgres  # vocab-guard:allow
```

Rename the database and the role, using a temporary superuser for the role:

```bash
docker start myflow-postgres  # vocab-guard:allow
docker exec myflow-postgres psql -U myflow -d postgres -c 'ALTER DATABASE myflow RENAME TO flow;'  # vocab-guard:allow
docker exec myflow-postgres psql -U myflow -d postgres -c "CREATE ROLE cutover_tmp SUPERUSER LOGIN PASSWORD 'tmp';"  # vocab-guard:allow
docker exec -e PGPASSWORD=tmp myflow-postgres psql -U cutover_tmp -d postgres -c 'ALTER ROLE myflow RENAME TO flow;'  # vocab-guard:allow
docker exec -e PGPASSWORD=tmp myflow-postgres psql -U cutover_tmp -d postgres -c "ALTER ROLE flow WITH PASSWORD 'flow';"  # vocab-guard:allow
docker exec myflow-postgres psql -U flow -d postgres -c 'DROP ROLE cutover_tmp;'  # vocab-guard:allow
```

The temporary role must be dropped from a `flow` session, not its own — the same rule that blocked
the rename in the first place.

Now copy the volume with the source stopped, and bring the renamed stack up:

```bash
docker stop myflow-postgres  # vocab-guard:allow
docker volume create stats_flow-pgdata
docker run --rm -v stats_myflow-pgdata:/from -v stats_flow-pgdata:/to alpine sh -c 'cd /from && cp -a . /to'
docker compose -f stats/docker-compose.yml up -d --remove-orphans
```

**Verify the copy before trusting it, and before deleting anything:**

```bash
docker exec flow-postgres psql -U flow -d flow -tAc 'select count(*) from stage_runs;'
docker exec flow-postgres psql -U flow -d flow -tAc 'select count(*) from changes;'
```

Both must match what the old container reported. **Keep `stats_myflow-pgdata` and the dump until they
do** — that volume is the live data, and the dump is the only copy outside Docker. Remove the old
volume only once the new stack has been serving correctly for a while.

Move the fallback state root in the same step:

```bash
mv ~/Agents/myflow ~/Agents/flow
```

That directory holds records and journal entries a failed write left behind, and the daemon replays
the journal when it can reach the database again. **Re-measure before you move it** — a non-empty
journal means a write never reached the store:

```bash
ls ~/Agents/myflow/state/*/*.json | wc -l
find ~/Agents/myflow/state -name '*.journal' -size +0
```

At the time this runbook was written that directory held 56 records and two journal files, both
empty — so the move was a formality rather than a rescue. That was true of this machine at that
moment, not of the mechanism.
<!-- measured: ls and find over ~/Agents/myflow/state at the time of writing @ branch spectre/kan-289-reproduce-rather-than-read-in-slot-template -->

**Until this step completes, no new `/flow` run can isolate a workspace.** `scripts/workspace.sh`
addresses the container by name, and the branch ends naming `flow-postgres` — which does not exist
until you rename it here.

### Running the Go suite before you finish

The test helpers' fallback DSN names the **declared** role and database — `flow:flow@.../flow`, what
a fresh compose stack creates. Against a container you have not renamed yet that host is unreachable,
and these suites **skip rather than fail**: `internal/store` alone reports `ok` while running four
tests and skipping 155. A green package that ran almost nothing is worse than a red one, so until
step 4 is done, set both overrides:

```bash
export FLOW_STATS_ADMIN_DSN="postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"  # vocab-guard:allow
export FLOW_STATS_DSN="postgres://myflow:myflow@localhost:5433/myflow?sslmode=disable"  # vocab-guard:allow
cd stats && go test ./... -count=1
```

Both are needed, and they do different jobs: `FLOW_STATS_ADMIN_DSN` drives the per-test database
creation and every connection derived from it, `FLOW_STATS_DSN` the compose-stack reachability check.
With both set the suite runs with zero skips. After step 4, neither is needed.

### 5. Reinstall the launchd agent, then remove the old binary

```bash
launchctl unload ~/Library/LaunchAgents/com.tweety53.myflowd.plist  # vocab-guard:allow
rm ~/Library/LaunchAgents/com.tweety53.myflowd.plist  # vocab-guard:allow
cp stats/launchd/com.tweety53.flowd.plist ~/Library/LaunchAgents/
launchctl load ~/Library/LaunchAgents/com.tweety53.flowd.plist
rm "$HOME/.local/bin/myflow"
```

### What needs no step from you

The managed block in `~/.claude/CLAUDE.md` still carries the old `<!-- myflow:begin -->` delimiters.
The installer migrates them itself on its next run: it rewrites the two marker lines in place and
then refreshes the block as usual, so the file keeps one block rather than gaining a second orphaned
one. Your existing `CLAUDE.md.myflow.bak` is left exactly as it is — it holds your genuine
pre-install content, and a fresh backup would only copy a file the installer already manages.

### What is deliberately left carrying the old name

Existing Jira issues keep the labels they were filed with, so the board holds both taxonomies and a
label search must match either form. The `stage_runs` rows recording `/myflow-do`, `/myflow-start`,
`/myflow-finish` and `/myflow-fast` keep those values, because they record which command actually
ran. The frozen `openspec/` tree and the historical records under `docs/` keep theirs for the same
reason.

## Making a change

**Edit the files here — this repo is the source of truth — then run `./setup.sh global`.**

```bash
cd /path/to/agents-data
$EDITOR skills/flow/SKILL.md   # or any rule / command / skill
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
| `alwaysApply: true` (`lint-fix-priority`, `flow-manual-review`) | the managed blocks in `~/.claude/CLAUDE.md` and `~/.codex/AGENTS.md` (plus a live symlink in `~/.cursor/rules/`) | `./setup.sh global` |
| opt-in (`kotlin-backend-development-standard`) | the managed block in the `CLAUDE.md` and `AGENTS.md` of **each project that named it** in `.flow/project.md` | `./setup.sh <harness> /path/to/that/project`, once per adopting project |

An opt-in rule edited here therefore changes nothing for any project until that project's
install is re-run — and the projects that adopted it are listed nowhere but in their own
`.flow/project.md` files, so a change to a widely-adopted opt-in rule needs a sweep.

`AGENTS.md` / `CLAUDE.md` carry their own flow summary tables — update those by hand
when the pipeline description changes, and keep every command file in `commands/` and
`commands-claude/` consistent with the skill it points at. A command that contradicts its
skill is a defect, not a shorthand.
