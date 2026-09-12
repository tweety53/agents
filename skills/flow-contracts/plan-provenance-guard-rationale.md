# Plan provenance — the guard — rationale

This file is the reasoning behind `skills/flow-contracts/plan-provenance-guard.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## The guard's scope, and why it is narrow

## The quotation exemption

### Why the delimiters are double quotes only

The exemption shipped with a second delimiter class, CommonMark inline code spans. That class then
failed **open** five times, each caught by a different reviewer: parity counting instead of
nearest-enclosing-pair; overlapping span regions, because interior runs were not consumed; an empty
list conflating *no code spans* with *code-span state indeterminate*; backslash-escaped delimiters
paired as live ones; and backticks inside an HTML comment treated as live delimiters. All five lived
in code-span handling, because that class had turned a provenance guard into a re-implementation of
a CommonMark **inline** parser — and CommonMark has further inline contexts still unhandled
(character references, autolinks, link destinations), so a sixth was a matter of time.

What settled it was a measurement rather than an argument: every false positive that ever justified
this exemption is double-quote delimited. Across the archived changes, four exemptions fire and all
four have the form `"194 tests"`. Not one needed a code span. The class was added during design
without evidence, and it was the sole source of all five escapes, so it was removed rather than
patched a sixth time.

That removal has a visible cost, and it is a **loud** one: a number you wrote inside backticks is
now reported. Quote it, reword it, or tag it.

Both remaining whole-line vetoes are deliberately **coarse** — a construct carrying a delimiter
refuses the whole line rather than resolving which characters inside it are markup. Coarseness is
the lesson of the five failures, not an oversight: a veto's only possible output is the absence of
an exemption, so the worst it can do is report a number a stricter reader would have let through.

### The class-wide veto

Fail-closed has a consequence worth stating plainly, because it is the one an author actually hits.

**If a delimiter class's delimiters do not all pair on a line, that class yields no region on that
line at all.** Not "the unpaired one is ignored" — the whole class is vetoed for that line. So a
genuinely closed, unambiguous quotation can lose its exemption because an unrelated stray delimiter
of the same class appears somewhere else on the same line.

This is deliberate. Pairing delimiters in order, *without* the class-wide veto, still admits a bare
asserted number: a stray delimiter simply pairs with a later quotation's opener and manufactures a
region that encloses the number sitting between them. On a line like

```markdown verified:the shape the exemption's fail-closed rule is written against
a "quote" and a stray " mark, then 99 tests ran, "done"
```

the intended pairing is genuinely undetermined — the stray delimiter could belong on either side of
the number — and the fail-closed answer to an undetermined line is to report the claim.

The trade is a **loud false positive**, fixable by rewording the line or balancing the delimiter,
in place of a **silent false negative**, which nothing would ever surface.

### The escape veto

A backslash-escaped delimiter is a *literal character*, not markup. Pairing one as though it were
markup invents a region that does not exist, and a number sitting in open prose inside that invented
region is silently exempted. On

    Use \" for a literal quote; the run reported 99 tests, then printed \" done

there is no quotation at all — both quote marks are literal — yet the claim was exempted until this
was fixed.

**A line containing any backslash-escaped delimiter is vetoed outright: no class exempts anything on
it.** Not "the escaped one is skipped". Refusing the whole line is what makes the veto fail-closed by
construction rather than by argument: its only output is the absence of an exemption, and an absence
can only ever withdraw a finding, never grant one.

The visible cost is one shape: a quotation that pairs perfectly loses its exemption because an
escaped delimiter appears somewhere else on the same line. Reword it or tag it; that is the
loud-false-positive side of the same trade as above.

Counting backslashes matters: `\\` is an escaped **backslash**, so the delimiter after it is live
and still opens a quotation. An odd run of backslashes escapes what follows, an even run does not.

### The angle-bracket veto

**A line containing a `<` anywhere on it gets no quotation exemption.** That is the whole rule. Not
"a `<` that opens a tag", not "a construct carrying a delimiter" — a `<`.

The reason it is that blunt is the reason it exists. CommonMark §6.6 passes raw HTML through
untouched: an HTML comment, a tag, a declaration, a CDATA section and a processing instruction are
all opaque, and nothing inside one is parsed as an inline delimiter. A delimiter written inside one
is therefore literal content, and pairing it with a live delimiter elsewhere invents a region exactly
as an escape does:

    See <!--"--> the benchmark ran 77 tests <!--"-->.

where the two quote marks are comment text and `77 tests` stands in open prose. Deciding that
correctly means knowing where each construct starts and ends, and the guard used to work that out
with a regex per construct. §6.6 permits a `>` inside a single-quoted attribute value, so a tag can
run past the first `>` — and

    <a b='>"'> the benchmark ran 77 tests "

exited 0 with the bare claim unreported, one of 150 arrangements of that shape. It was the sixth
fail-open this exemption has had and the fifth caused by approximating a spec grammar with a regex,
so the grammar was deleted rather than improved. Nothing is approximated now, so nothing can leak.

The cost you will actually meet: a line that merely mentions a `<` — an ordinary tag, an
`<!-- measured: … -->` comment, or a mathematical `a < b` — withdraws its own line's exemption, even
though nothing on it is ambiguous to a human reader. That is a finding reported where a strict reader
would have exempted it: loud, visible, and the only direction this exemption is allowed to be wrong
in. Reword the line or drop the `<`; do not reach for a suppression marker.

### What the vetoes cost, measured

This document does not get to assert a number it cannot show you how to re-derive. The measurement
compares, for every `CLAIM_RE` match on every line of every tracked Markdown file, what the guard
exempts against what the *same* left-to-right pairing would have exempted with every veto removed.
Every match in the second set and not the first is an enclosure a veto withdrew:

```python verified:run against this tree; the twelve hits below are its output
# Load the guard, then re-implement its quote scanner with the vetoes removed:
# escapes and raw HTML are ignored, a second opener simply re-arms `pending`,
# and a leftover `pending` is dropped instead of vetoing the class.
# Then, for each tracked .md file (`git ls-files '*.md'`) and each line:
for m in CLAIM_RE.finditer(line):
    strict = _is_quoted(m.start(), m.end(), quotation_regions(line))
    lax    = any(_encloses(m.start(), m.end(), r) for r in lenient_classes(line))
    if lax and not strict:
        print(path, lineno, m.group(0))   # an enclosure a veto withdrew
