# Agent Instructions (Claude Code)

This file is the active instruction set for Claude Code sessions in this project.
It contains mandatory rules and an index of project-specific skills.

---

## Mandatory Rules

### Lint Fix Priority

The fix-first lint policy is a **global rule**, installed into the managed block in
`~/.claude/CLAUDE.md` from `<agents repo>/rules/lint-fix-priority.mdc`. It is not restated here — one
source of truth, so the policy cannot drift between the global copy and this file.

What is project-specific is which commands it means: `<project>/.flow/project.md`'s `## lint`
section lists them.

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
To invoke a skill, type its slash command (`/flow`, `/flow-fast`, …); the harness loads its
`SKILL.md` through the Skill tool.

Every skill below but `flow-plan` and `flow-contracts` requires the `spectre` CLI to be
installed. Those two need none — reading a spectre tree, or a contract file, is reading markdown.

### Skill index

| Skill directory | Trigger | Purpose |
|-----------------|---------|---------|
| `skills/flow/` | `/flow` | Single-command pipeline: brainstorming behind a design gate, implementation under SDD + TDD behind the review panel resolved from the settings store, and integrate/archive across the same three-state pipeline, pausing only at the human gates. Re-run to resume, fix, or integrate. Carries the reviewer prompts + `engineering-principles.md` |
| `skills/flow-fast/` | `/flow-fast` | Minimal-ceremony `/flow` variant: one invocation from Jira key to landed change. A git worktree for isolation only, implementation and review panel as the plan's class decides, project lint plus targeted tests, the project's default landing route, cleanup. Marks every `flow.*` stage `/flow` marks and keeps the Jira transitions; no spectre artifacts or state file |
| `skills/flow-status/` | `/flow-status` | Read-only state report for open changes |
| `skills/flow-plan/` | `/flow-plan` | Thinking-partner mode — explore ideas, investigate, no implementation; a captured session creates the change at `STARTED` for `/flow` to resume |
| `skills/flow-settings/` | `/flow-settings` | Reads/writes the harness-wide default model and reviewer slots every `/flow` run reads from. Standalone, not a pipeline stage |
| `skills/flow-self-review/` | `/flow-self-review` | Runs a self-review pass a `/flow` run deferred, inline on this session's model, from the saved context bundle. Standalone, not a pipeline stage |
| `skills/flow-contracts/` | *(on demand)* | The pipeline itself (`pipeline.md` — **load first** for `/flow`) plus the state file, project configuration, Jira, plan-provenance and build-green contracts, `jira-followups.md` when `/flow`'s integrate run 1 files or joins a follow-up, `finish-contract-run1.md`/`finish-contract-run2.md` for `/flow`'s two-run integrate/archive procedure, and `workspace-isolation.md` when a run needs a worktree's own database, cache index, bucket or ports. Load the one file you need — and never a `-rationale.md` appendix, which carries a contract's or a skill's reasoning for whoever edits it and is not loaded by a run |

### How to invoke a skill

Type its slash command. Each file in `commands-claude/` names one skill plus its accepted states,
and the Skill tool loads that skill by name.

### Superpowers general skills

The Superpowers plugin provides general-purpose workflow skills (brainstorming, TDD,
subagent-driven-development, etc.). These are referenced by the `/flow` skill above.

Install it per `<agents repo>/README.md`'s Claude Code section.

After install, general skills auto-trigger from their descriptions. Project-specific `/flow`
skills are loaded on demand by their slash commands as described above.
