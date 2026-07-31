## MODIFIED Requirements

### Requirement: The review diff excludes planning artifacts

The diff a command presents for review at `IN_PROGRESS` SHALL exclude paths under `openspec/`,
`docs/manual-test/` and `docs/superpowers/`.

The exclusion SHALL be achieved by **not staging** those paths, not by filtering them out of a
display command. A filtered display leaves the artifacts in the staging area, where every other view
of it — a graphical client, `git status`, the IDE's staged-changes pane — shows them again, so the
reviewer sees exactly what the exclusion was meant to remove.

The list of excluded paths SHALL be fixed in the contract rather than configured per project. The
pipeline chooses these paths itself, so no project can differ.

`/myflow-do` SHALL therefore stage with those paths excluded, and its handoff SHALL present the
staged diff with no filter of its own. `/myflow-finish` SHALL stage and commit the excluded paths in
a separate commit from the implementation, so nothing is lost by leaving them unstaged earlier.

The plan was read at `STARTED`; presenting it again as code to review is noise that hides the
implementation diff it is mixed into.

#### Scenario: Planning artifacts are absent from the review diff

- **WHEN** `/myflow-do` hands off at `IN_PROGRESS` and gives commands to inspect the staged diff
- **THEN** `proposal.md`, `tasks.md`, the delta specs, the manual test guide and the preserved
  session records do not appear

#### Scenario: Planning artifacts are absent from the staging area, not merely from one command

- **WHEN** the operator inspects the staging area by any means at `IN_PROGRESS`
- **THEN** the excluded paths are not staged, so no view of the staging area shows them

#### Scenario: The review command needs no filter

- **WHEN** `/myflow-do`'s handoff prints the command to inspect the staged diff
- **THEN** that command carries no exclusion pathspec, because there is nothing staged to exclude

#### Scenario: Planning artifacts are still committed

- **WHEN** `/myflow-finish` runs its first run
- **THEN** the excluded paths are staged and committed in a commit separate from the implementation
