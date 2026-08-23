# myflow-review-panel-roster Specification

## Purpose
The review panel's roster is a recorded, per-change choice rather than a fixed list. This
capability governs the preset a change records, what each preset's roster is, how conditional
slots are offered under the lighter presets, the shape of the per-task review, and the rule that
no preset may move the handoff bar.

## Requirements

### Requirement: A change records a review panel roster preset

A change SHALL record a **review panel roster preset** in its state file, under the top-level key
`reviewPanelRoster`, whose value SHALL be exactly one of `light`, `standard` or `full`.

The preset SHALL be asked once, by `/myflow-start`, on the run that **creates** the change — the
same run that asks the planning effort and the three model roles, and determined the same way: the
state file does not exist. It SHALL be offered as named options with `light` marked as the
recommendation, never as open prose. A revision round SHALL NOT ask; it SHALL state the recorded
value and proceed at it.

Every other command SHALL carry the field forward verbatim, exactly as it carries `jiraIssue`,
`planningEffort` and `models`.

An **absent** `reviewPanelRoster` SHALL read as `light` rather than making the file unparseable, and
SHALL NOT trigger a migration pass over existing state files. A value outside the three SHALL read
as `light` and SHALL NOT make the file unparseable.

`/myflow-status` SHALL surface the recorded preset, and SHALL name the default as the default when
nothing is recorded.

#### Scenario: The creating run asks once

- **WHEN** `/myflow-start` runs for a change whose state file does not exist
- **THEN** it asks which roster preset to use, with `light` marked as the recommendation
- **AND** it records the answer under `reviewPanelRoster`

#### Scenario: A revision round reuses the recorded preset

- **WHEN** `/myflow-start` runs again for a change already at `STARTED`
- **THEN** it does not ask for a preset
- **AND** it states which preset is being reused

#### Scenario: An absent preset reads as the default

- **WHEN** a command reads a state file carrying no `reviewPanelRoster` key
- **THEN** the preset is read as `light`
- **AND** the file is not reported as unparseable on that account

#### Scenario: The preset survives every later command

- **WHEN** `/myflow-do` or `/myflow-finish` writes the state file
- **THEN** `reviewPanelRoster` is re-emitted as read, rather than reset

### Requirement: Each preset names the panel's required slots

`/myflow-do` SHALL select the review panel's required slots from the recorded preset:

| Preset | Required slots |
|--------|----------------|
| `light` | Primary, Principles, Code review (low) |
| `standard` | Primary, Principles, Bugbot |
| `full` | Primary, Bugbot, Principles |

Every preset SHALL dispatch exactly three required slots, and no preset SHALL reduce that number.
`full` SHALL reproduce the roster in force before this capability existed.

`skills/myflow-do/SKILL.md` SHALL remain canonical for the panel, and SHALL carry this table. The
state-file contract SHALL carry the field's shape and SHALL NOT restate what a preset means.

#### Scenario: The default preset runs the light roster

- **WHEN** a change records `light`, or records no preset at all
- **THEN** the panel dispatches Primary, Principles and Code review (low) as its required slots
- **AND** Bugbot is not dispatched as a required slot

#### Scenario: The full preset reproduces the previous behaviour

- **WHEN** a change records `full`
- **THEN** the panel dispatches Primary, Bugbot and Principles as its required slots, exactly as it
  did before this capability existed

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

### Requirement: The lighter presets offer conditional slots instead of auto-including them

Under `light` and `standard`, `/myflow-do` SHALL evaluate the conditional-slot trigger table against
the panel diff unchanged, and SHALL then present every slot whose trigger fired in **one**
multi-select prompt, naming each slot and the trigger that fired it, with including all of them the
recommended answer.

Under `full`, a fired trigger SHALL auto-include its slot, which is the behaviour in force before
this capability existed. The trigger table's existing borderline *ask* rows SHALL keep their current
behaviour under `full`.

Which optional slots were included, which were excluded, and why, SHALL be recorded as it is today,
under every preset. A slot the operator declined SHALL be recorded as declined rather than as
untriggered.

#### Scenario: A fired trigger is offered under a lighter preset

- **WHEN** a change records `light` and the diff touches a migration, firing the Adversarial trigger
- **THEN** the operator is asked, in one prompt, which triggered slots to include
- **AND** the prompt names Adversarial and the trigger that fired it

#### Scenario: Several fired triggers share one prompt

- **WHEN** three triggers fire under `standard`
- **THEN** all three slots appear in a single multi-select prompt rather than in three prompts

#### Scenario: The full preset auto-includes

- **WHEN** a change records `full` and a trigger fires
- **THEN** that slot is included without a prompt

#### Scenario: A declined slot is recorded as declined

- **WHEN** the operator excludes a triggered slot from the prompt
- **THEN** the panel record states that the trigger fired and the operator declined the slot

### Requirement: The lighter presets collapse the per-task review to one reviewer

Under `light` and `standard`, the per-task review `/myflow-do` runs during subagent-driven
development SHALL be a **single** combined reviewer per task, covering spec compliance and code
quality together, dispatched on the model recorded under `models.reviewPanel`.

Under `full`, the spec-compliance and code-quality reviewers SHALL both run, which is the behaviour
in force before this capability existed.

The SDD ledger SHALL record which shape ran for each task, so the choice is auditable after the
fact.

#### Scenario: One reviewer per task under a lighter preset

- **WHEN** a task completes under `light` or `standard`
- **THEN** one combined reviewer reviews it for spec compliance and code quality together
- **AND** the ledger line for that task records that the combined shape ran

#### Scenario: Both reviewers run under the full preset

- **WHEN** a task completes under `full`
- **THEN** the spec-compliance and code-quality reviewers both run

### Requirement: No preset moves the handoff bar

No roster preset SHALL change the bar at which `/myflow-do` hands off. Handoff SHALL require zero
open findings at any severity under every preset, and a minor finding SHALL block exactly as a
critical one does.

No preset SHALL change the escalation ladder, the fix-round rules, the panel record's format, the
marker-line rules, or the operator handback.

A preset sizes how much reading the panel does; it SHALL NOT be a way to skip review. A preset able
to lower the handoff bar would be the latter.

#### Scenario: A light panel blocks on a single open minor

- **WHEN** a change records `light` and one marker line reads `open` on a minor finding
- **THEN** `/myflow-do` does not hand off

#### Scenario: The record format is preset-independent

- **WHEN** the panel record is written under any preset
- **THEN** it carries the same findings table, marker block and total line the format already
  requires
