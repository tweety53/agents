# myflow-contract-economy Specification

## Purpose
How a contract file under `skills/myflow-contracts/` is partitioned into a normative core that
every `/myflow-*` command loads and a rationale appendix that none of them do, what distinguishes
the two, the edits a move may make, and the byte budget that keeps the cores from growing back.
## Requirements
### Requirement: A contract file separates its normative core from its rationale

A file the pipeline loads on a run — a contract under `skills/myflow-contracts/`, or a
`SKILL.md` a `/myflow-*` command reads — that carries both rules and the reasoning behind them SHALL
be split into two files: a **core** at the original path, and a **rationale appendix** beside it,
named by appending `-rationale` before the extension.

**The rule is about what a run loads, not about which directory the file sits in.** A
`SKILL.md` is read in full on every invocation of its command exactly as a contract is, carries the
same rule-plus-justification prose, and costs the same per run. A rule scoped to one directory would
have left the largest single per-command file — `skills/myflow-do/SKILL.md` — outside it for no
reason other than where it lives.

An appendix beside a `SKILL.md` does not change how skills resolve: the harness loads `SKILL.md` and
nothing else, and a skill directory may already hold other files.

One test decides which file a passage belongs in, applied to sentence groups — a paragraph, a
bullet, a table — and never to clauses inside a sentence:

> A passage is **core** if removing it would change what an agent does. A passage is **rationale**
> if removing it changes no agent behaviour.

Core therefore holds tables, ordered step lists, handoff templates, verdict tables, code blocks an
agent runs, and every prohibition. Rationale holds justification of an ordering, alternatives
considered and rejected, history such as "this reverses an earlier decision", measurements, and
meta-commentary about the document itself.

**The test is behavioural, not typographic.** A prose paragraph stating a rule is core.

#### Scenario: A rule stated in prose stays in the core

- **WHEN** a paragraph states a prohibition in prose rather than in a table
- **THEN** that paragraph is in the core file
- **AND** it is not moved to the appendix on the grounds of being prose

#### Scenario: A justification moves to the appendix

- **WHEN** a paragraph explains why one step is ordered before another, and removing it would leave
  the ordering itself still stated
- **THEN** that paragraph is in the appendix

#### Scenario: The appendix path is derived from the core path

- **WHEN** `skills/myflow-contracts/<name>.md` is split
- **THEN** its appendix is at `skills/myflow-contracts/<name>-rationale.md`

#### Scenario: A skill file splits the same way

- **WHEN** `skills/<skill>/SKILL.md` is split
- **THEN** its appendix is at `skills/<skill>/SKILL-rationale.md`
- **AND** the harness still loads `SKILL.md` alone

### Requirement: The split is a verbatim partition, with citation repointing as its only edit

Every sentence of the original file SHALL land in exactly one of the two resulting files,
byte-for-byte unchanged. No passage SHALL be summarised, reworded, merged, or authored afresh.

**Exactly two edits are permitted to a moved passage**, and no others:

1. **Repointing a citation** whose target moved. A backticked path, and the bold section token beside
   it, MAY change to name the file that now holds the target.
2. **Deleting a position word** — `above` or `below` — that the move made false, because what it
   locates now sits in a different file.

Every other character of a moved passage SHALL be unchanged.

**The second edit SHALL be a deletion, never a substitution.** The word is removed; it is not
replaced with a filename or any other phrasing. A deletion cannot alter a rule — it removes a
locator that has become false, leaving the citation's path to locate the target, which is the
convention a cross-file citation in this corpus already follows.

**A passage that another passage depends on by reference is classified with that passage.** Where a
core passage names a term, a count, or a fact that only a moved passage defines, the moved passage
SHALL stay in the core: an appendix is never loaded at runtime, so a core reader would otherwise meet
an undefined term. This applies symmetrically — a passage that refers to its neighbour by position
stays with that neighbour.

**One consequence of the citation edit is stated rather than left to be discovered: where
repointing changes the length of the bold token, the paragraph containing it MAY be re-wrapped.**
The corpus is hard-wrapped, so a longer token would otherwise leave an over-long line. The re-wrap
is confined to the paragraph holding the repointed citation, and its lines are accounted for in the
line-multiset check exactly as the repointed line itself is — a `<`/`>` pair whose two sides differ
only by wrapping.

