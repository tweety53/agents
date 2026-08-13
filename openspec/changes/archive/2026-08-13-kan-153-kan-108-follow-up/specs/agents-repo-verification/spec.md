## ADDED Requirements

### Requirement: The panel-record marker helpers have one definition

The helpers that read a review panel record's marker lines — counting matching lines, extracting
matching text, and collecting finding identifiers — SHALL be defined **once**, in a sourced library,
and SHALL NOT be reimplemented in any guard that needs them.

`scripts/lib/panel-record.sh` SHALL export:

- `count_matching <file> <ere>` — the number of lines of `<file>` matching `<ere>`;
- `grep_lines_of <file> <mode> <ere>` — every match, one per line, where `mode` is `match` (the
  matched substring only) or `line` (the whole matching line);
- `ids_of <file> <ere> <shape>` — the finding identifiers of every matching line, where `shape` is
  `digits` (bare digits, sorted, **duplicates retained**) or `ids-unique` (`F<n>`, sorted and
  de-duplicated).

Both shapes SHALL exist because the two callers ask different questions: a duplicate-identifier
check reads a list that retains duplicates, and a set-comparison check reads one that does not. A
library exporting only one shape would silently change a caller's answer.

Every helper SHALL pass `-a` to `grep`, so a record carrying a stray NUL byte cannot put `grep` into
binary mode where it reports "no match" whatever the file contains. Every helper SHALL distinguish
`grep`'s exit 1 ("no match", which is an answer) from an exit of 2 or more (a real error, which is a
refusal), and SHALL propagate the latter to its caller rather than returning a count of zero. Every
helper SHALL pass `--` before the path, so a record path beginning with `-` is never read as an
option.

`scripts/check-unfinished-work.sh` and `scripts/check-panel-reproducers.sh` SHALL source that
library and SHALL NOT carry local copies of those helpers. Neither guard's exit-code contract,
reported output, nor observable behaviour SHALL change as a result. A guard that cannot source the
library SHALL exit 2 — it cannot answer — and SHALL NOT continue with a reduced set of checks.

Every numeric field these guards read from a record SHALL bound the digit run it accepts, so that
the two guards agree on what a well-formed total is. `check-unfinished-work.sh`'s `findings-total`
pattern SHALL accept `(0|[1-9][0-9]{0,14})`, the bound `check-panel-reproducers.sh` already applies
to `reproducers-total`. A total outside that bound SHALL be reported as not a plain count, through
whichever disposition that guard's own contract defines for a defect it can read.

**The unbounded pattern is a divergence between two guards reading one record format, and is not a
crash.** `check-unfinished-work.sh` compares the declared total to the marker count as strings, so
an over-long total is compared and reported rather than evaluated; the arithmetic crash
`check-panel-reproducers.sh` was bounded against does not exist on this path. The requirement stands
on the divergence alone: one record format read by two guards that disagree on which totals are
well-formed is the same drift this library exists to close.

#### Scenario: Both guards read the same marker helper definitions

- **WHEN** either guard counts marker lines, extracts matching text, or collects finding identifiers
- **THEN** it calls the sourced library's helper
- **AND** no copy of that helper exists in the guard's own file

#### Scenario: The duplicate-retaining shape is preserved

- **WHEN** `check-unfinished-work.sh` collects identifiers to find repeats
- **THEN** it requests the `digits` shape, and the list it receives retains duplicates

#### Scenario: The de-duplicated shape is preserved

- **WHEN** `check-panel-reproducers.sh` collects identifiers to compare two sets
- **THEN** it requests the `ids-unique` shape, and the list it receives is `F<n>` and de-duplicated

#### Scenario: A grep error is a refusal, not a count of zero

- **WHEN** a helper's `grep` exits with a status of 2 or more
- **THEN** the helper returns non-zero and the calling guard refuses, rather than reading the failure
  as "nothing matched"

#### Scenario: An oversized declared total is not a plain count

- **WHEN** a panel record declares `findings-total:` followed by a digit run longer than 15 digits
- **THEN** `check-unfinished-work.sh` reports it as not a plain count, in the same disposition it
  gives every other malformed total — an `OUTSTANDING:` verdict at exit 0, that guard's stated
  contract for outstanding work found, exit 1 being its sibling's contract and not its own

