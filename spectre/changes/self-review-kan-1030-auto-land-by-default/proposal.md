# self-review-kan-1030-auto-land-by-default

## Why

kan-1030's self-review pass found four non-`big` defects, F1–F4:

- **F1** (flow-fix) — `skills/flow/review-panel.md` runs `sync-panel-base.sh` before the
  before-review planning commit, so the uncommitted ticks in the change's directory make git refuse
  the rebase and the guard exit 2.
- **F2** (flow-fix) — the land-straight-away path of `skills/flow/verify-and-handoff.md` prints no
  block, so a run's `?` open questions and 🤖 auto-resolutions never reach the operator, and
  integrate's handoffs carry no `Decisions:` field to catch them.
- **F3** (flow-fix) — kan-1030's archived `narrative.md` says no scratch `flowd` was built, against
  its own `live-verification.md`, and names `close-task.sh` missing from the worktree, which held it.
- **F4** (flow-speed) — `flow.run-instructions` rebuilds and restarts the project stack on a run that
  lands straight away, and a refused start ends that run at `IN_PROGRESS`, blocking a landing no
  reviewer waits for.

## What changes

- F1: the before-review planning commit opens **Check base movement first**, at panel entry and at
  every round boundary; `git-boundaries.md`'s boundary row names the base sync.
- F2: a run leaving a `?` line stops at `IN_PROGRESS` with its block whatever `reviewBeforeLanding`
  holds; one with only 🤖 lines prints them before continuing into integrate. `pipeline.md` and
  `skills/flow/SKILL.md` state the stop where they state the flag's effect.
- F3: the two narrative lines are corrected, in a planning commit of their own.
- F4: a run that will land straight away skips the stack start, its freshness check and the
  refused-start stop.
