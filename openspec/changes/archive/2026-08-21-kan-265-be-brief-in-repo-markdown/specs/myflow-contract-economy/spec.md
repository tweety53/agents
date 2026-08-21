# myflow-contract-economy delta — kan-265-be-brief-in-repo-markdown

## ADDED Requirements

### Requirement: Repository prose states a thing once

`rules/be-brief.mdc` SHALL carry a clause governing written prose, alongside the clause it already
carries for what an agent says. The clause SHALL name exactly two subjects:

- **this repository's own Markdown**, named explicitly rather than generalised to any repository's
  documentation, because `be-brief.mdc` installs globally into every project's `CLAUDE.md` and a
  general clause would govern projects it was never evaluated against;
- **the artifacts a `/myflow-*` run generates, in any project** — `proposal.md`, `design.md`,
  `tasks.md`, delta specs, panel records and self-review reports — because those are written by this
  repository's own pipeline wherever they land.

Another project's pre-existing documentation SHALL NOT be subject to the clause.

The clause SHALL state brevity as **non-repetition**: a file states a thing once and does not
restate what another file states canonically. `be-brief.mdc`'s existing sentence exempting written
files SHALL be reworded to state **completeness** — every requirement, scenario, exit code and
worked example present — so the two are readable as the separate properties they are. A file SHALL
be required to satisfy both, and the rule's text SHALL NOT leave them reading as alternatives.

The clause SHALL permit **deletion** and SHALL NOT permit **paraphrase**. A passage may be removed;
it SHALL NOT be reworded to say the same thing in fewer words, because a normative sentence re-said
differently is a changed requirement that no guard detects, while a deletion is visible in a diff
and provable against the normative inventory.

The clause SHALL name what is never cut: any normative sentence, any exit-code contract, any
ordering constraint, any scenario, any worked example, and any recorded reason a rejected
alternative was rejected.

#### Scenario: A passage restating a canonical statement elsewhere

- **WHEN** a file restates what another file states canonically and already cites
- **THEN** the restating passage may be deleted

#### Scenario: A normative sentence is never cut

- **WHEN** a passage carries a SHALL, SHALL NOT, MUST or MUST NOT sentence
- **THEN** it is out of scope for the trim, however repetitive it reads

#### Scenario: Density rewriting is not brevity

- **WHEN** a passage is reworded to say the same thing in fewer words
- **THEN** the clause is violated, because only deletion is permitted

#### Scenario: A generated artifact in another project

- **WHEN** a `/myflow-*` run writes `tasks.md` into a consuming project
- **THEN** that file is subject to the clause
- **AND** the same project's pre-existing `README.md` is not

### Requirement: A byte budget ratchets every owned Markdown file

`scripts/check-contract-budget.sh` SHALL fail when a file it covers exceeds its declared budget.

- It SHALL cover **every** `*.md` and `*.mdc` file this repository owns — `skills/`, `rules/`,
  `openspec/specs/`, `commands/`, `commands-claude/`, `.myflow/` and the repository root — and SHALL
  exclude `node_modules/`, `.superpowers/`, `openspec/changes/archive/` and `docs/superpowers/`.
  Those exclusions SHALL be structural rather than a list of paths, so a file added inside an
  excluded tree needs no edit to the guard. A covered file with no budget entry SHALL be a failure,
  so a file added later cannot silently escape the ratchet.
- Coverage SHALL NOT be limited to files the pipeline loads per run. Regrowth is the failure mode
  and it does not respect that boundary: at the time this requirement was widened, 79 owned files totalling 1.16 MB sat outside the ratchet, of which `openspec/specs/` alone was 462 KB.
- Budgets SHALL be declared as a path-to-maximum-bytes table **inside the script**. One place to
  read; raising a budget is then a visible diff in the guard rather than an edit to the file being
  ratcheted.
- The table SHALL be **hand-maintained**. The guard SHALL NOT offer a mode that rewrites the table
  from current sizes: a regenerate command makes regeneration the path of least resistance, and a
  ratchet whose bound is recomputed on demand does not ratchet.
- Budgets SHALL be per file. A directory total SHALL NOT stand in for the files under it, since one
  file may balloon while another shrinks with the total unchanged.
- It SHALL measure **bytes**, not lines and not tokens — deterministic, and dependent on no
  tokenizer.
- Each budget SHALL be the size the file actually has when the change declaring it lands, **plus
  25%**. The guard is a ratchet against regrowth and SHALL NOT be a target a split can fail against;
  a target-first budget creates pressure to push normative text into an appendix to make a number.
- **The table SHALL be regenerated once, as the last action of a change that alters any covered
  file.** A table computed while later edits are still landing goes stale silently, and it did so
  three times in one change before this rule was written.
- A budget row SHALL be written only after the change's normative-sentence inventory has been shown
  unchanged for that file's group, so a bad trim is not locked in by the ratchet that follows it.
- It SHALL be argument-free and self-scoped from its own location, like `check-references.sh` and
  `check-vocabulary.sh`, with an opt-in environment override for its own test harness alone.
- It SHALL refuse a covered path that is a symlink rather than following it, before reading it.
- Its exit codes SHALL be `0` every file within budget, `1` a file over budget or missing an entry,
  `2` the guard cannot answer at all.

#### Scenario: A file over its budget fails the guard

- **WHEN** a covered file exceeds its declared budget
- **THEN** the guard prints the path, the budget and the actual size
- **AND** it exits 1

#### Scenario: A new file with no budget fails the guard

- **WHEN** a `.md` file is added to a covered location with no entry in the budget table
- **THEN** the guard exits 1 naming that file

#### Scenario: A skill file is covered

- **WHEN** `skills/myflow-do/SKILL.md` grows past its declared budget
- **THEN** the guard exits 1 naming it

#### Scenario: A spec file is covered

- **WHEN** a file under `openspec/specs/` grows past its declared budget
- **THEN** the guard exits 1 naming it

#### Scenario: An archived change is not covered

- **WHEN** a `.md` file under `openspec/changes/archive/` exceeds any size
- **THEN** the guard reports nothing about it, archived changes being a record rather than
  maintained text

#### Scenario: The guard runs from any working directory

- **WHEN** the guard is invoked from outside the repository
- **THEN** it scans the repository it lives in
- **AND** it does not report a false clean by scanning nothing

#### Scenario: The guard is a declared lint step

- **WHEN** `.myflow/project.md`'s `## lint` section is read
- **THEN** `scripts/check-contract-budget.sh` is listed
- **AND** `scripts/test-check-contract-budget.sh` is listed under `## test`

## REMOVED Requirements

### Requirement: A byte budget ratchets every file the pipeline loads per run

**Reason**: The requirement is widened, not dropped. Its coverage rule bounded the ratchet to files
the pipeline loads on a run — `skills/myflow-contracts/` plus each `skills/*/SKILL.md` and its
rationale sibling — and regrowth does not respect that boundary: 79 owned files totalling 1.16 MB
sat outside it, `openspec/specs/` alone accounting for 462 KB. Since the bound is in the
requirement's own title, it is removed and restated rather than edited in place.

**Migration**: Read **A byte budget ratchets every owned Markdown file** above. Every rule this
requirement stated is carried there unchanged — the in-script table, byte measurement, the
size-plus-25% rule, one regeneration as a change's last action, self-scoping, the symlink refusal
and the three exit codes. What is new is the widened coverage, the structural exclusions, the
hand-maintained table, the per-file bound, and the inventory precondition on writing a row.
