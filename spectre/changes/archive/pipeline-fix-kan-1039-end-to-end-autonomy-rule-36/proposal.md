# pipeline-fix-kan-1039-end-to-end-autonomy-rule-36

## Why

`flow record dispatch begin` refuses every planning dispatch a creating `/flow` run records.
`skills/flow/brainstorm.md` requires each Explore/Plan dispatch under `flow.brainstorm` to be
recorded immediately before its Agent call, and those dispatches run before `spectre new`. At that
point the change's worktree, created by `kickoff-worktree.sh` on branch `spectre/<name>`, holds
none of `spectre/changes/<name>`, `spectre/changes/archive/<name>` or `.superpowers/sdd/<name>`, so
`requireChangeCheckout` (`stats/cmd/flow/record.go`) finds no checkout carrying the change and
exits 2. Observed on 2026-10-27 in the kan-1039 creating run, after kickoff and before
`spectre new`, with and without `-C <worktree>`. The check dates from 2071f3e0 (2026-10-15).

## What changes

- `changeCheckouts` (`stats/cmd/flow/record.go`) also counts a worktree whose checked-out branch is
  `refs/heads/spectre/<change>` as carrying the change, so the kickoff worktree is accepted from
  the moment `kickoff-worktree.sh` creates it.
- `requireChangeCheckout`'s comment and refusal message name the branch as a fourth accepted form.
- One regression test in `stats/cmd/flow/record_test.go` fails without the fix.
- Nothing is repaired retroactively: the refused rows were never written, so nothing stored is wrong.
- Out of scope: `-effort xhigh` falling outside `recordEfforts` (the kan-1039 change decides it).
