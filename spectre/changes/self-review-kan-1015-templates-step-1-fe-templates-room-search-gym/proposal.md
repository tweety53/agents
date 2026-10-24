# self-review-kan-1015-templates-step-1-fe-templates-room-search-gym

## Why

The self-review pass of `/flow` change `kan-1015-templates-step-1-fe-templates-room-search-gym`
found fifteen fixable pipeline defects, one product defect its sweep must fix in this run, and one
wrong cause recorded in the change's archived narrative. Each pipeline defect let the run go wrong
without any gate noticing: panel findings never reached the store, dispatches ran with no ledger
row, appended tasks skipped the plan guards, a verifier attributed a departure as pre-existing
without a merge-base capture and re-baselined it, and a scoped re-run ran the whole suite.

## What changes

- The integrate gate (`check-unfinished-work.sh`) refuses a panel report that counts more
  Critical and Important findings than the store holds, as `check-panel-findings-closed.sh`
  already does; the change's 13 missing finding rows are recorded and its panel record re-rendered.
- `hooks/enforce-agent-baseline.py` denies a flow dispatch whose row has no open begin; the
  change's 15 unrecorded dispatches are back-filled.
- `check-planning-commit-location.sh` runs `check-plan-shape` and `check-task-build-green` on a
  commit that appends a task; the archived plan's tasks 15–24 are repaired.
- `check-plan-shape`'s F11 accepts an ordering reachable through the After graph.
- The PIXEL PROBE paragraph no longer accepts a geometry assertion in place of rendered pixels.
- `/flow`'s implement stage runs `check-base-behind.sh` before task 1.
- The verifier may start the merge-base stack its attribution needs, and `check-verify-report.sh`
  refuses a pre-existing attribution with no merge-base evidence.
- A visual-fix task runs its capture spec before it closes, and measures on the 2× pair, never a
  composite.
- `hooks/protect-main-checkout.py` denies a commit on a `spectre/<name>` branch with no
  `Task-Id:`/`Flow-Key:` trailer and no `chore` subject.
- The closed dispatch list names the verifier's `-rest` and `-rest-<n>` keys.
- `measure-visual-properties.sh --props elements` finds the app surface on a scrim frame and
  emits the per-element pairs the verifier's matrix needs.
- The panel record's pass log carries a verdict line derived from the finding rows.
- A verifier row records its round base (`-round-base`), and `check-visual-verify-dispatched.sh`
  refuses a change whose newest verifier ran scoped.
- `check-verify-report.sh` refuses a scoped re-run's report whose mapped-spec count differs from
  what `specs` prints for the round base.
- The kan-466 weight-tab "Log a weigh-in" button is traced at the merge base, its cause fixed, and
  its correct baseline restored.
- The archived kan-1015 narrative states the T2 Use fill's real cause.
