## MODIFIED Requirements

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

## ADDED Requirements

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
