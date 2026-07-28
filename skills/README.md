# myflow skills

**myflow** = OpenSpec + Superpowers **Basic Workflow** bridge with a **three-state** machine: a
proposal gate, one combined review-and-test gate, and a finish command that integrates before it
archives.

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (it integrates on its first run)
```

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
elsewhere). Every other command → Sonnet, and **every review-panel reviewer runs on Sonnet** too.
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
| `/myflow-start <name>` | `myflow-start` | Brainstorm → design gate → OpenSpec artifacts → writing-plans enriched tasks → publish the proposal artifact → `STARTED`. Re-run to revise, republishing to the **same** URL. |
| *(gate)* | you | Read the proposal artifact |
| `/myflow-do <name>` | `myflow-do` | Worktree → validate plan → SDD + TDD → **review panel** → **manual test guide** → lint + tests → `git add` → `IN_PROGRESS`. Re-run to fix; a fix never moves the state. |
| *(gate)* | you | Review the staged diff **and** run the apps against the guide |
| `/myflow-finish <name>` | `myflow-finish` | **Run 1:** ask how to land the branch (PR by default, merge, or manual), commit, push, take that route, stop. **Run 2** (once merged): verify, sync specs, archive, **commit + push the archive**, **remove the worktrees**, → `FINISHED`. |
| `/myflow-status [name]` | `myflow-status` | Read-only report of where every open change is |
| `/myflow-info` | `myflow-info` | Read-only explanation, read from the installed contract |
| `/opsx:explore` | `openspec-explore` | Thinking-partner mode — no implementation, no state |

## Skills

```
skills/
├── myflow-start/      ← /myflow-start
├── myflow-do/         ← /myflow-do  (carries the review-panel prompts + engineering-principles.md)
├── myflow-finish/     ← /myflow-finish
├── myflow-status/     ← /myflow-status (read-only)
├── myflow-info/       ← /myflow-info   (read-only)
├── myflow-contracts/  ← on-demand contracts; `pipeline.md` is canonical for the state machine
└── openspec-explore/  ← /opsx:explore
```

`myflow-contracts/` holds `pipeline.md` (the state machine — load it first for any `/myflow-*`
command), plus `state-file.md`, `state-self-heal.md`, `project-configuration.md` and
`jira-integration.md`. Load the one file a step needs, not the directory.

## What runs when

- **Verification** — the project's lint and test commands run inside `/myflow-do`, and nothing
  re-runs them later. `/myflow-finish` has **no** verification gate: no tests, no linters, no
  coverage check before a PR or a merge.
- **The review panel** — three required slots (primary, Bugbot, Principles) plus conditional slots
  selected by the trigger table in `skills/myflow-do/SKILL.md`, which is canonical for it. **Every
  slot runs on Sonnet**, regardless of the parent model.
- **Commits** — only `/myflow-finish`, and `/myflow-do` when a PR is already open (a staged-only
  fix would be invisible on it).

All skills require the `openspec` CLI (`npm install -g openspec`).
