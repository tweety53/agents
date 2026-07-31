# Signal three redesign — report

`scripts/check-unfinished-work.sh`, signal three ("are any review findings still open"), plus the
`## Known incomplete` fence scraper. Operator-approved redesign, change
`kan-17-finish-gate-jira-and-commit-hygiene`.

## Why the shape changed, not the code

Signal three had been rewritten twice and had failed **open** six distinct ways across three review
passes. Every fix closed the shape it was shown and left the class alive, because the guard was
recovering one fact — is any finding still open — by hand-rolling a GFM table parser in awk over a
document whose grammar is defined in another file's prose.

The six, plus two more the last pass found:

1. `Open` capitalised.
2. `open (needs discussion)`.
3. An unescaped `|` in an earlier cell shifting the columns.
4. A reordered header.
5. A row missing its **leading** pipe, which ended table tracking and hid every row below it.
6. A row **both** detached from the table and malformed — caught in-table as `unparseable`, dropped
   silently out-of-table, each half of the parser leaving it to the other.
7. A findings table inside a GFM blockquote, invisible once a legitimate table existed elsewhere.
8. `is_status()`'s `==`, which compares through the locale's collating sequence, so a zero-width
   character in a status reported OUTSTANDING under UTF-8 and vanished under `LC_ALL=C`.

This repository had already paid for the lesson once: `.myflow/project.md` records
`check-plan-provenance`'s Bash classifier taking five review passes and seven fix waves before it
was replaced with a real parser. And `check-cleanup-complete.sh`, in this same change, already
solved the analogous problem the right way — the emitting side writes machine-readable markers
(`registry-row-checked:`) and the guard counts them.

## The design

**The emitting side writes the fact down.** `/myflow-do`'s panel writes, beside each human table
row, one anchored line, and one checksum for the block:

```
findings-total: <n>
finding-status: F<n> open
finding-status: F<n> fixed
finding-status: F<n> withdrawn <reason>
```

The findings table gains an `F<n>` identifier as its first cell and is otherwise unchanged. It is
the record's **human-readable rendering**; the marker line is the record of state.

**The guard counts markers.** Signal three is now a flat list of independent anchored `grep`
invocations and integer comparisons. There is no state machine, no ordering dependence, and no
notion of a block or a table. The only coupling left to the table's shape is one anchored prefix,
`^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`, used solely to hold the row identifiers against the marker
identifiers — and every way of getting that wrong makes the two lists differ, which is OUTSTANDING.

Why a declared total rather than a counted one: counting the human rows means re-deriving the
table's boundaries, which is the deleted defect class returning under another name. The declaration
is one anchored grep, it gives a record with genuinely zero findings a voice (`findings-total: 0`),
and every way of getting it wrong — forgetting to bump it, bumping it without writing the marker —
fails toward OUTSTANDING.

## What was deleted

From `check-unfinished-work.sh`: `cells_of`, `is_delimiter`, `is_header`, `ends_table`,
`is_status`, the orphan rule, the unparseable accounting, the escaped-pipe rule, the header-row
match, the table-boundary state machine, and the six-value awk state tuple. `grep -niE
'cells_of|is_delimiter|is_header|ends_table|orphan|unparseable|is_status'` over the guard returns
one prose mention and no code.

From `test-check-unfinished-work.sh`: every case whose subject was cell splitting, boundary pipes,
table termination or the orphan net. They are replaced by cases asserting that **the shape of the
table cannot change the count** — which is the property the redesign buys.

## Fail-safe cases, and where each is tested

Every one of these is OUTSTANDING. The one direction this guard may never fail in is the
reassuring one.

| Cannot confidently answer | Reported as | Case |
|---|---|---|
| No declared total — the record never spoke | `does not declare a findings total` | 4n |
| A total declared twice, or not a plain count | `more than once` / `not a plain count` | 4p |
| Total disagrees with the marker count, either direction | both numbers named | 4q |
| Table and marker block name different findings (row with no marker, marker with no row, duplicated identifier) | `do not name the same findings` | 4r, 4i |
| Marker lines not one unbroken block | `must be one unbroken block` | 4s-0 |
| A status that is not one of the three legal values, compared as bytes | `none of open, fixed or withdrawn` | 4c, 4d, 4j, 4m |
| A `withdrawn` marker with no reason | `withdrawn finding(s) with no reason` | 4l |
| A line naming the token without being a marker (indented, blockquoted, no identifier) | `line(s) naming finding-status:` | 4k |
| The record unreadable | exit 2, no verdict line | 8d |
| A record with a NUL byte | counts and identifier lists still read | 4s-ii, 4s-iii |
| CRLF line endings | checksum still runs | 4s-i |

