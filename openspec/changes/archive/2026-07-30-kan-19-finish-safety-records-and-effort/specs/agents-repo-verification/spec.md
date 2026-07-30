## MODIFIED Requirements

### Requirement: This repository declares its own myflow project configuration

`.myflow/project.md` SHALL exist in this repository and SHALL declare `## apps`, `## test`,
`## lint`, `## standards`, and `## jira`. `## test` SHALL name `scripts/test-setup.sh`,
`scripts/test-check-references.sh`, `scripts/test-check-plan-provenance.sh`,
`scripts/test-check-finish-preflight.sh` and `scripts/test-preserve-session-records.sh`. `## lint`
SHALL name `scripts/check-vocabulary.sh`, `scripts/check-references.sh` and
`scripts/check-plan-provenance.sh`, and SHALL state explicitly that no auto-fix command exists.

`## lint` SHALL name every guard that **scans the repository tree** and can therefore run at any
time. A check that requires context no tree scan provides — a branch, a worktree, a resolved base ref
— SHALL NOT be named in `## lint`, because a lint step that cannot run outside a change would fail on
every unrelated invocation. Such a check is invoked by the command that owns it, and its harness is
named in `## test` like any other.

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

#### Scenario: A new tree-scanning guard is reachable through the declared configuration

- **WHEN** a guard script that scans the repository tree is added to this repository
- **THEN** it is named in `## lint` and its harness in `## test`
- **AND** a myflow run verifying this repository invokes it without any per-guard special-casing

#### Scenario: A context-requiring check is tested but not linted

- **WHEN** a check is added that needs a branch, worktree or resolved base ref to run
- **THEN** its harness is named in `## test`
- **AND** it is not named in `## lint`, because it cannot run against a bare tree
