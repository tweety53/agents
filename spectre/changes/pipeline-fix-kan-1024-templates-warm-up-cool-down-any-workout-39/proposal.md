# pipeline-fix-kan-1024-templates-warm-up-cool-down-any-workout-39

## Why

`scripts/close-task.sh` runs close-task in-process, and close-task checks a committed
`full-app-suite.zip` by calling `Registry["rebuild-full-app-suite-zip"]` in the same process
(`stats/internal/guard/closetask.go`). That rebuild needs `FLOW_GUARD_SCRIPT_DIR` to find
`resolve-visual-screenshots.sh` (`stats/internal/guard/rebuildfullappsuitezip.go`), and only
`scripts/rebuild-full-app-suite-zip.sh` exports it. So every real close of a task committing the
zip stops on exit 2. gymie kan-1024 Task 14 (playwright commit 9f063188):
`rebuild-full-app-suite-zip: FLOW_GUARD_SCRIPT_DIR is unset` /
`close-task: task 14: rebuild-full-app-suite-zip exited 2 — cannot judge full-app-suite.zip` /
`close-task: STOP`; the same run with `FLOW_GUARD_SCRIPT_DIR` set passed. The Go tests inject
the variable into the guard's env, so none exercised the shim.

## What changes

- `scripts/close-task.sh` exports `FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"`, as
  `rebuild-full-app-suite-zip.sh`, `integrate-change.sh` and `run-verify-list.sh` do.
- `stats/internal/guard/close_task_test.go` gains a test that runs the real shim on a hand-built
  zip with no `FLOW_GUARD_SCRIPT_DIR` in its environment and wants exit 1 (rebuild ran, differed).
- No repair: the defect wrote nothing; it only stopped closes, which the operator re-ran with the
  variable set.
