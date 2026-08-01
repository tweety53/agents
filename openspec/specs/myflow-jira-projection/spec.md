# myflow-jira-projection Specification

## Purpose
TBD - created by archiving change kan-17-finish-gate-jira-and-commit-hygiene. Update Purpose after archive.
## Requirements
### Requirement: The pipeline projects its state onto a tracker issue at three points, and nowhere else

When a change has a linked issue, `/myflow-start` SHALL move it to **In Progress**,
`/myflow-finish` SHALL move it to **In Review** on its first run and to **Done** after the archive
is written. No other command SHALL transition the issue.

`/myflow-do` SHALL NOT transition the issue at any time, including on a fix run. It MAY write the
issue **description** when a fix adds scope, which is a separate concern from status.

The linked key SHALL live in the state file's `jiraIssue` field, SHALL be resolved only by
`/myflow-start`, and SHALL be carried forward verbatim by every other command.

#### Scenario: A fix round does not move the issue backwards

- **WHEN** `/myflow-do` runs for a change whose issue is already In Review
- **THEN** no transition is attempted
- **AND** the issue remains In Review

#### Scenario: A change with no linked issue makes no Jira call

- **WHEN** any command runs for a change whose `jiraIssue` is null
- **THEN** no Jira request of any kind is made

### Requirement: The issue reaches In Progress when planning begins, not when it ends

`/myflow-start` SHALL transition the issue to In Progress **immediately after the issue key is
resolved**, before brainstorming produces any artifact — so the board reflects the work while a
long planning run is in flight.

The transition SHALL remain non-blocking: a failure SHALL degrade to a single skipped-with-reason
line, and the run SHALL continue and write its state file exactly as it would have. The earlier
call position SHALL NOT make any state write depend on a Jira call succeeding.

A revision round SHALL find the issue already at or past In Progress and SHALL make no call.

#### Scenario: Planning begins and the board follows

- **WHEN** `/myflow-start` resolves a linked issue that is at To Do
- **THEN** the issue moves to In Progress before the first brainstorming question is asked

#### Scenario: A failed transition does not stop planning

- **WHEN** the transition call fails for any reason
- **THEN** one `⚠ Jira: skipped — <reason>` line is emitted
- **AND** brainstorming continues and the state file is written at the end of the run as normal

#### Scenario: A revision round makes no call

- **WHEN** `/myflow-start` is re-run for a change already at `STARTED` whose issue is In Progress
- **THEN** no transition is attempted and the handoff says the status was already correct

### Requirement: The issue reaches In Review when run 1 completes, whichever route was taken

`/myflow-finish` SHALL transition the issue to In Review at the end of a **successful** first run,
independently of how the branch landed — pull request, merge and push, or handled manually.

The transition SHALL NOT be conditioned on a pull request existing. Conditioning it on a PR is what
allowed a merge-and-push change to reach Done without ever passing through In Review.

A run 1 that stops before its chosen route completes — a rejected push, a merge conflict, a failed
PR creation — SHALL NOT transition the issue.

#### Scenario: Merge and push still reaches In Review

- **WHEN** run 1 completes by merging the branch and pushing, opening no pull request
- **THEN** the issue is transitioned to In Review

#### Scenario: The manual route still reaches In Review

- **WHEN** run 1 completes by pushing the branch only, leaving integration to the operator
- **THEN** the issue is transitioned to In Review

#### Scenario: A failed run 1 does not claim review

- **WHEN** run 1 stops because its push was rejected
- **THEN** no transition is attempted and the failure is reported with the command's own output

### Requirement: Transitions are forward-only, matched by name, and an unrecognised status is asked about

Transitions SHALL be resolved by reading the issue's available transitions and matching the target
**by name**, case-insensitively, allowing the usual spellings. A numeric transition identifier SHALL
NEVER be hardcoded.

