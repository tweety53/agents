## REMOVED Requirements

### Requirement: The review stage runs all five guards

**Reason**: **There is no review stage.** `/myflow-review` is removed, and no verification gate
runs before a PR or a merge — so a requirement naming the stage, and the five guards it ran, has
no subject left. `scripts/test-state-advance.sh` is also gone with the script it asserted against,
`skills/myflow-state-advance/state-advance.sh`, so the count in the name is doubly historical.

Retaining the name and rewriting the body in place was the original plan, but that would have
dropped the scenarios "Reference guard is part of the verification gate" and "A failing guard
blocks the stage" from inside a `MODIFIED` block — indistinguishable, to any reader or validator,
from losing them by accident. Removing the requirement and adding its successor states the same
outcome explicitly.

**Migration**: Replaced by "The repository's guards run during implementation" below. The guards
themselves are unchanged apart from the retired harness; only when they run changes — during
`/myflow-do` rather than at a review stage.

## ADDED Requirements

### Requirement: The repository's guards run during implementation

This repository's guards — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/test-check-references.sh` and `scripts/test-setup.sh` — SHALL be the project's `## lint`
and `## test` commands in `.myflow/project.md`, run during `/myflow-do` as part of implementing
and verifying the work. A non-zero exit from any of them SHALL be treated as a failing
verification that blocks the `IN_PROGRESS` handoff.

`scripts/test-state-advance.sh` is not among them: the script it asserted against,
`skills/myflow-state-advance/state-advance.sh`, is deleted with the `myflow-state-advance`
capability. Four guards remain.

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
