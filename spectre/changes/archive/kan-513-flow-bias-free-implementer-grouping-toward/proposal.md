# kan-513-flow-bias-free-implementer-grouping-toward

Source: KAN-513

## Why

Decide step 4's free implementer grouping defaults close to 1:1 with
`plan-dispatch-bundles.sh`'s raw file-overlap bundles even when several bundles have no real
parallel-execution need — bundles are serial by default (`**After:**` absent ⇒ waits on every
plan-earlier task), and only 10 of 88 archived plans use `**After:**` at all. The observed example
— `{1,2} {3,4} {5} {6} {8,9}` — became 5 implementer dispatches that bought zero wall-clock
parallelism (no two of those bundles could ever run concurrently) while paying five context
bundles, five engineering-principles loads and four dispatch boundaries. `implement.md`'s
two-in-flight-per-wave cap already bounds useful splitting, so a third mutually-ready group gains
nothing over being appended to one of the first two.

## What changes

- New `scripts/plan-dispatch-groups.py` (+ `.sh` wrapper), importing `compute_bundles` from
  `plan-dispatch-bundles.py`, computing a mechanical default grouping via chain-merge then
  fold-to-≤2-per-ready-set.
- Decide step 4 (`skills/flow/brainstorm-planner.md`) runs the new script for the mechanical
  grouping; the planner may only **split** a mechanical group (never merge across it), recorded as
  `groups_override` with a one-line reason.
- `implement.md`'s Waves paragraph gains a clause excluding a group's own members from the ids it
  waits on (the self-wait case, now the norm with merged groups).
- Brainstorming's interactive checklist and its convergence-and-approval confirm are both skipped
  when the found staging note is **fully seeded** (note + plan + decision, all three present) —
  `skills/flow/brainstorm-planner.md`, `skills/flow-contracts/pipeline.md` and
  `skills/flow/brainstorm.md`'s mark text. A partial seed still runs the full checklist.
- The adopted staging note (`docs/superpowers/research/kan-513.md` and its `kan-513/` plan
  directory) is deleted once this change's own artifacts have adopted its content.