The order SHALL be **four positions** — To Do, then In Progress, then In Review, then Done — and
each position SHALL be identified by the names mapped onto it. `TO DO URGENT` SHALL map to the To Do
position. An issue already at or past the target SHALL NOT be transitioned.

The mapping SHALL be by **enumerated name**, never by inference. In particular a position SHALL NOT
be derived from Jira's `statusCategory`, which groups a custom `TO DO URGENT` with `In Progress`
under `indeterminate` — an inference that would report such an issue as already at In Progress and
freeze the board at that status for the whole change. Enumerating `TO DO URGENT` at the To Do
position is the opposite operation: it states the position rather than deducing it, which is the same
mechanism this capability already uses for the follow-up join search's To Do set.

The names mapped onto the To Do position SHALL be stated **once** in
`skills/myflow-contracts/jira-integration.md`, and every other site needing that set SHALL cite it
rather than enumerate it a second time.

When the issue's **current** status maps to no position — matching none of the names, including the
mapped synonyms — its position in the order is undefined. The command SHALL NOT infer one and SHALL
instead show the operator the issue key, the current status and the intended target, and ask whether
to transition. Only an explicit yes SHALL transition it; anything else SHALL leave the status
untouched and emit one skipped-with-reason line.

This SHALL NOT increase the number of interactive Jira questions the pipeline may ask. The bound of
exactly two carve-outs — this ask and the join confirmation — is unchanged; what changes is that
fewer statuses reach this one.

#### Scenario: An urgent To Do transitions without a question

- **WHEN** `/myflow-start` resolves an issue whose current status is `TO DO URGENT` and the target is
  In Progress
- **THEN** the issue is transitioned to In Progress and no question is asked

#### Scenario: A custom status is asked about rather than guessed

- **WHEN** the issue's current status matches no name mapped onto any of the four positions
- **THEN** the operator is shown the key, the current status and the target, and asked whether to
  transition

#### Scenario: The position is not derived from the status category

- **WHEN** an issue sits at `TO DO URGENT`, whose `statusCategory` is `indeterminate`
- **THEN** its position is To Do because the name is enumerated there, not In Progress as the
  category would imply

#### Scenario: Declining leaves the status untouched

- **WHEN** the operator declines the transition for a status outside the mapping
- **THEN** no transition is made, one skipped-with-reason line is emitted, and the run continues

#### Scenario: A recognised status past the target is left alone

- **WHEN** the target is In Progress and the issue is already In Review
- **THEN** no transition is attempted and no question is asked

### Requirement: Issues the pipeline creates carry the parent issue's labels plus `AI-generated`

When a `/myflow-*` command creates a tracker issue, that issue SHALL be labelled with every label
carried by the change's linked issue, plus `AI-generated`.

No label SHALL be invented. The parent's labels exist by construction; `AI-generated` SHALL be
applied only because it is already in use in the project. When the change has no linked issue, the
created issue SHALL carry `AI-generated` alone.

The created issue SHALL be linked to the change's issue when one exists.

#### Scenario: A follow-up inherits its parent's routing label

- **WHEN** a command creates a follow-up issue during a change whose issue carries a routing label
- **THEN** the created issue carries that label and `AI-generated`

#### Scenario: No parent means no inherited labels

- **WHEN** a command creates an issue for a change with no linked issue
- **THEN** the created issue carries `AI-generated` and no other label

### Requirement: Jira is never a gate

No state write, commit, push, pull request, merge or archive SHALL depend on a Jira call
succeeding. Every failure path — no linked issue, integration unreachable, transition rejected,
target transition name not offered, issue not found — SHALL degrade to exactly one
`⚠ Jira: skipped — <reason>` line, after which the command SHALL continue and write its state
exactly as it would have.

A failure SHALL NOT be retried in a loop, SHALL NOT roll anything back, and SHALL NOT abort the run.

#### Scenario: An unreachable tracker does not stop an archive

