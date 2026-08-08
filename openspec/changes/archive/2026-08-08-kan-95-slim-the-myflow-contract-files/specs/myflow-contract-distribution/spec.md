## MODIFIED Requirements

### Requirement: The pipeline diagram and its stage table live in `pipeline.md`, and nowhere else

`README.md` SHALL carry the pipeline's state diagram and, beneath it, a two-level stage table, as
prose written for a human reader rather than for a command. `skills/myflow-contracts/pipeline.md`
SHALL NOT carry either, and no other file in this repository SHALL carry a second copy of either.

**The move is what makes the stage tables free.** No `/myflow-*` command reads `README.md`, so the
diagram and both levels cost nothing per run. They previously sat in `pipeline.md` so that
`/myflow-info` could present them at invocation time; that command no longer exists, and the tables
therefore have no runtime consumer at all.

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
wrong silently, and accepting it one level down would make the rule inconsistent with itself.

Every citation SHALL use the named-section form `scripts/check-references.sh` verifies, so a cited
section that is renamed or removed fails the guard rather than rotting.

**These passages are rewritten for a human reader rather than moved verbatim, and that is
deliberate.** They leave the loaded corpus entirely for a document no command reads, which
**A mixed passage MAY be handled by rule extraction** (`myflow-contract-economy`) places outside the
verbatim-partition rule rather than in exception to it. Each rewritten passage SHALL still carry a
per-move ledger row naming `README.md` as its destination.

#### Scenario: The README carries both levels

- **WHEN** `README.md` is read
- **THEN** it carries the state diagram, a level-1 row per command, and a level-2 expansion for each
  of the eight stages named above

#### Scenario: The contract file carries no copy

- **WHEN** `skills/myflow-contracts/pipeline.md` is read
- **THEN** it contains neither the state diagram nor either level of the stage table

#### Scenario: A tuned threshold is cited, not copied

- **WHEN** the review panel's level-2 expansion describes how a conditional slot is selected
- **THEN** it states that the slot is conditional and cites the section that lists its triggers
- **AND** it does not restate the changed-line counts

#### Scenario: A renamed cited section fails the guard

- **WHEN** a section a level-2 expansion cites is renamed
- **THEN** `scripts/check-references.sh` reports the citation as unresolved

#### Scenario: No command loads the stage tables

- **WHEN** any `/myflow-*` command runs
- **THEN** it loads no file carrying the state diagram or either level of the stage table

### Requirement: Extracted contracts live in a dedicated on-demand skill

The extracted content SHALL live in `skills/myflow-contracts/`, one file per contract.
`skills/myflow-contracts/SKILL.md` SHALL be an index that names each file and states when it is
needed, and SHALL NOT restate the contract bodies.

**The index is the inventory, and this requirement SHALL NOT enumerate the directory's contents.**
An enumeration here went stale the moment a contract was added and stayed stale through three
changes that each added one — `build-green.md`, `plan-provenance.md` and `workspace-isolation.md`
were all absent from the list that claimed to be exhaustive. What is required instead is that the
index and the directory agree in both directions: every file present is named in `SKILL.md`, and
every file `SKILL.md` names is present.

`pipeline.md` SHALL be canonical for the three states, the command→state transition table and git
boundaries. **The finish contract SHALL live in `finish-contract.md`**, which is canonical for it:
the preflight signals, both runs' procedures, base-branch resolution and worktree cleanup are
reachable only from `/myflow-finish`, and `pipeline.md` is read by every command.

**The per-state handoff block templates SHALL live in `handoff-blocks.md`**, which is canonical for
them: only `/myflow-status` needs the full set, every producing command carrying just the one block
it prints, and `pipeline.md` is read by every command.

**`state-self-heal.md` SHALL NOT exist.** State self-heal is removed from the pipeline, and a
contract file that no command loads and whose mechanism nothing performs is not kept against a
future that may never arrive.

A contract carrying both rules and the reasoning behind them SHALL be split into a core and a
rationale appendix, per
**A contract file separates its normative core from its rationale**
(`openspec/specs/myflow-contract-economy/spec.md`). Both halves live in this same directory, and
the appendix is a file of this skill like any other.

**A rationale appendix belonging to a skill lives beside that skill's `SKILL.md`, not here.** This
directory holds contracts; a skill's reasoning is not one, and collecting them here would make this
directory a catch-all whose index and budget guard then have to police files that belong elsewhere.

