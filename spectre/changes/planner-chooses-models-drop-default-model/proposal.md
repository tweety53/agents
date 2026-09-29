# planner-chooses-models-drop-default-model

## Why

- `ea661bc6` (2026-09-24) fixed a forced non-default rerun model by pinning **every** Decide pair to
  `DEFAULT_MODEL`. That removed the planner's per-dispatch model choice, which is intended
  behaviour: the operator's global model rule (opus by default, sonnet for simple predictable
  work, opus on every first-pass review) cannot be expressed while one model governs every role.
- With the planner choosing, `default_model` (store column, `## model` project key) has no job left.
- `self_review_model` is deprecated: the operator runs `/flow-self-review` on a model picked by hand
  (`/model`), and the deferred bundle is the only self-review path they use. The inline pass, the
  run/skip prompt and the `## self review` key are dead weight.

## What changes

- **Decide picks every model.** Implementer and fixer (on `sdd`), each panel dispatch, the rerun pair and each implementer group carry a planner-chosen
  `{model, effort, reason}`, `model` ∈ {`opus`, `sonnet`}, under the bounds in `design.md`.
- **No `DEFAULT_MODEL`.** Micro-class dispatches and any dispatch with no recorded pair run on the
  literal `opus`. `VERIFY_MODEL` stays `opus`. A session instruction still overrides, per run.
- **`flow record decision` refuses** a decision whose model fields name anything but opus/sonnet.
- **Self-review is deferred, always** — `/flow` archive step 9 and `/flow-fast` write and land the
  context bundle; `/flow-self-review` is the only reasoning pass.
- **Removed:** `flow_settings.default_model` and `.self_review_model` (migration), their API /
  client / CLI fields, `flow settings models`, the `## model`, `## self review model` and
  `## self review` project keys, `check-model-keys.sh` (+ Go guard) and
  `check-model-resolution-shell.sh` (+ harness). `/flow-settings` manages reviewer slots only.
- `ValidModels` shrinks to {`opus`, `sonnet`} and is the set the decision refusal enforces.
