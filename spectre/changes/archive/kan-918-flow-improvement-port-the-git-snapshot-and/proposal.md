# kan-918-flow-improvement-port-the-git-snapshot-and

## Why

- Every new guard and every extension of one is Go in `flow-guard` (`.flow/project.md`), so a guard
  still in bash must be ported before it can be extended; each port also takes a bash harness off
  `scripts/run-guard-tests.sh`.
- Five guards protecting a worktree during review and fix dispatches are still bash (KAN-918).

## What changes

- Five scripts run as Go ports inside `flow-guard`: `check-plan-unchanged`, `check-tree-markers`,
  `guard-autosquash`, `check-panel-diff-size`, `check-worktree-processes` — each CLI contract
  unchanged: arguments, output lines, stdout/stderr split, side effects, exit codes.
- Each `scripts/<name>.sh` becomes a `flow_guard_exec` shim keeping its header; its bash harness is
  replaced by an in-process Go test, every case ported.
- Before/after suite timings are recorded in `design.md`.
