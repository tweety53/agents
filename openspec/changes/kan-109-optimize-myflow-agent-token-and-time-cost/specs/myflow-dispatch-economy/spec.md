## ADDED Requirements

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
