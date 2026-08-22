## ADDED Requirements

### Requirement: The `IN_PROGRESS` handoff reports journalled record writes

The `IN_PROGRESS` handoff block SHALL carry one line naming how many of this change's record writes
are still sitting in the journal because the store could not be reached.

The line SHALL be present whether or not any write was journalled. A run whose record writes all
reached the store SHALL say so, so that the line's absence can never be mistaken for a clean run —
the failure this reporting exists to close is precisely one that leaves no trace.

The line SHALL be defined **once**, in the contract that defines the block, so that every command
rendering an `IN_PROGRESS` handoff carries it. That includes the read-only status report, which
renders the same block, and the composite command's own no-artifact variant — the one shape no other
command prints.

This reporting SHALL NOT make a record write blocking. A journalled write still exits 0 and still
leaves the run unaffected; what this requirement adds is visibility at the gate the operator reads,
not a gate.

#### Scenario: Two record writes fell back to the journal

- **WHEN** `/myflow-do` hands off after two of this change's record writes were journalled
- **THEN** the handoff names that two writes are journalled

#### Scenario: Every record write reached the store

- **WHEN** `/myflow-do` hands off and no record write was journalled
- **THEN** the handoff still carries the line, stating that none are outstanding

#### Scenario: The status report renders the same line

- **WHEN** the read-only status report renders an `IN_PROGRESS` change's block
- **THEN** it carries the same line, from the same contract, rather than omitting it
