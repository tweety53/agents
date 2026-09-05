# kan-451-add-a-per-project-incident-table-and-a-guard

## Why

KAN-451 is a `flow-stats-app` self-review finding from KAN-423. Two things happened during that
change's run that the pipeline had no record of afterwards:

1. `check-task-commit-fields.sh`'s Baseline revert/test/reset cycle was re-entered mid-flight and
   wiped the untracked planning directory, costing about 55 minutes of hand recovery (KAN-438).
   <!-- measured: KAN-438's own description, "~55 minutes lost recovering by hand" @ Jira KAN-438 — a live-event figure, not re-runnable -->
   The same hazard was already written in project memory before the run, and the first conductor
   dispatch did not carry it — a memory note nobody re-reads before dispatching. Nothing in the
   store counts how often a guard has hurt a project, so the next dispatch on that project starts
   with the same blank slate.
2. `check-unfinished-work.sh` reported `OUTSTANDING` on the frontend worktree because the
   planning directory lives in the backend repo only (KAN-439). The verdict was hand-verified
   structural — 19/19 tasks ticked, `flow record findings` answering `[]` — and overridden. The
   same false positive has been adjudicated by hand on KAN-393, KAN-260, KAN-164, KAN-151 and
   KAN-81; no row anywhere says so, so every recurrence is rediscovered from scratch.

The store already holds every dispatch and every panel finding as rows (`dispatches`, `findings`,
migration `0010_run_records.sql`) precisely so that "the record is somewhere I did not look" stops
being a failure mode. Guard verdicts and incidents are the two records still living only in
memory and chat.

## What changes

- The store gains two tables: `guard_verdicts` — every verdict a guard reached, with a
  false-positive flag and reason — and `incidents` — a per-project row of guard, symptom,
  recovery taken and minutes lost.
- `flow record` gains five verbs: `verdict`, `verdict false-positive`, `verdicts`, `incident`,
  `incidents`. The writes take the same never-block journal path every other record write takes;
  the reads fail loudly like `findings`.
- `check-unfinished-work.sh` records its own verdict on every run, and on `OUTSTANDING` prints one
  advisory stderr line naming how many prior false positives this guard has on this project and
  the last recorded reason. Its verdict line, exit codes and their meanings do not change.
- `gather-dispatch-context.sh` carries a `## incidents` section, so every conductor and panel
  dispatch on a project is handed that project's incidents without anyone re-reading memory.
- `skills/flow/integrate.md`'s unfinished-work gate shows the prior-false-positive line with the
  breakdown, and its **Continue** course records the false-positive flag with the operator's
  reason when they say the verdict was verified structural.
- The API exposes the two tables under `/api/v1/verdicts/{project}` and
  `/api/v1/incidents/{project}`. No SPA view.
