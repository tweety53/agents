# Design — planner-chooses-models-drop-default-model

## Context

- `ea661bc6` pinned every Decide pair to `DEFAULT_MODEL`; the operator wants the planner to choose
  models, `default_model` and `self_review_model` removed, and self-review deferred-only.

## Model choice

- The Decide step (`skills/flow/brainstorm-planner.md`) assigns `{model, effort, reason}` to every
  pair: implementer, fixer, each `panel.dispatches[]` entry, `panel.rerun_dispatch`, each `groups[]`
  entry. `model` ∈ {`opus`, `sonnet`}.
- Bounds:
  - implementer, fixer, groups — `opus` by default; `sonnet` only for simple, predictable work
    near-certain to succeed first try, the `reason` saying so;
  - a pair carrying a hard seam (concurrency, platform interop, data model, performance-sensitive
    code) — `opus`, always;
  - first-pass panel dispatches — `opus`, always;
  - rerun pair — `opus` or `sonnet`, effort `low` fixed; no must-differ-from-pass-1 constraint (the
    defect `ea661bc6` fixed stays fixed).
- Step 2 records the implementer and **fixer** pairs on `sdd` only; both stay `skipped — inline`
  on an inline decision, whose panel fixes the parent applies on its own model
  (`fixer-pair-sdd-only`).
- **Micro** records no pairs; its panel and every dispatch with no recorded pair
  (review-panel.md's no-decision dispatch) run on the literal `opus`.
- A plain-language session instruction replaces the affected pairs for this run, recorded in
  `overrides`. The model handshake compares against the dispatch's own recorded (or overridden)
  model.
- Unchanged: `VERIFY_MODEL=opus`, the zcode harness mapping, effort rules.
- Removed from `decision.json`: `resolved`. Removed from the Decide preamble and run summary:
  `DEFAULT_MODEL`/`MODEL_SOURCE`; the preamble's `models:` line becomes `reviewers <REVIEWERS>`.

## Record-time refusal

- `ApplyDecisionRecord` (`stats/internal/api/records.go`) parses the decision body's model fields:
  `implementer.model`, `fixer.model`, `panel.dispatches[].model`, `panel.rerun_dispatch.model`,
  `groups[].model`. A string-valued `implementer`/`fixer`/`panel` (`skipped — inline`, `default`)
  and a `null` `groups` carry no model and pass.
- A model outside `store.ValidModels` → `ErrInvalidRecord` (HTTP 400) naming the JSON path and the
  value. Writes only; historical rows are never re-checked.
- Server-side, so every client is covered, not only the CLI.

## Settings store

- Migration `0034` drops `flow_settings.default_model` and `flow_settings.self_review_model`.
  `flow_settings` keeps `reviewers`.
- `store.Settings` = `{Reviewers}`; `DefaultModel`, `ErrInvalidModel` on settings writes,
  `SelfReviewModel` go. `ValidModels` = {`opus`, `sonnet`}.
- API/client DTO drops `defaultModel`, `selfReviewModel`. CLI drops `-model`,
  `-self-review-model` and `flow settings models`.
- Compatibility: an old `flow` binary sending `defaultModel` is refused (400, unknown field — the
  decoder is strict); an old reader sees the field absent. `./setup.sh global` reinstalls.

## Self-review

- Deferred is the only path. `/flow` archive step 9 (`skills/flow/archive.md`,
  `skills/flow-contracts/finish-contract-run2.md`) and `/flow-fast` always fetch the bundle, add
  `## Session narrative`, write `docs/self-review/<name>-context.md` and land it.
- Removed: the run/skip/defer prompt, the inline reasoning pass, archive's filing/rating ask and
  report write, `SELF_REVIEW_MODEL` resolution, the `## self review` and `## self review model`
  keys.
- The handoff `Self-review` line always reads `deferred — docs/self-review/<name>-context.md`.
- `/flow-self-review` unchanged in behaviour; its wording that compares against a "same-run pass"
  is updated.

## Guards removed

- `scripts/check-model-keys.sh` + `stats/internal/guard/modelkeys.go` + tests: no model key left.
- `scripts/check-model-resolution-shell.sh` + `scripts/test-check-model-resolution-shell.sh`: both
  blocks it extracts lose their model resolution; `REVIEWERS` is a one-line jq read.
- Both leave `.flow/project.md`'s `## lint` list; their `KNOWN-BUGS.md` entries are removed.

## Decisions

### Planner chooses every dispatch's model

**ID:** planner-chooses-models
**Status:** active
**Chosen:** the Decide step picks `model` per pair within {opus, sonnet} — matches the operator's
global model rule, which varies model by role and task difficulty.
**Considered:** keep `DEFAULT_MODEL` for every pair (status quo since `ea661bc6`) — cannot express
opus-by-default with sonnet for simple work; `DEFAULT_MODEL` as fallback only — rejected by the
operator in favour of removing the setting entirely; `DEFAULT_MODEL` as a ceiling/floor — adds a
knob nobody asked for.

### Remove default_model everywhere

**ID:** remove-default-model
**Status:** active
**Chosen:** drop the store column, API/CLI fields and the `## model` project key; literal `opus`
where no pair is recorded — nothing reads the setting once the planner chooses.
**Considered:** remove from skills only, keep the column — leaves a dead setting `/flow-settings`
still offers.

### Inline fixer is planner-chosen

**ID:** inline-fixer-planner-chosen
**Status:** superseded by fixer-pair-sdd-only
**Chosen:** step 2 records a fixer pair on every non-micro class — the inline panel-fix dispatch
otherwise has no model source.
**Considered:** literal `opus` for the inline fixer — loses the sonnet-for-simple-fixes option.
**Superseded because:** the gated per-task review found no inline panel-fix dispatch exists — an
inline run's parent applies panel fixes itself on its own model (`implement.md` **Inline — the
parent implements**), so an inline fixer pair governs nothing.

### Fixer pair on sdd only

**ID:** fixer-pair-sdd-only
**Status:** active
**Chosen:** step 2 records the fixer pair only when execution is `sdd`, the one mode with a
panel-fix subagent; inline runs record `skipped — inline` — operator's answer at the task review
(2026-09-29).
**Considered:** inline runs dispatch a panel-fix subagent on the fixer pair — adds a dispatch and
its cost to every inline fix round.

### Refuse off-policy models at record time

**ID:** refuse-off-policy-models
**Status:** active
**Chosen:** `ApplyDecisionRecord` refuses any model outside `ValidModels` = {opus, sonnet} —
enforced where every client writes, not only stated in skill text.
**Considered:** skill text only — an unenforced bound drifts; CLI-side check — misses non-CLI
clients.

### Self-review is deferred only

**ID:** self-review-deferred-only
**Status:** active
**Chosen:** `/flow` and `/flow-fast` always save the bundle; `/flow-self-review` is the only pass,
run on the model the operator picks for that session.
**Considered:** deferred as default with run/skip kept — operator: deferred is "the only option".

### Remove self_review_model

**ID:** remove-self-review-model
**Status:** active
**Chosen:** drop the column, fields and `## self review model` key — deprecated; the deferred pass
runs on the session's hand-picked model.
**Considered:** none raised.

### ValidModels becomes the planner bound

**ID:** valid-models-opus-sonnet
**Status:** active
**Chosen:** shrink `ValidModels` to {opus, sonnet}; with both model settings gone its only reader
is the decision refusal, so the bound lives in one place.
**Considered:** a separate planner set beside a wider `ValidModels` — two sets, one with no
reader.

## Open questions
