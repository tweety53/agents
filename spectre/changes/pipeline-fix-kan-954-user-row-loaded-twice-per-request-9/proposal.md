# pipeline-fix-kan-954-user-row-loaded-twice-per-request-9

## Why

In the gymie `/flow` run of `kan-954-user-row-loaded-twice-per-request`, the self-review fix loop
on the change's own branch recorded `first-review-effort` as `low` in its plan,
`spectre/changes/self-review-kan-954-user-row-loaded-twice-per-request/design.md` on
`spectre/kan-954-user-row-loaded-twice-per-request` — where `skills/flow-self-review/SKILL.md` step
3 item 1 puts it. `scripts/check-loop-dispatch.sh` still refused the `review-1` dispatch on
`flow-low`: `PAIR — … runs on flow-medium, as the loop's design.md first-review-effort decision
records (flow-medium when none)`. `lpFirstReviewType` (`stats/internal/guard/loop.go`) reads
`spectre/changes/<change>/design.md`, and for an own-branch loop the change its rows are filed
under is `<name>`, not the loop's plan `self-review-<name>`. On a `/flow-fast` change it also looks
on the wrong branch: the change's own branch there is `<name>`, not `spectre/<name>`.

## What changes

- `lpFirstReviewType` reads the loop's plan at `spectre/changes/<loop branch name>/design.md` —
  `self-review-<name>` for a self-review on the change's own branch, the loop branch otherwise —
  and, for an own-branch loop, looks for it on `spectre/<name>` and then `<name>`.
- `TestCheckLoopDispatchOwnBranchInAnotherProject` places its plan where the skill puts it, and a
  new test covers the `/flow-fast` own branch.
