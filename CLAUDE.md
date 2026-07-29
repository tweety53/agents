# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `agents/rules/lint-fix-priority.mdc`. It is not restated here — one
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
     that lacks a `CLAUDE.md`, so a standard hardcoded to one stack would be wrong in
     every other project. A standard meant to apply across *many* projects belongs in
     `agents/rules/` as an opt-in rule instead, activated per project by naming it in
     `.myflow/project.md`'s `## standards` section — `kotlin-backend-development-standard.mdc`
     is the worked example of that pattern. -->

This project has not declared one yet. Until it does, follow the language's published
conventions and the patterns already present in the surrounding code.

---

## Project Skills (OpenSpec / /myflow workflow)

These skills live in `skills/` next to this file (or in `.claude/skills/` if installed there).
To invoke a skill: **read its `SKILL.md` file** then follow the instructions within.

All skills require the `openspec` CLI to be installed.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/myflow-start/` | `/myflow-start` | Brainstorm → design gate → OpenSpec artifacts → writing-plans → publish the proposal artifact. Re-run to revise, republishing to the **same** URL |
| `skills/myflow-do/` | `/myflow-do` | Worktree → SDD + TDD → review panel → **manual test guide** → lint + tests → `git add`. Re-run to fix; a fix never moves the state. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/myflow-finish/` | `/myflow-finish` | **Run 1** integrate (PR by default, merge, or manual). **Run 2** (once merged) sync specs, archive, commit + push the archive, remove the worktrees |
| `skills/myflow-status/` | `/myflow-status` | Read-only state report for open changes |
| `skills/myflow-info/` | `/myflow-info` | Read-only — reads `skills/myflow-contracts/pipeline.md` and explains the pipeline |
| `skills/myflow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for any `/myflow-*` command) plus the state file, self-heal, project configuration and Jira contracts. Load the one file you need |
| `skills/openspec-explore/` | `/opsx:explore` | Thinking-partner mode — explore ideas, investigate, no implementation, no state |

### /myflow commands summary

**Pipeline (3 states):** `STARTED` → `IN_PROGRESS` → `FINISHED`

```text
/myflow-start  → STARTED      you: read the proposal artifact
/myflow-do     → IN_PROGRESS  you: review the staged diff and run the apps
/myflow-finish → FINISHED     terminal (it integrates on its first run)
```

Each command ends in the state named after it. **The human gate is a property of the state**, so
no command exists whose only job is to record that a review happened — there is no `*-done`
command. `/myflow-do` emits both the staged diff and the manual test guide, so reviewing and
testing are one sitting. **Every command is re-entrant, and a fix never moves the state.**

Also follow `rules/myflow-manual-review.mdc` (always-on) — it is a stub, so **load
`skills/myflow-contracts/pipeline.md` first**; that file holds the states, transitions, git
boundaries, the handoff shape and the finish contract, and is canonical.

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived)
change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name; anything else is reported
rather than ignored.

**Model:** `/myflow-start` → **Opus** (brainstorming/design benefits from stronger reasoning).
Every other command's **session** → **Sonnet**, and **every review-panel reviewer runs on Sonnet**
regardless of the parent model. In Claude Code the *session* model is enforced via `model:`
frontmatter on each command; Cursor and Codex don't support per-command model selection yet, so
switch manually.

**The implementer subagents `/myflow-do` dispatches run on Opus** — or the harness's strongest
available model — named explicitly on each dispatch rather than inherited. Frontmatter cannot set
a subagent's model, so this rule is carried by the dispatch itself and recorded, per dispatch, in
the SDD ledger. It deliberately overrides superpowers:subagent-driven-development's "least
powerful model that can handle each role". See "Model policy" in
`skills/myflow-contracts/pipeline.md`, which is canonical.

| Command | What it does |
|---------|-------------|
| `/myflow-start <name>` | Brainstorm → design approval gate → OpenSpec artifacts → writing-plans enriched tasks → publishes a proposal artifact → `state: STARTED`. Re-run at `STARTED` to revise the plan, republishing to the **same** URL |
| *(gate)* | **You** read the proposal artifact |
| `/myflow-do <name>` | git worktree → validate plan → SDD + TDD → **review panel** (primary + Bugbot + Principles required; Security, Adversarial and extra principle lenses conditional; all on Sonnet) → writes `docs/manual-test/<name>.md` → runs the project's lint + test commands → **`git add -A`** → `state: IN_PROGRESS`. Re-run to fix — a fix leaves the state unchanged. **No commits**, unless a `prUrl` is already recorded, in which case the fix is committed and pushed to that branch |
| *(gate)* | **You** review the staged diff **and** run the apps against the guide |
| `/myflow-finish <name>` | **Run 1 — branch not merged:** asks how it should land (open a PR *(default)*, merge and push, or handle it manually), commits the staged work, takes that route, **stops** at `IN_PROGRESS`. **Run 2 — branch merged:** verifies the merge, syncs delta specs, archives (with any nested `<name>-fix-N`), **commits and pushes the archive**, **removes the worktrees and branches** after four gating safety checks plus a disclosure of the ignored files `--force` will destroy → `state: FINISHED`. **Runs no tests, linters or coverage check** |
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

Install Superpowers in Claude Code:
```
/plugin install prime-radiant-inc/superpowers
```

After install, general skills auto-trigger from their descriptions. Project-specific `/myflow-*`
skills are loaded on demand by reading their `SKILL.md` as described above.
