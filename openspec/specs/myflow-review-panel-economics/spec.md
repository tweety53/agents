# myflow-review-panel-economics Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: Panel slots are assigned a model tier by the kind of work they do

Review panel slots whose work is judgment across the whole change SHALL run on the parent model.
Slots whose work is a narrowed breadth pass over a bounded diff SHALL run on the provider's
economy tier.

Slot 0 (Primary) and slot 2 (Principles) SHALL inherit the parent model and SHALL be dispatched
with no `model` override. Slot 4 (Adversarial) and slots 5+ (extra principle lenses) SHALL be
dispatched with the economy-tier slug resolved from the economic model mapping.

#### Scenario: Adversarial runs on the economy tier

- **WHEN** the panel dispatches slot 4
- **THEN** the dispatch carries the economy-tier `model` slug for the parent provider family

#### Scenario: Principles still inherits the parent model

- **WHEN** the panel dispatches slot 2
- **THEN** no `model` override is passed

#### Scenario: Primary still inherits the parent model

- **WHEN** the panel dispatches slot 0
- **THEN** no `model` override is passed

### Requirement: The economic model mapping states its own scope correctly

The mapping section in `skills/openspec-apply-superpowers/SKILL.md` SHALL be titled and worded to
cover slots 4 and 5+, and the guardrail forbidding an omitted `model` on an economy-tier reviewer
SHALL name slot 4 alongside slots 5+.

#### Scenario: Mapping scope names slot 4

- **WHEN** the economic model mapping section is read
- **THEN** its heading and body identify slots 4 and 5+ as its subjects
- **AND** no sentence restricts it to "slots 5+ only"

#### Scenario: Guardrail covers slot 4

- **WHEN** the panel guardrails are read
- **THEN** omitting `model` on slot 4 is listed as an error

### Requirement: The always-on rule layer states the same slot range

`rules/myflow-manual-review.mdc` SHALL name slots 4 and 5+ as the economy-tier slots. The rule
layer is read before any skill, so a rule that still scopes the economy tier to slots 5+ would
override the skill in practice regardless of what the skill says.

#### Scenario: The rule layer names slot 4

- **WHEN** the panel paragraph in `rules/myflow-manual-review.mdc` is read
- **THEN** it names the Adversarial reviewer (slot 4) alongside slots 5+ as economy-tier
- **AND** no sentence restricts the economy tier to the lens reviewers alone

### Requirement: Each economy slot carries its own justification

The reason a slot is economy-tier SHALL be the reason that applies to that slot. Slot 4's
justification SHALL describe the Adversarial reviewer's own bounded, already-scoped search for
regressions and test theater, and SHALL NOT reuse the narrowed-lens reasoning written for
slots 5+.

#### Scenario: Slot 4's stated reason is its own

- **WHEN** the economic model mapping's justification is read
- **THEN** slot 4 and slots 5+ are justified separately, each in its own terms

### Requirement: Skills that reference the panel agree with it

`skills/openspec-apply-fix-superpowers/SKILL.md` and
`skills/openspec-fast-path-superpowers/SKILL.md` SHALL continue to defer to
`openspec-apply-superpowers` as canonical, and any prose in them naming the economy-tier slot
range SHALL name slots 4 and 5+.

#### Scenario: Deferring skills carry no stale slot range

- **WHEN** either skill is searched for the phrase "slots 5+"
- **THEN** no occurrence describes the economic model mapping's scope as slots 5+ alone

#### Scenario: The principles prompt claims only what is true of itself

- **WHEN** `principles-reviewer-prompt.md`'s `[ECONOMIC_MODEL_SLUG]` note is read
- **THEN** it names slots 5+ as the economy slots this template serves and points at
  `adversarial-reviewer-prompt.md` for slot 4

#### Scenario: Fast path is otherwise unaffected

- **WHEN** `/myflow-fast-path` runs without `full-panel`
- **THEN** it still dispatches only the three required slots, none of which is slot 4