**This is not a licence to re-wrap anything else.** The lines around a *moved* passage are never
re-wrapped; that rule is what the paragraph-granularity requirement rests on, and this does not
touch it. What is permitted here is only the reflow the citation edit itself forces on the paragraph
it edits.

**A mixed passage — one carrying both a rule and the argument for it — MAY be handled by rule
extraction, under three conditions and no others.** A mixed passage cannot be classified by the
core/rationale test without either keeping its argument in the core or losing its rule to the
appendix, and this is the bounded third answer:

1. The **original passage SHALL move to the appendix byte-for-byte unchanged**, subject to the same
   two permitted edits above and no more. Every sentence of the original file therefore still lands
   in exactly one resulting file, unmodified.
2. The **core SHALL gain a statement of the rule** the original carried, together with whatever
   operative detail the rule cannot be followed without, and a citation to the appendix heading now
   holding the argument. It SHALL NOT restate the argument.

   **The bound is on what kind of content stays, not on a sentence count.** A rule an agent cannot
   act on without knowing its mechanism keeps that mechanism in the core, by the same behavioural
   test that classifies any other passage: if removing it would change what an agent does, it is
   core. A count would fail exactly where it is most load-bearing — "both commits are guarded" is
   not actionable without "each commit is preceded by a staged-changes test, and the whole sequence
   is one `&&` chain" — and would push the operative half into a file no run loads, which is the
   defect this whole requirement exists to prevent. What the core may not gain is a second telling
   of *why*: that is the drift the extraction bounds.
3. What the core gained SHALL be **quoted in the per-move ledger** alongside the appendix passage it
   replaces, so it is checked against that passage rather than taken on trust.

**Where the rule survives verbatim in the core, the appendix keeps the argument alone.** Condition 1
sends the original to the appendix unchanged so that no wording is lost; where the core's extracted
statement already carries those sentences word for word, keeping them in the appendix too stores one
rule in two independently editable places, which is the drift this whole requirement exists to
prevent. The appendix then begins at the argument. **The test is whether any wording would be lost:**
if the core carries every sentence the appendix would have restated, trimming loses nothing and
removes a copy; if it does not, condition 1 governs and the original moves whole.

#### Scenario: A duplicated rule sentence is not kept in both files

- **WHEN** rule extraction leaves the core carrying the original's rule sentences verbatim
- **THEN** the appendix carries only the argument, beginning after those sentences
- **AND** no sentence present in the core is repeated in the appendix

**Rule extraction is not permitted where the whole passage is core**, and SHALL NOT be used to
shorten a passage that carries no argument. Extraction answers a mixed passage; a wholly normative
passage stays in the core whole, and a wholly rationale passage moves whole.

**A passage removed from the loaded corpus entirely — into a document no `/myflow-*` command
reads — is outside this requirement rather than an exception to it.** This requirement governs the
partition of a contract into a core and an appendix, both of which remain in the corpus. Authoring
fresh prose in a file no command loads is not a partition of a contract, and the per-move ledger
records such a passage with its destination document named.

#### Scenario: A stale position word is deleted rather than reworded

- **WHEN** a passage moves to another file and a citation in it says `above`
- **THEN** the word `above` is deleted
- **AND** it is not replaced with the name of the file that now holds the target

#### Scenario: A definition stays with the term that uses it

- **WHEN** a core passage uses a term whose only definition is in a passage classified as rationale
- **THEN** that defining passage stays in the core

#### Scenario: A mixed passage is extracted rather than misclassified

- **WHEN** a passage states a prohibition and then argues for it across three paragraphs
- **THEN** the whole passage appears in the appendix byte-for-byte
- **AND** the core carries the prohibition, any operative detail it cannot be followed without, and a
  citation to that appendix heading
- **AND** the core carries no restatement of the argument
- **AND** the ledger row for that passage quotes what the core gained

#### Scenario: A wholly normative passage is not extracted

- **WHEN** a passage carries a rule and no argument for it
- **THEN** it stays in the core unchanged
- **AND** no shorter sentence is substituted for it

A rewritten core would produce two independently authored statements of one rule — the failure a
contract split exists to avoid — so the constraint is met by construction wherever extraction is not
used, and by the ledger where it is.

