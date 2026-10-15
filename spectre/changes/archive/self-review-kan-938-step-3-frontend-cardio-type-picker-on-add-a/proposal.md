# self-review-kan-938-step-3-frontend-cardio-type-picker-on-add-a

## Why

The self-review pass over kan-938 found six defects, none `big` — four `flow-fix`, two `flow-stats-app`:

- F1: `close-task`'s implementer `flow record dispatch end` carries no `-C <canonical>`, so a run from
  another checkout (the agents repository) files the row under that checkout's project key.
- F2: the unfinished-work guard's signal four counts only `panel-…` reviewer rows, but its breakdown
  message and `scripts/check-unfinished-work.sh`'s header say any reviewer row.
- F3: a cross-repo visual-verify fix-round re-run does not say whose round base is substituted into
  `specs`' `<merge-base>`.
- F6: `close-task` closes an appended task no dispatch row names; nothing says so until integrate's
  unfinished-work gate.
- F7: every `flow` write resolves the project key from its working directory with no sanity check, so a
  call from the wrong repository files run records where the change's own reads never look.
- F8: `flow record dispatch end -outcome` takes free text; the ledger carries `done` and `approve` beside
  the contracts' `completed`.

## What changes

- `close-task` ends the implementer's dispatch row with `-C <canonical>`.
- Signal four's message names the `panel-<round>-<slot>` key it counts; the script header says so.
- The verifier's step 7 and the fix loop state that `specs` takes the `<frontend-root>` worktree's own
  round base, unscoped when it has none.
- `close-task` warns, never blocks, when it closes an appended task no dispatch row names — signal
  four's predicate, shared with the guard.
- `flow record dispatch end` refuses an `-outcome` outside the contracts' closed set.
- `flow record dispatch begin|end`, `finding`, `status` and `pass` refuse a change no checkout of the
  `-C` repository carries.
