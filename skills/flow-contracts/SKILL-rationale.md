# flow-contracts — rationale

Reasoning behind `skills/flow-contracts/SKILL.md`: where the index came from, how it is kept honest,
and the appendix convention the contract files follow. Moved here verbatim from the run-loaded
file; **no `/flow*` run loads this file.** Each heading names the source file and the section the
passage came from.

## SKILL.md — preamble

> flow's pipeline and its contract definitions, split out of `rules/flow-manual-review.mdc` so the
> always-on rule layer carries only the trigger, without being asked to load anything.

## SKILL.md — Index

> Each file is **canonical** for its own contract. Where a skill and one of these files disagree, the
> file wins — and a skill should **point at** these files rather than restate them, because a second
> copy of a procedure drifts even when nobody edits it wrongly.

## SKILL.md — Rationale appendices

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
| [plan-provenance-guard-rationale.md](plan-provenance-guard-rationale.md) | [plan-provenance-guard.md](plan-provenance-guard.md) |
| [workspace-isolation-rationale.md](workspace-isolation-rationale.md) | [workspace-isolation.md](workspace-isolation.md) |

They exist for whoever *edits* a contract — the
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

**The same split applies one level up.** `skills/flow/` carries a `SKILL-rationale.md` that a
`/flow*` run never loads — the reasoning behind that skill's files, kept beside them. **A skill's appendix lives beside its own
`SKILL.md`, not under `skills/flow-contracts/`** — this directory indexes only the contracts' own
appendices, so it never appears in the table above or anywhere in this directory's listing.

## SKILL.md — Keeping this index honest

This file is the entry point to the contracts, so a stale entry here misdirects every command that
starts from it. When a contract file gains or loses a section, update the row above in the same
change — `<agents repo>/scripts/check-references.sh` catches a **named** section that no longer exists, but it
cannot catch a description that is merely out of date.
