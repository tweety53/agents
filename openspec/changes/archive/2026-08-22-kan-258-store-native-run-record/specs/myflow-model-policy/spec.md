## MODIFIED Requirements

### Requirement: Every dispatch records the model it used

The **store** SHALL record the model used for each subagent dispatch, implementer and reviewer alike,
alongside the task or slot it ran. Where the dispatcher cannot know the model — a slot dispatched by
`subagent_type`, which carries its own agent definition — the record SHALL carry
`unknown (agent-defined)`. It SHALL NOT record a guess.

A model policy that nothing records is a policy nothing can verify. The absence of this record is how
the implementer rule came to be missing: the panel's model choices were partly written down, while
the implementers' were not recorded anywhere, so no one could tell whether any policy had been
followed.

**Scope: the record outlives the change, and no longer depends on a file having survived.** A
dispatch's model is written to the store as the dispatch closes, attached to the change by its own
identifier rather than by a path. It is therefore answerable after the worktree is removed and after
the change is archived, whether or not any Markdown rendering of it was produced. The rendered
ledger remains the form a reader reads without a daemon; it is a view of this record, not the record
itself. The render duty belongs to `myflow-finish-cleanup` and `myflow-run-record`; this requirement
depends on neither for the model to be answerable.

**The question this record answers SHALL be answerable by query, not only by reading.** "Which
changes ran a task on a given model" SHALL be a question the store can answer across changes, which
is what distinguishes a recorded policy from a policy that merely left a trail somewhere.

The `unknown (agent-defined)` value exists for the same reason it always did, and durability SHALL
NOT weaken it. Slots the panel dispatches by `subagent_type` resolve their model from their own
definition, which the dispatcher never reads. Recording a plausible model for them would put a value
in the audit trail that nothing measured — precisely the failure mode this policy exists to police,
and one this repository's own panel record has already made once. A record that persists is a
stronger reason to leave an unobserved entry unobserved, not a weaker one. No rendering step SHALL
fill such a value in on the way out of the store either.

#### Scenario: A ledger entry is written for a completed task

- **WHEN** an implementer task completes and its dispatch is recorded
- **THEN** that record names the model the implementer ran on
- **AND** the ledger entry a reader sees is a rendering of that record, not a separately authored
  line that could disagree with it

#### Scenario: A slot's model is not knowable to the dispatcher

- **WHEN** a review-panel slot is dispatched by `subagent_type` and carries its own agent definition
- **THEN** its dispatch record carries `unknown (agent-defined)`
- **AND** it does not record a model the dispatcher had no way to observe

#### Scenario: A reader asks which model implemented a given task

- **WHEN** the change has been integrated and archived and its worktree removed
- **THEN** the store answers from the record rather than from the transcript
- **AND** the answer is available to a reader who was not present for the session, and does not
  depend on any file having been preserved

#### Scenario: A reader asks the same question across every change

- **WHEN** a reader asks which changes ran a task on a given model
- **THEN** the store answers across changes, rather than the question requiring a grep of whichever
  archived records happen to have survived

#### Scenario: Durability does not fill in an unobserved entry

- **WHEN** a ledger is rendered from records containing `unknown (agent-defined)` entries
- **THEN** those entries are rendered as they stand
- **AND** no step substitutes a model the dispatcher never read
