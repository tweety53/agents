# pipeline-fix-kan-1051-rework-the-trend-dashboard-53

## Why

`check-unfinished-work` (`stats/internal/guard/unfinishedwork.go`, signal four's commit half) counts
an appended task as committed when it is already ticked in the plan at a dropped re-run's range base,
read by `git show <base>:<plan path>`. After integrate's drop the range base is the drop commit itself
(`archivedBase`, `stats/internal/guard/integratechange.go`), which removed the change directory, so
the `show` fails and every appended task ticked before the drop is reported as having no
`Task-Id:` commit. kan-1051's `integrate-change.sh prepare` stopped on
`STOP: unfinished-work … (Task 8, 9)`: `git show dffb70df:./spectre/changes/kan-1051-rework-the-trend-dashboard/tasks.md`
exits 128 while `dffb70df^` holds the plan with Tasks 1–9 ticked, and a false-positive verdict does
not pass the gate.

## What changes

- When the plan is absent at the range base, the guard reads it at the base's parent.
- One test case whose drop commit really removes the plan; it is OUTSTANDING without the fix. The
  existing case passes only because its fake drop commit still holds the plan.
