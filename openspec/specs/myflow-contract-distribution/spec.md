# myflow-contract-distribution Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: Always-on rule layer carries only always-on judgment

`rules/myflow-manual-review.mdc` SHALL contain only the sections that govern judgment an agent
must exercise without being asked to load anything: Model policy, Change name resolution, Pipeline
stages, Stage transitions, Fix re-entry, IntelliJ commands, Stage boundaries, the per-stage
sections from Do through Finish, Full cycle gates, and Opt-out.

The four contract sections — State file, State self-heal, Project configuration, and Jira
integration — SHALL NOT appear in full in that file.

#### Scenario: Contract bodies are absent from the always-on rule

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** it contains no `jq` state-write template, no `## standards` entry-form table, no
  containment rules, and no Jira transition table
- **AND** its total size is at most 32 KB

#### Scenario: Judgment sections are retained

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** the Pipeline stages table, the Stage transitions table, and the Stage boundaries table
  are present in full and unchanged in meaning

### Requirement: Extracted contracts live in a dedicated on-demand skill

The four extracted sections SHALL live in `skills/myflow-contracts/`, one file per section:
`state-file.md`, `state-self-heal.md`, `project-configuration.md`, and `jira-integration.md`.
`skills/myflow-contracts/SKILL.md` SHALL be an index that names each file and states when it is
needed, and SHALL NOT restate the contract bodies.

#### Scenario: One file per contract section

- **WHEN** `skills/myflow-contracts/` is listed
- **THEN** it contains exactly `SKILL.md`, `state-file.md`, `state-self-heal.md`,
  `project-configuration.md`, and `jira-integration.md`

#### Scenario: A consumer needing one contract loads only that contract

- **WHEN** a skill needs the state-file shape and nothing else
- **THEN** it loads `state-file.md` alone
- **AND** it does not load `jira-integration.md`

### Requirement: Each extracted section leaves a discoverable stub

At the location each moved section previously occupied, `rules/myflow-manual-review.mdc` SHALL
retain the original `##` heading, one sentence naming what the contract governs, and the exact
path of the contracts file that now holds it.

#### Scenario: Stub names the contract and its file

- **WHEN** an agent reads the always-on rule and reaches the `## Jira integration` heading
- **THEN** it finds a sentence stating that the contract governs issue resolution, transitions,
  and description sync
- **AND** it finds the path `skills/myflow-contracts/jira-integration.md`

#### Scenario: All four stubs are present

- **WHEN** `rules/myflow-manual-review.mdc` is searched for the four moved section headings
- **THEN** all four headings are still present, each followed by a stub naming its contracts file

### Requirement: Canonical authority moves with the contract text

Where a moved section previously declared itself canonical over the skills that reference it, that
declaration SHALL move into the contracts file and be reworded to name the contracts file as
canonical. The stub SHALL NOT claim authority over text it no longer contains.

#### Scenario: Contracts file asserts canonical authority

- **WHEN** `skills/myflow-contracts/project-configuration.md` is read
- **THEN** it states that it is the canonical definition of `.myflow/project.md` and that it wins
  over any skill that disagrees

#### Scenario: Stub does not claim canonical authority

- **WHEN** the `## Project configuration` stub in `rules/myflow-manual-review.mdc` is read
- **THEN** it points at the contracts file rather than declaring itself canonical

### Requirement: The contracts skill is listed where skills are discovered

`myflow-contracts` SHALL appear in the skill index of `CLAUDE.md`, `AGENTS.md`, and `README.md`,
alongside `myflow-state-advance`, described as the on-demand contract definitions. A contract an
agent never learns exists is a contract it never loads, which is the failure this split's own
mitigation depends on preventing.

#### Scenario: The skill index names the contracts skill

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md`, or `README.md` is read
- **THEN** `skills/myflow-contracts/` is listed with its purpose

### Requirement: A contract referenced from a subagent prompt resolves by absolute path

Where a skill instructs a subagent to read a contract file, it SHALL give the **absolute**
installed path — not a repo-relative `skills/…` path, which does not open from a project
worktree — and SHALL instruct the agent to **stop and report** rather than proceed if the file
cannot be read.

#### Scenario: The standards containment rule is fail-closed

- **WHEN** `skills/openspec-apply-superpowers/SKILL.md` instructs the resolution of
  `## standards` entries
- **THEN** it names the absolute installed location of `project-configuration.md`
- **AND** it instructs the agent to stop rather than resolve entries without the containment rule

### Requirement: The contracts skill installs into every harness without new installer code

`setup.sh` SHALL install `skills/myflow-contracts/` into every harness skill directory through the
existing `install_skills` path, with no installer changes beyond what that path already does.

#### Scenario: Global install places the contracts skill in all three harnesses

- **WHEN** `setup.sh global` runs against a sandboxed `HOME`
- **THEN** `myflow-contracts` is present under `.claude/skills/`, `.cursor/skills/`, and
  `.codex/skills/`

#### Scenario: The contracts skill is not inlined into the managed block

- **WHEN** `setup.sh global` runs against a sandboxed `HOME`
- **THEN** the managed block in `.claude/CLAUDE.md` does not contain the contract bodies
- **AND** that block is at least 25 KB smaller than before this change

