## MODIFIED Requirements

### Requirement: The review stage runs all five guards

**There is no review stage.** `/myflow-review` is removed, and no verification gate runs before a
PR or a merge.

This repository's guards — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/test-check-references.sh` and `scripts/test-setup.sh` — SHALL instead be the project's
`## lint` and `## test` commands in `.myflow/project.md`, run during `/myflow-do` as part of
implementing and verifying the work. A non-zero exit from any of them SHALL be treated as a
failing verification that blocks the `IN_PROGRESS` handoff.

`scripts/test-state-advance.sh` is no longer among them: the script it asserted against,
`skills/myflow-state-advance/state-advance.sh`, is deleted with the `myflow-state-advance`
capability. Four guards remain, and the requirement's name is retained so the delta applies
cleanly; both the count and the stage in the name are historical.

#### Scenario: The guards run during implementation

- **WHEN** `/myflow-do` verifies work in this repository
- **THEN** all four scripts run and their output is shown as evidence

#### Scenario: Each guard's own assertion harness is invoked

- **WHEN** the guards run
- **THEN** `scripts/test-check-references.sh` is among them, so that harness is not a guard
  nothing invokes

#### Scenario: A retired harness is not invoked

- **WHEN** the guards run
- **THEN** `scripts/test-state-advance.sh` is not run and does not exist

#### Scenario: A failing guard blocks the handoff

- **WHEN** `scripts/check-references.sh` exits non-zero during `/myflow-do`
- **THEN** the change does not hand off to the human gate and the failure is reported

#### Scenario: Nothing re-runs them before integration

- **WHEN** `/myflow-finish` integrates the change
- **THEN** it runs none of the guards, because verification happened during `/myflow-do`