- **WHEN** run 2 cannot reach the tracker to set Done
- **THEN** the archive is committed and pushed, the state is written `FINISHED`, and one
  skipped-with-reason line is reported

### Requirement: The issue description is appended to, never rewritten

`/myflow-start` and `/myflow-do` SHALL be the only commands that write the issue description, and
only when the run added scope the issue does not already describe. The write SHALL append a dated
bullet under an `## Added during implementation` heading, creating that heading once at the end of
the description when absent.

Because the underlying operation replaces the whole description, the payload SHALL be asserted
before every such write to contain the description just read as an **exact prefix**, byte for byte,
and to be **strictly longer** than it. If either assertion fails — including a read that was
truncated, elided, or returned in a form that cannot be reproduced verbatim — no write SHALL be made
and one skipped-with-reason line SHALL be emitted instead.

The pre-edit description SHALL be echoed verbatim into the handoff on any run that writes it, so the
transcript is the recovery path.

#### Scenario: A lossy read skips the write

- **WHEN** the existing description cannot be reproduced verbatim
- **THEN** no description write is made
- **AND** `⚠ Jira: description sync skipped — could not reproduce the existing description verbatim`
  is reported

#### Scenario: A run that added no scope writes nothing

- **WHEN** a run ends without having added scope the issue does not describe
- **THEN** the description is not written at all

### Requirement: A follow-up the pipeline files is named for the issue it came from

An issue any `/myflow-*` command files as a **follow-up** SHALL be titled `<KEY> follow-up`, where
`<KEY>` is the change's linked Jira issue. With no linked issue it SHALL be titled
`myflow follow-up`.

This requirement SHALL govern every site that files a follow-up. Today the only such site is finish
run 1's unfinished-work gate; a site added later inherits the naming rather than choosing its own.

Labelling is unchanged: the created issue carries the parent's labels plus `AI-generated`, or
`AI-generated` alone when no issue is linked.

#### Scenario: A follow-up is filed for a linked change

- **WHEN** the unfinished-work gate files a follow-up for a change linked to `KAN-31`
- **THEN** the created issue is titled `KAN-31 follow-up`

#### Scenario: A follow-up is filed for an unlinked change

- **WHEN** a follow-up is filed for a change with no linked issue
- **THEN** the created issue is titled `myflow follow-up`

### Requirement: An existing follow-up in To Do is joined rather than duplicated

Before creating a follow-up, the pipeline SHALL search the project for an existing pipeline-filed
follow-up and, on a match, SHALL **confirm the newest match with the operator** before writing to
it. Only an explicit confirmation SHALL join it; anything else SHALL file a new follow-up instead.
A follow-up matches when it carries the `AI-generated` label, is titled as a follow-up, and sits at
a To Do status.

**"A To Do status" means the names mapped onto the To Do position**, as stated once in
`skills/myflow-contracts/jira-integration.md` and required by
**Requirement: Transitions are forward-only, matched by name, and an unrecognised status is asked
about** above. The set SHALL be cited from that single statement rather than enumerated a second
time here, so the two sites cannot disagree. It SHALL NOT be derived from Jira's `statusCategory`.
An issue at any status outside that set SHALL NOT be a join candidate, and no question SHALL be
asked about it — the run simply files a new follow-up.

**This does not reopen the unrecognised-status rule.** That rule governs *transitions*, and forbids
inferring a position for a status outside the mapping, because such an inference freezes the board
for a whole change. A search filter performs no transition; the worst it can do is miss a join
candidate, after which a new follow-up is filed.

#### Scenario: The join set and the transition mapping cannot disagree

- **WHEN** the set of To Do names changes
- **THEN** both the transition mapping and the join search follow the single statement, because the
  join search cites it rather than carrying its own copy

#### Scenario: An existing follow-up is joined after confirmation

