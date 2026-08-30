# kan-370-self-review-model

## Why

`/flow`'s archive-phase self-review (the 5-angle retrospective run after `FINISHED` is written —
`skills/flow/archive.md` step 9 / `skills/flow-contracts/finish-contract.md` Run 2 step 9)
currently runs as an inline reasoning pass in the same session as the rest of `/flow`. Claude
Code's per-command frontmatter only sets the session's own model, so self-review has no model of
its own today — it inherits whatever the running `/flow` session happens to be on. There is no way
to pin self-review to a stronger model (e.g. Opus) independent of the session model, or to persist
that choice across runs.

## What changes

- `flow_settings` gains a `selfReviewModel` field (empty means "inherit `defaultModel`"), stored
  and validated the same way `defaultModel` already is.
- `flow settings set` gains a `-self-review-model` flag; `flow settings get` reports the new field.
- `/flow-settings` asks a third question — "Self-review model" — offering the known models plus
  "Inherit default model", seeded from the current value.
- `/flow`'s model resolution reads the new field into `SELF_REVIEW_MODEL`, resolving to
  `DEFAULT_MODEL` when empty.
- `skills/flow/archive.md` step 9 dispatches the 5-angle reasoning pass as a subagent on
  `SELF_REVIEW_MODEL` instead of running it inline in the main session. The subagent returns the
  five angles' findings; the main session still runs the per-angle filing prompts (`AskUserQuestion`
  cannot be driven by a subagent) and commits the report exactly as today.

Out of scope: the review-panel (code-review) model and the implementer/panel-fix model roles —
none of those change.
