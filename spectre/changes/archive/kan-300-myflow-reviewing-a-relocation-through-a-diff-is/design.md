## Context

Full problem statement, alternatives and out-of-scope items are in
`docs/superpowers/specs/2026-08-30-kan-300-myflow-reviewing-a-relocation-through-a-diff-is-design.md`
(committed to `main` before this change's branch existed, per this pipeline's own convention).
This file carries the decisions and open questions the brainstorming round settled, not a second
telling of the problem.

## Decisions

### The comparison unit is a passage, not a sentence

**ID:** relocation-comparison-unit-is-passage
**Status:** active
**Chosen:** paragraph/bullet/table-row granularity, matching `myflow-contract-economy`'s per-move
ledger — one-line rationale: the two artifacts audit the same kind of move and should speak the
same unit, and passage granularity is what the ledger's own scenarios already test against.
**Considered:** true sentence-level splitting (the ticket's own title wording) — rejected because
it would produce a second, incompatible unit alongside the ledger's own passage unit, doubling the
granularity a reviewer has to reconcile.

### Declaration is plan-level, not per-task

**ID:** relocation-declaration-is-plan-level
**Status:** active
**Chosen:** one `**Relocation:** yes|no` line in `tasks.md`'s header, required and explicit on
every plan.
**Considered:** a per-task field naming source/destination files — rejected as heavier to author
and check than the ticket's motivating case needs; per-task specificity is unnecessary because the
comparison already scopes itself from every task's existing `**Files:**` field.

### Generation never blocks the panel

**ID:** relocation-comparison-never-blocks
**Status:** active
**Chosen:** a generation failure (unreadable merge-base blob, empty scope, malformed header)
prints one line and the panel dispatches without the file — the same fallback shape
`dispatch-context.md`'s own missing-bundle case already uses.
**Considered:** stopping the run on a generation failure — rejected; the comparison is a review
aid, and a plan wrongly declared `**Relocation:** yes` should not be able to block the panel
outright.

## Open questions

<!-- none — the brainstorming round converged with an explicit approval and no deferred items -->
