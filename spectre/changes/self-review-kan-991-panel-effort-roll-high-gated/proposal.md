# self-review-kan-991-panel-effort-roll-high-gated

## Why

The self-review pass over kan-991 found five non-`big` defects, each held against the tree
(`design.md`, **Context**):

- kan-991's plan cited a guard retired in `52f9660b` in a verify step, and every section-D guard
  passed it: `check-plan-shape` checks fields and shape only, `check-references.sh` only Markdown
  section references.
- The docs-only panel reduction was refused for kan-991 because `check-panel-docs-only` treats the
  change's own `verbatim-moves.txt` — the acknowledgement list `check-verbatim-moves.sh` reads —
  as code.
- `brainstorm-planner.md` and `implement.md` let the gated per-task reviewer bundle take `high`,
  but the decision JSON records no pair for that bundle, so a raised effort has nowhere to live.
- `flow record pass` accepts any non-empty `-note`; kan-991's pass log carries two `x` rows.
- kan-991's plan left `skills/flow/SKILL-rationale.md` out of `**Files:**` though the reasoning it
  removed lived there; no planning rule says to list a rationale appendix.

## What changes

- `check-plan-shape.sh` also runs `flow-guard plan-script-paths`, which refuses a live plan's task
  step citing a `scripts/…` path the repository does not carry and no task's `**Files:**` names.
- `check-panel-docs-only` counts a `<spec-root>/changes/<change>/verbatim-moves.txt` as
  documentation; `review-panel.md` and the script header say so.
- The gated per-task reviewer bundle runs at the panel roll's effort and never `high`; the
  plan-gate `high` question covers `panel.dispatches` entries only.
- `flow record pass` refuses a one-word `-note`.
- Section D of `brainstorm-planner.md` tells the planner to list and update the `*-rationale.md`
  appendix recording a changed rule's reasoning.
