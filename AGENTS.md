# Agent Instructions (Codex)

This file is the active instruction set for Codex sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Where a Codex session gets its rules

Codex reads `<project>/AGENTS.md` — the project's own, plus `~/.codex/AGENTS.md` globally. It does
not read `~/.claude/CLAUDE.md` or `~/.cursor/rules/`.

`<agents repo>/setup.sh global` writes the always-on rules into a managed block in `~/.codex/AGENTS.md`,
delimited by `<!-- myflow:begin -->` / `<!-- myflow:end -->`, using the same ordering check
and self-poisoning guard as the Claude Code block. So a Codex session gets
`flow-manual-review.mdc` and `lint-fix-priority.mdc` globally, and opt-in rules (such as
the Kotlin backend standard) are deliberately excluded — a project activates those by
naming them in `<project>/.flow/project.md`'s `## standards` section.

Project-mode `<agents repo>/setup.sh codex` installs skills and this file but no rules, symmetrically
with `claude-code`. Run `<agents repo>/setup.sh global` for the rule layer.

### What a global install actually leaves a Codex session with

Be precise about this, because the two halves are asymmetric:

| | After `<agents repo>/setup.sh global` |
|---|---|
| **Rules** | ✅ present — the managed block in `~/.codex/AGENTS.md` carries `flow-manual-review.mdc` and `lint-fix-priority.mdc` |
| **Skills** | ✅ present — `install_global` links every directory in `skills/` into `~/.codex/skills/`, alongside `~/.claude/skills/` and `~/.cursor/skills/` |
| **Commands** | ❌ absent — `~/.claude/commands/` and `~/.cursor/commands/` only. There is no `~/.codex/commands/` layer. |

So a Codex session has the rules and the skills, but no slash-command layer: typing
`/flow` will not resolve, even though the skill it delegates to is installed.

**What to do today:** invoke the skill directly instead of through a command — read its
`SKILL.md` out of the globally installed tree and follow it, e.g.

```
Read file: ~/.codex/skills/flow/SKILL.md
(then follow the instructions in that file)
```

That path is a symlink into this checkout, so the content is always current. Each command
file in `commands-claude/` is a thin wrapper naming exactly one skill plus its accepted
states, so reading the skill directly loses nothing but the shorthand.

Do **not** work around the missing command layer with a per-project `<agents repo>/setup.sh codex`
install. A project-local copy shadows the global one and then goes stale: installs are additive, so
an entry deleted from this checkout leaves its symlink behind at every destination it was ever
installed to. `<agents repo>/setup.sh`'s own `link_into` and `prune_stale_links` exist to prevent
exactly that shadowing, and a second install per project reintroduces it.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.codex/AGENTS.md` from `<agents repo>/rules/lint-fix-priority.mdc`. It is not restated here — one
source of truth, so the policy cannot drift between the global copy and this file.

What is project-specific is which commands it means. Full list in `<project>/.flow/project.md`'s `## lint`
section — named here so this file states them rather than leaving a placeholder:

```bash
cd stats && gofmt -w .                                  # auto-fix, Go source only — nothing else in
                                                          # this repository has an auto-fix command
scripts/check-vocabulary.sh                              # plus every other scripts/check-*.sh guard
scripts/check-references.sh                               # named in .flow/project.md's `## lint`
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
     that lacks an `<project>/AGENTS.md`, so a standard hardcoded to one stack would be wrong in
     every other project. A standard meant to apply across *many* projects belongs in
     `<agents repo>/rules/` as an opt-in rule instead, activated per project by naming it in
     `<project>/.flow/project.md`'s `## standards` section — `kotlin-backend-development-standard.mdc`
     is the worked example of that pattern. -->

This project has not declared one yet. Until it does, follow the language's published
conventions and the patterns already present in the surrounding code.

---

## Project Skills (spectre / /flow workflow)

These skills live in `skills/` next to this file (or in `<project>/.codex/skills/` if installed there).
To invoke a skill: **read its `SKILL.md` file** then follow the instructions within.

