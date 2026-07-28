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
| `skills/openspec-propose/` | `/opsx:propose` | Propose a change — create all planning artifacts in one step |
| `skills/openspec-propose-superpowers/` | `/myflow-start` | Propose with Superpowers Basic Workflow #1 (brainstorming) + #3 (writing-plans) |
| `skills/openspec-propose-fix-superpowers/` | `/myflow-start-fix` | Revise the proposal after Gate A review, republish the artifact to the same URL, stay at `awaiting-proposal-review` |
| `skills/openspec-apply-change/` | `/opsx:apply` | Implement tasks from an OpenSpec change |
| `skills/openspec-apply-superpowers/` | `/myflow-do` | Implement with Superpowers #2–#6 (**stage with `git add`**; no commits; Gate B then Gate C next) |
| `skills/openspec-apply-fix-superpowers/` | `/myflow-do-fix` | Fix a Gate B / C / D finding — document in proposal (append or nested) → Superpowers #4–#6 in the **existing** worktree. Stages only at Gate B and C (no commits); **commits and pushes to the PR branch at Gate D** |
| `skills/openspec-fast-path-superpowers/` | `/myflow-fast-path` | Shortened single-session flow for small features — minimal artifacts, inline TDD, 3-agent review, ends at a PR (**never merges**); escalates to `/myflow-do` on size triggers |
| `skills/openspec-manual-test-superpowers/` | `/myflow-manual-test` | Gate C — write `docs/manual-test/<name>.md`, reply with link only |
| `skills/openspec-archive-change/` | `/opsx:archive` | Archive a completed change (delta sync + move to archive) |
| `skills/openspec-review-superpowers/` | `/myflow-review` | Verify Gate C → coverage check → tests/linters → **commit + push + open PR** (never merges unless `automerge`) |
| `skills/openspec-archive-superpowers/` | `/myflow-finish` | Verify the PR merged (Gate D) → delta sync → archive (also archives nested `<name>-fix-N` sub-changes) |
| `skills/openspec-full-cycle-superpowers/` | `/myflow-full` | start → do → Gate B review → Gate C manual test → review (commit+push+PR) → Gate D, stop (or `review-done` with `automerge`) |
| `skills/openspec-explore/` | `/opsx:explore` | Thinking-partner mode — explore ideas, investigate, no implementation |
| `skills/openspec-sync-specs/` | `/opsx:sync-specs` | Sync delta specs from a change to main specs |
| `skills/openspec-update-change/` | `/opsx:update` | Revise existing planning artifacts; keep them coherent |
| `skills/myflow-status/` | `/myflow-status` | Read-only stage report for open changes |
| `skills/myflow-info/` | `/myflow-info` | Reads the rule file and explains the pipeline |
| `skills/myflow-state-advance/` | *(internal)* | Pure state write used by every `*-done`/`*-manual-review` command: validates the incoming stage, writes the new one, prints the next step |
| `skills/myflow-contracts/` | *(on demand)* | The pipeline's contract definitions — state file, state self-heal, project configuration, Jira integration. Load the one file you need; each is canonical for its own contract |

### /myflow commands summary

**Pipeline (12 stages):** `awaiting-proposal-review` (Gate A) → `proposal-done` → `awaiting-do-review` (Gate B) → `do-review-started` → `do-done` → [`awaiting-fix-review` → `fix-review-started`] → `awaiting-manual-test` (Gate C) → `manual-test-done` → `awaiting-pr-review` (Gate D) → `review-done` → `finished`

**`*-done` and `*-manual-review` commands are pure state writes.** They call `myflow-state-advance` to update `stage`/`updatedAt`/`updatedBy` only — no verification, no artifact reading, no git. They exist to record a human confirmation as a discrete fact, separate from `/myflow-finish` independently verifying the PR merged.

Also follow `rules/myflow-manual-review.mdc` (always-on stage boundaries).

`<name>` is **optional** on every command below — if omitted, the sole active (non-archived) change relevant to that stage is used automatically; if there are multiple, you're asked which.

**Model:** `/myflow-start` → **Opus** (brainstorming/design benefits from stronger reasoning). Every other command → **Sonnet**. In Claude Code this is enforced via `model:` frontmatter on each command; Cursor and Codex don't support per-command model selection yet, so switch manually — see "Model policy" in `rules/myflow-manual-review.mdc`. Note `/myflow-full` runs as one session on Sonnet throughout, including Phase A — for brainstorming-heavy new work, run `/myflow-start` standalone first.

