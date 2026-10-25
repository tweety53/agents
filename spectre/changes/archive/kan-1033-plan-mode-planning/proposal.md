# kan-1033-plan-mode-planning

## Why

- Flow planning (`/flow`'s creating run, `/flow-plan`, and the self-review and in-run fix
  planners) does every read itself and dispatches nothing before the plan gate. Wide investigation
  therefore fills the planning session's own context with file dumps, and a design is drawn from one
  line of thought.
- Claude Code's plan mode has practices flow planning lacks:
  - parallel Explore agents that bring back conclusions only;
  - Plan agents that design the change from different angles before one design is recommended;
  - a plan that always names the critical files, the existing code to reuse and how the change is
    verified end to end;
  - a harness-enforced read-only phase that ends in an approve-or-reject-with-feedback step.
- The operator wants flow planning to be its current practices plus those, with nothing removed.

## What changes

- All three planning entry points may spawn fresh subagents through the Agent tool during planning.
  A `fork` stays banned (it copies the whole conversation and saves no context).
- Before the first design question, 1–3 Explore agents investigate in parallel. For a non-trivial
  change, 1–3 Plan agents then each design it from a different angle (simplest, most robust, best
  fit with existing patterns). The session recommends one design and records why the others were
  rejected. The self-review and in-run fix planners auto-pick, as they do every other choice.
- Every Explore and Plan agent runs at `low` effort on `sonnet` (40%) or `opus` (60%), rolled
  reproducibly per dispatch by a new `plan-agent-pair.sh`.
- `design.md` gains four sections: `## Recommended approach`, `## Critical files`,
  `## Existing code to reuse` and `## Verification` (end to end). `check-plan-shape.sh` enforces
  them on every change whose `tasks.md` declares the new shape.
- On Claude Code, `/flow`'s creating run and `/flow-plan` enter real plan mode (`EnterPlanMode`)
  for investigation and design. `ExitPlanMode` is the design approval (approve, or reject with
  feedback for another round). Every existing gate stays, the plan review gate ("Push artifacts?")
  included.
- The self-review and in-run fix planners run on new `flow-planner-low` / `flow-planner-medium` / `flow-planner-high`
  agent definitions that carry the Agent tool. Implementers, fixers and reviewers keep their
  no-delegation definitions.
