## MODIFIED Requirements

### Requirement: Run 1 refuses to integrate silently over unfinished work

Before it asks how the branch should land, and before any git action, `/myflow-finish` SHALL
determine whether the change carries unfinished work, by running a script rather than by asserting
it in prose. The script SHALL be run once per worktree recorded in the state file's `worktrees` map.

Two signals SHALL count as unfinished:

- unchecked items in `openspec/changes/<name>/tasks.md` and in any nested `<name>-fix-N` sub-change
- findings whose recorded status is open in `.superpowers/sdd/final-review-panel.md`

The script SHALL print exactly one verdict line — `CLEAR` when no signal fires, or `OUTSTANDING`
with a per-signal breakdown — and SHALL exit non-zero **with no verdict line** when it cannot read
the worktree. A run that receives no verdict line SHALL stop and ask the operator; it SHALL NOT read
missing output as either verdict, and SHALL check the exit code as well as the line.

A file it cannot find SHALL count as outstanding rather than as clear. Treating silence as clearance
is the failure this requirement exists to prevent.

On `OUTSTANDING` the run SHALL show the breakdown and offer the operator exactly three courses:
continue and integrate anyway; stop so the work can be finished; or file a tracker issue carrying
the outstanding items and then continue. Filing SHALL follow the labelling rules for issues the
pipeline creates.

Choosing to stop SHALL leave the change at `IN_PROGRESS` with no git action performed.

#### Scenario: An unticked plan stops the run before it asks how to land

- **WHEN** run 1 begins for a change whose `tasks.md` has unticked boxes
- **THEN** the breakdown is shown and the three courses are offered before the landing question is
  asked, and before any git command runs

#### Scenario: A clean change is not interrupted

- **WHEN** every signal is clear
- **THEN** run 1 proceeds directly to the landing question with no extra prompt

#### Scenario: An unreadable worktree stops the run

- **WHEN** the script exits non-zero without printing a verdict line
- **THEN** the run stops and asks the operator, and integrates nothing

#### Scenario: A missing plan counts as outstanding

- **WHEN** the change's `openspec/changes/<name>/tasks.md` cannot be found
- **THEN** that signal counts as outstanding and the operator is prompted once

#### Scenario: Filing a task records the outstanding work and continues

- **WHEN** the operator chooses to file a tracker issue
- **THEN** an issue carrying the outstanding items is created, linked to the change's issue, and
  the run continues to the landing question
