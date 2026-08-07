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

`pipeline.md` SHALL be canonical for the three states, the command→state transition table, git
boundaries, and the finish contract.

A contract carrying both rules and the reasoning behind them SHALL be split into a core and a
rationale appendix, per
**A contract file separates its normative core from its rationale**
(`openspec/specs/myflow-contract-economy/spec.md`). Both halves live in this same directory, and
the appendix is a file of this skill like any other.

#### Scenario: The index and the directory agree

- **WHEN** `skills/myflow-contracts/` is listed and `SKILL.md` is read
- **THEN** every `.md` file in the directory other than `SKILL.md` is named in the index
- **AND** every file the index names is present in the directory

#### Scenario: A consumer needing one contract loads only that contract

- **WHEN** a skill needs the state-file shape and nothing else
- **THEN** it loads `state-file.md` alone
- **AND** it does not load `jira-integration.md`

#### Scenario: A consumer needing a contract's rules loads no appendix

- **WHEN** a `/myflow-*` command needs the pipeline contract
- **THEN** it loads `pipeline.md`
- **AND** it does not load `pipeline-rationale.md`

### Requirement: The contracts skill installs into every harness without new installer code

`setup.sh` SHALL install `skills/myflow-contracts/` into every harness skill directory through the
existing `install_skills` path, with no installer changes beyond what that path already does.

**This covers every file the directory holds, present and future.** `install_skills`
(`setup.sh:188`) iterates `skills/*/` and links each whole skill *directory*, so a file added to
`skills/myflow-contracts/` ships to all three harnesses with no installer edit. A split that adds
core, appendix and single-command files therefore requires no installer change, and that is
asserted here rather than implemented.

#### Scenario: Global install places the contracts skill in all three harnesses

- **WHEN** `setup.sh global` runs against a sandboxed `HOME`
- **THEN** `myflow-contracts` is present under `.claude/skills/`, `.cursor/skills/`, and
  `.codex/skills/`

#### Scenario: The contracts skill is not inlined into the managed block

- **WHEN** `setup.sh global` runs against a sandboxed `HOME`
- **THEN** the managed block in `.claude/CLAUDE.md` does not contain the contract bodies
- **AND** that block is at least 25 KB smaller than before the change that extracted them

#### Scenario: Files added to the contracts skill ship without an installer edit

- **WHEN** `setup.sh global` runs against a sandboxed `HOME` after files are added to
  `skills/myflow-contracts/`
- **THEN** every `.md` file in that directory is reachable under `.claude/skills/myflow-contracts/`,
  `.cursor/skills/myflow-contracts/` and `.codex/skills/myflow-contracts/`
- **AND** `setup.sh` carries no per-file list of contract names
