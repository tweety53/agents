# Plan provenance

**This file is canonical for plan provenance.** Skills and guards reference it by name; none of
them restate the tag vocabulary. If a rule below and a skill ever disagree, this file wins.

## The four tags

Every fenced code block in a plan carries a provenance tag on its info string, and every numeric
claim carries one in an HTML comment within the two lines after it:

- `verified:<how>` — on a fenced block's info string. The snippet was checked against something
  real; `<how>` names the check, e.g. `verified:javap intellij.platform.diff.jar`.
- `unverified:<what-to-check>` — on a fenced block's info string. The snippet is the implementer's
  best guess, not a checked fact; `<what-to-check>` names exactly what to confirm before trusting
  it, e.g. `unverified:confirm the member is a property, not a function`.
- `measured:<command> @ <ref>` — in an HTML comment within the two lines after a numeric claim
  (the claim's own line, the line after it, or the line after that — so a blank note line between
  claim and comment is tolerated). The number came from actually running `<command>` at `<ref>`
  (a commit, tag, branch, or other named point), e.g. `<!-- measured: ./gradlew test @ c515c42 -->`.

  **The guard checks only that a `measured:`/`predicted:` comment is present.** It does not parse
  the `<command>` or the `@ <ref>` half — so that shape is a *convention this contract states and a
  human enforces*, not something the script can hold you to. It is written down here as the shape
  to follow, and named as advisory so the contract is not defined two ways: a plan whose comment
  omits the ref passes the guard and is still wrong under this contract.

  **Choosing a ref while the work is uncommitted.** A plan under `/myflow-do` sits on a branch whose
  commits do not exist yet, so naming the merge base is worse than useless: the commands being cited
  frequently do not exist there, and `git cat-file -e <merge-base>:<script>` fails outright. Name
  the **branch** (`@ branch openspec/<change-name>`), which resolves both while the work is in
  flight and after it merges. Name a commit only when the measurement really was taken at that
  commit and the command really does exist there — and say which, as in
  `@ merge-base d38372a (the count BEFORE this change)`. When a measurement genuinely cannot be
  re-run — it read a machine-local file, or observed a live event — say that in the tag instead of
  naming a ref that implies otherwise.
- `predicted:<what-confirms-it>` — in an HTML comment within the two lines after a numeric claim,
  same window as `measured:` above. The number is an expectation, not a measurement;
  `<what-confirms-it>` names what would confirm it, e.g.
  `<!-- predicted: ./gradlew test after task 1 -->`.

## Tag syntax examples

````markdown verified:authored in-tree for this change
```kotlin verified:javap intellij.platform.diff.jar
override val toolWindowIds: Array<String>
```

```kotlin unverified:confirm the member is a property, not a function
override fun getToolWindowIds(): Array<String>
```

Baseline: 197 tests, 0 failures
<!-- measured: ./gradlew test @ c515c42 -->

After the deletion: 186 tests
<!-- predicted: ./gradlew test after task 1 -->
````

## The asymmetry rule

Every fenced code block in a plan needs a tag — `verified:<how>` or `unverified:<what-to-check>` —
but a number with no tag may not appear in a plan **at all**. A block can be labelled `unverified`
and stay in the plan, because it is honest about its own uncertainty. A number has no such escape
hatch: an untagged number is not "unverified", it is unattributed, and an implementer reading it
cannot tell whether it was run or guessed.

**Why the asymmetry, and not just "label everything":** in a prior run (KAN-6, an IntelliJ
plugin), a plan claimed "194 tests" as a baseline. That number was invented, not mislabelled —
nobody ran the suite and wrote down a stale count; the figure never came from a run at all.
Labelling alone would not have caught it, because a fabricated number can be labelled exactly as
confidently as a real one. What catches it is refusing the number a home unless the label names a
command and a ref that can be re-run — `measured:<command> @ <ref>` forces the claim to point at
something checkable, and `predicted:<what-confirms-it>` forces it to name the check that has not
happened yet. A plan that cannot produce either for a number should not state the number.

The same run also carried the motivating failure for the block rule: a plan snippet said a marker
key could be read from `DiffRequest`. It cannot — the platform never propagates chain user data to
requests. Had the implementer transcribed that snippet verbatim, the feature's guard would have
been permanently false and the whole feature a no-op **that every unit test still passed** — a
false guard produces no failing assertion, only a feature that silently never fires. It was caught
only because that one task happened to carry a hand-written instruction telling the implementer to
check. The tag vocabulary exists so that instruction is never a special case again: every snippet
either says how it was checked, or says plainly that it was not.

## The guard's scope, and why it is narrow

The guard that enforces this contract reads three files per non-archived change — that change's own
`tasks.md`, `design.md` and `proposal.md` — and explicitly excludes `openspec/changes/archive/`. It
does not scan the whole repository, and it does not scan other changes' plans. Both rules apply to
all three files identically: every fenced block needs `verified:`/`unverified:`, and no untagged
number may appear in any of them.

All three make claims **about the world**, and a claim is no more attributable for sitting in the
design than in the tasks. Scanning only `tasks.md` left the two artifacts the tasks are *derived
from* unchecked — which is how an invented baseline can enter a plan through the door the guard
does not watch and be transcribed into the tasks as settled fact.