- **WHEN** a follow-up is to be filed and an `AI-generated` follow-up already sits at a To Do status
- **THEN** the operator is shown that issue's key, title and status and asked whether to add this
  run's items to it
- **AND** on an explicit yes the existing issue is joined
- **AND** no second issue is created

#### Scenario: A declined confirmation files a new follow-up

- **WHEN** the operator is shown a candidate and does not explicitly confirm it
- **THEN** the candidate is not written to at all
- **AND** a new follow-up is filed instead
- **AND** the run continues and writes its state exactly as it would have

#### Scenario: A session that cannot ask does not join

- **WHEN** a candidate matches but the session cannot put the question to the operator
- **THEN** the candidate is not joined
- **AND** a new follow-up is filed instead

#### Scenario: A failed search files nothing

- **WHEN** the search for an existing follow-up fails
- **THEN** no issue is created and no issue is joined
- **AND** one skipped-with-reason line names the search failure
- **AND** the run continues and writes its state exactly as it would have

#### Scenario: A follow-up from a different change is still joined

- **WHEN** the only open follow-up was filed for a different change's issue
- **THEN** it is offered as the candidate and joined on confirmation

#### Scenario: An urgent To Do counts

- **WHEN** the only candidate sits at `TO DO URGENT`
- **THEN** it is treated as a To Do status and offered as the candidate, because that name is mapped
  onto the To Do position
- **AND** no transition is performed on it

#### Scenario: A candidate past To Do is not joined

- **WHEN** the only follow-up in the project sits at `In Progress`
- **THEN** it is not a join candidate
- **AND** a new follow-up is filed without asking

#### Scenario: The newest of several is the candidate

- **WHEN** more than one follow-up matches
- **THEN** the newest is the one put to the operator

#### Scenario: A crafted title cannot impersonate the question

- **WHEN** the candidate's title contains newlines, terminal escape sequences, a bidirectional
  override or a zero-width joiner, or text shaped like one of the confirmation's own options
- **THEN** those characters are folded to spaces and the title is truncated to its bounded length
- **AND** it is shown in a block of its own, visibly separate from the question and its options

#### Scenario: A crafted title cannot close the block it is shown in

- **WHEN** the candidate's title is a run of the fence delimiter character long enough to close the
  display block
- **THEN** the run is folded below that length before the title is shown
- **AND** the already-recorded count and both answer options are still printed outside the block,
  where the operator can read them

#### Scenario: No earlier step can manufacture a delimiter run behind the fold

- **WHEN** the candidate's title carries two short delimiter runs separated only by whitespace, or by
  a zero-width formatting character
- **THEN** the separator is replaced with a space rather than deleted, so the two runs stay separate
  and no longer run is created for the delimiter fold to miss
- **AND** the delimiter fold still runs after both earlier steps, so any run long enough to delimit a
  block is folded before the title is shown

#### Scenario: A search without a project constraint is not performed

- **WHEN** the pipeline searches for an existing follow-up to join
- **THEN** the query names the project the follow-up would otherwise be filed in
- **AND** an issue in another project the connection can reach is not a candidate, whatever its
  label, title and status

#### Scenario: A title that merely contains the words is not a candidate

- **WHEN** the only `AI-generated` issue at `To Do` is titled so that a follow-up title appears
  inside a longer title
- **THEN** it does not match, because the comparison is on the exact title shapes
- **AND** a new follow-up is filed without asking about it

### Requirement: A join appends, and retitles the issue once it holds more than one source

A join SHALL append the new items to the joined issue's description under a dated `## From <KEY>`
heading, creating that heading if absent, and SHALL leave everything preceding it byte-for-byte
unchanged.

The write SHALL be governed by the same pre-write assertion every description write in this
capability carries: the payload contains the description just read as an exact prefix, and is
strictly longer than it. A failed assertion SHALL make no write at all and SHALL emit the standard
skipped-with-reason line.

