# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `<agents repo>/rules/lint-fix-priority.mdc`. It is not restated here — one
source of truth, so the policy cannot drift between the global copy and this file.

What is project-specific is which commands it means. Full list in `<project>/.myflow/project.md`'s `## lint`
section — named here so this file states them rather than leaving a placeholder:

```bash
cd stats && gofmt -w .                                  # auto-fix, Go source only — nothing else in
                                                          # this repository has an auto-fix command
scripts/check-vocabulary.sh                              # plus every other scripts/check-*.sh guard
scripts/check-references.sh                               # named in .myflow/project.md's `## lint`
cd stats && go vet ./... && gofmt -l .                   # must exit clean before claiming Go work done
cd stats/web && npx tsc -b                                # must exit clean before claiming SPA work done
```

Pre-approved suppressions and documented deviations live in `CONTRIBUTING.md`. Do not
expand that list without user approval.

---

### Project-specific standards

<!-- Replace this section with the coding standard this project actually follows:
     module layout, layering rules, naming conventions, framework constraints, and the
     test command to run before claiming completion.

     This template ships generic on purpose. `<agents repo>/setup.sh` copies it into any project root
     that lacks a `<project>/CLAUDE.md`, so a standard hardcoded to one stack would be wrong in
     every other project. A standard meant to apply across *many* projects belongs in
     `<agents repo>/rules/` as an opt-in rule instead, activated per project by naming it in
     `<project>/.myflow/project.md`'s `## standards` section — `kotlin-backend-development-standard.mdc`
     is the worked example of that pattern. -->

This project has not declared one yet. Until it does, follow the language's published
conventions and the patterns already present in the surrounding code.

---

## Project Skills (OpenSpec / /myflow workflow)

These skills live in `skills/` next to this file (or in `<project>/.claude/skills/` if installed there).
To invoke a skill: **read its `SKILL.md` file** then follow the instructions within.

All skills require the `openspec` CLI to be installed.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/myflow-start/` | `/myflow-start` | Proposal: brainstorming behind a design gate, the OpenSpec artifacts, and a published proposal artifact. Re-run to revise, republishing to the **same** URL |
| `skills/myflow-do/` | `/myflow-do` | Implementation: SDD + TDD in an isolated worktree, the review panel, and a staged diff whose handoff carries the run instructions. Re-run to fix; a fix never moves the state. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/myflow-finish/` | `/myflow-finish` | **Run 1** integrates the branch (PR by default, merge, or manual). **Run 2**, once merged, archives the change and removes what the pipeline created |
| `skills/myflow-fast/` | `/myflow-fast` | Composite: chains `/myflow-start`'s brainstorming and `/myflow-do`'s implementation into one invocation, and `/myflow-finish`'s two runs into one when the chosen route needs no external merge — same three states, same state file, no human gate in between. Publishes no proposal artifact. See `skills/myflow-fast/SKILL.md` |
| `skills/myflow-status/` | `/myflow-status` | Read-only state report for open changes |
| `skills/myflow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for any `/myflow-*` command) plus the state file, project configuration, Jira, plan-provenance and build-green contracts, `jira-followups.md` when `/myflow-finish` run 1 files or joins a follow-up, `finish-contract.md` for `/myflow-finish`'s two-run procedure and no other command, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |
| `skills/openspec-explore/` | `/opsx:explore` | Thinking-partner mode — explore ideas, investigate, no implementation, no state |

### /myflow commands summary

**Pipeline (3 states):** `STARTED` → `IN_PROGRESS` → `FINISHED`

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (second run — it integrates first)
```

**That digest is the one piece of pipeline content this file copies, and the copy is deliberate.**
Claude Code loads this file into every session in this project, before any `/myflow-*` command runs
and before anything loads `skills/myflow-contracts/pipeline.md` — being present without that load is
the whole job of the block, which is why the always-on rule `rules/myflow-manual-review.mdc` carries
the same three lines. **What the copy reproduces is the states and the transitions, not the
wording.** The canonical block under **States** (`skills/myflow-contracts/pipeline.md`) closes its
third line by pointing at the finish contract — a pointer that resolves inside that file and
nowhere else — so a copy read *before* that file loads states the fact instead, and the two are
deliberately **not** kept byte-identical: a difference in what the three lines *say* is drift worth
reporting; a difference in how the third one is phrased is not. Everything else is cited rather than
copied: the state diagram lives under **States** (`skills/myflow-contracts/pipeline.md`), and each
command's own stage sequence is spelled out in its own `SKILL.md` under `skills/`. `<agents repo>/README.md`'s
"How the pipeline works" section covers the same ground for a reader browsing the repository
directly — it is not copied by `<agents repo>/setup.sh` into any project, so it is not cited here as a source of
record.

