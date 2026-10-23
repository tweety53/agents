# self-review-kan-1028-mutation-failure-modes-extra-dispatch

## Why

The self-review pass of `kan-1028-mutation-failure-modes-extra-dispatch` found five fixable
pipeline defects:

- A decided panel recorded before kan-1028 drops `failure-modes` and `mutation` from its roster
  (`skills/flow/review-panel.md`), but its `panel.dispatches` bundles still name them and are
  dispatched literally, so both roles run twice and two dispatches of one round create the same
  `<worktree>-mutation-<round>` copy.
- `/flow-status`'s `IN_PROGRESS` template in `skills/flow-contracts/handoff-blocks.md` still prints a
  `## Learning — <topic>` section, where the run's own handoff now prints `**Learning:** <absolute
  path>`.
- `/flow-settings` still offers `failure-modes` and `mutation` from `ValidReviewers` and says every
  id it offers is dispatched, though since kan-1028 every roster drops both.
- The session-token rule asks only for "a short, unique literal string"; kan-1028's resumed run
  reused `mf-k1028b`, the dead session's token, and its rows collided.
- A task added mid-run on a creating run — kan-1028's Task 6 — leaves the only decision row's
  `inputs` stale; `implement.md` has no rule to re-decide.

## What changes

- A role dropped from a decided roster leaves its `panel.dispatches` bundle too.
- The `/flow-status` template prints `**Learning:** <absolute path>`.
- `failure-modes` and `mutation` are retired from `ValidReviewers`, and migration `0044` strips
  them from a stored reviewer list.
- A session token ends in the run's start time, UTC to the second.
- A task added after the decision row is re-decided before it runs; kan-1028 gains a corrected
  decision row.
