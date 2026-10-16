# pipeline-fix-kan-954-user-row-loaded-twice-per-request-8

## Why

A self-review pass inside a gymie `/flow` run (change `kan-954-user-row-loaded-twice-per-request`)
could not run its fix loop. The pass's findings were fixes to gymie's own dev tooling, which ride
the change's own branch (`skills/flow-self-review/SKILL.md` step 3, **Pipeline defects found
mid-run** in `skills/flow-contracts/pipeline.md`), so no `self-review-<name>` branch exists in the
agents repository. Every loop tool still resolved flowd's project from the agents checkout:
`hooks/enforce-loop-dispatch.py` hands `scripts/check-loop-dispatch.sh` the agents checkout, whose
guard (`stats/internal/guard/loop.go`) runs `flow record dispatches` and `flow record dispatch
begin` with `-C <agents main checkout>`. `flow` refused — no checkout of the agents repository
carries change `kan-954-user-row-loaded-twice-per-request` — and the hook denied
`self-review-kan-954-user-row-loaded-twice-per-request-plan`. `scripts/pipeline-fix.sh end` and
`escalate` fail the same way for such a key, and `pipeline-fix.sh verify` diffs a gymie worktree
against the agents repository's `origin/HEAD`. The parent then stopped and asked the operator
instead of fixing the pipeline in-run.

## What changes

- A self-review loop on the change's own branch — no `self-review-<name>` branch in the agents
  repository — reads and writes its flowd rows, its escalation markers and its
  `first-review-effort` decision in the project of the caller's working directory: the hook runs
  the checker in the PreToolUse payload's `cwd`, and `pipeline-fix.sh end`/`escalate` use the
  directory they are run from. Pipeline-fix loops and `self-review-<name>` loop branches keep
  filing under the agents repository, unchanged.
- `pipeline-fix.sh verify <worktree>` diffs the worktree against its own `origin/HEAD`.
- `check-loop-landing.sh` is unaffected: it already resolves flowd's project from the repository
  being pushed.
- `skills/flow-self-review/SKILL.md` states that a loop tool's refusal or failure is a pipeline
  defect fixed in-run, never a stop or a question to the operator.
