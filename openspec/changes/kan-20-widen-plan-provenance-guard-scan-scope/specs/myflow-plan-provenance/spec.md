## MODIFIED Requirements

### Requirement: A number with no provenance may not appear in a plan

A numeric claim SHALL carry `measured:` or `predicted:`. A number carrying neither SHALL NOT appear
in the plan at all.

Stating an unsourced number is itself the violation. Labelling alone would not have helped the case
this rule exists for: a baseline of "194 tests" was not mislabelled, it was invented, and the plan
that carried it was wrong at every task that read it.

A number that is **quoted rather than asserted** is not a claim, and SHALL NOT require a tag. A
numeric string enclosed in a matched pair of double quotes — straight (`U+0022`) or curly
(`U+201C`/`U+201D`) — is being reproduced, not stated. Without this carve-out the rule demands a tag
that cannot be written truthfully: the documentation of this very contract must reproduce the
invented baseline in order to explain why the rule exists, and neither `measured:` (asserting a run
that never happened) nor `predicted:` (naming a confirmation that will never come) is an honest
label for it.

Matched double quotes SHALL be the whole delimiter set. A CommonMark inline code span SHALL NOT
delimit a quotation: a number written inside backticks is an ordinary numeric claim and SHALL be
reported. Single quotes and apostrophes SHALL NOT delimit a quotation either, because prose
apostrophes are unpaired by nature and admitting them would make pairing meaningless.

The carve-out SHALL be recognised on a single line only, and SHALL fail closed: it may only ever
remove a finding, and only where the number is demonstrably enclosed. Three vetoes SHALL enforce
that, each able only to withdraw the carve-out and never to grant it:

- **Class-wide.** Where the delimiters of a class do not all pair on a line, that class SHALL yield
  no region on that line — not merely ignore the unpaired delimiter.
- **Escape.** A backslash-escaped delimiter is a literal character (CommonMark §2.4). A line
  carrying one SHALL lose the carve-out entirely. An odd run of backslashes escapes the delimiter
  after it; an even run leaves it live.
- **Angle bracket.** A delimiter inside a raw-HTML construct — an HTML comment, tag, declaration,
  CDATA section or processing instruction (CommonMark §6.6) — is literal content, and locating those
  constructs cannot be done with an approximate grammar: six fail-opens have come from trying. A line
  containing a `<` ANYWHERE ON IT SHALL therefore lose the carve-out entirely, whether or not the `<`
  opens a construct, whether or not that construct closes, and whether or not any delimiter sits
  inside it. The guard SHALL NOT parse HTML in order to make this decision.

Both whole-line vetoes SHALL be coarse, and the angle-bracket veto SHALL be the coarsest rule that
states the property: it may ask only whether a `<` is present, never what the `<` is part of or where
that part ends.

Where a violation is reported on a line whose carve-out was withdrawn by a veto, the guard SHALL
also report which veto fired, what would resolve it, and where the contract is. It SHALL NOT change
the exit code or the failure count on that account. Rewording the line SHALL be offered as a remedy
wherever adding a tag is offered, because a tag written for a number that was already correctly
quoted is a false statement of provenance.

#### Scenario: A measured baseline names how it was measured

- **WHEN** a plan states a test-count or other measured baseline
- **THEN** it carries `measured:<command> @ <ref>` naming the command and the revision

#### Scenario: A forecast is distinguishable from a measurement

- **WHEN** a plan states a number that could not be measured while planning, such as a count after a
  change that has not been made
- **THEN** it carries `predicted:<what-confirms-it>` rather than `measured:`
- **AND** the reader can tell the two apart without consulting anything else

#### Scenario: An unsourced number is removed rather than tagged

- **WHEN** a number cannot be measured and no command would confirm it
- **THEN** it does not appear in the plan

#### Scenario: A quoted number is reproduced, not claimed

- **WHEN** a plan encloses a numeric string in a matched pair of double quotes, straight or curly
- **THEN** the guard does not require a provenance tag for it
- **AND** the same number written bare on the same line is still a violation

#### Scenario: A number in a code span is an ordinary claim

- **WHEN** a plan writes a numeric string inside a CommonMark inline code span, such as
  `` `85 lines` ``, with no provenance comment
- **THEN** the guard reports it as an untagged numeric claim

#### Scenario: An unmatched delimiter does not create an exemption

- **WHEN** a numeric claim follows a single unmatched quote character on its line
- **THEN** the guard still reports it as an untagged numeric claim
- **AND** it reports that the carve-out was withdrawn because the line's delimiters do not pair

#### Scenario: An escaped delimiter withdraws the carve-out from its whole line

- **WHEN** a line carries a backslash-escaped quotation delimiter, such as `\"`
- **THEN** the guard reports every numeric claim on that line, including one enclosed by a pair of
  delimiters that would otherwise have matched
