## MODIFIED Requirements

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
