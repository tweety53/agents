# orphan-workspaces

## Why

Run 2 cleanup removes only the finishing change's own workspace. Workspaces no live worktree
derives — left by `-before` worktrees or by earlier runs whose cleanup never ran — accumulate in
the project's shared services; gymie's 15-slot cache index pool had six such orphans, removed by
hand. The operator wants them cleaned automatically.

## What changes

- An optional fourth command-table verb, `orphans`: a project command that prints one workspace id
  per line for every workspace on the machine that no live worktree derives.
- `integrate-change.sh cleanup`, after the change's own `remove`, runs `orphans` from the main
  checkout and runs `remove` for each id printed, reporting `WORKSPACE-ORPHAN-REMOVED: <id>` or
  `WORKSPACE-ORPHAN-REMOVE-FAILED: <id> — exit <rc>`. No failure stops cleanup.
- `check-workspace-isolation` accepts the verb; the command table in
  `skills/flow-contracts/finish-contract-run2.md` states it.
