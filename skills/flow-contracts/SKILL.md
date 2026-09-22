---
name: flow-contracts
description: The flow pipeline itself plus its contract definitions — the three states and their transitions, the state file shape, project configuration, Jira integration, follow-up issues, plan provenance, the build-green tag, and workspace isolation. Load the one file you need; each is canonical for its own contract, and a run never loads a rationale appendix. Referenced by the stubs in rules/flow-manual-review.mdc.
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
| [pipeline.md](pipeline.md) | **Run any `/flow*` command — load this first.** The three states and what each means, the command→state transition table, the wrong-state handoff, the handoff output shape, IntelliJ commands, guard resolution and stage marks |
| [finish-contract-run1.md](finish-contract-run1.md) | `/flow`'s integrate run: the preflight-signal decision, run 1's procedure, base-branch resolution, and resolving a change's worktrees. **Loaded by `skills/flow/integrate.md` and no other command** |
| [finish-contract-run2.md](finish-contract-run2.md) | `/flow`'s archive run: run 2's procedure and worktree cleanup. **Loaded by `skills/flow/archive.md` and no other command** |
| [handoff-blocks.md](handoff-blocks.md) | The per-state handoff block templates and the rules governing their regeneration: the three per-state templates, the run-only rule, the missing-rather-than-dropped rule and the `IN_PROGRESS` rendering-selection table. **Loaded by `/flow-status` and no other command** |
| [state-file.md](state-file.md) | Read or write a change's state file: its path, its full shape, monotonic state writes, carry-forward |
| [project-configuration.md](project-configuration.md) | Resolve `<project>/.flow/project.md` — every key it defines, including standards-entry resolution and containment, the workspace-isolation and visual-verification tables, and the single-line-literal keys. Resolution rules only — see below for authoring guidance |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |
| [jira-followups.md](jira-followups.md) | File or join a follow-up issue for work a run left outstanding: the naming, the scoped join search, the confirmation, and the three ordered writes a join makes. **Loaded by `/flow`'s integrate run and no other command** |
| [plan-provenance.md](plan-provenance.md) | Write a plan's provenance tags: the four tags, the asymmetry rule, the implementer's duty, and what to do when a measurement contradicts the plan |
| [plan-provenance-guard.md](plan-provenance-guard.md) | What check-plan-provenance.py enforces: the guard's scope, the quotation exemption and its vetoes, what the guard does not do |
| [build-green.md](build-green.md) | Write or check a plan's build-state tags: the tag vocabulary, the merge-partner rule, and the guard's scope |
| [workspace-isolation.md](workspace-isolation.md) | Resolve a worktree's own database, cache index, bucket or ports: the workspace id, what it derives, why the cache index is probed rather than derived, the empty id, and creation and cleanup |
| [git-boundaries.md](git-boundaries.md) | Which git actions each command may take, the guarded two-commit chain that enforces the split between implementation and planning artifacts, and the branch backup. **Loaded by `/flow`'s implement phase, bare `/flow` and `/flow-fast`** |
| [model-policy.md](model-policy.md) | Which model each role runs on, their defaults, how an override applies, and per-harness enforcement. **Loaded by `/flow`'s creating run, `/flow`'s implement phase and `/flow-fast`** |
| [artifacts-registry.md](artifacts-registry.md) | Every artifact the pipeline creates, with what creates it, where it lives, and what removes it. **Loaded by `/flow`'s implement phase, `/flow`'s archive run and `/flow-fast`** |
| [session-records.md](session-records.md) | The outcome table for `flow record render`, and what each outcome means for the caller. **Loaded by `/flow`'s implement phase, bare `/flow` and `/flow-fast`** |
| [worktree-resolution.md](worktree-resolution.md) | How any step resolves the set of worktrees belonging to a change. **Loaded by `/flow`'s implement phase, bare `/flow`, `/flow-status` and `/flow-fast`** |
| [project-configuration-authoring.md](project-configuration-authoring.md) | Write or edit a project's own `<project>/.flow/project.md` — rationale for a key's shape and worked examples not consulted by any run. **Loaded by no `/flow*` command** |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins.

**A `/flow*` run never loads an appendix.**