A genuine zero-findings panel says `findings-total: 0` and carries no markers, which clears (4o) —
a positive declaration, where silence is not.

Counts are asserted, not presence. Every countable case names its number (`1 open finding(s)`,
`3 open finding(s)`, `3 line(s) naming …`), and the harness header states the rule so a new case
cannot quietly drop back to a bare substring. That was the pass-3 finding about the suite itself.

## The `## Known incomplete` fence scraper

It survives the redesign — a different signal — and its boolean toggle is replaced by CommonMark's
real rule: the opening marker's character and length are tracked; a closer must use the same
character, be at least as long, and carry nothing but whitespace; the at-most-three-spaces indent
rule applies; an unclosed fence runs to end of file; a backtick opener whose info string contains a
backtick opens nothing.

Both reproductions were confirmed to print `CLEAR:` before the fix and OUTSTANDING after
(cases 8c-i, 8c-ii): a closing marker indented four spaces inside an unclosed fence, and a ````
fence documenting a ``` one. Three opposite-direction cases (8c-iii, 8c-iv) stop the fix being
"never close a fence".

The header comment overclaimed that fence tracking closed the hole. It now states the rule the code
implements, names the two shapes the toggle let through, and records what it deliberately does not
model (fences nested in list items or blockquotes) together with why that is the safe direction.

## Attacking the design

Twenty-eight shapes tried. Three worked; all three are closed. Full list.

**Worked, and fixed:**

- **A fenced marker standing in for a missing one.** Because signal three has no notion of a block,
  a fenced `finding-status: F2 fixed` elsewhere in the record substituted for F2's real marker when
  that marker was never written: identifiers matched, the checksum matched, and an open finding read
  `CLEAR:`. **Fixed** by requiring the marker lines to occupy consecutive lines — the fence's own
  delimiter falls between the quoted marker and the block and breaks the run. Closes it without
  teaching this signal about fences. Case 4s-0, with 4s-0b asserting the opposite direction.
- **CRLF silently disabled the checksum.** `read -r _ n` does not split `\r`, so the declared count
  arrived as `2<CR>` and `[ … -ne … ]` exited 2 with "integer expression expected" — inside an `if`,
  which errexit does not see, so the branch was skipped. **Fixed** by taking the count with
  parameter expansion (the fix) and comparing as strings (defence in depth). Case 4s-i, both
  directions.
- **One NUL byte voided the identifier check.** Without `-a`, grep reports "no match" on a binary
  file whatever it contains; both identifier lists came back empty, empty equalled empty, and the
  table-against-markers check passed vacuously. **Fixed** with `-a` on every grep. Case 4s-iii.
  Found by mutation-testing the suite, not by reading the code.

**Tried, safe:** a marker in a blockquote; an indented marker; a marker with no identifier; a
duplicated `F<n>`; a marker numbered for no row; a row with no marker; the declared total too high
and too low; leading, trailing and internal extra whitespace; a marker split across two lines; a tab
instead of a space after the colon; `Open`, `WITHDRAWN`, `open (needs discussion)`; a zero-width
space and a non-breaking space inside a status; a Cyrillic homoglyph inside `fixed`; Arabic-Indic
digits in an identifier; `F01` against an `F1` row; a total of `+1`; a total with a leading zero; a
second total line inside a fence; a whole record in a blockquote; a row added after the marker
block; no trailing newline; a stray line inside the table; a row detached by a blank line or a
heading; the roster and convergence tables; a prose line shaped like a row; the change-name
allowlist under both `C` and `en_US.UTF-8`; and the guide fence variants `~~~~`/`~~~`, a closer
carrying an info string, a 3-space opener with a 0-space closer, and a section inside an HTML
comment.

**Mutation testing.** Thirteen mutations of the guard were run against the suite. Ten were caught.
Three survived and each was investigated rather than waved away:

- Dropping `-a` from `ids_of` — a **real** silent skip; the suite now catches it (case 4s-iii).
- Dropping `export LC_ALL=C` — a **real** gap, and it showed the comment was overclaiming. Measured
  on bash 3.2: the change name `démo` is rejected by the allowlist under `LC_ALL=C` and **accepted**
  under `en_US.UTF-8`, because `case`'s `A-Za-z0-9` are collating ranges. The suite now asserts the
  allowlist under both locales (case 8e-i), and the comment now says what the pin actually buys —
  the status comparison was fixed by the redesign, not by the locale pin.
- Reverting `!=` to `-ne` — **not** a defect: the parameter-expansion extraction is what fixes CRLF,
  and with clean digits the two comparisons are equivalent. Mutating the extraction instead is
  caught. The comment was corrected to stop claiming `!=` is the fix.

Dropping `-a` from `count_matching` and from the total read also survives, and that is correct: when
binary mode drives every *count* to zero the record looks like one with no declared total, which is
OUTSTANDING. The flag stays on those greps so that one guard does not read one file two ways.

## Round two — verification review, and the operator's decision

### A seventh under-count: an identifier reused symmetrically

The verification reviewer found one more, independently. `ROW_IDS != MARKER_IDS` compares two
sorted **multisets**, so a duplicate present on *both* sides cancels out. Two genuinely distinct
findings both labelled `F1` — against the spec's own "unique within the record" — with `F1` also
written twice in the marker block, matched perfectly:

```
| F1 | Lens B   | Minor    | a.sh:1    | genuinely fixed finding |
| F1 | Security | Critical | app.py:12 | genuinely OPEN finding, same id |

