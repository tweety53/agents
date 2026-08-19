## ADDED Requirements

### Requirement: A fix round mutation-proves every behaviour it changes

A panel fix round SHALL mutation-prove every **executable** behaviour its fix changes, not only the
test cases the round adds. Executable behaviour means anything a test could fail on — a guard
script, Go, TypeScript, shell. A fix round that changes only prose or markdown has no executable
behaviour to mutate and SHALL record the exemption form rather than recording nothing.

Each mutation SHALL alter **one** mechanism. Where a single revert would change state that more
than one check reads, the round SHALL split it into surgical mutations, one per mechanism. A
mutation touching shared state can produce a **false pass** — a case that passes by
cross-contamination rather than by the mutated mechanism working — and a false pass reports coverage
that does not exist, which is worse than recording no mutation at all.

The **parent** SHALL perform the mutations, after the fix subagent reports. The fix subagent SHALL
name the executable behaviours its fix changed; the parent SHALL then mutate each one and establish
that an existing test fails. A round SHALL NOT certify its own mutations, for the same reason
`/myflow-do` already re-runs every dispatched finding's reproducer itself rather than accepting the
subagent's account of its own success.

Where a mutation survives — no existing test fails — the round SHALL add the test that catches it
before the round closes. A surviving mutation found by this requirement SHALL NOT be raised as a new
finding and SHALL NOT consume an additional panel pass; this is deliberately a different disposition
from a surviving mutant Bugbot reports, because the round already has the changed behaviour in hand
and a finding would spend a full round re-discovering it.

This requirement SHALL bind every roster preset, including one that dispatches no Bugbot. The
obligation belongs to the fix round, not to a slot. It SHALL NOT add a slot to any preset and SHALL
NOT change any preset's required set. It SHALL NOT introduce a mutation-testing framework, and no
mutation score SHALL be computed.

This requirement SHALL bind a fix round's own changes and nothing else. It SHALL NOT reach task
implementation, the red-task-partner's fixups, or guard changes generally.

#### Scenario: A structural change alongside added cases is proved too

- **WHEN** a fix round both adds test cases and changes a mechanism those cases are written
  alongside — a delimiter, a protocol, a control flow
- **THEN** the parent mutates the changed mechanism as well as running the added cases
- **AND** the round does not close on the added cases passing alone

#### Scenario: A revert touching shared state is split

- **WHEN** reverting a changed behaviour in one step would also alter state a second check reads
- **THEN** the round splits it into surgical mutations, one per mechanism
- **AND** no single mutation is relied on to prove more than one mechanism

#### Scenario: A surviving mutation is repaired inside the round

- **WHEN** the parent mutates a changed behaviour and no existing test fails
- **THEN** the round adds a test that fails under that mutation before it closes
- **AND** no new `F<n>` finding is raised for it and no additional panel pass is spent

#### Scenario: A prose-only fix round

- **WHEN** a fix round changes only prose or markdown
- **THEN** it records the exemption form rather than recording nothing

#### Scenario: A preset without Bugbot is bound too

- **WHEN** a change records a roster preset whose slots do not include Bugbot
- **THEN** its fix rounds carry this obligation unchanged
- **AND** no slot is added to that preset

### Requirement: The pass log records each fix round's mutations

The parent SHALL record each fix round's mutations in the pass log entry it already writes in
`.superpowers/sdd/final-review-panel.md`, one line per executable behaviour the round changed,
followed by a declared count:

```
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

The exemption SHALL use the `none — <reason>` literal form the record already uses for
`finding-reproducer:`, so the record carries one exemption shape rather than two.

These lines SHALL be written in the pass log entry and SHALL NOT be written inside the marker block.
`check-unfinished-work.sh` requires every `finding-status:` marker to occupy one unbroken run of
consecutive lines, and derives the findings table's identifiers from lines matching
`^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A line that split that block, or that matched that row
shape, would change a guard's verdict on a record this requirement does not otherwise touch.

The record SHALL NOT be read by a guard script. This requirement SHALL NOT add one, and SHALL NOT
extend an existing guard to parse these lines. What the record exists for is to make the question
askable by the next round's reviewers, which is how the omission this requirement addresses was
found in the first place.

The fix round SHALL NOT close, and `/myflow-do` SHALL NOT reach the handoff, while an executable
behaviour the round changed carries neither a `fix-mutation:` line nor an exemption. This is the
round's own completeness condition, checked where the parent writes the record; it SHALL NOT be
expressed as a new class of finding and SHALL NOT add a gate to the handoff sequence.

#### Scenario: The lines land in the pass log

- **WHEN** a fix round that changed executable behaviour completes
- **THEN** its pass log entry carries one `fix-mutation:` line per changed behaviour and a
  `fix-mutations-total:` count
- **AND** those lines sit outside the marker block

#### Scenario: The existing guard's verdict is unchanged

- **WHEN** `check-unfinished-work.sh` reads a panel record carrying `fix-mutation:` and
  `fix-mutations-total:` lines
- **THEN** its verdict is identical to its verdict on the same record without them

#### Scenario: An undeclared behaviour holds the round open

- **WHEN** a fix round changed an executable behaviour and the pass log carries neither a
  `fix-mutation:` line nor an exemption for it
- **THEN** the parent does not close the fix round
- **AND** the run does not reach the handoff

#### Scenario: No guard is added

- **WHEN** this requirement is in force
- **THEN** no guard script reads `fix-mutation:` or `fix-mutations-total:`
