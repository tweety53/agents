## ADDED Requirements

### Requirement: A message inside more than one window attributes to the window that started last

Where a transcript message's timestamp falls inside more than one stage window of the same session,
the daemon SHALL attribute it to the window with the latest start instant. Where two such windows
start at the same instant, the higher attempt SHALL win.

Iteration order, row identity and insertion order SHALL NOT decide the outcome. A stage run left
open by a dropped end mark therefore costs its own stage's measurement and nothing beyond it: it
SHALL NOT absorb usage that a later window of the same session also contains.

#### Scenario: An open stage run is left behind by a dropped end mark

- **WHEN** a stage run stays open because its end mark was never issued, and a later stage of the
  same session opens its own window and writes assistant messages
- **THEN** those messages' usage is attributed to the later stage run
- **AND** the stage run left open receives none of it

#### Scenario: A replayed begin opens a second attempt

- **WHEN** a journalled begin mark is replayed, opening a second attempt of a stage
- **AND** the replay carries the original mark's own start instant, so both attempts begin at the
  same moment
- **THEN** the earlier attempt is closed as superseded at that same instant, which leaves it an
  empty interval no message can fall inside
- **AND** every message from that instant onward is attributed to the replayed attempt

The two attempts are therefore never both open, and the same-instant rule below is not what
resolves this case. That rule is what remains for the windows this one does not cover: a run
recorded before this capability existed, a run carrying no session token, and the interval before a
supersede commits.

#### Scenario: Two windows recorded at the same instant

- **WHEN** a message falls inside two windows of one session whose start instants are identical
- **THEN** the window with the higher attempt is attributed the message

### Requirement: A begin mark closes the runs of its own session that it supersedes

A `stage begin` SHALL close every still-open stage run that carries the same session token and
started no later than the new run, recording `superseded` as that run's outcome and the new run's
own start instant as its end. The new run's insert and the closing of what it supersedes SHALL be
one atomic write: either both land or neither does.

The runs SHALL be matched by session token, not by session id, since a session id is bound after
the fact and is routinely absent when a mark is written. A run carrying no session token SHALL NOT
be closed by this rule.

A begin mark carrying a start instant earlier than an open run's SHALL leave that run open — a
replayed mark closes only what it genuinely preceded.

`superseded` SHALL be distinct from `abandoned`: an abandoned run is one the daemon closed after its
session went silent, and the rework-rate view reads that outcome directly; a superseded run is one
whose end mark never came, discovered by the next mark of the same session.

#### Scenario: The next stage of a session begins

- **WHEN** a stage run of a session is still open and the same session marks the beginning of
  another stage
- **THEN** the open run is closed with outcome `superseded` and an end instant equal to the new
  run's start instant
- **AND** the new stage run is recorded as it would have been anyway

#### Scenario: A begin mark is replayed from the journal

- **WHEN** a journalled begin mark is replayed carrying a start instant earlier than an open run of
  the same session
- **THEN** that open run stays open

#### Scenario: A mark carries no session token

- **WHEN** a stage run recorded without a session token is still open and another run of any session
  begins
- **THEN** the run without a token stays open, to be closed by an end mark or by the sweeper

#### Scenario: A superseded stage is not counted as abandoned

- **WHEN** the rework-rate view aggregates over stage runs including superseded ones
- **THEN** only runs recorded as abandoned are counted as abandoned

The scenario is titled for what it asserts. The same view's rework count is
`attempt > 1` with no outcome filter, so a superseded run at attempt 2 counts as rework exactly as
any other second attempt does — which is correct, since it really was a second attempt, and is not
what this scenario governs.
