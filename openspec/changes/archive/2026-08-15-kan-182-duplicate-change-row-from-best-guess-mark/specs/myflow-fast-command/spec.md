## ADDED Requirements

### Requirement: The state gate reads with a guess but marks with the resolved name

`/myflow-fast`'s state gate SHALL read the change's state before marking, using whatever name the
invocation gives it — an argument, or a best guess on a run that has not yet resolved one. The read
creates nothing: `myflow state get` on a name the store does not hold reports no state and writes
no record, which is itself the signal that the run is a creating one.

The state-gate marks SHALL carry a resolved change name. On a creating run, where the name does not
exist until the Jira key resolves into `<key>-<slug>`, the `begin` and `end` pair SHALL fire back to
back at the point the name is fixed, before the resolve-change stage's own marks — the same shape
`/myflow-start` section A already requires of its creating run, and for the same reason.

At `IN_PROGRESS` the name is resolved before the gate in every path — an argument names the change,
and a bare invocation resolves it from names that already exist — so those marks SHALL fire where
the gate runs.

A creating run's state-gate stage therefore records a duration close to zero, because the mark
brackets nothing. That is accepted: a truthful zero is better than a duration bought by creating a
change row for a name nobody chose.

#### Scenario: A creating run gates on state before any name exists

- **WHEN** `/myflow-fast` is invoked for work that has no change yet
- **THEN** it reads the state with its best guess, takes the creating branch, and emits no mark
  until the resolved change name exists

#### Scenario: A creating run marks its state gate

- **WHEN** the Jira key resolves and the change name is fixed
- **THEN** the state-gate begin and end marks fire back to back under that name, before the
  resolve-change marks

#### Scenario: A fix run gates on state at `IN_PROGRESS`

- **WHEN** `/myflow-fast` is invoked at `IN_PROGRESS`
- **THEN** the state-gate marks fire where the gate runs, since the change name is already resolved
