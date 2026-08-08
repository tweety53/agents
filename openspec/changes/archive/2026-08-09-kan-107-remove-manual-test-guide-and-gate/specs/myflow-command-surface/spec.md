## MODIFIED Requirements

### Requirement: Git actions are bounded by state

`/myflow-start` SHALL perform no git actions beyond writing planning artifacts. `/myflow-do` SHALL
create or resume a worktree and stage with `git add`, and SHALL commit and push **only** when the
state file already records a `prUrl`. `/myflow-finish` SHALL commit, push, open a PR or merge on
its first run, and commit and push the archive and remove worktrees, branches and the remote branch
on its second.

No command other than `/myflow-finish`, and `/myflow-do` when a PR is already open, SHALL create a
commit.

Every staging pass `/myflow-do` performs SHALL exclude the paths the review diff excludes, so the
staging area holds implementation only. `/myflow-finish` SHALL stage those paths on its first run
and commit them **separately** from the implementation, implementation first.

`/myflow-do`'s commit-and-push exception SHALL make the same two commits, so a fix pushed to an open
PR keeps its code commits free of planning artifacts.

#### Scenario: A fix before integration stays staged

- **WHEN** `/myflow-do` completes for a change at `IN_PROGRESS` with no `prUrl` recorded
- **THEN** the changes are staged and uncommitted, and `git status` shows them as staged

#### Scenario: A fix after integration reaches the PR

- **WHEN** `/myflow-do` completes for a change whose state file records a `prUrl`
- **THEN** the fix is committed and pushed to the PR branch as two commits, implementation first

#### Scenario: Staging holds implementation only

- **WHEN** `/myflow-do` stages at the end of any run
- **THEN** no path under `openspec/` or `docs/superpowers/` is staged

#### Scenario: Finish commits both, in order

- **WHEN** `/myflow-finish` commits on its first run
- **THEN** the implementation is committed first and the planning artifacts second

### Requirement: One skill per command, named after it

Each pipeline command SHALL load exactly one skill, and that skill's directory SHALL be named
after the command: `skills/myflow-start/`, `skills/myflow-do/`, `skills/myflow-finish/`.

Content that a surviving skill delegates to a removed skill SHALL be inlined into the surviving
skill **before** the removed skill is deleted.

#### Scenario: Finish does not delegate to a deleted skill

- **WHEN** `/myflow-finish` runs after the restructure
- **THEN** the delta-sync and archive-move steps it previously delegated to `openspec-sync-specs`
  and `openspec-archive-change` are present in `skills/myflow-finish/`, and neither removed skill
  is referenced

#### Scenario: Start does not delegate to a deleted skill

- **WHEN** `/myflow-start` runs after the restructure
- **THEN** the artifact-creation steps it previously delegated to `openspec-propose` are present
  in `skills/myflow-start/`

#### Scenario: Do carries the run-instruction resolution

- **WHEN** `/myflow-do` runs after the restructure
- **THEN** the step that resolves the run instructions is present in `skills/myflow-do/`, and
  `openspec-manual-test-superpowers` no longer exists
