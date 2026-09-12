# Plan provenance — the guard

**This file is canonical for what `<agents repo>/scripts/check-plan-provenance.py` enforces.** The tag
vocabulary itself is canonical in **The four tags** (`skills/flow-contracts/plan-provenance.md`).

## The guard's scope, and why it is narrow

The guard that enforces this contract reads three files per non-archived change — that change's own
`tasks.md`, `design.md` and `proposal.md` — and explicitly excludes `<project>/spectre/changes/archive/`. It
does not scan the whole repository, and it does not scan other changes' plans. Both rules apply to
all three files identically: every fenced block needs `verified:`/`unverified:`, and no untagged
number may appear in any of them.

All three make claims **about the world**, and a claim is no more attributable for sitting in the
design than in the tasks. Scanning only `tasks.md` left the two artifacts the tasks are *derived
from* unchecked — which is how an invented baseline can enter a plan through the door the guard
does not watch and be transcribed into the tasks as settled fact.

**`<project>/spectre/specs/` is excluded**, and the reason is a difference in kind rather than a
difference in risk. Spec text **legislates** rather than describes: a requirement that forbids a
shape has to be able to name that shape, so an untagged fence in a spec is usually the requirement
doing its job, not an unattributed claim. Widening into `<project>/spectre/specs/` would therefore buy hits that are
correct by construction — the definition of over-firing.

That is where the over-firing history bears. `<agents repo>/scripts/check-references.sh`'s own header records
that an earlier guard in this repository, once widened past the shape it was designed for, produced
28 false failures on this repo's own tree — none of them a genuine defect. The only way to silence
a false failure under time pressure is a suppression marker, and a suppression marker placed to
silence a false hit also switches off whatever real check shares that line. A guard that over-fires
does not just annoy; it manufactures the conditions for its own defeat.

Two things follow from that history, and they are the reason this scope is what it is rather than
merely narrow for its own sake:

1. **`<project>/spectre/specs/` stays out**, because a rule that fires on text doing its job cannot be fixed by the
   author — there is nothing to correct.
2. **Two exemptions keep the wider scan from over-firing on prose.** An issue key followed by a
   unit word (`KAN-6 errors`) is an identifier, not a quantity; and a number reproduced inside a
   quotation is not a number asserted.

## The quotation exemption

Design and proposal prose quotes things — including the invented baseline this contract exists to
explain. A number that a line **reproduces** rather than **asserts** needs no tag, so:

- **A number inside a matched pair of double quotes is a quotation.** Straight (`"`) and curly
  (`“ ”`) both count, and they are the **whole** delimiter set.
- **A code span is not a delimiter.** A backtick is an ordinary character here, so a number written
  `` `85 lines` `` is an ordinary unattributed claim and is reported. Inline code spans *were* a
  delimiter class and were removed — see **Why the delimiters are double quotes only**
  (`plan-provenance-guard-rationale.md`).
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

**Each veto is whole-line and coarse.** A line carrying an unpaired delimiter of a class, a
backslash-escaped delimiter, or a `<` anywhere on it gets no exemption from any class — not "the
offending character is skipped". The cost is a loud false positive, fixable by rewording the line or
balancing the delimiter; the alternative is a silent false negative that nothing surfaces. Counting
backslashes matters: an odd run escapes what follows, an even run does not. Reword or tag; do not
reach for a suppression marker, and do not weaken the guard.

See **Why the delimiters are double quotes only**, **The class-wide veto**, **The escape veto** and
**The angle-bracket veto** (`skills/flow-contracts/plan-provenance-guard-rationale.md`) for the
failures each rule settles and what they cost, measured.

## What the guard does not do

The guard checks that provenance is **stated**: every code block carries `verified:` or
`unverified:`, every number is followed by `measured:` or `predicted:` — in all three files alike.
It does not, and cannot, check that the stated provenance is **true**.
`verified:javap intellij.platform.diff.jar` passes the guard whether or not `javap` was actually run — no script can confirm a verification was
performed, only that a claim of one was written down. The guard converts "silently unverified"
into "loudly unlabelled or falsely labelled"; only a human reviewing the plan's own claims can tell
labelled-and-true from labelled-and-false.

**It does not verify a `measured:` tag's `@ <ref>`** — or its `<command>`, or that either exists at
all. The guard's whole test is that a `measured:`/`predicted:` comment is *present* in the window
after the claim. It never resolves the ref, never checks the command exists at it, and never runs
anything. The `<command> @ <ref>` shape is a convention this file states and a human enforces; a
comment that omits the ref entirely still passes.

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
plain, expected `<project>/spectre/changes/<name>/<filename>` path: the right basename, not a symlink, not
reached through a change directory that resolves outside `<project>/spectre/changes/`, and not some other
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

