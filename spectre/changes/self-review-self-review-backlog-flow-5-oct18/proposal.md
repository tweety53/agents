# self-review-self-review-backlog-flow-5-oct18

## Why

`self-review-backlog-flow-5-oct18`'s `tasks.md` `## Global constraints` told its implementer that
task commits carry no `Task-Id:` trailer — the self-review loop fixer's convention
(`skills/flow-self-review/SKILL.md`, step 3 item 1) — while the change ran under `/flow`, where
`scripts/close-task.sh` refuses a task commit without one. The implementer met the contradiction
(the change's narrative). Nothing refuses such a plan: `check-plan-shape.sh` passes it.

## What changes

- `check-plan-shape.sh` refuses a live `/flow` plan stating, outside a fence, that task commits
  carry no `Task-Id:` trailer, through a new `flow-guard plan-task-id-rule` subcommand.
- A `/flow` plan is told from a loop plan by its change's `spectre/<name>` branch, local or on
  origin — the discriminator `hooks/enforce-loop-dispatch.py` uses — so a `self-review-<name>` or
  `pipeline-fix-<name>-<k>` loop plan, which carries the rule legitimately, still passes, and a
  `/flow` change named `self-review-…` is still judged a `/flow` plan.
