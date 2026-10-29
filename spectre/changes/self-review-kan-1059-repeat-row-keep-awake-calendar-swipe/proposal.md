# self-review-kan-1059-repeat-row-keep-awake-calendar-swipe

## Why

The self-review pass of `kan-1059-repeat-row-keep-awake-calendar-swipe` found seven fixable defects:

- 1a: `check-panel-fix-single-dispatch` refuses a panel-fix row whose commit carries a
  `Flow-Key: panel-fix-<r-1>` trailer — the very repair `check-unfinished-work` signal eleven
  prescribes — and nothing catches the one-round-low trailer while the commit can still be amended
  (kan-1044: three, kan-1059: two).
- 1b: a panel re-review's `flow record dispatch begin` is admitted while the fix round's
  `fix-mutation:` rows are unrecorded; only `flow record mutation` polices the order, after the fact
  (twice "late mutation proof" in kan-1059).
- 1c: `skills/flow/unfinished-work-gate.md` closes a verify failure that does not recur as a guard
  false positive, which `skills/flow/verify-fix-loop.md` forbids ("An observed failure is never
  closed as 'not reproduced'").
- 1d: gymie-playwright `fidelity/kan-932-forms.spec.ts`'s `onPausedClock` bounds a paused-clock wait
  by a 15 s wall-clock deadline, which a loaded machine exhausts (wr-E3b failed once in the full
  suite, passed alone).
- 1e: gymie-playwright `api.ts` defaults `API` to port 8080 whatever `GATEWAY_PORT` says, so
  `fixtures.ts` throws on a run that exports only `GATEWAY_PORT`.
- 1f: `check-task-commit-fields` searches a `**Tests:**` token written `Class.method` literally, so
  the spelling the pipeline's own `fix-mutation:` lines use is never found.
- 5a: `close-task` accepts an appended task whose only dispatch rows carry no commit; kan-1059's
  rows `task-7-implementer-fix-2` and `task-8-implementer-fix-3` render "Commit: no commit".

## What changes

- `check-panel-fix-single-dispatch` accepts a trailer one round below the row's key;
  `flow record dispatch begin` refuses a re-review while a commit authored since the reviewed round
  began carries a panel-fix trailer of another round.
- `flow record dispatch begin` refuses a re-review while the reviewed round's fixed Critical or
  Important findings have no fix-mutation row.
- The unfinished-work gate's false-positive route keeps only a verdict over a passing spec; a
  non-recurring verify failure is fixed anyway.
- `onPausedClock` bounds its wait by ticked fake-clock time, not wall time.
- `api.ts` derives its default port from `GATEWAY_PORT`.
- A `Class.method` test token is found when both segments are.
- `close-task` refuses an appended task none of whose rows carries a commit; rows 8623 and 8632 are
  repaired to their task commits.
