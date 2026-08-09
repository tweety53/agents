# myflow-task-commits Specification

## Purpose
TBD - created by archiving change kan-100-myflow-get-rid-of-staging-use-commits. Update Purpose after archive.
## Requirements

### Requirement: Each task commits after finishing, before review

`/myflow-do` SHALL replace its `NO-COMMITS` implementer dispatch clause with a `COMMIT-PER-TASK`
clause. After an implementer finishes a task (RED-GREEN-REFACTOR complete), it SHALL commit that
task's work with a subject of the form `task(<n>): <short subject>` and a `Task-Id: <n>` trailer,
where `<n>` is the task's dotted id from its `tasks.md` heading, **before** the task is dispatched
for review.

#### Scenario: A finished task is committed before review

- **WHEN** an implementer finishes a task's RED-GREEN-REFACTOR cycle
- **THEN** it commits the task's changes with a `task(<n>): <subject>` subject and a `Task-Id: <n>`
  trailer
- **AND** only after that commit exists is the task dispatched for review

#### Scenario: No uncommitted work is left after a task

- **WHEN** a task's commit is made
- **THEN** the worktree carries no uncommitted changes belonging to that task

### Requirement: Review reads a real commit diff, not a checkpoint snapshot

Per-task and per-fix-round review SHALL read `git diff <task-base>..<task-sha>` (or the
autosquashed result of it) directly. `/myflow-do` SHALL NOT use `skills/myflow-do/scripts/checkpoint`
or `skills/myflow-do/scripts/uncommitted-review-package`; both scripts are removed.

#### Scenario: A reviewer reads a commit range

- **WHEN** a task's commit exists and review is dispatched
- **THEN** the reviewer is given a diff of that task's own commit range, produced directly from git
  history rather than from a snapshot script

#### Scenario: The retired scripts are absent

- **WHEN** `/myflow-do` runs under this requirement
- **THEN** `skills/myflow-do/scripts/checkpoint` and
  `skills/myflow-do/scripts/uncommitted-review-package` do not exist in the skill directory

### Requirement: A fix-round change commits as a fixup and is autosquashed immediately

When a review finding requires a code change to an already-committed task, the fix SHALL be
committed with `git commit --fixup=<task-sha>` and immediately folded into the task's commit via
`git rebase --autosquash`, before the next review pass reads the diff.

#### Scenario: A fix round folds into its task's commit

- **WHEN** a reviewer raises a finding against a committed task and the fix is applied
- **THEN** the fix is committed as `git commit --fixup=<task-sha>`
- **AND** `git rebase --autosquash` runs before the next review pass, leaving one commit for that
  task

#### Scenario: Review never sees a trailing fixup commit

- **WHEN** a re-review is dispatched after a fix round
- **THEN** the diff it reads reflects the task commit with the fixup already folded in, never a
  separate fixup commit alongside it

### Requirement: A red task's commit folds into its green partner

A task tagged `Build: red` SHALL have its commit folded into its `Squash-with:` partner's commit via
the same fixup-and-autosquash mechanism used for fix rounds, rather than remaining a separate commit
on the branch.

#### Scenario: A red task's commit is folded into its partner

- **WHEN** a task tagged `Build: red` with `Squash-with: Task <N>` is committed
- **THEN** its commit is folded into Task `<N>`'s commit via `--fixup`/`--autosquash`, leaving one
  commit for the merged pair

### Requirement: `/myflow-do` still never pushes, merges, or opens a PR

Committing per task SHALL NOT change `/myflow-do`'s boundary against push, merge, or PR creation:
those remain `/myflow-finish`'s responsibility exclusively, on every run — first run and fix run
alike.

#### Scenario: A first run commits but does not push

- **WHEN** `/myflow-do` runs from `STARTED` and commits every task
- **THEN** no push, merge, or PR-creation command is run

#### Scenario: A fix run commits but does not push

- **WHEN** `/myflow-do` runs from `IN_PROGRESS` with no `prUrl` recorded and commits a fix
- **THEN** no push, merge, or PR-creation command is run
