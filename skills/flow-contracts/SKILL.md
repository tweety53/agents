---
name: flow-contracts
description: The flow pipeline itself plus its contract definitions — the three states and their transitions, the state file shape, project configuration, Jira integration, follow-up issues, and workspace isolation. Load the one file you need; each is canonical for its own contract. The reasoning behind them lives in docs/rationale/. Referenced by the stubs in rules/flow-manual-review.mdc.
allowed-tools: Bash(jq:*), Bash(git:*)
license: MIT
metadata:
  author: gymie
  version: "2.0"
---

# flow contracts

flow's pipeline and its contract definitions.

**Load the one file you need — not this whole directory.** The exception is `pipeline.md`: every
`/flow*` command needs it, so load it first, always.

## Index

| File | Load it when you need to |
|------|--------------------------|
| [pipeline.md](pipeline.md) | **Run any `/flow*` command — load this first.** The three states and what each means, the command→state transition table, the wrong-state handoff, the handoff output shape, guard resolution and stage marks |
| [finish-contract-run1.md](finish-contract-run1.md) | `/flow`'s integrate run: the preflight-signal decision, run 1's procedure — the self-review pass included — base-branch resolution, and resolving a change's worktrees. **Loaded by `skills/flow/integrate.md`; `/flow-self-review` also reads single sections of it** |
| [finish-contract-run2.md](finish-contract-run2.md) | `/flow`'s cleanup run: run 2's procedure and worktree cleanup. **Loaded by `skills/flow/cleanup.md`** |
| [state-file.md](state-file.md) | Read or write a change's state file: its full shape, monotonic state writes, carry-forward |
| [project-configuration.md](project-configuration.md) | Resolve `<project>/.flow/project.md` — every key it defines and the single-line-literal keys. |
| [project-configuration-standards.md](project-configuration-standards.md) | Resolve a `## standards` entry: the three entry forms and containment. **Loaded at the principles dispatch** |
| [project-configuration-isolation.md](project-configuration-isolation.md) | The `## workspace isolation` tables: the cell forms, row validation and what is left to the agent. **Loaded only when writing or reviewing a `## workspace isolation` section** |
| [project-configuration-visual.md](project-configuration-visual.md) | The `## visual verification` tables, the `mockups` sidecar and the full app suite. **Loaded when `flow.verify` begins** |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |
| [jira-integration-finish.md](jira-integration-finish.md) | The join carve-out and echo, and labels on created issues. **Cited by every command that creates an issue** |
| [workspace-isolation.md](workspace-isolation.md) | Resolve a worktree's own database, cache index, bucket or ports: the workspace id, what it derives, why the cache index is probed rather than derived, the empty id, and creation and cleanup |
| [git-boundaries.md](git-boundaries.md) | Which git actions each command may take, that the change directory is never committed, and the branch backup. **Loaded by `/flow`'s creating run, `/flow`'s implement phase, bare `/flow`, and `/flow-plan`** |
| [git-boundaries-commit-chain.md](git-boundaries-commit-chain.md) | The guarded commit chain that commits the implementation and leaves the gitignored change directory out. **Loaded by bare `/flow`'s run 1, and by `/flow`'s implement phase only when a `prUrl` is recorded** |
| [model-policy.md](model-policy.md) | Which model each role runs on, their defaults, how an override applies, and per-harness enforcement. **Loaded by `/flow`'s creating run and `/flow`'s implement phase** |
| [artifacts-registry.md](artifacts-registry.md) | Every artifact the pipeline creates, with what creates it, where it lives, and what removes it. **Loaded by `/flow`'s cleanup run** |
| [session-records.md](session-records.md) | The outcome table for `flow record render`, and what each outcome means for the caller. **Loaded by `/flow`'s implement phase, and bare `/flow`** |
| [worktree-resolution.md](worktree-resolution.md) | How any step resolves the set of worktrees belonging to a change. **Loaded by `/flow`'s creating run, `/flow`'s implement phase and bare `/flow`** |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins.