### Requirement: An appendix mirrors its core's heading tree

An appendix SHALL carry the same `##` and `###` headings as its core, in the same order. Each
appendix section SHALL hold only the reasoning belonging to the core section of the same name.

A section that is wholly normative SHALL leave its appendix heading **present with no body**,
rather than absent — the same missing-rather-than-dropped rule this pipeline applies to handoff
fields, and what makes it visible that a section was examined rather than skipped.

Mirroring is what lets a citation resolve against whichever file its path names, and lets an editor
read core and appendix side by side.

#### Scenario: Every core heading has an appendix counterpart

- **WHEN** the `##` and `###` headings of a core file are compared with those of its appendix
- **THEN** the two sequences are identical

#### Scenario: A wholly normative section leaves an empty appendix heading

- **WHEN** a core section carries no rationale at all
- **THEN** its heading is present in the appendix with no body beneath it

### Requirement: A `/myflow-*` run never loads a rationale appendix

A rationale appendix exists for whoever edits the file it belongs to. No `/myflow-*` command SHALL
load one as part of a run, and no skill SHALL instruct an agent to load one. This covers a contract's
appendix and a skill's appendix alike.

This rule is **stated and deliberately not enforced**, the same treatment this corpus gives its
other judgment rules. A naming-convention guard was considered and rejected as machinery around a
rule no observed failure has broken.

#### Scenario: The rule is stated where a command's agent will read it

- **WHEN** `skills/myflow-contracts/SKILL.md` is read
- **THEN** it states that a `/myflow-*` run never loads an appendix, and names appendices as
  editor material

#### Scenario: No skill instructs an appendix load

- **WHEN** the `/myflow-*` skill files are searched for a `-rationale.md` path in a load instruction
- **THEN** none is found

#### Scenario: A skill does not load its own appendix

- **WHEN** `skills/<skill>/SKILL.md` is read
- **THEN** it does not instruct the run to read `SKILL-rationale.md`

### Requirement: A section reachable from only one command lives in its own file

Where a contract section's rules are reachable from exactly one command, that section SHALL be
moved whole — its rules **and** its reasoning together — into a file of its own, so that every
other command stops loading it.

**Such a file SHALL NOT be core/appendix split merely because it is large.** Splitting buys a saving
on only the one command that loads it, and costs that command a second file to read.

**It SHALL be split where its reasoning is separable from its rules** — where a passage can be
lifted whole, leaving behind a rule an agent can still act on, without the surrounding procedure
losing anything it needs. That is the test, and it is the same behavioural one that governs every
other split: a passage is core if removing it would change what an agent does.

**The test is separability, never the loading.** No appendix in this corpus is loaded by any run —
that is the base invariant of an appendix — so "the appendix would not be loaded" is true of every
conceivable split and cannot decide one. Stating it as the test would license splitting a file this
same requirement forbids splitting, two sentences apart. What the never-loaded property does is make
the saving one-sided once separability is established: the "extra file" the previous rule declines to
pay for is not paid, because the command reads only the smaller core.

`### Follow-up issues` in `jira-integration.md` is reachable only from `/myflow-finish` run 1 and is
457 lines / 35467 bytes — 67% of that file — so `/myflow-start` and `/myflow-do` stop loading it. It
stays whole because its reasoning is **not** separable: the why and the what sit in the same
paragraphs, sentence by sentence, with no line a passage could be lifted along.

`handoff-blocks.md` is reachable only from `/myflow-status` and **is** split, because its design
history — why regeneration beats storage, why the open-questions count is not run-only, why the
merge-base pre-check precedes the ancestor test — is not needed to render a block, and
`/myflow-status` pays for it on every detail view.

#### Scenario: Follow-up issues moves to its own file

- **WHEN** `skills/myflow-contracts/jira-integration.md` is read after this change
- **THEN** it does not contain the `Follow-up issues` section
- **AND** `skills/myflow-contracts/jira-followups.md` contains it, with its reasoning inline

#### Scenario: A single-command file whose reasoning is interleaved is not split further

- **WHEN** `skills/myflow-contracts/jira-followups.md` is read
- **THEN** it carries both its rules and its reasoning
- **AND** no `jira-followups-rationale.md` exists

#### Scenario: A single-command file whose reasoning is separable is split

