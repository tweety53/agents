# Agent Instructions (Codex)

This file is the active instruction set for Codex sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Where a Codex session gets its rules

Codex reads `AGENTS.md` — the project's own, plus `~/.codex/AGENTS.md` globally. It does
not read `~/.claude/CLAUDE.md` or `~/.cursor/rules/`.

`setup.sh global` writes the always-on rules into a managed block in `~/.codex/AGENTS.md`,
delimited by `<!-- myflow:begin -->` / `<!-- myflow:end -->`, using the same ordering check
and self-poisoning guard as the Claude Code block. So a Codex session gets
`myflow-manual-review.mdc` and `lint-fix-priority.mdc` globally, and opt-in rules (such as
the Kotlin backend standard) are deliberately excluded — a project activates those by
naming them in its `.myflow/project.md` `## standards` section.

Project-mode `setup.sh codex` installs skills and this file but no rules, symmetrically
with `claude-code`. Run `setup.sh global` for the rule layer.

### What a global install actually leaves a Codex session with

Be precise about this, because the two halves are asymmetric:

| | After `setup.sh global` |
|---|---|
| **Rules** | ✅ present — the managed block in `~/.codex/AGENTS.md` carries `myflow-manual-review.mdc` and `lint-fix-priority.mdc` |
| **Skills** | ✅ present — `install_global` links every directory in `skills/` into `~/.codex/skills/`, alongside `~/.claude/skills/` and `~/.cursor/skills/` |
| **Commands** | ❌ absent — `~/.claude/commands/` and `~/.cursor/commands/` only. There is no `~/.codex/commands/` layer. |

So a Codex session has the rules and the skills, but no slash-command layer: typing
`/myflow-do` will not resolve, even though the skill it delegates to is installed.

**What to do today:** invoke the skill directly instead of through a command — read its
`SKILL.md` out of the globally installed tree and follow it, e.g.

```
Read file: ~/.codex/skills/myflow-do/SKILL.md
(then follow the instructions in that file)
```

That path is a symlink into this checkout, so the content is always current. Each command
file in `commands-claude/` is a thin wrapper naming exactly one skill plus its accepted
states, so reading the skill directly loses nothing but the shorthand.

Do **not** work around the missing command layer with a per-project `setup.sh codex` install:
`specs/myflow-global-install/spec.md` requires that projects not retain their own copies
of myflow skills, commands, or rules once the global install exists.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.codex/AGENTS.md` from `agents/rules/lint-fix-priority.mdc`. It is not restated here — one
source of truth, so the policy cannot drift between the global copy and this file.

What is project-specific is which commands it means. Record them in this project's
`.myflow/project.md`, then name them here — for example:

```bash
<auto-fix command>     # the formatter, run first
<check command>        # must pass before you claim the work is done
```

Pre-approved suppressions and documented deviations live in `CONTRIBUTING.md`. Do not
expand that list without user approval.

---

### Project-specific standards

<!-- Replace this section with the coding standard this project actually follows:
     module layout, layering rules, naming conventions, framework constraints, and the
     test command to run before claiming completion.

     This template ships generic on purpose. `setup.sh` copies it into any project root
     that lacks an `AGENTS.md`, so a standard hardcoded to one stack would be wrong in
     every other project. A standard meant to apply across *many* projects belongs in
     `agents/rules/` as an opt-in rule instead, activated per project by naming it in
     `.myflow/project.md`'s `## standards` section — `kotlin-backend-development-standard.mdc`
     is the worked example of that pattern. -->

This project has not declared one yet. Until it does, follow the language's published
conventions and the patterns already present in the surrounding code.

---

## Project Skills (OpenSpec / /myflow workflow)

These skills live in `skills/` next to this file (or in `.codex/skills/` if installed there).
To invoke a skill: **read its `SKILL.md` file** then follow the instructions within.

