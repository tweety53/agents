---
name: myflow-contracts
description: The myflow pipeline itself plus its contract definitions — the three states and their transitions, the state file shape, state self-heal, project configuration, and Jira integration. Load the one file you need; each is canonical for its own contract. Referenced by the stubs in rules/myflow-manual-review.mdc.
allowed-tools: Bash(jq:*), Bash(git:*)
license: MIT
metadata:
  author: gymie
  version: "2.0"
---

# myflow contracts

myflow's pipeline and its contract definitions, split out of `rules/myflow-manual-review.mdc` so the
always-on rule layer carries only the trigger, without being asked to load anything.

**Load the one file you need — not this whole directory.** The exception is `pipeline.md`: every
`/myflow-*` command needs it, so load it first, always.

## Index

| File | Load it when you need to |
|------|--------------------------|
| [pipeline.md](pipeline.md) | **Run any `/myflow-*` command — load this first.** The three states and what each means, the command→state transition table, the wrong-state handoff, git boundaries, the handoff output shape, IntelliJ commands, and the finish contract (the two runs, and the worktree-removal checks) |
| [state-file.md](state-file.md) | Read or write a change's state file: its path, its full shape, monotonic state writes, carry-forward |
| [state-self-heal.md](state-self-heal.md) | Validate a state file against on-disk artifacts, or handle a missing/contradicted one |
| [project-configuration.md](project-configuration.md) | Resolve `.myflow/project.md` — apps, run, stop, test, lint, standards, jira — including standards-entry resolution and containment |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins — and a skill should **point at** these files rather than restate them, because a second
copy of a procedure drifts even when nobody edits it wrongly.

## Keeping this index honest

This file is the entry point to the contracts, so a stale entry here misdirects every command that
starts from it. When a contract file gains or loses a section, update the row above in the same
change — `scripts/check-references.sh` catches a **named** section that no longer exists, but it
cannot catch a description that is merely out of date.
