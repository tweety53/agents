## ADDED Requirements

### Requirement: A guard a command invokes SHALL be reachable from that command's installed skill

Every guard script a `/myflow-*` command invokes SHALL be reachable from the skill directory
that command's own skill occupies, at `<skill-dir>/scripts/<name>`, in every harness the
installer targets. A guard reachable only from this repository's own checkout is not
installed, because the project a command runs against is almost never this repository.

The one real copy of each guard SHALL remain at the repository's root `scripts/` directory.
A skill's `scripts/` directory SHALL hold relative symlinks into it, tracked in version
control, so that the installer's existing whole-directory symlink carries them with no
installer step of its own and no second copy exists to drift.

A guard that resolves a sibling from its own directory — a shared library, a `.py` companion,
another guard — SHALL have that sibling symlinked beside it in every skill directory carrying
it. A guard shipped without its siblings fails at the moment it is needed, which is the
failure this requirement exists to prevent.

#### Scenario: A guard is invoked through an installed skill directory

- **WHEN** a command invokes a guard at `<skill-dir>/scripts/<name>` in an installed harness,
  where `<skill-dir>` is itself a symlink into this repository
- **THEN** the guard executes
- **AND** any sibling it resolves from its own directory resolves too

#### Scenario: A skill invokes a guard it does not carry

- **WHEN** a skill's text invokes a guard for which that skill's `scripts/` directory holds no
  symlink
- **THEN** the repository's own lint reports it by name and fails

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

Prose that describes **this repository's own** lint and test guards is not an invocation and
SHALL keep its repository-relative path, because there it names a file in this repository
rather than a guard a command runs.

#### Scenario: A contract loaded by two commands names one guard

- **WHEN** a contract file names a guard by basename
- **AND** it is loaded once by `/myflow-finish` and once by `/myflow-status`
- **THEN** each command resolves the guard inside its own skill directory

#### Scenario: An invoking call site carries a repository-relative path

- **WHEN** a skill's text invokes a guard by a path relative to a repository root
- **THEN** the repository's own lint reports that call site and fails

### Requirement: A guard SHALL NOT derive a repository root from a fixed depth above itself

A guard SHALL NOT compute a repository root as a fixed number of levels above its own
directory. A guard reachable through more than one directory has more than one such answer,
and the wrong one is a directory that exists — so the guard proceeds against it and reports a
confident wrong result rather than an error.

Where a guard needs a repository root and was given none, it SHALL derive it from its own
**resolved physical** location, so that the answer does not depend on which path it was
invoked through.

#### Scenario: A guard is invoked through a skill's scripts directory

- **WHEN** a guard that derives a default repository root is invoked at
  `<skill-dir>/scripts/<name>` with no explicit root argument
- **THEN** the root it derives is this repository's root
- **AND** it is not the skill directory one level above the guard

### Requirement: A command SHALL report a missing guard once, at the start of the run

A `/myflow-*` command SHALL check, once at the start of its run, that every guard it can
invoke is present in its own `scripts/` directory. On a complete set it SHALL print nothing.

On any absence it SHALL print exactly one block, naming every missing guard, the directory
searched, and the command that installs them, and stating that the affected checks will be
performed by hand. It SHALL then continue.

The check is a report and SHALL NOT be a gate. Each contract's existing hand-run fallback
still governs what happens at the call site, and the handoff SHALL say that those checks were
run manually. A guard is never skipped for want of the script.

#### Scenario: Every guard is present

- **WHEN** a command starts and finds every guard it can invoke
- **THEN** it prints nothing about guards and proceeds

#### Scenario: Some guards are missing

- **WHEN** a command starts and finds that some of the guards it can invoke are absent
- **THEN** it prints one block naming each missing guard, the directory searched, and the
  install command
- **AND** the run continues under the hand-run fallback
- **AND** the handoff records that those checks were run manually

#### Scenario: A missing guard is not reported twice

- **WHEN** a command has printed the missing-guard block and later reaches a call site for one
  of those guards
- **THEN** it performs that check by hand without printing the block again
