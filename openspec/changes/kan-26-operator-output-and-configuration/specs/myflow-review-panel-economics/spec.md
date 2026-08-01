## MODIFIED Requirements

### Requirement: Every review-panel slot runs on Sonnet

Every review-panel slot SHALL run on **Sonnet** by default. The roster and the trigger table that
selects optional slots are unchanged; only model selection is.

Slots the panel spawns directly — the primary reviewer, the principles reviewer, the adversarial
reviewer, and any extra principle lens — SHALL pass Sonnet explicitly. Slots dispatched by
`subagent_type` (Bugbot, Security Review) carry their own agent definitions and SHALL NOT be
given a model override.

No slot SHALL inherit the parent model, and no slot SHALL resolve a model from the parent's
provider family.

The default SHALL yield to an explicit operator override, and only to that — whether the operator
recorded a review-panel model in the change's state file or instructed one during the session.
`myflow-model-policy` is canonical for how such an override is given, recorded and applied; this
requirement defers to it rather than restating the mechanism.

**Only the slots that take an override are affected.** Bugbot and Security Review are dispatched by
`subagent_type` and SHALL still receive no model override, whatever is recorded.

The default is Sonnet for the reason it always was, and an override does not weaken it: a reviewer
supplies one of many independent readings of a finished diff, so the panel's cost must not scale
with whichever model the operator happens to be running. An override is a deliberate, recorded
decision for one change, not an inheritance path — which is what "no slot SHALL inherit the parent
model" continues to forbid.

#### Scenario: Every directly-spawned slot names Sonnet

- **WHEN** the panel dispatches the primary, principles, adversarial or extra-lens reviewers for a
  change that records no panel model
- **THEN** each is given Sonnet explicitly, and none omits the model in order to inherit the
  parent's

#### Scenario: A stronger parent model does not raise the panel's cost

- **WHEN** the parent agent is running on Opus and no override was given
- **THEN** the panel still runs entirely on Sonnet

#### Scenario: A recorded override raises the panel

- **WHEN** the change's state file records a review-panel model other than Sonnet
- **THEN** the slots that take an override are dispatched on that model
- **AND** this is an override rather than inheritance, because the parent's model played no part in
  selecting it

#### Scenario: An override does not reach the subagent_type slots

- **WHEN** a panel model is recorded and Bugbot or Security Review is dispatched
- **THEN** no model override is passed to that slot