**`specs/` under a change is excluded**, and the reason is a difference in kind rather than a
difference in risk. Spec text **legislates** rather than describes: a requirement that forbids a
shape has to be able to name that shape, so an untagged fence in a spec is usually the requirement
doing its job, not an unattributed claim. Widening into `specs/` would therefore buy hits that are
correct by construction — the definition of over-firing.

That is where the over-firing history bears. `scripts/check-references.sh`'s own header records
that an earlier guard in this repository, once widened past the shape it was designed for, produced
28 false failures on this repo's own tree — none of them a genuine defect. The only way to silence
a false failure under time pressure is a suppression marker, and a suppression marker placed to
silence a false hit also switches off whatever real check shares that line. A guard that over-fires
does not just annoy; it manufactures the conditions for its own defeat.

Two things follow from that history, and they are the reason this scope is what it is rather than
merely narrow for its own sake:

1. **`specs/` stays out**, because a rule that fires on text doing its job cannot be fixed by the
   author — there is nothing to correct.
2. **The widening removed its own false-positive classes first.** Two exemptions landed *before*
   the scan set grew: an issue key followed by a unit word (`KAN-6 errors`) is an identifier and
   not a quantity, and a number reproduced inside a quotation is not a number asserted. The extra
   prose the guard now reads is prose it can already classify without over-firing.

## The quotation exemption

Design and proposal prose quotes things — including the invented baseline this contract exists to
explain. A number that a line **reproduces** rather than **asserts** needs no tag, so:

- **A number inside a matched pair of double quotes is a quotation.** Straight (`"`) and curly
  (`“ ”`) both count, and they are the **whole** delimiter set.
- **A code span is not a delimiter.** A backtick is an ordinary character here, so a number written
  `` `85 lines` `` is an ordinary unattributed claim and is reported. Inline code spans *were* a
  delimiter class and were removed — see
  **Why the delimiters are double quotes only** (`plan-provenance.md`), below.
- **Single quotes and apostrophes are not delimiters.** Prose apostrophes are unpaired by nature,
  so admitting them would make pairing meaningless on most lines this repository writes.
- **A backslash-escaped delimiter is not a delimiter, and its line is exempted from nothing.**
  CommonMark §2.4 makes `\"` a literal quote, so it can neither open nor close anything. A line
  carrying one loses the exemption **entirely** — see below.
- **A line containing a `<` is exempted from nothing.** CommonMark §6.6 passes an HTML comment, tag,
  declaration, CDATA section or processing instruction through untouched, so a `"` inside one is
  literal content — and working out which characters those are is a grammar that has leaked six
  times. A `<` anywhere on the line therefore withdraws the exemption **entirely**, whatever the `<`
  turns out to be — see below.
- **The exemption is line-scoped.** A pair that opens on one line and closes on the next is not
  recognised. This keeps the exemption a pure line-local test rather than a second piece of parser
  state that would have to stay in agreement with the fence tracker.
- **It fails closed.** The exemption can only ever *remove* a finding, so anything it gets wrong is
  a claim the guard silently accepts. When the enclosure is not certain, the number is reported.

Its purpose is narrow and specific: it lets this contract's own documentation, and the plans that
discuss it, reproduce the invented baseline they exist to explain without having to attribute a
figure nobody ever measured.

**When a veto costs you an exemption, the guard tells you which one.** Alongside the ordinary
`numeric claim with no measured:/predicted: provenance comment` line it prints a note naming the
veto that fired and what to do about it, and the exit-1 remedy banner offers rewording as well as
tagging. Do not answer a veto by writing a `measured:` tag for a number that was already correctly
quoted — that is the tag vocabulary being used untruthfully to silence a guard, and it is worse
than the finding.

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

### The class-wide veto — the property that will surprise you

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

### The escape veto — the second thing that will surprise you

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

### The angle-bracket veto — the third thing that will surprise you

**A line containing a `<` anywhere on it gets no quotation exemption.** That is the whole rule. Not
"a `<` that opens a tag", not "a construct carrying a delimiter" — a `<`.

The reason it is that blunt is the reason it exists. CommonMark §6.6 passes raw HTML through
untouched: an HTML comment, a tag, a declaration, a CDATA section and a processing instruction are
all opaque, and nothing inside one is parsed as an inline delimiter. A delimiter written inside one
is therefore literal content, and pairing it with a live delimiter elsewhere invents a region exactly
as an escape does:

    See <!--"--> the benchmark ran 77 tests <!--"-->.

where the two quote marks are comment text and `77 tests` stands in open prose. Deciding that
correctly means knowing where each construct STARTS AND ENDS, and the guard used to work that out
with a regex per construct. §6.6 permits a `>` inside a single-quoted attribute value, so a tag can
run past the first `>` — and

    <a b='>"'> the benchmark ran 77 tests "

exited 0 with the bare claim unreported, one of 150 arrangements of that shape. It was the sixth
fail-open this exemption has had and the fifth caused by approximating a spec grammar with a regex,
so the grammar was deleted rather than improved. Nothing is approximated now, so nothing can leak.

