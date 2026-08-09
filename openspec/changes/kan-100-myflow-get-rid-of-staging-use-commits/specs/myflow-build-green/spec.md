## MODIFIED Requirements

### Requirement: Every task in a plan declares its build state

`/myflow-start`'s writing-plans stage SHALL tag every task in `tasks.md` with `**Build:** green` or
`**Build:** red`. A task tagged `red` SHALL additionally carry a `**Squash-with:** Task <N>` field
(defined in `myflow-task-commit-fields`), where `<N>` names the task whose commit this task's
commit folds into.

A `green` tag SHALL mean the project builds and this task's own verification command can run,
given only the state left by the preceding tasks. A `red` tag SHALL mean this task alone does not
leave a green build, and its `Squash-with:` field SHALL name the merge partner — the tag itself no
longer carries the partner inline.

#### Scenario: A task's ordering could not leave the build green alone

- **WHEN** a task in `tasks.md` cannot leave the project building and its own verification
  runnable using only the preceding tasks
- **THEN** the plan tags it `**Build:** red` and carries `**Squash-with:** Task <N>`, naming the
  task it is dispatched with as one unit

#### Scenario: An ordinary task leaves the build green

- **WHEN** a task in `tasks.md` leaves the project building and its own verification runnable
  using only the preceding tasks
- **THEN** the plan tags it `**Build:** green`

### Requirement: A guard enforces the build-state tags before the proposal publishes

A guard script SHALL parse `tasks.md` and fail the run when: a task has no `**Build:**` tag; a
task is tagged `red` with no `**Squash-with:**` field; a named `Squash-with:` partner does not
exist in the plan, or is not itself tagged `green`; or a `red` chain never resolves to `green`
before the plan's last task.

`/myflow-start` SHALL run this guard, when the project declares one, before publishing the
proposal artifact — the same point at which the plan-provenance guard runs — and SHALL fix any
hit before publishing.

The guard SHALL NOT attempt to verify that a `green` tag is true. It verifies only that every task
declares a build state and that no declared-red task is left unresolved.

#### Scenario: A missing tag blocks publish

- **WHEN** the build-green guard runs against a `tasks.md` where a task carries no `**Build:**` tag
- **THEN** the guard fails, naming the task, and `/myflow-start` does not publish the proposal
  artifact until the plan is fixed

#### Scenario: An unresolved red task blocks publish

- **WHEN** the build-green guard runs against a `tasks.md` where a task is tagged
  `**Build:** red` and its `**Squash-with:** Task <N>` names a task `<N>` that does not exist, or
  is itself tagged `red`
- **THEN** the guard fails, naming the unresolved task

#### Scenario: A resolved red chain passes

- **WHEN** the build-green guard runs against a `tasks.md` where every `red`-tagged task's
  `Squash-with:` names a partner that exists and is tagged `green`
- **THEN** the guard passes

#### Scenario: The guard does not claim to verify truth

- **WHEN** a task is tagged `**Build:** green` but its verification command would in fact fail
- **THEN** the guard passes regardless — it checks that the tag is present and consistent, not
  that the underlying claim is true
