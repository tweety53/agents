## MODIFIED Requirements

### Requirement: Run 1 asks how to land the branch, before doing anything

On its first run `/myflow-finish` SHALL ask, **before performing any git action**, how the branch
should be landed:

- **Open a pull request** *(default, recommended)* — commit, push, open a PR, record `prUrl`, stop
- **Merge and push** — commit, push, merge into the base branch, push that
- **Handle it manually** — commit and push the branch only, then stop and say what is left to do

The unfinished-work gate SHALL run **before** this question. The operator SHALL NOT be asked how to
land a branch and only then told that the branch carries unfinished work.

Having asked once, the run SHALL proceed to completion without a further interruption. The answer
SHALL never be remembered between runs and never inferred from anything else.

Before committing anything, every route SHALL first reshape the branch: `git reset --soft` to the
recorded merge base, collapsing every per-task and fixup commit `/myflow-do` made back into the
working tree. Both routes that commit SHALL then produce **two** commits from that reshaped state:
the implementation first, then the change's planning artifacts and session records. Neither route
SHALL mix the two, and neither SHALL leave any of `/myflow-do`'s per-task commits on the final
branch.

#### Scenario: The default opens a PR and stops

- **WHEN** the operator accepts the default
- **THEN** the branch is committed, pushed and a PR opened; `prUrl` is recorded; nothing is merged

#### Scenario: Merging is chosen

- **WHEN** the operator chooses to merge and push
- **THEN** the branch is merged into the base branch and pushed, and the next `/myflow-finish`
  run will archive

#### Scenario: Manual handling commits but does not integrate

- **WHEN** the operator chooses to handle it manually
- **THEN** the branch is committed and pushed, no PR is opened and nothing is merged, and the
  handoff states what the operator must do before running `/myflow-finish` again

#### Scenario: The base branch is resolved, never assumed

- **WHEN** finish needs the base branch
- **THEN** it resolves the repository's default branch rather than assuming `main` or `develop`

#### Scenario: The gate precedes the question

- **WHEN** run 1 starts for a change carrying unfinished work
- **THEN** the unfinished-work breakdown is shown and answered before the landing question appears

#### Scenario: Implementation and planning artifacts land in separate commits

- **WHEN** any committing route runs
- **THEN** one commit carries the implementation and a second carries the planning artifacts and
  preserved session records

#### Scenario: Per-task commits are reshaped away before committing

- **WHEN** `/myflow-do` left the branch carrying multiple per-task and fixup commits
- **THEN** run 1's `git reset --soft` to the recorded merge base collapses them into the working
  tree before the implementation commit is made
- **AND** the final branch carries no trace of the individual task commits — only the two commits
  this requirement produces