**The read the assertion is made against SHALL be taken immediately before the write**, never one
taken earlier in the run. An earlier read may already be historical, and the write replaces the
whole field, so asserting against it destroys any intervening edit while reporting success. This
narrows the window and SHALL NOT be described as closing it: a compare-and-swap requires the tracker
to reject a write whose field moved underneath it, which the available tooling does not offer.

**A join SHALL NOT echo the joined issue's pre-edit description into the handoff.** That text was
authored by whoever could file the ticket and the pipeline never wrote a line of it, so reproducing
it verbatim places unreviewed third-party content among the run's own trusted output. The handoff
SHALL instead report the issue key, its status, any title change, and the section this run appended.
The tracker's own edit history is the recovery path for the pre-existing text.

**A join SHALL be idempotent under retry, and each of its three writes SHALL be guarded
separately.** Before appending, the pipeline SHALL check whether the joined issue already carries
this change's items and, if it carries all of them, SHALL skip the append. Without that check a
re-entered run 1 matches its own prior follow-up and appends a second identical section on every
retry.

**That check SHALL match on the change and the items, and SHALL NOT be scoped to the current date.**
The section heading carries the date it was written; matching only a section bearing today's date
makes a retry after a date rollover append a second, content-identical section — the duplicate the
check exists to prevent.

**The comparison SHALL be per item, by membership over the union of every `## From <KEY>` section
carrying this change's key — never equality of the whole list, and never the mere existence of such
a section.** Each item SHALL be normalised before comparison by stripping its list bullet, trimming,
and collapsing internal whitespace runs, and compared exactly thereafter. Both looser and stricter
readings fail, in opposite directions: comparing the whole list for equality appends everything
again whenever the outstanding list shifted between attempts, duplicating the items already there;
treating the existence of a section as a match silently drops the item that appeared since the last
attempt, which is the item the append exists for.

**A partial match SHALL be a defined outcome, not a rounding.** Where some of this change's items
are already recorded and others are not, the pipeline SHALL append **only the absent ones**, under a
new dated section rather than by editing an existing one, and SHALL report that it did so. Appending
into an existing section is not a pure suffix and would put the pre-write assertion's guarantee at
risk.

**A skipped append SHALL NOT skip the retitle or the label union.** Those two SHALL be re-attempted
whenever they are not yet satisfied — the retitle is self-guarding by title comparison and a label
union is idempotent — and the outcome SHALL continue to be reported as a **partial join** until both
are. Treating "the items are already there" as "write nothing" abandons a partial join permanently
and downgrades the warning an earlier run correctly emitted to an unmarked no-write line, which is
the invisibility the partial-join outcome exists to prevent. Only when all three writes are satisfied
SHALL the run report that nothing was written.

A join makes up to three writes and SHALL make them in this order: append the items, then retitle,
then union the labels. On the first join the issue's title SHALL become `myflow follow-up`, because
it no longer belongs to the single source its original title named; a title already reading
`myflow follow-up` SHALL be left as it is, **and so SHALL one naming the joining change's own key**,
which still names the only source the issue has and matches the search either way. The joined
issue's labels SHALL be unioned with the
labels the new follow-up would have carried — without this, an issue holding a second change's
outstanding work is invisible to anyone filtering by that change's labels.

The append goes first because the other two writes only make it findable. If the append fails, the
retitle and the label union SHALL NOT be attempted and the outcome is a failed join. If the append
succeeds and a later write fails, the outcome SHALL be reported as a **partial join**, naming which
write failed — never as a success, because a silently failed label union produces exactly the
invisibility the union exists to prevent.

**The partial-join report SHALL also name what the append did**, that being independent of which
later write failed: the whole list appended, only the absent ones appended, or the items already
recorded from an earlier attempt. Every combination of the two therefore has a reported form, and an
implementer meeting a partial append followed by a failed retitle composes no message of their own.