#### Scenario: The index and the directory agree

- **WHEN** `skills/myflow-contracts/` is listed and `SKILL.md` is read
- **THEN** every `.md` file in the directory other than `SKILL.md` is named in the index
- **AND** every file the index names is present in the directory

#### Scenario: The finish contract is indexed and attributed to one command

- **WHEN** the index is read
- **THEN** `finish-contract.md` is named
- **AND** it is stated to be loaded by `/myflow-finish` and no other command

#### Scenario: The handoff blocks are indexed and attributed to one command

- **WHEN** the index is read
- **THEN** `handoff-blocks.md` is named
- **AND** it is stated to be loaded by `/myflow-status` and no other command

#### Scenario: A consumer needing one contract loads only that contract

- **WHEN** a skill needs the state-file shape and nothing else
- **THEN** it loads `state-file.md` alone
- **AND** it does not load `jira-integration.md`

#### Scenario: A consumer needing a contract's rules loads no appendix

- **WHEN** a `/myflow-*` command needs the pipeline contract
- **THEN** it loads `pipeline.md`
- **AND** it does not load `pipeline-rationale.md`

#### Scenario: A start or do run does not load the finish contract

- **WHEN** `/myflow-start` or `/myflow-do` runs
- **THEN** it loads `pipeline.md`
- **AND** it does not load `finish-contract.md`

#### Scenario: A producing command does not load the handoff blocks

- **WHEN** `/myflow-start`, `/myflow-do` or `/myflow-finish` runs
- **THEN** it does not load `handoff-blocks.md`

#### Scenario: The self-heal contract is gone

- **WHEN** `skills/myflow-contracts/` is listed
- **THEN** no `state-self-heal.md` exists
- **AND** no file in the repository cites it

#### Scenario: A skill's appendix is not filed under the contracts skill

- **WHEN** `skills/myflow-contracts/` is listed
- **THEN** no `SKILL-rationale.md` belonging to another skill is present

### Requirement: Always-on rule layer carries only the trigger and the pointers

`rules/myflow-manual-review.mdc` SHALL contain only what an agent needs in order to recognise
that myflow governs the work and to know where the rest lives: the state diagram, the
instruction to load `skills/myflow-contracts/pipeline.md` before acting, and the table of
narrower contracts.

The pipeline itself — the state definitions, the command→state transition table, the mismatch
handoff, git boundaries per state, and the finish contract — SHALL NOT appear in that file, nor
SHALL the three contract sections (State file, Project configuration, Jira integration).

**There are three rather than four because state self-heal is removed from the pipeline.** The
mechanism had one performer, `/myflow-status`, which no longer validates a state file against
artifacts, so neither the contract nor a pointer to it survives anywhere — including here.

The rule's frontmatter `description` SHALL name the three states, not the retired twelve stages.

#### Scenario: Contract and pipeline bodies are absent from the always-on rule

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** it contains no `jq` state-write template, no `## standards` entry-form table, no
  containment rules, no Jira transition table, and no command→state transition table
- **AND** its total size is at most 8 KB

#### Scenario: The trigger and the pointers are retained

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** the three-state diagram is present, the instruction to load `pipeline.md` first is
  present, and each of the three narrower contracts is named with its exact path

#### Scenario: The removed contract is not pointed at

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** it names no state self-heal contract and carries no row for one

#### Scenario: The frontmatter names the current vocabulary

- **WHEN** the rule's frontmatter `description` is read
- **THEN** it names `STARTED`, `IN_PROGRESS` and `FINISHED`, and no retired
  stage value appears in it

### Requirement: The contracts skill is listed where skills are discovered

`myflow-contracts` SHALL appear in the skill index of `CLAUDE.md`, `AGENTS.md`, and
`README.md`, described as the on-demand contract definitions. The retired
`myflow-state-advance` skill SHALL NOT appear alongside it in any of those indexes, and neither
SHALL the removed `myflow-info` skill.

#### Scenario: The skill index names the contracts skill

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md`, or `README.md` is read
- **THEN** `skills/myflow-contracts/` is listed with its purpose

#### Scenario: The index names the surviving skills only

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md` or `README.md` is read
- **THEN** `myflow-contracts` is listed
- **AND** `myflow-state-advance` and `myflow-info` are both absent
