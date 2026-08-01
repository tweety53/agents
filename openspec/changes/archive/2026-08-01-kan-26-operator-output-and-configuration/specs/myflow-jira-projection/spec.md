## ADDED Requirements

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

**The search SHALL carry an explicit project constraint**, naming the project the follow-up would
otherwise be filed in: the key(s) the project's `## jira` configuration names, and with none named
there the project-key prefix of the change's linked issue. The tracker query searches whatever the
session's connection can reach, which on a multi-project site is every project it can query, so an
unconstrained search selects its write target from a population far larger than the one this
requirement describes — and every statement here about who can plant a candidate is scoped to
members of that project. Scoping the search to where the follow-up would be filed introduces no new
failure mode, a run that can determine no project key having had nowhere to file either.

The confirmation SHALL show the candidate's key, its title and its current status. It exists because
the search selects a **write target** by label, title and status, every one of which any member of
that project can set, over a scan deliberately unnarrowed within it — so the confirmation is the only
check that the issue the pipeline is about to append to, retitle and relabel is the intended one. A
declined confirmation SHALL NOT be treated as a failure: filing a new follow-up is the other correct
outcome, a duplicate being visible and mergeable where a write to the wrong issue is neither.

**The candidate's title is externally-authored text and SHALL be constrained before it is shown.**
Characters in Unicode categories `Cc` **and** `Cf` SHALL be folded to spaces — the second named
explicitly, because bidirectional overrides and zero-width joiners reorder or hide text without
being what a reader would call a control character. Whitespace runs SHALL then be collapsed, any run
of three or more of the fence delimiter characters SHALL then be folded below the length at which it
can delimit a block, and the result SHALL then be truncated to a bounded length. **That order is
itself a requirement:** the whitespace collapse SHALL follow the category fold, being defined over
the spaces that fold produces; the delimiter fold SHALL follow both, so that no step able to bring
characters together runs behind it; and only truncation — a suffix cut, able to shorten a run but
never to bridge two — SHALL follow the delimiter fold. As specified neither the category fold nor
the whitespace collapse deletes anything to zero width, so neither can manufacture a delimiter run
today; the ordering is what keeps that true of a later change to either step. The result SHALL be
rendered in a block of its own rather than interpolated into the question, so no run of it can be
read as the pipeline's own prose or as one of the offered options.

**Neutralising the fence delimiters is a requirement of the display block, not a refinement of it.**
The title is the sole content line of that block and the already-recorded count is printed after the
block closes, so a title that is itself a valid closing delimiter ends the block early, turns the
template's own closing delimiter into an opening one, and swallows the count and both answer options
into a dangling block — removing the one signal that makes forged already-recorded evidence visible
before the write.

The same file already bounds a summary-derived slug because it reaches a path; a string an operator
reads in order to authorise a write is bounded for the same reason.

**"Titled as a follow-up" SHALL mean an exact match on the two named title shapes**, after trimming
surrounding whitespace — never a substring or containment match. A tracker query operator that
matches on containment MAY narrow the search, but the exact comparison SHALL decide, so an issue
merely mentioning the words does not enter the candidate space.

The search SHALL NOT be narrowed by which change filed the candidate. A follow-up filed for a
different issue is a match, which is the point: outstanding work accumulates in one place rather
than in an issue per change.

**A search that fails is a third outcome and SHALL NOT be read as "no match".** On a failed search
the pipeline SHALL file nothing and SHALL emit one skipped-with-reason line naming the failure.
Reading it as "no match" files a new follow-up on every transient tracker failure, reintroducing
exactly the duplicate proliferation this requirement exists to prevent, and doing so silently
because a created issue looks like a success. The outstanding list still reaches the planning
commit's message and the handoff, so the durable record does not depend on the tracker.

**"A To Do status" means exactly two names — `To Do` and `TO DO URGENT`.** The set SHALL be
enumerated rather than derived from Jira's `statusCategory`, which groups a custom `TO DO URGENT`
with `In Progress` under `indeterminate`. An issue at any other status SHALL NOT be a join
candidate, and no question SHALL be asked about it — the run simply files a new follow-up.

**This does not reopen the unrecognised-status rule.** That rule governs *transitions*, and forbids
inferring a position in the four-name order for a status outside it, because such an inference
freezes the board for a whole change. A search filter performs no transition; the worst it can do is
miss a join candidate, after which a new follow-up is filed.

#### Scenario: An existing follow-up is joined after confirmation

- **WHEN** a follow-up is to be filed and an `AI-generated` follow-up already sits at `To Do`
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
- **THEN** it is treated as a To Do status and offered as the candidate
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