- **WHEN** `skills/myflow-contracts/handoff-blocks.md` is read
- **THEN** it carries the per-state templates and the rules for rendering them
- **AND** its design history lives in `skills/myflow-contracts/handoff-blocks-rationale.md`
- **AND** no `/myflow-*` run loads that appendix

#### Scenario: A start run does not load the follow-up contract

- **WHEN** `/myflow-start` runs
- **THEN** it loads `jira-integration.md` and not `jira-followups.md`

### Requirement: A passage another command depends on is hoisted before a single-command move

Before a section is moved out on the grounds that one command needs it, every citation **into** that
section SHALL be resolved, and any passage a **different** command genuinely depends on SHALL be
hoisted into the core first.

A citation is a real dependency when the citing file is loaded by a command that would not load the
destination. A citation whose only role is to point at reasoning is not.

**A guard cannot answer this question.** `scripts/check-references.sh` resolves a citation whenever a
heading of that name exists in the named file, and the heading-mirroring rule guarantees one does —
so the guard is green whether or not the citation still reaches its substance. The resolution is
therefore made by reading each citation, and the change that moves a section SHALL name them
individually rather than asserting that the guard passed.

#### Scenario: A passage two loaded contracts cite is not moved out

- **WHEN** a section is a candidate for a single-command file
- **AND** a contract loaded by a different command cites into that section
- **THEN** the cited passage stays in the core

#### Scenario: Each citation into a moved section is checked by reading

- **WHEN** a section is moved to a single-command file
- **THEN** every citation into it is examined individually
- **AND** a green `check-references.sh` is not offered as evidence that they still resolve to their
  substance

### Requirement: A move or eviction is recorded in a per-move ledger

Any task that moves, evicts or extracts prose from a file the pipeline loads SHALL emit a ledger
table in its own output, before its diff reaches the review panel. The table SHALL carry one row per
removed passage, with these four columns:

| Column | Contents |
|--------|----------|
| Removed passage | the passage's first eight words, verbatim |
| Source heading | the heading it sat under in the original file |
| Destination | the file and heading now holding it, or `— (rule extracted)` with what the core gained quoted, or `— (rewritten for <document>)` naming the document |
| Pointer left | what the core carries naming the rule, or `none` where the passage carried no rule |

The review panel SHALL check the table against the diff **in both directions**: every deletion in
the diff SHALL have a row, and every row's destination SHALL exist. A ledger row whose destination
cannot be found, and a diff deletion with no row, SHALL each be a finding.

**The ledger exists because the reference guard cannot see this change's failure mode.**
`scripts/check-references.sh` resolves a citation whenever a heading of that name exists, so a stub
carrying a heading and a pointer resolves identically to a stub carrying a heading and nothing. The
ledger is what makes the loss question answerable from an artifact rather than from a reviewer's
attention.

**A mechanical no-loss check SHALL NOT be substituted for the ledger.** Asserting that every removed
line survives somewhere is exact for a pure move and fails on rule extraction, which authors a line
that did not previously exist; the similarity threshold such a check would need cannot be set
without guessing.

#### Scenario: A pure move is recorded

- **WHEN** a rationale paragraph moves to an appendix unchanged
- **THEN** its ledger row names the appendix file and the heading it now sits under
- **AND** the Pointer left column names the sentence left in the core, or `none`

#### Scenario: An extraction quotes its replacement

- **WHEN** a mixed passage is handled by rule extraction
- **THEN** its ledger row reads `— (rule extracted)` and quotes the new core sentence

#### Scenario: A deletion with no ledger row is a finding

- **WHEN** the diff removes a passage that appears in no ledger row
- **THEN** the review panel records a finding

#### Scenario: A ledger row whose destination is absent is a finding

- **WHEN** a ledger row names an appendix heading that does not exist in the appendix
- **THEN** the review panel records a finding

### Requirement: Repository prose states a thing once

`rules/be-brief.mdc` SHALL carry a clause governing written prose, alongside the clause it already
carries for what an agent says. The clause SHALL name exactly two subjects:

- **this repository's own Markdown**, named explicitly rather than generalised to any repository's
  documentation, because `be-brief.mdc` installs globally into every project's `CLAUDE.md` and a
  general clause would govern projects it was never evaluated against;