A failed creation, a failed join or a partial join SHALL degrade exactly as every other Jira write
does — one `⚠ Jira: skipped — <reason>` or partial-join line, with the run continuing and writing
its state as it would have.

#### Scenario: Items are appended under a dated heading

- **WHEN** an existing follow-up is joined
- **THEN** the new items appear under a dated `## From <KEY>` heading
- **AND** the text preceding that heading is unchanged

#### Scenario: A join that cannot be reproduced verbatim writes nothing

- **WHEN** the joined issue's description cannot be reproduced byte-for-byte as a prefix
- **THEN** no write is made
- **AND** one skipped-with-reason line is emitted

#### Scenario: The title becomes generic on the first join

- **WHEN** an issue titled `KAN-31 follow-up` is joined by a follow-up from `KAN-34`
- **THEN** its title becomes `myflow follow-up`

#### Scenario: An already-generic title is left alone

- **WHEN** an issue titled `myflow follow-up` is joined again
- **THEN** its title is unchanged

#### Scenario: A change does not retitle its own follow-up

- **WHEN** an issue titled `KAN-34 follow-up` is joined by a follow-up from `KAN-34`
- **THEN** its title is unchanged, no retitle call being made
- **AND** the join's remaining writes proceed as they otherwise would

#### Scenario: Labels are unioned

- **WHEN** a join carries a label the joined issue lacks
- **THEN** that label is added
- **AND** the labels it already carried are retained

#### Scenario: A failed join does not block the run

- **WHEN** the join is refused by the tracker
- **THEN** one skipped-with-reason line is emitted
- **AND** the command continues and writes its state exactly as it would have

#### Scenario: A partial join is not reported as a success

- **WHEN** the items are appended but the label union is refused
- **THEN** the outcome is reported as a partial join, naming the write that failed
- **AND** it is not reported as a completed join
- **AND** the command continues and writes its state exactly as it would have

#### Scenario: A failed append attempts nothing further

- **WHEN** the description append fails
- **THEN** neither the retitle nor the label union is attempted
- **AND** the outcome is a failed join

#### Scenario: A retried run does not append twice

- **WHEN** run 1 is re-entered and joins the follow-up its own previous attempt already appended to
- **AND** the items are the same, and the retitle and the label union are already satisfied
- **THEN** no write is made
- **AND** the run reports that the issue already carries this run's items

#### Scenario: A retry the next day still does not append twice

- **WHEN** run 1 is re-entered on a later date and the joined issue already carries this change's
  section with the same items under an earlier date
- **THEN** the append is skipped
- **AND** no second section is created for the new date

#### Scenario: A retry whose outstanding list gained an item appends only that item

- **WHEN** run 1 is re-entered and the joined issue already carries some but not all of this run's
  items
- **THEN** only the items not already recorded are appended, under a new dated section
- **AND** the already-recorded items are not written a second time
- **AND** the run reports that new items were added and the rest were already recorded

#### Scenario: A retry whose outstanding list shrank does not re-append

- **WHEN** run 1 is re-entered after the operator resolved some items, so this run's list is a
  subset of what the issue already carries
- **THEN** the append is skipped rather than repeated for the whole list

#### Scenario: A partial append followed by a failed retitle has a reported form

- **WHEN** a retry appends only the items not already recorded and the retitle is then refused
- **THEN** the outcome is reported as a partial join
- **AND** the report names both that new items were added while the rest were already recorded, and
  that the retitle failed

#### Scenario: A retry completes a partial join

- **WHEN** an earlier run appended the items but its label union was refused, and run 1 is
  re-entered
- **THEN** the append is skipped and the label union is attempted again
- **AND** the outcome is reported as a partial join until the union succeeds, never as an unmarked
  no-write

#### Scenario: The joined issue's existing description is not echoed

- **WHEN** a join writes to an issue filed by someone else
- **THEN** the handoff reports the key, the status, any title change and the appended section
- **AND** it does not reproduce the description that was there before the write