| Command | What it does |
|---------|-------------|
| `/myflow-start <name>` | Brainstorm → design approval gate → OpenSpec artifacts → writing-plans enriched tasks → publishes a proposal artifact → `stage: awaiting-proposal-review` |
| *(Gate A)* | **You** read the proposal artifact |
| `/myflow-start-fix <name>` | Revise the proposal after Gate A feedback, republish the artifact to the **same** URL, stay at `awaiting-proposal-review` |
| `/myflow-start-done <name>` | *Pure state write.* Confirms the proposal was reviewed → `stage: proposal-done` |
| `/myflow-do <name>` | git worktree → validate plan → SDD + TDD → **strict review panel** (primary + Bugbot + Principles required; Security, Adversarial and extra principle lenses conditional) → **`git add -A` (staged + uncommitted; no #7)** → `stage: awaiting-do-review` |
| *(Gate B)* | **You** open the worktree in IDE and review staged changes (`git diff --cached`) |
| `/myflow-do-manual-review <name>` | *Pure state write.* Confirms Gate B review is in progress → `stage: do-review-started` |
| `/myflow-do-done <name>` | *Pure state write.* Confirms the implementation diff was reviewed → `stage: do-done` |
| `/myflow-do-fix <name>` | Fix something found at Gate B (manual review), Gate C (manual test), or Gate D (PR review) — documents it in `proposal.md`/`tasks.md` (or a linked nested `<name>-fix-N` sub-change, your choice) → resumes the **same** worktree → SDD + TDD → strict review panel (targeted re-run by default; always full for Gate D origins or with `full-panel`). **Gate B and C: staged, no commits. Gate D: commits and pushes to the PR branch** (the one place this command commits; it still never merges). Records `originStage`; sets `stage: awaiting-fix-review`. Loop as many rounds as needed at any of the four origins. |
| `/myflow-do-fix-manual-review <name>` | *Pure state write.* Confirms review of the fix is in progress → `stage: fix-review-started` |
| `/myflow-do-fix-done <name>` | *Pure state write.* Confirms the fix was reviewed → returns to the stage `originStage` recorded, then clears `originStage` |
| `/myflow-manual-test <name>` | Write `docs/manual-test/<name>.md` (run apps + checklist); always asks whether to skip Gate C (default No); reply with **link only** → `stage: awaiting-manual-test` |
| *(Gate C)* | **You** run the apps and check off items in the guide |
| `/myflow-manual-test-done <name>` | *Pure state write.* Confirms manual testing is complete → `stage: manual-test-done` |
| `/myflow-review <name>` | Verifies every Gate C box is checked (or the guide is marked `SKIPPED`) — checks test coverage against delta specs (routes gaps to `/myflow-do-fix`) — then tests/linters → **commit + push + open PR** → `stage: awaiting-pr-review` (or, with `automerge`, commits + pushes + **merges** → `stage: review-done`, no PR) |
| *(Gate D)* | **You** review the PR and merge it — skipped entirely when `automerge` was used |
| `/myflow-review-done <name>` | *Pure state write.* Confirms the PR was reviewed (and merged) → `stage: review-done` |
| `/myflow-finish <name>` | Verifies the PR actually merged → delta sync → archive (also archives any nested `<name>-fix-N` sub-changes together) |
| `/myflow-full <name>` | Full cycle: Gate A (proposal) → do → Gate B (review) → Gate C (manual test) → review, ending at Gate D (PR open, stop) — or at `review-done` with `automerge`. Never auto-invokes any `*-done`/`*-manual-review` command — those remain separate human confirmations. |
| `/myflow-fast-path <name>` | **Fast path for small, well-understood features.** Minimal `proposal.md` + `tasks.md` → worktree → inline TDD → primary + Bugbot + Principles review (the three required panel slots) → tests/lint → commit + push + **open PR** → `stage: awaiting-pr-review` (`reviewed: false`, `tested: "skipped"`, `fastPath: true`). One human stop instead of five. `checkpoint` adds a staged-diff stop; stops and asks whenever the change stops looking small. Never merges. |
| `/myflow-status <name>` | Read-only stage report for open changes |
| `/myflow-info` | Reads the rule file and explains the pipeline |

**Flags:** `skip-propose`, `propose-only`, `skip-review` (skips Gate B only; the flag itself is the human's explicit opt-out, so the cycle writes `stage: do-done` with `gates.reviewed: false` — see "Opt-out (explicit only)" in the rule file), `skip-manual-test` (pre-answers the Gate C skip prompt with Yes, writing `stage: manual-test-done` with `gates.tested: "skipped"`, for the same reason; review still runs and still checks coverage), `automerge` (opt-in only, on `/myflow-review`/`/myflow-full` — commits, pushes, and **merges**, skipping Gate D and ending at `review-done`; never implied by any other flag), `full-panel` (forces every roster slot, including both extra principle lenses, over the whole-branch diff on every re-run, instead of the default targeted re-run), `commit-during-apply` (legacy), `checkpoint` (on `/myflow-fast-path` — adds a Gate B staged-diff stop before anything is pushed; the run is resumable by re-invoking the command)

### How to invoke a skill

Read the skill file, then follow it:

```
Read file: skills/openspec-propose-superpowers/SKILL.md
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