```

Lines in this repository lose an enclosure that way. All but one are counter-examples written down
on purpose — the shapes this document quotes in order to explain each
veto — which is what it means for a counter-example to be one: it demonstrates the shape its veto
exists to catch, so being withdrawn is the demonstration working.

| Line | Veto | What it is |
|------|------|-----------|
| this file, an archived design document, or an SDD ledger | class-wide, escape, angle-bracket | the shapes each veto exists to catch, quoted on purpose (eleven lines) |
| an archived `tasks.md` table row whose trailing `<!-- measured: … -->` comment quotes `"197 tests"` | angle-bracket | a genuine quotation, withdrawn because its line carries a `<` |

The last is the honest cost of the angle-bracket veto. Its quote characters pair perfectly; what
withdraws the exemption is the `<` that opens the comment they sit in. It lives in an **archived**
change, which the guard does not scan, so no scanned file loses an exemption today.

The same measurement taken the other way round — how much the coarse rule costs over what a full
CommonMark reader would exempt — needs a genuine §6.6 raw-HTML boundary detector to answer, which
this repository does not have and could not pin (the same reason `check-plan-provenance.py`'s
`CLAIM_RE` comment gives for staying regex-based rather than reaching for a real parser). No number
is stated here for that reason: the cost is real — a line that merely mentions a `<` unrelated to
any actual delimiter loses its exemption regardless — but it is rare, and rewording the line is the
fix whenever you hit one. Do not reach for a suppression marker, and do not weaken the guard.

## What the guard does not do