- **the artifacts a `/myflow-*` run generates, in any project** — `proposal.md`, `design.md`,
  `tasks.md`, delta specs, panel records and self-review reports — because those are written by this
  repository's own pipeline wherever they land.

Another project's pre-existing documentation SHALL NOT be subject to the clause.

The clause SHALL state brevity as **non-repetition**: a file states a thing once and does not
restate what another file states canonically. `be-brief.mdc`'s existing sentence exempting written
files SHALL be reworded to state **completeness** — every requirement, scenario, exit code and
worked example present — so the two are readable as the separate properties they are. A file SHALL
be required to satisfy both, and the rule's text SHALL NOT leave them reading as alternatives.

The clause SHALL permit **deletion** and SHALL NOT permit **paraphrase**. A passage may be removed;
it SHALL NOT be reworded to say the same thing in fewer words, because a normative sentence re-said
differently is a changed requirement that no guard detects, while a deletion is visible in a diff
and provable against the normative inventory.

The clause SHALL name what is never cut: any normative sentence, any exit-code contract, any
ordering constraint, any scenario, any worked example, and any recorded reason a rejected
alternative was rejected.

#### Scenario: A passage restating a canonical statement elsewhere

- **WHEN** a file restates what another file states canonically and already cites
- **THEN** the restating passage may be deleted

#### Scenario: A normative sentence is never cut

- **WHEN** a passage carries a SHALL, SHALL NOT, MUST or MUST NOT sentence
- **THEN** it is out of scope for the trim, however repetitive it reads

#### Scenario: Density rewriting is not brevity

- **WHEN** a passage is reworded to say the same thing in fewer words
- **THEN** the clause is violated, because only deletion is permitted

#### Scenario: A generated artifact in another project

- **WHEN** a `/myflow-*` run writes `tasks.md` into a consuming project
- **THEN** that file is subject to the clause
- **AND** the same project's pre-existing `README.md` is not

### Requirement: A byte budget ratchets every owned Markdown file

`scripts/check-contract-budget.sh` SHALL fail when a file it covers exceeds its declared budget.

- It SHALL cover **every** `*.md` and `*.mdc` file this repository owns — `skills/`, `rules/`,
  `openspec/specs/`, `commands/`, `commands-claude/`, `.myflow/` and the repository root — and SHALL
  exclude `node_modules/`, `.superpowers/`, `openspec/changes/archive/` and `docs/superpowers/`.
  Those exclusions SHALL be structural rather than a list of paths, so a file added inside an
  excluded tree needs no edit to the guard. A covered file with no budget entry SHALL be a failure,
  so a file added later cannot silently escape the ratchet.
- Coverage SHALL NOT be limited to files the pipeline loads per run. Regrowth is the failure mode
  and it does not respect that boundary: at the time this requirement was widened, 79 owned files totalling 1.16 MB sat outside the ratchet, of which `openspec/specs/` alone was 462 KB.
- Budgets SHALL be declared as a path-to-maximum-bytes table **inside the script**. One place to
  read; raising a budget is then a visible diff in the guard rather than an edit to the file being
  ratcheted.
- The table SHALL be **hand-maintained**. The guard SHALL NOT offer a mode that rewrites the table
  from current sizes: a regenerate command makes regeneration the path of least resistance, and a
  ratchet whose bound is recomputed on demand does not ratchet.
- Budgets SHALL be per file. A directory total SHALL NOT stand in for the files under it, since one
  file may balloon while another shrinks with the total unchanged.
- It SHALL measure **bytes**, not lines and not tokens — deterministic, and dependent on no
  tokenizer.
- Each budget SHALL be the size the file actually has when the change declaring it lands, **plus
  25%**. The guard is a ratchet against regrowth and SHALL NOT be a target a split can fail against;
  a target-first budget creates pressure to push normative text into an appendix to make a number.
- **The table SHALL be regenerated once, as the last action of a change that alters any covered
  file.** A table computed while later edits are still landing goes stale silently, and it did so
  three times in one change before this rule was written.
- A budget row SHALL be written only after the change's normative-sentence inventory has been shown
  unchanged for that file's group, so a bad trim is not locked in by the ratchet that follows it.
