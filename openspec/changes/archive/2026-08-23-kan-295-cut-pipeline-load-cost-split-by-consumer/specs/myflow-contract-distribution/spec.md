# myflow-contract-distribution delta — kan-295-cut-pipeline-load-cost-split-by-consumer

## MODIFIED Requirements

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

`pipeline.md` SHALL be canonical for the three states and the command→state transition table.
**Git boundaries SHALL live in `git-boundaries.md`**, which is canonical for it: the git actions
each command may take, by state and condition, and the guarded two-commit chain are reachable only
from `/myflow-do`, `/myflow-finish` and `/myflow-fast`, and `pipeline.md` is read by every command,
including `/myflow-start` and `/myflow-status`, neither of which commits. **The finish contract
SHALL live in `finish-contract.md`**, which is canonical for it: the preflight signals, both runs'
procedures, base-branch resolution and worktree cleanup are reachable only from `/myflow-finish`,
and `pipeline.md` is read by every command.

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

#### Scenario: Git boundaries is indexed and attributed to its consuming commands

- **WHEN** the index is read
- **THEN** `git-boundaries.md` is named
- **AND** it is stated to be loaded by `/myflow-do`, `/myflow-finish` and `/myflow-fast`, and not
  by `/myflow-start` or `/myflow-status`

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

#### Scenario: A start or status run does not load git boundaries

- **WHEN** `/myflow-start` or `/myflow-status` runs
- **THEN** it does not load `git-boundaries.md`, since neither command commits

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
