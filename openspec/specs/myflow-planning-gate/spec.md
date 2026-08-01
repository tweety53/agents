# myflow-planning-gate Specification

## Purpose
TBD - created by archiving change kan-17-finish-gate-jira-and-commit-hygiene. Update Purpose after archive.
## Requirements

### Requirement: An unresolved question is put to the operator, never resolved by assumption

When `/myflow-start` reaches a question its inputs do not answer — scope, decomposition, where a
requirement belongs, which of two behaviours was meant — it SHALL put that question to the operator
rather than pick an answer and proceed.

This SHALL hold at every planning effort level. A lower level MAY group questions into fewer rounds
and MAY batch related questions into one prompt; it SHALL NOT convert a question into an assumption.

Work that does **not** depend on the answer SHALL still be done rather than held behind it, and the
question SHALL be put at the point the answer is first needed.

#### Scenario: An ambiguous ask is asked about

- **WHEN** an issue's text can be read two ways that would produce materially different work
- **THEN** `/myflow-start` puts the readings to the operator rather than choosing one

#### Scenario: A low effort level still asks

- **WHEN** a change is planned at the lowest effort level and an unresolved question arises
- **THEN** the question is still put to the operator, grouped with others rather than dropped

#### Scenario: Independent work is not blocked on an answer

- **WHEN** a question blocks only part of the planning
- **THEN** the parts that do not depend on it are completed rather than deferred behind it

### Requirement: Every approval or choice is offered as options, not as open prose

Wherever a `/myflow-*` command asks the operator to approve something or to choose between courses
of action, it SHALL present named options to select from, with the recommended option named as such.

An open-ended question — of the form "does this look right, or is there something you want
changed?" — SHALL NOT be the mechanism by which an approval gate is passed.

Each option SHALL state what will happen if it is chosen, so the choice can be made without reading
the surrounding text.

#### Scenario: The design approval gate offers options

- **WHEN** `/myflow-start` presents a design for approval
- **THEN** the operator is offered named options, including approving as-is and requesting changes

#### Scenario: A recommendation is marked

- **WHEN** a command offers options and one of them is recommended
- **THEN** that option is identified as the recommendation rather than left for the operator to infer

#### Scenario: An open-ended prompt does not stand in for an approval

- **WHEN** a command needs approval to proceed
- **THEN** it does not substitute a free-text question for the options

### Requirement: Brainstorming iterates until no question is left holding

`/myflow-start`'s brainstorming stage SHALL iterate rather than pass once. After **every**
planning-stage exchange — a round of clarifying questions, the approval of a design section, and the
operator's review of the written spec — the command SHALL determine whether it now holds a question
its inputs do not answer. While it holds one, it SHALL open another round rather than proceed to
approaches, to the design, or out of the stage, **on the command's own judgment**. The stage SHALL
end while a question is still held only on an explicit operator answer — declining the round offer
named under **Rounds are softly bounded — later ones are offered, not taken** below, which records
what remains open rather than settling it — and never because this test itself concluded the stage
was done.

This SHALL be **one test applied after every exchange**, not a separate rule per gate. A design
section that reopens a question and an answer that opens a question are the same event, so a gate
added later inherits the loop instead of bypassing it by not being enumerated.

The test SHALL NOT be satisfied by the command recording a question **pre-emptively** — to dodge
asking one the operator has not seen; a question recorded that way is still held. A question the
operator has **explicitly deferred** — an "I cannot answer this" response at any round, or declining
the round offer named under **Rounds are softly bounded — later ones are offered, not taken**
below — stops counting as held once recorded under `## Open questions`, so the next test does not
reopen a round on the very question the operator just said they could not answer. Recording is what
such a deferred question receives, per the requirement below; it is not an alternative to asking one
the operator has not seen, and it is never a substitute the command reaches for on its own.

#### Scenario: An answer that opens a question opens a round

- **WHEN** an operator's answer in a question round leaves the command holding a question its inputs
  do not answer
- **THEN** another question round is opened rather than proceeding to approaches or to the design

#### Scenario: A design section that reopens a question routes back

- **WHEN** the operator's response to a design section raises a question rather than a correction
- **THEN** the command opens a question round rather than revising the design over the gap

#### Scenario: The spec review reopens the loop

- **WHEN** the operator's review of the written spec raises a question
- **THEN** the command opens a question round, and the spec is revised only once it is answered

#### Scenario: Scope added at a gate is treated as an exchange

- **WHEN** the operator adds scope at any planning gate and that scope leaves the command holding an
  unanswered question
- **THEN** the convergence test opens a round on it before the stage ends

#### Scenario: An explicitly deferred question stops counting as held

- **WHEN** the operator answers a question with an explicit "I cannot answer this" and it is recorded
  under `## Open questions`
- **THEN** the next convergence test does not reopen a round on that question
- **AND** a question the command has not yet put to the operator still counts as held, so a
  pre-emptive recording of that different question does not satisfy the test
### Requirement: The stage ends by confirmation, not by the command's own judgment