- It SHALL be argument-free and self-scoped from its own location, like `check-references.sh` and
  `check-vocabulary.sh`, with an opt-in environment override for its own test harness alone.
- It SHALL refuse a covered path that is a symlink rather than following it, before reading it.
- Its exit codes SHALL be `0` every file within budget, `1` a file over budget or missing an entry,
  `2` the guard cannot answer at all.

#### Scenario: A file over its budget fails the guard

- **WHEN** a covered file exceeds its declared budget
- **THEN** the guard prints the path, the budget and the actual size
- **AND** it exits 1

#### Scenario: A new file with no budget fails the guard

- **WHEN** a `.md` file is added to a covered location with no entry in the budget table
- **THEN** the guard exits 1 naming that file

#### Scenario: A skill file is covered

- **WHEN** `skills/myflow-do/SKILL.md` grows past its declared budget
- **THEN** the guard exits 1 naming it

#### Scenario: A spec file is covered

- **WHEN** a file under `openspec/specs/` grows past its declared budget
- **THEN** the guard exits 1 naming it

#### Scenario: An archived change is not covered

- **WHEN** a `.md` file under `openspec/changes/archive/` exceeds any size
- **THEN** the guard reports nothing about it, archived changes being a record rather than
  maintained text

#### Scenario: The guard runs from any working directory

- **WHEN** the guard is invoked from outside the repository
- **THEN** it scans the repository it lives in
- **AND** it does not report a false clean by scanning nothing

#### Scenario: The guard is a declared lint step

- **WHEN** `.myflow/project.md`'s `## lint` section is read
- **THEN** `scripts/check-contract-budget.sh` is listed
- **AND** `scripts/test-check-contract-budget.sh` is listed under `## test`

### Requirement: A section reachable from fewer than every command lives in its own file

Where a contract section loaded by every `/myflow-*` command carries rules reachable from fewer than
all of them, that section SHALL be moved into a file of its own, loaded by exactly the commands that
reach it.

This generalises **A section reachable from only one command lives in its own file**, which governs
the one-command case and is unchanged by this requirement. The saving is the same in kind and
smaller in degree: a section reachable from three of five commands stops being loaded by two.

**The threshold is reachability, never size.** A large section every command reaches stays in the
core; a small section two commands reach does not. `Rendering the session records` is 1486 bytes and
is moved, because `/myflow-start` and `/myflow-status` never render a record; `Stage marks` is 6630
bytes and stays, because every producing command marks a stage on every run.

**Reachability is established by reading, never by counting citations.** A command reaches a section
when a step it runs acts on that section's rules. A citation from a file the command loads is
evidence; the absence of one is not proof, because a command may act on a rule it was told about
elsewhere.

The moved file SHALL name, in its own opening prose, which commands load it, so a reader holding the
file can tell whether their command is one of them.

**Each command's own `SKILL.md` SHALL carry the instruction to load it**, at the step needing it.
`rules/myflow-manual-review.mdc`'s contract table SHALL NOT gain a row for such a file: that table is
read in every session whether or not a `/myflow-*` command runs, so a row there converts a per-run
saving into a permanent cost.

#### Scenario: A section two of five commands reach is moved

- **WHEN** a section in `pipeline.md` carries rules only `/myflow-do` and `/myflow-finish` act on
- **THEN** it is moved to its own file under `skills/myflow-contracts/`
- **AND** `/myflow-start`, `/myflow-fast` and `/myflow-status` no longer load it

#### Scenario: A large section every command reaches stays in the core

- **WHEN** a section is among the largest in `pipeline.md`
- **AND** every `/myflow-*` command acts on its rules
- **THEN** it stays in `pipeline.md`
- **AND** its size is not offered as a reason to move it

#### Scenario: The moved file names its own consumers

- **WHEN** a file produced by this rule is read
- **THEN** its opening prose names the commands that load it

#### Scenario: The always-on rule table gains no row

- **WHEN** `rules/myflow-manual-review.mdc` is read after such a move
- **THEN** its contract table carries no row for the moved file
- **AND** the instruction to load it is found in each consuming command's `SKILL.md` instead

#### Scenario: Reachability is not decided by citation count

- **WHEN** a section is considered for a move
- **THEN** each citation into it is read to decide whether the citing command acts on its rules
- **AND** a count of citations is not offered as the decision

