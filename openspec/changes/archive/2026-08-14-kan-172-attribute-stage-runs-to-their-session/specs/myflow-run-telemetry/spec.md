## ADDED Requirements

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

A stage run whose session cannot be established SHALL be recorded with no session rather than with a
guessed one. Attributing one session's usage to another's stage run is worse than recording none:
both figures are then wrong and nothing distinguishes them.

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

### Requirement: A stage run is bound to its session after the mark, within a bounded window

Binding SHALL be permitted to complete after the mark is recorded, because the evidence that
identifies the session may not exist at the moment the mark is made.

Binding SHALL be attempted for a bounded period, after which the stage run SHALL be left
unattributed. An unbounded search SHALL NOT be performed, so that a harness which will never produce
the evidence costs a bounded amount of work rather than a permanent one.

Once bound, usage SHALL be attributed exactly as it is for a stage run whose session was known from
the start; binding SHALL introduce no second attribution path.

#### Scenario: The evidence appears after the mark

- **WHEN** the evidence identifying a stage run's session becomes available after the mark was
  recorded
- **THEN** the stage run is bound to that session, and usage falling in its window is attributed to
  it

#### Scenario: A harness that produces no such evidence

- **WHEN** a stage is marked on a harness that produces no session evidence at all
- **THEN** binding is attempted for the bounded period and then abandoned, and the stage run remains
  recorded and unattributed

#### Scenario: Binding does not disturb what is already attributed

- **WHEN** a stage run is bound to its session
- **THEN** usage already recorded against other stage runs is unchanged

### Requirement: Every documented stage is a stage, not a description of a command

The documented stage list SHALL name discrete stages for every command that marks them. A command's
entry SHALL NOT be prose describing the command as a whole, because a command whose entry is one
sentence can only ever record one stage and its per-stage statistics answer nothing.

#### Scenario: A command's documented stages

- **WHEN** a command that emits stage marks is examined against the documented list
- **THEN** its entry names discrete stages, each of which a command can begin and end

#### Scenario: A command whose recorded stages span its whole run

- **WHEN** a command records a single stage covering substantially its entire run
- **THEN** that is a defect in the documented stage list rather than a property of the command

### Requirement: A stage is identified by a stable key, not by its prose name

Every documented stage SHALL have a key that is unique across the whole pipeline and a human-readable
name. A mark SHALL identify its stage by key; the interface SHALL present the name.

A key SHALL be declared rather than derived from the name. A key derived from the name changes when
the name is improved, which is the outcome keys exist to prevent.

A stage's name MAY be reworded without changing what is recorded. Rewording SHALL NOT split that
stage's recorded history.

A key SHALL NOT be reused for a different stage. Two stages sharing a key would merge two different
things in every statistic, silently.

Where one command runs a stage another command defines, both SHALL record the same key, so runs of
the same stage under different commands are directly comparable.

#### Scenario: A stage's name is improved

- **WHEN** a documented stage's name is reworded
- **THEN** runs recorded before and after the rewording remain the same stage

#### Scenario: The same stage under two commands

- **WHEN** two different commands each run the same documented stage
- **THEN** both record it under the same key, and their runs are comparable as one stage

#### Scenario: A duplicate key

- **WHEN** two documented stages are given the same key
- **THEN** that is reported as a defect rather than accepted

#### Scenario: A mark naming an unknown key

- **WHEN** a mark names a key no documented stage declares
- **THEN** it is rejected, naming the documented alternatives
