## ADDED Requirements

### Requirement: Every review-panel slot runs on Sonnet

Every review-panel slot SHALL run on **Sonnet**. The roster and the trigger table that selects
optional slots are unchanged; only model selection is.

Slots the panel spawns directly — the primary reviewer, the principles reviewer, the adversarial
reviewer, and any extra principle lens — SHALL pass Sonnet explicitly. Slots dispatched by
`subagent_type` (Bugbot, Security Review) carry their own agent definitions and SHALL NOT be
given a model override.

No slot SHALL inherit the parent model, and no slot SHALL resolve a model from the parent's
provider family.

#### Scenario: Every directly-spawned slot names Sonnet

- **WHEN** the panel dispatches the primary, principles, adversarial or extra-lens reviewers
- **THEN** each is given Sonnet explicitly, and none omits the model in order to inherit the
  parent's

#### Scenario: A stronger parent model does not raise the panel's cost

- **WHEN** the parent agent is running on Opus
- **THEN** the panel still runs entirely on Sonnet

### Requirement: No provider-family model mapping survives

There SHALL be no mapping from a parent provider family to an economy tier, because there is no
longer more than one tier. Surviving prose SHALL state only that every slot runs on Sonnet, and
SHALL NOT describe per-slot tier reasoning, provider detection, or a fallback slug.

#### Scenario: No provider-family table survives

- **WHEN** the apply skill and its reviewer prompts are searched
- **THEN** no table maps a parent provider family to a model slug, and no instruction tells the
  agent to detect its own provider

### Requirement: Skills that describe the panel defer to the canonical roster

Every skill that dispatches or describes the review panel SHALL state that all slots run on
Sonnet, and SHALL agree with the roster and trigger table in `skills/myflow-do/SKILL.md`, which is
canonical for the panel.

#### Scenario: The panel is described in exactly one place

- **WHEN** a skill needs to describe the panel
- **THEN** it defers to `skills/myflow-do/SKILL.md` rather than restating the roster

## REMOVED Requirements

### Requirement: Panel slots are assigned a model tier by the kind of work they do

**Reason**: There are no longer two tiers to assign between. The requirement split the panel into
slots that inherit the parent model (Primary, Principles) and slots dispatched on an economy tier
(Adversarial, extra lenses); its scenarios asserted exactly that split — "Primary still inherits
the parent model", "Principles still inherits the parent model", "Adversarial runs on the economy
tier". This change puts every slot on Sonnet and forbids inheriting the parent model at all, so
all three assertions become false rather than merely reworded.

**Migration**: Replaced by "Every review-panel slot runs on Sonnet" above. Slots dispatched by
`subagent_type` (Bugbot, Security Review) keep their own agent definitions and take no override.

### Requirement: The economic model mapping states its own scope correctly

**Reason**: The requirement governed how the mapping section scoped itself — that its heading and
body name slots 4 and 5+, and that the guardrail forbidding an omitted `model` names slot 4. With
one tier there is no mapping section to scope and no omitted-`model` error to guard: omitting the
override is now correct for the `subagent_type` slots and wrong for every other, decided by how
the slot is dispatched rather than by which tier it belongs to.

**Migration**: Replaced by "No provider-family model mapping survives" above, which asserts the
mapping's absence rather than its scope.

### Requirement: The always-on rule layer states the same slot range

**Reason**: The always-on rule layer no longer describes the review panel at all. It carries only
the pipeline trigger and the contract pointers, so there is no slot range in it that could agree
or disagree with the panel.

**Migration**: None required. The panel's roster lives in `skills/myflow-do/SKILL.md`, which is
canonical for it.

### Requirement: Each economy slot carries its own justification

**Reason**: With every slot on Sonnet there are no economy slots, so there is nothing to justify.
The requirement existed to stop one slot's rationale being borrowed for another; a single tier
removes the possibility.

**Migration**: None required.

### Requirement: Skills that reference the panel agree with it

**Reason**: The requirement was written in terms of a slot range the deferring skills had to keep
in sync — "any prose naming the economy-tier slot range SHALL name slots 4 and 5+" — and two of
the three skills it named, `openspec-apply-fix-superpowers` and `openspec-fast-path-superpowers`,
are removed by this change along with the fast path itself. Its scenarios assert a stale slot
range, a `[ECONOMIC_MODEL_SLUG]` note that no longer exists, and fast-path behaviour that has no
subject.

**Migration**: Replaced by "Skills that describe the panel defer to the canonical roster" above,
which states the same deference without a slot range to keep in sync.
