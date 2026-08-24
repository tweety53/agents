# myflow-state-store Specification

## Purpose
TBD - created by archiving change kan-16-myflow-stats-app. Update Purpose after archive.
## Requirements
### Requirement: The change state record lives in PostgreSQL, reached through a daemon

The authoritative home of a change's state record SHALL be a PostgreSQL database owned by a
long-running daemon. Every `/myflow-*` command SHALL read and write that record through a CLI that
speaks HTTP to the daemon, and SHALL NOT read or write the database directly.

The record SHALL carry the same fields the state file carried — `state`, `branch`, `worktrees`,
`artifactUrl`, `jiraIssue`, `planningEffort`, `models`, `reviewPanelRoster`, `prUrl`, `updatedAt`,
`updatedBy` — and SHALL be governed by the same closed-schema rule, the same monotonic-state rule,
and the same carry-forward rule the state file contract already states. Only the mechanism changes.

The record SHALL be keyed by project and change name together, so two projects may hold a change of
the same name without collision.

A write SHALL be refused when it would move the record backwards **in either dimension**: to a state
earlier in the pipeline than the stored one, or — at the same state — to an earlier recorded
instant. The recorded instant is the primary ordering and the pipeline state is the tiebreaker, so a
replayed or duplicated write can never silently overwrite a newer record with older field values.

`updatedAt` SHALL be stamped by the CLI on every write, at a precision no coarser than the precision
any other writer of the record uses, and SHALL overwrite whatever value the caller supplied. The
stamped value SHALL be the one carried by the store row, the on-disk fallback file and the journal
entry alike, so a journal entry replayed later orders by the instant its write happened at rather
than the instant of the replay. A caller MAY still supply the field; it is ignored rather than
refused, so a journal entry written before this rule still replays.

#### Scenario: A command reads state through the CLI

- **WHEN** any `/myflow-*` command needs a change's state and the daemon is reachable
- **THEN** it obtains the record through the CLI's `state get` command
- **AND** the fields returned are exactly those the state file contract names, with the same
  nullability

#### Scenario: A write renders the whole record

- **WHEN** a command writes state
- **THEN** it sends the complete object, carrying forward every field it does not itself own
- **AND** a field absent from the payload **is** cleared, which is why carrying forward is the
  caller's obligation rather than a protection the store provides

#### Scenario: A monotonic violation is refused

- **WHEN** a write would set a state earlier in the pipeline than the state already stored
- **THEN** the store refuses the write and the stored state is unchanged

#### Scenario: An out-of-order write at the same state is refused

- **WHEN** a write carries the same state as the stored record but an earlier recorded instant
- **THEN** the store refuses it and the stored record is unchanged
- **AND** the refusal is reported as the same kind of refusal a state regression produces, so a
  caller cannot mistake it for a transport failure

#### Scenario: The CLI stamps the recorded instant

- **WHEN** a caller writes state through the CLI with a payload carrying its own `updatedAt`
- **THEN** the CLI replaces that value with the current instant before the payload is sent
- **AND** the value sent to the store, the value written to the on-disk fallback file, and the value
  appended to the journal are the same instant

#### Scenario: A same-state write following a sub-second stored instant is accepted

- **WHEN** a change's stored record carries a recorded instant with sub-second precision — as a
  stage mark's synthetic bootstrap writes — and a write at that same state arrives within the same
  wall-clock second
- **THEN** the write is accepted and the stored record carries the new field values
- **AND** it is accepted because the arriving instant is stamped at the same precision, not because
  the monotonic rule was relaxed

### Requirement: The pipeline never blocks on the store

A command SHALL NOT fail, stall, or alter its behaviour because the daemon or the database is
unreachable. On any store failure — daemon down, database down, timeout, or a non-2xx response —
the CLI SHALL write the state to the local on-disk JSON path, append the intent to a journal file,
print exactly one warning line, and **exit 0**.

This posture SHALL be at least as strong as the Jira contract's "never a gate", because a state
write ends every command: a write that could fail the run would strand a change at an unwritten
state with its work already done.

#### Scenario: The daemon is unreachable during a state write

- **WHEN** a command writes state and the daemon cannot be reached
- **THEN** the CLI writes the on-disk JSON file and a journal entry, prints one warning line, and
  exits 0
- **AND** the command continues and completes exactly as it would have

#### Scenario: The database is down but the daemon is up

- **WHEN** the daemon is reachable but its database is not
- **THEN** the daemon returns a failure the CLI treats identically to being unreachable, and the
  same journal path is taken

#### Scenario: A read falls back to the local file

- **WHEN** a command reads state and the store cannot be reached
- **THEN** the CLI returns the record from the local on-disk JSON file
- **AND** reports that the value came from the fallback rather than the store

### Requirement: The journal is reconciled without losing pipeline progress

The daemon SHALL replay journal entries at startup and whenever it regains a database connection.
Conflicts SHALL resolve by `updatedAt`, with the monotonic-state rule as the tiebreaker.

A state already stored SHALL NEVER be moved backwards by a journal entry recording an earlier state,
regardless of that entry's timestamp.

