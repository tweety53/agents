# myflow-state-integrity Specification

## Purpose

Define what a state-file write owes the fields it does not own, and what it must announce
when it cannot recover one — so an inference never silently destroys a value it did not infer.

## Requirements

### Requirement: A self-heal rewrite carries forward every field it did not infer

State self-heal infers `state` only. When it rewrites a state file, it SHALL re-emit every other
field exactly as read — `branch`, `worktrees`, `artifactUrl`, `jiraIssue`, `prUrl` and `effort` —
because a write renders the whole object and a field omitted from the render is erased permanently.

This duty already binds every write under the State file contract. It is restated at the self-heal
path because that is the one write whose input is *inference* rather than the prior file, and an
inference has no source for a field it does not infer.

`effort` is carried forward like the others, with one difference that follows from its absence being
legal: a prior file that did not carry the key SHALL be rewritten without inventing a level, because
a fabricated default is indistinguishable from an operator's choice to anyone reading it later.

#### Scenario: An inferred state does not disturb the other fields

- **WHEN** self-heal corrects a state file whose recorded state is contradicted by artifacts
- **THEN** the written file records the inferred state
- **AND** `artifactUrl`, `jiraIssue`, `prUrl`, `effort`, `branch` and `worktrees` are byte-identical
  to what was read

#### Scenario: A read-only command's self-heal obeys the same duty

- **WHEN** a read-only command performs the self-heal correction its contract permits
- **THEN** it carries the unowned fields forward exactly as any other write would

#### Scenario: An absent effort is not invented

- **WHEN** self-heal rewrites a state file that carried no `effort` key
- **THEN** the rewritten file records no chosen level
- **AND** no default is substituted

### Requirement: A field that cannot be recovered is announced, never silently nulled

When the prior state file is missing or unparseable, the fields self-heal does not infer have no
source. The rewrite SHALL NOT record `null` for them silently. It SHALL name every field it could
not recover in the correction it announces, `effort` included.

A silent `null` here destroys the published proposal link, the link to the tracker issue, and the
record of the effort the operator chose, and nothing downstream can tell an absent value from one
that was never set. Announcing the loss is what makes it recoverable from the transcript.

#### Scenario: An unrecoverable field is named in the announcement

- **WHEN** self-heal rewrites a state file whose prior contents could not be read
- **AND** `artifactUrl`, `jiraIssue` and `effort` therefore have no source
- **THEN** the correction announced to the user names all three fields as unrecovered

#### Scenario: A recoverable field is not announced as lost

- **WHEN** self-heal rewrites a state file it could read
- **THEN** no field is reported as unrecovered, because every one of them was carried forward

#### Scenario: A legally absent effort is not reported as a loss

- **WHEN** self-heal rewrites a readable state file that predates the `effort` field
- **THEN** `effort` is not named among the unrecovered fields, because its absence is a legal value
  rather than a failed recovery

#### Scenario: The loss is visible rather than inferred from a later failure

- **WHEN** a state file loses a field to an unreadable prior read
- **THEN** the operator learns of it at the moment of correction
- **AND** not later, from a command that cannot find the proposal artifact
