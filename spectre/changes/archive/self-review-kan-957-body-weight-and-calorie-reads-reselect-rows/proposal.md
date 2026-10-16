# self-review-kan-957-body-weight-and-calorie-reads-reselect-rows

## Why

The self-review pass over kan-957 found three non-`big` defects, each reproduced against the tree
(`design.md`, **Context**):

- gymie's `scripts/count-request-queries.sh` drops every statement whose SQL begins on a new line —
  a Kotlin `"""` native query. Postgres logs `LOG:  execute <unnamed>: ` with the SQL on
  tab-indented continuation lines; `count_slice` trims the first line to "" and discards it as the
  pool's validation query, and the continuations are not attributed because nothing is held.
  kan-957's own live check undercounted for that reason.
- `flow record summary` is prescribed (**Every summary a run prints is recorded as it prints**,
  `skills/flow-contracts/pipeline.md`, added by `937955e9`), yet 35 of 36 recent gymie changes have
  no summary row, and nothing in `scripts/` or `hooks/` refuses a change that never wrote one — the
  self-review bundle's first source goes missing silently.
- **Live check** (`skills/flow/verify-and-handoff.md`) asks for before/after figures but names no
  source for the before one; kan-957's run checked base sources out over the change worktree and
  restored them by hand.

## What changes

- gymie: `count_slice` counts an empty statement line followed by a tab-indented continuation as a
  statement whose text is that continuation; an empty one with no continuation stays the
  uncounted validation query. `scripts/test-count-request-queries.sh` gains the case.
- `check-unfinished-work.sh` gains signal six: a change whose store holds no summary row is
  `OUTSTANDING`, read through `flow self-review bundle`. **Write `IN_PROGRESS`** records the
  `**Summary:**` bullets before that guard runs, and integrate's `STOP: unfinished-work`
  (`skills/flow/integrate.md`) names how the breakdown's summary line is closed.
- **Live check** states that a before figure comes from a detached worktree at the merge base,
  never from base sources checked out over the change worktree.