All skills require the `openspec` CLI to be installed.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/myflow-start/` | `/myflow-start` | Proposal: brainstorming behind a design gate, the OpenSpec artifacts, and a published proposal artifact. Re-run to revise, republishing to the **same** URL |
| `skills/myflow-do/` | `/myflow-do` | Implementation: SDD + TDD in an isolated worktree, the review panel, a **manual test guide** and a staged diff. Re-run to fix; a fix never moves the state. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/myflow-finish/` | `/myflow-finish` | **Run 1** integrates the branch (PR by default, merge, or manual). **Run 2**, once merged, archives the change and removes what the pipeline created |
| `skills/myflow-status/` | `/myflow-status` | Read-only state report for open changes |
| `skills/myflow-info/` | `/myflow-info` | Read-only — reads `skills/myflow-contracts/pipeline.md` and explains the pipeline |
| `skills/myflow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for any `/myflow-*` command) plus the state file, self-heal, project configuration, Jira, plan-provenance and build-green contracts, `jira-followups.md` when `/myflow-finish` run 1 files or joins a follow-up, `finish-contract.md` for `/myflow-finish`'s two-run procedure and no other command, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |
| `skills/openspec-explore/` | `/opsx:explore` | Thinking-partner mode — explore ideas, investigate, no implementation, no state |

### /myflow commands summary

**Pipeline (3 states):** `STARTED` → `IN_PROGRESS` → `FINISHED`

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (second run — it integrates first)
```

**That digest is the one piece of pipeline content this file copies, and the copy is deliberate.**
Codex loads this file into every session in this project, before any `/myflow-*` command runs and
before anything loads `skills/myflow-contracts/pipeline.md` — being present without that load is the
whole job of the block, which is why the always-on rule `rules/myflow-manual-review.mdc` carries the
same three lines. **What the copy reproduces is the states and the transitions, not the wording.**
The canonical block under **States** (`skills/myflow-contracts/pipeline.md`) closes its third line
by pointing at the finish contract — a pointer that resolves inside that file and nowhere else —
so a copy read *before* that file loads states the fact instead, and the two are deliberately
**not** kept byte-identical. Expect both renderings in one session: `/myflow-info` reads the
canonical block live and prints what it finds. A difference in what the three lines *say* is drift
worth reporting; a difference in how the third one is phrased is not. Everything else is cited
rather than copied: the state diagram and the per-command stage table live under
**Pipeline flow** (`skills/myflow-contracts/pipeline.md`), and `README.md` and `skills/README.md`
carry that citation with no digest at all, because a document read on demand can load the file it
would be copying.

Each command ends in the state named after it. **The human gate is a property of the state**, so
no command exists whose only job is to record that a review happened — there is no `*-done`
command. `/myflow-do` emits both the staged diff and the manual test guide, so reviewing and
testing are one sitting. **Every command is re-entrant, and a fix never moves the state.**

Also follow `rules/myflow-manual-review.mdc` (always-on) — it is a stub, so **load
`skills/myflow-contracts/pipeline.md` first**; that file holds the states, transitions, git
boundaries and the handoff shape, and is canonical for them. The finish contract lives in
`skills/myflow-contracts/finish-contract.md`, canonical for itself and loaded by `/myflow-finish`
alone.

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived)
change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name; anything else is reported
rather than ignored.

**Model:** `/myflow-start` → **Opus** (brainstorming/design benefits from stronger reasoning).
Every other command's **session** → **Sonnet**, and **every review-panel reviewer runs on the
panel's model — Sonnet by default** — regardless of the parent model, the default being what a
change may override by recording its own panel model. In Claude Code the *session* model is
enforced via `model:` frontmatter on each command; Cursor and Codex don't support per-command model
selection yet, so switch manually.

