# kan-389-flow-dispatch-a-dedicated-subagent-for-test

## Why

`flow.verify` and `flow.visual-verify` (`skills/flow/verify-and-handoff.md`) run the project's
`## lint`/`## test` commands and the Playwright `setup`/`verify`/`capture` steps inline in the
conductor, which `skills/flow/implement.md` dispatches on `DEFAULT_MODEL` — whatever the settings
store or a session override resolved for implementation and review that run. Those runs are
mechanical: run a command, read output, report pass/fail. They neither need nor should inherit the
operator's implementation model; their cost and behaviour should be predictable regardless of it.
KAN-389.

## What changes

- `skills/flow/verify-and-handoff.md`: `flow.verify` and `flow.visual-verify` each dispatch a
  `verifier` subagent — one per worktree per stage, `subagent_type: general-purpose`, Agent-tool
  `model: sonnet` literal — to run the commands. The conductor keeps every stage mark,
  `prepare-workspace.sh`, the visual stage's resolve/trigger skips and its commit step, and every
  block decision; it reads the verifier's `## Report`.
- `skills/flow/SKILL.md` **Model resolution**: `VERIFY_MODEL=sonnet`, documented as a fixed
  literal beside `DEFAULT_MODEL`, `SELF_REVIEW_MODEL` and `PLANNING_MODEL` — not read from the
  settings store or `<project>/.flow/project.md`, not overridable by a session instruction.
- `stats/cmd/flow/record.go`: `verifier` joins `recordRoles`, so the dispatch is recorded and
  costed like every other; `record_test.go` pins it; `skills/flow/implement.md`'s `-role` sentence
  names it.

Out of scope: what `## test`/`## visual verification` commands run for any project; the
resolution of `DEFAULT_MODEL`, `SELF_REVIEW_MODEL` or `PLANNING_MODEL`; the review panel's
reproducer runs (conductor-inline, not a dispatch) and the archive self-review subagent (runs no
tests).
