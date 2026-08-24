## ADDED Requirements

### Requirement: A dispatch records the diff base its subagent read from

A dispatch row SHALL be able to carry the **diff base** its subagent read from: the commit sha a
review slot's diff was computed against. It SHALL be written by the opening call, where the caller
knows it, and SHALL be optional — a dispatch that read a whole-change diff, and every dispatch that
is not a review slot, records none.

The rendered SDD ledger SHALL emit the recorded base for a dispatch that carries one, beside the
commit it produced, and SHALL emit nothing for a dispatch that carries none. The rendered review
panel record SHALL NOT be changed by this field, so the marker contract the panel guards parse is
untouched.

An absent base SHALL mean *not recorded* and never a sha of zero length matched against another
absent one, for the reason an absent agent identifier already means *not reported*.

Nothing in the pipeline SHALL schedule off this field. It is the durable audit trail for what a
review slot read; a panel runs inside one command invocation and reads its own in-session value,
which is what keeps this write on the never-block guarantee every record write carries — an
unreachable store journals the intent and the panel proceeds unaffected.

#### Scenario: A panel slot records what it read from

- **WHEN** a review slot is dispatched against a delta computed from sha `A`
- **THEN** its dispatch row records `A` as the diff base, and the rendered ledger names it

#### Scenario: A dispatch that read the whole change records no base

- **WHEN** the primary reviewer is dispatched against the whole-change diff
- **THEN** its row records no diff base, and the rendered ledger emits no base line for it

#### Scenario: An implementer dispatch is unaffected

- **WHEN** an implementer subagent is recorded for a task
- **THEN** its row carries no diff base and its ledger entry is unchanged

#### Scenario: The panel record is unchanged

- **WHEN** the review panel record is rendered for a change whose panel dispatches recorded bases
- **THEN** the record's marker lines, totals and findings table are byte-for-byte what they would be
  without the field
