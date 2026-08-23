# myflow-contract-economy delta — kan-295-cut-pipeline-load-cost-split-by-consumer

## ADDED Requirements

### Requirement: A section reachable from fewer than every command lives in its own file

Where a contract section loaded by every `/myflow-*` command carries rules reachable from fewer than
all of them, that section SHALL be moved into a file of its own, loaded by exactly the commands that
reach it.

This generalises **A section reachable from only one command lives in its own file**, which governs
the one-command case and is unchanged by this requirement. The saving is the same in kind and
smaller in degree: a section reachable from three of five commands stops being loaded by two.

**The threshold is reachability, never size.** A large section every command reaches stays in the
core; a small section two commands reach does not. `Rendering the session records` is 1486 bytes and
is moved, because `/myflow-start` and `/myflow-status` never render a record; `Stage marks` is 6630
bytes and stays, because every producing command marks a stage on every run.

**Reachability is established by reading, never by counting citations.** A command reaches a section
when a step it runs acts on that section's rules. A citation from a file the command loads is
evidence; the absence of one is not proof, because a command may act on a rule it was told about
elsewhere.

The moved file SHALL name, in its own opening prose, which commands load it, so a reader holding the
file can tell whether their command is one of them.

**Each command's own `SKILL.md` SHALL carry the instruction to load it**, at the step needing it.
`rules/myflow-manual-review.mdc`'s contract table SHALL NOT gain a row for such a file: that table is
read in every session whether or not a `/myflow-*` command runs, so a row there converts a per-run
saving into a permanent cost.

#### Scenario: A section two of five commands reach is moved

- **WHEN** a section in `pipeline.md` carries rules only `/myflow-do` and `/myflow-finish` act on
- **THEN** it is moved to its own file under `skills/myflow-contracts/`
- **AND** `/myflow-start`, `/myflow-fast` and `/myflow-status` no longer load it

#### Scenario: A large section every command reaches stays in the core

- **WHEN** a section is among the largest in `pipeline.md`
- **AND** every `/myflow-*` command acts on its rules
- **THEN** it stays in `pipeline.md`
- **AND** its size is not offered as a reason to move it

#### Scenario: The moved file names its own consumers

- **WHEN** a file produced by this rule is read
- **THEN** its opening prose names the commands that load it

#### Scenario: The always-on rule table gains no row

- **WHEN** `rules/myflow-manual-review.mdc` is read after such a move
- **THEN** its contract table carries no row for the moved file
- **AND** the instruction to load it is found in each consuming command's `SKILL.md` instead

#### Scenario: Reachability is not decided by citation count

- **WHEN** a section is considered for a move
- **THEN** each citation into it is read to decide whether the citing command acts on its rules
- **AND** a count of citations is not offered as the decision
