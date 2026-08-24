# myflow-daemon-single-instance Specification

## Purpose
How the stats daemon refuses to start beside a live instance of itself: what it records, which single
condition is a refusal and which are stale files it logs and starts past, and the ordering that keeps
a refused start from touching the database at all.

## Requirements

### Requirement: The daemon refuses to start beside a live instance

The stats daemon SHALL record, at startup, its own process id and the path of its own executable in
a file derived from the port it is about to serve, and SHALL remove that file when it shuts down
gracefully.

The daemon SHALL refuse to start when that file names a process that is both alive and running the
recorded executable, and SHALL say which process id holds it.

The daemon SHALL treat the file as stale — logging it, overwriting it and continuing to start —
in every other case: when the file is absent, when it does not parse, when the recorded process id
is not alive, or when it is alive but is not running the recorded executable. A process id recycled
onto an unrelated process SHALL NOT prevent the daemon from starting.

The identity check SHALL compare against the executable path recorded in the file rather than a
fixed program name, so that an instance started from a differently-named copy of the same daemon is
still recognised as one.

The file's path SHALL be derived from the daemon's resolved port, so that two instances configured
for different ports do not contend for one file. The path SHALL NOT be settable by the caller.

#### Scenario: A second daemon on the same port

- **WHEN** a daemon is running and another is started with the same port
- **THEN** the second refuses to start, naming the running process id
- **AND** the running daemon is unaffected

#### Scenario: Two daemons on different ports

- **WHEN** a daemon is running on one port and another is started on a different port
- **THEN** both run, each holding its own file

#### Scenario: The recorded process id has been recycled

- **WHEN** the recorded process id is alive but belongs to an unrelated program
- **THEN** the daemon logs the file as stale, overwrites it, and starts

#### Scenario: A daemon that was killed without shutting down

- **WHEN** a daemon is killed so abruptly that it removes nothing
- **AND** a new daemon is started on the same port
- **THEN** the new daemon starts, because the recorded process id is no longer alive

#### Scenario: An instance started from a renamed copy of the daemon

- **WHEN** a running instance was started from a copy of the daemon under a different filename
- **AND** another daemon is started with that instance's port
- **THEN** the second refuses to start

### Requirement: A refused start touches no database

The daemon SHALL acquire its port before it opens its database connection, so that a start refused
for a port or instance conflict performs no migration, no seeding and no transcript harvesting.

The daemon SHALL still refuse a non-loopback bind address before opening any listener.

#### Scenario: The port is held by another process

- **WHEN** the daemon is started and its port is already held
- **THEN** it exits reporting the bind failure
- **AND** the database it was configured against is unchanged: no migration ran, no pricing was
  seeded, and no transcript was harvested

#### Scenario: A non-loopback bind address

- **WHEN** the daemon is configured with a non-loopback host
- **THEN** it refuses before opening a listener

