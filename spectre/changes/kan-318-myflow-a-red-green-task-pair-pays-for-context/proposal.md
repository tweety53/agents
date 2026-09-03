# kan-318-myflow-a-red-green-task-pair-pays-for-context

## Why

A `Build: red` task and its `Squash-with:` green partner are one unit of work dispatched as two
implementers. The red dispatch reads the plan, the specs, the principles and the surrounding code
to write failing tests; the green dispatch re-reads all of it minutes later to implement against
them. KAN-212's 5 red-partner dispatches cost 30.5M tokens, 15% of the run, almost all of it
context the green partner re-acquired; the after-the-fact fold (`--fixup` + `--autosquash`) cost
that run a correction round when it had to step over an intervening fix commit. The contracts
(`skills/flow-contracts/build-green.md`, `openspec/specs/myflow-build-green`) already say the pair
is "dispatched together as a single unit" — `plan-dispatch-bundles.py` never delivered it, because
it joins on `**Files:**` overlap only and a test file and its source file rarely share a path.
KAN-318.

## What changes

- `scripts/plan-dispatch-bundles.py`: a `**Squash-with:**` field is a bundling edge — a red task
  lands in its partner's bundle. `scripts/test-plan-dispatch-bundles.sh` pins it.
- `skills/flow/implement.md` §4: one implementer runs the red task (tests written, run, failure
  reported) then its partner, and makes one commit — the partner's `**Commit:**` subject and
  `Task-Id`. The fold choreography and the `red-partner` dispatch paragraph go; `-role` no longer
  names `red-partner`. Per-task review is one reviewer per commit in a bundle; a red task is
  ticked with its partner.
- `skills/flow-contracts/build-green.md`: names the bundler as the mechanism behind "dispatched
  together as a single unit".

Out of scope: `check-task-commit-fields.py` and `check-task-build-green.py` (a pair producing one
commit is the fold they already resolve); `recordRoles` in `stats/cmd/flow/record.go` keeps
`red-partner` for old rows; `openspec/` is frozen. KAN-319 (a red task's per-task review
duplicates its partner's) is not implemented — it ceases to exist once this lands, and can be
closed against this change.
