# kan-991-panel-effort-roll-high-gated

## Why

Decide gives every pass-1 review-panel dispatch `medium` effort whenever `effort_roll < 90`, and
above it lets the planner choose freely, `high` included (`skills/flow/brainstorm-planner.md`,
**Model and effort**, checked against base `bc79a4d6`). A hard-seam panel dispatch may take `high`
on the planner's say alone. Panels therefore run costlier than most changes need, and `high`
reaches a run without the operator ever seeing it. Separately, brainstorming asks the operator
about display-only formatting — numeric precision, empty-cell placeholders — which nobody needs to
weigh in on (KAN-991).

## What changes

- Every pass-1 panel dispatch, and every gated per-task reviewer bundle, takes its effort from the
  existing `effort_roll`: `low` below 60, `medium` at 60 or above — 60/40 per change.
- `high` reaches a panel dispatch only by the operator's explicit approval at the plan gate, or on
  a fix run after a lower-effort attempt failed to solve the issue.
- Brainstorming decides display-only formatting itself and records the choice; it never asks.
