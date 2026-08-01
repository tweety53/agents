## ADDED Requirements

### Requirement: The pipeline diagram and its stage table live in `pipeline.md`, and nowhere else

`skills/myflow-contracts/pipeline.md` SHALL carry the pipeline's state diagram and, beneath it, a
two-level stage table. No other file in this repository SHALL carry a second copy of either.
`README.md` SHALL point at `pipeline.md` rather than reproduce them.

**Level 1** SHALL be one row per command, listing its stages in order with the human gates marked,
and SHALL mark which stages carry a level-2 expansion. **Level 2** SHALL expand the stages that
themselves hide substructure. Those stages SHALL be:

| Command | Stage |
|---------|-------|
| `/myflow-start` | brainstorm |
| `/myflow-start` | writing-plans |
| `/myflow-do` | SDD + TDD per task |
| `/myflow-do` | the review panel |
| `/myflow-finish` | the preflight verdict |
| `/myflow-finish` run 1 | the unfinished-work gate |
| `/myflow-finish` run 1 | the landing routes |
| `/myflow-finish` run 2 | cleanup |

A level-2 expansion SHALL state **structure** and SHALL cite the owning file for **tuned
thresholds**. Structure is the shape that changes only when the pipeline changes: which review slots
are required and which conditional, that every slot runs on the panel's model — Sonnet by default —
except those dispatched by `subagent_type`, that no handoff occurs while any finding is open at any
severity, that escalation widens the panel's breadth rather than its model, that a `REFUSE` verdict
stops the run. Thresholds are the tuned values that move independently: the changed-line counts that
select a conditional slot, the per-slot trigger lists, and the conditions that force a full re-run.

An expansion SHALL NOT reproduce a table another file owns. A copied threshold is a copy that goes
wrong silently, which is the same failure that justifies the README carrying no diagram; accepting
it one level down would make the rule inconsistent with itself.

Every citation SHALL use the named-section form `scripts/check-references.sh` verifies, so a cited
section that is renamed or removed fails the guard rather than rotting.

Placing the diagram here is what lets `/myflow-info` show it: that command reads `pipeline.md` at
invocation time and is forbidden from answering from memory, so a diagram held only in `README.md`
is one it can never present.

#### Scenario: The contract file carries both levels

- **WHEN** `pipeline.md` is read
- **THEN** it carries the state diagram, a level-1 row per command, and a level-2 expansion for each
  of the eight stages named above

#### Scenario: The README carries no copy

- **WHEN** `README.md` is read
- **THEN** it points at `pipeline.md` for the diagram and the stage table
- **AND** it contains neither

#### Scenario: A tuned threshold is cited, not copied

- **WHEN** the review panel's level-2 expansion describes how a conditional slot is selected
- **THEN** it states that the slot is conditional and cites the section that lists its triggers
- **AND** it does not restate the changed-line counts

#### Scenario: A renamed cited section fails the guard

- **WHEN** a section a level-2 expansion cites is renamed
- **THEN** `scripts/check-references.sh` reports the citation as unresolved

#### Scenario: The pipeline reference command can show the diagram

- **WHEN** `/myflow-info` is asked how the pipeline works
- **THEN** the diagram it presents comes from the `pipeline.md` it read during that invocation
