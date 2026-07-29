# myflow-model-policy Specification

## Purpose

Define which model each subagent role runs on, why the implementer and reviewer defaults
point in opposite directions, how an operator overrides either, and the record that makes the
policy auditable for the session that made it.

## Requirements

### Requirement: Implementer subagents run on the strongest available model

Implementer subagents dispatched by `/myflow-do` SHALL run on Opus, or the harness's strongest
available model. The model SHALL be named explicitly at dispatch, never left to inheritance.

This overrides superpowers:subagent-driven-development's instruction to select "the least powerful
model that can handle each role". That guidance optimises for the cost of a single dispatch. This
pipeline sends every implementer's output to a review panel, so an implementation defect is not
avoided by cheapness — it is found later by the panel and repaired by a fix wave that re-runs it.
The cheaper implementer buys a more expensive review.

It also overrides two further instructions in the same upstream skill, which
`skills/myflow-contracts/pipeline.md` §Model policy names: dispatch the final review on the most
capable model, and escalate the model in fix rounds 4-5.

Non-normative contrast — the panel's own model is **not** set here. Every review-panel slot runs
on Sonnet, and `myflow-review-panel-economics` is canonical for that; this requirement does not
restate it, weaken it, or depend on it. The two defaults point in opposite directions
deliberately, and the reason is worth stating even though the rule is elsewhere: a reviewer
supplies one of many independent readings of a finished diff, so panel cost must not scale with
whichever model the operator happens to be running, while an implementer produces the diff once,
where capability compounds.

#### Scenario: An implementer is dispatched for a plan task

- **WHEN** `/myflow-do` dispatches an implementer for a task
- **THEN** the dispatch names Opus, or the strongest model the harness offers
- **AND** the model is stated in the dispatch rather than inherited from the parent session

#### Scenario: A mechanical task does not lower the default

- **WHEN** a task's plan text already contains the complete code to write
- **THEN** the implementer still runs on the strongest available model
- **AND** the economy tiering described in subagent-driven-development does not apply

### Requirement: An explicit operator instruction overrides either default

Both defaults SHALL yield to an explicit operator instruction, in either direction — raising the
panel to a stronger model, or lowering an implementer for genuinely mechanical work. The
instruction SHALL be recorded with the dispatch it governs.

An override nobody wrote down is indistinguishable from a mistake by anyone reading the record
later.

#### Scenario: The operator raises the panel

- **WHEN** the operator instructs that the panel run on a stronger model
- **THEN** the slots are dispatched on that model
- **AND** the instruction is recorded alongside the dispatch

#### Scenario: No instruction is given

- **WHEN** no operator instruction addresses the model
- **THEN** the defaults apply — implementers strongest, panel Sonnet

### Requirement: Every dispatch records the model it used, for the session that made it

The SDD ledger SHALL record the model used for each subagent dispatch, implementer and reviewer
alike, alongside the task or slot it ran. Where the dispatcher cannot know the model — a slot
dispatched by `subagent_type`, which carries its own agent definition — the entry SHALL record
`unknown (agent-defined)`. It SHALL NOT record a guess.

A model policy that nothing records is a policy nothing can verify. The absence of this record is
how the implementer rule came to be missing: the panel's model choices were partly written down,
while the implementers' were not recorded anywhere, so no one could tell whether any policy had
been followed.

**Scope, stated plainly: this record is session-scoped and does not survive the change.** The SDD
ledger lives under `.superpowers/`, which is gitignored, in a worktree `/myflow-finish` run 2
removes. It is therefore a record for the operator *during* the change and for the review panel
reading it — not an after-the-fact audit trail, and no requirement here pretends otherwise. This
is the same disclosure task 7 made about its own unverifiable claim, for the same reason: a
capability the artifacts cannot actually provide is worse than an absent one, because a reader
plans around it. Making it durable (copying the ledger into the repository at archive time) was
considered and deliberately deferred — it adds a write to the one command that performs
irreversible operations, and it belongs in the change that owns `/myflow-finish`'s archive step.

The `unknown (agent-defined)` value exists for the same reason. Slots the panel dispatches by
`subagent_type` resolve their model from their own definition, which the dispatcher never reads.
Recording a plausible model for them would put a value in the audit trail that nothing measured —
precisely the failure mode this whole change exists to police.

#### Scenario: A ledger entry is written for a completed task

- **WHEN** an implementer task completes and its ledger line is written
- **THEN** that line names the model the implementer ran on

#### Scenario: A slot's model is not knowable to the dispatcher

- **WHEN** a review-panel slot is dispatched by `subagent_type` and carries its own agent definition
- **THEN** its ledger entry records `unknown (agent-defined)`
- **AND** it does not record a model the dispatcher had no way to observe

#### Scenario: A reader asks which model implemented a given task

- **WHEN** the change is still in flight and its worktree exists
- **THEN** the ledger answers from the record rather than from the transcript
- **AND** once the change is archived the record is gone, which the requirement states rather than
  leaving to be discovered
