# self-review-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets

## Why

The self-review pass of `kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets` found five fixable
defects:

- F1-a: the transcript-based restamp of a stamped dispatch end is a no-op since agents `6560ecb7`
  made `Store.EndDispatch` keep a closed row's end; agent rows closed since 2026-10-28T15:30Z end at
  the close call, not at their agent's last write (panel-6 and panel-10 slots within 50 ms of each
  other) <!-- measured: dispatches ended_at of the panel-6 and panel-10 reviewer rows, self-review pass sr-findings.md F1-a @ 2026-10-28 -->.
- F1-b: inline implementer rows misrecord the parent's work — a phantom aborted row carrying task 7's
  tokens (8410), a duplicate completed row on task 8's commit (8437), and two rows begun seconds
  before their commits (8449, 8458).
- F1-c: three frontend panel-fix commits (`3b9a3985`, `ae761ec7`, `25394c9a`) have no dispatch row,
  and the unfinished-work gate did not flag them.
- F1-e: `measure-visual-properties.sh` reads every frame under a scrim as one full-height band, so a
  modal-sheet frame gets no band or seam comparison.
- A8-a: an in-run fix for a visual departure seen on a running clock closes on the pinned-clock
  capture spec, where the departure never shows; three extra visual-verify rounds followed.

## What changes

- The daemon's restamp of a stamped end moves the row's end again; a new `flowd restamp-ends` repairs
  the rows closed since `6560ecb7`.
- A live inline implementer row closed with its commit starts no later than the back-fill's derived
  start; a back-fill duplicating a task's completed commit, and a live begin of a task already
  completed, are refused; a new `flowd drop-dispatch` removes a phantom or duplicate row, moving its
  tokens to the kept one. Rows 8410, 8437, 8449 and 8458 are repaired.
- `check-unfinished-work` flags a `Flow-Key: panel-fix-<r>` commit with no dispatch row of that key,
  reading the pre-reshape pin when one exists; the three missing rows are back-filled.
- Band and seam measurement crops a scrim frame to the sheet it dims.
- A task fixing a departure observed on a running clock reproduces it red under that clock and closes
  only on bounded repeated passes.
