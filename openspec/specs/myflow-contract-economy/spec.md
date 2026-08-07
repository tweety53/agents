# myflow-contract-economy Specification

## Purpose
How a contract file under `skills/myflow-contracts/` is partitioned into a normative core that
every `/myflow-*` command loads and a rationale appendix that none of them do, what distinguishes
the two, the edits a move may make, and the byte budget that keeps the cores from growing back.

## Requirements
### Requirement: A contract file separates its normative core from its rationale

A contract file under `skills/myflow-contracts/` that carries both rules and the reasoning behind
them SHALL be split into two files: a **core** at the original path, and a **rationale appendix**
at the same path with a `-rationale` suffix before the extension.

One test decides which file a passage belongs in, applied to sentence groups — a paragraph, a
bullet, a table — and never to clauses inside a sentence:

> A passage is **core** if removing it would change what an agent does. A passage is **rationale**
> if removing it changes no agent behaviour.

Core therefore holds tables, ordered step lists, handoff templates, verdict tables, code blocks an
agent runs, and every prohibition. Rationale holds justification of an ordering, alternatives
considered and rejected, history such as "this reverses an earlier decision", measurements, and
meta-commentary about the document itself.

**The test is behavioural, not typographic.** A prose paragraph stating a rule is core: the two
files in this change measure 78% and 90% prose, and much of that prose is normative.

#### Scenario: A rule stated in prose stays in the core

- **WHEN** a contract paragraph states a prohibition in prose rather than in a table
- **THEN** that paragraph is in the core file
- **AND** it is not moved to the appendix on the grounds of being prose

#### Scenario: A justification moves to the appendix

- **WHEN** a contract paragraph explains why one step is ordered before another, and removing it
  would leave the ordering itself still stated
- **THEN** that paragraph is in the appendix

#### Scenario: The appendix path is derived from the core path

- **WHEN** `skills/myflow-contracts/<name>.md` is split
- **THEN** its appendix is at `skills/myflow-contracts/<name>-rationale.md`

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

#### Scenario: A stale position word is deleted rather than reworded

- **WHEN** a passage moves to another file and a citation in it says `above`
- **THEN** the word `above` is deleted
- **AND** it is not replaced with the name of the file that now holds the target

#### Scenario: A definition stays with the term that uses it

- **WHEN** a core passage uses a term whose only definition is in a passage classified as rationale
- **THEN** that defining passage stays in the core

A rewritten core would produce two independently authored statements of one rule — the failure a
contract split exists to avoid — so the constraint is met by construction rather than by review
diligence, and the diff is reviewable as a move.

**The unit that moves SHALL be a whole blank-line-delimited paragraph, bullet or table, never part
of one.** A paragraph mixing a rule with its justification SHALL stay in the core entire, and the
lines that remain around any moved passage SHALL NOT be re-wrapped.

**This follows from the verification rather than from taste, and its cost is accepted.** A contract
file is hard-wrapped, so a sentence boundary almost never coincides with a line break; splitting
mid-paragraph requires re-wrapping the lines that stay, and a re-wrapped line is indistinguishable
from a lost one in the check that proves nothing was dropped. A sentence-granular partition would
have to give up the only mechanical guarantee this change has against silent rule loss across
thousands of lines, which is worth more than the bytes such a split would move. What it costs is
that rationale welded into a rule's own paragraph stays in the core.

#### Scenario: A paragraph mixing a rule and its justification stays whole

- **WHEN** a core paragraph states a rule in one sentence and justifies it in the next
- **THEN** the whole paragraph remains in the core
- **AND** no part of it appears in the appendix

#### Scenario: Surviving lines are not re-wrapped

- **WHEN** a passage is moved out of a core file
- **THEN** the lines remaining around it keep their original wrapping

#### Scenario: A moved passage is byte-identical

- **WHEN** the multiset of non-blank lines of core and appendix together is compared with that of
  the original file at the previous revision
