# myflow-dispatch-economy Specification

## Purpose
TBD - created by archiving change kan-109-optimize-myflow-agent-token-and-time-cost. Update Purpose after archive.
## Requirements
### Requirement: One implementer subagent per worktree at a time

`/myflow-do` SHALL keep at most **one** implementer subagent in flight against a given worktree at
any moment. Before dispatching the next implementer into a worktree, the parent SHALL have received
the previous implementer's commit sha for that worktree.

Dispatches into **different** worktrees SHALL remain free to run concurrently. The constraint is on
a shared worktree, and therefore on a shared build directory — not on concurrency as such.

This requirement SHALL explicitly override `superpowers:subagent-driven-development`'s parallel
dispatch guidance and `superpowers:dispatching-parallel-agents` wherever the tasks in question share
a worktree. A skill that dispatches implementers SHALL state the override rather than leaving the
two guidances to be reconciled at dispatch time.

The reason is measured rather than presumed: two implementers on one build directory produced
assertions left red at file seams, agents idling on another agent's mid-edit compile, and multiple
build-daemon sessions corrupting test-result XML, which made whole suite runs untrustworthy and
forced re-runs. Serialization removes the collision; it does not change what any implementer does.

This requirement SHALL NOT change the commit-per-task model, the per-task review shape, the review
panel, or the handoff bar.

#### Scenario: A second implementer waits for the first

- **WHEN** two dispatches are due against the same worktree
- **THEN** the second is dispatched only after the first has reported its commit sha

#### Scenario: Different worktrees still run concurrently

- **WHEN** a change affects three worktrees and each has work due
- **THEN** one implementer may be in flight in each of the three at the same time

#### Scenario: The override is stated, not inferred

- **WHEN** the skill that dispatches implementers is read
- **THEN** it names the upstream parallel-dispatch guidance it overrides for same-worktree tasks

### Requirement: Implementer dispatches are bundled by declared file overlap

`/myflow-do` SHALL group the plan's unchecked tasks into **bundles** by the overlap of their
declared `**Files:**` sets, and SHALL dispatch one implementer per bundle rather than one per
`tasks.md` checkbox.

Two tasks SHALL join the same bundle when they declare any common path. The relation SHALL be
transitive: three tasks A, B and C where A and B share a path and B and C share a different path
SHALL form one bundle.

A task's `**Allowed-collateral:**` glob SHALL NOT contribute to the join. It names paths a
legitimate sweep may also touch, not paths the task owns, and joining on it would collapse unrelated
tasks into one bundle.

Bundles SHALL be dispatched in the order of each bundle's lowest task id, so a plan whose tasks are
ordered by dependency stays ordered.

The grouping SHALL be computed by a script rather than by eye. The script SHALL take a `tasks.md`
path, and with no argument SHALL resolve every non-archived change's `tasks.md`, the same resolution
`check-task-build-green.sh` already performs. It SHALL print one line per bundle naming that
bundle's task ids, and SHALL exit **0** when bundles were computed, **1** when a task carries no
`**Files:**` field — naming that task — and **2** when it cannot answer at all.

A task with no `**Files:**` field SHALL be a reported failure rather than a guess. The field is
already required of every task by `myflow-task-commit-fields`, so its absence is a plan defect.

The script SHALL NOT be a lint step. It computes a grouping rather than judging a file's text, and a
lint run has no change in flight to compute one for. It SHALL be covered by its own test harness.

Bundling SHALL NOT change the commit-per-task model: an implementer handed a bundle still makes one
commit per task, carrying that task's own `Task-Id:` trailer, and a `Build: red` task still folds
into the commit its `**Squash-with:**` field names.

#### Scenario: Tasks sharing a file are dispatched together

- **WHEN** two unchecked tasks each declare `src/Foo.kt` under `**Files:**`
- **THEN** they form one bundle and one implementer is dispatched for both

#### Scenario: Overlap is transitive

- **WHEN** task 1 and task 2 share one path, and task 2 and task 3 share a different path
- **THEN** all three form a single bundle

#### Scenario: Disjoint tasks are dispatched separately

- **WHEN** no two unchecked tasks declare a common path
- **THEN** each task forms its own bundle

#### Scenario: Allowed collateral does not join bundles

- **WHEN** two tasks declare disjoint `**Files:**` sets and one of them carries an
  `**Allowed-collateral:**` glob matching a path the other declares
- **THEN** they remain in separate bundles

#### Scenario: A task with no Files field stops the computation

- **WHEN** an unchecked task carries no `**Files:**` field
- **THEN** the script exits 1 and names that task, rather than grouping it by guess

#### Scenario: Checked tasks are not bundled

- **WHEN** a plan contains completed tasks marked `[x]`
- **THEN** they take no part in any bundle

#### Scenario: One commit per task survives bundling

- **WHEN** an implementer is dispatched for a bundle of three tasks
- **THEN** it makes three commits, one per task, each carrying that task's own `Task-Id:` trailer

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