Every skill below but `flow-research` and `flow-contracts` requires the `spectre` CLI to be
installed. Those two need none — reading a spectre tree, or a contract file, is reading markdown.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/flow/` | `/flow` | Single-command pipeline: brainstorming behind a design gate, implementation under SDD + TDD behind the fixed 3-slot review panel, and integrate/archive across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/flow-status/` | `/flow-status` | Read-only state report for open changes |
| `skills/flow-research/` | `/flow-research` | Thinking-partner mode — explore ideas, investigate, no implementation, no state; stages research notes for `/flow`'s brainstorming to seed from |
| `skills/flow-settings/` | `/flow-settings` | Reads/writes the harness-wide default model and reviewer slots every `/flow` run reads from. Standalone, not a pipeline stage |
| `skills/flow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for `/flow`) plus the state file, project configuration, Jira, plan-provenance and build-green contracts, `jira-followups.md` when `/flow`'s integrate run 1 files or joins a follow-up, `finish-contract.md` for `/flow`'s two-run integrate/archive procedure, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |

### /flow commands summary

**Pipeline (3 states):** `STARTED` → `IN_PROGRESS` → `FINISHED`

```text
/flow  (no state)          → STARTED → IN_PROGRESS   you: review the staged diff and run the apps
/flow  <fix instructions>  → IN_PROGRESS (unchanged)  you: review the staged diff and run the apps
/flow  (bare, IN_PROGRESS) → IN_PROGRESS or FINISHED  terminal only on the merge-and-push route
```

**That digest is the one piece of pipeline content this file copies, and the copy is deliberate.**
Codex loads this file into every session in this project, before `/flow` runs and
before anything loads `skills/flow-contracts/pipeline.md` — being present without that load is the
whole job of the block, which is why the always-on rule `rules/flow-manual-review.mdc` carries the
same three lines. **What the copy reproduces is the states and the transitions, not the wording**,
and the two are deliberately **not** kept byte-identical: a difference in what the lines *say* is
drift worth reporting; a difference in how they are phrased is not. Everything else is cited
rather than copied: the state diagram lives under **How the pipeline works** (`<agents repo>/README.md`),
and `/flow`'s own stage sequence is spelled out across `skills/flow/*.md`.

Also follow `rules/flow-manual-review.mdc` (always-on) — it is a stub, so **load
`skills/flow-contracts/pipeline.md` first**; that file holds the states, transitions, git
boundaries and the handoff shape, and is canonical for them. The finish contract lives in
`skills/flow-contracts/finish-contract.md`, canonical for itself and loaded by `/flow`'s
integrate/archive phase alone.

`<name>` is **optional** on `/flow` and `/flow-status` — if omitted, the sole active (non-archived)
change relevant to that state is used automatically; if there are multiple, you're asked which.

**No command takes a flag.** The only argument is the change name (or, on a creating `/flow` run, a
description or Jira key); anything else is reported rather than ignored.

**Model:** See "Model resolution" in `skills/flow/SKILL.md`, canonical for `/flow`; "Model policy" in
`skills/flow-contracts/model-policy.md` for the per-harness enforcement notes that still apply.

| Command | What it does |
|---------|-------------|
| `/flow <name>` | No state creates the change and writes `STARTED` immediately, then — same invocation — runs brainstorming (fully interactive, unchanged) and implementation behind the fixed 3-slot review panel (Primary, Principles, Code review (low); Bugbot and Security only when explicitly asked), ending at `IN_PROGRESS`. Asks no planning-effort, model, or review-panel-roster question on a creating run, and publishes no proposal artifact. An argument at `IN_PROGRESS` is a fix run — state unchanged. Bare at `IN_PROGRESS`, it asks how to land the branch — open PR *(default)*, merge and push, or manual — and, on merge-and-push, continues the same invocation through archive to `FINISHED`; open PR and manual stop and hand off. **Runs no tests, linters or coverage check outside implementation's own verify stage** |
| *(gate)* | **You** — creating run or fix: review the staged diff **and** run the apps; integrate with open PR or manual: wait for the branch to merge (or finish your manual steps); merge-and-push: nothing — the state is terminal |
| `/flow-status <name>` | Read-only state report for open changes |

The branch's merge status alone decides which `/flow` integrate/archive run happens — so a PR you
merged on the forge and a merge it performed itself are indistinguishable to it, which is correct.

**Verification runs during `/flow`'s implementation phase and nowhere else.** The integrate/archive
phase has no verification gate: re-running tests immediately before the one irreversible step
repeats finished work, and a gap found there routes back to a fix run anyway.

### How to invoke a skill

Read the skill file, then follow it:

```
Read file: skills/flow/SKILL.md
(then follow the instructions in that file)
```

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/flow` skill above.

Install Superpowers for Codex from its fork repo, per the Superpowers README (look for the
Codex install section), and enable multi-agent support in `~/.codex/config.toml`:
```toml
[features]
multi_agent = true
```

After install, general skills auto-trigger from their descriptions. Project-specific `/flow`
skills are loaded on demand by reading their `SKILL.md` as described above.
