# self-review-kan-1024-templates-warm-up-cool-down-any-workout

## Why

The self-review pass of `kan-1024-templates-warm-up-cool-down-any-workout` found ten fixable
defects:

- F-b: `kickoff-worktree.sh --from` writes a new `link.md` naming only `.` and the repository it
  adds, so worktrees added at kickoff before the change directory existed are left out of the
  merge order.
- F-c: two playwright flakes — `fidelity/templates.spec.ts`'s `sq` catch-up loop and
  `tests/offline-floor-actions.spec.ts`'s `floor_notice` wait — fail under load.
- F-d: the AD3 capture snaps `/template?state=PUBLIC` unfiltered, so any public template in the
  worktree's database lands in `ad3-unpublish-sheet`.
- F-e: the admin-frontend launch environment added to `startService` in gymie's
  `gradle/dev-lifecycle.gradle.kts` has no test.
- F-i: `flow record dispatch begin` admitted `panel-fix-1` before its review round had any
  Critical or Important row, and admitted the same key again for a later round.
- F-j: an in-run visual-verify fix task whose `**Files:**` omits the change's
  `visual-verification.md` passes `check-plan-shape.sh`.
- F-k: the panel record's verdict line counts rows another row of the same round supersedes.
- F-l: the abandoned-stage sweep ends a stage run at the sweep's instant, overlapping later stages
  of the same change.
- F-m: `EndDispatch` accepts an end before the row's start, and the renderer prints a negative
  duration.
- F-n: reproducer script paths are reused across rounds — the rendered slot prompt never states
  the round number — and the guard checks neither cross-round reuse nor closed rows.

## What changes

- Kickoff's merge-order writer names every repository the change's state record holds.
- The two flakes wait on the state they need instead of on a tick loop or a long timeout.
- The AD3 capture removes the public rows the frame does not draw before it snaps.
- The admin-frontend launch environment is built in buildSrc and tested there.
- `flow record dispatch begin` refuses a `panel-fix-<r>` key whose review round `<r>-1` holds no
  Critical or Important row, and a key another session of the change already used.
- `check-plan-shape.sh` refuses an `In-run fix <k> — visual-verify` task that omits
  `visual-verification.md` from `**Files:**`.
- The panel verdict skips superseded rows of the same round.
- The sweep ends an abandoned stage at the next stage's start, and a `flowd` command clips the
  rows already written.
- `EndDispatch` clamps a same-second end to the start and refuses an earlier one; the renderer
  flags a stored end before the start; kan-1024's sub-second row is retimed.
- The rendered slot prompt states the round number, `check-panel-reproducers.sh` refuses a
  reproducer path two rounds record, and kan-1024's F8 and F11 rows are repaired.