findings-total: 2
finding-status: F1 fixed
finding-status: F1 fixed
```

→ `CLEAR: … no finding is open`, over an open Critical, with the word `open` never appearing in a
marker so the open count could not help either. Distinct from case 4r, where the duplication is
asymmetric and was already caught.

Reproduced first, then fixed: uniqueness is now asserted on **each side independently**, naming the
reused identifier, so neither side can be excused by the other carrying the same repeat. Cases 4r-i
(both sides), 4r-ii (each side alone) and 4r-iii (sparse, out-of-order but distinct identifiers must
still reach CLEAR, so the rule is uniqueness and not a contiguous sequence). Recorded as row F41.

### The `Status` column is gone

Operator decision, and it removes residual risk 2 below rather than documenting it. The findings
table is now `ID | Slot | Severity | Location | Note`, and a finding's state is written once, on its
marker line. The reasoning is in the spec text: a status cell beside the marker is a second surface
that can silently disagree with the line that governs, and the protection was asymmetric — a marker
reading `open` blocks whatever the cell says, so the machine's direction was covered, but nothing
covered a human who reads `fixed` in the table and takes it at face value. This change spent a whole
task establishing that a rule is stated once.

Updated in all three places so they cannot drift: the spec requirement and its scenarios, the mirror
in `skills/myflow-do/SKILL.md`, and the live record — whose 42 rows were rewritten with every
non-status cell verified byte for byte against the pre-migration copy, zero mismatches. Harness
fixtures follow the new format; the two that still carry a stale `Status` cell are deliberately
malformed records kept to show that such a cell reaches no count. Nothing else in the repository
references the column except two historical records — the preserved session copy under `docs/` and
the change's own `tasks.md` — which are accounts of what was done at the time and were left alone.
Recorded as row F42.

One latent bug was found while rewriting the fixtures and fixed: `write_panel` briefly used
`seq 1 $#`, and BSD `seq 1 0` prints "1 0", so a panel with no findings would have been written with
two rows. It iterates over `"$@"` instead.

## Residual risks

1. **A finding the record never mentions.** Delete the row, the marker and decrement the total, or
   simply write `fixed` on a marker for a finding that is open, and the guard reports `CLEAR:`. No
   document format can detect a finding that was never written down; this is the same class as any
   emitter lying about its own state, and it is why the panel's fix subagents are forbidden from
   writing `withdrawn` at all. Not closable here.
2. ~~**The table's `Status` cell can diverge from its marker.**~~ **Closed in round two** by
   removing the column on the operator's decision. A finding's state is now written once, on its
   marker line, so there is no second surface to disagree with it.
3. **The panel record's path is not symlink-checked.** A symlink at
   `.superpowers/sdd/final-review-panel.md` pointing at a clean decoy reads as that decoy. This is
   unchanged from before the redesign, the path is not derived from the PR-controlled change name
   (the allowlist covers that), and planting the symlink requires write access to the worktree — at
   which point the record itself can be rewritten directly. Recorded, not fixed.
4. **`## Known incomplete work` is accepted as the section**, because the heading is matched as a
   prefix. Pre-existing and unchanged; it widens what counts as the section, and the section's
   content is still checked.

## The live record

`.superpowers/sdd/final-review-panel.md` is migrated: 42 rows, 42 markers, `findings-total: 42`,
zero open findings, no repeated identifier on either side, and it parses clean under the new guard.
The 31 pre-existing rows keep their text and gain identifiers and markers. Rows F32-F38 are pass 3's
findings — the sixth evasion, the blockquoted table, the collation-dependent comparison, the
orphan-rule false positive, the harness asserting presence rather than counts, the fence toggle, and
the header overclaim. Rows F39-F40 are the two defects found by attacking the redesign afterwards.
Rows F41-F42 are round two: the symmetric duplicate identifier, and the status column's removal. It
is the first document written in the new format.