The cost you will actually meet: a line that merely MENTIONS a `<` — an ordinary tag, an
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

Twelve lines in this repository lose an enclosure that way. Eleven are counter-examples written down
on purpose — the shapes this document and the change's delta spec quote in order to explain each
veto — which is what it means for a counter-example to be one: it demonstrates the shape its veto
exists to catch, so being withdrawn is the demonstration working.

| Line | Veto | What it is |
|------|------|-----------|
| this file (five lines), the live delta spec (`openspec/specs/myflow-plan-provenance/spec.md`, two lines), an archived copy of that spec (two lines), an archived design document (one line), or an SDD ledger (one line) | class-wide, escape, angle-bracket | the shapes each veto exists to catch, quoted on purpose (eleven lines) |
| `openspec/changes/archive/2026-07-29-kan-14-plan-provenance/tasks.md`, a table row whose trailing `<!-- measured: … -->` comment quotes `"197 tests"` | angle-bracket | a genuine quotation, withdrawn because its line carries a `<` |

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

The guard checks that provenance is **stated**: every code block carries `verified:` or
`unverified:`, every number is followed by `measured:` or `predicted:`. It does not, and cannot,
check that the stated provenance is **true**. `verified:javap intellij.platform.diff.jar` passes
the guard whether or not `javap` was actually run — no script can confirm a verification was
performed, only that a claim of one was written down. The guard converts "silently unverified"
into "loudly unlabelled or falsely labelled"; only a human reviewing the plan's own claims can tell
labelled-and-true from labelled-and-false.

**It does not verify a `measured:` tag's `@ <ref>`** — or its `<command>`, or that either exists at
all. The guard's whole test is that a `measured:`/`predicted:` comment is *present* in the window
after the claim. It never resolves the ref, never checks the command exists at it, and never runs
anything. The `<command> @ <ref>` shape is a convention this file states and a human enforces; a
comment that omits the ref entirely still passes.

Widening the scan to `design.md` and `proposal.md` bought **more files checked for stated
provenance**. It bought no truth-checking whatsoever, in any of them. Do not read the wider scope as
a stronger guarantee: the guarantee is the same one, applied to two more files.

**It does not always scan a file to the end.** A fence-like run of backticks/tildes that the
guard's container-prefix grammar cannot resolve — behind a prefix shape it does not recognise, a
bare line indented 4+ columns with no container syntax of its own, or a line whose own container
prefix starts 4+ columns in (where CommonMark reads the whole line as an indented code block and
the marker on it as literal text) — makes the guard refuse to
guess whether that line opens or closes a fence (exit code 4). That refusal stops scanning **that
one file** at the line it could not classify: no line after it in that file is scanned on that run.
The scan is not abandoned — every other change directory keeps being scanned, and a real violation
found anywhere still outranks this exit code — but a genuinely unattributed claim sitting after the
unresolved line will not be reported until the unresolved line is fixed and the guard re-run. This
is disclosed here because it is a property of what the guard *does*, not merely of its exit codes:
treat exit 4 as "fix this line, then run it again," not as "this file has no other problems."

**It refuses to open a planning artifact it cannot trust the path of (exit code 3).** Before
reading any of a change's three scanned files, the guard confirms that file really sits at the
plain, expected `openspec/changes/<name>/<filename>` path: the right basename, not a symlink, not
reached through a change directory that resolves outside `openspec/changes/`, and not some other
non-regular file (a FIFO, a device node) sitting where a plan should be. Any of those shapes is a
containment refusal, not an ordinary violation — a PR-controlled `design.md` that is actually a
symlink to something else could otherwise make the guard read (or hang reading) content its author
never intended to be scanned as this change's plan.

A containment refusal stops the scan only for **that one candidate file** — the refused file's own
siblings in the same change directory are still scanned, every other change directory is still
scanned, and every violation, classification abort, or unreadable file found anywhere is still
reported — but exit code 3 **outranks every other exit code this guard can produce** (1, 2, and 4
included), even when one of those was also found elsewhere in the same run.
A symlink escape must never be downgraded to a mere violation or environment code just because an
unrelated tag was also missing somewhere else; that is exactly the shape that would let a real
escape hide behind an unrelated, easily-fixed nit and get lost in the noise.

**What an operator should do on exit 3:** treat it as a security finding first, not a provenance
nit. Read the reported path(s) — the message names exactly which file failed containment, and why
(wrong basename, symlink, directory escape, or non-regular file) — and fix the containment problem
itself (replace the symlink with a real file, correct the directory structure) before any other
violation or abort also listed in the same run's output. The other findings remain true and still
need fixing, but they are not why the process exited non-zero this time.

## The implementer's duty

Because the guard cannot verify truth, the obligation falls on whoever writes the tag: write
`verified:<how>` only after actually performing `<how>`, and write `measured:<command> @ <ref>`
only after actually running `<command>` at `<ref>` and reading the result. Writing either tag
without doing the check it names is worse than leaving the block `unverified` or the number
`predicted` — it tells the next reader a check happened when it did not, which is exactly the
failure this contract exists to prevent.
