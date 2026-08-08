# myflow skills

**myflow** = OpenSpec + Superpowers **Basic Workflow** bridge with a **three-state** machine: a
proposal gate, one combined review-and-test gate, and a finish command that integrates before it
archives.

The three commands, the state each ends in, and the gate that follows it are under
**States** (`myflow-contracts/pipeline.md`); the state diagram and the per-command stage table are
under **How the pipeline works** (`README.md`). This file copies none of them — it is read
on demand, beside the document it would be copying, exactly as `README.md` is. What does carry the
three-line digest is the layer that is loaded into a session before anything reads the pipeline: the
always-on rule, and a project's own `CLAUDE.md` / `AGENTS.md`.

Each command ends in the state named after it. **The human gate is a property of the state**, not
a separate stage — so no command exists whose only job is to record that a review happened.

**Every command is re-entrant.** Re-run `/myflow-start` to revise the proposal, `/myflow-do` to fix
something. A fix never moves the state.

**No command takes a flag.** The only argument is the optional change name.

See also: `myflow-manual-review.mdc` — authored at `rules/myflow-manual-review.mdc` in this repo,
installed by `setup.sh global` to `~/.cursor/rules/` and inlined into the managed block in
`~/.claude/CLAUDE.md`. It is a **stub**: the pipeline itself lives in
`skills/myflow-contracts/pipeline.md`, loaded on demand by every `/myflow-*` command.

`<name>` is **optional** on every `/myflow-*` command below — if omitted, the sole active
(non-archived) change relevant to that state is used automatically; if there are multiple, you're
asked which.

**Model:** `/myflow-start` → Opus (enforced via frontmatter in Claude Code; manual switch
elsewhere). Every other command's **session** → Sonnet, and **every review-panel reviewer runs on
the panel's model — Sonnet by default** — a change being free to record its own panel model. But
the **implementer subagents `/myflow-do` dispatches run on Opus** — or the harness's strongest
model — named explicitly at dispatch. Frontmatter sets the session's model and cannot set a
subagent's, so that one is enforced by the dispatch and recorded in the SDD ledger.
See "Model policy" in `myflow-contracts/pipeline.md`, which is canonical.
See "Model policy" in `skills/myflow-contracts/pipeline.md`.

## Superpowers Basic Workflow map

| Step | Skill | myflow command |
|------|-------|----------------|
| **1** | brainstorming | `/myflow-start` |
| **3** | writing-plans | `/myflow-start` (enrich), `/myflow-do` (validate) |
| **2** | using-git-worktrees | `/myflow-do` |
| **4** | subagent-driven-development | `/myflow-do` |
| **5** | test-driven-development | `/myflow-do`, every implementer dispatch |
| **6** | requesting-code-review + the panel | `/myflow-do` |
| **8** | verification-before-completion | `/myflow-do` |

`finishing-a-development-branch` is **never** invoked — integration is `/myflow-finish`'s job.

## Command map

| Command | Skill | What it does |
|---------|-------|--------------|
| `/myflow-start <name>` | `myflow-start` | Turns an idea into an approved plan: brainstorming behind a design gate, the OpenSpec artifacts, and a published proposal artifact. Ends at `STARTED`; re-run to revise, republishing to the **same** URL. |
| *(gate)* | you | Read the proposal artifact |
| `/myflow-do <name>` | `myflow-do` | Implements that plan under SDD + TDD behind the **review panel**, leaving a manual test guide beside a staged diff. Ends at `IN_PROGRESS`; re-run to fix, which never moves the state. |
| *(gate)* | you | Review the staged diff **and** run the apps against the guide |
| `/myflow-finish <name>` | `myflow-finish` | Integrates the branch on its first run, asking how to land it (PR by default, merge, or manual); on its second, once merged, archives the change and removes what the pipeline created. |
| `/myflow-status [name]` | `myflow-status` | Read-only report of where every open change is |
| `/opsx:explore` | `openspec-explore` | Thinking-partner mode — no implementation, no state |

Each row says what a command is *for*. Its stages, in order, are stated once under
**Level 1 — the stages of each command** (`README.md`) and are deliberately not
repeated here.

## Skills

```
skills/
├── myflow-start/      ← /myflow-start
├── myflow-do/         ← /myflow-do  (carries the review-panel prompts + engineering-principles.md)
├── myflow-finish/     ← /myflow-finish
├── myflow-status/     ← /myflow-status (read-only)
├── myflow-contracts/  ← on-demand contracts; `pipeline.md` is canonical for the state machine
└── openspec-explore/  ← /opsx:explore
```

`myflow-contracts/` holds `pipeline.md` (the state machine — load it first for any `/myflow-*`
command), plus `state-file.md`, `project-configuration.md`,
`jira-integration.md`, `plan-provenance.md`, `build-green.md` and `workspace-isolation.md`. Load the
one file a step needs, not the directory.

## What runs when

- **Verification** — the project's lint and test commands run inside `/myflow-do`, and nothing
  re-runs them later. `/myflow-finish` has **no** verification gate: no tests, no linters, no
  coverage check before a PR or a merge.
- **The review panel** — three required slots (primary, Bugbot, Principles) plus conditional slots
  selected by the trigger table in `skills/myflow-do/SKILL.md`, which is canonical for it. **Every
  slot runs on the panel's model — Sonnet by default** — regardless of the parent model.
- **Commits** — only `/myflow-finish`, and `/myflow-do` when a PR is already open (a staged-only
  fix would be invisible on it).

All skills require the `openspec` CLI (`npm install -g openspec`).