A replayed entry SHALL be removed from the journal only after the store has accepted or explicitly
refused it, so an interrupted replay repeats rather than loses work.

#### Scenario: A stale journal entry cannot regress a finished change

- **WHEN** the journal holds an `IN_PROGRESS` entry for a change the store already records as
  `FINISHED`
- **THEN** replay leaves the stored `FINISHED` record unchanged
- **AND** the entry is retired from the journal rather than retried indefinitely

#### Scenario: Replay is interrupted

- **WHEN** the daemon stops partway through replaying the journal
- **THEN** entries not yet accepted or refused remain in the journal
- **AND** the next replay processes them without duplicating the ones already applied

### Requirement: A change spanning repositories is one unit of work

A change that affects more than one repository SHALL be stored as **one** change record, never as
one record per repository. The record SHALL carry the set of affected repositories, each with its
own root and its own recorded merge base.

The change's project key SHALL identify the project whose state directory owns the record, and
SHALL NOT be read as the list of affected repositories.

A recorded merge base MAY be absent for a repository, and an absent merge base SHALL mean *no merge
base recorded* — never a value to be inferred or computed.

The repository set SHALL be written in the same transaction as the change itself, and SHALL follow
the same whole-object rule the record's own fields follow: a repository absent from a write is
removed from the set.

#### Scenario: A two-repository change is stored once

- **WHEN** a change affecting two repositories is written
- **THEN** exactly one change record exists for it
- **AND** its repository set names both repositories with their respective merge bases

#### Scenario: The owning project is not the repository list

- **WHEN** a change owned by one project affects a second repository
- **THEN** the change's project key still names the owning project alone
- **AND** the second repository is present in the repository set

#### Scenario: A repository set is replaced wholesale

- **WHEN** a change previously affecting two repositories is written naming only one
- **THEN** the omitted repository is no longer in the set

#### Scenario: The repository set is transactional with the change

- **WHEN** a write of a change and its repository set fails partway
- **THEN** no reader observes the change with a partially-written repository set

### Requirement: Stage runs attribute to a repository or to the whole unit

A stage run SHALL record the repository it ran in when it ran in exactly one, and SHALL record no
repository when it belongs to the change as a whole.

An absent repository on a stage run SHALL mean *the whole unit of work*, and SHALL NOT be read as an
unknown or missing repository.

#### Scenario: A stage that ran in one repository

- **WHEN** a stage ran inside a single repository of a multi-repository change
- **THEN** its stage run names that repository

#### Scenario: A stage that belongs to the whole change

- **WHEN** a stage belongs to the change rather than to any one repository
- **THEN** its stage run records no repository, and aggregation attributes it to the whole unit

### Requirement: A recorded merge base is a sha or nothing

Each value of the `worktrees` map records that worktree's merge base. A value SHALL be either JSON
`null` — meaning *no merge base recorded*, exactly as the state file contract already defines it —
or a 40-character lowercase hexadecimal sha. No other value SHALL reach the store.

The CLI SHALL refuse a violating write **before the store is touched**, naming the offending
worktree path and the rejected value, and SHALL exit non-zero without writing the on-disk fallback
file and without appending a journal entry. A malformed merge base is a caller mistake, not a store
outage, and SHALL NOT take the never-block fallback path that a store outage takes — journalling it
would hide the bad value until the finish gate, which is the failure this requirement exists to
prevent.

The store SHALL additionally refuse a violating write with an error distinguishable from an invalid
state and from a monotonic refusal. For a write made through the CLI this refusal is unreachable,
because the CLI refused first; it covers the path that bypasses the CLI, namely journal replay of a
hand-edited or out-of-band-modified fallback file. Such an entry SHALL be retired from the journal
as a definitive outcome rather than replayed indefinitely.

A validated merge base SHALL NOT be inferred, defaulted, or repaired. `null` remains the refusal to
infer one, never a licence to.

#### Scenario: A sha is accepted

- **WHEN** a write carries a `worktrees` value that is a 40-character lowercase hexadecimal string
- **THEN** the CLI sends it and the store records it as that worktree's merge base

#### Scenario: An unrecorded merge base is accepted

- **WHEN** a write carries a `worktrees` value of JSON `null`
- **THEN** the CLI sends it and the store records no merge base for that worktree
- **AND** every rule the pipeline already states for a missing recorded merge base applies unchanged

#### Scenario: A worktree path in the merge base position is refused at the CLI

- **WHEN** a write carries a `worktrees` value that is neither `null` nor a 40-character lowercase
  hexadecimal string — an absolute path, a short sha, an uppercase sha, or an empty string
- **THEN** the CLI refuses the write, names the offending worktree path and the rejected value, and
  exits non-zero
- **AND** no on-disk fallback file is written and no journal entry is appended

#### Scenario: Journal replay refuses a hand-edited merge base

- **WHEN** journal replay applies an entry whose `worktrees` carries a value that is neither `null`
  nor a 40-character lowercase hexadecimal string
- **THEN** the store refuses it with an error distinguishable from an invalid state and from a
  monotonic refusal
- **AND** the entry is retired from the journal rather than replayed again

