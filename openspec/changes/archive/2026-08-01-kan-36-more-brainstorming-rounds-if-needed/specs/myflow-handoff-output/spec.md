## MODIFIED Requirements

### Requirement: The handoff block is defined once and rendered by two commands

The shape of the handoff block for each state SHALL be defined in exactly one place —
`skills/myflow-contracts/pipeline.md` — as a per-state template. Both the command that ends in a
state and `/myflow-status <name>` SHALL render the operator's handoff from that template.

`/myflow-status <name>` SHALL regenerate the block from the state file and the artifacts on disk. It
SHALL NOT read back a stored copy of the text a previous run emitted, and no command SHALL store one.

**Each producing skill SHALL cite that definition at the block it carries.** A block sitting in a
skill with no citation is a second, independently authored definition however faithfully it matches
today, because nothing tells the next editor of the skill that the template exists. A skill's block
SHALL carry the same labels, the same fields and the same order as the template; it MAY enumerate,
in place of a placeholder, the literal alternatives its own command writes, and that enumeration
SHALL stay within the value space the placeholder describes. A skill listing fewer cases than the
template describes narrows the contract and is a defect in the skill.

A stored copy would reproduce the original exactly and then go wrong silently the moment anything it
names moves — a worktree removed, an artifact republished, a PR opened. A regenerated block has no
copy to keep in step, which is why regeneration is the mechanism rather than an implementation
detail of it.

The regenerated block SHALL carry, for the change's current state:

| State | Contents |
|-------|----------|
| `STARTED` | the change name, the artifact URL, the count of decisions recorded, **the count of open questions**, the planning-effort line, the three model choices, the main-checkout IntelliJ command, and `/myflow-do <name>` as the last line |
| `IN_PROGRESS`, run 1 not yet done | the change name, the task progress, the git state, the worktree path, the test guide's path, the review command matching the git state, the worktree IntelliJ command, and `/myflow-finish <name>` as the last line |
| `IN_PROGRESS`, run 1 done | the change name, the pull request URL, and `/myflow-finish <name>` as the last line |
| `FINISHED` | nothing — `FINISHED` changes are omitted from this report, as they already are |

The open-questions count SHALL be derived from the change's `design.md` — the entries under
`## Open questions` whose status is still `open` — and SHALL therefore be **regenerable rather than
run-only**: it is read from an artifact on disk, exactly as the decisions count is, so
`/myflow-status <name>` renders it rather than omitting it. It SHALL read `none` when no question is
open, by the same missing-rather-than-dropped rule the other fields follow.

**This table lists what a regenerated block carries, which is the template's fields minus its
run-only ones** — the panel roster and the pre-edit Jira description on the `IN_PROGRESS` review
rendering, the Jira line and the pre-edit description on `STARTED`, and the landing route and
outstanding list on the run-1 rendering. The command that ends in the state prints those too;
`/myflow-status` omits them,
per the run-only rule above. The block's **heading** is the one line the rows do not itemise,
because it is not a field: which heading the `IN_PROGRESS` renderings carry is the selection rule
stated below, not a value read from anywhere. Every other field is listed, the change name
included — a field a row omits reads as one the block need not carry, which is how a constant that
every rendering does carry came to be absent from all three rows. The table is otherwise kept in
step with the per-state templates it summarises: the templates are the definition, and a field added
to one is added to the other in the same change. Reading this table as the field list to trim a
template down to inverts that relationship and reopens the drift this requirement exists to close.

`IN_PROGRESS` covers two different waits — a diff waiting on review, and a branch already handed off
and waiting on a merge. **The branch's merge status decides which rendering applies whenever that
status is known**, a merged branch being integrated by definition; `prUrl` is the discriminator only
where the merge status is inconclusive. The selection rule, and what the remaining one-way `prUrl`
limitation costs, are stated in
**The block each state renders** (`skills/myflow-contracts/pipeline.md`) and SHALL NOT be restated
here, so there is only one place to correct when they change.

#### Scenario: Re-running status reproduces the handoff

- **WHEN** `/myflow-status <name>` runs for a change at `IN_PROGRESS` whose work is staged and
  uncommitted, so `HEAD` is still the merge base recorded for its worktree, and which records no
  `prUrl`
- **THEN** it prints the worktree path, the guide path and the staged-diff command
- **AND** its last line is `/myflow-finish <name>`

#### Scenario: A change waiting on a merge regenerates the shorter block

- **WHEN** `/myflow-status <name>` runs for a change at `IN_PROGRESS` whose branch is not merged and
  which records a `prUrl`
- **THEN** it prints the pull request URL
- **AND** it does not print the worktree path, the guide path or the staged-diff command
- **AND** its last line is `/myflow-finish <name>`

#### Scenario: A manually landed branch is not shown as a diff waiting on review

