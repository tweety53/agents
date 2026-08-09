## MODIFIED Requirements

### Requirement: The first invocation chains brainstorming into implementation

`/myflow-fast`, invoked with no existing state file, SHALL run `/myflow-start`'s brainstorming
stage's clarifying questions and design presentation interactively, one question at a time, exactly
as `/myflow-start` runs them, with no auto-answering. `/myflow-fast` SHALL NOT stop for a separate
explicit design-approval confirm after the design is presented — it SHALL proceed directly into
artifact creation once the design has been presented, unless the operator raises an objection during
that presentation. This departs from `/myflow-start`, which still stops at the confirm; the
divergence is scoped to `/myflow-fast` alone.

Once the OpenSpec artifacts are created, `/myflow-fast` SHALL continue, within the same invocation
and without requiring a further command from the operator, into `/myflow-do`'s SDD+TDD
implementation and review panel, exactly as `/myflow-do` runs them. The run SHALL end at
`IN_PROGRESS` with the same staged-diff-plus-run-instructions handoff `/myflow-do` produces — the
operator's checkpoint for this invocation is that review, not a pre-implementation confirm.

`/myflow-fast` SHALL skip publishing the proposal HTML artifact that `/myflow-start` publishes,
because the operator who answers the brainstorming questions is present in the same session that
produces the design.

#### Scenario: One invocation reaches IN_PROGRESS from nothing

- **WHEN** `/myflow-fast <description>` is invoked for a change with no state file
- **THEN** the operator is asked brainstorming's clarifying questions interactively within that same
  run, the design is presented, and implementation and the review panel run unattended within the
  same run without a separate approval confirm, ending at `IN_PROGRESS` with no additional command
  required in between

#### Scenario: No proposal artifact is published

- **WHEN** `/myflow-fast` completes the first invocation for a change
- **THEN** the state file's `artifactUrl` is `null` and no HTML artifact was published

#### Scenario: The design-approval confirm does not stop the run

- **WHEN** `/myflow-fast` finishes presenting the brainstormed design and the operator has raised no
  objection during that presentation
- **THEN** it proceeds directly into creating the OpenSpec artifacts, with no separate "does this
  look right?" prompt

### Requirement: Recorded defaults favor speed, still overridable

On the run that creates a change, `/myflow-fast` SHALL NOT ask the planning-effort, model, or
review-panel-roster questions `/myflow-start` asks interactively. It SHALL record the recommended
defaults directly: `sonnet` for `models.implementation`, `models.reviewPanel` and `models.panelFix`,
and `light` for `reviewPanelRoster`. An explicit operator instruction given in the same session SHALL
still override a named field, recorded with the dispatch exactly as any other operator override is.

#### Scenario: A creating run records defaults without asking

- **WHEN** `/myflow-fast` creates a change with no state file
- **THEN** it records `planningEffort: default`, `models.implementation: sonnet`,
  `models.reviewPanel: sonnet`, `models.panelFix: sonnet`, and `reviewPanelRoster: light` without
  presenting any of the four questions to the operator

#### Scenario: An explicit instruction still overrides a default

- **WHEN** the operator gives an explicit instruction naming a different value for one of the four
  fields before or during the run
- **THEN** that field is recorded with the operator's value instead of the recommended default, and
  the override is recorded with its dispatch

#### Scenario: Defaults are recommended, not forced

- **WHEN** `/myflow-fast` records the four defaults on a creating run
- **THEN** each recorded value is the recommended default only, never a value the operator is
  prevented from overriding — the absence of a question does not make a default mandatory, per the
  override scenario above
