# self-review-kan-989-stats-ui-dashboard-review

## Why

The self-review pass over kan-989 found nine non-`big` defects, each held against the tree
(`design.md`, **Context**):

- F1: the human-gate fix run skipped visual verification and recorded no inline `verify` row,
  ran `check-unfinished-work.sh` without its `<worktree>=<merge-base>` ranges — exit 2 — and
  wrote `IN_PROGRESS` anyway. Given the ranges the guard prints `CLEAR:`: it sees neither gap.
- F2: fix-run plan appends are committed without `check-plan-shape.sh`; three of kan-989's
  planning commits carry a plan that guard refuses.
- F3: a reviewer row keyed `panel-fix1-primary` is no panel round to
  `records.PanelDispatchRound`, so the round ran with no pass row and nothing noticed.
- F4: the guard and `cmd/flow` package tests post `guard_runs` telemetry into the dev store
  (6,065 rows).
- F5: `stats/web/playwright.config.ts`'s `maxDiffPixelRatio: 0.02`, widened for an antialiasing
  flake, let a stale empty Reviewers baseline pass.
- F6: implementer keys `task-11-12` and `task-12` are off convention; `task-11-12` reads as task
  11 alone, and dispatch 18 recorded the wrong commit.
- F7: the failed `verify` row stayed open while `verify-fix-1` opened.
- F8: the visual verifier reported exit 126 for a command the harness denied — an exit code the
  command never returned.
- S1: the inline verify list runs serially, about ten minutes, past the 600 s tool timeout.

## What changes

- `check-unfinished-work.sh` derives its own worktree's range when none is passed, folds in
  `check-visual-verify-dispatched.sh`'s verdict, and finds a code commit no later completed
  `verify` row covers unfinished.
- `document-fix.md` runs `check-plan-shape.sh` and `check-plan-provenance.sh` before the fix-run
  planning commit.
- `flow record dispatch begin` refuses a reviewer `panel-` key with no numeric round, and an
  implementer key outside `task-<ids>-implementer[-fix-<k>]` and the loop keys.
- Both test packages run with `FLOW_GUARD_TELEMETRY=off`.
- The Playwright suite renders deterministically and compares at zero tolerance against
  re-captured baselines.
- `verify-fix-loop.md` closes the failing row first; `visual-verify-verifier.md` reports a denied
  command as blocked with the denial text; `verify-and-handoff.md` runs the lint and test lists
  as two concurrent, polled groups.
