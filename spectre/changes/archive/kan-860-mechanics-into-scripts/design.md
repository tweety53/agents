# Design — kan-860-mechanics-into-scripts

## Context

The per-group notes in this directory are canonical for each row's shape — interface, exit
contract, parity target, tests and the prompt passage it replaces. This file adds only what the
notes do not say: the scope decision, the operator's answers, and one correction.

| Group | Note | Rows |
|---|---|---|
| Planning | `design-planning.md` | BP-21/22, BP-23, BR-M1, MX4 |
| Implement | `design-implement.md` | MX1, MX3, MX2/OS05, MX7 |
| Panel | `design-panel.md` | RP35, late-fix trigger, `render-slot-prompt.sh`, Placeholders cut |
| Verify | `design-verify.md` | VH-25, VH-24, VV-18 |
| Finish | `design-finish.md` | MI1/MA2, MA1, MR2, FF1 |
| Basesync | `design-basesync.md` | `aside-planning-artifacts` port, base-rebase core, MX6, RP34, sync onto base |

Each note's "WIP — nothing implemented" status and its "Rows left" list describe the tree at
`452e31d7`, when planning stopped. They are superseded by **all-designed-rows-in-scope** below.
Every row they name is planned in `tasks.md`.

Base: `67a082a6` (main after KAN-855–859 landed and were archived). The notes' passage locations
were written against `452e31d7`, whose parent `ae805186` is the KAN-859 tip on main, so they still
resolve.

## Decisions

### Every designed row lands in this one change

**ID:** all-designed-rows-in-scope
**Status:** active
**Chosen:** one change, one plan, every row the six notes design — matches the proposal's scope table.
**Considered:** one change per group — six integrate cycles for one epic child, no parallelism gained
that SDD waves do not already give; one group first — leaves the proposal's scope unmet.

### KAN-855–859 land before KAN-860 starts

**ID:** land-trims-first
**Status:** active
**Chosen:** the cloud branch's KAN-855–859 commits fast-forwarded main (`ae805186`) and their
change directories were archived (`67a082a6`); KAN-860 branches from that — the proposal's premise
("KAN-855–KAN-859 have landed") is then true, and no moved text is edited twice.
**Considered:** stack KAN-860 on the cloud branch — its integrate would land five unrelated
changes; rebase only KAN-860 onto old main — the notes were written against the trimmed files and
would need re-deriving.

### close-task.sh pushes only after both commit guards pass

**ID:** close-task-push-after-guards
**Status:** active
**Chosen:** the push runs after `check-task-commit-planning-paths`, so a refused commit is never
pushed — the one deliberate order change in MX3 (`design-implement.md` item 2).
**Considered:** byte parity with today's order — keeps a push that a later refusal contradicts.

### The visual preflight moves all four checks into Go

**ID:** visual-preflight-all-checks
**Status:** active
**Chosen:** VV-18 as `design-verify.md` designs it, checks 2–3 included with their mechanical
base-URL and origins rules — operator's choice.
**Considered:** checks 1 and 4 only, 2–3 left as prose judgment — less drift risk, less prose cut;
leaving VV-18 out — the recipe stays hand-typed.

### Worktree cleanup moves into one guard

**ID:** remove-change-worktrees-included
**Status:** active
**Chosen:** MR2 as `design-finish.md` designs it: disclose (exit 3) then `--proceed` after the
operator's one ask; the irreplaceable/preserved judgment and the ask stay prose.
**Considered:** leaving cleanup hand-typed — lower risk on a destructive path, but the 4 KB recipe
the issue names stays.

### A failed findings read renders as unknown

**ID:** handoff-deferred-unknown
**Status:** active
**Chosen:** `flow record handoff-lines` prints `**Deferred:** unknown — the findings could not be
read` and still exits 0 — a failed read is never shown as zero deferred.
**Considered:** exit 2 — turns a store hiccup into a handoff the parent must hand-assemble.

### Bash parity tests pin the retired bash at ae805186

**ID:** parity-ref-on-main
**Status:** active
**Chosen:** a parity test that runs a retired bash script reads it from `ae805186` — on main, so it
cannot be garbage-collected — in place of the notes' `452e31d7`, which exists only on the cloud
branch.
**Considered:** `452e31d7` as the notes say — unreachable from main once the cloud branch is deleted.

## Open questions
