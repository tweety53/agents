## MODIFIED Requirements

### Requirement: Panel slots are assigned a model tier by the kind of work they do

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

### Requirement: The economic model mapping states its own scope correctly

**Removed in substance:** there is no longer a mapping from parent provider family to an economy
tier, because there is no longer more than one tier. Any surviving prose SHALL state only that
every slot runs on Sonnet, and SHALL NOT describe per-slot tier reasoning, provider detection, or
a fallback slug.

#### Scenario: No provider-family table survives

- **WHEN** the apply skill and its reviewer prompts are searched
- **THEN** no table maps a parent provider family to a model slug, and no instruction tells the
  agent to detect its own provider

### Requirement: Skills that reference the panel agree with it

Every skill that dispatches or describes the review panel SHALL state that all slots run on
Sonnet, and SHALL agree with the roster and trigger table in `skills/myflow-do/SKILL.md`, which is
canonical for the panel.

#### Scenario: The panel is described in exactly one place

- **WHEN** a skill needs to describe the panel
- **THEN** it defers to `skills/myflow-do/SKILL.md` rather than restating the roster

## REMOVED Requirements

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
