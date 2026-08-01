## ADDED Requirements

### Requirement: The creating run asks and records three model choices

`/myflow-start` SHALL ask three questions on the run that **creates** a change — one per model role —
and SHALL record the answers in the change's state file. Each SHALL be asked as its own question,
with the default named and marked as the recommendation.

| Role | Default |
|------|---------|
| implementation — the implementer subagents `/myflow-do` dispatches | Opus, or the harness's strongest available model |
| review panel — every slot that takes a model override | Sonnet |
| panel fixes — the subagents that repair panel findings | Opus, or the harness's strongest available model |

"Creates" SHALL be determined by the absence of the state file, exactly as the planning-effort
question determines it. A revision round SHALL state the recorded values and SHALL NOT ask again.
Every other command SHALL carry the recorded values forward verbatim, as it does the linked Jira
issue.

**The panel-fix default is the strongest available model, and SHALL NOT be lowered by this
capability.** The role's name describes the agent that *applies* a fix, which is an implementer, and
the requirement that implementer subagents run on the strongest available model already governs it.
Fix rounds escalate the panel's breadth rather than its model, because implementers sit at the
ceiling from the first round; a fix-wave default of Sonnet would contradict both.

**These fields record intent; the SDD ledger records what happened.** This capability already
requires that an override be recorded, because an override nobody wrote down is indistinguishable
from a mistake — these fields are that writing-down, made durable rather than transcript-only. They
SHALL NOT replace the per-dispatch ledger line, which remains the only evidence of the model a
dispatch actually ran on.

**Slots dispatched by `subagent_type` are unaffected.** They carry their own agent definitions, take
no model override from any mechanism, and SHALL continue to record `unknown (agent-defined)`. A
recorded panel model SHALL NOT be passed to them, and SHALL NOT cause a model to be written for them
in the ledger.

#### Scenario: The creating run asks three questions

- **WHEN** `/myflow-start` runs for a change that has no state file
- **THEN** it asks separately which model to use for implementation, for the review panel, and for
  panel fixes
- **AND** each question names its default and marks it as the recommendation

#### Scenario: A revision round does not ask again

- **WHEN** `/myflow-start` is re-run for a change already at `STARTED`
- **THEN** it states the recorded model choices and does not ask for them

#### Scenario: A dispatch uses the recorded model

- **WHEN** `/myflow-do` dispatches an implementer for a change whose state file records an
  implementation model
- **THEN** that model is named on the dispatch
- **AND** the ledger line records the model actually used

#### Scenario: The recorded value does not replace the ledger

- **WHEN** a reader asks which model implemented a given task
- **THEN** the answer comes from the ledger entry for that dispatch
- **AND** not from the state file, which records what was chosen rather than what ran

#### Scenario: A subagent_type slot takes no recorded override

- **WHEN** the panel dispatches Bugbot or Security Review for a change recording a panel model
- **THEN** no model override is passed to that slot
- **AND** its ledger entry still records `unknown (agent-defined)`

#### Scenario: The panel-fix default is not lowered

- **WHEN** no operator choice addresses the panel-fix role
- **THEN** fix-wave subagents run on the strongest available model

## MODIFIED Requirements

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

Non-normative contrast — the panel's own model is **not** set here. Every review-panel slot runs on
the panel's model, **Sonnet by default**, and `myflow-review-panel-economics` is canonical for that;
this requirement does not restate it, weaken it, or depend on it. The default is named as a default
rather than an absolute because a change may record its own panel model, which this capability makes
possible — an unqualified "runs on Sonnet" here would contradict the requirement below the moment
one is recorded. The two defaults point in opposite directions deliberately, and the reason is worth
stating even though the rule is elsewhere: a reviewer supplies one of many independent readings of a
finished diff, so panel cost must not scale with whichever model the operator happens to be running,
while an implementer produces the diff once, where capability compounds.

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

An instruction MAY be given in either of two ways, and both SHALL be honoured: as a model choice
recorded in the change's state file by `/myflow-start`'s creating run, or as an instruction given in
the session. A recorded choice SHALL apply to every run of the change without being restated, which
is the point of recording it. A session instruction SHALL take precedence over a recorded choice for
the run in which it is given, and SHALL be recorded with the dispatch exactly as before — it is the
narrower and later of the two.

#### Scenario: The operator raises the panel

- **WHEN** the operator instructs that the panel run on a stronger model
- **THEN** the slots are dispatched on that model
- **AND** the instruction is recorded alongside the dispatch

#### Scenario: No instruction is given

- **WHEN** no operator instruction addresses the model
- **THEN** the defaults apply — implementers strongest, panel Sonnet

#### Scenario: A recorded choice applies without being restated

- **WHEN** `/myflow-do` runs for a change whose state file records a panel model
- **THEN** the slots that take an override are dispatched on that model
- **AND** the operator does not have to repeat the instruction

#### Scenario: A session instruction beats the recorded choice

- **WHEN** the operator instructs a different model during a run of a change that records one
- **THEN** the session instruction governs that run
- **AND** it is recorded with the dispatch
