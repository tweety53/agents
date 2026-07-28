## MODIFIED Requirements

### Requirement: Always-on rule layer carries only always-on judgment

`rules/myflow-manual-review.mdc` SHALL contain only what an agent needs in order to recognise
that myflow governs the work and to know where the rest lives: the state diagram, the
instruction to load `skills/myflow-contracts/pipeline.md` before acting, and the table of
narrower contracts.

The pipeline itself — the state definitions, the command→state transition table, the mismatch
handoff, git boundaries per state, and the finish contract — SHALL NOT appear in that file, nor
SHALL the four contract sections (State file, State self-heal, Project configuration, Jira
integration).

The rule's frontmatter `description` SHALL name the three states, not the retired twelve stages.

#### Scenario: Contract and pipeline bodies are absent from the always-on rule

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** it contains no `jq` state-write template, no `## standards` entry-form table, no
  containment rules, no Jira transition table, and no command→state transition table
- **AND** its total size is at most 8 KB

#### Scenario: The trigger and the pointers are retained

- **WHEN** `rules/myflow-manual-review.mdc` is read after this change
- **THEN** the three-state diagram is present, the instruction to load `pipeline.md` first is
  present, and each of the four narrower contracts is named with its exact path

#### Scenario: The frontmatter names the current vocabulary

- **WHEN** the rule's frontmatter `description` is read
- **THEN** it names `STARTED`, `IN_PROGRESS` and `FINISHED`, and no retired
  stage value appears in it

### Requirement: Extracted contracts live in a dedicated on-demand skill

The extracted content SHALL live in `skills/myflow-contracts/`, one file per contract:
`pipeline.md`, `state-file.md`, `state-self-heal.md`, `project-configuration.md`, and
`jira-integration.md`. `skills/myflow-contracts/SKILL.md` SHALL be an index that names each
file and states when it is needed, and SHALL NOT restate the contract bodies.

`pipeline.md` SHALL be canonical for the three states, the command→state transition table, git
boundaries, and the finish contract.

#### Scenario: One file per contract

- **WHEN** `skills/myflow-contracts/` is listed
- **THEN** it contains exactly `SKILL.md`, `pipeline.md`, `state-file.md`,
  `state-self-heal.md`, `project-configuration.md`, and `jira-integration.md`

#### Scenario: A consumer needing one contract loads only that contract

- **WHEN** a skill needs the state-file shape and nothing else
- **THEN** it loads `state-file.md` alone
- **AND** it does not load `jira-integration.md`

### Requirement: Each extracted section leaves a discoverable stub

At the location each moved section previously occupied, `rules/myflow-manual-review.mdc` SHALL
retain a `##` heading, one sentence naming what the contract governs, and the exact path of the
contracts file that now holds it.

#### Scenario: Stub names the contract and its file

- **WHEN** an agent reads the always-on rule and reaches the Jira integration pointer
- **THEN** it finds a sentence stating that the contract governs issue resolution, transitions,
  and description sync
- **AND** it finds the path `skills/myflow-contracts/jira-integration.md`

#### Scenario: The pipeline pointer is unmissable

- **WHEN** the always-on rule is read
- **THEN** it states that `skills/myflow-contracts/pipeline.md` must be loaded before any other
  step of any `/myflow-*` command, and that acting on a remembered copy is the failure the
  split exists to prevent

### Requirement: The contracts skill is listed where skills are discovered

`myflow-contracts` SHALL appear in the skill index of `CLAUDE.md`, `AGENTS.md`, and
`README.md`, described as the on-demand contract definitions. The retired
`myflow-state-advance` skill SHALL NOT appear alongside it in any of those indexes.

#### Scenario: The index names the surviving skills only

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md` or `README.md` is read
- **THEN** `myflow-contracts` is listed and `myflow-state-advance` is absent
