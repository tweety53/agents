# kan-380-flow-self-review-model-overridable-per-project

Jira: KAN-380.

## Why

- `SELF_REVIEW_MODEL` (**Model resolution**, `skills/flow/SKILL.md`) reads one field, the settings
  store's `selfReviewModel`, and inherits `DEFAULT_MODEL` when it is empty. A project cannot set
  the model its archive-phase self-review runs on without changing the harness-wide store.
- `PLANNING_MODEL` already resolves session instruction > `## planning model` in
  `<project>/.flow/project.md` > store `planningModel` > literal `fable`, with a verified `opus`
  fallback on the planner dispatch (KAN-374). Self-review is the one model-bearing dispatch left
  with a single-tier resolver and no fallback.

## What changes

- **`SELF_REVIEW_MODEL` resolves in `PLANNING_MODEL`'s shape**: session instruction >
  `## self review model` in `<project>/.flow/project.md` > store `selfReviewModel` > literal
  `fable`. An unreachable store resolves to `fable` too, named a fallback.
- **The self-review subagent's model is verified.** Its report opens with `Model: <name>`; a
  mismatch re-dispatches once on `opus`; a second mismatch proceeds on whatever answered and names
  it. No dispatch record is added.
- **An empty store `selfReviewModel` means `fable`**, not "inherit `defaultModel`" — the meaning
  `planningModel`'s empty already carries. `/flow-settings`' "Inherit default model" option
  becomes "Store default (fable)". Every site that says inherit is reworded; no Go logic, no
  migration.
- **`skills/flow-contracts/project-configuration.md` gains the `## self review model` key**, and
  this repository's `.flow/project.md` sets it to `fable`.
