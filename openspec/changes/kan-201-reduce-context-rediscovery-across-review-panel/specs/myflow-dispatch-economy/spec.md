## ADDED Requirements

### Requirement: A dispatching stage gathers one shared context bundle

`/myflow-do` SHALL gather this change's planning context into a single bundle at
`<worktree>/.superpowers/sdd/dispatch-context.md` before dispatching any subagent, and SHALL produce
it with a script rather than by hand, so that every dispatch in a stage is given demonstrably
identical inputs.

The bundle SHALL carry, in order, the change's `proposal.md`, `design.md`, `tasks.md`, every delta
spec under the change's `specs/` directory, and the content of `engineering-principles.md`.

The absolute path of `engineering-principles.md` SHALL be passed to the gather script by its caller
and SHALL NOT be re-derived inside the script. The rule locating that file belongs to the skill that
dispatches, and a second copy of it inside a script would be free to drift from the first.

The `## standards` entries resolved for the principles slot SHALL NOT be carried in the bundle. They
resolve through the entry-form table and containment rule of the project-configuration contract, they
are read by one slot rather than by every dispatch, and re-implementing that containment rule inside
the gather script would duplicate the contract it depends on.

The bundle SHALL state, in its header, the instant it was generated and the git sha of the worktree's
`HEAD`, so any reader can establish which state of the tree the bundle describes.

#### Scenario: The bundle carries every planning source

- **WHEN** the gather script runs against a change whose directory holds a proposal, a design, a plan
  and one delta spec
- **THEN** the bundle carries all four, plus the engineering principles content, each under its own
  heading

#### Scenario: The principles path is validated, never derived

- **WHEN** the gather script is invoked
- **THEN** it uses the principles path its caller passed, and derives no such path of its own

#### Scenario: Standards files are absent from the bundle

- **WHEN** the project declares `## standards` entries
- **THEN** the bundle carries none of them, and the principles slot resolves them as it did before

#### Scenario: The bundle states what it describes

- **WHEN** a bundle is read
- **THEN** its header names the generating instant and the `HEAD` sha it was generated against

### Requirement: The bundle is rebuilt per dispatching stage, never once per run

`/myflow-do` SHALL rebuild the bundle at the start of the stage that dispatches implementers, at the
start of the stage that dispatches the review panel, and before each fix round, writing to the same
path each time.

A bundle gathered once per run SHALL NOT be reused across those points. A fix documented before it is
implemented edits `proposal.md` and `tasks.md`, so a run-scoped bundle would leave every later
dispatch reading a plan that no longer exists — a correctness failure, not merely a stale-cost one.

The bundle SHALL NOT be rebuilt per dispatch. Rebuilding between two slots of the same panel pass
would leave those slots reading different inputs, defeating the property that makes a shared bundle
meaningful.

#### Scenario: A fix round re-gathers before dispatching

- **WHEN** a fix round is about to dispatch, after the plan was edited to document the fix
- **THEN** the bundle is rebuilt first, and carries the edited plan

#### Scenario: Two slots of one pass read the same bundle

- **WHEN** a panel pass dispatches three slots
- **THEN** all three are given the same bundle, generated once at that stage's start

### Requirement: The bundle is advisory, never authoritative

Every dispatch prompt that names the bundle SHALL state that the dispatch may open any file the
bundle names, and that it MUST still read the actual diff and the actual code it is reviewing or
changing.

A dispatch SHALL NOT be instructed that the bundle is its only permitted planning source. Sharing a
dispatch's *inputs* is the whole intent; sharing its *conclusions* is not, and a panel's value depends
on each slot reasoning independently over the real artefact.

#### Scenario: A slot still reads the diff

- **WHEN** a review slot is dispatched with the bundle
- **THEN** its prompt requires it to read the review diff itself, and the bundle does not stand in for
  that diff

#### Scenario: A dispatch may open a bundled file directly

- **WHEN** a dispatch judges the bundle incomplete for its purpose
- **THEN** nothing in its prompt forbids it opening the original file

### Requirement: A missing bundle never stops a run

Where the gather script cannot be located, `/myflow-do` SHALL report its absence through the existing
guard presence check and SHALL dispatch with the prompt shape it would have used before this
capability existed.

The gather script SHALL exit 2 on a malformed invocation — a missing argument, a change name outside
the allowlist, or a path that does not resolve where it was required to — and SHALL otherwise always
exit 0, reporting a source it could not find as skipped rather than as a failure. A change may
legitimately carry no design document.

The gather script SHALL validate each path it is given by normalizing it lexically, resolving it
semantically, and refusing any divergence between the two, so that a symlink anywhere between the
trusted root and the leaf is refused rather than followed.

#### Scenario: The script is not installed

- **WHEN** the gather script is absent from the running command's own scripts directory
- **THEN** the guard presence check names it, and dispatches proceed without a bundle

#### Scenario: A source is absent

- **WHEN** the change carries no `design.md`
- **THEN** the bundle reports that source skipped and the script exits 0

#### Scenario: A path resolves through a symlink

- **WHEN** any passed path diverges between its lexical and its resolved form
- **THEN** the script refuses the invocation and exits 2