- **THEN** the two multisets are equal, excluding lines whose only difference is a repointed
  citation path

#### Scenario: A repointed citation resolves

- **WHEN** a core passage cites a section that moved to the appendix
- **THEN** the citation names the appendix path
- **AND** `scripts/check-references.sh` resolves it against a real heading in that file

#### Scenario: Rewording is a defect

- **WHEN** a line differs from the original in any way other than a repointed citation path
- **THEN** the split is not complete

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

A rationale appendix exists for whoever edits a contract. No `/myflow-*` command SHALL load one as
part of a run, and no skill SHALL instruct an agent to load one.

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

### Requirement: A section reachable from only one command lives in its own file

Where a contract section's rules are reachable from exactly one command, that section SHALL be
moved whole — its rules **and** its reasoning together — into a file of its own, so that every
other command stops loading it.

Such a file is **not** core/appendix split. Splitting it would buy a saving only on the one command
that loads it, while costing that command an extra file.

`### Follow-up issues` in `jira-integration.md` is reachable only from `/myflow-finish` run 1 and is
457 lines / 35467 bytes — 67% of that file — so `/myflow-start` and `/myflow-do` stop loading it.

#### Scenario: Follow-up issues moves to its own file

- **WHEN** `skills/myflow-contracts/jira-integration.md` is read after this change
- **THEN** it does not contain the `Follow-up issues` section
- **AND** `skills/myflow-contracts/jira-followups.md` contains it, with its reasoning inline

#### Scenario: A single-command file is not split further

- **WHEN** `skills/myflow-contracts/jira-followups.md` is read
- **THEN** it carries both its rules and its reasoning
- **AND** no `jira-followups-rationale.md` exists

#### Scenario: A start run does not load the follow-up contract

- **WHEN** `/myflow-start` runs
- **THEN** it loads `jira-integration.md` and not `jira-followups.md`

### Requirement: A byte budget ratchets every contract file

`scripts/check-contract-budget.sh` SHALL fail when a file under `skills/myflow-contracts/` exceeds
its declared budget.

- It SHALL cover **every** `*.md` in that directory, cores, appendices and `SKILL.md` alike. A file
  present in the directory with no budget entry SHALL be a failure, so a contract added later
  cannot silently escape the ratchet.
- Budgets SHALL be declared as a path-to-maximum-bytes table **inside the script**. One place to
  read; raising a budget is then a visible diff in the guard rather than an edit to the file being
  ratcheted.
- It SHALL measure **bytes**, not lines and not tokens — deterministic, and dependent on no
  tokenizer.
- Each budget SHALL be the size the file actually has when this change lands, **plus 25%**. The
  guard is a ratchet against regrowth and SHALL NOT be a target the split itself can fail against;
  a target-first budget creates pressure to push normative text into an appendix to make a number.
- It SHALL be argument-free and self-scoped from its own location, like `check-references.sh` and
  `check-vocabulary.sh`, with an opt-in environment override for its own test harness alone.
- Its exit codes SHALL be `0` every file within budget, `1` a file over budget or missing an entry,
  `2` the guard cannot answer at all.

#### Scenario: A file over its budget fails the guard

- **WHEN** a file under `skills/myflow-contracts/` exceeds its declared budget
- **THEN** the guard prints the path, the budget and the actual size
- **AND** it exits 1

#### Scenario: A new contract file with no budget fails the guard

- **WHEN** a `.md` file is added to `skills/myflow-contracts/` with no entry in the budget table
- **THEN** the guard exits 1 naming that file

#### Scenario: The guard runs from any working directory

- **WHEN** the guard is invoked from outside the repository
- **THEN** it scans the repository it lives in
- **AND** it does not report a false clean by scanning nothing

#### Scenario: The guard is a declared lint step

- **WHEN** `.myflow/project.md`'s `## lint` section is read
- **THEN** `scripts/check-contract-budget.sh` is listed
- **AND** `scripts/test-check-contract-budget.sh` is listed under `## test`
