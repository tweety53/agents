## MODIFIED Requirements

### Requirement: This repository declares its own myflow project configuration

`.myflow/project.md` SHALL exist in this repository and SHALL declare `## apps`, `## test`,
`## lint`, `## standards`, `## jira`, and `## workspace isolation`. `## test` SHALL name
`scripts/test-setup.sh`, `scripts/test-check-references.sh`,
`scripts/test-check-plan-provenance.sh` and `scripts/test-check-finish-preflight.sh`. `## lint`
SHALL name `scripts/check-vocabulary.sh`, `scripts/check-references.sh` and
`scripts/check-plan-provenance.sh`, and SHALL state explicitly that no auto-fix command exists for
the non-Go half of the repository.

`## lint` SHALL name every guard that **scans the repository tree** and can therefore run at any
time. A check that requires context no tree scan provides — a branch, a worktree, a resolved base ref
— SHALL NOT be named in `## lint`, because a lint step that cannot run outside a change would fail on
every unrelated invocation. Such a check is invoked by the command that owns it, and its harness is
named in `## test` like any other.

**This repository holds a runnable application**, so `## apps` SHALL describe it alongside the
skills-and-rules half that has no port and no URL, and the `## workspace isolation` section SHALL
declare the resources an apply worktree runs against its own copy of. A repository that ships an
application and declares no isolation gives every apply worktree the main checkout's own database and
port, which is the failure the isolation contract exists to prevent.

The `## workspace isolation` section SHALL record, in the prose beside its tables, which resources
this repository has deliberately not isolated, so that an absent row reads as a decision rather than
an oversight.

#### Scenario: myflow reads this repository's own commands

- **WHEN** a myflow change in this repository reaches its verification step
- **THEN** the verification it runs is the guard scripts and their harnesses
- **AND** no Gradle, npm, or other unrelated project's command is invoked

#### Scenario: The absent auto-fix step is explicit

- **WHEN** `.myflow/project.md` `## lint` is read
- **THEN** it states which half of the repository has an auto-fix command and which has none
- **AND** an agent following the Lint Fix Priority rule can tell the auto-fix step is
  inapplicable rather than skipped

#### Scenario: The apps section describes the repository's runnable application

- **WHEN** `.myflow/project.md` `## apps` is read
- **THEN** it names the stats daemon, its repository root and its local URL, alongside the
  skills-and-rules half that has neither
- **AND** it names the guard scripts and the sandboxed `setup.sh` run as that half's verification
  surface

#### Scenario: The isolation declaration is validated rather than merely present

- **WHEN** the repository's lint runs `scripts/check-workspace-isolation.sh`
- **THEN** the declared section is checked against the contract's rules and reported well formed
- **AND** the run is no longer a silent pass earned by declaring nothing

#### Scenario: A new tree-scanning guard is reachable through the declared configuration

- **WHEN** a guard script that scans the repository tree is added to this repository
- **THEN** it is named in `## lint` and its harness in `## test`
- **AND** a myflow run verifying this repository invokes it without any per-guard special-casing

#### Scenario: A context-requiring check is tested but not linted

- **WHEN** a check is added that needs a branch, worktree or resolved base ref to run
- **THEN** its harness is named in `## test`
- **AND** it is not named in `## lint`, because it cannot run against a bare tree
