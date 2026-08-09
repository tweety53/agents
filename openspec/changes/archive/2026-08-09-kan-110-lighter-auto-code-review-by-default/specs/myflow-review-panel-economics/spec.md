## ADDED Requirements

### Requirement: Bugbot's brief includes reasoned mutation testing

Where the panel dispatches Bugbot, its dispatch prompt SHALL include a mutation-testing brief: for
each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a boundary, remove
a branch — and establish whether an existing test fails as a result. A mutation no test catches is a
**surviving mutant**.

This SHALL be reasoned mutation testing performed by the reviewer. No mutation-testing framework
SHALL be added, adopted or executed by this requirement, and no mutation score SHALL be computed.

The brief SHALL be carried by the dispatch prompt rather than by the agent definition. Bugbot is
dispatched by `subagent_type` and carries its own definition, which the dispatcher does not edit;
the prompt is therefore the only place this instruction can live, and passing a model override
remains forbidden.

A surviving mutant SHALL be raised as an ordinary finding — an `F<n>` row in the findings table and
a marker line in the marker block — and SHALL block the handoff under the existing zero-open-findings
bar until a test is added or the operator withdraws the finding with a reason. It SHALL NOT be
recorded as an advisory note outside the findings table: a second class of finding that nothing
enforces would sit beside a bar that enforces every other one.

Presets that do not dispatch Bugbot SHALL NOT acquire the mutation reading by another route, and
this requirement SHALL NOT add a slot to any preset. `myflow-review-panel-roster` is canonical for
which presets dispatch Bugbot.

#### Scenario: The brief reaches Bugbot through its prompt

- **WHEN** the panel dispatches Bugbot
- **THEN** its prompt carries the mutation-testing brief
- **AND** no model override is passed to it

#### Scenario: A surviving mutant blocks the handoff

- **WHEN** Bugbot finds a changed behaviour whose mutation no existing test would catch
- **THEN** the panel record carries an `F<n>` row and an `open` marker line for it
- **AND** `/myflow-do` does not hand off until it is fixed or withdrawn with a reason

#### Scenario: No mutation tooling is introduced

- **WHEN** the panel runs with the mutation brief in force
- **THEN** no mutation-testing framework is installed or executed, and no mutation score is reported

#### Scenario: A preset without Bugbot gets no mutation reading

- **WHEN** a change records a preset whose required slots do not include Bugbot, and no conditional
  slot adds it
- **THEN** no slot is briefed for mutation testing
