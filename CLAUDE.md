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

`flowd` on `127.0.0.1:4173`, the `flow-postgres` container on host port 5433, and the default
`flow` database inside it are the **dev workspace's** service and storage — the store every
`flow` call in every project writes state, stage marks and records through. No agent action stops
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
| `skills/flow/` | `/flow` | Single-command pipeline: brainstorming behind a design gate, implementation under SDD + TDD behind the review panel resolved from the settings store, and integrate/archive across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/flow-status/` | `/flow-status` | Read-only state report for open changes |
| `skills/flow-research/` | `/flow-research` | Thinking-partner mode — explore ideas, investigate, no implementation, no state; stages research notes for `/flow`'s brainstorming to seed from |
| `skills/flow-settings/` | `/flow-settings` | Reads/writes the harness-wide default model and reviewer slots every `/flow` run reads from. Standalone, not a pipeline stage |
| `skills/flow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for `/flow`) plus the state file, project configuration, Jira, plan-provenance and build-green contracts, `jira-followups.md` when `/flow`'s integrate run 1 files or joins a follow-up, `finish-contract-run1.md`/`finish-contract-run2.md` for `/flow`'s two-run integrate/archive procedure, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |

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