- **WHEN** `/myflow-status <name>` runs for a change whose run 1 took the *handle it manually*
  route, so its branch carries commits of its own, is proven not merged, and records no `prUrl`
- **THEN** the regenerated block is the one for a branch waiting on the merge
- **AND** it is not the review-and-test one, whose heading would describe committed and pushed work
  as staged

#### Scenario: One invocation does not contradict itself about the merge

- **WHEN** `/myflow-status <name>` runs for a change left at `IN_PROGRESS` by a run 2 that stopped
  on a cleanup leftover, so its branch is merged
- **THEN** the table reports that the next `/myflow-finish` will archive
- **AND** the regenerated block reports the branch as merged rather than as waiting on the merge
- **AND** neither is derived from `prUrl` while the merge status is known

#### Scenario: The merge-and-push route does not tell the operator to wait for its own merge

- **WHEN** `/myflow-finish` run 1 takes the merge-and-push route, which merges into the base branch
  within that run
- **THEN** the heading it prints is the merged-and-waiting-on-run-2 alternative
- **AND** it does not say the branch is waiting on a merge beside a route line reading *merged and
  pushed*

#### Scenario: A branch with no commits of its own is not reported as merged

- **WHEN** `/myflow-status <name>` runs for a change at `IN_PROGRESS` whose work is staged but never
  committed, so `HEAD` is still the merge base recorded for its worktree
- **THEN** the merge status reads *not merged*, whatever the ancestor test would return
- **AND** the regenerated block is the one for a diff waiting on review, not the one saying the
  branch was merged and is waiting on run 2

#### Scenario: An unresolvable recorded merge base is inconclusive, not merged

- **WHEN** the merge base recorded for a worktree no longer resolves there
- **THEN** the merge status reads *inconclusive* and no ancestor test decides it
- **AND** the branch is not reported as merged

#### Scenario: Two worktrees that disagree do not yield one worktree's answer

- **WHEN** a change records two worktrees and one is proven merged while the other is not
- **THEN** the change's merge status reads *not merged*
- **AND** the detail view names which worktree gave which answer

#### Scenario: A committed branch is not sent a staged-diff command

- **WHEN** the regenerated block is the one for a diff waiting on review, and the change's work is
  already committed and pushed
- **THEN** the git line names that state
- **AND** the review command it prints is one that shows the committed range, not one that prints
  nothing

#### Scenario: A skill's block cites the template

- **WHEN** a contributor reads the handoff block in `skills/myflow-do/SKILL.md`
- **THEN** it names the per-state template as the definition
- **AND** the labels and fields it carries are the template's

#### Scenario: A run-only value is omitted rather than named missing

- **WHEN** a regenerated block would carry a value only the run that emitted it could know — such as
  the landing route, the outstanding list run 1 integrated over, or the panel roster that run
  selected
- **THEN** the block omits that line
- **AND** it does not report the value as missing

#### Scenario: The Jira transition is not regenerated

- **WHEN** `/myflow-status <name>` regenerates the `STARTED` block for a change linked to an issue
- **THEN** it omits the Jira line rather than reporting it missing or naming a transition
- **AND** it makes no Jira call to recover one, the state file recording the issue key and no
  transition history

#### Scenario: The panel roster is not regenerated from a record that may not exist

- **WHEN** `/myflow-status <name>` regenerates the review rendering for a change whose worktree has
  been removed
- **THEN** it omits the panel line rather than reporting it missing
- **AND** it does not depend on the panel record, which is gitignored and may legitimately be absent

#### Scenario: Nothing is stored to go stale

- **WHEN** a change's worktree has been removed since the run that emitted the original handoff
- **THEN** the regenerated block reflects the state as it now stands
- **AND** no stored copy of the earlier text is printed

#### Scenario: A missing value is named

- **WHEN** a change at `STARTED` has no recorded artifact URL
- **THEN** the block reports the artifact URL as missing
- **AND** the line is not silently dropped

#### Scenario: The table form is unchanged

- **WHEN** `/myflow-status` runs with no change name
- **THEN** it prints the table
- **AND** it does not regenerate a handoff block for each change

#### Scenario: Regeneration performs no action

- **WHEN** the regenerated block names a command to run
- **THEN** `/myflow-status` prints it
- **AND** does not execute it, stage anything, or write any state

#### Scenario: The STARTED block reports open questions

- **WHEN** `/myflow-start` ends a run whose `design.md` records two questions still at `open`
- **THEN** its handoff carries an open-questions line reading `2`, immediately after the decisions
  count

#### Scenario: A change with nothing open reports none

- **WHEN** the `## Open questions` section is empty
- **THEN** the open-questions line reads `none` rather than being omitted

#### Scenario: The count is regenerated rather than remembered

- **WHEN** `/myflow-status <name>` regenerates the `STARTED` block for a change whose open questions
  were answered by a later revision round
- **THEN** the count reflects the entries still at `open` in `design.md` as it now stands, not the
  count the original run printed