#### Scenario: A missing library is a refusal

- **WHEN** a guard is invoked and `scripts/lib/panel-record.sh` cannot be sourced
- **THEN** the guard exits 2 and reports that it cannot answer

### Requirement: A guard reads this repository's own Markdown for structural damage

`scripts/check-markdown-integrity.py` SHALL parse every `skills/**/*.md` and `rules/*.mdc` file into
blocks — fenced code, table, heading, list item, blockquote, paragraph — and SHALL report structural
damage that changes how the text reads. It SHALL be written in Python 3 using the standard library
only, following the precedent set by `scripts/check-plan-provenance.py`, because the classification
it performs is block structure and not a line pattern.

It SHALL report each violation as `file:line: <signal>: <what is wrong>`, and SHALL detect exactly
these four signals:

- **broken code span** — a single-backtick code span containing a backslash-escaped backtick, since
  backslash is not an escape inside a code span under CommonMark and the span therefore closes
  early, rendering the remainder as plain text; or an odd backtick count on a line outside a fenced
  block.
- **orphaned blockquote body** — a line beginning `> ` whose immediately preceding non-blank line is
  un-marked prose that does not end a sentence, i.e. a blockquote whose first line lost its marker.
- **unterminated paragraph** — a paragraph, outside fenced blocks, tables, headings and list items,
  whose last line ends with none of `.` `:` `?` `!` `)` `|` `"` `'` or a closing backtick.
- **dangling promise** — a paragraph that is the last block before a heading or the end of the file,
  whose final sentence ends in a colon, with no list, fenced block, table or paragraph following.

The dangling-promise signal SHALL be defined by that structural rule alone and SHALL NOT key on
phrases such as "see … for why" or "the reason is". This repository's contracts cite one another in
that form throughout, and a phrase rule would fire on correct text, which the lint policy could only
answer with suppression markers it forbids.

It SHALL take an optional project root, defaulting to this repository, so it runs against a bare tree
and can be named in `## lint`. It SHALL exit **0** when every scanned file is clean, **1** when
violations are found, each reported with its line, and **2** when it cannot answer at all — an
unreadable file, an argument it does not accept, or a scope root that does not exist.

It SHALL be named in `.myflow/project.md` under `## lint`, and its harness
`scripts/test-check-markdown-integrity.sh` SHALL be named under `## test`.

A violation SHALL be fixed by repairing the text — closing the span, restoring the blockquote
marker, completing the sentence, supplying what the colon promised — and never by narrowing the
guard's scope or adding a suppression marker.

#### Scenario: A code span broken by an escaped backtick is caught

- **WHEN** a scanned file contains a single-backtick code span with a backslash-escaped backtick
  inside it
- **THEN** the guard exits 1 and reports that file and line as a broken code span

#### Scenario: A blockquote that lost its first marker is caught

- **WHEN** a `> ` line follows un-marked prose that does not end a sentence
- **THEN** the guard exits 1 and reports that file and line as an orphaned blockquote body

#### Scenario: A sentence torn mid-clause is caught

- **WHEN** a paragraph outside any fence, table, heading or list item ends with no terminal
  punctuation
- **THEN** the guard exits 1 and reports that file and line as an unterminated paragraph

#### Scenario: A colon promising content that never arrives is caught

- **WHEN** the last block before a heading is a paragraph whose final sentence ends in a colon, and
  no list, fence, table or paragraph follows it
- **THEN** the guard exits 1 and reports that file and line as a dangling promise

#### Scenario: An ordinary citation ending a section does not fire

- **WHEN** a section's last paragraph ends with a sentence of the form `See **X** (`file.md`) for
  why …`, terminated by a full stop
- **THEN** the guard reports nothing for that paragraph

#### Scenario: Content inside a fence is not scanned for prose signals

- **WHEN** a fenced code block contains an unterminated line, an unbalanced backtick, or a `>`
  character
- **THEN** the guard reports nothing for it

#### Scenario: An unreadable file is a refusal

- **WHEN** a file inside the scanned scope exists but cannot be read
- **THEN** the guard exits 2 and reports that it cannot answer, rather than reporting the file clean
