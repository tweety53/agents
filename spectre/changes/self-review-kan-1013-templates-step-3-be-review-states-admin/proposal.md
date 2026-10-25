# self-review-kan-1013-templates-step-3-be-review-states-admin

## Why

kan-1013's self-review pass found five non-`big` defects. Four are real: a plan named a
`**Files:**` path in a directory that does not exist and no guard said so; the run recorded
`execution: inline` and then dispatched seven implementers and a panel-fix subagent anyway; the
operator's session cap (opus at low effort until 2026-10-28) bound only the panel because nothing
records or enforces a session-wide ceiling; and two explorer rows were begun and ended in the same
instant after the agents had already returned, so the ledger shows them as 0s dispatches. The
fifth — the worktree gateway on the shared 8080 — is the documented behaviour of the handed-off
stack (`design.md`, **Context**).

## What changes

- `check-plan-shape` refuses a `**Files:**` path whose directory is missing but exists with one
  non-final segment dropped (the `src/app/app/...` shape).
- `flow record dispatch begin` refuses an `implementer` or `panel-fix` row with a subagent behind
  it while the change's recorded decision says `execution: inline`, and refuses any row whose model
  or effort is above the decision's new `session_cap`.
- flowd refuses a completed `dispatch end` landing under three seconds after its begin on a session
  whose transcript the harvester reads — the after-the-fact begin/end pair.
- kan-1013's records are repaired: its two explorer rows are retimed to their real instants, its
  round-1 pass log gains a correction naming `panel-fix-1`, and its narrative's execution line
  states what ran.
