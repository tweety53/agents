## MODIFIED Requirements

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

## ADDED Requirements

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
