# kan-829-flow-improvement-fix-pipeline-defects-at-the

## Why

A run that hits a defect in the pipeline itself — a skill, contract, guard, the `flow` CLI, a rule
or a saved memory — has the freshest context to fix it, yet nothing tells it to: the agent baseline
treats `<agents repo>` as "a repository you were not given" from any other project, so the defect
is reported and waits for self-review to file a `flow-fix` ticket and a later change to fix it.
kan-745's run fixed three such defects at their source mid-run (`flow-decide-default-model`,
`remove-code-review-low`, `panel-handback-auto`) and self-review flagged that as the practice to
reproduce (KAN-829).

## What changes

- `/flow` and `/flow-fast` fix a pipeline defect they hit within the run, in its own
  `<agents repo>` worktree branch, and land it by that repo's default landing route without
  asking, while the run continues.
- Whether to fix in-run is predicted before any work from the fix's blast radius: 60+ files, a
  reach outside `<agents repo>`, or an operator-only design choice makes it `big`, and a `big` fix
  is deferred to self-review as today.
- The fix loop is one-shot at every step: a fix subagent, then fresh reviewers, inline or fresh
  fixers and fresh re-reviewers until clean — no resumed agent, no parked question.
- Each in-run fix is one narrative line; self-review reports a fixed one as fixed and never offers
  it for filing.
