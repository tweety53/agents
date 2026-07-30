## REMOVED Requirements

### Requirement: Every dispatch records the model it used, for the session that made it

**Reason**: The trailing qualifier "for the session that made it" is the part this change repeals.
The record is no longer session-scoped: `/myflow-finish` run 1 preserves the SDD ledger into the
repository, so the model record survives the worktree's removal and the change's archival. Editing
the requirement in place would leave a heading asserting a limitation its own body no longer
describes.

**Migration**: Replaced by *Every dispatch records the model it used* below, which keeps the
recording duty and the `unknown (agent-defined)` rule byte-for-byte in substance and replaces only
the scope paragraph and the scenario that stated the record is gone once the change is archived. No
ledger content, format, or field changes.

## ADDED Requirements

### Requirement: Every dispatch records the model it used

The SDD ledger SHALL record the model used for each subagent dispatch, implementer and reviewer
alike, alongside the task or slot it ran. Where the dispatcher cannot know the model — a slot
dispatched by `subagent_type`, which carries its own agent definition — the entry SHALL record
`unknown (agent-defined)`. It SHALL NOT record a guess.

A model policy that nothing records is a policy nothing can verify. The absence of this record is
how the implementer rule came to be missing: the panel's model choices were partly written down,
while the implementers' were not recorded anywhere, so no one could tell whether any policy had
been followed.

**Scope: the record outlives the change.** The ledger is authored under `.superpowers/`, which is
gitignored, in a worktree `/myflow-finish` run 2 removes — but it SHALL be preserved into the
repository before that happens, so it serves the operator and the review panel during the change and
remains answerable afterwards. The preservation duty itself belongs to `myflow-finish-cleanup`; this
requirement depends on it rather than restating it.

The `unknown (agent-defined)` value exists for the same reason it always did, and durability SHALL
NOT weaken it. Slots the panel dispatches by `subagent_type` resolve their model from their own
definition, which the dispatcher never reads. Recording a plausible model for them would put a value
in the audit trail that nothing measured — precisely the failure mode this policy exists to police,
and one this repository's own panel record has already made once. A record that persists is a
stronger reason to leave an unobserved entry unobserved, not a weaker one.

#### Scenario: A ledger entry is written for a completed task

- **WHEN** an implementer task completes and its ledger line is written
- **THEN** that line names the model the implementer ran on

#### Scenario: A slot's model is not knowable to the dispatcher

- **WHEN** a review-panel slot is dispatched by `subagent_type` and carries its own agent definition
- **THEN** its ledger entry records `unknown (agent-defined)`
- **AND** it does not record a model the dispatcher had no way to observe

#### Scenario: A reader asks which model implemented a given task

- **WHEN** the change has been integrated and archived and its worktree removed
- **THEN** the preserved ledger in the repository answers from the record rather than from the
  transcript
- **AND** the answer is available to a reader who was not present for the session

#### Scenario: Durability does not fill in an unobserved entry

- **WHEN** a ledger is preserved into the repository and contains `unknown (agent-defined)` entries
- **THEN** those entries are preserved as they stand
- **AND** no step substitutes a model the dispatcher never read
