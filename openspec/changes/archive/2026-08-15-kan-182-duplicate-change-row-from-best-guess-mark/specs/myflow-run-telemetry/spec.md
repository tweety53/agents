## ADDED Requirements

### Requirement: A mark names a resolved change, never a guess

The `<change>` argument of a stage mark SHALL be a resolved change name. A command SHALL NOT emit a
mark naming a change name it has not yet resolved — a best guess, a bare issue key awaiting its
slug, or any other stand-in for a name the run is still working out.

Marking is not a read. Where the store has no such change, the begin handler bootstraps a change row
so the mark has something to attach to, so a guessed name becomes a change row that outlives the run
that guessed it: it appears among the open changes, carries a next command, and is never archived,
because no change directory bears that name.

This is the sibling of "A state gate reads the state before it marks", and rests on the same
property of marking: a mark writes. That requirement keeps a command from *reading* a state its own
mark authored; this one keeps a command from *creating* a change nobody named.

The bootstrap itself SHALL remain unchanged. A synthetic row is how an unattributable mark is made
visible rather than dropped, and suppressing it — in the store, in a query, or in a view — would
remove the report while leaving the cause in place.

A skill's `stage begin` call sites SHALL be checked mechanically for a change argument written as a
placeholder that names a guess, alongside the existing checks on that call's session token and
harness.

#### Scenario: A command resolves its change name after gating on state

- **WHEN** a command must read a change's state before that change's name is finally resolved
- **THEN** it performs the read with whatever name it has, and emits its state-gate marks only once
  the name is resolved

#### Scenario: A skill names a guess in a mark

- **WHEN** a `stage begin` call in skill source writes its change argument as a placeholder naming a
  guess
- **THEN** the guard over stage-mark calls rejects it, naming the file and line

#### Scenario: A mark names a change the store has never seen

- **WHEN** a mark carries a resolved change name the store does not hold
- **THEN** the synthetic change row is bootstrapped exactly as before, and the mark is recorded
  rather than dropped
