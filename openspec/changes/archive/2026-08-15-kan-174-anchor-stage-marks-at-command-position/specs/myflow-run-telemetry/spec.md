## MODIFIED Requirements

### Requirement: A stage mark carries a correlator that identifies its session

Every `stage begin` SHALL carry a correlator identifying the session it was made in, and the harness
it ran under.

The correlator SHALL identify a **session**, not a mark: one correlator is generated per command run
and passed on every mark that run makes. Once a correlator is bound to a session, every later mark
carrying it SHALL bind immediately, without repeating the work that established the binding.

A correlator SHALL be generated fresh per run. A correlator reused across concurrent runs would
resolve to more than one session and so identify none.

The correlator SHALL be a literal value written by the calling command — never a value produced by
expansion at execution time, which would be recorded identically for every caller and so identify
nothing.

The harness SHALL likewise be the harness the command is actually running under, never a fixed value
written into shared command text, since one source is installed into several harnesses.

**A command SHALL be recognised as a mark by the invocation it performs, not by where that
invocation sits in the command text.** Marks are emitted inside shell blocks carrying variable
assignments, directory changes and other statements; a rule that requires the invocation to lead the
text rejects the shapes actually emitted, and a mark that is not recognised binds nothing at all.

Recognition SHALL NOT be satisfied by a command that merely mentions a correlator. Where recognition
admits a command that only reproduces a mark's text without performing it, that SHALL be documented
rather than left implicit — **a false negative here is silent and total, while a false positive is
narrow and, where two sessions match, already refused by the ambiguity rule.**

#### Scenario: A mark inside a larger shell block

- **WHEN** a mark is emitted after variable assignments, a directory change, or on a later line of a
  multi-statement command
- **THEN** it is recognised, and its stage run binds to the session that emitted it

#### Scenario: A mark is emitted

- **WHEN** a pipeline command marks the start of a stage
- **THEN** the mark carries its run's correlator and the harness it ran under

#### Scenario: Later marks in the same run

- **WHEN** a run makes further marks after its correlator has been bound to a session
- **THEN** those marks are attributed to that session without re-establishing the binding

#### Scenario: Two sessions mark stages in the same directory at the same time

- **WHEN** two sessions each mark a stage while working in one directory
- **THEN** each stage run is bound to the session that marked it, and neither receives the other's
  usage

#### Scenario: A correlator that could identify more than one session

- **WHEN** a correlator matches more than one session
- **THEN** no session is recorded for that stage run, and the ambiguity is reported rather than
  resolved by choosing

## ADDED Requirements

### Requirement: A state gate reads the state before it marks

A command that both gates on a change's state and marks its own stages SHALL read that state
**before** emitting its first mark.

A mark SHALL NOT cause the state it is about to be gated on to come into existence. Where marking
creates a record for a change that has none, a command that marks first observes a state its own
mark authored, and can refuse to proceed on the strength of it.

A record whose only author is a mark's own side effect SHALL NOT satisfy a state gate that expects a
state a pipeline command wrote.

#### Scenario: A creating run marks its state gate

- **WHEN** a command that accepts no existing state marks its own state-gate stage
- **THEN** it still proceeds as a creating run, rather than refusing on a state its mark created