Also follow `rules/myflow-manual-review.mdc` (always-on) — it is a stub, so **load
`skills/myflow-contracts/pipeline.md` first**; that file holds the states, transitions, git
boundaries and the handoff shape, and is canonical for them. The finish contract lives in
`skills/myflow-contracts/finish-contract.md`, canonical for itself and loaded by `/myflow-finish`
alone.

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived)
change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name; anything else is reported
rather than ignored.

**Model:** See "Model policy" in
`skills/myflow-contracts/pipeline.md`, which is canonical.

| Command | What it does |
|---------|-------------|
| `/myflow-start <name>` | **Asks the planning effort level once**, on the run that creates the change — the three levels and which of them is the default are defined under **Planning effort** in State file (`skills/myflow-contracts/state-file.md`) — it sizes the thinking inside the gates and never the gates themselves; that same run also asks which model implements the change, which one the review panel runs on and which one applies panel fixes, the three roles and their defaults being stated once under **Model policy** (`skills/myflow-contracts/pipeline.md`); it also asks which review-panel roster the panel runs under — `light`, `standard` or `full`, `light` the default — defined under **5. The review panel** (`skills/myflow-do/SKILL.md`); a revision round reuses the recorded level, models and roster instead of asking again. Its stages, in order, begin at **A. Resolve the change** (`skills/myflow-start/SKILL.md`); the run ends at `state: STARTED`. Re-run at `STARTED` to revise the plan, republishing to the **same** URL |
| *(gate)* | **You** read the proposal artifact |
| `/myflow-do <name>` | Implements the change under SDD + TDD behind a **review panel** sized by the recorded `reviewPanelRoster` — `light` *(default)*, `standard` or `full`, each dispatching exactly three required slots on the panel's model, Sonnet by default, per **5. The review panel** (`skills/myflow-do/SKILL.md`), which is canonical for what each preset means — which hands off only at **zero open findings at any severity**. Its stages, in order, begin at **1. Load context and validate the plan** (`skills/myflow-do/SKILL.md`); the run ends at `state: IN_PROGRESS`. Re-run to fix — a fix leaves the state unchanged |
| *(gate)* | **You** review the staged diff **and** run the apps |
| `/myflow-finish <name>` | Which run happens is `<agents repo>/scripts/check-finish-preflight.sh`'s verdict, taken once per worktree in the resolved set — `RUN1`, `RUN2` from every worktree, or `REFUSE`, which stops and asks you rather than guessing; an empty resolved set is never a vacuous `RUN2` either, per **Resolving a change's worktrees** (`skills/myflow-contracts/pipeline.md`). Verdict `RUN1` integrates the branch — asking first how it should land, the choices being open a PR *(default)*, merge and push, or handle it manually — and stops at `IN_PROGRESS`. Verdict `RUN2`, reached once the branch has merged, archives the change, removes what the pipeline created, and ends at `state: FINISHED`. The stages of run 1, in order, begin at **Run 1 — integrate**, and the stages of run 2 begin at **Run 2 — archive and clean up** (both `skills/myflow-finish/SKILL.md`); which artifact is removed, when, and on what condition is likewise stated once under **Temporary artifacts registry** (`skills/myflow-contracts/pipeline.md`). **Runs no tests, linters or coverage check** |
| `/myflow-fast <name>` | Composite command: chains `/myflow-start`'s brainstorming (fully interactive, unchanged) directly into `/myflow-do`'s implementation and review panel on a creating run, and chains `/myflow-finish`'s run 1 into run 2 when the landing route is merge-and-push, with no operator action between the chained stages. Accepts **no state** (creates the change) or **`IN_PROGRESS`** — an argument at `IN_PROGRESS` is fix instructions, a bare invocation asks how to land the branch. Publishes no proposal artifact. Its stages, in order, are cited by section in `skills/myflow-fast/SKILL.md`; a creating or fix run ends at `state: IN_PROGRESS` — a fix leaves the state unchanged. A bare invocation ends at `state: IN_PROGRESS` or `state: FINISHED`, depending on the route chosen. Re-run to fix or to integrate |
| `/myflow-status <name>` | Read-only state report for open changes |

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

Install Superpowers in Claude Code:
```
/plugin install prime-radiant-inc/superpowers
```

After install, general skills auto-trigger from their descriptions. Project-specific `/myflow-*`
skills are loaded on demand by reading their `SKILL.md` as described above.
