## ADDED Requirements

### Requirement: Subagent token usage is attributed to the dispatch that incurred it

The harvester SHALL read each transcript line's `agentId` field and SHALL attribute every assistant
message carrying one to that dispatch, in addition to the whole-run total and the per-model breakout
it already produces.

The token delta computed for a stage run SHALL carry a per-dispatch breakout keyed by `agentId`,
written into the stage run's metrics bag under `dispatches.<agentId>`, and built the same way the
per-model breakout is built.

This breakout SHALL be additive. The metrics bag's existing `tokens.main`, `tokens.sidechain` and
`models.<model>` keys SHALL keep their current meaning and SHALL continue to be written exactly as
before, so that every existing aggregation reads unchanged and a stage run recorded before this
capability existed is not retroactively wrong.

A record carrying no `agentId` — a message from the parent session itself — SHALL contribute to the
whole-run total and to its model's breakout, and SHALL create no dispatch entry. No placeholder
identifier SHALL be fabricated for it.

#### Scenario: Two dispatches in one stage are told apart

- **WHEN** a review-panel stage dispatches three slots and one fix round
- **THEN** the stage run's metrics carry four dispatch entries, each with its own token figures, and
  the sidechain total continues to equal what it did before

#### Scenario: A parent-session message creates no dispatch entry

- **WHEN** an assistant message carries no `agentId`
- **THEN** it is counted in the run total and its model's bucket, and in no dispatch entry

#### Scenario: Existing keys are unchanged

- **WHEN** a stage run is harvested under this capability
- **THEN** its `tokens.main`, `tokens.sidechain` and `models.*` keys hold what they would have held
  before it existed

### Requirement: A dispatch's descriptors are read from its transcript's sibling meta file

Where a harvested transcript is a subagent transcript, the harvester SHALL read the sibling meta file
alongside it for that dispatch's agent type, description, model and spawn depth, and SHALL record
them with the dispatch's token figures.

A meta file that is absent or unreadable SHALL NOT prevent the dispatch's tokens from being
attributed. The dispatch entry SHALL carry its token figures with the descriptors omitted rather than
guessed, consistent with the rule that an absent value is never recorded as a value.

#### Scenario: Descriptors accompany the tokens

- **WHEN** a subagent transcript has its sibling meta file
- **THEN** the dispatch entry carries that dispatch's agent type, description, model and spawn depth

#### Scenario: A missing meta file still yields tokens

- **WHEN** the sibling meta file is absent
- **THEN** the dispatch entry carries its token figures and no descriptors, and no descriptor is
  invented
