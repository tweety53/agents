## MODIFIED Requirements

### Requirement: The handoff block is defined once and rendered by two commands

The shape of the handoff block for each state SHALL be defined in exactly one place —
`skills/myflow-contracts/handoff-blocks.md` — as a per-state template. Both the command that ends in
a state and `/myflow-status <name>` SHALL render the operator's handoff from that template.

**Only `/myflow-status` loads that file.** A producing command needs the one block it prints and
carries it in its own skill, citing the definition; the full set is needed only by the command that
may render any of them.

`/myflow-status <name>` SHALL regenerate the block from the state file and the artifacts on disk. It
SHALL NOT read back a stored copy of the text a previous run emitted, and no command SHALL store one.

**Each producing skill SHALL cite that definition at the block it carries.** A block sitting in a
skill with no citation is a second, independently authored definition however faithfully it matches
today, because nothing tells the next editor of the skill that the template exists. A skill's block
SHALL carry the same **folded lines**, the same fields within each folded line and the same order as
the template; it MAY enumerate, in place of a placeholder, the literal alternatives its own command
writes, and that enumeration SHALL stay within the value space the placeholder describes. A skill
listing fewer cases than the template describes narrows the contract and is a defect in the skill.

**Related fields SHALL be folded onto one line, and no field SHALL be dropped.** A folded line
carries several values under one label, separated by ` · `. Folding is a layout rule and never a
content rule: a value the block carried before the fold is carried after it, and the
missing-rather-than-dropped rule below applies to each value within a folded line exactly as it
applied to each value on its own line.

**A folded line SHALL group values of one kind — on-disk or run-only — and never both.** The
run-only rule is what makes this load-bearing: `/myflow-status` omits a run-only field and renders an
on-disk one, so a line mixing the two could be neither omitted nor rendered without misreporting
something. The `Jira` line and the pre-edit Jira description are therefore each their own line,
never folded into the on-disk `Recorded` line beside them.

**The pre-edit Jira description SHALL never be folded and never be dropped.** `editJiraIssue`
replaces the whole description field and there is no local backup, so the handoff transcript is the
recovery path; it is printed verbatim in a fenced block, on any run that wrote a description, and on
no other.

A stored copy would reproduce the original exactly and then go wrong silently the moment anything it
names moves — a worktree removed, an artifact republished, a PR opened. A regenerated block has no
copy to keep in step, which is why regeneration is the mechanism rather than an implementation
detail of it.

The regenerated block SHALL carry, for the change's current state:

| State | Contents |
|-------|----------|
| `STARTED` | the change name, the artifact URL, one folded line carrying the count of decisions recorded, **the count of open questions**, the planning effort and the three model choices, the main-checkout IntelliJ command, and `/myflow-do <name>` as the last line |
| `IN_PROGRESS`, run 1 not yet done | the change name, one folded line carrying the task progress and the git state, the worktree path, the test guide's path, the review command matching the git state, the worktree IntelliJ command, and `/myflow-finish <name>` as the last line |
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
`/myflow-status` omits them, per the run-only rule above. The block's **heading** is the one line the
rows do not itemise, because it is not a field: which heading the `IN_PROGRESS` renderings carry is
the selection rule stated below, not a value read from anywhere. Every other field is listed, the
change name included — a field a row omits reads as one the block need not carry, which is how a
constant that every rendering does carry came to be absent from all three rows. The table is
otherwise kept in step with the per-state templates it summarises: the templates are the definition,
and a field added to one is added to the other in the same change. Reading this table as the field
list to trim a template down to inverts that relationship and reopens the drift this requirement
exists to close.

`IN_PROGRESS` covers two different waits — a diff waiting on review, and a branch already handed off
and waiting on a merge. **The branch's merge status decides which rendering applies whenever that
status is known**, a merged branch being integrated by definition; `prUrl` is the discriminator only
where the merge status is inconclusive. The selection rule, and what the remaining one-way `prUrl`
limitation costs, are stated in
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and SHALL NOT be
restated here, so there is only one place to correct when they change.

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

#### Scenario: A producing command carries only its own block

- **WHEN** `/myflow-start` runs
- **THEN** it renders the `STARTED` block from the copy in its own skill, citing the definition
- **AND** it does not load `handoff-blocks.md`

#### Scenario: Folding drops no value

- **WHEN** a `STARTED` block is rendered for a change recording a planning effort and three models
- **THEN** the effort and all three models appear on the folded `Recorded` line
- **AND** a value the state file does not carry is reported as missing within that line rather than
  omitted from it

#### Scenario: A folded line does not mix on-disk and run-only values