**The implementer subagents `/myflow-do` dispatches run on Opus** — or the harness's strongest
available model — named explicitly on each dispatch rather than inherited. Frontmatter cannot set
a subagent's model, so this rule is carried by the dispatch itself and recorded, per dispatch, in
the SDD ledger. It deliberately overrides superpowers:subagent-driven-development's "least
powerful model that can handle each role". See "Model policy" in
`skills/myflow-contracts/pipeline.md`, which is canonical.

| Command | What it does |
|---------|-------------|
| `/myflow-start <name>` | **Asks the planning effort level once**, on the run that creates the change — the three levels and which of them is the default are defined under **Planning effort** in State file (`skills/myflow-contracts/state-file.md`) and are deliberately not repeated here — it sizes the thinking inside the gates and never the gates themselves; that same run also asks which model implements the change, which one the review panel runs on and which one applies panel fixes, the three roles and their defaults being stated once under **Model policy** (`skills/myflow-contracts/pipeline.md`) and deliberately not repeated here; a revision round reuses the recorded level and models instead of asking again. Its stages, in order, are stated once under **Level 1 — the stages of each command** (`skills/myflow-contracts/pipeline.md`) and are deliberately not repeated here; the run ends at `state: STARTED`. Re-run at `STARTED` to revise the plan, republishing to the **same** URL |
| *(gate)* | **You** read the proposal artifact |
| `/myflow-do <name>` | Implements the change under SDD + TDD behind a **review panel** (primary + Bugbot + Principles required; Security, Adversarial and extra principle lenses conditional; all on the panel's model, Sonnet by default), which hands off only at **zero open findings at any severity**, and **stages the implementation, excluding the planning paths** `openspec/`, `docs/manual-test/` and `docs/superpowers/`. Its stages, in order, are stated once under **Level 1 — the stages of each command** (`skills/myflow-contracts/pipeline.md`) and are deliberately not repeated here; the run ends at `state: IN_PROGRESS`. Re-run to fix — a fix leaves the state unchanged. **No commits**, unless a `prUrl` is already recorded, in which case the fix is committed and pushed to that branch as two commits, implementation first |
| *(gate)* | **You** review the staged diff **and** run the apps against the guide |
| `/myflow-finish <name>` | Which run happens is `scripts/check-finish-preflight.sh`'s verdict, taken once per recorded worktree — `RUN1`, `RUN2` from every worktree, or `REFUSE`, which stops and asks you rather than guessing. Verdict `RUN1` integrates the branch — asking first how it should land, the choices being open a PR *(default)*, merge and push, or handle it manually — and stops at `IN_PROGRESS`. Verdict `RUN2`, reached once the branch has merged, archives the change, removes what the pipeline created, and ends at `state: FINISHED`. The stages of both runs, in order, are stated once under **Level 1 — the stages of each command** (`skills/myflow-contracts/pipeline.md`) and are deliberately not repeated here; which artifact is removed, when, and on what condition is likewise stated once under **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`). **Runs no tests, linters or coverage check** |
| `/myflow-status <name>` | Read-only state report for open changes |
| `/myflow-info` | Read-only — reads `skills/myflow-contracts/pipeline.md` and explains the pipeline |

The branch's merge status alone decides which `/myflow-finish` run happens — so a PR you merged on
the forge and a merge it performed itself are indistinguishable to it, which is correct.

**Verification runs during `/myflow-do` and nowhere else.** `/myflow-finish` has no verification
gate: re-running tests immediately before the one irreversible step repeats finished work, and a
gap found there routes back to `/myflow-do` anyway.

### How to invoke a skill

Read the skill file, then follow it:

```
Read file: skills/myflow-start/SKILL.md
(then follow the instructions in that file)
```

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/myflow-*` skills above.

Install Superpowers for Codex from its fork repo, per the Superpowers README (look for the
Codex install section), and enable multi-agent support in `~/.codex/config.toml`:
```toml
[features]
multi_agent = true
```

After install, general skills auto-trigger from their descriptions. Project-specific `/myflow-*`
skills are loaded on demand by reading their `SKILL.md` as described above.
