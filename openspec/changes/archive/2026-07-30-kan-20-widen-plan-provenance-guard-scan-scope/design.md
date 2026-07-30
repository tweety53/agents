# Design — widen the plan-provenance guard's scan scope

## What the measurements showed

Every figure below comes from the guard's own classifier: a throwaway script imported
`scripts/check-plan-provenance.py` and called `ProvenanceGuard.check_file` directly, rather than
approximating the rule with a grep.

Real violations the widened scope would report across the archived changes:

| Candidate scope | Violations | Classification aborts |
|---|---|---|
| `design.md` | 3 | 0 |
| `proposal.md` | 3 | 0 |
| change `specs/**` | 1 | 0 |

<!-- measured: a machine-local throwaway script importing scripts/check-plan-provenance.py via
     importlib and calling ProvenanceGuard.check_file over openspec/changes/archive/*/ @ c1f5aa6.
     Not preserved; task 1 re-establishes these as real test cases, after which the harness is the
     reproducible record. -->

KAN-20 predicted this would "fire heavily", projecting five to seventy-one hits per `design.md` from
a bare-number proxy. The real `CLAIM_RE` requires a digit group followed by one of a small closed
set of unit words, so the true cost is an order of magnitude smaller. Zero classification aborts
also means the container-prefix grammar already handles these file types — the mechanical risk is
not where the ticket expected it.

Of the hits, most are false positives in two classes:

- **Self-referential quotation.** The historical invented baseline — reproduced as a quoted string —
  is flagged in KAN-14's `design.md`, `proposal.md` and delta spec, because the provenance
  documentation must quote it to explain itself. No tag is honest for it.
- **Issue-key adjacency.** `catches all six KAN-6 errors` is flagged because the classifier reads a
  quantity out of the issue key. The module docstring claims this exclusion already exists; it holds
  only when the key is not followed by a unit word, so the code lacks the property the docstring
  describes.

The genuine catches are an unattributed ratio of loaded-to-total lines in kan-10's `design.md`, and
an elapsed-time claim in kan-14's `proposal.md`.

## Decisions

### Widen the scope and fix its false-positive classes, rather than re-aiming at tag truth

**ID:** `scope-not-truth`
**Status:** active
**Chosen:** Widen to `design.md`/`proposal.md` and fix both false-positive classes in the same
change — honest that this catches two real defects, not the four the ticket cites.
**Considered:** *Block rule only* — zero measured cost, but it catches neither real numeric defect
and is close to a no-op on this repository's evidence. *Re-aim at verifying the `@ <ref>` half of a
`measured:` tag* — this is what would have caught the ticket's most-instanced defect, but it amends
the deliberate, documented boundary that the guard never checks truth. *Both together* — the largest
scope, and one that wants decomposing into two changes rather than bundling.

### A change's `specs/` directory stays out of scope

**ID:** `specs-excluded`
**Status:** active
**Chosen:** Exclude `specs/` entirely. Spec text legislates rather than describes, and the only
measured hit across every archived delta spec is the quotation false positive.
**Considered:** *Include it* — immediately raises why `openspec/specs/`, the synced copy of the same
text, stays unscanned, a question with no good answer. *Include for the block rule only* — invents a
per-file rule split to save nothing measurable.

### The quoted-number false positive is fixed syntactically, not with a fifth tag

**ID:** `quotation-syntactic`
**Status:** superseded by `quotation-quotes-only`
**Chosen:** A number enclosed in a delimiter pair on its own line is a quotation, not a claim.
Mechanical, no vocabulary change, and an exemption that applies automatically never tempts anyone
toward a suppression marker.
**Considered:** *A fifth tag, `quoted:<source>`* — auditable, and nothing escapes silently, but it
amends "four tags exist, and no others" in both the contract and the spec to cover instances that
are all the same self-referential quotation. *A hybrid* — strictly more surface to specify and test,
for a case that has not yet arisen.

### The delimiter set is double quotes only; inline code spans are not delimiters

**ID:** `quotation-quotes-only`
**Status:** active
**Chosen:** Matched double quotes — straight (U+0022) and curly (U+201C/U+201D) — are the whole
delimiter set. A number written inside backticks is an ordinary unattributed claim and is reported.
Every measured false positive that ever justified this exemption is quote-delimited: across the
archived changes four exemptions fire and all four have the form `"194 tests"`, and not one needed a
code span. The code-span class was added during design without evidence, and it was the sole source
of five successive fail-opens — parity counting, overlapping span regions, an empty list conflating
"no spans" with "spans indeterminate", escaped delimiters paired as live ones, and backticks inside
an HTML comment treated as live delimiters. Deleting it removes the CommonMark inline surface those
five lived on. Three vetoes keep what remains fail-closed: a backslash-escaped delimiter, a
delimiter inside raw HTML, and a delimiter class that does not fully pair.
**Considered:** *Keep code spans and add a raw-HTML veto* — fixes the fifth escape and nothing else;
it keeps a CommonMark inline surface with contexts still unhandled (character references, autolinks,
link destinations), so it buys a sixth round rather than avoiding one. *Drop the exemption entirely*
— the smallest possible surface, rejected because the four measured false positives come straight
back, and a guard that over-fires on text doing its job is how suppression markers get written.

### The raw-HTML veto's trigger is a bare `<`, and the HTML grammar is deleted

**ID:** `quotation-angle-bracket-veto`
**Status:** active
**Chosen:** A line containing a `<` anywhere on it loses the quotation exemption. No tag grammar, no
comment/PI/CDATA/declaration alternatives, no attribute parsing — the veto never asks what the `<`
is part of or where that part ends, so there is nothing to approximate and nothing to leak. It is
trivially fail-closed and trivially auditable.

This replaces a per-construct regex that located §6.6 constructs in order to ask whether one carried
a delimiter. §6.6 permits a `>` inside a single-quoted attribute value, so a real tag can extend past
the first `>`; the tag alternative `</?[A-Za-z][^<>]*>` truncated there, the truncated prefix carried
no delimiter, and the veto cleared a line whose `"` was genuinely inside the tag:

```text verified:run against the pre-fix guard; it exited 0 with the claim unreported
<a b='>"'> the benchmark ran 77 tests "
```

A differential sweep against an accurate §6.6 grammar found 150 distinct fail-open arrangements of
that shape; the double-quoted attribute form self-vetoes, which is why it was missed for a wave.

The cost was measured before the rule was chosen, over every Markdown file in this repository (110 of
them): 38 numeric claims are exempted today, and exactly 2 sit on a line containing `<`. One is the
manual-test guide's own fixture for the behaviour being replaced; the other already carries a
provenance tag. Inside the guard's scan scope (`openspec/changes/*/{tasks,design,proposal}.md`) the
loss is zero. A differential fuzz over 346,632 lines
<!-- measured: /usr/bin/python3 .superpowers/sdd/fix-wave-4-fuzz.py, reported in .superpowers/sdd/fix-wave-4-report.md -->
built from a 44-token corpus — single- and double-quoted attribute values, `>` and `<` inside
attribute values, unterminated tags and comments, PIs, CDATA, declarations, bare `<`, interleaved
with straight quotes, curly quotes and backslash escapes — found 0 fail-open divergences against an
independent §6.6 oracle. The same corpus finds 443 against the pre-fix code, which is what makes the
0 evidence rather than an artifact of a token set that cannot express the defect.

**Considered:** *Fix the tag grammar* — extend the tag alternative to handle single-quoted attribute
values correctly. Rejected: this would have been the fifth consecutive attempt to approximate a spec
grammar with a regex, and grammar approximation is the documented root cause of every one of this
exemption's six fail-opens. Six failures is sufficient evidence that the approach does not work;
another round buys a seventh rather than avoiding one, and the measured cost of not needing a grammar
at all is essentially zero. *Adopt a real CommonMark parser* — no `cmark`/`pandoc`/markdown package is
installed anywhere on this machine and this repository has no dependency management, so it would
trade a known bug class for an unguaranteed runtime (see the guard's own module docstring).

## Scan scope

Three files per non-archived change directory: `tasks.md`, `design.md`, `proposal.md`. Both rules
apply to all three. `specs/**` is not scanned.

The archive exclusion needs no work. It is applied at the change-directory level, not per file, so
every file type added to the scan set inherits it automatically — the ticket's concern that it
"probably has to extend" does not arise.

Three mechanical consequences in the scan loop:

1. **Containment generalises.** The containment gate was per-file and named for `tasks.md`; it
   becomes `verify_scanned_file_containment`, taking the filename as a parameter. It must run
   against each candidate, or a `design.md` symlink walks past the check that exists to stop exactly
   that. The containment exit code keeps its current behaviour and its rank above every
   other code: refuse that one candidate, keep scanning.
2. **The scan-integrity invariant is restated.** "Change directories exist but none produced a
   `tasks.md`" becomes "none produced any candidate", so a change mid-planning cannot trip a failure
   that means nothing.
3. **The scanned-file counter and the success line** stop naming `tasks.md` specifically.

Exit codes, their ranking, and the behaviour where a classification abort stops only the current
file are all untouched.

## Classifier changes

**Issue-key adjacency** is a bug fix: `CLAIM_RE` gains a fixed-width negative lookbehind rejecting a
digit run preceded by an uppercase letter and a hyphen. The docstring sentence describing this
exclusion is corrected to describe what the code then actually does.

**The quotation exemption** is new behaviour. A match falling inside a delimiter pair on its own
line is a quotation:

- **Delimiters:** matched straight double quotes and matched curly double quotes (`U+201C`/`U+201D`),
  and nothing else. Curly quotes do not occur in the corpus today, but a curly-quoted number is the
  identical false positive and covering it costs a character class rather than a later fix. Inline
  code spans were a third class as first shipped and were removed — decision
  `quotation-quotes-only`, which supersedes `quotation-syntactic`.
- **Single quotes and apostrophes are excluded.** Prose apostrophes are unpaired by nature, so
  admitting them would make pairing meaningless across most lines of this repository's prose.
- **Line-scoped only**, keeping the exemption a line-local predicate rather than a second piece of
  parser state that must agree with the fence tracker.
- **Fail closed, through three vetoes.** A backslash-escaped delimiter (CommonMark §2.4) and a `<`
  anywhere on the line each withdraw the exemption from the whole line; a delimiter class whose
  delimiters do not all pair yields no region for that class. Both whole-line vetoes are deliberately
  coarse, and the second is now as coarse as a rule can be — it asks only whether a `<` is present,
  never what the `<` is part of — because approximating a spec grammar with a regex is what produced
  all six fail-opens (decision `quotation-angle-bracket-veto`).
- **The guard says which veto fired.** A withdrawn exemption prints a note naming the cause, the
  remedy and the contract, and the exit-1 banner offers rewording alongside tagging. Without it an
  author whose correctly-quoted number lost its exemption to an unrelated stray delimiter would
  write a `measured:` tag that is not true.
- **No fence interaction.** Inside an open fence the scan loop tests only for the closing line and
  never consults the numeric rule, so in-fence numbers are already never claim-checked.

Together these take the widened scope's hits down to exactly the genuine ones.

**Accepted cost:** a real claim written inside quotes escapes the guard. Weighted low because the
failure this contract exists to prevent was an asserted baseline, not a quoted one. The mirror cost
is a loud one: a number inside backticks, or beside a stray/escaped/HTML-embedded delimiter, is
reported even when a strict CommonMark reader would exempt it. Measured over every tracked Markdown
file, eight lines lose an enclosure to a veto and seven of them are counter-examples written on
purpose; the eighth sits in an archived change the guard does not scan.

## Documentation

`skills/myflow-contracts/plan-provenance.md` has its scope section rewritten for the three-file set,
with the over-firing rationale re-pointed: it now justifies excluding `specs/` and requiring the
exemption, rather than excluding `design.md` and `proposal.md`. New subsections document the
exemption, why its delimiters are double quotes only, and each of its three vetoes with the
measured cost of all three.

`scripts/check-plan-provenance.py`'s module docstring has its scope paragraph and its numeric-rule
paragraph corrected.

The delta spec carries the normative changes.

## Bootstrapping

Once implemented, the guard scans this change's own `design.md` and `proposal.md`, and `/myflow-do`
runs the lint step at the end of its run. These artifacts are written tag-clean under the new rules
from the start, which is also the most direct end-to-end proof that the widening works.
