## ADDED Requirements

### Requirement: A stage run opens onto its own dispatches

The route that shows one change's stage runs SHALL let a stage run be expanded to show the dispatches
it made, one row per dispatch, carrying that dispatch's description, agent type, model, token total
and cost, ordered by cost descending.

Per-dispatch cost SHALL be derived through the same pricing path every other cost figure uses,
applied to the dispatch's own recorded model.

A dispatch the store never measured SHALL be rendered as unavailable rather than as zero, exactly as
an unmeasured stage run already is, so that the absence of a measurement stays distinguishable from a
measurement of nothing.

This requirement SHALL NOT add a view. The count of views this capability serves is unchanged, and
per-dispatch cost is reached only by expanding a stage run on the route that already shows it.

#### Scenario: A panel stage expands to its slots

- **WHEN** a stage run that dispatched three review slots and a fix round is expanded
- **THEN** four rows are shown, the most expensive first, each naming its description, agent type and
  model

#### Scenario: An unmeasured dispatch is not drawn as zero

- **WHEN** a dispatch carries no token figures
- **THEN** its row renders as unavailable rather than as a zero cost

#### Scenario: No view is added

- **WHEN** the interface is loaded
- **THEN** it serves the same set of views it served before, and the per-dispatch rows appear only
  under an expanded stage run
