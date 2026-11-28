# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `<agents repo>/rules/lint-fix-priority.mdc`.

What is project-specific is which commands it means: `<project>/.flow/project.md`'s `## lint`
section lists them.

---

### Never stop the dev workspace's stats service or its storage

`flowd` on `127.0.0.1:4173`, the `flow-postgres` container on host port 5433, and the default
`flow` database inside it are the **dev workspace's** service and storage — the store every
`flow` call in every project writes state, stage marks and records through. No agent action stops
or drops them: not `docker compose down`, not `launchctl unload`, not a `kill` on the daemon's pid,
and not to make a later step succeed. Bringing them back up does not repair a run that already fell
through to the on-disk journal.

The reasoning and the boundary are the `## stop` section of `<agents repo>/.flow/project.md` —
canonical for it, and not restated here. Stopping the dev stack is an operator action; the commands
live in that file's `## run` section and in `<agents repo>/stats/README.md`.

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

---

## Project Skills (spectre / /flow workflow)

These skills live in `skills/` next to this file (or in `<project>/.claude/skills/` if installed there).
To invoke a skill, type its slash command (`/flow`, `/flow-interactive`, …); the harness loads its
`SKILL.md` through the Skill tool.

Every skill below but `flow-plan` and `flow-contracts` requires the `spectre` CLI to be
installed. Those two need none — reading a spectre tree, or a contract file, is reading markdown.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/flow/` | `/flow` | Single-command pipeline: brainstorming behind a design gate, implementation under SDD + TDD behind the decided review panel, and integrate/cleanup across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/flow-interactive/` | `/flow-interactive` | Conversational session: one invocation is one change, one Jira issue and one worktree, and a stream of requests the main session coordinates — each looked at live, then by one subagent implemented test-first, verified with the project's lint and targeted tests, landed by merge and push, reinstalled and, when it touches `ui paths`, checked on the live page — until an explicit stop runs one self-review pass, cleanup and `FINISHED`. Every request is recorded in flowd; no spectre artifacts or state file |
| `skills/flow-plan/` | `/flow-plan` | Thinking-partner mode — explore ideas, investigate, no implementation; a captured session creates the change at `STARTED` for `/flow` to resume |
| `skills/flow-self-review/` | *(run by the pipeline)* | Runs a change's self-review pass, a one-shot dispatch — inside `/flow`'s integrate and `/flow-interactive`'s stop, or standalone from the context bundle flowd saved at cleanup; fixes and lands every `flow-fix` finding that is not big, suggests the impactful big ones and other angles' findings (no Jira issue), records each in the flow store. Standalone, not a pipeline stage |
| `skills/flow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for `/flow`) plus the state file, project configuration and Jira contracts, `finish-contract-run1.md`/`finish-contract-run2.md` for `/flow`'s two-run integrate/cleanup procedure, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need; a contract's or a skill's reasoning lives in `docs/rationale/` |

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/flow` skill above.

Install it per `<agents repo>/README.md`'s Claude Code section.
