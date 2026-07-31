# myflow-handoff-output Specification

## Purpose
TBD - created by archiving change kan-8-myflow-updates. Update Purpose after archive.
## Requirements
### Requirement: Every handoff ends with the next command as its last line

Every `/myflow-*` command SHALL end its output with the next command to run, on its own line,
bare and copy-pasteable, with no prose, formatting or content after it.

An agent cannot drive a harness's autocomplete — nothing lets a running session prefill the
operator's input — so the last-line convention together with a small command surface SHALL be
the whole mechanism.

#### Scenario: Nothing follows the next command

- **WHEN** any pipeline command completes
- **THEN** the final line of its output is the next command, and no summary, hint or closing
  remark appears after it

#### Scenario: A terminal state names no next command

- **WHEN** `/myflow-finish` completes run 2 and the change is `FINISHED`
- **THEN** it reports that the change is finished rather than printing a next command

#### Scenario: Finish run 1 names itself as the next command

- **WHEN** `/myflow-finish` completes its integration run and stops
- **THEN** the final line is `/myflow-finish <name>` again, because that is what the operator runs
  once the branch is merged

### Requirement: Handoffs carry only what the operator must act on

A handoff SHALL contain what actually happened in one to three lines, the absolute paths of
anything the operator needs to open, and the next command. It SHALL NOT restate the plan,
enumerate completed internal steps, or repeat content available at a path it just gave.

#### Scenario: A guide is linked, never pasted

- **WHEN** `/myflow-do` writes a manual test guide
- **THEN** the reply contains the guide's absolute path and not its body

#### Scenario: A diff is described, never dumped

- **WHEN** `/myflow-do` completes
- **THEN** the reply gives the worktree path and the commands to inspect the diff, not the diff
  itself

### Requirement: Every path in every output is absolute

Every path a `/myflow-*` command prints — in handoffs, in generated manual test guides, in
IntelliJ open commands, and in run instructions — SHALL be absolute.

A relative path, a sibling-relative path such as `../<other-app>`, or a main-checkout path used
while an apply worktree holds the work SHALL NOT be emitted. App roots SHALL be resolved from
`git worktree list` or the state file's `worktrees` keys.

#### Scenario: A generated guide is runnable from anywhere

- **WHEN** a manual test guide gives a command to start an app
- **THEN** every path in that command is absolute and points at the apply worktree for that app

#### Scenario: The IntelliJ command is absolute

- **WHEN** a command prints an `open -na "IntelliJ IDEA"` line
- **THEN** its argument is an absolute path resolved from `git worktree list`

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

