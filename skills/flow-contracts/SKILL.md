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

flow's pipeline and its contract definitions, split out of `rules/flow-manual-review.mdc` so the
always-on rule layer carries only the trigger, without being asked to load anything.

**Load the one file you need — not this whole directory.** The exception is `pipeline.md`: every
`/flow*` command needs it, so load it first, always.

## Index

| File | Load it when you need to |
|------|--------------------------|
| [pipeline.md](pipeline.md) | **Run any `/flow*` command — load this first.** The three states and what each means, the command→state transition table, the wrong-state handoff, the handoff output shape, IntelliJ commands, guard resolution and stage marks |
| [finish-contract-run1.md](finish-contract-run1.md) | `/myflow-finish`'s run 1: the preflight-signal decision, run 1's procedure, base-branch resolution, and resolving a change's worktrees. **Loaded by `skills/flow/integrate.md` and no other command** |
| [finish-contract-run2.md](finish-contract-run2.md) | `/myflow-finish`'s run 2: run 2's procedure and worktree cleanup. **Loaded by `skills/flow/archive.md` and no other command** |
| [handoff-blocks.md](handoff-blocks.md) | The per-state handoff block templates and the rules governing their regeneration: the three per-state templates, the run-only rule, the missing-rather-than-dropped rule and the `IN_PROGRESS` rendering-selection table. **Loaded by `/flow-status` and no other command** |
| [state-file.md](state-file.md) | Read or write a change's state file: its path, its full shape, monotonic state writes, carry-forward |
| [project-configuration.md](project-configuration.md) | Resolve `<project>/.flow/project.md` — apps, run, stop, test, lint, standards, jira — including standards-entry resolution and containment. Resolution rules only — see below for authoring guidance |
| [jira-integration.md](jira-integration.md) | Resolve a linked issue, transition it, or sync its description |
| [jira-followups.md](jira-followups.md) | File or join a follow-up issue for work a run left outstanding: the naming, the scoped join search, the confirmation, and the three ordered writes a join makes. **Loaded by `/myflow-finish` run 1 and no other command** |
| [plan-provenance.md](plan-provenance.md) | Write a plan's provenance tags: the four tags, the asymmetry rule, the implementer's duty |
| [plan-provenance-guard.md](plan-provenance-guard.md) | What check-plan-provenance.py enforces: the guard's scope, the quotation exemption and its vetoes, what the guard does not do |
| [build-green.md](build-green.md) | Write or check a plan's build-state tags: the tag vocabulary, the merge-partner rule, and the guard's scope |
| [workspace-isolation.md](workspace-isolation.md) | Resolve a worktree's own database, cache index, bucket or ports: the workspace id, what it derives, why the cache index is probed rather than derived, the empty id, and creation and cleanup |
| [git-boundaries.md](git-boundaries.md) | Which git actions each command may take, and the guarded two-commit chain that enforces the split between implementation and planning artifacts. **Loaded by `/myflow-do`, `/myflow-finish` and `/myflow-fast`** |
| [model-policy.md](model-policy.md) | Which model each role runs on, their defaults, how an override applies, and per-harness enforcement. **Loaded by `/myflow-start`, `/myflow-do` and `/myflow-fast`** |
| [artifacts-registry.md](artifacts-registry.md) | Every artifact the pipeline creates, with what creates it, where it lives, and what removes it. **Loaded by `/myflow-do`, `/myflow-finish` run 2 and `/myflow-fast`** |
| [session-records.md](session-records.md) | The outcome table for `flow record render`, and what each outcome means for the caller. **Loaded by `/myflow-do`, `/myflow-finish` and `/myflow-fast`** |
| [worktree-resolution.md](worktree-resolution.md) | How any step resolves the set of worktrees belonging to a change. **Loaded by `/myflow-do`, `/myflow-finish`, `/flow-status` and `/myflow-fast`** |
| [project-configuration-authoring.md](project-configuration-authoring.md) | Write or edit a project's own `<project>/.flow/project.md` — rationale for a key's shape and worked examples not consulted by any run. **Loaded by no `/flow*` command** |

Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
file wins — and a skill should **point at** these files rather than restate them, because a second
copy of a procedure drifts even when nobody edits it wrongly.

## Rationale appendices

Where a contract is large enough that its reasoning outweighs its rules, that reasoning is split
out of the file a command loads. The table below is the list — like the contract table above it,
this sentence carries no count, because a count goes stale the first time an appendix is added and
nothing checks it:

| Appendix | Holds the reasoning behind |
|----------|----------------------------|
| [pipeline-rationale.md](pipeline-rationale.md) | [pipeline.md](pipeline.md) |
| [git-boundaries-rationale.md](git-boundaries-rationale.md) | [git-boundaries.md](git-boundaries.md) |
| [model-policy-rationale.md](model-policy-rationale.md) | [model-policy.md](model-policy.md) |
| [artifacts-registry-rationale.md](artifacts-registry-rationale.md) | [artifacts-registry.md](artifacts-registry.md) |
| [session-records-rationale.md](session-records-rationale.md) | [session-records.md](session-records.md) |
| [worktree-resolution-rationale.md](worktree-resolution-rationale.md) | [worktree-resolution.md](worktree-resolution.md) |
| [handoff-blocks-rationale.md](handoff-blocks-rationale.md) | [handoff-blocks.md](handoff-blocks.md) |
| [jira-integration-rationale.md](jira-integration-rationale.md) | [jira-integration.md](jira-integration.md) |
| [project-configuration-rationale.md](project-configuration-rationale.md) | [project-configuration.md](project-configuration.md) |
| [workspace-isolation-rationale.md](workspace-isolation-rationale.md) | [workspace-isolation.md](workspace-isolation.md) |

**A `/flow*` run never loads an appendix.** They exist for whoever *edits* a contract — the
justification of an ordering, the alternatives that were rejected, the history, and the measurements
behind a rule. A command that loads one has paid the cost the split exists to remove, which is why
no appendix appears in the **Index** above: that index is what a command reads, and this section is
for the editor.

**Nothing enforces this.** It is a judgment rule, like the others this corpus states and does not
guard.

An appendix carries the **same heading tree as its core**, in the same order, so a section's
reasoning is found under the heading it belongs to, and a section that is wholly normative leaves
its appendix heading present with no body — present rather than absent, so it is visible that the
section was examined rather than skipped.

**The same split applies one level up.** `skills/myflow-do/`, `skills/myflow-start/` and
`skills/myflow-finish/` each carry a `SKILL-rationale.md` that a `/flow*` run never loads — the
reasoning behind that skill's `SKILL.md`, kept beside it. **A skill's appendix lives beside its own
`SKILL.md`, not under `skills/flow-contracts/`** — this directory indexes only the contracts' own
appendices, so it never appears in the table above or anywhere in this directory's listing.

## Keeping this index honest

This file is the entry point to the contracts, so a stale entry here misdirects every command that
starts from it. When a contract file gains or loses a section, update the row above in the same
change — `<agents repo>/scripts/check-references.sh` catches a **named** section that no longer exists, but it
cannot catch a description that is merely out of date.
