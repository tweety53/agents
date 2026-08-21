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
`myflow-state-advance` skill SHALL NOT appear alongside it in any of those indexes, and neither
SHALL the removed `myflow-info` skill.

#### Scenario: The skill index names the contracts skill

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md`, or `README.md` is read
- **THEN** `skills/myflow-contracts/` is listed with its purpose

#### Scenario: The index names the surviving skills only

- **WHEN** the skill index in `CLAUDE.md`, `AGENTS.md` or `README.md` is read
- **THEN** `myflow-contracts` is listed
- **AND** `myflow-state-advance` and `myflow-info` are both absent

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

### Requirement: A guard a command invokes SHALL be reachable from that command's installed skill

Every guard script a `/myflow-*` command invokes SHALL be reachable from the skill directory
that command's own skill occupies, at `<skill-dir>/scripts/<name>`, in every harness the
installer targets. A guard reachable only from this repository's own checkout is not
installed, because the project a command runs against is almost never this repository.

The one real copy of each guard SHALL remain at the repository's root `scripts/` directory.
A skill's `scripts/` directory SHALL hold relative symlinks into it, tracked in version
control, so that the installer's existing whole-directory symlink carries them with no
installer step of its own and no second copy exists to drift.

A guard that resolves a sibling from its own directory — a shared library, a `.py` companion,
another guard — SHALL have that sibling symlinked beside it in every skill directory carrying
it. A guard shipped without its siblings fails at the moment it is needed, which is the
failure this requirement exists to prevent.

#### Scenario: A guard is invoked through an installed skill directory

- **WHEN** a command invokes a guard at `<skill-dir>/scripts/<name>` in an installed harness,
  where `<skill-dir>` is itself a symlink into this repository
- **THEN** the guard executes
- **AND** any sibling it resolves from its own directory resolves too

#### Scenario: A skill invokes a guard it does not carry

- **WHEN** a skill's text invokes a guard for which that skill's `scripts/` directory holds no
  symlink
- **THEN** the repository's own lint reports it by name and fails

### Requirement: A named guard resolves against the running command's own skill directory

`skills/myflow-contracts/pipeline.md` SHALL state, once, that a named guard resolves to
`<the running command's own skill directory>/scripts/<name>`. Every other file SHALL cite that
statement rather than restate it.

A skill or contract SHALL name an invoked guard by **basename**. It SHALL NOT give a path
relative to a repository root: such a path resolves only when the project being worked on is
this repository, which is the one case that never needs the guard shipped.

Resolution against the **running command's** skill directory is what lets a contract loaded by
more than one command name a guard at all. `skills/myflow-contracts/` is never a running
command and SHALL NOT carry a `scripts/` directory.

Prose that describes **this repository's own** lint and test guards is not an invocation, and SHALL
name the guard as `<agents repo>/scripts/<name>` rather than by a bare repository-relative path. A
bare path there resolves, for a reader standing in an installed project, against that project's own
tree — so the sentence names a file the reader may be able to write. Carrying the prefix says which
repository is meant, and it removes the need for any classifier to tell describing a guard from
running one.

#### Scenario: A contract loaded by two commands names one guard

- **WHEN** a contract file names a guard by basename
- **AND** it is loaded once by `/myflow-finish` and once by `/myflow-status`
- **THEN** each command resolves the guard inside its own skill directory

#### Scenario: An invoking call site carries a repository-relative path

- **WHEN** a skill's text invokes a guard by a path relative to a repository root
- **THEN** the repository's own lint reports that call site and fails

#### Scenario: Prose about this repository's own guard names its repository

- **WHEN** an installed file describes, without invoking, a guard belonging to this repository
- **THEN** it writes `<agents repo>/scripts/<name>`
- **AND** a bare `scripts/<name>` in that position is reported by the repository's own lint

### Requirement: A guard SHALL NOT derive a repository root from a fixed depth above itself

A guard SHALL NOT compute a repository root as a fixed number of levels above its own
directory. A guard reachable through more than one directory has more than one such answer,
and the wrong one is a directory that exists — so the guard proceeds against it and reports a
confident wrong result rather than an error.

Where a guard needs a repository root and was given none, it SHALL derive it from its own
**resolved physical** location, so that the answer does not depend on which path it was
invoked through.

#### Scenario: A guard is invoked through a skill's scripts directory

- **WHEN** a guard that derives a default repository root is invoked at
  `<skill-dir>/scripts/<name>` with no explicit root argument
- **THEN** the root it derives is this repository's root
- **AND** it is not the skill directory one level above the guard

### Requirement: A command SHALL report a missing guard once, at the start of the run

A `/myflow-*` command SHALL check, once at the start of its run, that every guard it can
invoke is present in its own `scripts/` directory. On a complete set it SHALL print nothing.

On any absence it SHALL print exactly one block, naming every missing guard, the directory
searched, and the command that installs them, and stating that the affected checks will be
performed by hand. It SHALL then continue.

The check is a report and SHALL NOT be a gate. Each contract's existing hand-run fallback
still governs what happens at the call site, and the handoff SHALL say that those checks were
run manually. A guard is never skipped for want of the script.

#### Scenario: Every guard is present

- **WHEN** a command starts and finds every guard it can invoke
- **THEN** it prints nothing about guards and proceeds

#### Scenario: Some guards are missing

- **WHEN** a command starts and finds that some of the guards it can invoke are absent
- **THEN** it prints one block naming each missing guard, the directory searched, and the
  install command
- **AND** the run continues under the hand-run fallback
- **AND** the handoff records that those checks were run manually

#### Scenario: A missing guard is not reported twice

- **WHEN** a command has printed the missing-guard block and later reaches a call site for one
  of those guards
- **THEN** it performs that check by hand without printing the block again


### Requirement: A skill directory SHALL carry symlinks only under its `scripts/` directory

A skill directory SHALL NOT contain a symlink directly under `skills/<skill>/`. Every symlink a
skill carries SHALL sit under `skills/<skill>/scripts/`, where **A guard a command invokes SHALL be
reachable from that command's installed skill** already governs it.

This closes the one placement that requirement does not reach. A file a skill's text names but does
not carry — most concretely `engineering-principles.md` and the reviewer-prompt files, which
`skills/myflow-do/SKILL.md` resolves **beside itself**, in `skills/myflow-do/`, including when
`/myflow-fast` is the command running that section — is resolved by reading where the text says it
lives, never by symlinking a copy into the command's own skill directory. That symlink makes a wrong
reading of the resolution rule work, which is what keeps the wrong reading alive.

The repository's own lint SHALL report such a symlink by path and target, and fail. Prose SHALL NOT
be the only countermeasure: `skills/myflow-do/SKILL.md` already states in terms that symlinking a
file in is the wrong fix, and a session created three such symlinks in
`skills/myflow-fast/` roughly five hours after that sentence was committed.

#### Scenario: A symlink appears directly under a skill directory

- **WHEN** `skills/<skill>/<name>` is a symlink, at the skill directory's top level rather than
  under its `scripts/` directory
- **THEN** the repository's own lint reports its path and its target, and fails

#### Scenario: A skill's scripts directory is unaffected

- **WHEN** `skills/<skill>/scripts/<name>` is a symlink into the repository's root `scripts/`
- **THEN** this requirement reports nothing, because that placement is the one the guard-reachability
  requirement above requires

#### Scenario: A skill carrying no symlink at its top level passes

- **WHEN** every entry directly under `skills/<skill>/` is a regular file or a directory
- **THEN** this requirement reports nothing
