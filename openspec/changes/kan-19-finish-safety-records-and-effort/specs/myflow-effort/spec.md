## ADDED Requirements

### Requirement: `/myflow-start` asks for an effort level once per change

`/myflow-start` SHALL ask which effort to apply on the run that **creates** a change, and SHALL NOT
ask again on any later run for that change.

"Creates" SHALL be determined by the absence of the change's state file, not by any inference about
the operator or the conversation. A revision round — `/myflow-start` re-entered at `STARTED` — SHALL
read the recorded value, state which level it is reusing, and proceed without asking.

The question SHALL be asked interactively, in the same way `/myflow-finish` asks how a branch should
land. Effort SHALL NOT become a command argument: the only argument any `/myflow-*` command accepts
remains the optional change name, and anything else is still reported rather than interpreted.

#### Scenario: The creating run asks

- **WHEN** `/myflow-start` runs for a change that has no state file
- **THEN** it asks which effort level to apply before producing any artifact

#### Scenario: A revision round does not ask again

- **WHEN** `/myflow-start` is re-run for a change already at `STARTED`
- **THEN** it does not ask for an effort level
- **AND** it states the level recorded on the creating run and proceeds at it

#### Scenario: Effort is never passed as an argument

- **WHEN** an operator invokes `/myflow-start` with a word that is not a known change name
- **THEN** that word is reported as unrecognised
- **AND** it is not interpreted as an effort level

### Requirement: Effort scales the reasoning spent inside the gates, never the gates themselves

The chosen level SHALL govern only how much reasoning `/myflow-start` spends on its own
brainstorming and plan enrichment. Three levels SHALL exist — `low`, `medium` and `high` — with
`medium` the default offered.

Every level SHALL run brainstorming, SHALL hold the design approval gate, SHALL run writing-plans,
and SHALL leave `tasks.md` at plan quality rather than a scaffold. A lower level SHALL mean fewer
rounds of questions and coarser grouping, never a gate that does not run.

An effort level able to switch a gate off would make the field a way to skip review rather than a way
to size the thinking inside it.

This requirement is normative for the levels, the default, and what a level may change. The
operational table the commands read — the same three levels rendered as the per-level behaviour
`/myflow-start` follows — SHALL live in one place, the **Effort** section of
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
- **AND** the recorded effort level does not remove a slot

#### Scenario: The highest level separates what lower levels group

- **WHEN** a change is planned at `high`
- **THEN** open questions are put one at a time and each design section is approved on its own

### Requirement: The chosen effort is recorded in the state file, and its absence is legal

The answer SHALL be written to an `effort` field in the change's state file by `/myflow-start`, and
SHALL be carried forward verbatim by every other command, exactly as `jiraIssue` is.

A state file that **does not carry the key at all** SHALL be read as "not recorded" and SHALL NOT be
treated as unparseable. This is an explicit exception to the closed-schema rule, which otherwise makes
both a missing documented field and an undocumented key unparseable.

Without that exception every state file written before this field existed would be routed through
self-heal, which announces unrecovered fields and rewrites from artifact inference — a loud correction
for a value nobody had the opportunity to set.

The field SHALL be a record and the default for a revision round. No command SHALL derive behaviour
from it beyond `/myflow-start`'s own reasoning depth.

#### Scenario: A pre-existing state file is not unparseable

- **WHEN** a command reads a state file written before the `effort` field existed
- **THEN** the file parses normally with `effort` read as not recorded
- **AND** no self-heal correction is announced on that account

#### Scenario: Later commands carry the value forward

- **WHEN** `/myflow-do` or `/myflow-finish` writes the state file for a change with a recorded effort
- **THEN** the recorded value is re-emitted unchanged
- **AND** it is not re-derived, defaulted, or dropped

#### Scenario: A change created without the ask records no level

- **WHEN** a change's state file is written by a path that never asked for an effort level
- **THEN** the field records that no level was chosen rather than a fabricated default