- **AND** it reports that the carve-out was withdrawn because of the escaped delimiter

#### Scenario: A delimiter inside raw HTML withdraws the carve-out from its whole line

- **WHEN** a line encloses a numeric claim between quote characters that sit inside HTML comments,
  such as `See <!--"--> the benchmark ran 77 tests <!--"-->.`
- **THEN** the guard reports the claim, because a delimiter inside raw HTML is literal content
- **AND** it reports that the carve-out was withdrawn because the line contains a `<`

#### Scenario: A `<` carrying no delimiter still withdraws the carve-out

- **WHEN** a line contains a `<` with no quotation delimiter anywhere near it, alongside a numeric
  string enclosed in a matched pair of double quotes
- **THEN** the guard reports that number as an untagged numeric claim
- **AND** it reports that the carve-out was withdrawn because the line contains a `<`

#### Scenario: A `>` inside a single-quoted attribute value cannot hide a delimiter

- **WHEN** a line carries a tag whose single-quoted attribute value contains both a `>` and a
  quotation delimiter, such as `<a b='>"'> the benchmark ran 77 tests "`
- **THEN** the guard reports the claim, because the tag's own extent is never computed and so cannot
  be computed wrongly

#### Scenario: A veto names a cause the author can see on the line

- **WHEN** the guard reports a numeric claim on a line whose carve-out a veto withdrew
- **THEN** the stated cause is a fact the author can confirm by reading that line — an escaped
  delimiter, unpaired delimiters, or a `<` — and never a construct the guard merely inferred

#### Scenario: A withdrawn carve-out names its remedy

- **WHEN** the guard reports a numeric claim on a line whose carve-out a veto withdrew
- **THEN** the report names the veto, what would resolve it, and the provenance contract
- **AND** the exit-1 remedy banner offers rewording the line as well as adding a tag

### Requirement: Provenance is enforced mechanically, and only where it applies

A guard SHALL fail when a plan contains an untagged fenced code block or an untagged numeric claim,
reporting `file:line`.

The guard SHALL check only that provenance is **stated**, never whether it is true. No script can
confirm that a verification was performed; a guard claiming otherwise would be both unimplementable
and dishonest about what it proves.

The guard's scope SHALL be a change's `tasks.md`, `design.md` and `proposal.md`, excluding archived
changes. Both rules SHALL apply to all three files: the asymmetry this specification draws is
between how a block is tagged and whether a number may appear at all, never between which files
each rule reads.

A change's `specs/` directory SHALL NOT be scanned. Spec text legislates rather than describes, and
a requirement forbidding a shape must be able to name that shape.

A guard that fires outside the file type its rule is about produces false failures, and false
failures are answered with suppression markers that switch off the real checks sharing those lines.
That risk is what bounds this scope, and what obliges a widening to remove its own false-positive
classes in the same change rather than after them.

The path-containment refusal that protects a scanned file — refusing a symlink, a path escaping the
changes directory, or a non-regular file — SHALL apply to every file the guard scans, not only to
`tasks.md`.

An issue key followed by a unit word SHALL NOT be read as a numeric claim. `KAN-6 errors` states no
quantity.

#### Scenario: An untagged block fails the guard

- **WHEN** the guard runs over a plan containing a fenced block with no provenance tag
- **THEN** it exits non-zero and reports the file and line

#### Scenario: A correctly tagged plan is silent

- **WHEN** the guard runs over a plan whose blocks and numbers all carry tags
- **THEN** it exits zero and reports no failures

#### Scenario: Archived plans are not scanned

- **WHEN** the guard runs in a repository containing archived changes with untagged plans
- **THEN** those plans are not scanned and do not cause a failure

#### Scenario: A change's design and proposal are scanned

- **WHEN** a non-archived change's `design.md` or `proposal.md` contains an untagged fenced block or
  an untagged numeric claim
- **THEN** the guard exits non-zero and reports that file and line

#### Scenario: Delta specs are not scanned

- **WHEN** a non-archived change's `specs/` directory contains an untagged numeric claim
- **THEN** the guard does not report it

#### Scenario: Documentation outside a plan is not scanned

- **WHEN** a skill file, contract or README contains an untagged fenced code block
- **THEN** the guard does not fail — its scope is a change's own planning artifacts

#### Scenario: A scanned file that fails containment is refused, not read

- **WHEN** a change's `design.md` or `proposal.md` is a symlink, escapes the changes directory, or
  is not a regular file
- **THEN** the guard refuses that file with the containment exit code rather than reading it
- **AND** every other change directory is still scanned

#### Scenario: An issue key is not a quantity

- **WHEN** a plan contains an issue key immediately followed by a unit word, such as `KAN-6 errors`
- **THEN** the guard does not report it as an untagged numeric claim
