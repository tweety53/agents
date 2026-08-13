# myflow-debugging-dispatch Specification

## Purpose
TBD - created by archiving change kan-158-wire-systematic-debugging-into-myflow-do. Update Purpose after archive.
## Requirements
### Requirement: The workflow table names systematic-debugging

`skills/myflow-do/SKILL.md`'s Superpowers Basic Workflow table SHALL carry a row naming
`superpowers:systematic-debugging` at step 7, stating that it applies to an unexpected test failure
during implementation (one RED-GREEN-REFACTOR did not plan) and to a review-panel finding confirmed
as a real defect.

#### Scenario: The table names the skill and its trigger

- **WHEN** the Superpowers Basic Workflow table in `skills/myflow-do/SKILL.md` is read
- **THEN** it carries a row for `superpowers:systematic-debugging` naming both triggers: an
  unexpected test failure, and a confirmed-defect review finding

### Requirement: The implementer dispatch names the skill for unexpected failures

Every implementer dispatch in section 4 SHALL carry a `REQUIRED SUB-SKILL` instruction naming
`superpowers:systematic-debugging`, to be invoked when a test fails for a reason
RED-GREEN-REFACTOR did not plan, before the implementer writes a fix.

This requirement SHALL NOT change the commit-per-task model, the per-task review shape, or require
invoking the skill on an ordinary expected RED step.

#### Scenario: An implementer dispatch carries the instruction

- **WHEN** the parent dispatches an implementer for a task in section 4
- **THEN** the dispatch prompt names `superpowers:systematic-debugging` as the required approach for
  a test failure RED-GREEN-REFACTOR did not plan

#### Scenario: An ordinary expected RED is unaffected

- **WHEN** an implementer's test fails exactly as RED-GREEN-REFACTOR's own RED step expects
- **THEN** the implementer proceeds with GREEN as before, without invoking
  `superpowers:systematic-debugging`

### Requirement: The fix-subagent dispatch names the skill for confirmed defects

The section 5 fix-subagent dispatch, which receives the union of surviving review findings, SHALL
carry an instruction naming `superpowers:systematic-debugging` for any finding confirmed as a real
defect, before the fix subagent writes a fix for it.

This requirement SHALL NOT change the review panel's roster, escalation ladder, fix-round
accounting, or the zero-open-findings handoff bar.

#### Scenario: The fix-subagent dispatch carries the instruction

- **WHEN** the parent dispatches the fix subagent with the surviving findings in section 5
- **THEN** the dispatch prompt names `superpowers:systematic-debugging` as the required approach for
  a finding confirmed as a real defect
