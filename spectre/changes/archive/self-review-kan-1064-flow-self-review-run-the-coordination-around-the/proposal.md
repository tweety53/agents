# self-review-kan-1064-flow-self-review-run-the-coordination-around-the

## Why

The self-review pass of `/flow` change `kan-1064-flow-self-review-run-the-coordination-around-the`
found four non-`big` defects:

1. `skills/flow-self-review/SKILL.md` says "The `fix` row carries no no-progress escalation", but
   its pairs paragraph makes "the `fix` row" cover `pipeline-fix-fix` too, whose `opus/flow-high`
   `No-progress escalation` row `skills/flow-contracts/model-policy.md` keeps, and
   `skills/flow-contracts/pipeline.md` makes that step canonical for the pipeline-fix loop's
   escalation — which is now forbidden instead of stated. `skills/flow/implement.md` and
   `skills/flow/SKILL-rationale.md` still name the self-review loop's escalation.
2. The planner tags a block `verified:` after applying it to a scratch copy but never runs the lint
   list there; kan-1064's task 2 then failed `check-installed-citations.sh` and
   `check-guard-symlinks.sh` at implementation.
3. The `flow` client never retries, so a `make -C stats restart` by another session fails any CLI
   call that lands in the daemon's restart gap (a `flow tasks count` in kan-1064).
4. `scripts/pipeline-fix.sh end <key>-pass` prints `PIPELINE-FIX-END-NA` and closes nothing, while
   `check-loop-dispatch.sh` records the pass and then refuses the plan dispatch over the still-open
   pass row.

## What changes

- The self-review SKILL scopes the no-escalation sentence to `self-review-fix` and states the
  `pipeline-fix-fix` row's `flow-high` no-progress escalation; the two cross-references follow.
- `skills/flow/brainstorm-planner.md` requires `run-verify-list.sh <scratch copy> lint` before a
  scratch-applied block is tagged `verified:`.
- `make restart` marks its restart and shortens the daemon gap; the client retries a refused dial
  while that mark is fresh, inside the caller's timeout, and still fails fast on a dead daemon.
- `pipeline-fix.sh end <key>-pass` closes the pass row where `check-loop-dispatch.sh` opened it.
