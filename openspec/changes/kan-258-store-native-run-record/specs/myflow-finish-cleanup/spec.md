## MODIFIED Requirements

### Requirement: Session records are preserved in the repository

The SDD ledger and the review panel record SHALL be **rendered from the store** into the repository,
so that they are readable there without a running daemon and survive the worktree's removal. They
SHALL NOT be copied out of a working directory, because neither is authored as a file in one.

The ledger render SHALL happen during `/myflow-finish` run 1, before it stages the work, so the
records land in the same commit as the implementation they describe and reach the base branch with
it. It SHALL also happen on `/myflow-do`'s commit path — the case where a pull request already exists
and a fix is committed and pushed — so a fix round raised after integration refreshes the records
rather than leaving them stale. The panel record's own render SHALL happen earlier, at panel close;
that firing point belongs to `myflow-run-record` and this requirement does not restate it.

The destination path for a change SHALL be fixed at the first render and reused on every later
render, so repeated runs overwrite in place rather than accumulating one dated file per round.

A kind the store holds no rows of SHALL be reported and SHALL NOT fail the run: a change may
legitimately have no panel record, and a step able to block an integration would be a worse failure
than the gap it closes. That report SHALL be distinguishable from every failure outcome, and — unlike
the file-copy step this replaces — it SHALL mean unambiguously that no record exists, never that one
exists somewhere that was not read.

Every destination the render writes to SHALL be required to resolve inside the repository root. A
path resolving outside it SHALL be refused: reported as a failure of that one render, distinctly from
the report of an empty record, while the remaining records are still rendered. The refusal exists
because the render runs automatically and its result is committed and pushed, so a planted symlink at
a destination would otherwise write an arbitrary file under the repository's name.

A change name SHALL be required to be a single plain path component before any directory is touched.
A name carrying a path separator or a glob metacharacter SHALL be rejected outright, because the
search for an already-rendered file is anchored on the name and a metacharacter in it would let one
change adopt and overwrite another change's rendered record.

Records other than the ledger and the panel record SHALL NOT be rendered. Per-task diffs duplicate
commits already present in git history.

Rendering SHALL NOT change what the ledger may contain. A dispatch whose model the dispatcher could
not observe SHALL still record `unknown (agent-defined)`; durability SHALL NOT be a reason to fill
such an entry with a plausible value.

The script that performed the file copy, and its test harness, SHALL be removed rather than kept as a
wrapper. The pipeline stage that invoked it SHALL keep its existing stage key and its position in
run 1, so that stage runs already recorded under that key remain valid.

#### Scenario: The ledger survives the change

- **WHEN** a change is integrated and later archived, and its worktree is removed
- **THEN** the SDD ledger is present in the repository at its rendered path
- **AND** a reader can still determine which model implemented each task, with no daemon running

#### Scenario: An empty record is reported, not fatal, and not ambiguous

- **WHEN** a change has no review panel finding in the store
- **THEN** the absence is reported in the run's output as "no rows for this change"
- **AND** the integration proceeds and the remaining records are still rendered
- **AND** that report cannot also mean a record was written somewhere that was not read

#### Scenario: A fix round refreshes rather than duplicates

- **WHEN** a fix is committed to a branch that already has a pull request, after records were already
  rendered
- **THEN** the existing rendered files are overwritten in place
- **AND** no second dated copy is created for the same change

#### Scenario: A destination outside the repository is refused, not followed

- **WHEN** one of the destination directories under `docs/superpowers/` resolves outside the
  repository root
- **THEN** nothing is written through it, the refusal is reported as a failure rather than as an empty
  record, and the remaining records are still rendered

#### Scenario: A change name that is not one plain component is rejected

- **WHEN** the render is invoked with a change name containing a path separator or a glob
  metacharacter
- **THEN** it is rejected before any directory is created or any file is written
- **AND** no other change's rendered record is read or overwritten

#### Scenario: An unobservable model stays unobserved

- **WHEN** a rendered ledger contains a slot dispatched by agent type, whose model the dispatcher
  never read
- **THEN** that entry records `unknown (agent-defined)`
- **AND** rendering does not substitute a guess

#### Scenario: The retired script is gone and nothing names it

- **WHEN** the reference and symlink guards run after the retirement
- **THEN** the copying script and its test harness are absent, and no skill, contract, guard-presence
  list or installer names either of them
