# myflow-review-panel-roster delta — kan-302-panel-code-review-slot-hangs-on-fork

## REMOVED Requirements

### Requirement: The light preset's third slot invokes the harness's code-review skill

**Reason:** The `code-review` skill forks its own background agent, which the panel dispatcher never
observes, so the slot could not report through the panel's contract. Observed three times on
KAN-295: one run returned its findings to the session while the slot stalled and was killed, one ran
four hours with no result, and a re-dispatch on this requirement's own fallback shape returned clean
in 2.5 minutes. Replaced by **The light preset's third slot is a general-purpose reviewer** below,
which makes that fallback shape the only shape.

## ADDED Requirements

### Requirement: The light preset's third slot is a general-purpose reviewer

Under the `light` preset, the third required slot SHALL be **Code review (low)**: a
`general-purpose` subagent, dispatched on the model recorded under `models.reviewPanel` and
defaulting to Sonnet, briefed to report high-confidence defects only against the panel's diff in the
worktree. It SHALL NOT invoke a skill.

The slot's name SHALL remain `Code review (low)`. It is the value `myflow record dispatch begin
-slot` writes into the store, and keeping it leaves dispatch rows written before this change
comparable with rows written after it.

Because the dispatcher names the model, the SDD ledger SHALL record that model for this slot and
SHALL NOT record `unknown (agent-defined)` — that value is reserved for slots dispatched by
`subagent_type`, whose agent definitions the dispatcher does not read.

Its findings SHALL be recorded exactly as any other slot's: an `F<n>` row in the findings table and
a marker line in the marker block. This requirement SHALL NOT change the panel record's format.

Because the slot invokes no skill, there SHALL be no harness-availability condition on it and no
substitution for the panel record to name. The panel SHALL NOT fall back to two required slots on
any account.

#### Scenario: The slot is dispatched with a named model

- **WHEN** the panel dispatches Code review (low)
- **THEN** it is a `general-purpose` subagent given `models.reviewPanel` explicitly
- **AND** it is given no skill to invoke
- **AND** the ledger line for it records that model rather than `unknown (agent-defined)`

#### Scenario: Its findings are ordinary findings

- **WHEN** Code review (low) raises a defect
- **THEN** the panel record carries an `F<n>` row and a marker line for it, in the same format every
  other slot's findings use

#### Scenario: The slot's shape does not depend on the harness

- **WHEN** the panel dispatches Code review (low) under any harness
- **THEN** it is the same `general-purpose` reviewer in each
- **AND** the panel record names no substitution, because none was made
- **AND** the panel still has three required slots
