# myflow-manual-test-guide Specification

## Purpose
TBD - created by archiving change kan-26-operator-output-and-configuration. Update Purpose after archive.
## Requirements
### Requirement: The guide is a behaviour checklist at capability scope

`docs/manual-test/<name>.md` SHALL consist of one tickable line per user-visible behaviour, grouped
by capability, scoped to the change's blast radius rather than to its plan tasks.

Each line SHALL be phrased as the check to perform, in the register an operator would use — for
example `check exercise update — the "key" field saves` — and SHALL NOT restate the requirement it
came from.

The guide SHALL NOT carry per-step command transcripts, expected-output blocks, or an explanation of
why a check exists. A change that touched every part of exercise CRUD SHALL list create, update,
filter, sort and delete, rather than one entry per plan task that produced them.

A guide written per plan task grows with the implementation rather than with the behaviour, which is
what made previous guides long without making them more thorough: several entries could exercise one
behaviour while another went unlisted.

#### Scenario: Behaviours are listed, not tasks

- **WHEN** a change implements exercise create, update, filter, sort and delete across eleven plan
  tasks
- **THEN** the guide carries one line per behaviour
- **AND** it does not carry one line per plan task

#### Scenario: A check states what to do, not why

- **WHEN** a guide line is written for a behaviour
- **THEN** it names the check to perform
- **AND** it carries no rationale, transcript or expected-output block

#### Scenario: Scope follows the blast radius

- **WHEN** a change modifies one field of an entity whose other behaviours it did not touch
- **THEN** the guide covers the behaviours the change could have affected
- **AND** it does not enumerate the whole application

### Requirement: The guide keeps the shapes its guards read

The guide SHALL keep an unordered checkbox list using the `- [ ]` and `- [x]` markers, and SHALL keep
its `## Known incomplete` section, written on every run and refreshed on every fix run.

It SHALL keep a preamble stating how to run whatever is in scope, with every path absolute per
`myflow-handoff-output`.

`scripts/check-unfinished-work.sh` reads both the checkboxes and the `## Known incomplete` section,
the latter through a scan implementing CommonMark's fence rule. Simplifying the guide's register
SHALL NOT alter either shape: the register is prose, the two shapes are machine-read, and changing
them would break a guard while appearing only to shorten a document.

#### Scenario: Checkbox syntax survives the simplification

- **WHEN** a guide is written under this capability
- **THEN** its checks use `- [ ]` and `- [x]`
- **AND** the unfinished-work guard counts them exactly as before

#### Scenario: The known-incomplete section is still written

- **WHEN** a run completes with nothing outstanding
- **THEN** the guide still carries a `## Known incomplete` section reading `None.`
- **AND** its absence would still be treated as outstanding rather than as clear

#### Scenario: The preamble keeps absolute paths

- **WHEN** the preamble gives a command to run something in scope
- **THEN** every path in that command is absolute

### Requirement: A repository with no runnable application states its checks as commands

Where a project has no runnable application, each check SHALL be stated as the command to run, one
line each, tickable in the same way.

The guide SHALL NOT be given an application-shaped structure a project does not have. This
repository is that case: it is the source of the myflow skills, commands and rules, and "running the
apps" means running its guard scripts, its assertion harnesses and a sandboxed installer pass.

#### Scenario: A source-only repository gets a runnable guide

- **WHEN** a guide is written for a repository whose `.myflow/project.md` declares no runnable
  application
- **THEN** each check is the command to run, one per line
- **AND** the guide does not describe an application, a port or a URL that does not exist
