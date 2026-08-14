## ADDED Requirements

### Requirement: A recorded but unmeasured run is distinguishable from an absent one

The interface SHALL distinguish three states rather than two: a period in which nothing was
recorded, a run that was recorded but received no measurement, and a value measured as zero.

A run that was recorded and never attributed SHALL NOT be presented the same way as a period with no
runs. Where every run in a period is unattributed, the interface SHALL say so, so that a
misconfiguration in recording is distinguishable from a genuinely quiet period.

#### Scenario: A run recorded without measurement

- **WHEN** a stage run exists for the period but received no usage
- **THEN** it is shown as recorded and unmeasured, distinctly from a run measured as zero

#### Scenario: A period whose runs are all unmeasured

- **WHEN** every stage run in the requested period is unattributed
- **THEN** the interface reports that runs were recorded but none was measured, rather than
  reporting that no data was recorded
