# kan-121-run-the-guards-own-parsers-over-tasks-md-at-plan

Run the guards' own parsers over tasks.md at plan time.

**Jira:** KAN-121 — absorbs KAN-120, KAN-122, KAN-137, KAN-140; related KAN-114.

## Why

A `tasks.md` that a downstream guard cannot parse is written, dispatched, and rejected only
**after** work has been done against it, forcing a mid-run reshape.

`check-task-build-green.sh` already runs at plan time — it is a `## lint` step and scans a whole
`tasks.md` in one pass. `check-task-commit-fields.sh` cannot: its signature is
`<worktree> <task-id> <commit-sha> [parent-sha]`, so there is no way to run its field parser over a
plan before a commit exists. Five plan-shape defects are therefore invisible until that commit
lands. `design.md` is canonical for what each one is and which are still live.

## What changes

- A new shipped guard, `scripts/check-plan-shape.sh` + `.py`, that **imports**
  `check-task-commit-fields.py` and `lib/plan_grammar.py` rather than reimplementing their grammar,
  and reports six plan-shape findings before any task is dispatched.
- One behaviour change in `check-task-commit-fields.py`: a `Tests:` field opening with the literal
  `none` declares zero tests, so a no-test task can say so without its backticks becoming test
  names that must appear in the commit's diff.
- The guard is wired into this repository's `## lint`, into `/flow`'s writing-plans step and its
  pre-dispatch step, and into `skills/flow/scripts/` as a shipped guard.
