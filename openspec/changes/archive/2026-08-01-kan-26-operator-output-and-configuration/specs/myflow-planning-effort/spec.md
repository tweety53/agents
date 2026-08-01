## ADDED Requirements

### Requirement: `/myflow-start` asks for a planning effort level once per change

`/myflow-start` SHALL ask which planning effort to apply on the run that **creates** a change, and
SHALL NOT ask again on any later run for that change.

"Creates" SHALL be determined by the absence of the change's state file, not by any inference about
the operator or the conversation. A revision round — `/myflow-start` re-entered at `STARTED` — SHALL
read the recorded value, state which level it is reusing, and proceed without asking.

The question SHALL be asked interactively, in the same way `/myflow-finish` asks how a branch should
land. Planning effort SHALL NOT become a command argument: the only argument any `/myflow-*` command
accepts remains the optional change name, and anything else is still reported rather than
interpreted.

The concept SHALL be called **planning effort** wherever it is named to an operator, so the question
says what the level governs rather than leaving "effort" to be read as the effort of the whole
change.

#### Scenario: The creating run asks

- **WHEN** `/myflow-start` runs for a change that has no state file
- **THEN** it asks which planning effort level to apply before producing any artifact

#### Scenario: A revision round does not ask again

- **WHEN** `/myflow-start` is re-run for a change already at `STARTED`
- **THEN** it does not ask for a planning effort level
- **AND** it states the level recorded on the creating run and proceeds at it

#### Scenario: Planning effort is never passed as an argument

- **WHEN** an operator invokes `/myflow-start` with a word that is not a known change name
- **THEN** that word is reported as unrecognised
- **AND** it is not interpreted as a planning effort level

### Requirement: Planning effort scales the reasoning spent inside the gates, never the gates themselves

The chosen level SHALL govern only how much reasoning `/myflow-start` spends on its own
brainstorming and plan enrichment. Three levels SHALL exist — `low`, `default` and `detailed` — with
`default` the level offered as the recommendation.

The level names SHALL say what they mean to an operator choosing between them: `default` is what a
change gets unless there is a reason to differ, and `detailed` describes the work done rather than
a position on a scale whose endpoints are not shown.

Every level SHALL run brainstorming, SHALL hold the design approval gate, SHALL run writing-plans,
and SHALL leave `tasks.md` at plan quality rather than a scaffold. A lower level SHALL mean fewer
rounds of questions and coarser grouping, never a gate that does not run.

A level able to switch a gate off would make the field a way to skip review rather than a way to
size the thinking inside it.

This requirement is normative for the levels, the default, and what a level may change. The
operational table the commands read — the same three levels rendered as the per-level behaviour
`/myflow-start` follows — SHALL live in one place, the **Planning effort** section of
`skills/myflow-contracts/state-file.md`, and SHALL follow this requirement rather than restate it as
a second source. A change to the levels or the default SHALL be made here first and carried into that
table; the two SHALL NOT be allowed to disagree.

The level SHALL NOT be read as an instruction by `/myflow-do` or `/myflow-finish`. Those commands
determine their own depth; in particular the review panel's breadth is decided by its own escalation
triggers and SHALL NOT be scaled from this field.

#### Scenario: The lowest level still runs every gate

- **WHEN** a change is planned at `low`
- **THEN** brainstorming runs, the design is presented and approved, writing-plans runs, and
  `tasks.md` carries exact paths and verification commands
- **AND** only the number of question rounds and the grouping differ from a higher level

#### Scenario: The review panel is not scaled by planning effort

- **WHEN** a change recorded at `low` reaches `/myflow-do`
- **THEN** the review panel's roster is decided by its own escalation triggers
- **AND** the recorded level does not remove a slot

#### Scenario: The highest level separates what lower levels group

- **WHEN** a change is planned at `detailed`
- **THEN** open questions are put one at a time and each design section is approved on its own

#### Scenario: The recommended level is the one named default

- **WHEN** the creating run offers the three levels
- **THEN** `default` is the option marked as the recommendation

### Requirement: The chosen planning effort is recorded in the state file, and its absence is legal

The answer SHALL be written to a `planningEffort` field in the change's state file by
`/myflow-start`, and SHALL be carried forward verbatim by every other command, exactly as
`jiraIssue` is.

A state file that **does not carry the key at all** SHALL be read as "not recorded" and SHALL NOT be
treated as unparseable. This is an explicit exception to the closed-schema rule, which otherwise makes
both a missing documented field and an undocumented key unparseable.

Without that exception every state file written before this field existed would be routed through
self-heal, which announces unrecovered fields and rewrites from artifact inference — a loud correction
for a value nobody had the opportunity to set.

A state file carrying the retired `effort` key SHALL be read as recording the equivalent level —
`medium` as `default`, `high` as `detailed`, `low` as `low` — and SHALL NOT be treated as
unparseable. It SHALL be rewritten under `planningEffort` on the next write that file receives, with
no migration pass and no announced correction.

**Every command that reads the level SHALL perform that fallback**, not read `planningEffort`
alone: a compatibility read promised in prose and implemented in no consumer reports a file that
recorded a real level as having recorded none, which is the one outcome the exception exists to
prevent. Where a file carries **both** keys, `planningEffort` SHALL win.

A value under the retired key **outside those three** SHALL be read as *not recorded*, and SHALL NOT
make the file unparseable. Declaring such a file unparseable would promise an announcement no
command emits — `/myflow-do` and `/myflow-finish` invoke self-heal nowhere, and `/myflow-status`
reads the file through a literal `jq` projection that ignores keys it does not name — whereas *not
recorded* needs no detection to be true and discards only a value that already mapped to no level.
`planningEffort` SHALL remain exempt from being named among the unrecovered fields so the
announcement's shape is unchanged.

The field SHALL be a record and the default for a revision round. No command SHALL derive behaviour
from it beyond `/myflow-start`'s own reasoning depth.

#### Scenario: A pre-existing state file is not unparseable

- **WHEN** a command reads a state file written before this field existed
- **THEN** the file parses normally with the level read as not recorded
- **AND** no self-heal correction is announced on that account

#### Scenario: The retired key is read as its equivalent level

- **WHEN** a command reads a state file carrying `effort` set to `medium`
- **THEN** the level is read as `default`, the file is not treated as unparseable, and no correction
  is announced
- **AND** the next write to that file records `default` under `planningEffort` and no `effort` key

#### Scenario: Every consumer performs the fallback

- **WHEN** `/myflow-status` reads a state file that recorded the level only under the retired key
- **THEN** it reads through the fallback rather than `planningEffort` alone
- **AND** it surfaces the mapped level rather than the raw value, and rather than "not recorded"

#### Scenario: A file carrying both keys resolves to the current one

- **WHEN** a command reads a state file carrying both `planningEffort` and `effort`
- **THEN** `planningEffort` governs
- **AND** the next write to that file re-emits it and drops the retired key

#### Scenario: An unmapped retired value reads as not recorded

- **WHEN** a command reads a state file whose retired key holds a value outside the three mapped
  ones
- **THEN** the file parses, the level reads as not recorded, and no correction is announced
- **AND** the raw value is never surfaced as a level

#### Scenario: Later commands carry the value forward

- **WHEN** `/myflow-do` or `/myflow-finish` writes the state file for a change with a recorded level
- **THEN** the recorded value is re-emitted unchanged
- **AND** it is not re-derived, defaulted, or dropped

#### Scenario: A change created without the ask records no level

- **WHEN** a change's state file is written by a path that never asked for a planning effort level
- **THEN** the field records that no level was chosen rather than a fabricated default
