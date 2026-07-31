# myflow-planning-gate Specification

## Purpose
TBD - created by archiving change kan-17-finish-gate-jira-and-commit-hygiene. Update Purpose after archive.
## Requirements

### Requirement: An unresolved question is put to the operator, never resolved by assumption

When `/myflow-start` reaches a question its inputs do not answer — scope, decomposition, where a
requirement belongs, which of two behaviours was meant — it SHALL put that question to the operator
rather than pick an answer and proceed.

This SHALL hold at every planning effort level. A lower level MAY group questions into fewer rounds
and MAY batch related questions into one prompt; it SHALL NOT convert a question into an assumption.

Work that does **not** depend on the answer SHALL still be done rather than held behind it, and the
question SHALL be put at the point the answer is first needed.

#### Scenario: An ambiguous ask is asked about

- **WHEN** an issue's text can be read two ways that would produce materially different work
- **THEN** `/myflow-start` puts the readings to the operator rather than choosing one

#### Scenario: A low effort level still asks

- **WHEN** a change is planned at the lowest effort level and an unresolved question arises
- **THEN** the question is still put to the operator, grouped with others rather than dropped

#### Scenario: Independent work is not blocked on an answer

- **WHEN** a question blocks only part of the planning
- **THEN** the parts that do not depend on it are completed rather than deferred behind it

### Requirement: Every approval or choice is offered as options, not as open prose

Wherever a `/myflow-*` command asks the operator to approve something or to choose between courses
of action, it SHALL present named options to select from, with the recommended option named as such.

An open-ended question — of the form "does this look right, or is there something you want
changed?" — SHALL NOT be the mechanism by which an approval gate is passed.

Each option SHALL state what will happen if it is chosen, so the choice can be made without reading
the surrounding text.

#### Scenario: The design approval gate offers options

- **WHEN** `/myflow-start` presents a design for approval
- **THEN** the operator is offered named options, including approving as-is and requesting changes

#### Scenario: A recommendation is marked

- **WHEN** a command offers options and one of them is recommended
- **THEN** that option is identified as the recommendation rather than left for the operator to infer

#### Scenario: An open-ended prompt does not stand in for an approval

- **WHEN** a command needs approval to proceed
- **THEN** it does not substitute a free-text question for the options
