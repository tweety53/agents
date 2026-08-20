## MODIFIED Requirements

### Requirement: A named guard resolves against the running command's own skill directory

`skills/myflow-contracts/pipeline.md` SHALL state, once, that a named guard resolves to
`<the running command's own skill directory>/scripts/<name>`. Every other file SHALL cite that
statement rather than restate it.

A skill or contract SHALL name an invoked guard by **basename**. It SHALL NOT give a path
relative to a repository root: such a path resolves only when the project being worked on is
this repository, which is the one case that never needs the guard shipped.

Resolution against the **running command's** skill directory is what lets a contract loaded by
more than one command name a guard at all. `skills/myflow-contracts/` is never a running
command and SHALL NOT carry a `scripts/` directory.

Prose that describes **this repository's own** lint and test guards is not an invocation, and SHALL
name the guard as `<agents repo>/scripts/<name>` rather than by a bare repository-relative path. A
bare path there resolves, for a reader standing in an installed project, against that project's own
tree — so the sentence names a file the reader may be able to write. Carrying the prefix says which
repository is meant, and it removes the need for any classifier to tell describing a guard from
running one.

#### Scenario: A contract loaded by two commands names one guard

- **WHEN** a contract file names a guard by basename
- **AND** it is loaded once by `/myflow-finish` and once by `/myflow-status`
- **THEN** each command resolves the guard inside its own skill directory

#### Scenario: An invoking call site carries a repository-relative path

- **WHEN** a skill's text invokes a guard by a path relative to a repository root
- **THEN** the repository's own lint reports that call site and fails

#### Scenario: Prose about this repository's own guard names its repository

- **WHEN** an installed file describes, without invoking, a guard belonging to this repository
- **THEN** it writes `<agents repo>/scripts/<name>`
- **AND** a bare `scripts/<name>` in that position is reported by the repository's own lint
