## MODIFIED Requirements

### Requirement: The pipeline has exactly three states

myflow SHALL track a change in exactly one of three states: `STARTED`, `IN_PROGRESS`, `FINISHED`.
Each pipeline command SHALL end in the state named after it, so the state names and the command
names are the same vocabulary.

The human gate SHALL be a property of the state rather than a separate stage: `STARTED` means the
proposal awaits reading, `IN_PROGRESS` means the staged diff awaits the human's review and the run,
`FINISHED` is terminal.

The state SHALL be held in a **state record** whose authoritative home is the store defined by
`myflow-state-store`, and whose on-disk JSON form is the fallback journal that store contract
defines. "The state file", wherever this specification and its siblings use the term, SHALL denote
that record rather than a particular storage medium: which medium holds it is governed by
`myflow-state-store` alone, and no requirement about the pipeline's states depends on the answer.

#### Scenario: Each command lands in its own state

- **WHEN** `/myflow-do` completes from `STARTED`
- **THEN** the state is `IN_PROGRESS`, and no separate command is required to record that a human
  reviewed or tested anything

#### Scenario: No state records a human confirmation

- **WHEN** the state record is read at any point in the pipeline, from the store or from its on-disk
  fallback
- **THEN** it contains no field asserting that a human reviewed the proposal, reviewed the diff,
  or ran the tests, and no command exists whose only effect is to write such a field

#### Scenario: The medium does not change the state machine

- **WHEN** a command reads the state record from the store, or from the on-disk fallback because the
  store was unreachable
- **THEN** the three states, the command-to-state mapping, and the human gates are identical in both
  cases
