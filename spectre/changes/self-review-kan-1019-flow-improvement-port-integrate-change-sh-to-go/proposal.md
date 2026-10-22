# self-review-kan-1019-flow-improvement-port-integrate-change-sh-to-go

## Why

kan-1019's self-review pass found six defects, none `big`:

- **A (flow-fix).** kan-1019's `## Live check` — two `prepare` runs, bash at `a03f3c4b` against the
  shim, compared — was deferred from verify to integrate ("running it twice here would mark
  integrate stages before integrate") and at integrate only the shim half ran; no
  `live-verification.md` exists. Nothing ties a design's live check to its record:
  `check-live-verification.sh` runs only when the verify procedure is followed by hand.
- **B (flow-fix).** `generate-relocation-comparison` exits 2 when a scoped file is absent from the
  worktree (`stats/internal/guard/relocationcomparison.go`, `grcReadWorktree`), so a relocation
  plan that deletes a file it moves text out of gets no comparison at all — kan-1019's panel lost
  it over the deleted `scripts/test-integrate-change.sh`.
- **D2 (flow-fix).** The three Important parity fixes of kan-1019's panel — F1 (cleanup through the
  shim from inside a worktree), F2 (an indented report row), F3 (a padded orphan id) — have no suite
  test; their only proof is untracked reproducers under `.superpowers/sdd/`.
- **C (flow-automation).** `check-plan-shape` does not check a Baseline chain: kan-1019's original
  plan declared tasks 3 and 4 both `before=1 after=2` under one `measured:` command and passed; the
  error surfaced only at task 4's commit.
- **E (flow-stats-app).** `skills/flow/review-panel.md` records every slot's `end` in one call once
  all reports exist, so every parallel reviewer row carries the round's last finish: kan-1019's
  sonnet slot and opus slot of round 0 both read the same end.
- **D1 (flow-speed).** `TestIntegrateChangePrepareCommit/narrative_format` runs six fixtures in
  series in one case, the critical path of the integrate-change tests; `rerun_pin` runs two.

## What changes

- `integrate-change prepare` stops at `STOP: live-check` while the change's design names a live check
  and `live-verification.md` is missing or fails `check-live-verification.sh`; the planner writes a
  live check runnable at verify, through a stated stand-in where a step needs the landed change or
  would write integrate state. kan-1019's own live check is run on that stand-in and its record
  committed.
- A scoped file the branch deleted reads as empty text in the relocation comparison; one missing on
  disk but still tracked stays exit 2.
- Three suite cases pin F1-F3.
- `check-plan-shape` F17 refuses a same-command Baseline chain whose `before` falls below the previous
  task's `after`.
- Each panel slot's `end` carries its own report file's modification time; kan-1019's two round-0
  rows are re-timed.
- The two multi-fixture cases run as parallel table entries.
