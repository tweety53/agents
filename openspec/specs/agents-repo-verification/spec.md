# agents-repo-verification Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: This repository declares its own myflow project configuration

`.myflow/project.md` SHALL exist in this repository and SHALL declare `## apps`, `## test`,
`## lint`, `## standards`, and `## jira`. `## test` SHALL name `scripts/test-setup.sh`,
`scripts/test-check-references.sh` and `scripts/test-check-plan-provenance.sh`. `## lint` SHALL name
`scripts/check-vocabulary.sh`, `scripts/check-references.sh` and
`scripts/check-plan-provenance.sh`, and SHALL state explicitly that no auto-fix command exists.

#### Scenario: myflow reads this repository's own commands

- **WHEN** a myflow change in this repository reaches its verification step
- **THEN** the verification it runs is the guard scripts and their harnesses
- **AND** no Gradle, npm, or other unrelated project's command is invoked

#### Scenario: The absent auto-fix step is explicit

- **WHEN** `.myflow/project.md` `## lint` is read
- **THEN** it states that this repository has no auto-fix command
- **AND** an agent following the Lint Fix Priority rule can tell the auto-fix step is
  inapplicable rather than skipped

#### Scenario: The apps section describes a repository with no runnable app

- **WHEN** `.myflow/project.md` `## apps` is read
- **THEN** it states that there is no runnable application and names the guard scripts and the
  sandboxed `setup.sh` run as the verification surface

#### Scenario: A new guard is reachable through the declared configuration

- **WHEN** a guard script is added to this repository
- **THEN** it is named in `## lint` and its harness in `## test`
- **AND** a myflow run verifying this repository invokes it without any per-guard special-casing
### Requirement: A guard fails when a cross-referenced section no longer exists

`scripts/check-references.sh` SHALL, for every backticked `.md` or `.mdc` path that resolves to a
real file inside the repository, require that at least one bold token the line **associates** with
that path matches a `#`–`####` heading in the referenced file. A bold token is associated only when
it is adjacent to the path in one of the shapes references are written in, and each token is
assigned to the nearest path it is associated with. A path the line associates no token with is not
checked. The script SHALL report every failure as `file:line` and exit non-zero.

The script SHALL refuse to read any path that normalizes outside the repository root, SHALL decide
that from the path's **shape** before any existence test so the verdict does not depend on what is
on the machine, and SHALL treat such a reference as a **failure** rather than a note — a clean exit
means every reference was checked. It SHALL exit `2` rather than report a clean run when its root
is unset-but-empty, is not a directory, or contains no Markdown at all.

The script SHALL take no arguments and SHALL own its scan set in a single `DEFAULT_TARGETS`
definition, matching the contract of the repository's other guards.

#### Scenario: A moved section is caught

- **WHEN** a section is removed from a file that other files reference by name
- **THEN** the script reports each referring `file:line` and exits non-zero

#### Scenario: A live reference passes

- **WHEN** every referenced section still exists in the file it is referenced from
- **THEN** the script exits `0`

#### Scenario: Reference phrasing variants are recognized

- **WHEN** a line reads "see **State file** in `…`", "per **State file** in `…`", or "defined once
  under **State file** in `…`"
- **THEN** all three are checked identically

#### Scenario: Emphasis that is not a section reference does not fail the guard

- **WHEN** a line carries a bold token used for emphasis and, elsewhere on the same line, a
  backticked path the token is not written next to
- **THEN** the path is not checked against that token and the line does not fail

#### Scenario: An associated token beside emphasis is still checked

- **WHEN** the same line carries both unrelated emphasis and a bold token written in a reference
  shape whose section no longer exists
- **THEN** the script reports the line and exits non-zero

#### Scenario: A token is assigned to its nearest path

- **WHEN** a line names two paths and a bold token sits adjacent to the second
- **THEN** the token is required to resolve in the second file only, not in the first

#### Scenario: A reference outside the repository root is refused, not read

- **WHEN** a line references a path that normalizes outside the repository root
- **THEN** the file is not opened, the refusal is reported as a failure, and the script exits
  non-zero

#### Scenario: Containment does not depend on what exists on the machine

- **WHEN** two such references are checked, one whose out-of-tree target exists and one whose does
  not
- **THEN** both are refused, reported, and fail identically

#### Scenario: An unscannable root never reports clean

- **WHEN** the root override is empty, names a nonexistent directory, or holds no Markdown files
- **THEN** the script exits `2` instead of reporting that all references resolve

#### Scenario: The script is argument-free and self-scoped

- **WHEN** the script is invoked with no arguments from any directory in the repository
- **THEN** it scans its own `DEFAULT_TARGETS` and does not require a path list

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

