# myflow-artifact-economy Specification

## Purpose
TBD - created by archiving change kan-77-sdd-ledger-canonical-path. Update Purpose after archive.
## Requirements
### Requirement: A change's artifacts are written brief

Every artifact a `/myflow-*` run writes SHALL be written brief — bullets over prose, no preamble, no
recap, and no restatement of what another artifact in the same change already says. This binds
`proposal.md`, `design.md`, the delta specs, `tasks.md`, the SDD ledger, the review panel record and
the self-review report.

The rule SHALL be stated in `skills/myflow-contracts/pipeline.md`, which every `/myflow-*` command
loads before any other step, so that it reaches every run without each skill carrying its own copy.

**Brevity SHALL NOT withhold a fact.** A run SHALL compress wording only. It SHALL NOT drop a
decision, a reason, a measured number, an alternative that was ruled out, or a caveat in order to be
shorter — a shorter artifact that has lost a fact is a defect, not a brief artifact.

The following SHALL NOT be compressed, because a guard or a contract parses them byte for byte:

- a task's `**Files:**`, `**Tests:**`, `**Regression:**`, `**Baseline:**`, `**Commit:**`,
  `**Build:**` and `**Squash-with:**` fields;
- a plan's `verified:` / `unverified:` / `measured:` / `predicted:` provenance tags;
- a decision's or an open question's `**ID:**` and `**Status:**` lines;
- a delta spec's normative statements and their scenarios;
- the review panel record's marker blocks and findings table.

This SHALL narrow the "code, commits, docs and specs stay full" carve-out that the be-brief rule
states, for the artifacts named above and for no other file. No guard SHALL be added to measure an
artifact's length: brevity is a judgment, and a byte budget on a per-change artifact would reward
dropping the facts the paragraph above requires kept.

#### Scenario: An artifact restates another artifact in the same change

- **WHEN** `design.md` would repeat the problem statement `proposal.md` already carries
- **THEN** it cites or omits it rather than restating it

#### Scenario: Brevity does not cost a fact

- **WHEN** shortening a decision entry would drop the alternative that was ruled out
- **THEN** the alternative is kept and the wording around it is compressed instead

#### Scenario: A guard-parsed field is never compressed

- **WHEN** a task's `**Regression:**` field would read more briefly as a fragment
- **THEN** the field is written in full, because `check-task-commit-fields.py` reads it

#### Scenario: No length guard is added

- **WHEN** this requirement is implemented
- **THEN** no script measures a change artifact's length, and no budget row is added for one

