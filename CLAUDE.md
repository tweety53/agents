# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `<agents repo>/rules/lint-fix-priority.mdc`. It is not restated here — one
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

### Never stop the dev workspace's stats service or its storage

`myflowd` on `127.0.0.1:4173`, the `myflow-postgres` container on host port 5433, and the default
`myflow` database inside it are the **dev workspace's** service and storage — the store every
`myflow` call in every project writes state, stage marks and records through. No agent action stops
or drops them: not `docker compose down`, not `launchctl unload`, not a `kill` on the daemon's pid,
and not to make a later step succeed. Bringing them back up does not repair a run that already fell
through to the on-disk journal.

**Only those.** A worktree's own derived `flow_<id>` database and bucket are per-change artifacts,
and `<project>/scripts/workspace.sh remove <id>` drops them during archive cleanup as it should.

The reasoning, that boundary, and why `/flow`'s worktree-cleanup check 5 therefore has nothing to
run, are the `## stop` section of `<project>/.flow/project.md` — canonical for it, and not
restated here. Stopping the dev stack is an operator action; the commands live in that file's
`## run` section and in `<project>/stats/README.md`.

---

### Project-specific standards

<!-- Replace this section with the coding standard this project actually follows:
     module layout, layering rules, naming conventions, framework constraints, and the
     test command to run before claiming completion.

     This template ships generic on purpose. `<agents repo>/setup.sh` copies it into any project root
     that lacks a `<project>/CLAUDE.md`, so a standard hardcoded to one stack would be wrong in
     every other project. A standard meant to apply across *many* projects belongs in
     `<agents repo>/rules/` as an opt-in rule instead, activated per project by naming it in
     `<project>/.flow/project.md`'s `## standards` section — `kotlin-backend-development-standard.mdc`
     is the worked example of that pattern. -->

This project has not declared one yet. Until it does, follow the language's published
conventions and the patterns already present in the surrounding code.

---

## Project Skills (spectre / /flow workflow)

These skills live in `skills/` next to this file (or in `<project>/.claude/skills/` if installed there).
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
Claude Code loads this file into every session in this project, before `/flow` runs
and before anything loads `skills/flow-contracts/pipeline.md` — being present without that load is
the whole job of the block, which is why the always-on rule `rules/flow-manual-review.mdc` carries
the same three lines. **What the copy reproduces is the states and the transitions, not the
wording**, and the two are deliberately **not** kept byte-identical: a difference in what the lines
*say* is drift worth reporting; a difference in how they are phrased is not. Everything else is
cited rather than copied: the state diagram lives under **How the pipeline works** (`<agents repo>/README.md`),
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

Install Superpowers in Claude Code:
```
/plugin install prime-radiant-inc/superpowers
```

After install, general skills auto-trigger from their descriptions. Project-specific `/flow`
skills are loaded on demand by reading their `SKILL.md` as described above.
