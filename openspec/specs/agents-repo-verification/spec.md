# agents-repo-verification Specification

## Purpose
TBD - created by archiving change kan-10-myflow-economical-updates. Update Purpose after archive.
## Requirements
### Requirement: This repository declares its own myflow project configuration

`.myflow/project.md` SHALL exist in this repository and SHALL declare `## apps`, `## test`,
`## lint`, `## standards`, and `## jira`. `## test` SHALL name `scripts/test-setup.sh`,
`scripts/test-check-references.sh`, `scripts/test-check-plan-provenance.sh`,
`scripts/test-check-finish-preflight.sh` and `scripts/test-preserve-session-records.sh`. `## lint`
SHALL name `scripts/check-vocabulary.sh`, `scripts/check-references.sh` and
`scripts/check-plan-provenance.sh`, and SHALL state explicitly that no auto-fix command exists.

`## lint` SHALL name every guard that **scans the repository tree** and can therefore run at any
time. A check that requires context no tree scan provides — a branch, a worktree, a resolved base ref
— SHALL NOT be named in `## lint`, because a lint step that cannot run outside a change would fail on
every unrelated invocation. Such a check is invoked by the command that owns it, and its harness is
named in `## test` like any other.

#### Scenario: myflow reads this repository's own commands

- **WHEN** a myflow change in this repository reaches its verification step
- **THEN** the verification it runs is the guard scripts and their harnesses
- **AND** no Gradle, npm, or other unrelated project's command is invoked

#### Scenario: The absent auto-fix step is explicit

- **WHEN** `.myflow/project.md` `## lint` is read
- **THEN** it states that this repository has no auto-fix command
- **AND** an agent following the Lint Fix Priority rule can tell the auto-fix step is
  inapplicable rather than skipped

#### Scenario: The apps section describes a repository with no runnable app

- **WHEN** `.myflow/project.md` `## apps` is read
- **THEN** it states that there is no runnable application and names the guard scripts and the
  sandboxed `setup.sh` run as the verification surface

#### Scenario: A new tree-scanning guard is reachable through the declared configuration

- **WHEN** a guard script that scans the repository tree is added to this repository
- **THEN** it is named in `## lint` and its harness in `## test`
- **AND** a myflow run verifying this repository invokes it without any per-guard special-casing

#### Scenario: A context-requiring check is tested but not linted

- **WHEN** a check is added that needs a branch, worktree or resolved base ref to run
- **THEN** its harness is named in `## test`
- **AND** it is not named in `## lint`, because it cannot run against a bare tree

### Requirement: A guard fails when a cross-referenced section no longer exists

`scripts/check-references.sh` SHALL, for every backticked `.md` or `.mdc` path that resolves to a
real file inside the repository, require that at least one bold token the line **associates** with
that path matches a `#`–`####` heading in the referenced file. A bold token is associated only when
it is adjacent to the path in one of the shapes references are written in, and each token is
assigned to the nearest path it is associated with. A path the line associates no token with is not
checked. The script SHALL report every failure as `file:line` and exit non-zero.

The script SHALL refuse to read any path that normalizes outside the repository root, SHALL decide
that from the path's **shape** before any existence test so the verdict does not depend on what is
on the machine, and SHALL treat such a reference as a **failure** rather than a note — a clean exit
means every reference was checked. It SHALL exit `2` rather than report a clean run when its root
is unset-but-empty, is not a directory, or contains no Markdown at all.

The script SHALL take no arguments and SHALL own its scan set in a single `DEFAULT_TARGETS`
definition, matching the contract of the repository's other guards.

#### Scenario: A moved section is caught

- **WHEN** a section is removed from a file that other files reference by name
- **THEN** the script reports each referring `file:line` and exits non-zero

#### Scenario: A live reference passes

- **WHEN** every referenced section still exists in the file it is referenced from
- **THEN** the script exits `0`

#### Scenario: Reference phrasing variants are recognized

- **WHEN** a line reads "see **State file** in `…`", "per **State file** in `…`", or "defined once
  under **State file** in `…`"
- **THEN** all three are checked identically

#### Scenario: Emphasis that is not a section reference does not fail the guard

- **WHEN** a line carries a bold token used for emphasis and, elsewhere on the same line, a
  backticked path the token is not written next to
- **THEN** the path is not checked against that token and the line does not fail

#### Scenario: An associated token beside emphasis is still checked

- **WHEN** the same line carries both unrelated emphasis and a bold token written in a reference
  shape whose section no longer exists
- **THEN** the script reports the line and exits non-zero

#### Scenario: A token is assigned to its nearest path

- **WHEN** a line names two paths and a bold token sits adjacent to the second
- **THEN** the token is required to resolve in the second file only, not in the first

#### Scenario: A reference outside the repository root is refused, not read

- **WHEN** a line references a path that normalizes outside the repository root
- **THEN** the file is not opened, the refusal is reported as a failure, and the script exits
  non-zero

#### Scenario: Containment does not depend on what exists on the machine

- **WHEN** two such references are checked, one whose out-of-tree target exists and one whose does
  not
- **THEN** both are refused, reported, and fail identically

#### Scenario: An unscannable root never reports clean

- **WHEN** the root override is empty, names a nonexistent directory, or holds no Markdown files
- **THEN** the script exits `2` instead of reporting that all references resolve

#### Scenario: The script is argument-free and self-scoped

- **WHEN** the script is invoked with no arguments from any directory in the repository
- **THEN** it scans its own `DEFAULT_TARGETS` and does not require a path list

### Requirement: The repository's guards run during implementation

This repository's guards — `scripts/check-vocabulary.sh`, `scripts/check-references.sh`,
`scripts/test-check-references.sh` and `scripts/test-setup.sh` — SHALL be the project's `## lint`
and `## test` commands in `.myflow/project.md`, run during `/myflow-do` as part of implementing
and verifying the work. A non-zero exit from any of them SHALL be treated as a failing
verification that blocks the `IN_PROGRESS` handoff.

`scripts/test-state-advance.sh` is not among them: the script it asserted against,
`skills/myflow-state-advance/state-advance.sh`, is deleted with the `myflow-state-advance`
capability. Four guards remain.

#### Scenario: The guards run during implementation

- **WHEN** `/myflow-do` verifies work in this repository
- **THEN** all four scripts run and their output is shown as evidence

#### Scenario: Each guard's own assertion harness is invoked

- **WHEN** the guards run
- **THEN** `scripts/test-check-references.sh` is among them, so that harness is not a guard
  nothing invokes

#### Scenario: A retired harness is not invoked

- **WHEN** the guards run
- **THEN** `scripts/test-state-advance.sh` is not run and does not exist

#### Scenario: A failing guard blocks the handoff

- **WHEN** `scripts/check-references.sh` exits non-zero during `/myflow-do`
- **THEN** the change does not hand off to the human gate and the failure is reported

#### Scenario: Nothing re-runs them before integration

- **WHEN** `/myflow-finish` integrates the change
- **THEN** it runs none of the guards, because verification happened during `/myflow-do`

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

