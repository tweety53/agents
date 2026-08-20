## MODIFIED Requirements

### Requirement: Git actions are bounded by state

`/myflow-start` SHALL perform no git actions beyond writing planning artifacts. `/myflow-do` SHALL
create or resume a worktree and stage with `git add`, and SHALL commit and push **only** when the
state file already records a `prUrl`. `/myflow-finish` SHALL commit, push, open a PR or merge on
its first run. On its second run it SHALL commit the archive on `chore/archive-<name>`, remove
worktrees, branches and the remote branch, and — after self-review has committed its report onto
that same branch — push `chore/archive-<name>` and open a pull request against the base branch.

**Run 2 SHALL NOT push the base branch.** The archive reaches the base branch only through that
pull request, never through a local push.

This is scoped to run 2 deliberately. **Run 1's merge-and-push route does push the base branch** —
it pushes the change branch, merges it into the base branch, and pushes that — and this change does
not touch it: that route is one of three the operator explicitly chooses between, and it is not the
unasked-for direct push run 2 used to perform.

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

#### Scenario: Run 2 never pushes the base branch

- **WHEN** `/myflow-finish` completes its second run
- **THEN** the only branch pushed is `chore/archive-<name>`, and the base branch is reached solely
  through the pull request that branch opens — run 1's merge-and-push route, which does push the
  base branch when the operator chooses it, is unaffected
