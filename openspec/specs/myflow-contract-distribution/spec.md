# myflow-contract-distribution Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
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
reachable only from `/myflow-finish`, and `pipeline.md` is read by every command. `pipeline.md`
SHALL retain the pipeline's stage table and every level-2 expansion, including those of the finish
stages, so `/myflow-info` can still describe both runs from the core alone.

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

#### Scenario: A skill's appendix is not filed under the contracts skill

- **WHEN** `skills/myflow-contracts/` is listed
- **THEN** no `SKILL-rationale.md` belonging to another skill is present

### Requirement: Each extracted section leaves a discoverable stub

At the location each moved section previously occupied, `rules/myflow-manual-review.mdc` SHALL
retain a `##` heading, one sentence naming what the contract governs, and the exact path of the
contracts file that now holds it.

#### Scenario: Stub names the contract and its file

- **WHEN** an agent reads the always-on rule and reaches the Jira integration pointer
- **THEN** it finds a sentence stating that the contract governs issue resolution, transitions,
  and description sync
- **AND** it finds the path `skills/myflow-contracts/jira-integration.md`

#### Scenario: All four stubs are present

- **WHEN** `rules/myflow-manual-review.mdc` is searched for the four moved contract section
  headings
- **THEN** all four headings are still present, each followed by a stub naming its contracts file

#### Scenario: The pipeline pointer is unmissable

- **WHEN** the always-on rule is read
- **THEN** it states that `skills/myflow-contracts/pipeline.md` must be loaded before any other
  step of any `/myflow-*` command, and that acting on a remembered copy is the failure the
  split exists to prevent

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

`myflow-contracts` SHALL appear in the skill index of `CLAUDE.md`, `AGENTS.md`, and
`README.md`, described as the on-demand contract definitions. The retired
`myflow-state-advance` skill SHALL NOT appear alongside it in any of those indexes.

#### Scenario: The skill index names the contracts skill

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md`, or `README.md` is read
- **THEN** `skills/myflow-contracts/` is listed with its purpose

#### Scenario: The index names the surviving skills only

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md` or `README.md` is read
- **THEN** `myflow-contracts` is listed and `myflow-state-advance` is absent

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

### Requirement: Always-on rule layer carries only the trigger and the pointers

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
