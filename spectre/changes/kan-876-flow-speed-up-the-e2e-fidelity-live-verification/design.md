## Context

- Planner rules live in `skills/flow/brainstorm-planner.md` section D; the "Write a live-verification
  task" paragraph puts a final task in every plan that touches runtime state.
- `plan-dispatch-groups.py`'s chain merge joins a bundle to the group holding what it waits on, so
  sibling verification bundles waiting on one feature task, or on each other, run in series in one
  group; Decide step 4 already lets the planner split a mechanical group for a parallel wave, and
  `sdd-dispatch.md`'s Waves already give each wave member its own throwaway worktree.
- `flow.verify` (`skills/flow/verify-and-handoff.md`) starts nothing; visual verify starts a stack
  only when nothing answers and stops it at its step 13; `flow.run-instructions` rebuilds and
  restarts every app on every run.
- One change: the planner rules, the stack reuse and the verify-stage live check only pay off
  together — fanning out tasks that each start a stack, or dropping the live task with nowhere to
  run its check, would make things worse.

## Decisions

### Verification tasks fan out

**ID:** verification-tasks-fan-out
**Status:** active
**Chosen:** end-to-end spec and fidelity-capture tasks declare `**After:**` only on the feature
tasks they exercise and seed their own data; Decide splits each such bundle into its own group, so
they run as one wave against one stack.
**Considered:** teaching `plan-dispatch-groups.py` to detect verification bundles — it cannot tell a
spec task from a feature task from `tasks.md` alone, and the planner's split override already
exists; keeping one serial group at `low` effort — what KAN-754 did, the slow path this ticket names.

### One stack per run

**ID:** stack-started-once
**Status:** active
**Chosen:** the parent starts the canonical worktree's stack once, before the first verification
task, and every later consumer — those tasks, `flow.verify`'s live check, visual verify,
`flow.run-instructions` — reuses it while `check-dev-stack-fresh.sh` reports it fresh.
**Considered:** each task starting its own stack — the KAN-754 behaviour; a new project-configuration
key for a shared stack — `## stop`/`## run` plus the fingerprint already say everything needed;
visual verify leaving the stack it started running — its `start` row may name a disposable stack
other than the one `## run` hands off (this repository's UI-test stack on port 4174), which must
still be stopped.

### The live check runs in verify

**ID:** live-check-in-verify
**Status:** active
**Chosen:** no live-verification task; the planner writes `design.md`'s `## Live check` (what to
exercise, before/after figures, what failure looks like, or one line saying there is no runtime) and
`flow.verify` runs it with the parent's own Bash calls, writing `<changeRoot>/live-verification.md`;
a failure look takes the verify fix loop.
**Considered:** folding it into the visual-verify verifier — a change touching runtime state but no
UI path never reaches that stage; a separate live-check dispatch — the figures are API and database
reads the parent's own calls make, and a browser walk is already visual verify's.

### No production waits in specs

**ID:** no-production-waits-in-specs
**Status:** active
**Chosen:** a spec depending on a production timeout, backoff or polling interval plans a test-only
override that shortens it, and an independent suite runs on the runner's parallel workers — planner
guidance only.
**Considered:** editing gymie's client timeout and Playwright config here — gymie's main already runs
`fullyParallel` on five workers and carries no spec that waits out a timeout; the one that does
(`a timed out replay is not applied twice`) exists only on KAN-754's unmerged branch, so the operator
routed its test-only timeout override to a KAN-754 fix run (2026-10-05).

## Open questions

## Live check

None — this change edits pipeline prose only; no service, daemon or store changes behaviour.