When the convergence test comes back empty, `/myflow-start` SHALL NOT silently proceed. It SHALL
state what it believes settled and ask the operator whether anything is still unclear, offering
named options with **moving on** as the recommended one, and SHALL end the stage at this prompt only
on an answer that chooses to move on. This confirm is one of two explicit operator answers that end
the stage — the other is declining the round offer under **Rounds are softly bounded — later ones
are offered, not taken** below, which ends the stage with a question still recorded open rather than
settled here.

Recommending *move on* SHALL be correct at this prompt precisely because the prompt is unreachable
while the command holds an unanswered question.

#### Scenario: The command confirms before leaving brainstorming

- **WHEN** the command holds no unanswered question
- **THEN** it states what it believes settled and asks whether anything is still unclear, rather
  than proceeding

#### Scenario: The operator names something at the confirm

- **WHEN** the operator answers the confirm by naming something still unclear
- **THEN** another question round is opened
### Requirement: Rounds are softly bounded — later ones are offered, not taken

Early rounds SHALL open without asking. From a threshold round onward, `/myflow-start` SHALL NOT
open a round silently: it SHALL show what is still open and what it would ask, and offer the round
as a named choice with **another round** as the recommended option.

Recommending *another round* SHALL be correct at this prompt because the offer is reachable only
while the command genuinely holds unanswered questions — the same reasoning that makes **Stop** the
recommendation at `/myflow-finish` run 1's unfinished-work gate.

The threshold SHALL be a tuned value stated in `skills/myflow-start/SKILL.md`, not in this
requirement, so it can move without amending the contract. There SHALL be no hard cap: no round
count SHALL end the stage on the command's own authority.

#### Scenario: An early round opens without a prompt

- **WHEN** the command opens a round below the threshold
- **THEN** it asks its questions without first offering the round as a choice

#### Scenario: A later round is offered with what is open

- **WHEN** the command would open a round at or beyond the threshold
- **THEN** it shows what is still open and what it would ask, and offers the round as a named choice
  with another round recommended

#### Scenario: No round count ends the stage by itself

- **WHEN** any number of rounds has been run and the operator has not chosen to move on
- **THEN** the stage does not end on the round count alone
### Requirement: A question left open is recorded in the change's design

A question still open when the brainstorming stage ends SHALL be recorded in the change's
`design.md` under `## Open questions`, beside `## Decisions`, one entry per question, carrying the
question, an immutable kebab-case **ID**, a **Status** of `open`, why it is open, and what it
affects.

The ID SHALL be assigned once and SHALL NOT change, because it is the match key a later round uses.
A later round that answers the question SHALL set that entry's status to `answered by <decision-id>`
and SHALL add the corresponding decision entry. An entry SHALL NEVER be deleted or rewritten once
recorded.

A stage that left nothing open SHALL record none, leaving the section empty rather than inventing
entries — the rule `## Decisions` already carries.

The recorded questions SHALL reach the published proposal artifact, so that stopping with something
open is visible at the gate the operator reads rather than held only in the session transcript.

#### Scenario: A deferred question is recorded

- **WHEN** the operator chooses to move on with a question unanswered
- **THEN** it is recorded under `## Open questions` in `design.md` with an ID and a status of `open`,
  and appears in the published proposal artifact

#### Scenario: A later round answers an open question

- **WHEN** a revision round answers a question recorded as open
- **THEN** that entry's status becomes `answered by <decision-id>`, a decision entry is added, and
  the original entry is left in place

#### Scenario: Nothing open records nothing

- **WHEN** the stage ends with no open question
- **THEN** the `## Open questions` section is empty rather than carrying invented entries
### Requirement: A revision round re-enters the loop, scoped to what was reopened

Re-running `/myflow-start` for a change already at `STARTED` SHALL apply the convergence test to
what the operator's feedback reopened, and to whatever that opens in turn. Settled parts of the plan
SHALL NOT be re-brainstormed.

Scoping SHALL be what makes the loop usable on a revision: a revision round exists because most of
the proposal was right, and re-asking answered questions would make the cheap path expensive enough
that operators stop taking it.

#### Scenario: A revision round asks about what was reopened

- **WHEN** a revision round's feedback leaves the command holding an unanswered question
- **THEN** a question round is opened on it before the artifacts are revised

#### Scenario: A revision round does not re-ask settled questions

- **WHEN** a revision round revises one part of the plan
- **THEN** questions already answered for the untouched parts are not asked again
### Requirement: Every planning effort level runs the loop

The convergence test, the confirm, the soft bound and the open-questions record SHALL apply at every
planning effort level. A level MAY change how many questions are grouped into a round — one at a
time at `detailed`, batched at `low` — and SHALL NOT change whether another round opens.

A level able to end the loop early would be a way to skip the gate rather than a way to size the
thinking inside it, which `myflow-planning-effort` already forbids.

#### Scenario: The lowest level still converges

- **WHEN** a change planned at `low` reaches the end of a question round holding an unanswered
  question
- **THEN** another round is opened, with questions batched rather than dropped

#### Scenario: The highest level converges without extra gates

- **WHEN** a change is planned at `detailed`
- **THEN** questions are put one at a time across more rounds, and the same confirm ends the stage
