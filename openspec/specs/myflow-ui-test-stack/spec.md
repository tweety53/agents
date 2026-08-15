# myflow-ui-test-stack Specification

## Purpose
A UI-testing stack separate from the live one: what is isolated and what deliberately is not,
how a session is pointed at it, that it resets on every bring-up and starts populated, and the
guard that keeps its destructive paths off any other database.

## Requirements


### Requirement: UI testing runs against a stack separate from the live one

UI testing of the stats application SHALL run against a database and an application instance
separate from the ones holding recorded data.

The separation SHALL isolate the logical resources — the database, the application port, the
transcript source and the fallback state directory — rather than duplicating the database service
that holds them.

The test stack's transcript source SHALL be separate from the live one. The daemon harvests
transcripts on a short interval, so a test stack left pointed at the live transcript source ingests
genuine session data on its own, without anyone asking it to, and reproduces the mixing this
separation exists to prevent. The same SHALL hold for the fallback state directory.

The test stack SHALL be created by the project's normal migration and seeding path, never by copying
the live database.

#### Scenario: Bringing up the test stack

- **WHEN** the operator brings up the UI-test stack
- **THEN** it serves the application on its own port, backed by its own database, and the live
  database and the live application instance are unchanged

#### Scenario: A UI test session writes data

- **WHEN** a UI test session creates projects, changes or stage runs
- **THEN** those rows exist only in the test database, and the live database gains no rows

#### Scenario: The test stack runs while the live stack is running

- **WHEN** both stacks are up at the same time
- **THEN** neither interferes with the other's port, database, transcript source or state directory

### Requirement: The test stack resets on every bring-up

Bringing up the UI-test stack SHALL discard the previous session's contents and recreate the test
database from migrations before seeding it.

A test stack whose contents persist between sessions accumulates the same way the live database did,
and the value of a separate stack is that its contents are known at the start of every session.

#### Scenario: A second session after a polluting first one

- **WHEN** a session leaves arbitrary rows in the test database and the stack is brought up again
- **THEN** the new session starts from the fixture alone, carrying nothing from the previous one

### Requirement: The test stack starts populated

The UI-test stack SHALL be seeded with a committed fixture covering the states and quantities the
application's views display, so that bringing it up yields a populated interface rather than an
empty one.

The fixture SHALL be expressed against the application's own persistence types rather than as raw
statements, so that a schema change breaks the fixture's build rather than leaving it to fail
silently at load time or, worse, to load into a shape the views render wrongly.

#### Scenario: The operator opens the seeded UI

- **WHEN** the operator opens the test stack's interface after bring-up
- **THEN** the views render the fixture's projects, changes across every pipeline state, and stage
  runs carrying usage

#### Scenario: A migration changes a persisted shape

- **WHEN** a migration changes a column the fixture writes
- **THEN** the fixture fails to build, rather than loading and rendering incorrectly

### Requirement: A single variable targets the test stack

The command-line client SHALL resolve its default daemon address from the environment, so that one
exported variable directs an entire session at the test stack.

An explicitly passed address SHALL take precedence over the environment variable, which SHALL take
precedence over the built-in default.

Requiring the address to be passed per invocation is what allowed the live database to be polluted:
the flag already existed and was simply not passed. A mechanism that must be remembered once per
session, rather than once per command, is the change.

#### Scenario: The variable is exported

- **WHEN** the variable naming the test stack's address is exported and a client command is run with
  no address flag
- **THEN** the command targets the test stack

#### Scenario: A flag is passed as well

- **WHEN** the variable is exported and a command also passes an explicit address
- **THEN** the explicit address is used

#### Scenario: Neither is set

- **WHEN** neither the variable nor the flag is set
- **THEN** the command targets the live stack, exactly as it does today

### Requirement: Destructive test-stack paths refuse to act on any other database

Every path that seeds or destroys the UI-test database SHALL verify that its target names the test
database, and SHALL refuse with a reported reason otherwise.

The check SHALL run before any statement is issued against the target.

This makes the separation structural rather than conventional. The live database was polluted while
the rule "do not point this at the live database" was in force as advice; a guard that refuses is
the difference between a mistake that is prevented and one that is discovered afterwards by reading
the data.

The application instance itself SHALL NOT carry this restriction — serving the live database is its
ordinary purpose. The guard belongs to the seeding and teardown paths alone.

#### Scenario: Seeding aimed at the live database

- **WHEN** the seeder's target names the live database
- **THEN** it refuses, names the reason, and writes nothing

#### Scenario: Teardown aimed at the live database

- **WHEN** the teardown's target names the live database
- **THEN** it refuses, names the reason, and drops nothing

#### Scenario: Both aimed at the test database

- **WHEN** either path's target names the test database
- **THEN** it proceeds