- **WHEN** `/myflow-status <name>` regenerates a `STARTED` block
- **THEN** it renders the folded `Recorded` line in full
- **AND** it omits the `Jira` line and the pre-edit description line entirely, rather than rendering
  a partial folded line

### Requirement: Every pipeline command prints the tab commands at the start of its run

`/myflow-start`, `/myflow-do` and `/myflow-finish` SHALL print, immediately after their announcement
line, two commands for the operator to paste: `/rename <change-name>` and `/color` with a single
fixed colour.

They SHALL be printed at the **start** of the run rather than in the handoff, because labelling a
tab is useful before a long run rather than after it. This is a deliberate exception to nothing:
the handoff requirements govern what a command prints when it *ends*, and these lines are not part
of a handoff.

The colour SHALL be one fixed value for every command and every change, signifying only that a
pipeline command owns the tab, and SHALL be one the harness accepts.

The contract SHALL record why these are printed rather than invoked, so a later reader does not
treat the printing as an oversight to correct:

- the harness's `SlashCommand` tool exposes only commands of `type: "prompt"`, and both of these are
  `type: "local"` or `"local-jsx"`; and
- writing a terminal escape sequence directly is not available either, the harness providing no
  writable `/dev/tty` to a command.

`/myflow-status` SHALL NOT print them. A read-only report does not own the tab.

#### Scenario: A run opens with the two commands

- **WHEN** `/myflow-do` starts a run for a change
- **THEN** its announcement is followed by a `/rename` line carrying the change name and a `/color`
  line carrying the fixed colour

#### Scenario: The reason they are printed is recorded

- **WHEN** a reader asks why the commands are not invoked by the run itself
- **THEN** the contract states that the tool exposes only `type: "prompt"` commands and that no
  writable `/dev/tty` is available
- **AND** the answer does not depend on the session transcript in which it was discovered

#### Scenario: The read-only command does not print them

- **WHEN** `/myflow-status` runs
- **THEN** no `/rename` or `/color` line is printed

## ADDED Requirements

### Requirement: The status table reads the state file and local git, never the network

`/myflow-status` with no argument SHALL render its table from each change's state file plus local
git, and SHALL make no network call. The `gh pr list` probe SHALL NOT be performed.

The table SHALL carry exactly these columns, in this order:

| Column | Source |
|--------|--------|
| Change | the change name |
| Jira | `jiraIssue`, verbatim, or `—` |
| State | `state` |
| PR | the number parsed from the recorded `prUrl`, or `—` when it is `null` |
| Next | the next command, from the state and the merge status |
| Updated | `updatedAt` and `updatedBy` |

**The worktree path SHALL NOT be a column.** It is the widest value in the row and its branch
component is always `openspec/<change>`, so the column restated the change name at the cost of
wrapping every row. The absolute path SHALL still appear in the detail view rendered when a change
name is given, where the absolute-path requirement binds it.

**The `PR` column SHALL NOT report whether a pull request is open, merged or closed.** That answer
requires the network call this requirement removes, and reporting a state the command did not
observe is worse than reporting none. A change recording a `prUrl` SHALL show its number; the detail
view SHALL say where to look for its state.

**Merge status SHALL be retained**, because it is local git and it is what splits the `IN_PROGRESS`
row's next-command answer between integrating and archiving. Its three ordered steps, the rule that
an unresolved recorded merge base makes the answer inconclusive, and the multi-repo combination rule
are unchanged.

**`/myflow-status` SHALL perform no state-file write of any kind.** It reports what the file says. A
file that is missing or unparseable SHALL be named in the command's own output and skipped, never
rebuilt from inference.

#### Scenario: A status run makes no network call

- **WHEN** `/myflow-status` runs with no argument for a change recording a `prUrl`
- **THEN** no `gh` invocation and no other network call is made
- **AND** the PR column shows the recorded pull request's number

#### Scenario: The table has no worktree column

- **WHEN** `/myflow-status` renders its table
- **THEN** the row carries Change, Jira, State, PR, Next and Updated, and no worktree or branch column

#### Scenario: The detail view still gives the absolute worktree path

- **WHEN** `/myflow-status <name>` runs for a change with a recorded worktree
- **THEN** the detail view prints that worktree's absolute path

#### Scenario: An unreadable state file is named rather than rebuilt

- **WHEN** a file in the state directory cannot be parsed
- **THEN** the command names that file in its output and omits the change from the table
- **AND** it writes nothing to that file

#### Scenario: Merge status still splits the next command

- **WHEN** `/myflow-status` renders a row for a change at `IN_PROGRESS` whose branch is proven merged
- **THEN** the Next column says the next `/myflow-finish` will archive
