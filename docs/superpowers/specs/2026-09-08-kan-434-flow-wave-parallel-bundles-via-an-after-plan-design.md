# kan-434 — wave-parallel bundles via an `After:` plan field — design

Date: 2026-09-08. Change: `kan-434-flow-wave-parallel-bundles-via-an-after-plan` (KAN-434).
Source: `docs/superpowers/research/flow-speedup.md` wall-clock lever 1, decided in its section 8.
Canonical artifact form: this change's `spectre/changes/<name>/design.md` (Decisions and Open
questions live there); this document records the design presented at the approval gate.

## Problem

`implement.md` §4 allows at most one implementer in flight per worktree, so file-disjoint dispatch
bundles run serially even when they cannot conflict. A quarter of recent changes are 6–9 such
bundles (kan-372, kan-374, kan-377, kan-378, kan-379); each pays one implementer-length plus a
~40-second boundary for no reason the bundling rule imposes. Parallelism cannot be derived from
`**Files:**` disjointness alone — two file-disjoint tasks may still depend on each other's symbols
— so independence must be declared.

## Design

### The field

A task body may carry one optional `**After:**` field:

```markdown
**After:** Task 2, 3
```

or `**After:** none`. Grammar mirrors `**Squash-with:**` exactly: `AFTER_FIELD_RE` captures the
line's remainder, `AFTER_VALUE_RE` gates `Task <id>[, <id>...]` or the literal `none`, `after_ids`
extracts ids only from a gated value, and `select_after` picks the first non-fenced gating line,
line-scoped, with a non-gating first candidate returned for the malformed report. All four live in
`scripts/lib/plan_grammar.py` beside the `SQUASH_WITH_*` definitions; no guard carries its own
copy. **Absent field = every earlier task.** The default is applied where after-sets are resolved
(`plan-dispatch-bundles.py`), never in the grammar, so "serial by default" has one definition.

### Resolution and output

Each bundle's after-set is the union over its member tasks: a declared `Task <ids>` contributes
its ids; a member with no field contributes every plan-order earlier task id; a member with `none`
contributes nothing (a `**Squash-with:**` red pair therefore merges both members' declarations).
`plan-dispatch-bundles.py` prints one `after <k>: <ids>` line (ids numerically sorted, or `none`)
beneath each unchanged `bundle <k>: <ids>` line. The script computes; it does not validate
(dangling ids and cycles are shape findings) and does not schedule (that is the conductor's rule).

### Validation

`check-plan-shape.py` gains four findings, canonical in its docstring: F7, a second gating
`**After:**` line in one body (the selector keeps the first; the second would be silently
ignored); F8, a non-gating value; F9, an id naming no task in the plan; F10, a cycle over the
**resolved** after-graph with defaults materialized — so `2 After: 5` with `5` default-serial is
caught, and self-reference folds into the same DFS. `check-task-commit-fields.py`'s `FIELD_RE`
alternation gains `After` so the field parses as a field boundary and the F3a indented-field
finding covers indented copies; `plan-dispatch-bundles.py`'s `ANY_FIELD_RE` gains it too. The
field is optional everywhere: a plan without it validates identically to today.

### Wave dispatch

Readiness: a bundle is ready when every id in its `after <k>:` line has landed (committed and
guard-passed, whether by direct commit or by pick).

- **Singleton wave** — dispatch into the canonical worktree exactly as today: the commit lands
  directly, no throwaway, no pick. A no-field plan produces only singleton waves and its dispatch
  mechanics are byte-identical to today, which is what makes the compatibility invariant
  structural.
- **Shared wave** — one message launches every ready bundle's implementer, each into its own
  throwaway worktree created by the review-panel throwaway-worktree sequence
  (`git -C <worktree> worktree add --detach <worktree>-wave-bundle-<k> HEAD`, staged/untracked
  transplant), after which the project's resolved `## worktree setup` command runs once in the
  copy (this repository: `cd stats && make web-build` — a fresh worktree's embedded-SPA
  `go test` fails without it). Per-bundle gather is unchanged.
- **Picks** — as members return, each is cherry-picked onto the change branch in plan order: a
  member is picked once every plan-earlier member of its wave is picked. The unchanged
  `check-task-commit-fields.sh` call runs on each picked commit and the dispatch `end` records the
  picked sha. A pick conflict or guard failure hands the bundle back to its own implementer, its
  throwaway worktree rebased onto the advanced branch HEAD, for re-commit and re-pick; copies are
  removed once picked, or after handback resolves.
- **Full suite** — a singleton plan-last bundle keeps implementer-carried `FULL SUITE` unchanged.
  When the plan-last bundle belongs to a shared wave, the conductor runs the resolved `## test`
  list once on the canonical worktree after that wave's final pick passes the guard; a failure is
  the same verbatim-output `## Question` handback as today. Same serial wall clock as today's
  last-bundle run; the panel-on-green guarantee holds.
- A member failing pre-commit (BLOCKED) follows the existing handback; siblings are unaffected and
  dependent later waves simply stay unready. Turn discipline is unchanged: one bounded foreground
  wait per report file, never a turn ended with a wave child in flight.

### Plan template

`brainstorm-planner.md` §D's field family gains `**After:**`: optional, `Task <ids>` or `none`,
absent = serial; annotate file-disjoint tasks with no caller/helper relationship, and annotate a
`**Squash-with:**` pair consistently.

## Testing

- `test-plan-dispatch-bundles.sh`: declared sets, default expansion, `none`, red-pair union,
  resumed-plan checked-task handling.
- `test-check-plan-shape.sh`: F7–F10 each; a clean `After:`-using plan; an absent-field plan
  staying clean.
- `test-check-task-commit-fields.sh`: an `**After:**` line adjacent to a `**Files:**` bullet run
  leaves Files parsing intact.
- `check-contract-budget.sh` measured after the doc edits, rows raised only if tripped; the
  normative inventory captured before and after the `implement.md` prose edits, differences
  restored.
