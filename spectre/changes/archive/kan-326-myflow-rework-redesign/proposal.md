# kan-326-myflow-rework-redesign

## Why

The myflow pipeline (`/myflow-start` → `/myflow-do` → `/myflow-finish`, plus `/myflow-fast`,
`/myflow-status`, `/myflow-research`) has accreted stages and config surface — model-role
questions, review-panel roster presets, a published-proposal step — that no longer earn their
complexity for how the operator actually works day to day. Separately, the stats app's UI has
grown to 8 dashboards, most unused. KAN-326 rescopes the pipeline to what's actually used, renames
the command surface, and gives review-panel/model configuration a real home instead of being
re-asked and re-recorded on every change.

## What changes

- **Command surface.** `/myflow-start`, `/myflow-do`, `/myflow-finish` deleted as commands;
  `/myflow-fast` renamed `/flow`. `/myflow-status` → `/flow-status`, `/myflow-research` →
  `/flow-research` (pure rename — functionally already the replacement for `openspec-explore`).
  New: `/flow-settings`, `/flow-fast` (shorter brainstorming + Primary-only reviewer).
- **Pipeline stages.** Remove `start.ask-options` and `start.publish-proposal`. Redefine `STARTED`
  as a kickoff marker written at invocation rather than after design approval. Fix the review panel
  at 3 required slots by default; Bugbot/Security become on-demand only, dispatched whenever the
  operator asks, at any point in the run. Move `do.run-instructions` to immediately before
  `do.write-in-progress`. Fold `finish.move-in-review` into `finish.landing-routes`. Merge
  `do.workspace-export` and `do.lint-and-test` into one verify stage.
- **Settings store.** New global-defaults store in the stats app for implementer/fixer/reviewer
  models (collapsed to one `models.default` field) and the reviewers list, validated against a
  strict enumerated list. `/flow-settings` writes it; any `/flow` run can override via
  plain-language session instruction, recorded per-run. The per-change state file drops
  `models.implementation`/`reviewPanel`/`panelFix` and `reviewPanelRoster`.
- **`/flow-research` staging.** Formalize writing research notes to a location outside
  `spectre/changes/`; `/flow`'s brainstorming stage detects and seeds from staged notes for a
  topic without skipping its own interactive round.
- **Stats UI/backend cut.** Keep `StateBoard`, `Trend`, `CacheEfficiency`, `RunDetail`. Remove
  `CostPerChange`, `ModelComparison`, `PanelEconomics`, `ReworkRate` and their backend support.
- **MD cleanup.** Rework-driven cleanup inline with the above; a separate general verbosity pass
  afterward, out of this change's scope.

Full design record: `docs/superpowers/research/kan-326-myflow-rework.md` (research staging notes)
and `design.md` in this change.
