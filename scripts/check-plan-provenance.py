#!/usr/bin/env python3
"""check-plan-provenance.py — fail when a live change's planning
artifacts carry a fenced code block or a numeric claim with no stated
provenance.

Exit codes:
  0  clean — every scanned file states its provenance (or there is
     nothing in flight to scan).
  1  violations found — one or more untagged fences or unattributed
     numeric claims. Fix the named file:line; never suppress.
  2  environment — the guard cannot determine anything about the tree:
     CHECK_PLAN_PROVENANCE_ROOT set but empty or not a directory,
     openspec/changes/ missing, or change directories exist but none of
     them yielded any of `SCANNED_FILENAMES` at the expected depth
     (scan-integrity). This last case is grouped under "environment"
     rather than given its own code: it means the root or the changes
     tree is misconfigured (a broken glob, a planning artifact at the
     wrong depth), the same class of "cannot determine" as a missing
     openspec/changes/ — not a malicious payload (that's 3) and not a
     parsing question (that's 4). Also covers a file that passed
     containment but cannot be read (permission denied, I/O error) or
     cannot be decoded as UTF-8: both are "cannot determine this file's
     content", the OS/encoding-level counterpart of a misconfigured root,
     not a security escape and not an ambiguous parse.
  3  containment/security — a symlinked planning artifact, a change
     directory that resolves outside its expected path, or a non-regular
     file where one of `SCANNED_FILENAMES` should be. The scan refuses to
     open something PR-author-controlled that reaches outside the tree it
     was told to scan. A containment refusal stops only THAT CANDIDATE
     (pass-11 fix wave, Critical 2 — the same shape an unreadable/
     undecodable file or a classification abort already use): the
     candidate's own SIBLINGS in the same change directory are still
     scanned, every OTHER change directory is still scanned, and every
     violation/abort/environment failure found anywhere is still
     reported alongside it — but exit 3 OUTRANKS ALL OF THEM (1, 2 and
     4) for the process's own exit code.
     Two review slots disagreed on this and both were right about half of
     it: exit 3 must not be downgraded just because a violation or an
     abort was ALSO found elsewhere (downgrading a symlink escape because
     an unrelated tag is missing would blind any CI that greps for 3),
     and the scan must not stop just because containment failed
     somewhere (a PR author could otherwise hide a later directory's own
     violations behind one containment escape sorted earlier). Both hold
     at once: keep scanning, keep reporting everything found, and let 3
     win the exit code.
  4  content-classification — a fence-like line this guard's grammar
     cannot resolve (see classify_line below): either it sits behind a
     container prefix the strip grammar does not recognise, or it is a
     BARE line (no container syntax of its own) that reduces to a fence
     run at 4+ columns of indentation, where whether a list is in scope
     cannot be determined without cross-line state this parser
     deliberately does not keep (pass-8 fix wave, Critical 1). Either way
     the guard refuses to guess whether the line opens or closes a fence
     rather than silently reading it as prose or as content. A
     classification abort stops only THAT file's scan (pass-8 fix wave,
     Critical 2 — the same shape an unreadable/undecodable file already
     used): every OTHER change directory is still scanned, and a real
     violation (exit 1) found anywhere always outranks this code,
     exactly as it already outranks an unreadable file (exit 2) — but a
     containment refusal (exit 3) outranks THIS code too, for the reason
     given above.

Full priority ladder, highest first: 3 (containment/security) > 1
(violations) > 4 (content-classification) > 2 (environment) > 0 (clean).
Every code below 3 in this ladder can still be found and printed
alongside a higher one that wins the exit — the ladder governs only
which single number the process exits with, never what gets reported.
A caller that treats "non-zero" uniformly (the only caller in this repo
today) is unaffected by any of this; the codes exist so a caller that
wants to distinguish — e.g. page on 3, merely log 2 — can, without
parsing prose.

Rule (canonical definition: skills/myflow-contracts/plan-provenance.md —
do not restate it here; a second copy is a Single Source of Truth
violation and is exactly how the two-line/one-line drift this guard once
had happened).

The motivating failure (KAN-6, an IntelliJ plugin): a plan claimed a
`DiffRequest` snippet could read a marker key. It could not — the
platform never propagates chain user data to requests. Had the
implementer transcribed that snippet verbatim, the feature's guard would
have been permanently false and the whole feature a silent no-op that
every unit test still passed — a false guard produces no failing
assertion, only a feature that never fires. It was caught only because
that one task happened to carry a hand-written instruction telling the
implementer to check. The tag vocabulary — and this guard — exist so
that instruction is never a special case again: every snippet either
says how it was checked, or says plainly that it was not, and every
number names the command that would confirm it.

Scope is `openspec/changes/*/{tasks,design,proposal}.md` — the three
artifacts named in `SCANNED_FILENAMES` — and never
`openspec/changes/archive/`. All three make claims about the world, and
a claim is no more attributable for sitting in the design than in the
tasks; scanning only tasks.md left the two artifacts the tasks are
derived FROM unchecked.

`specs/` under a change is still excluded, and the reason is a
difference in kind rather than in risk: spec text legislates rather than
describes, so a requirement forbidding a shape must be able to name that
shape, and an untagged fence there is the requirement doing its job.

check-references.sh's own header records that a guard once widened past
the shape it was designed for produced 28 false failures on this repo's
own tree, and the only way to silence a false failure under time
pressure is a suppression marker — one that then switches off whatever
real check shares that line. That is why this widening removed its own
false-positive classes FIRST: the issue-key exclusion in the numeric
rule and the quotation exemption below both landed before the scan set
grew, so the extra prose being read is prose this guard can already
classify without over-firing.

The numeric rule is intentionally narrow for the same reason: it
matches only a digit group directly followed by one of a small closed
set of unit words (tests, failures, errors, files, lines, seconds,
minutes, ms), and only when that digit group itself is not just the
tail end of some other token. So it does not match "step 1.2", "line
452", "IU-262", "KAN-14", "sha256", or a bare version number — the
first four because no unit word follows, "sha256" because its digits
are the tail of a token. A digit group directly preceded by an
uppercase letter and a hyphen is an issue key rather than a quantity,
and the leading lookbehind excludes it even when a unit word DOES
follow, so "KAN-6 errors" reads as an identifier and not as six errors.
None of these are numeric claims a plan needs to attribute, and
flagging them is exactly the over-firing failure mode described below.

=============================================================================
WHY THE QUOTATION EXEMPTION IS QUOTES-ONLY
=============================================================================

**This block is the single authoritative account of the quotation
exemption's history.** Every other docstring, comment and contract section
that needs it refers to it BY THIS HEADING and does not restate it. Four
rounds of edits previously had to keep six near-verbatim copies of the
narrative below in sync, which is the same Single Source of Truth defect
this guard's own subject matter is about.

A number a line REPRODUCES rather than ASSERTS needs no provenance tag, so
a number inside a matched delimiter pair is exempt. The exemption may only
ever REMOVE a finding, so every error it makes is a claim silently
accepted — it must fail closed.

It shipped with two delimiter classes: matched double quotes (straight and
curly) and CommonMark inline code spans. The code-span class then failed
OPEN five times, each caught by a different reviewer:

  1. parity counting instead of nearest-enclosing-pair;
  2. overlapping span regions — interior runs were not consumed, so a run
     CommonMark treats as content was reused as an independent opener;
  3. `[]` conflating "this line has no code spans" with "this line's
     code-span state is indeterminate";
  4. backslash-escaped delimiters paired as live ones;
  5. backticks inside raw HTML (`<!-- ... -->`) treated as live
     delimiters — `See <!--``--> the benchmark ran 77 tests ``.` exited 0
     with the bare claim unreported.

All five lived in code-span handling, because that class had made this
guard a re-implementation of a CommonMark INLINE parser — and CommonMark
has further inline contexts still unhandled here (character references,
autolinks, link destinations), so a sixth was a matter of time.

The evidence that ended it: every measured false positive that ever
justified this exemption is DOUBLE-QUOTE delimited. Measured over the
archived changes, four exemptions fire and all four have the form
`"194 tests"`. Not one needed a code span. The code-span delimiter was
added during design without evidence, and it was the sole source of all
five fail-opens, so it was removed outright (design decision
`quotation-quotes-only`).

The SIXTH arrived anyway, in the raw-HTML veto that replaced part of the
code-span class — not in what it decided, but in how it LOCATED §6.6
constructs: a per-construct regex whose tag alternative truncated at the
first `>`, where §6.6 permits a `>` inside a single-quoted attribute
value. `<a b='>"'> the benchmark ran 77 tests "` exited 0 with the bare
claim unreported, one of 150 fail-open arrangements a differential sweep
against an accurate §6.6 grammar found.

That is the same defect as the other five wearing different clothes: a
spec grammar approximated by a regex. Six failures is enough evidence
that the approach does not work, so the sixth fix DELETES the grammar
instead of improving it (design decision
`quotation-angle-bracket-veto`). Nothing is approximated, so nothing
leaks.

What ships, therefore:

  - **Delimiters are matched double quotes only** — straight (U+0022) and
    curly (U+201C/U+201D). A backtick is an ordinary character; a number
    inside `` `85 lines` `` is an ordinary unattributed claim.
  - **Three vetoes**, each of which can only ever WITHDRAW an exemption:
    a backslash-escaped delimiter (`_has_escaped_delimiter`), a `<`
    ANYWHERE ON THE LINE (`_has_angle_bracket`), and a delimiter class
    whose delimiters do not all pair (`_quote_regions`). The first two
    are whole-line; the third is per class.
  - **Both whole-line vetoes are COARSE on purpose**, and the raw-HTML
    one is now as coarse as a rule can be: it asks only whether a `<` is
    present, never what the `<` is part of or where that part ends. Every
    question it declines to ask is a question one of the six fail-opens
    answered wrongly.

=============================================================================
WHY THIS FILE IS PYTHON, NOT BASH — the reshape's own history
=============================================================================

This guard was originally a ~150-line Bash routine, in the same style as
this repository's other guards. Five review panel passes and seven fix
waves later, it had shipped and then fixed the following distinct defect
classes (the enumeration below is the count — this docstring does not
restate it as a number, because a number sitting next to its own
enumeration is exactly the kind of unsourced claim that drifts out of
sync with the list it describes, which is what happened here: an
earlier draft said "NINE" while the list below has always had ten
entries):

  1. column-0-only fence detection (indented fences invisible)
  2. non-whitespace prefixes invisible (blockquote/list fences invisible)
  3. task-checkbox lines swallowing every later violation in the file
  4. PROVENANCE_RE with no left boundary (a comment *denying* provenance
     satisfied the rule that grants it)
  5. prose over-firing on any non-letter-led line (an abort reachable on
     ordinary punctuation-led prose)
  6. the same over-firing again, one level deeper, on container-prefixed
     prose (a bullet or blockquote that merely *mentions* a fence run)
  7. an abort reachable on CONTENT lines inside an already-open,
     correctly tagged, correctly closed fence
  8. a quadratic DoS in blockquote-prefix stripping (bash regex re-run
     once per marker on a shrinking string: 0.32s at N=2000, 27.58s at
     N=16000)
  9. container scoping ignored on the CLOSING test — a fence opened with
     no container prefix could be falsely closed by unrelated content
     that merely reduced to a bare fence run after blind stripping
 10. a second, independent quadratic: bash 3.2 (what macOS ships) makes
     sequential indexed-array access O(N) per access, so reading a file
     into a bash array and walking it by index is O(N^2) overall
     (10k lines: 0.68s: 160k lines: 47.56s)

Pass 5 alone found three of these (7, 9, 10) on a SETTLED tree, after
four of seven review slots had already called the guard clean. Every
round's fix was structurally better than the last — the wave-3 inversion
in particular replaced an allowlist-of-known-prefixes (which can only
ever answer for prefixes someone already enumerated, so it shares
whatever blind spot the parser itself has) with a genuine grammar. It
still produced three more Criticals, because Bash gives a fence
classifier nowhere to put real block structure: no persistent container
stack, O(N) string algorithms that are actually O(N^2) under bash 3.2,
and ERE as the only pattern language.

The operator's call, recorded in full in design.md's "Post-review
reshape" section: stop patching and replace the classifier with a real
block-structure parser in Python 3 (standard library only — no
`cmark`/`pandoc`/markdown package is installed anywhere on this machine,
and this repository has no dependency management, so a library import
would trade a known bug class for an unguaranteed runtime). Python gives
this parser three things Bash structurally cannot:

  - a real container context, recorded per fence at OPEN time, so the
    closing test only ever strips the container syntax that fence was
    actually opened inside — this is the structural fix for defect 9,
    though "structural" describes the SHAPE of the fix, not a guarantee
    of correctness on every depth: an early version of this fix recorded
    only whether a blockquote prefix was present, as a boolean, and a
    fence opened two blockquote levels deep was falsely closed by a
    one-level-deep line, because both are "truthy". The context now
    records the blockquote depth as a COUNT and requires the close to
    match that count exactly (see FenceContext below);
  - O(1) list/string indexing, so reading and walking a file is linear
    regardless of size (fixes defect 10);
  - `re` as a real regex engine with anchors and bounded-repetition that
    behave identically whether the string is 10 characters or 10,000,
    so the same "one bounded match, not one match per marker" fix that
    closed defect 8 in Bash is simply how `re` already works here.

This history is kept, not deleted, because it is the evidence for why
this file exists in this form. The CLI contract above — exit codes, the
CHECK_PLAN_PROVENANCE_ROOT override, the scan scope, the containment
rules, the tag vocabulary — is frozen and byte-identical to what the
Bash version shipped; only the fence/container classifier underneath it
changed shape.

=============================================================================
STRUCTURE OF THE PARSER
=============================================================================

`classify_line()` is the only place that decides whether a line (when
NOT already inside an open fence) is prose, a fence opener, or an
unclassifiable prefix the guard must abort on. It works by attempting to
strip, from the line's start, exactly the container syntax this parser
understands — any number of blockquote markers, then at most one list
marker (bullet or ordered, simple or compound: `3.4.`, `1.2.3)`),
optionally followed by a task checkbox, then any number of further
blockquote markers — and asking whether what is left begins with a fence
run (three or more backticks or tildes).

A line is only ever a *candidate* for this in the first place when its
own start already attempts container syntax the stripper recognises (a
blockquote marker, a list marker, or a checkbox) or IS the fence run
itself. Prose that merely mentions a fence run later in the line —
whether it starts with a letter, a quote, a parenthesis, bold markdown,
or a container-prefix bullet/blockquote — is never a candidate, because
the fence run does not begin the line's content immediately after
whatever prefix was stripped. That is the fix for defect 6: the
distinction is where the fence run sits relative to the stripped prefix,
never what character led the line.

Among genuine candidates, stripping either reaches the fence run (it is
a fence, tag and all) or it does not — and if it does not, the abort
fires ONLY when what is left still looks like an attempted, unresolved
piece of container syntax (a second list marker glued behind the first,
a bare checkbox with no preceding list marker). If what is left is
ordinary prose (an unprefixed word), the line is prose, not an abort —
this is what lets container-prefixed prose ("- Confirmed the fence style
uses ``` markers.") pass cleanly while a doubled marker ("- - ```bash")
still aborts loudly, and it is checked against a bounded, constant
number of strip attempts per line (never a loop over the line's own
length), so there is no re-introduction of defect 8's shape here.

`check_file()` maintains the fence state as a `FenceContext` recorded at
OPEN time (which container tokens were consumed to reach the opening
run) and consults it, not `classify_line()`, for every line while a
fence is open. A content line is tested for closing by attempting to
strip ONLY the container tokens the context says were present at open
time, in the same order — never a blind, context-free strip. That is
the fix for defect 9: a fence opened with no container prefix accepts
only an unprefixed line as its close, so unrelated content that merely
reduces to a bare fence run after generic stripping (a quoted blockquote
fence run, as CONTENT, inside a bare-opened block) is correctly read as
content, never as a false close. The same context also records the
blockquote depth opened at, as a count, and requires the same count on
close — a fence opened two blockquote levels deep is not closed by a
one-level-deep line; see FenceContext and is_closing_line below.
`classify_line()`'s abort path is never
consulted for a content line (fixes defect 7): the only question inside
an open fence is the CommonMark close test, never "is this an
ambiguous opener".

KNOWN LIMITATION, carried forward unchanged from the Bash version and
still deliberate: three-or-more-level container nesting on a single
fence line (`- > - [ ] ```bash`) aborts (exit 4). The strip grammar has
exactly three slots (leading blockquote run, one list marker plus
optional checkbox, trailing blockquote run) because that is what the
existing fixtures and this repository's own plans need; a fourth slot
would require an unbounded per-level loop, and the Bash version's own
history (defect 8) is the reason that trade was rejected once already.
Verified absent from every planning artifact in this repository,
including archived ones — a real, loud, documented refusal on a shape
nothing here produces, not a silent gap.

KNOWN LIMITATION — containment/read TOCTOU: `verify_scanned_file_containment`
resolves and checks `path` and then `check_file` opens it separately; a
symlink swap between the two calls is a real, narrow race. Accepted, not
fixed: this guard's threat model is static PR-checked-out content
reviewed at a point in time, not a file system under concurrent hostile
mutation during the scan itself. Closing it would mean holding an open
file descriptor across both checks (`os.path.samestat` against an
`fstat`), which is real work for a race this scope does not need to
defend against.

KNOWN LIMITATION — read size: `check_file` refuses (exit 2) a scanned
file larger than `MAX_SCANNED_FILE_BYTES` before reading it, rather than
reading an arbitrarily large file into memory in one call. This is
defense in depth, not a realistic ceiling — every planning artifact in
this repository is a few hundred KB at most — chosen because the check
costs a few lines and an `os.path.getsize` call, not because an
unbounded regular file is a observed problem here.
"""

from __future__ import annotations

import bisect
import os
import re
import stat
import sys
from dataclasses import dataclass
from typing import Optional, Tuple

# MAX_SCANNED_FILE_BYTES — a defense-in-depth cap on how large a single
# planning artifact this guard will read into memory in one call. Every
# such file in this repository is a few hundred KB at most; this is not a
# realistic ceiling for ordinary use, it exists so an oversized file
# (however it got that way) is refused (exit 2) rather than read
# unbounded — see the module docstring's "KNOWN LIMITATION — read size".
MAX_SCANNED_FILE_BYTES = 10 * 1024 * 1024  # 10 MiB

# SCANNED_FILENAMES — the planning artifacts a change is scanned for, and
# the single source of truth for that set in CODE: the scan loop and the
# scan-integrity failure both derive from it rather than restating it.
#
# The module docstring's scope paragraph does NOT derive from it — it
# hardcodes the three filenames, because a docstring cannot interpolate a
# module constant. That is a restatement, and it is named as one here rather
# than claimed away: a fourth filename added to this tuple must be written
# into that paragraph by hand, and nothing mechanical will notice if it is
# not.
#
# `specs/` is deliberately absent. Spec text legislates rather than
# describes: a requirement forbidding a shape must be able to name that
# shape, so an untagged fence there is the requirement doing its job, not
# an unverified claim. tasks.md, design.md and proposal.md are the three
# artifacts that make claims ABOUT the world — the claims this guard
# exists to make attributable.
SCANNED_FILENAMES = ("tasks.md", "design.md", "proposal.md")

# --- Regexes -----------------------------------------------------------
#
# UNITS / CLAIM_RE — the small closed set of words that turn a digit
# group into a numeric claim.
#
# The left/right boundary was originally an ALLOWLIST of ASCII
# whitespace/punctuation (`[\s{re.escape(string.punctuation)}]`), mirroring
# the Bash version's `[[:space:][:punct:]]` classes. That allowlist is the
# same defect shape this guard has already fixed once for PROVENANCE_RE
# (docstring item 4): a boundary defined by ENUMERATION rather than by "not
# part of a token" can only ever answer for characters someone already
# listed. An ASCII-only allowlist misses every non-ASCII punctuation mark —
# an em dash (U+2014), a curly quote (U+2018), a fullwidth paren (U+FF08) —
# so a claim glued to one of those evaded detection entirely, and this
# repository's own prose uses em dashes and curly quotes routinely.
#
# The fix defines the boundary NEGATIVELY instead: the character
# immediately before the digit run (or after the unit word) must not be
# word-constituent (`\w` — letter, digit, or underscore, Unicode-aware).
# `(?<!\w)`/`(?!\w)` are zero-width, so they also naturally handle
# start/end of string without a separate `(^|...)`/`(...|$)` alternation.
UNITS = r"tests|failures|errors|files|lines|seconds|minutes|ms"
# The digit run is bounded ({0,20} repetitions of an optional comma plus a
# digit) rather than open-ended ([0-9,]*). This is the third quadratic this
# guard has had to fix (see the module docstring's "WHY THIS FILE IS PYTHON"
# section) — the reshape removed two Bash ones and introduced this one:
# `re`'s backtracking engine, given an open-ended digit/comma run with no
# trailing unit word, retries the `[\s]+` requirement once per position in
# the run, which is O(length) per match attempt and O(length) attempts, i.e.
# O(length^2) overall on ordinary PR-authored content (a pasted list of IDs
# or a long comma-separated number list) — no adversarial input required.
# Bounding the repetition count makes each attempt O(1), restoring the O(n)
# behaviour a single-pass scan is supposed to have. The general lesson this
# guard has now hit twice in two different languages: changing the
# implementation language changes which defect classes are reachable, but it
# does not remove backtracking/quadratic risk as a category — Bash's ERE
# engine could not produce this shape, Python's `re` can, and a language
# migration must budget for verifying its *new* failure modes, not just
# confirm the old ones are gone.
#
# A FOURTH one has since been caught, in review, before it shipped: the
# quotation exemption's region-membership test was a linear scan called once
# per character index, which is quadratic in the number of code spans on a
# single line. See `_region_at`. The lesson generalises past regex engines to
# any "scan a list once per element of another list" shape, which is why that
# fix is a `bisect` probe and not a faster loop.
#
# `(?<![A-Z]-)` is the issue-key exclusion. `(?<!\w)` alone cannot make
# this call: the character before the digits in `KAN-6 errors` is a
# hyphen, which is not word-constituent, so the boundary is satisfied and
# the key's number is read as a quantity. An uppercase letter followed by
# a hyphen immediately before a digit run is the issue-key shape
# (`KAN-6`, `IU-262`), never a measurement, so it is excluded here rather
# than left to the generic boundary. The lookbehind is fixed-width, which
# Python's `re` requires; every other boundary decision stays with
# `(?<!\w)`.
CLAIM_RE = re.compile(
    rf"(?<![A-Z]-)(?<!\w)[0-9](?:,?[0-9]){{0,20}}[\s]+(?:{UNITS})(?!\w)"
)

# PROVENANCE_RE — anchored on the HTML comment opener, not merely
# "contains measured:/predicted: somewhere". Without this anchor, a
# comment that NAMES the words while denying provenance
# (`<!-- not actually measured: just a guess -->`,
# `<!-- unmeasured: revisit -->`) would satisfy the rule exactly as well
# as a real one.
PROVENANCE_RE = re.compile(r"<!--\s*(measured|predicted):")

# _QUOTE_PAIRS — the quotation delimiters that make a number a QUOTATION
# rather than a claim, and the WHOLE delimiter set: matched double quotes,
# straight and curly. Single quotes are deliberately absent: prose
# apostrophes ("the guard's scope") are unpaired by nature, so admitting them
# would make pairing meaningless on most lines of this repository's
# documentation. Inline code spans were a third class and are gone — see the
# module docstring's "WHY THE QUOTATION EXEMPTION IS QUOTES-ONLY", which is
# the only place that history is written down.
_QUOTE_PAIRS = (('"', '"'), ("“", "”"))

# _DELIMITER_CHARS — every delimiter character, DERIVED from `_QUOTE_PAIRS`
# rather than re-listed, so a future delimiter class cannot be added to the
# pairing scanner while the escape veto below keeps answering for the old set.
#
# The angle-bracket veto does NOT consume this: it never looks at delimiters at
# all (see `_ANGLE_BRACKET`), which is precisely what makes it answer for every
# delimiter class, present and future, without being told about any of them.
_DELIMITER_CHARS = "".join(sorted({c for pair in _QUOTE_PAIRS for c in pair}))

# VETO_* — why a line's quotation exemption was withdrawn, in the operator's
# words. These strings are printed (see `_veto_note`), so they are phrased as
# the cause an author can act on rather than as an internal enum.
#
# EVERY REASON MUST NAME SOMETHING THE AUTHOR CAN VERIFY BY READING THEIR OWN
# LINE (pass-18 fix wave, Important). `VETO_RAW_HTML` — "a quotation delimiter
# inside raw HTML" — did not: it was returned for the old helper's SECOND
# answer as well ("this line opened a construct it never closed"), which fires
# on `the range a <b and c covers 5 tests`, a line carrying no quotation
# delimiter anywhere. The operator was sent to look for something provably not
# there. The replacement states the literal character that triggered it.
VETO_ESCAPE = "a backslash-escaped quotation delimiter"
VETO_ANGLE_BRACKET = "a `<` character on the line"
VETO_UNBALANCED = "unbalanced quotation delimiters"

# _VETO_REMEDIES — the fix for each veto. Lens C, pass-17 fix wave: the
# distinguishing information existed in-process and was discarded before
# printing, so an author whose correctly-quoted number lost its exemption to
# an unrelated stray delimiter had no way to learn that from the tool — and
# the exit-1 banner's only advice ("add the missing tag") would have had them
# write a `measured:` tag for a number that was already quoted, which is the
# tag vocabulary being used untruthfully to silence a guard.
_VETO_REMEDIES = {
    VETO_ESCAPE: "unescape it or reword the line",
    VETO_ANGLE_BRACKET: "remove the `<` or reword the line",
    VETO_UNBALANCED: "balance them or reword the line",
}

# _ESCAPED_DELIMITER_RE — the escape veto's trigger (pass-16 fix wave,
# Critical 1 — item 4 of the module docstring's "WHY THE QUOTATION EXEMPTION
# IS QUOTES-ONLY", which is the only place that history is written out).
#
# CommonMark 0.31.2 §2.4 makes a backslash followed by ASCII punctuation a
# LITERAL character: `\"` is a quotation mark, not a quotation delimiter.
# `_quote_regions`'s character scan did not know that, so it paired escaped
# delimiters as live ones and manufactured regions that do not exist —
# exempting numbers standing in open prose.
#
# `(?<!\\)(\\+)` captures a MAXIMAL backslash run: the lookbehind pins the
# match to the run's first character, so `group(1)` is the whole run and its
# parity is CommonMark's own. An ODD run escapes the delimiter after it; an
# EVEN one is a sequence of escaped backslashes leaving that delimiter LIVE
# (`\\` then `"` still opens a quotation). Parity, not mere adjacency, is
# what makes the veto stop at the escapes and no further.
#
# DELIBERATE OVER-APPROXIMATION, in the safe direction only: §2.4 escapes
# ASCII punctuation, so a backslash before a CURLY quote (U+201C/U+201D) is
# not an escape at all — the backslash is literal and the curly quote stays
# live. This pattern vetoes that line anyway, because keeping the delimiter
# set derived from `_QUOTE_PAIRS` is worth more than the one exemption it
# costs, and every error a veto can make is a finding REPORTED rather than a
# finding removed.
_ESCAPED_DELIMITER_RE = re.compile(
    r"(?<!\\)(\\+)[" + re.escape(_DELIMITER_CHARS) + r"]"
)

# _ANGLE_BRACKET — the raw-HTML veto's ENTIRE trigger, and deliberately not a
# regex (pass-18 fix wave, Critical: the SIXTH fail-open).
#
# CommonMark 0.31.2 §6.6 makes an HTML comment, a tag, a declaration, a CDATA
# section and a processing instruction RAW HTML: their contents are passed
# through untouched and nothing inside one is parsed as an inline delimiter.
# A `"` inside `<!-- ... -->` is therefore literal content, and pairing it with
# a live delimiter elsewhere on the line invents a region exactly as the escape
# case did. That much was always right. What was wrong was how the constructs
# were LOCATED.
#
# The previous version matched them with a per-construct regex whose tag
# alternative was `</?[A-Za-z][^<>]*>`. §6.6 permits `>` inside a
# SINGLE-QUOTED attribute value, so a real tag can extend past the first `>`
# and the pattern truncated there. When the truncated prefix carried no
# delimiter, the veto cleared it, advanced past it, found no further `<` and
# returned False — leaving a `"` that is genuinely inside the tag live as a
# delimiter:
#
#     <a b='>"'> the benchmark ran 77 tests "
#
# exited 0 with the bare claim unreported. A differential sweep against an
# accurate §6.6 grammar found 150 distinct fail-open arrangements of that
# shape; the DOUBLE-quoted attribute form self-vetoes, which is why it was
# missed. It was the fifth consecutive attempt to approximate a spec grammar
# with a regex, and every fail-open this exemption has ever had came from one.
#
# So the grammar is GONE, not improved. The rule is now the coarsest one that
# can be stated:
#
#     IF THE LINE CONTAINS A `<` AT ALL, THE EXEMPTION IS WITHDRAWN.
#
# There is no tag grammar, no comment/PI/CDATA/declaration alternative and no
# attribute parsing — nothing to approximate, so nothing to leak. It is
# trivially fail-closed and trivially auditable, it is O(1) work per line with
# no engine behind it (this file has already fixed four quadratics; a veto
# cannot introduce a fifth if it does not scan), and it answers for every
# delimiter class at once because it never looks at delimiters.
#
# The cost was measured before it was chosen, over all 110 markdown files in
# this repository: 38 numeric claims are exempted, and exactly 2 sit on a line
# containing `<`. One of those is the manual-test guide's own fixture line for
# the behaviour being replaced; the other already carries a provenance tag, so
# it reports nothing. Inside the guard's actual scan scope
# (`openspec/changes/*/{tasks,design,proposal}.md`) the loss is zero.
_ANGLE_BRACKET = "<"


def _has_escaped_delimiter(text: str) -> bool:
    """True when `text` carries a backslash-escaped quotation delimiter — an
    ODD run of backslashes immediately before one.

    See `_ESCAPED_DELIMITER_RE` for the escape rule and `quotation_regions`
    for what the answer is used for (a whole-line veto).
    """
    return any(
        len(m.group(1)) % 2 == 1 for m in _ESCAPED_DELIMITER_RE.finditer(text)
    )


def _has_angle_bracket(text: str) -> bool:
    """True when `text` contains a `<` — the whole raw-HTML veto.

    This asks nothing about HTML. It does not decide whether the `<` opens a
    construct, where that construct ends, or whether a delimiter is inside it,
    because deciding any of those requires a §6.6 grammar and five successive
    approximations of one are what produced every fail-open this exemption has
    had (see `_ANGLE_BRACKET` for the sixth, which is why this function has no
    regex in it at all).

    The trade this makes is stated rather than hidden: a line that merely
    mentions `<` in prose loses its quotation exemption. That is a finding
    REPORTED where a strict CommonMark reader would have exempted it — a loud
    false positive, the only direction this exemption is allowed to be wrong
    in — and it was measured to cost this repository nothing inside the scan
    scope before it was chosen.
    """
    return _ANGLE_BRACKET in text


# _REGION_HI — the "larger than any end offset" second element used to make
# a `bisect` probe land after every region whose START equals the probed
# index. `sys.maxsize` is a plain int, so the tuple comparison stays a
# total order on `(int, int)`.
_REGION_HI = sys.maxsize


def _regions_are_sorted_disjoint(regions: list) -> bool:
    """True when `regions` is sorted by start, pairwise disjoint, and made
    only of non-inverted half-open spans — `_region_at`'s precondition,
    stated as something executable rather than as prose in three docstrings.
    """
    previous_end = 0
    for start, end in regions:
        if start < previous_end or end < start:
            return False
        previous_end = end
    return True


def _region_at(index: int, regions: list) -> Optional[Tuple[int, int]]:
    """The region of `regions` containing `index`, or `None`.

    `regions` must be sorted by start and pairwise DISJOINT — which every
    region list this module builds is, by construction (`_quote_regions`
    emits strictly left-to-right, non-overlapping runs). Disjointness is what
    makes one `bisect` probe sufficient: the last region whose start is
    `<= index` is the ONLY one that can contain it.

    THE PRECONDITION IS ENFORCED, NOT MERELY DOCUMENTED (Lens B, pass-17 fix
    wave). It used to be stated in three docstrings and checked nowhere,
    which matters more here than it would elsewhere: this is the single
    chokepoint every exemption flows through, and a violated precondition
    makes the `bisect` probe return a silently WRONG answer in the fail-open
    direction — a region reported as containing an index it does not. That is
    the shape of four prior defects in this exemption, and "silently wrong,
    fail-open" is the one outcome this code may not have.

    Two checks, deliberately split by cost:

      - here, an O(1) LOCAL check on the probed region and its immediate
        neighbours, which fails CLOSED (returns `None` — "not enclosed") so a
        violation withdraws an exemption instead of granting one;
      - at the producer, an O(N) full check (`_quote_regions`'s `assert`),
        which fails FAST and loudly, because a producer emitting an
        out-of-order list is programmer error rather than input.

    The O(1) check cannot be an O(N) validation: `_quote_regions` calls this
    once per character index, so validating the whole list here would restore
    exactly the O(N^2) the `bisect` probe below was introduced to remove.

    Important 3 (pass-15 fix wave): that probe replaced a linear
    `any(start <= index < end for ...)` scan, which made the exemption
    quadratic in the number of regions on a single line.

    ONE figure, with the line shape it was measured on (pass-16 fix wave,
    Minor A). Three different numbers for this same nominal measurement were
    in circulation — 3.38 s, 6.709 s and 3.13 s — because each was taken on a
    differently-crafted line and none said which, in a guard whose entire
    subject is numeric claims that do not name their source. They are
    superseded by this one, which states its shape:

        line  = "`x` " * 4000 + "99 tests"   (4000 regions, 16,008 chars)
        timed = quotation_regions(line), linear membership vs this `bisect`
        3.15 s before, 0.014 s after; /usr/bin/python3 3.9.6, macOS 25.5.0

    Doubling the region count quadrupled the "before" time (0.05 / 0.20 /
    0.79 / 3.15 / 12.83 s at 500 / 1000 / 2000 / 4000 / 8000 regions) and
    doubled the "after" time, which is the quadratic-vs-linear claim itself
    rather than any single number. Hours on a line still well inside
    `MAX_SCANNED_FILE_BYTES`, on a guard that runs as a CI gate over
    PR-authored content. That is a denial of service, and it would have been
    the FOURTH quadratic in this guard (see the module docstring's "WHY THIS
    FILE IS PYTHON" enumeration and `CLAIM_RE`'s comment block for the
    previous three). `bisect` is standard library, which this file is
    restricted to.

    The measurement was taken while code spans were still a delimiter class,
    which is what made 4000 regions on one line easy to construct; the
    quadratic it records is a property of the probe, not of that class, and
    the probe is unchanged.
    """
    i = bisect.bisect_right(regions, (index, _REGION_HI)) - 1
    if i < 0:
        return None
    start, end = regions[i]
    if end < start:
        return None
    if i > 0 and regions[i - 1][1] > start:
        return None
    if i + 1 < len(regions) and regions[i + 1][0] < end:
        return None
    return (start, end) if start <= index < end else None


def _encloses(start: int, end: int, regions: list) -> bool:
    """True when the half-open span `[start, end)` lies wholly inside one
    region of the sorted, disjoint list `regions`."""
    region = _region_at(start, regions)
    return region is not None and end <= region[1]


def _quote_regions(text: str, opener: str, closer: str) -> Optional[list]:
    """The interior of every matched `opener`/`closer` pair on `text`, or
    `None` when any delimiter of that class is left over — the CLASS-WIDE
    VETO.

    `None` and `[]` are different answers, and the difference is what
    `_check_numeric_claim` prints: `[]` means "this class has no delimiters
    on this line, so it exempts nothing", while `None` means "this class's
    delimiters do not all pair, so nothing it might have enclosed can be
    trusted". Both exempt nothing — the veto is fail-closed either way — but
    only the second is something an author can act on, and discarding that
    distinction before printing is what left an author staring at "no
    provenance comment" for a number they had correctly quoted.

    The returned regions are sorted by start and pairwise disjoint — the
    scan emits them strictly left to right and never reopens a closed one —
    which is the precondition `_region_at`'s single `bisect` probe needs, and
    which the `assert` below turns from a comment into a check (Lens B,
    pass-17 fix wave; see `_region_at` for why this producer asserts while
    the consumer fails closed).

    THE CLOSER/OPENER ASYMMETRY (pass-16 fix wave, Minor B). A leftover
    OPENER vetoes the class (`pending is not None` below returns `None`), but
    a closer arriving with no pending opener is silently SKIPPED — the `elif`
    chain simply falls through it. That asymmetry is deliberate and safe, and
    the reason is that the two leftovers are not the same kind of ignorance:

      - A leftover opener leaves the class's region EXTENTS undetermined:
        text after it is inside an unterminated quotation whose end is not on
        this line, so nothing downstream of it can be said to be enclosed or
        not. Undetermined means veto.
      - A leftover closer removes nothing. Every region this scan emits is
        bounded by an opener and the very next closer after it, with no
        opener in between (a second opener vetoes the class outright). A
        number inside such a region is therefore literally between a curly
        open quote and a curly close quote — a genuine enclosure — whatever
        stray closer happened to appear earlier in the line. Skipping it
        cannot manufacture a region; it can only fail to consume a character
        that could not have opened anything anyway.

    This is why it is NOT the same shape as the straight-quote parity
    failure `_is_quoted` describes. There, opener and closer are the SAME
    character, so a stray delimiter is indistinguishable from an opener and
    the pairing itself becomes ambiguous. Here the roles are typed by the
    character, so a stray `”` is unambiguously not an opener. This asymmetry
    is reachable only for the curly class; for a same-character class the
    "closer with nothing pending" branch is unreachable by construction,
    because such a character is taken as an opener.

    Checked, not merely argued: exhaustively over every arrangement of up to
    six curly delimiters around a claim, every exemption this function
    granted had the claim's nearest live delimiter to the left an opener and
    its nearest to the right a closer — i.e. no exemption was ever granted
    without a genuine enclosing pair.
    """
    regions = []
    pending = None
    for i, char in enumerate(text):
        if pending is None:
            if char == opener:
                pending = i
        elif char == closer:
            regions.append((pending + 1, i))
            pending = None
        elif char == opener:
            # A second opener before the first is closed (possible only for
            # a distinct opener/closer pair). Nesting is not a shape this
            # guard can resolve, so the class is vetoed.
            return None
    if pending is not None:
        return None
    assert _regions_are_sorted_disjoint(regions), (
        "quotation regions must be sorted and pairwise disjoint: "
        f"{regions!r}"
    )
    return regions


@dataclass(frozen=True)
class QuotationScan:
    """One line's quotation analysis: the regions inside which a number is
    reproduced rather than asserted, and why — if at all — the exemption was
    withdrawn.

    `classes` holds one sorted, pairwise-disjoint region list per delimiter
    class that survived. A vetoed class contributes NO list, so an empty
    `classes` exempts nothing: the fail-closed answer is the absence of
    regions rather than a sentinel that every consumer has to remember to
    check. This replaced an `Optional[list]` whose `None` meant "the whole
    line is indeterminate" (pass-17 fix wave): with the code-span class gone
    that sentinel had one producer and one consumer, and an empty tuple says
    the same thing without a second way to spell "exempt nothing".

    `veto` is the reason, or `None` when no veto fired. It exists so the
    guard can tell an author WHY a correctly-quoted number was reported —
    see `_veto_note` and `_check_numeric_claim`.

    The classes are kept as separate lists rather than merged into one:
    within a class the regions are disjoint (which is what lets `_region_at`
    answer with a single `bisect` probe), but a straight-quoted passage may
    legitimately contain a curly-quoted one, so a merged list would not be.
    """

    classes: Tuple[list, ...]
    veto: Optional[str]


def quotation_regions(text: str) -> QuotationScan:
    """Analyse `text` for the regions inside which a number is a QUOTATION
    rather than an assertion.

    Computed ONCE PER LINE and threaded into `_is_quoted` (Important 3,
    pass-15 fix wave). `_check_numeric_claim` used to call `_is_quoted` once
    per `CLAIM_RE` match, each call recomputing every region on the line from
    scratch. The helpers stay pure — the regions are a parameter, not a cache
    hidden inside a function — so the only thing that changed is who owns the
    one computation.

    Three vetoes, in order, each able only to WITHDRAW an exemption. The
    first two are whole-line and are checked before any pairing is attempted;
    the third is per class. See the module docstring's "WHY THE QUOTATION
    EXEMPTION IS QUOTES-ONLY" for the history that produced them, and
    `_has_escaped_delimiter` / `_has_angle_bracket` / `_quote_regions` for each
    one's own rule.

    Why VETO rather than "strike the offending delimiters and pair what
    remains": a veto's output is the absence of regions, and an absence can
    only ever withdraw an exemption. Whatever a veto's analysis gets wrong,
    the worst outcome is a claim REPORTED that a strict CommonMark reader
    would have exempted — a loud false positive — never a claim silently
    accepted. Striking, by contrast, needs two passes to agree about a
    structure neither can compute alone, and "two passes that must agree" is
    the shape of every fail-open in the history above.

    KNOWN LIMITATION, and it is that worst outcome made explicit: a whole
    line loses its exemption for a delimiter problem anywhere on it — or, for
    the angle-bracket veto, for a single `<` anywhere on it — however soundly
    a quotation elsewhere on the same line pairs. Harness cases 221, 223 and
    224 pin the escape form, the angle-bracket form and the case where the `<`
    is nowhere near a delimiter, so a later reader meets the trade-off as an
    assertion rather than rediscovering it as a surprise.
    """
    if _has_escaped_delimiter(text):
        return QuotationScan(classes=(), veto=VETO_ESCAPE)
    if _has_angle_bracket(text):
        return QuotationScan(classes=(), veto=VETO_ANGLE_BRACKET)
    classes = []
    veto = None
    for opener, closer in _QUOTE_PAIRS:
        regions = _quote_regions(text, opener, closer)
        if regions is None:
            veto = VETO_UNBALANCED
        else:
            classes.append(regions)
    return QuotationScan(classes=tuple(classes), veto=veto)


def _veto_note(relfile: str, lineno: int, veto: str) -> str:
    """The operator-facing explanation of a withdrawn exemption.

    Names the cause and the remedy, and points at the contract — the three
    things missing from `_check_numeric_claim`'s message, which said only
    that provenance was absent and so sent an author to tag a number that
    was already correctly quoted.
    """
    return (
        f"{relfile}:{lineno}: note: the quotation exemption was withdrawn "
        f"on this line ({veto}) — {_VETO_REMEDIES[veto]}; "
        "see skills/myflow-contracts/plan-provenance.md"
    )


def _is_quoted(start: int, end: int, scan: QuotationScan) -> bool:
    """True when the half-open offset span `[start, end)` sits inside a
    matched delimiter pair on the line `scan` was computed from — see
    `quotation_regions`, which is where that line's text is read. This
    predicate never sees the text itself, so it cannot disagree with the
    regions it was handed.

    Line-scoped by construction: a pair spanning a line break is not
    recognised, which keeps this a pure predicate rather than a second
    piece of parser state that would have to agree with the fence tracker.

    FAILS CLOSED, and the strength of that is the whole point: this
    function can only ever REMOVE a finding, so anything it gets wrong is
    a claim the guard silently accepts. It therefore answers "is this
    offset inside a genuinely matched pair?" by pairing the line's
    delimiters in order — NOT by the cheaper test of "is the count of
    openers before it odd, and does a closer appear anywhere after it?".
    That parity test fails OPEN on an ordinary shape: in

        a "quote" and a stray " mark, then 99 tests ran, "done"

    the odd count before `99` and the closer after it are both satisfied,
    yet the number is plainly asserted rather than reproduced. Pairing in
    order is not sufficient on its own either — it would pair the stray
    delimiter with the opener of the final quotation and manufacture the
    same bogus region. So a class whose delimiters do not ALL pair yields
    no region at all on that line: with a leftover delimiter, the intended
    pairing is undetermined (the stray one could belong on either side of
    the number), and the fail-closed answer to an undetermined line is to
    report the claim.

    The two delimiter classes are independent of each other — an unbalanced
    line for one of them does not withdraw an enclosure the other
    demonstrates. The two whole-line vetoes are not classes at all: they
    remove every class, which `quotation_regions` expresses by returning no
    classes to iterate.
    """
    return any(_encloses(start, end, regions) for regions in scan.classes)


# FENCE_TAG_RE — a fence's info string must carry verified:/unverified:
# as a whole token, immediately preceded by start-of-string or
# whitespace, so `preverified:` cannot satisfy it as a mere substring.
FENCE_TAG_RE = re.compile(r"(^|\s)(un)?verified:")

# --- Container-prefix grammar --------------------------------------------
#
# _WS_RE — leading whitespace, consumed once per strip attempt.
#
# NOTE (Critical 1, pass-14 fix wave): this regex is a RECOGNISER, not a
# consumer. Every place a whitespace run is actually consumed goes through
# `_scan_ws_capped`/`_consume_ws_capped` below, which require an explicit
# column budget. See the "WHITESPACE BUDGETS" block below that pair for
# why, and for the one deliberately-unbudgeted site.
_WS_RE = re.compile(r"\s+")

# _BQ_MARKER_RE — ONE blockquote marker: the `>` itself, and nothing else.
#
# Critical 1 (pass-14 fix wave): this pattern used to be `\s*>` — the
# leading `\s*` was an UNBOUNDED whitespace run whose width was fed
# straight into the running column and into the recorded blockquote
# DEPTH. CommonMark allows 0-3 columns of plain indentation before a
# block-level `>` (https://spec.commonmark.org/0.30/#block-quotes: "up to
# three spaces of initial indent"); at 4+ columns the `>` is literal text
# inside an indented code block, not a marker, and the run ENDS there.
# With `\s*` the parser instead read `>` + 5 spaces + `>` as depth 2,
# where CommonMark says depth 1 with indented content. The gap is now
# consumed by `_consume_bq_run` through `_scan_ws_capped(..., _BQ_GAP_MAX)`
# — measured in COLUMNS (tab-aware), refused past the budget — so the
# marker regex itself no longer carries any whitespace at all.
#
# The whitespace AFTER the `>` is deliberately NOT part of this pattern.
# CommonMark credits the marker with exactly ONE COLUMN of it
# (https://spec.commonmark.org/0.30/#block-quotes: "the character >,
# together with a following space"), and one column is not one character:
# a tab is a single character worth up to four columns, and spec example
# 6 (`>\t\tfoo` → a code block containing "  foo") shows the tab's
# REMAINING columns staying behind as ordinary indentation of the quoted
# content. This regex used to be `(\s*>\s?)+`, whose `\s?` swallowed a
# tab whole however wide it was — a width measured in characters fed
# into the (correct) running-column cursor, which is Critical 1 of the
# pass-13 fix wave. `_consume_bq_run` below does that one-column step
# against the running column instead, where a partially consumed tab can
# be left in place for the rest of the line to account for.
#
# Scanning a run marker-by-marker keeps the O(length) total cost the
# single `(\s*>\s?)+` match had (the Bash version's quadratic DoS, defect
# 8, came from re-matching a fresh regex against a SHRINKING STRING once
# per marker): `_consume_bq_run` matches at an advancing POSITION in one
# string, never re-slicing it, and every iteration consumes at least the
# one `>` it matched — so the whole run is still walked exactly once.
# It also removes the nested quantifier (`\s*` inside a `+`), which was
# itself a backtracking hazard on a near-miss run.
_BQ_MARKER_RE = re.compile(r">")

# _WS_CHAR_RE — exactly one whitespace character, used to test the single
# character following a `>` (see _consume_bq_run) with the same notion of
# "whitespace" `_WS_RE` and `_BQ_MARKER_RE` use, rather than a second,
# subtly different one (`str.isspace()` accepts characters `\s` does not).
_WS_CHAR_RE = re.compile(r"\s")

# _LIST_MARKER_RE — a bullet (`-`, `*`, `+`) or an ordered marker, simple
# (`3.`, `3)`) or compound (`3.4.`, `1.2.3)`), and NOTHING after it.
#
# Critical 1 (pass-14 fix wave): the trailing `\s+` this pattern used to
# carry was an UNBOUNDED whitespace run whose full width was committed
# into `list_content_col`. CommonMark bounds it
# (https://spec.commonmark.org/0.30/#list-items, "Basic case" rule 1 vs
# "Item starting with indented code" rule 2): a marker followed by W
# columns of whitespace establishes a content column of `marker_end + W`
# only while 1 <= W <= 4; at W >= 5 the content column is `marker_end + 1`
# and everything past it is INDENTED CODE inside the item. Spec example 7
# of the Tabs section (`-\t\tfoo` → `<li><pre><code>  foo`) is the tab
# form of exactly that. Consumption now goes through
# `_consume_list_marker_ws`, which applies that rule; this regex is only
# the marker itself.
_LIST_MARKER_RE = re.compile(r"[-*+]|[0-9]+(\.[0-9]+)*[.)]")

# _LIST_RE — the RECOGNISER: a list marker immediately followed by at
# least one whitespace character. The whitespace requirement is what makes
# `--` (glued, no separating space) fail this pattern entirely — it is not
# a "doubled marker" under this grammar, it simply is not a marker. Used
# only to answer "is there a list marker here?" (`_starts_container_syntax`
# and `strip_container_prefix`'s gate); `m.end()` is never consumed from
# it, so `\s` rather than `\s+` is deliberate — one character is all a
# yes/no needs, and a `+` here would invite a future edit to consume the
# unbounded run again.
_LIST_RE = re.compile(r"([-*+]|[0-9]+(\.[0-9]+)*[.)])\s")

# _CHECKBOX_MARKER_RE / _CHECKBOX_RE — a task checkbox, only ever stripped
# immediately after a list marker (see strip_container_prefix below); a
# bare `[ ]` with no preceding list marker is a shape this grammar does
# not resolve. Split for the same reason as the list marker above: the
# marker for consuming, the recogniser for asking.
_CHECKBOX_MARKER_RE = re.compile(r"\[[ xX]\]")
_CHECKBOX_RE = re.compile(r"\[[ xX]\]\s")

# _FENCE_RUN_RE — a run of 3+ backticks or 3+ tildes, anywhere in a line
# (used as a cheap pre-filter) or anchored at a string's start (used
# once stripping has been attempted).
_FENCE_RUN_ANYWHERE_RE = re.compile(r"`{3,}|~{3,}")
_FENCE_RUN_START_RE = re.compile(r"^(`{3,}|~{3,})")

# CommonMark caps an UN-CONTAINED fence's own leading indentation at 0-3
# columns: at 4+ columns with no container prefix, the line is an INDENTED
# CODE BLOCK, and a run of backticks/tildes there is literal text, never a
# fence boundary. `_UNCONTAINED_FENCE_INDENT_MAX` is that cap. See
# classify_line's use of `_advance_col` below for why this must be a
# measured COLUMN width (tabs expand to the next multiple of 4) rather than
# a raw character count.
_UNCONTAINED_FENCE_INDENT_MAX = 3

# _LIST_GAP_MAX — CommonMark tolerates 0-3 columns of plain indentation
# before a list marker, including directly after a blockquote strip. This
# is a DIFFERENT budget from `_UNCONTAINED_FENCE_INDENT_MAX` above (which
# caps a fence run's OWN indentation relative to its container) — this one
# caps how much whitespace may sit between a blockquote run and the list
# marker that follows it before the two are recognised as one nested
# container. A blockquote marker itself only ever takes ONE such column
# (Critical 2, pass-9 fix wave; the marker's own one-column allowance is
# `_consume_bq_run`'s, and was `_BQ_RUN_RE`'s `\s?` before pass 13): a
# two-space gap between `>` and `-` fell through that budget entirely, so
# neither the blockquote run nor `_LIST_RE`
# matched, the line was misread as unresolvable prose, the fence never
# opened, and its closing delimiter was later read as a fresh untagged
# opener that silently swallowed everything after it. See
# strip_container_prefix for where this is applied and why a gap of 4+ is
# deliberately NOT absorbed the same way.
_LIST_GAP_MAX = 3

# _BQ_GAP_MAX — CommonMark's 0-3 columns of plain indentation permitted
# in front of a block-level `>` marker, INCLUDING one that follows a
# previous `>` in the same run ("up to three spaces of initial indent",
# https://spec.commonmark.org/0.30/#block-quotes; the same 0-3 budget
# applies to every block-start token, per
# https://spec.commonmark.org/0.30/#indented-code-blocks, which is what
# reserves column 4). At 4+ columns the `>` is literal text inside an
# indented code block and the blockquote run ENDS at that point.
# Critical 1, pass-14 fix wave: this budget did not exist — `_BQ_MARKER_RE`
# carried an unbounded `\s*` instead, so `>` + 5 spaces + `>` recorded
# depth 2 where CommonMark reads depth 1.
_BQ_GAP_MAX = 3

# _LIST_MARKER_WS_MAX — CommonMark's "Basic case" bound on the whitespace
# a list marker may credit to its own content column
# (https://spec.commonmark.org/0.30/#list-items, rule 1: content begins
# W columns after the marker "where 1 <= W <= 4"). A run of 5+ columns
# falls to rule 2 ("Item starting with indented code"), where the content
# column is exactly one past the marker and the rest is indented code.
# `_consume_list_marker_ws` implements both branches; this is rule 1's
# bound. Critical 1, pass-14 fix wave.
_LIST_MARKER_WS_MAX = 4

# _MEASURE_ALL — the explicit "no column budget" sentinel for
# `_scan_ws_capped`. Passing it is how a call site STATES that it wants a
# whitespace run measured in full, rather than getting that behaviour by
# omission. See the WHITESPACE BUDGETS note under `_scan_ws_capped`; a
# grep for this name enumerates every intentionally unbudgeted run in the
# parser. There are exactly two, both in the same category: a run whose
# TRUE width is needed in order to REFUSE it (the outer indentation in
# `strip_container_prefix` and the residual indentation in
# `classify_line`, each compared against
# `_UNCONTAINED_FENCE_INDENT_MAX`). Clipping either to its budget would
# hand the comparison a width of 3 for every over-indented line and
# silently delete the refusal, which is the opposite of this wave's fix.
_MEASURE_ALL = -1


def _advance_col(col: int, consumed: str) -> int:
    """Advance the running ABSOLUTE column `col` past `consumed` —
    whatever text a strip step just removed from the front of the
    remaining line — expanding a tab to the next multiple-of-4 column
    (CommonMark's tab-expansion rule for block structure) and any other
    character by exactly one column.

    This is the single running cursor the pass-12 fix wave threads
    through the whole prefix pipeline (`strip_container_prefix`,
    `classify_line`, `is_closing_line`), replacing the deleted
    `_leading_ws_width`, which measured a (whitespace-only) slice's width
    by expanding tabs as if that slice began at column 0. That assumption
    held only for the OUTERMOST whitespace strip, the one place a slice
    genuinely starts at the line's true column 0. Every other call site
    handed it a MID-LINE slice — `list_candidate[:m.end()]` (gap+marker),
    `s[:m2.end()]` (checkbox), `stripped[:resid_ws.end()]` (residual) —
    where a tab's real expansion depends on how far into the line it
    actually sits, not on the length of whatever substring a later strip
    step happens to be looking at. A brute-force sweep against the three
    reproduced shapes (tab after a bullet/ordered marker, tab after a
    checkbox, tab after a blockquote marker preceding an untagged fence)
    found 50 under-counts and 40 over-counts of the true content column
    before this fix (see the module's fix-wave-pass12 report).

    Because `_advance_col` takes the CURRENT absolute column as its
    starting reference rather than assuming 0, the same function is
    correct at every call site regardless of how much of the line has
    already been consumed — there is exactly one place tab expansion is
    computed in this entire parser, and it is always seeded from the
    true running position of the cursor on the real line, never from a
    fresh slice's own zero."""
    for ch in consumed:
        if ch == "\t":
            col += 4 - (col % 4)
        else:
            col += 1
    return col


def _scan_ws_capped(
    s: str, pos: int, col: int, max_cols: int
) -> Tuple[int, int, int]:
    """Scan a whitespace run in `s` starting at `pos`, consuming AT MOST
    `max_cols` COLUMNS of it, and report `(new_pos, width, new_col)`.

    WHITESPACE BUDGETS (Critical 1, pass-14 fix wave — the fifth
    recurrence of one defect family)
    ------------------------------------------------------------------
    Five separate Criticals across passes 7, 9, 10, 11 and 14 have had
    the identical shape: *a whitespace run measured in one place and not
    bounded where it is used*. The pass-7 ambient tracker's uncapped
    trailing gap after a marker (deleted); the gap before a list marker
    (`_LIST_GAP_MAX`); `_leading_ws_width`'s mid-line call sites (fixed
    by making the cursor absolute); `_BQ_RUN_RE`'s `\\s?` measuring
    characters instead of columns; and — pass 14 — the runs still living
    inside `_LIST_RE`, `_CHECKBOX_RE` and `_BQ_MARKER_RE`.

    Each of those was fixed where it was found. The family survived
    because the parser had no place that made "how many columns is this
    run allowed to be worth?" a question a call site had to answer. This
    function is that place: it is the ONLY way a whitespace run is
    consumed anywhere in the parser, and `max_cols` has no default. A
    site that genuinely wants the whole run must say so by passing
    `_MEASURE_ALL`, which is greppable — the family is now "a run
    consumed without a stated budget", and that is not expressible.

    `_WS_RE`, correspondingly, is demoted to a recogniser: it may be
    matched to ask *whether* whitespace is present and how wide the run
    is in total, but its `m.end()` is never sliced off as consumed text.
    Same split as `_LIST_MARKER_RE`/`_LIST_RE` and
    `_CHECKBOX_MARKER_RE`/`_CHECKBOX_RE`.

    Two properties this keeps from the code it replaces:

      - Columns, not characters. A tab is expanded from the running
        absolute `col` (`_advance_col`'s rule), never from a slice's own
        zero.
      - A tab STRADDLING the budget is not consumed. Its character stays
        at `new_pos` and only the columns that fit are added to `col`, so
        a later `_advance_col` over that same tab credits exactly its
        REMAINING columns and no tab stop is counted twice. This is the
        technique `_consume_bq_run` already used for the marker's own
        one-column allowance, generalised.

    Position-based (rather than returning `s[pos:]`) so a caller walking
    a run of markers never re-slices a shrinking string — the quadratic
    the Bash version shipped as defect 8. `_consume_ws_capped` below is
    the string-front convenience wrapper for the single-shot call sites.
    """
    width = 0
    while pos < len(s):
        ch = s[pos]
        if not _WS_CHAR_RE.match(ch):
            break
        step = (4 - (col % 4)) if ch == "\t" else 1
        if max_cols != _MEASURE_ALL and width + step > max_cols:
            # The character does not fit the budget. For a tab, take the
            # columns that DO fit and leave the character in place (see
            # the docstring); for anything else `step` is 1, so nothing
            # fits and nothing is taken.
            take = max_cols - width
            col += take
            width += take
            break
        col += step
        width += step
        pos += 1
    return pos, width, col


def _consume_ws_capped(s: str, col: int, max_cols: int) -> Tuple[str, int, int]:
    """`_scan_ws_capped` at the front of `s`, reported in the
    `(remainder, width, new_col)` shape every strip step in this parser
    uses. See `_scan_ws_capped` for the budget contract."""
    pos, width, col = _scan_ws_capped(s, 0, col, max_cols)
    return s[pos:], width, col


def _consume_list_marker_ws(s: str, col: int) -> Tuple[str, int, int]:
    """Consume the whitespace a list marker credits to its own content
    column, per CommonMark's two list-item rules
    (https://spec.commonmark.org/0.30/#list-items).

    `s` begins immediately after the marker; `col` is the running
    absolute column there. Let W be the FULL width, in columns, of the
    whitespace run that follows:

      - rule 1 ("Basic case"), 1 <= W <= `_LIST_MARKER_WS_MAX` (4): the
        content column is W columns past the marker — the whole run is
        the marker's, and is consumed;
      - rule 2 ("Item starting with indented code"), W >= 5: the content
        column is exactly ONE column past the marker, and the remaining
        W-1 columns are indented-code indentation belonging to the item's
        CONTENT, not to the marker. Only that one column is consumed; the
        rest stays in the remainder for the caller to account for.

    Rule 2 is why `-` followed by five spaces and a tagged fence must not
    open a fence: the item's content column is 2, the backtick run sits 4+
    columns past it, and CommonMark reads it as literal indented code. It
    is also spec example 7 of the Tabs section — `-\\t\\tfoo` renders as
    `<li><pre><code>  foo</code></pre></li>`, i.e. content column 2 with
    the tabs' remaining columns left behind as code indentation, which is
    exactly the remainder this function leaves in place.

    Measuring W before choosing the branch is not the defect family this
    wave closes: both branches are bounded (by 4 and by 1 respectively),
    and the measurement selects between them rather than being spent.
    """
    m = _WS_RE.match(s)
    if not m:
        return s, 0, col
    full_width = _advance_col(col, s[:m.end()]) - col
    budget = full_width if full_width <= _LIST_MARKER_WS_MAX else 1
    return _consume_ws_capped(s, col, budget)


def _commit_span(s: str, end: int, col: int) -> Tuple[str, int, int]:
    """Commit the prefix span `s[:end]` — whatever a single strip step
    just matched — reporting `(remainder, width_in_columns, new_col)`.

    Important 5 (pass-13 fix wave): every prefix step in
    `strip_container_prefix` (the outer whitespace, the list-marker gap,
    the marker itself, its checkbox) needs the same three things from a
    match, and each used to spell them out separately — slice the text,
    hand it to `_advance_col`, subtract to get a width, assign the new
    column. Written four times, the widths and the cursor drifted apart
    into "several derived values sharing a name" (`col_after_bq_pre +
    gap_width` reconstructed a column the cursor could have simply been
    advanced to). Stated once here, the cursor is only ever ADVANCED, and
    a width is only ever the distance it moved.

    The gap step is the one that must be able to NOT commit — it is
    measured before the code knows whether a list marker follows it. That
    is expressed by calling this on a trial copy of the cursor and using
    the result only if the marker matched (see `strip_container_prefix`),
    which is why this returns a value rather than mutating anything: a
    trial that is discarded leaves no trace on the real cursor.
    """
    new_col = _advance_col(col, s[:end])
    return s[end:], new_col - col, new_col


@dataclass(frozen=True)
class FenceContext:
    """The container syntax present when a fence was opened.

    Recorded once, at open time, and consulted — never recomputed from
    a content line's own prefix — every time a candidate closing line is
    tested. `bq_pre`/`bq_post` are the NUMBER of blockquote markers
    stripped before/after the (at most one) list marker, not merely
    whether any were present: a fence opened with `bq_pre = 0` (a bare,
    unprefixed opener) accepts only an unprefixed closing line; content
    that merely LOOKS like a container-prefixed fence run (e.g. a quoted
    blockquote fence, appearing as ordinary CONTENT) is never mistaken
    for that fence's close, because the close test never attempts to
    strip a blockquote marker unless this context says one was present
    at open time — and a fence opened N blockquote levels deep requires
    the closing line to carry exactly N levels too, not merely "at least
    one" or "some". An earlier version of this parser recorded `bq_pre`/
    `bq_post` as booleans; a fence opened two levels deep was then falsely
    closed by a one-level-deep line, because both are truthy. Tracking the
    count (not just presence) is what closes that gap — for any depth, not
    only the specific case a fixture happened to exercise.

    `list_content_col` (Critical 1, pass-9 fix wave; widened at pass-10 and again
    at pass-11 — see strip_container_prefix — each pass carrying one more
    of the widths that feed it) IS carried, unlike the
    boolean `has_list` this docstring used to say was deliberately
    omitted as dead state — that was true before this field existed,
    and is corrected here rather than left to contradict the code below
    it. `list_content_col` is the COLUMN WIDTH consumed by the gap (if any was
    tolerated before it — see strip_container_prefix and Critical 2 of
    the pass-10 fix wave) plus the (at most one) list marker plus its
    optional checkbox, exactly as they appeared on the OPENING line — 0
    when no list marker was present at open time.
    It matters because a continuation/closing line never repeats a list
    marker (ordinary Markdown list-item continuation is by indentation,
    not by re-typing the bullet), so the closing test cannot learn the
    list's content column from the closing line itself; it has to be
    told what the opening line established. `is_closing_line` requires
    exactly this many columns of plain indentation, in addition to
    whatever `bq_pre`/`bq_post` require, before applying the ordinary
    0-3 column fence-indent budget to whatever remains.

    This is per-fence state captured from a marker physically present
    on THIS fence's own opening line, consulted only while this fence is
    open, and discarded when it closes — structurally identical to how
    `bq_pre`/`bq_post` already work, and not to be confused with the
    deleted pass-7 ambient tracker (`_next_list_content_col`), which
    guessed a list's content column for lines carrying no marker of
    their own, threaded across blank lines and dedents. That tracker
    caused three Criticals and was deleted for exactly that reason (see
    classify_line's docstring). `list_content_col` never infers anything about a
    line that has no list marker on it; it only remembers what a marker
    already on the opening line established, for as long as that one
    fence stays open.

    NOT carried here: the opening line's own outermost whitespace width
    (Minor 9, pass-14 fix wave — this used to be a fourth field,
    `outer_ws_width`). Its docstring claimed a lifetime role — "the width
    the OPENING line was admitted at" — that nothing ever consumed: its
    only read was `classify_line`'s, on the context of the very line that
    had just produced it, one statement later. A per-fence field read only
    on the same line that created it is a local with extra reach, and it
    invited exactly the misreading the docstring then had to spend a
    paragraph denying (that a closing line must match the opener's outer
    indent — it must not; CommonMark budgets those 0-3 columns per line
    independently). This is the same finding, and the same remedy, as the
    `has_list` boolean two paragraphs above: `strip_container_prefix` now
    RETURNS that width alongside the context, and `classify_line` holds it
    in a local for the two statements it is needed.
    """

    bq_pre: int
    bq_post: int
    list_content_col: int = 0


EMPTY_CONTEXT = FenceContext(0, 0, 0)


def _consume_bq_run(
    s: str, col: int, expected: Optional[int] = None
) -> Optional[Tuple[str, int, int]]:
    """Consume one run of blockquote markers (see _BQ_MARKER_RE) at the
    start of `s`, walking the running absolute column `col` through it,
    and report `(remainder, count, new_col)`, where `count` is the NUMBER
    of `>` markers consumed (0 if no run matched at all).

    Each marker consumes its `>` plus at most ONE COLUMN of the
    whitespace that follows — not one whitespace CHARACTER (Critical 1,
    pass-13 fix wave; see `_BQ_MARKER_RE` for the CommonMark citation).
    For a space the two are the same thing. For a TAB they are not: a tab
    starting at column c is worth `4 - c % 4` columns, and only its first
    belongs to the marker. When more than one column remains, the tab
    CHARACTER is deliberately left at the front of the remainder and only
    `col` is advanced past its first column: `_advance_col` expands a tab
    to the next multiple of 4 from wherever the cursor now sits, so the
    caller's own whitespace/gap accounting picks up exactly the columns
    the marker did not take, with no substitution or re-splicing of the
    line. A tab with exactly one column left (the cursor sits at 3 mod 4)
    is consumed outright — leaving it would re-advance the cursor a full
    tab stop it has already passed.

    The indentation BEFORE each marker — including a marker following an
    earlier `>` in the same run — is budgeted at `_BQ_GAP_MAX` columns
    (Critical 1, pass-14 fix wave). It used to be `_BQ_MARKER_RE`'s own
    unbounded `\\s*`, which meant `>` + 5 spaces + `>` counted as depth 2
    where CommonMark reads depth 1 with indented content, inflating both
    the recorded depth and the running column. Past the budget the run
    simply ENDS, and — because the gap is scanned on a trial cursor and
    only committed once a `>` is actually found there — no whitespace is
    consumed by a marker that did not materialise.

    When `expected` is given, the run must match that count EXACTLY — a
    fence opened N blockquote levels deep closes only at exactly N, never
    a different depth (shallower or deeper) — and a mismatch (including
    no run at all when one was expected) reports failure by returning
    `None` rather than a count.

    Critical 2 (pass-14 fix wave): "exactly N" is now consumed as well as
    demanded. This loop used to consume the WHOLE run and only then
    compare the total against `expected`, which made a fence carrying both
    a `bq_pre` and a `bq_post` structurally unclosable: on a closing line
    like `>   > ``` `, `is_closing_line`'s `bq_pre` call (expected=1)
    swallowed the nested `>` that belonged to `bq_post` too, reported 2,
    failed the comparison and returned `None` — for every candidate
    closer, at every indentation, so the fence never closed and every
    remaining line in the file was swallowed as its content. Stopping at
    `expected` leaves the nested marker in the remainder for the
    `bq_post` step that owns it. With `expected` unset the loop still
    walks the entire run, which is what `strip_container_prefix` wants:
    it is discovering the depth, not testing one.

    This is the one piece of "match a run, count the `>`s" knowledge
    `strip_container_prefix` and `is_closing_line` both need, twice each
    (bq_pre and bq_post) — factored so it is stated once, not four times.
    """
    pos = 0
    count = 0
    while True:
        # Trial-then-commit: the pre-marker gap is scanned onto a trial
        # cursor and only adopted once a `>` is confirmed to sit behind
        # it. A gap wider than `_BQ_GAP_MAX` leaves `gap_pos` short of the
        # `>`, so the match below fails and the run ends — which is the
        # intended reading (indented code, not a marker).
        gap_pos, _gap_width, gap_col = _scan_ws_capped(s, pos, col, _BQ_GAP_MAX)
        m = _BQ_MARKER_RE.match(s, gap_pos)
        if not m:
            break
        col = gap_col + 1  # the `>` itself is exactly one column
        pos = m.end()
        count += 1
        nxt = s[pos:pos + 1]
        if nxt and _WS_CHAR_RE.match(nxt):
            if nxt == "\t" and 4 - (col % 4) > 1:
                # A tab with columns to spare: the marker takes its first
                # column, the character stays for the caller's accounting.
                col += 1
            else:
                col = _advance_col(col, nxt)
                pos += 1
        if expected is not None and count == expected:
            break
    if expected is not None and count != expected:
        return None
    return s[pos:], count, col


def strip_container_prefix(line: str) -> Tuple[str, FenceContext, int, int]:
    """Strip, from the START of `line`, exactly the container syntax
    this parser understands, and report which pieces were consumed.

    Returns `(remainder, context, col)`, where `col` is the running
    ABSOLUTE column reached once every piece of container syntax this
    call actually consumed (leading whitespace, blockquote run, gap,
    list marker, checkbox, trailing blockquote run) has been walked —
    never recomputed later from a fresh slice assumed to start at column
    0 (see `_advance_col`). `classify_line` picks the cursor up exactly
    where this function leaves it to measure the residual whitespace
    before the fence run itself, rather than re-deriving a width from
    `remainder` in isolation.

    Order: leading whitespace, any number of blockquote markers, at
    most ONE list marker (with an optional checkbox immediately after
    it), then any number of further blockquote markers. The list-marker
    step is capped at one deliberately — a doubled bullet or doubled
    ordered marker is left unstripped rather than silently consumed
    twice (see classify_line's abort path).

    `bq_pre`/`bq_post` in the returned context are the COUNT of `>`
    markers consumed in each run (not merely whether the run matched),
    so a fence opened N levels deep can later be required to close at
    exactly N levels deep — see FenceContext and is_closing_line.

    Before attempting the list marker, up to `_LIST_GAP_MAX` columns of
    whitespace immediately in front of it are tolerated and stripped
    (Critical 2, pass-9 fix wave) — CommonMark allows 0-3 columns of
    plain indentation before a list marker, including directly after a
    blockquote run, and the blockquote marker itself only ever takes one
    such column on its own (see `_consume_bq_run`). A gap of 4+ columns is deliberately left
    untouched rather than silently absorbed: that is a different
    construct (indented content, not a nested list marker), so the list
    match below is attempted — and correctly fails — against the
    ORIGINAL, unstripped text, falling through to the same prose/abort
    resolution `classify_line` already applies to any other line whose
    container syntax does not resolve.

    `list_content_col` in the returned context is the COLUMN WIDTH consumed by the
    gap (if any was absorbed above) plus the list marker plus its optional
    checkbox, exactly as they appeared on THIS line — 0 when no list
    marker was found (Critical 2, pass-10 fix wave: the gap's width was
    previously dropped here, computed only from `list_candidate` — the
    text AFTER the gap was already sliced off — which left the recorded
    content column short by exactly the gap's width). See
    FenceContext's docstring for why this is needed on the closing
    side now, and why it is not the deleted ambient tracker.

    Critical 1 (pass-11 fix wave): `list_content_col` must ALSO include the width
    of the very first whitespace strip below — this function's outermost
    leading indentation — whenever `bq_pre == 0`. That first strip's width
    was previously measured and then discarded outright (not merely
    computed from the wrong slice, as at pass-10, but never captured at
    all). When no blockquote is present, that outer indentation IS the
    list item's own leading whitespace — the same column CommonMark
    counts toward the marker's content column — so omitting it left
    `list_content_col` short by exactly that width, and `is_closing_line`'s
    `resid_ws_width >= context.list_content_col` test then accepted a closing/
    continuation candidate indented that many columns too little. When a
    blockquote IS present (`bq_pre > 0`), this outer indentation is the
    PRE-`>` indentation, not indentation inside the list item — `classify_line`
    strips an equal-width run on the opening side of a container-prefixed
    line via its own leading-whitespace match, and `is_closing_line` strips
    an equal-width run before testing `bq_pre`, so that width is already
    accounted for symmetrically on both sides without being added to
    `list_content_col`. Adding it there too would double-count it and falsely
    require MORE indentation than the list item actually establishes — the
    mirror-image bug to the one this fix closes. Verified both branches
    (see the marker-indent x closer-column matrix in the harness).
    """
    s = line
    col = 0
    # `_MEASURE_ALL` (Critical 1, pass-14 fix wave) is the ONE
    # deliberately unbudgeted whitespace run in this parser, and it is
    # stated rather than obtained by omission — grep the name to
    # enumerate every such site. It is unbudgeted because both consumers
    # need the run's TRUE width in order to REFUSE it: `classify_line`
    # compares it against `_UNCONTAINED_FENCE_INDENT_MAX` and returns an
    # abort (exit 4) past the budget, and `is_closing_line` performs the
    # same comparison on the closing side. Clipping the run here would
    # hand both of them a width of 3 for every over-indented line and
    # silently destroy the refusal. The budget therefore lives at the two
    # points of use, which is precisely what the other four sites lacked.
    s, outer_ws_width, col = _consume_ws_capped(s, col, _MEASURE_ALL)

    s, bq_pre, col = _consume_bq_run(s, col)

    # Critical 2 (pass-9 fix wave): try the list match against a
    # gap-tolerant view first. If nothing matches there, `m` is None and
    # the code below falls back to matching directly against `s` — the
    # untouched, pre-gap-strip text — which reproduces the prior
    # (correct) behaviour for a line with no list marker at all, and
    # deliberately does NOT absorb a 4+-column gap into anything.
    #
    # Minor 9 (pass-13 fix wave): this is a deliberate TRY-THEN-COMMIT,
    # named as such rather than left as an implicit idiom. The gap has to
    # be measured before the code can know whether a list marker follows
    # it, so it is measured onto a TRIAL cursor (`gap_col`) that the real
    # `col` is only assigned from once the marker has actually matched
    # (Important 5, same wave). If no marker follows, the trial is simply
    # never read and the real cursor was never moved — where this used to
    # keep `col_after_bq_pre`/`gap_width` around and RECONSTRUCT a column
    # out of them in one branch while discarding them in the other.
    #
    # Critical 1 (pass-14 fix wave): the trial is now taken through
    # `_consume_ws_capped` with `_LIST_GAP_MAX` as its budget rather than
    # matching the whole run and comparing afterwards. Over-budget is
    # still all-or-nothing — a 4+ column gap must be left entirely
    # untouched, not clipped to 3 — and that is detected by asking whether
    # any whitespace SURVIVED the budgeted consume: if it did, the run was
    # wider than the budget and the trial is discarded. (A tab straddling
    # the budget stays in the remainder for the same reason, so it is
    # caught by the same test.) This is the same answer as before, reached
    # without ever measuring an unbudgeted run.
    list_candidate, gap_width, gap_col = s, 0, col
    trial_s, trial_width, trial_col = _consume_ws_capped(s, col, _LIST_GAP_MAX)
    if not _WS_RE.match(trial_s):
        list_candidate, gap_width, gap_col = trial_s, trial_width, trial_col

    # Whether a list marker was found gates the checkbox/bq_post strip
    # below (both only ever apply immediately after one). `list_content_col`
    # (unlike the old `has_list` boolean this function used to discard)
    # IS returned now — see FenceContext's docstring.
    #
    # Critical 2 (pass-10 fix wave): `list_content_col` must include `gap_width`,
    # the columns consumed by the gap-tolerant strip ABOVE, not merely the
    # marker+checkbox width measured from `list_candidate` (which starts
    # AFTER the gap was already sliced off). Before this fix, a fence
    # opened behind a 1-3 column gap recorded a content column exactly
    # `gap_width` columns short of where the list item's content actually
    # starts; `is_closing_line`'s `resid_ws_width >= context.list_content_col`
    # test then accepted a closing/continuation candidate indented one
    # (or more) columns too little as a valid close — a false EARLY close
    # that misreads the fence's real remaining content as fresh, unrelated
    # lines (see this function's own docstring and FenceContext's).
    #
    # Critical 1 (pass-12 fix wave): the marker's own width — and the
    # checkbox's, below — is now measured by advancing the RUNNING column
    # from wherever the gap left it (`gap_col`), not
    # by handing `_advance_col`/`_leading_ws_width` a fresh slice
    # (`list_candidate[:m.end()]`) and letting it assume that slice starts
    # at column 0. A tab inside a marker's own trailing whitespace (e.g.
    # `-\t`) or between a checkbox and the fence run expands relative to
    # its TRUE position on the line; measuring it from a slice's own start
    # under-or-over-counts it depending on how far into the line that slice
    # actually began (see `_advance_col`'s docstring — 50 under-counts, 40
    # over-counts on the reproduced shapes).
    #
    # Critical 1 (pass-14 fix wave): the marker's and the checkbox's own
    # trailing whitespace is no longer part of `_LIST_RE`/`_CHECKBOX_RE`
    # (an unbounded `\s+` in each, committed whole into
    # `list_content_col`). It is consumed by `_consume_list_marker_ws`,
    # which applies CommonMark's two list-item rules: 1-4 columns belong
    # to the marker's content column, 5+ columns mean the content column
    # is one past the marker and the rest is the item's own indented-code
    # indentation. Before this fix, `-` followed by five spaces recorded
    # `list_content_col = 6`, so a fence run sitting there was measured at
    # a relative indent of 0 and opened — where CommonMark reads content
    # column 2 and a relative indent of 4, i.e. literal indented code.
    # Anything the fence then swallowed (a numeric claim, an untagged
    # fence, the rest of the file) went unreported.
    bq_post = 0
    list_content_col = 0
    m = _LIST_MARKER_RE.match(list_candidate)
    if m and _WS_CHAR_RE.match(list_candidate[m.end():m.end() + 1] or ""):
        # The marker matched, so the trial gap above is committed: the
        # real cursor takes the trial's value and carries on from there.
        col = gap_col
        s, marker_width, col = _commit_span(list_candidate, m.end(), col)
        s, marker_ws_width, col = _consume_list_marker_ws(s, col)
        list_content_col = gap_width + marker_width + marker_ws_width
        if bq_pre == 0:
            # Critical 1 (pass-11 fix wave): fold in the outermost
            # leading-whitespace width captured above — see this
            # function's docstring for why this is gated on `bq_pre == 0`
            # (a blockquote's own pre-`>` indentation is already dropped
            # symmetrically on both the open and close side and must not
            # be added here as well).
            list_content_col += outer_ws_width

        # The checkbox is treated as an extension of the marker by this
        # grammar (a myflow tasks.md convention GFM itself parses as
        # inline text), so its trailing whitespace is budgeted by exactly
        # the same rule — 1-4 columns credited to the content column, 5+
        # meaning one column credited and the rest left as the item's own
        # indentation. Critical 1, pass-14 fix wave: this was an unbounded
        # `\s+` inside `_CHECKBOX_RE`, the third reproduction of the run.
        m2 = _CHECKBOX_MARKER_RE.match(s)
        if m2 and _WS_CHAR_RE.match(s[m2.end():m2.end() + 1] or ""):
            s, checkbox_width, col = _commit_span(s, m2.end(), col)
            list_content_col += checkbox_width
            s, checkbox_ws_width, col = _consume_list_marker_ws(s, col)
            list_content_col += checkbox_ws_width

        s, bq_post, col = _consume_bq_run(s, col)
    # No `else` branch: when no list marker follows, the trial gap was
    # never committed — neither `s` nor `col` was ever moved past the
    # `bq_pre` strip, so there is nothing to undo.

    return s, FenceContext(bq_pre, bq_post, list_content_col), col, outer_ws_width


def _starts_container_syntax(s: str) -> bool:
    """Does `s` (already whitespace-stripped) attempt recognised
    container syntax at its very start — a blockquote marker, a list
    marker, or a checkbox? Used both for candidate eligibility (does
    this line attempt container syntax at all) and, after one strip
    attempt, to tell an unresolved nested attempt (abort) apart from
    genuine prose (never an abort)."""
    return bool(
        _BQ_MARKER_RE.match(s) or _LIST_RE.match(s) or _CHECKBOX_RE.match(s)
    )


@dataclass(frozen=True)
class LineClass:
    kind: str  # "prose" | "fence" | "abort"
    char: str = ""
    run_len: int = 0
    info: str = ""
    context: FenceContext = EMPTY_CONTEXT
    # `reason` only matters when kind == "abort": "prefix" is the
    # original abort (a fence-like run sits behind a container prefix
    # this parser's grammar cannot resolve — a doubled list marker, a
    # bare checkbox); "indent" is Critical 1 of the pass-8 fix wave (a
    # BARE line — no container syntax of its own — reduces to a fence
    # run at 4+ columns, and whether a list is in scope cannot be
    # determined without the cross-line state this parser no longer
    # keeps); "outer" is Minor 6 of the pass-13 fix wave — the same
    # unknowable question on a line that DOES carry a container marker,
    # whose own container prefix begins 4+ columns in (Critical 2,
    # pass-12). It is a separate reason only because the remedy differs:
    # "indent"'s message offers adding a marker, which is nonsense advice
    # for a line that already has one. check_file uses this to choose the
    # right message.
    reason: str = ""


_PROSE = LineClass("prose")
_ABORT_PREFIX = LineClass("abort", reason="prefix")
_ABORT_INDENT = LineClass("abort", reason="indent")
_ABORT_OUTER_INDENT = LineClass("abort", reason="outer")


def classify_line(line: str) -> LineClass:
    """Classify a line that is NOT already inside an open fence.

    This is the inversion described in the module header: rather than
    matching the line's prefix against an allowlist of recognised
    shapes (which can only ever answer for prefixes someone already
    enumerated), it asks whether stripping the container syntax this
    parser actually understands reaches a fence run — and, when it does
    not, whether what is left still looks like an unresolved attempt at
    MORE container syntax (abort) or is simply prose that happens to
    mention a fence run later on the line (never an abort, regardless
    of what character led the line or whether it carries a valid
    container prefix of its own).

    The fence's own indentation is capped at
    `_UNCONTAINED_FENCE_INDENT_MAX` (CommonMark's 0-3 column budget).
    Which column it is measured against, and what happens when a BARE
    line (no container syntax of its own) exceeds the cap, is Critical 1
    of the pass-8 fix wave — an operator decision, not a rediscovered
    bug:

      - A line that carries its own container syntax (a blockquote
        and/or list marker on THIS line): measured relative to whatever
        `strip_container_prefix` leaves after stripping exactly that
        syntax — the container's own content column, established by the
        marker itself. Unambiguous; unchanged from before.
      - A BARE line, indentation 0-3: a fence, unconditionally. This is
        safe regardless of any enclosing list — if a list were in scope
        its content column would be >= 1, making the relative indent
        even smaller, still within the 0-3 budget. No cross-line
        knowledge needed.
      - A BARE line, indentation 4 or more: AMBIGUOUS. With no list in
        scope this is an indented code block; with a list in scope it
        could be a legitimate container-relative fence — and this parser
        keeps no cross-line state (see the module docstring's
        "Post-review reshape"/pass-8 history) that could tell the two
        apart. It refuses to guess: `classify_line` returns an "abort"
        classification (reason "indent") rather than silently picking
        either reading.

    An earlier version of this parser (pass 7) tried to resolve the last
    case with an ambient, top-level list-item content-column tracker
    threaded across lines (`_next_list_content_col`). That tracker
    itself produced three separate silent misses in one review pass — a
    thematic break (`- - -`) opening a phantom list scope, an uncapped
    trailing-whitespace gap after a marker inflating the tracked column,
    and a scalar (not a stack) losing all state on any nested-list
    dedent. The operator was shown a stack-based repair and chose
    deletion instead: a guard that cannot tell whether a list is in
    scope must refuse, not guess, and refusing is available here in a
    way it structurally is not for cross-line ambiguity elsewhere in
    this parser (see the "KNOWN LIMITATION" notes). This function now
    carries no cross-line state of any kind.
    """
    if not _FENCE_RUN_ANYWHERE_RE.search(line):
        return _PROSE

    ws_stripped = _WS_RE.match(line)
    rest0 = line[ws_stripped.end():] if ws_stripped else line

    # Does `rest0` (whitespace already stripped) attempt real container
    # syntax at all? Computed once, up front: it decides both (a) which
    # column the indentation cap below is measured against, and (b) —
    # unchanged from before — candidate eligibility for the abort path
    # further down.
    has_container_prefix = _starts_container_syntax(rest0)

    # One full container-prefix strip. In the no-container case this is
    # a no-op beyond the same whitespace `rest0` already stripped, so
    # `stripped` and `context` are safe to use unconditionally below.
    # `col` is where the running absolute-column cursor sits once that
    # strip finished — never re-derived from `stripped` in isolation
    # (Critical 1, pass-12 fix wave; see strip_container_prefix and
    # `_advance_col`).
    stripped, context, col, outer_ws_width = strip_container_prefix(line)

    # Minor 8 (pass-13 fix wave): the outer-indent cap below reads the
    # width `strip_container_prefix` already measured for exactly this
    # line's leading whitespace, rather than measuring the same run a
    # second time here. The two were always equal — that is the problem
    # with keeping both: one quantity computed in two places is the shape
    # every Critical on this code path has taken.
    ws_width = outer_ws_width

    # `stripped` may still carry RESIDUAL leading whitespace — the
    # container's own indentation before the fence run itself (e.g. the
    # 2 columns left over after a blockquote marker consumes `>` plus at
    # most one space: `>   ``` `). `_FENCE_RUN_START_RE` is anchored at
    # position 0 with no whitespace tolerance, so that residual run must
    # be stripped before testing for a fence run, exactly as `rest0` was
    # stripped once already for the no-container case above. Measured by
    # advancing the running cursor `col` (where `strip_container_prefix`
    # left it), not by handing a fresh slice of `stripped` to a
    # width function that would otherwise assume it starts at column 0
    # (Critical 1, pass-12 fix wave — this was one of the three broken
    # call sites the fix-wave brief named directly: `stripped[:resid_ws.end()]`).
    #
    # `_MEASURE_ALL` (pass-14): the second and last deliberately
    # unbudgeted run, and for the same reason as the outer one — the
    # width is measured in order to REFUSE past
    # `_UNCONTAINED_FENCE_INDENT_MAX` (the `rel_width` test below falls
    # through to the prose/abort decision), and the full remainder is
    # additionally what `_starts_container_syntax` must be tested against
    # further down (Critical 3, pass-10). Clipping either would destroy a
    # refusal rather than tighten one.
    fence_candidate, resid_ws_width, _resid_col = _consume_ws_capped(
        stripped, col, _MEASURE_ALL
    )

    m = _FENCE_RUN_START_RE.match(fence_candidate)
    if m:
        if has_container_prefix:
            # Critical 2 (pass-12 fix wave): the OUTER indentation — this
            # line's own leading whitespace, BEFORE any container syntax
            # is recognised at all — is capped at
            # `_UNCONTAINED_FENCE_INDENT_MAX` here too, not only on the
            # bare-line branch below. CommonMark allows 0-3 columns of
            # un-contained indentation at the top level; at 4+ columns
            # with no enclosing list in scope (a question this parser
            # deliberately cannot answer without the deleted cross-line
            # state — see this function's own docstring), the content is
            # an INDENTED CODE BLOCK, and any marker-shaped token sitting
            # there is literal code content, not real container syntax —
            # a marker happening to be present must not answer a question
            # ("is a list in scope?") the parser structurally cannot
            # answer. Before this fix, `has_container_prefix` alone gated
            # this branch, so a marker at outer indent 4+ was ALWAYS
            # measured only against its OWN container-relative indent
            # (`resid_ws_width`), silently accepting exactly the shape the
            # bare branch already refuses at the same outer indent — the
            # same unknowable question, answered two different ways
            # depending on whether an irrelevant token happened to be
            # present. Refusing here, unconditionally, before even
            # looking at `resid_ws_width`, makes the answer depend only on
            # the outer indent, exactly as the bare case already does.
            if ws_width > _UNCONTAINED_FENCE_INDENT_MAX:
                # Minor 6 (pass-13 fix wave): a DISTINCT abort reason from
                # the bare branch below. The question refused is the same
                # one ("is a list in scope?"), but this line demonstrably
                # carries a container marker, so the bare branch's message
                # — "no container prefix on this line", remedied by
                # "adding an explicit list/blockquote marker on this line"
                # — states a falsehood and names a fix the operator has
                # already applied. See check_file for the two texts.
                return _ABORT_OUTER_INDENT
            # Relative to what strip_container_prefix left after
            # consuming exactly the container syntax this line itself
            # carries — the container's own content column. Unambiguous:
            # this line's own syntax resolves it with no cross-line
            # knowledge needed.
            rel_width = resid_ws_width
            if rel_width <= _UNCONTAINED_FENCE_INDENT_MAX:
                run = m.group(1)
                return LineClass(
                    "fence", char=run[0], run_len=len(run),
                    info=fence_candidate[len(run):], context=context,
                )
            # Indentation exceeds the container's own 0-3 column budget:
            # an indented code block relative to that container, not a
            # fence — falls through to the prose/abort decision below
            # exactly like the plain "stripping never reached a fence
            # run" case.
        else:
            # BARE line — no container syntax of its own. Critical 1
            # (pass 8): 0-3 columns is a fence unconditionally (see
            # classify_line's docstring for why this is safe regardless
            # of any enclosing list); 4+ columns is ambiguous and this
            # parser now refuses to guess rather than consult the
            # (deleted) ambient list-column tracker.
            if ws_width <= _UNCONTAINED_FENCE_INDENT_MAX:
                run = m.group(1)
                return LineClass(
                    "fence", char=run[0], run_len=len(run),
                    info=fence_candidate[len(run):], context=context,
                )
            return _ABORT_INDENT

    # Stripping did not reach a fence run (or did, but only via
    # indentation deep enough to be an indented code block relative to
    # its own container — literal text, not a fence). Eligibility: this
    # line is a candidate at all only if its start attempted container
    # syntax the stripper recognises — anything else is prose that
    # merely mentions a fence run mid-sentence (a letter, a quote, a
    # parenthesis, bold markdown, an unterminated compound number, or —
    # per the indentation cap above — too many columns of indentation
    # relative to its container), never a candidate, never an abort.
    if not has_container_prefix:
        return _PROSE

    # It attempted container syntax and one full strip did not resolve
    # to a fence run. Distinguish a genuinely unresolved nested attempt
    # (the remainder STILL starts with more container syntax — a second
    # list marker glued behind the first, a bare checkbox with no list
    # marker ahead of it) — abort — from ordinary prose that happens to
    # follow a valid, single container prefix (the remainder starts with
    # an ordinary word) — never an abort, and never treated as fence
    # content either; it is scanned as normal prose for numeric claims.
    #
    # Tested against `fence_candidate` (residual whitespace already
    # stripped), not the raw `stripped` (Critical 3, pass-10 fix wave).
    # A list/blockquote/checkbox marker can sit behind residual whitespace
    # that `strip_container_prefix` deliberately did NOT absorb — a 4+
    # column gap between a blockquote run and a list marker, the one shape
    # `_LIST_GAP_MAX` intentionally leaves untouched (see
    # strip_container_prefix). `_starts_container_syntax` itself requires
    # its marker at position 0 with no leading-whitespace tolerance, so
    # testing it against `stripped` (gap still attached) silently read
    # that unresolved nested attempt as ordinary prose — the fence never
    # opened, and the ambiguity CommonMark cannot itself resolve (is a
    # list in scope behind that gap, or is this an indented code block?)
    # was decided silently instead of refused. Stripping the residual
    # whitespace first before this test — exactly as already done before
    # testing for a fence run — makes the two questions ("is this a
    # fence?" and "is this still-unresolved container syntax?") consult
    # the same normalised view of the remainder, and resolves this
    # ambiguity the same way the bare (no-container-prefix) case already
    # does at 4+ columns: a loud abort, never a silent guess. Ordinary
    # prose that merely mentions a fence run stays prose exactly as
    # before, because it never had a container-syntax token sitting at
    # this position in the first place, whitespace or no whitespace in
    # front of it.
    if _starts_container_syntax(fence_candidate):
        return _ABORT_PREFIX
    return _PROSE


def _advance_or_none(
    s: str, expected: int, col: int, max_ws_cols: int
) -> Optional[Tuple[str, int, int]]:
    """Consume one blockquote run of exactly `expected` markers from the
    start of `s`, then the whitespace immediately after it, and report
    `(remainder, residual_ws_width, new_col)` — or `None` if the run does
    not match `expected` exactly (see `_consume_bq_run`) or if that
    trailing whitespace is wider than `max_ws_cols` COLUMNS.

    `max_ws_cols` has no default (Critical 1, pass-14 fix wave). This
    trailing run was previously matched in full with `_WS_RE` and handed
    to `_commit_span`, i.e. consumed without a stated budget, and was
    left to whatever cap the caller happened to apply afterwards — the
    shape of every recurrence in this defect family. Over-budget is
    all-or-nothing (a refusal, not a clip): a partially consumed run
    would leave whitespace for the NEXT step's own budget to absorb a
    second time, which is how one over-indented closing line could
    accumulate slack across steps.

    `col` is the running absolute column (see `_advance_col`) at the
    point `s` begins; both the blockquote run and the trailing
    whitespace are walked through it in order, so a tab inside either
    span expands relative to where it TRUE sits on the line, never
    relative to `s`'s own start (Critical 1, pass-12 fix wave — this
    function used to measure the trailing whitespace with
    `_leading_ws_width(s[:m.end()])`, which assumed `s` began at column
    0; it does not, once a blockquote run has already been consumed by a
    PRIOR call in the same close test — see `is_closing_line`).

    Factored out of `is_closing_line`, which previously repeated this
    exact four-line shape (try, bail on `None`, unwrap, strip trailing
    whitespace) twice — once for `bq_pre`, once for `bq_post` — differing
    only in which `FenceContext` field it read. `_consume_bq_run` already
    factored out the regex-matching knowledge; this factors out the
    calling convention wrapped around it, so it is written once rather
    than twice (Important 4, pass-8 fix wave).
    """
    result = _consume_bq_run(s, col, expected)
    if result is None:
        return None
    s, _, col = result
    s, resid_ws_width, col = _consume_ws_capped(s, col, max_ws_cols)
    if _WS_RE.match(s):
        # Whitespace survived the budget: the run was wider than
        # `max_ws_cols` columns. Refuse rather than clip (see docstring).
        return None
    return s, resid_ws_width, col


def is_closing_line(line: str, context: FenceContext, fence_char: str, fence_len: int) -> bool:
    """CommonMark's closing rule, scoped to the container context the
    fence was actually OPENED in: a fence opened with no container
    prefix is closed only by an unprefixed run; a fence opened inside a
    blockquote is closed only by a line that continues that SAME
    blockquote — at the SAME depth, not merely "some" depth. `bq_pre`/
    `bq_post` on `context` are counts, and the closing line's own
    blockquote run must match that count exactly: a fence opened two
    blockquote levels deep (`> > ```bash`) is not closed by a
    one-level-deep line (`> ```), only by another two-level line. Depth
    mismatch in either direction (shallower OR deeper than the opener)
    fails the close and the line is read as fence content instead.

    The list marker itself never repeats on continuation lines (ordinary
    Markdown list-item continuation is by indentation, not by repeating
    the bullet) — but the COLUMN that marker established is mandatory,
    not optional (Critical 1, pass-9 fix wave). Before this fix, a fence
    opened on a checkbox line (content column 6, e.g. `- [ ] `) accepted
    a closing/continuation line indented to ANY column within the bare
    0-3 budget, including columns that are not actually inside the list
    item at all under CommonMark — the false-positive direction is the
    worse one, because it silently misreads unrelated dedented content as
    still being part of the fence. `context.list_content_col` (0 when no list
    marker was present at open time) is required as a flat, mandatory
    run of whitespace columns — consumed after `bq_pre`, before
    `bq_post` and before the ordinary 0-3 budget, exactly where that
    list marker's own width sat on the opening line. A closing candidate
    indented to fewer than `list_content_col` columns simply is not recognised as
    a close (the same "not resolved, read as content" outcome as any
    other depth mismatch here) — never a guess, and never the deleted
    ambient tracker, since `list_content_col` was captured from THIS fence's own
    opening line, not inferred from this content line's own prefix (see
    FenceContext's docstring).

    The closing run's OWN indentation, past `bq_pre`/`list_content_col`/
    `bq_post`, is capped at `_UNCONTAINED_FENCE_INDENT_MAX` columns —
    the same CommonMark budget `classify_line` applies on the opening
    side (Important 3, pass-8 fix wave). Before that cap existed,
    `is_closing_line` stripped leading whitespace with no limit at all,
    so an over-indented backtick run (six columns, well past any
    container's 0-3 budget) closed a fence early; the file's real closer
    was then read as a fresh, untagged opener. Measured against the
    residual whitespace immediately before the candidate fence-character
    run — after any required `bq_pre`/`list_content_col`/`bq_post` strip —
    exactly mirroring how `classify_line` measures the opening side.

    A single running absolute column, seeded at 0 (the true start of
    `line`), is threaded through every strip below (Critical 1, pass-12
    fix wave) via `_advance_col` and `_advance_or_none` — never
    recomputed from a fresh mid-line slice, which is what let a tab
    after `bq_pre`'s own marker run under/over-count the true closing
    column before this fix.
    """
    s = line
    col = 0
    # Critical 1 (pass-14 fix wave): budgeted, like every other
    # whitespace consume in this parser. The widest LEGAL leading run on
    # a closing line is the opener's list content column plus
    # CommonMark's own 0-3 columns — either of fence indentation (no
    # `bq_post`) or of indentation before the nested `>` (`bq_post` set,
    # where `_BQ_GAP_MAX` is the same 3). Anything wider cannot be a
    # closer at all, and refusing here rather than clipping means the
    # surplus is not left behind for a later step's budget to absorb a
    # second time.
    max_leading_ws = context.list_content_col + _UNCONTAINED_FENCE_INDENT_MAX
    s, resid_ws_width, col = _consume_ws_capped(s, col, max_leading_ws)
    if _WS_RE.match(s):
        return False

    if context.bq_pre:
        # Important 2 (pass-13 fix wave): the outer-indent cap, which
        # until now existed only on the OPENING side (Critical 2, pass
        # 12). The whitespace just consumed sits before this line's own
        # `>` marker, so it is un-contained indentation in exactly the
        # sense `classify_line` caps at 0-3 columns: at 4+ columns
        # CommonMark reads the line as an indented code block and its `>`
        # as literal text, which cannot continue — let alone close — a
        # blockquoted fence. Before this, an opener at outer indent 4 was
        # refused (exit 4) while a CLOSER at outer indent 4 was accepted
        # silently, so the same unknowable question got two answers
        # depending on which side of the fence it was asked from.
        #
        # This cap is applied to the blockquote prefix and NOWHERE ELSE,
        # deliberately:
        #   - with no container prefix at all (`bq_pre == 0` and
        #     `list_content_col == 0`), the whole leading run IS the
        #     fence's own indentation and the final
        #     `resid_ws_width > _UNCONTAINED_FENCE_INDENT_MAX` test below
        #     already caps it at the same 0-3 columns;
        #   - inside a list with no blockquote, the leading run is NOT
        #     un-contained indentation — it is the list item's own
        #     content column, and CommonMark measures the closing fence
        #     relative to THAT column, not to the line's start. The
        #     existing `list_content_col` test plus the 0-3 budget below
        #     already expresses that exactly. Capping the raw leading run
        #     there would refuse valid closers: a fence opened at
        #     `"  - ```"` has `list_content_col == 4`
        #     (the opener's own 2 columns of outer indentation folded
        #     in), so its legitimate closers run to 7 columns of
        #     indentation — 3 relative to the item's content column, and
        #     well inside CommonMark's budget.
        # The opening line's own outer indentation is deliberately NOT
        # compared against this line's, because CommonMark budgets those
        # 0-3 columns per line rather than requiring the two sides to
        # agree — which is why it is not recorded on the context at all
        # (Minor 9, pass-14 fix wave; see FenceContext).
        if resid_ws_width > _UNCONTAINED_FENCE_INDENT_MAX:
            return False
        result = _advance_or_none(s, context.bq_pre, col, max_leading_ws)
        if result is None:
            return False
        s, resid_ws_width, col = result

    # Critical 1 (pass-9 fix wave): the list marker's own content column,
    # recorded at open time, is a MANDATORY run of whitespace columns
    # here — a closing/continuation line indented to fewer columns than
    # the opener's list marker established is not inside that list item
    # under CommonMark, and must not be recognised as this fence's close.
    if context.list_content_col:
        if resid_ws_width < context.list_content_col:
            return False
        resid_ws_width -= context.list_content_col

    if context.bq_post:
        # Critical 3 (pass-14 fix wave): the residual left after
        # `list_content_col` was subtracted above is CAPPED HERE, before
        # `bq_post` is consumed. This branch used to overwrite
        # `resid_ws_width` with `_advance_or_none`'s result and never
        # look at the value it was discarding, so the whitespace between
        # the list item's content column and the nested `>` was the one
        # run in the whole close test with no cap on it at all: measured
        # 0 to 49+ extra columns accepted, where every other context
        # stops at 3. The same text as an OPENER aborts (exit 4), so the
        # two sides answered the same question differently — the exact
        # asymmetry the pass-13 wave was meant to have eliminated, left
        # standing in the one branch it did not reach.
        #
        # The budget is CommonMark's ordinary 0-3 columns of block-start
        # indentation before a `>` (`_BQ_GAP_MAX`), and it is the same
        # budget `_consume_bq_run` applies to that gap on the opening
        # side — stated in the same terms, on both sides, so the
        # asymmetry cannot come back by one side being edited alone.
        if resid_ws_width > _BQ_GAP_MAX:
            return False
        result = _advance_or_none(
            s, context.bq_post, col, _UNCONTAINED_FENCE_INDENT_MAX
        )
        if result is None:
            return False
        s, resid_ws_width, col = result

    if resid_ws_width > _UNCONTAINED_FENCE_INDENT_MAX:
        return False

    if fence_char == "`":
        m = re.match(r"^`+", s)
    else:
        m = re.match(r"^~+", s)
    if not m:
        return False
    run_len = len(m.group(0))
    if run_len < fence_len:
        return False
    tail = s[run_len:]
    return tail.strip() == ""


# --- Guard ---------------------------------------------------------------


class ProvenanceGuard:
    """Scans a change's planning artifacts for stated provenance,
    accumulating a single running failure count across every file in the
    scan — the count a mid-scan containment abort (exit 3, from
    `verify_scanned_file_containment`) reports before it exits, so an
    abort that fires after several files have already been scanned does
    not drop the summary of what was already found.

    `abort_messages` (Critical 2, pass-8 fix wave) is the analogous
    record for a content-classification abort (exit 4, from
    `classify_line` via `check_file`): before this fix, that abort called
    `sys.exit(4)` directly from inside the per-line scan loop, so it
    always terminated the ENTIRE scan on the spot — skipping every later
    change directory (its own claims never reported) and outranking a
    real violation already found elsewhere (exit 4 instead of exit 1,
    even though "violations found" is documented as the more severe
    code). `classify_line`'s abort now only records a message here and
    then `check_file` returns (stopping just that file's scan, the same
    shape Important 8 already established for an `OSError`/decode
    failure — see `check_file`'s own docstring); every OTHER change
    directory keeps being scanned. `main()` decides the final exit code
    once, after every directory has had a chance to be scanned: a real
    violation anywhere still outranks a classification abort.

    `containment_errors` (Critical 2, pass-11 fix wave) is the same
    accumulator shape applied to `verify_scanned_file_containment`'s eight
    `sys.exit(3)` sites, which — until this fix — each terminated the
    entire scan the moment the FIRST symlinked/escaping/non-regular
    candidate was found, in whatever directory sorted first: every LATER
    change directory was never scanned, and any violation it carried was
    never reported. Two review slots looked at this and disagreed on the
    remedy: one argued exit 3 must keep outranking everything else (a
    symlink escape must never be downgraded to a mere violation or
    environment code just because something else was also found), the
    other argued the scan itself must not stop. Both are correct and
    both are implemented: `verify_scanned_file_containment` now records the
    message here and returns (never exits), so `main()`'s scan loop —
    nested since the scan set widened, one candidate per entry in
    `SCANNED_FILENAMES` inside one iteration per change directory — moves
    on to the next CANDIDATE, not merely the next directory, exactly as
    it already does for `file_errors`; and `main()`'s end-of-scan
    priority decision checks `containment_errors` FIRST, before
    violations/aborts/file errors, so exit 3 still wins even though the
    scan no longer stops the instant it is found. Every violation, abort
    or file error recorded from OTHER directories — and from the refused
    candidate's own siblings in the SAME directory — is still printed:
    the containment refusal does not silence them, it only outranks them
    for the exit code.
    """

    def __init__(self) -> None:
        self.failures = 0
        self.abort_messages: list = []
        self.containment_errors: list = []

    def verify_scanned_file_containment(
        self, path: str, changes_dir: str, filename: str
    ) -> bool:
        """Confirms `path` really sits at exactly
        <changes_dir>/<name>/<filename> before it is opened —
        `find -L`-style traversal is what lets the walk descend into a
        symlinked change directory in the first place (a
        PR-author-controlled slug that never gets scanned would be a
        guard that reports clean having never opened the plan), but
        following links means the walk is no longer bounded to the named
        target. Two escapes are closed here rather than left for
        `check_file` to discover the hard way: the change directory
        itself resolving outside the expected path, and the scanned file
        itself being a symlink (an out-of-tree read at best, an
        unbounded-read hang at worst — a symlink to /dev/zero never
        yields EOF). The path's own BASENAME is confirmed too, so
        "exactly `<changes_dir>/<name>/<filename>`" is a claim every
        component of is actually checked.

        Both refuse (exit 3, via `containment_errors` — see this class's
        own docstring) rather than silently skipping the file, because a
        skip and a clean run are indistinguishable to the caller. Returns
        `False` on any containment refusal (the caller must not open
        `path`) and `True` when containment is confirmed; it no longer
        calls `sys.exit` itself (Critical 2, pass-11 fix wave) — doing so
        here stopped the ENTIRE scan at the first escaping/symlinked
        file, silently dropping every later change directory. The
        caller (`main()`) continues scanning the remaining directories
        either way, and decides the final exit code once, after every
        directory has had a chance to be scanned.

        `filename` is the artifact being confirmed — one of
        `SCANNED_FILENAMES`. Every refusal names it rather than a literal
        `tasks.md`, so an operator reading the message knows which of a
        change's three planning artifacts was refused.
        """
        if os.path.islink(path):
            print(
                f"{path}: {filename} is a symlink — refusing to open it "
                "(out-of-tree read / unbounded-read risk)",
                file=sys.stderr,
            )
            self.containment_errors.append(path)
            return False

        directory = os.path.dirname(path)
        name = os.path.basename(directory)

        # The BASENAME half of "sits at exactly
        # <changes_dir>/<name>/<filename>". Everything below confirms the
        # containing DIRECTORY; without this the method would confirm a
        # claim it never checked. `main()` builds every candidate from
        # `SCANNED_FILENAMES`, so no caller can reach this today — which is
        # precisely why it is cheap to assert rather than to weaken the
        # docstring: this is the security-relevant gate, and a future
        # caller that derives `filename` from anything less literal would
        # otherwise inherit a check that never looked at the name it was
        # told to confirm.
        if os.path.basename(path) != filename:
            print(
                f"{path}: is not the {filename} it was checked as — "
                "refusing to open it",
                file=sys.stderr,
            )
            self.containment_errors.append(path)
            return False

        # Minor 5 (pass-8 fix wave): `os.path.isdir` has the exact same
        # swallow-every-OSError-and-report-False shape already fixed
        # elsewhere in this method (see Important 9's `os.stat`/`os.lstat`
        # fixes above and Critical 1's fix in `main()` below) — a
        # permission-denied `stat` on `real_dir`/`real_changes_dir` would
        # report "not a directory", identically to a path that genuinely
        # isn't one, hiding which of the two actually happened. Both are
        # exit 3 either way (containment cannot be confirmed), but an
        # explicit `os.stat` + `stat.S_ISDIR` keeps this instance
        # consistent with the rest of the function rather than a narrower,
        # not-yet-converted survivor of the same shape.
        real_dir = os.path.realpath(directory)
        try:
            dir_st = os.stat(real_dir)
        except OSError:
            print(f"{path}: cannot resolve containing directory", file=sys.stderr)
            self.containment_errors.append(path)
            return False
        if not stat.S_ISDIR(dir_st.st_mode):
            print(f"{path}: cannot resolve containing directory", file=sys.stderr)
            self.containment_errors.append(path)
            return False

        real_changes_dir = os.path.realpath(changes_dir)
        try:
            changes_st = os.stat(real_changes_dir)
        except OSError:
            print(f"{path}: cannot resolve {changes_dir}", file=sys.stderr)
            self.containment_errors.append(path)
            return False
        if not stat.S_ISDIR(changes_st.st_mode):
            print(f"{path}: cannot resolve {changes_dir}", file=sys.stderr)
            self.containment_errors.append(path)
            return False

        expected = os.path.join(real_changes_dir, name)
        if real_dir != expected:
            print(
                f"{path}: resolves to {real_dir}, outside {expected} — "
                "refusing to scan a path reached through a symlink escape",
                file=sys.stderr,
            )
            self.containment_errors.append(path)
            return False

        # Not `os.path.isfile(path)`: like `os.path.lexists`/`os.path.isdir`
        # elsewhere in this guard, `isfile` swallows every `OSError` (a
        # permission-denied `stat` included) and reports `False` — the
        # same message this guard would then print for a GENUINE
        # non-regular file (a FIFO, a device node). Exit 3 is correct
        # either way (both are containment-adjacent refusals to open
        # something at this path), but the message should not claim "not
        # a regular file" when the real reason is "could not tell" (see
        # Important 9's identical fix for the `os.listdir` case above).
        try:
            st = os.stat(path)
        except OSError as exc:
            print(
                f"{path}: cannot stat after resolution ({exc}) — refusing to open it",
                file=sys.stderr,
            )
            self.containment_errors.append(path)
            return False

        if not stat.S_ISREG(st.st_mode):
            print(
                f"{path}: not a regular file after resolution — refusing to open it",
                file=sys.stderr,
            )
            self.containment_errors.append(path)
            return False

        return True

    def _check_numeric_claim(
        self,
        text: str,
        lines: list,
        i: int,
        total: int,
        relfile: str,
        lineno: int,
    ) -> None:
        """Report an unattributed numeric claim in `text` — either a fence's
        own info string (the opening line's content after its tag, checked
        while `in_fence` is being set) or an ordinary prose line outside any
        fence. Both call sites need the identical shape: does `text` contain
        a numeric claim (`CLAIM_RE`) and, if so, does a `measured:`/
        `predicted:` comment (`PROVENANCE_RE`) appear within the two-line
        lookahead window starting at `lines[i]` (the claim's own line, or
        one of the two after it)?

        Every `CLAIM_RE` match is examined, not merely the first: a match
        that `_is_quoted` reports as enclosed in a delimiter pair is a
        number the line REPRODUCES rather than asserts, and skipping it
        must not stop the scan — otherwise a quoted number would mask a
        bare one sitting beside it on the same line.

        The line's quotation regions are computed ONCE, here, and threaded
        into every `_is_quoted` call (Important 3, pass-15 fix wave). This
        loop used to call `_is_quoted(text, ...)` per match, and each call
        recomputed every delimiter class on the line from scratch — the
        per-match half of the quadratic that made a single crafted line a
        denial of service on a CI gate. `quotation_regions` is also skipped
        entirely when the line has no match at all, which is almost every
        line.

        WHEN THE EXEMPTION WAS WITHDRAWN, SAY SO (Lens C, pass-17 fix wave).
        There are several distinct routes to the message below — no
        exemption was ever possible, a delimiter class did not fully pair, an
        escaped delimiter vetoed the line, raw HTML vetoed the line — and
        only the first is what the message alone implies. The distinguishing
        fact exists right here in `scan.veto` and used to be discarded before
        printing, so an author whose correctly-quoted number lost its
        exemption to an unrelated stray delimiter elsewhere on the line had
        no way to learn that from the tool, and the remedy banner's only
        advice would have had them write a provenance tag for a number that
        was already quoted. The note is printed alongside the violation, on
        its own line, and changes neither the exit code nor the failure
        count: this is message only.

        Extracted (Important 7, pass-12 fix wave) because this exact
        lookahead-and-print shape was duplicated near-verbatim in
        `check_file` for the info-string case and the prose case, with a
        comment noting the duplication rather than removing it — this
        file's own history is defects from "a value computed once, needed
        twice, not kept in sync" (see `_advance_col`'s docstring for the
        column-width instance of the same shape), so a duplicated CHECK is
        exactly as risky as a duplicated computed value: a future edit to
        the lookahead window or the message text could easily update one
        copy and miss the other.
        """
        claim = None
        scan = None
        for candidate in CLAIM_RE.finditer(text):
            if scan is None:
                scan = quotation_regions(text)
            if not _is_quoted(candidate.start(), candidate.end(), scan):
                claim = candidate
                break
        if claim is None:
            return
        satisfied = False
        for j in range(i, min(i + 3, total)):
            if PROVENANCE_RE.search(lines[j]):
                satisfied = True
                break
        if not satisfied:
            print(
                f"{relfile}:{lineno}: numeric claim with no "
                "measured:/predicted: provenance comment"
            )
            if scan.veto is not None:
                print(_veto_note(relfile, lineno, scan.veto))
            self.failures += 1

    def check_file(self, path: str, repo_root: str) -> Optional[str]:
        """Walks the file line by line, tracking fence state with
        CommonMark's closing rule scoped to the fence's own open-time
        container context (see is_closing_line). Both the open and the
        close are recognised regardless of leading indentation or
        container prefix — see classify_line/strip_container_prefix
        above.

        The read is split into two classified failure modes, neither of
        which is allowed to reach `main()` as an uncaught exception: a
        traceback is indistinguishable from this guard's own exit 1
        ("violations found") to a caller that inspects only the exit
        code, and this is the one script whose entire purpose is an exit
        contract a caller can trust. Both are ENVIRONMENT failures (exit
        2 — see the module docstring's exit-code enumeration; an earlier
        version of THIS docstring said exit 4 for the decode case, which
        contradicted both the actual `sys.exit(2)` below and the module
        docstring — corrected, not restated, since two docstrings in one
        file disagreeing about the exit contract is exactly the defect
        class this guard exists to prevent):

          - `OSError` (permission denied, I/O error, and similar) after
            containment has already passed: the guard cannot determine
            anything about this file's content, the same "cannot know"
            class as a missing `openspec/changes/` — not a malicious
            payload (that is exit 3, already ruled out by
            `verify_scanned_file_containment` above) and not a parsing
            question (that is exit 4, `classify_line`'s abort path).
          - Content that cannot be decoded as UTF-8: scanning it anyway
            with `errors="replace"` would silently substitute
            replacement characters into the very claims this guard
            exists to read correctly, which could mask a real,
            unattributed claim. Refusing outright is the same
            fail-loud principle `classify_line`'s abort path already
            follows for an unrecognised container prefix.

        Neither failure mode calls `sys.exit` directly any more (that
        was Important 8's defect: one bad file terminated the entire
        scan, silently dropping every OTHER change directory from being
        scanned and folding any violations already found into a bare
        environment exit). Instead this method returns the classified
        message as a string and lets `main()` decide the exit AFTER
        every file has been given a chance to scan: if real violations
        were found anywhere, they are reported (exit 1) alongside the
        unreadable file, never masked by it; only when nothing else was
        found does the file failure alone become exit 2. A `None` return
        does NOT mean this file was fully scanned (Minor 7, pass-9 fix
        wave corrects an earlier version of this sentence that said it
        did): it means no OSError/decode failure was returned as a
        string. `None` is also the return value when `classify_line`
        aborts below and this method returns early — in that case the
        file was scanned only up to the abort point, not to EOF, and
        `self.abort_messages` (not this return value) is what records
        that it happened. This method may still have recorded ordinary
        provenance violations in `self.failures` either way.

        `encoding="utf-8-sig"` strips a single leading BOM (U+FEFF) if
        present and is a no-op otherwise. Without it, a BOM-prefixed
        fence-opener line falls through `classify_line`'s whitespace/
        container-prefix stripping (U+FEFF is Unicode category Cf, not
        whitespace, and is not container syntax), so the line is
        misread as prose and the fence never opens; the next BARE fence
        run in the file is then read as a fresh, untagged opener, and
        everything after it — including a genuinely unattributed
        numeric claim — is silently swallowed as that opener's content.
        That is a silent miss, not a false positive, and BOM-prefixed
        Markdown is routine output from editors and Windows tooling.
        """
        relfile = os.path.relpath(path, repo_root)

        try:
            size = os.path.getsize(path)
        except OSError as exc:
            # Minor 6 (pass-11 fix wave): `str(exc)` on an `OSError` raised
            # from a call on `path` embeds `path` itself — the ABSOLUTE
            # path — inside the parenthesised errno text (e.g. "[Errno 13]
            # Permission denied: '/abs/.../tasks.md'"), even though `msg`'s
            # own prefix is `relfile`, repo-relative. `exc.strerror` is
            # just the OS's errno description with no filename attached,
            # consistent with the repo-relative framing this message
            # otherwise establishes; falls back to `str(exc)` only for the
            # rare OSError that carries no `strerror` at all.
            msg = f"{relfile}: cannot stat file: {exc.strerror or exc}"
            print(msg, file=sys.stderr)
            return msg

        if size > MAX_SCANNED_FILE_BYTES:
            msg = (
                f"{relfile}: {size} bytes exceeds the "
                f"{MAX_SCANNED_FILE_BYTES}-byte cap — refusing to read it "
                "unbounded"
            )
            print(msg, file=sys.stderr)
            return msg

        try:
            with open(path, "rb") as f:
                raw = f.read()
        except OSError as exc:
            # Minor 6 (pass-11 fix wave): see the identical fix just above
            # for `os.path.getsize`'s OSError — `exc.strerror` avoids
            # re-embedding the absolute `path` in a message whose prefix
            # is deliberately repo-relative.
            msg = f"{relfile}: cannot read file: {exc.strerror or exc}"
            print(msg, file=sys.stderr)
            return msg

        try:
            text = raw.decode("utf-8-sig")
        except UnicodeDecodeError as exc:
            msg = (
                f"{relfile}: cannot decode as UTF-8 ({exc}) — refusing to "
                "scan with replacement characters that could mask a claim"
            )
            print(msg, file=sys.stderr)
            return msg

        lines = text.split("\n")
        # split() on a trailing newline produces one extra empty
        # trailing element; drop it so line counts match a file
        # that ends with a single trailing newline (the ordinary
        # case), mirroring the Bash version's `read -r` loop.
        if lines and lines[-1] == "":
            lines.pop()

        total = len(lines)

        in_fence = False
        fence_char = ""
        fence_len = 0
        fence_tagged = False
        fence_start_line = 0
        fence_context = EMPTY_CONTEXT

        i = 0
        while i < total:
            lineno = i + 1
            line = lines[i]

            if in_fence:
                # Critical fix (defect 9 / defect 7): while inside an
                # open fence, a line is CONTENT, never a fence-OPENING
                # candidate, and therefore never an abort candidate —
                # classify_line (with its eligibility gate and abort
                # path) is for detecting ambiguous OPENERS and is
                # deliberately never consulted here. The only question
                # for a content line is the CommonMark close test,
                # scoped to the container context recorded when this
                # fence was opened.
                if is_closing_line(line, fence_context, fence_char, fence_len):
                    if not fence_tagged:
                        print(
                            f"{relfile}:{fence_start_line}: fenced code block "
                            "has no verified:/unverified: tag"
                        )
                        self.failures += 1
                    in_fence = False
                    fence_char = ""
                    fence_len = 0
                    fence_tagged = False
                    fence_context = EMPTY_CONTEXT
                i += 1
                continue

            # Not in a fence: classify_line may legitimately abort here —
            # the three cases this guard refuses to guess about are an
            # ambiguous OPENING prefix (reason "prefix"), a bare
            # fence-like run indented 4+ columns with no container
            # syntax of its own (reason "indent" — Critical 1, pass-8 fix
            # wave; see classify_line's docstring), and a line whose OWN
            # container prefix starts 4+ columns in (reason "outer" —
            # Critical 2, pass-12, given its own message at Minor 6,
            # pass-13, because the "indent" text is false for it).
            # Critical 2 (same
            # wave): this no longer calls `sys.exit(4)` directly — doing
            # so from inside this loop skipped every later change
            # directory outright and outranked a real violation found
            # elsewhere. The message is recorded, and — exactly like the
            # OSError/decode-failure return below (Important 8's
            # pattern) — this method returns rather than continuing to
            # read THIS file: once one line's fence-open-or-close
            # reading is unknown, whether any later line is fence
            # CONTENT or fresh prose is unknown too, and guessing past it
            # would fabricate misleading violations of its own (an
            # unrelated later bare fence run misread as a fresh,
            # never-closed opener). Every OTHER change directory is
            # unaffected and keeps being scanned; `main()` decides the
            # final exit code once, after every directory has had a
            # chance to be scanned, so a real violation found in another
            # file still outranks this abort (exit 1 over exit 4).
            cls = classify_line(line)
            if cls.kind == "abort":
                # Important 6 (pass-9 fix wave): the message now says what
                # to CHANGE, not only what the guard could not resolve —
                # exit 4 is routine by the operator's own fail-loud design
                # (Critical 1, pass-8 fix wave), so the population that
                # hits it just grew, and a diagnostic with no remedy is a
                # worse deal for that larger population. Critical 3 (pass-9
                # fix wave): the message also says scanning of THIS FILE
                # stopped here — no line after this point in this file is
                # scanned on this run (see check_file's docstring and
                # skills/myflow-contracts/plan-provenance.md's "What the
                # guard does not do") — rather than leaving that only
                # implied by the fact that no more of this file's output
                # appears.
                if cls.reason == "indent":
                    msg = (
                        f"{relfile}:{lineno}: fence-like run indented 4+ "
                        "columns with no container prefix on this line — "
                        "cannot determine whether a list is in scope; "
                        "refusing to guess whether it opens or closes a "
                        "fence. Fix by dedenting to 0-3 columns, or by "
                        "adding an explicit list/blockquote marker on this "
                        "line so the container is unambiguous. Scanning of "
                        "this file stopped here; re-run after fixing to see "
                        "anything further on in it."
                    )
                elif cls.reason == "outer":
                    # Minor 6 (pass-13 fix wave): the container-prefix-but-
                    # over-indented case, which used to be handed the
                    # "indent" text above verbatim — telling an operator
                    # whose line reads "\t- \t```bash" both that it has "no
                    # container prefix on this line" and to fix it by
                    # "adding an explicit list/blockquote marker on this
                    # line". Both clauses are false for this branch, and
                    # this is the third instance of one message reused
                    # across two causes (pass 10's remedy string, pass 11's
                    # summary verb), so the harness now asserts this text,
                    # not merely the exit code it accompanies.
                    msg = (
                        f"{relfile}:{lineno}: fence-like run behind a "
                        "container prefix starting 4+ columns in — at that "
                        "outer indentation CommonMark reads the whole line "
                        "as an indented code block, so the list/blockquote "
                        "marker on it is literal text rather than container "
                        "syntax; cannot determine whether a list is in "
                        "scope; refusing to guess whether it opens or closes "
                        "a fence. Fix by dedenting this line — marker "
                        "included — to 0-3 columns. Scanning of this file "
                        "stopped here; re-run after fixing to see anything "
                        "further on in it."
                    )
                else:
                    # Important 4 (pass-10 fix wave): this used to share
                    # the "indent" branch's remedy verbatim, including
                    # "Fix by dedenting to 0-3 columns" — but indentation
                    # is never the cause here. This branch fires for a
                    # doubled/glued marker (a second list marker right
                    # behind the first, e.g. "- - ```bash"), a bare
                    # checkbox with no preceding list marker (e.g.
                    # "[ ] ```bash"), or — since Critical 3 of this same
                    # wave — a list/blockquote marker sitting behind a gap
                    # wider than this parser's 0-3 column budget (e.g.
                    # a blockquote followed by 4+ columns then "- ```").
                    # A doubled marker at column 0 told to "dedent" names
                    # a fix that does not exist for its actual cause; name
                    # the real ones instead.
                    msg = (
                        f"{relfile}:{lineno}: line has a fence-like run of "
                        "backticks/tildes behind a prefix this guard does not "
                        "recognise — a doubled/glued container marker, a bare "
                        "checkbox with no preceding list marker, or a list "
                        "marker sitting behind a gap wider than 3 columns — "
                        "refusing to guess whether it opens or closes a "
                        "fence. Fix by removing the doubled marker, adding "
                        "the missing list marker before the checkbox, or "
                        "narrowing the gap to 0-3 columns, so the container "
                        "is unambiguous. Scanning of this file stopped here; "
                        "re-run after fixing to see anything further on in "
                        "it."
                    )
                print(msg, file=sys.stderr)
                self.abort_messages.append(msg)
                return None

            if cls.kind == "fence" and cls.run_len >= 3:
                in_fence = True
                fence_char = cls.char
                fence_len = cls.run_len
                fence_start_line = lineno
                fence_context = cls.context
                fence_tagged = bool(FENCE_TAG_RE.search(cls.info))

                # A numeric claim can sit on the fence's own OPENING line,
                # in its info string (e.g. ```bash verified:ran 500 tests).
                # This is otherwise never checked: the info string is only
                # ever tested against FENCE_TAG_RE above, and once in_fence
                # is True this same line is never re-examined as prose. Use
                # the identical lookahead window as the prose case below —
                # the line itself is lines[i], so the window is the same
                # range.
                self._check_numeric_claim(cls.info, lines, i, total, relfile, lineno)
            else:
                # Outside a fence: check for an unattributed numeric
                # claim, within the two-line lookahead window.
                self._check_numeric_claim(line, lines, i, total, relfile, lineno)

            i += 1

        # A fence left open at end of file is malformed input, not this
        # guard's concern to diagnose further than leaving it unflagged
        # as a fence — but report it so it is not silently ignored.
        if in_fence:
            print(f"{relfile}:{fence_start_line}: fenced code block never closed")
            self.failures += 1

        return None


# _EXIT_PRIORITY — the full priority ladder, highest first, as DATA rather
# than as logic repeated once per channel. Each entry is
# (exit_code, channel_name); `_resolve_exit_code` below walks this list
# exactly once and returns the exit code of the first channel that has
# anything to report. See `_resolve_exit_code`'s own docstring (Important
# 3, pass-12 fix wave) for why this collapses four hand-written
# priority-ordering blocks into one.
_EXIT_PRIORITY = (
    (3, "containment_errors"),
    (1, "failures"),
    (4, "abort_messages"),
    (2, "file_errors"),
)


def _report_containment_refusal(guard: "ProvenanceGuard", file_errors: list) -> None:
    """Print the exit-3 summary and exit 3.

    Two call sites reach exit 3 (Critical 5, pass-14 fix wave): the
    end-of-scan ladder below, and `main()`'s validation of the
    `openspec/changes` ANCHOR itself, which fires before any directory is
    scanned. Stated once here rather than twice, for the reason this file
    keeps rediscovering — a message and an exit code written in two places
    drift, and this guard's whole value is an exit contract a caller can
    trust. The counts it prints are read from `guard`, so the
    before-the-scan call naturally reports zeroes for channels that never
    had a chance to accumulate anything.
    """
    if guard.failures != 0:
        print(f"\n{guard.failures} plan-provenance violation(s) found.", file=sys.stderr)
    if guard.abort_messages:
        print(
            f"{len(guard.abort_messages)} line(s) could not be classified "
            "(content-classification, see above).",
            file=sys.stderr,
        )
    if file_errors:
        print(
            f"{len(file_errors)} file(s) could not be read "
            "(environment failure, see above).",
            file=sys.stderr,
        )
    print(
        f"\n{len(guard.containment_errors)} path(s) failed containment "
        "(symlinked, escaping their change directory, or non-regular — see "
        "above) — refusing to report anything but a security failure. Every "
        "violation, abort or file error listed above from OTHER directories "
        "was still found and still needs fixing; fix the containment "
        "escape(s) first.",
        file=sys.stderr,
    )
    sys.exit(3)


def _resolve_exit_code(
    containment_errors: list, failures: int, abort_messages: list, file_errors: list
) -> int:
    """Decide the single exit code `main()` reports, from the full ladder
    — 3 (containment/security) > 1 (violations) > 4
    (content-classification) > 2 (environment) > 0 (clean) — in exactly
    ONE place (Important 3, pass-12 fix wave).

    Before this fix, `main()` decided this same ladder via four
    hand-written `if` blocks, each one restating both the priority order
    AND the "N other things were also found" knowledge as its own
    ad-hoc chain of conditions. A fourth failure channel
    (`containment_errors`, added at pass 11) already had to be threaded
    by hand through all four of those blocks; Lens B of the pass-11
    review predicted a further channel would eventually arrive and called
    the four-copies shape blocking for exactly that reason. It has now
    arrived (this pass's own review flagged the shape, not a new
    channel) — a fifth channel from here on is one more tuple in
    `_EXIT_PRIORITY` and one more `elif` in `main()`'s print dispatch,
    never a fifth copy of the ladder itself.

    This function decides ONLY the winning exit code — which channel's
    findings are severe enough to determine the process's own exit
    status. It does not print anything and does not decide which
    channels get REPORTED: every channel with anything in it is still
    printed by `main()` regardless of which one wins (see the module
    docstring's "every code below 3 in this ladder can still be found and
    printed alongside a higher one that wins the exit"); this function
    answers only "which single number does the process exit with".

    Returns 0 (clean) when none of the four channels has anything to
    report.
    """
    channel_present = {
        "containment_errors": bool(containment_errors),
        "failures": failures != 0,
        "abort_messages": bool(abort_messages),
        "file_errors": bool(file_errors),
    }
    for exit_code, name in _EXIT_PRIORITY:
        if channel_present[name]:
            return exit_code
    return 0


def resolve_repo_root() -> str:
    """REPO_ROOT is always resolved from this script's own location by
    default — this is what makes the guard argument-free and self-scoped
    from any cwd. CHECK_PLAN_PROVENANCE_ROOT is an explicit, opt-in
    override honoured only when set — it exists solely so the companion
    harness (test-check-plan-provenance.sh) can point the guard at a
    sandboxed fixture tree under TMPDIR without touching this repo; never
    set it for a normal invocation.
    """
    if "CHECK_PLAN_PROVENANCE_ROOT" in os.environ:
        root = os.environ["CHECK_PLAN_PROVENANCE_ROOT"]
        if root == "":
            print("CHECK_PLAN_PROVENANCE_ROOT is set but empty", file=sys.stderr)
            sys.exit(2)
        return root

    script_dir = os.path.dirname(os.path.realpath(__file__))
    return os.path.dirname(script_dir)


def main() -> None:
    # Keep stdout/stderr interleaving honest when the caller merges both
    # streams (e.g. `... 2>&1`) — Python block-buffers stdout by default
    # once it is not a tty.
    try:
        sys.stdout.reconfigure(line_buffering=True)
    except AttributeError:
        pass

    repo_root = resolve_repo_root()

    # A root that is not a directory would make the scan find nothing and
    # the guard report a clean run over zero files — a false "all clear",
    # which is the one outcome a guard must never produce.
    if not os.path.isdir(repo_root):
        print(f"not a directory: {repo_root}", file=sys.stderr)
        sys.exit(2)

    guard = ProvenanceGuard()
    changes_dir = os.path.join(repo_root, "openspec", "changes")

    # Critical 5 (pass-14 fix wave): validate the ANCHOR itself before
    # anything is anchored on it.
    #
    # `verify_scanned_file_containment` decides every per-file containment
    # question by comparing against `os.path.realpath(changes_dir)` — and
    # never checked that `changes_dir` was itself where it claimed to be.
    # `os.path.isdir` below follows symlinks, so `openspec/changes` being
    # a symlink was simply accepted, and every subsequent comparison was
    # made against the symlink's TARGET. Measured consequences:
    #
    #   - `changes/` -> an empty decoy inside the repo, with the real
    #     plan elsewhere in the tree: exit 0, "nothing in flight". A
    #     FALSE CLEAN RUN — the one outcome this module's docstring says
    #     must never happen.
    #   - `changes/` -> a directory outside the repo root: read out of
    #     tree, exit 1, with `openspec/changes/...`-relative paths in the
    #     output that actively conceal the escape.
    #
    # The strictly WEAKER version of the same escape — a symlink one
    # level deeper, at `changes/<name>` — was already refused with exit 3
    # by `verify_scanned_file_containment`. Refusing the weak form and
    # following the strong one is not a defensible line. (The security
    # slot rated this Important because the trigger is conspicuous in a
    # diff; a reviewer noticing a symlink is not a control this guard
    # provides, and severity follows the outcome, so it is Critical.)
    #
    # Recorded into `containment_errors` rather than exiting here, so it
    # resolves to exit 3 through the existing priority ladder — the same
    # code, from the same accumulator, as every other containment refusal.
    real_repo_root = os.path.realpath(repo_root)
    expected_changes_dir = os.path.join(real_repo_root, "openspec", "changes")
    if os.path.islink(changes_dir):
        print(
            f"{changes_dir}: openspec/changes is a symlink — refusing to "
            "scan through it (a decoy target reports a false clean run; an "
            "out-of-tree target reads outside the repository)",
            file=sys.stderr,
        )
        guard.containment_errors.append(changes_dir)
        _report_containment_refusal(guard, [])
    real_changes_dir = os.path.realpath(changes_dir)
    if os.path.isdir(changes_dir) and real_changes_dir != expected_changes_dir:
        print(
            f"{changes_dir}: resolves to {real_changes_dir}, not "
            f"{expected_changes_dir} — refusing to scan a changes "
            "directory reached through a symlink escape",
            file=sys.stderr,
        )
        guard.containment_errors.append(changes_dir)
        _report_containment_refusal(guard, [])

    # No openspec/changes directory at all: genuinely cannot determine
    # anything — the same "cannot know" case as REPO_ROOT not being a
    # directory above.
    if not os.path.isdir(changes_dir):
        print(
            f"openspec/changes not found under {repo_root} — refusing to "
            "report a clean run",
            file=sys.stderr,
        )
        sys.exit(2)

    # Count non-archived change directories FIRST, and treat zero of them
    # as a genuine, ordinary state — not an error. Symlinked entries are
    # followed (os.path.isdir follows symlinks, mirroring `find -L`'s
    # -type d), so a symlinked change directory (PR-author-controlled
    # slug) is counted like any other rather than silently skipped.
    #
    # Zero non-archived changes is this repository's steady state between
    # `/myflow-finish` and the next `/myflow-start` — "nothing in flight"
    # is not "cannot determine", and reporting exit 2 for it would make
    # this guard's declared place in `## lint` spuriously red on an
    # ordinary day. It is NOT the same state as "changes exist but none
    # of them yielded any scanned file" below, which stays a hard
    # failure: that shape means a broken glob or a planning artifact at
    # the wrong depth, not "nothing to do".
    try:
        entries = sorted(os.listdir(changes_dir))
    except OSError as exc:
        # Important 9: an `OSError` here is NOT always "not found" — a
        # `changes_dir` that exists but is unreadable (permission denied)
        # is a different, distinguishable failure, and printing "not
        # found" for it points an operator at the wrong fix (a path
        # problem when the real one is a permissions problem). Exit code
        # is unchanged (2, "cannot determine anything"); only the
        # message is corrected to name what actually happened.
        if isinstance(exc, PermissionError):
            print(
                f"{changes_dir}: permission denied ({exc}) — refusing to "
                "report a clean run",
                file=sys.stderr,
            )
        else:
            print(
                f"openspec/changes not found under {repo_root} — refusing to "
                "report a clean run",
                file=sys.stderr,
            )
        sys.exit(2)

    # Critical 1 fix: build change_dirs from an explicit, exception-aware
    # `os.stat` per entry rather than `os.path.isdir`, which SWALLOWS
    # every `OSError` (permission denied on the entry, a dangling symlink
    # target, a circular symlink/ELOOP) and reports `False` — exactly the
    # same value it reports for an entry that legitimately isn't a
    # directory (a stray regular file). Those two cases must not be
    # conflated: an entry that IS present in the directory listing but
    # cannot be classified is a SCAN-INTEGRITY failure (a PR author can
    # otherwise hide an entire change behind one broken/circular symlink,
    # or a `chmod 000`, and the guard would report "all provenance
    # stated" having never opened it — the false clean run the module
    # docstring calls "the one outcome a guard must never produce"), not
    # a skip. A genuine non-directory entry (`os.stat` succeeds, mode
    # says regular file) is still skipped silently — that part of the
    # original behaviour was correct and stays.
    #
    # Critical 1 (pass-10 fix wave): this used to call `sys.exit(2)`
    # directly, right here, inside the scan loop — the one gate
    # `file_errors`'s whole design (see ProvenanceGuard's docstring and
    # check_file's OSError/decode handling) was supposed to have already
    # replaced. That direct exit had two consequences, both violations of
    # the priority ladder `_resolve_exit_code` now decides in one place
    # (Minor 9, pass-12 fix wave: this comment used to say "violations
    # outrank classification aborts outrank environment failures" — a
    # three-code ladder that predates `containment_errors` entirely and
    # had gone stale, reading as if it were the complete ladder when a
    # fourth code, exit 3, already outranks all three of these and had for
    # a full fix wave; see `_resolve_exit_code` for the current, complete
    # ladder): a change directory that sorted AFTER the broken entry was never scanned at
    # all (its own violation never reported), and a change directory that
    # sorted BEFORE it — already scanned, already carrying a real
    # violation — still lost to exit 2 instead of exit 1. The fix is to
    # record the message in `file_errors` (declared once, above this
    # loop, precisely so both this gate and the candidate-lstat gate below
    # share the one accumulator `check_file` already uses) and `continue`
    # to the next entry — never adding the unclassifiable entry to
    # `change_dirs` (it cannot be scanned; it can be reported as a
    # failure) but never abandoning every remaining directory either.
    file_errors = []
    change_dirs = []
    for e in entries:
        if e == "archive":
            continue
        entry_path = os.path.join(changes_dir, e)
        try:
            st = os.stat(entry_path)  # follows symlinks, like `find -L`
        except OSError as exc:
            # Minor 6 (pass-11 fix wave): `exc.strerror` (not `exc`
            # itself) — see check_file's identical fix — keeps this
            # message's interpolated OS text errno-only, consistent with
            # the repo-relative prefix this message otherwise uses.
            msg = (
                f"{os.path.relpath(entry_path, repo_root)}: cannot "
                f"determine whether this entry is a change directory "
                f"({exc.strerror or exc}) — refusing to report a clean run"
            )
            print(msg, file=sys.stderr)
            file_errors.append(msg)
            continue
        if stat.S_ISDIR(st.st_mode):
            change_dirs.append(e)

    # Zero non-archived changes is only the genuine "nothing in flight"
    # steady state when nothing ALSO failed classification above — an
    # entry this guard could not classify is not the same state as an
    # empty directory, and must not be reported as one (see the
    # docstring's "false clean run" warning). When `file_errors` is
    # non-empty, `change_dirs` may be empty too (every entry was
    # unclassifiable): the scan loop below then simply does nothing, and
    # the end-of-scan priority block decides the exit code, exactly the
    # same as any other combination of results.
    if not change_dirs and not file_errors:
        print(
            f"check-plan-provenance: no non-archived changes under "
            f"{changes_dir} — nothing in flight"
        )
        sys.exit(0)

    # Each change directory contributes one candidate per entry in
    # `SCANNED_FILENAMES`, and each candidate keeps the identical
    # per-file semantics the single-candidate version established: an
    # absent file is an ordinary skip, an undeterminable one is a
    # scan-integrity failure recorded in `file_errors`, and a containment
    # refusal skips that one candidate while the scan continues. A change
    # mid-planning that has a design.md and no tasks.md is therefore a
    # scanned change, not a broken glob.
    scanned = 0
    for name in change_dirs:
        for filename in SCANNED_FILENAMES:
            candidate = os.path.join(changes_dir, name, filename)
            # Critical 1 fix, second half: `os.path.lexists` has the exact
            # same swallow-every-OSError-and-say-False shape as
            # `os.path.isdir` above — a candidate this guard cannot even
            # LOOK UP because its parent directory is `chmod 000` reports
            # "does not exist", identically to a directory that genuinely
            # has no such file, and the entry is silently skipped rather
            # than refused. `os.lstat` distinguishes them:
            # `FileNotFoundError` (or `NotADirectoryError`, a path
            # component that isn't a directory) means genuinely absent —
            # an ordinary, unremarkable skip; any OTHER `OSError`
            # (permission denied being the realistic case) means "cannot
            # determine", the same scan-integrity failure as above, and
            # must abort loudly instead of being read as "nothing here".
            #
            # OPERATOR-VISIBLE COUNT: this gate is now per CANDIDATE, so
            # one undeterminable change directory (a `chmod 000` parent)
            # emits one "cannot determine whether <filename> exists"
            # message per entry in `SCANNED_FILENAMES` — three where the
            # tasks.md-only scan emitted one. That is the required
            # consequence of per-candidate semantics, not duplication:
            # each message names a different candidate, and collapsing
            # them would mean deciding for the operator that the three
            # failures share one cause, which this gate cannot know. The
            # exit code (2) and the diagnosis are unchanged.
            #
            # Critical 1 (pass-10 fix wave): same defect as the entry-stat
            # gate above — this used to be a direct `sys.exit(2)`, skipping
            # every remaining change directory and outranking a violation
            # already found in one scanned earlier. Now: record into the
            # same `file_errors` accumulator and `continue` to the next
            # candidate, exactly the pattern `check_file`'s own
            # OSError/decode handling already uses.
            try:
                os.lstat(candidate)
            except (FileNotFoundError, NotADirectoryError):
                continue
            except OSError as exc:
                # Minor 6 (pass-11 fix wave): `exc.strerror`, not `exc` —
                # see check_file's identical fix above.
                msg = (
                    f"{os.path.relpath(candidate, repo_root)}: cannot "
                    f"determine whether {filename} exists ({exc.strerror or exc}) "
                    "— refusing to report a clean run"
                )
                print(msg, file=sys.stderr)
                file_errors.append(msg)
                continue

            # Critical 2 (pass-11 fix wave): a containment refusal must not
            # stop the scan — `verify_scanned_file_containment` records the
            # refusal (see ProvenanceGuard's docstring) and returns `False`
            # rather than exiting, so a symlinked/escaping/non-regular
            # candidate in an early directory no longer prevents every
            # LATER directory from being scanned. `check_file` must not be
            # called on a `path` that failed containment — that refusal
            # exists precisely so this guard never opens it.
            if not guard.verify_scanned_file_containment(
                candidate, changes_dir, filename
            ):
                continue
            err = guard.check_file(candidate, repo_root)
            if err is not None:
                file_errors.append(err)
            scanned += 1

    # Change directories exist, but none of them produced ANY of
    # `SCANNED_FILENAMES`: a misconfigured root or a planning artifact at
    # the wrong depth, and must never be reported as a clean run —
    # distinct from "zero non-archived changes" above, which is handled
    # before this point precisely so it cannot be confused with this
    # genuinely broken shape. A change that has only some of the three
    # (a design.md and no tasks.md, say) does not trip this: it produced
    # one of them, so it was scanned.
    #
    # Guarded by `not file_errors` (pass-10 fix wave, Critical 1): when
    # every change directory failed classification above (an
    # all-unreadable-entries fixture, not a broken glob), `file_errors`
    # already explains why nothing was scanned, and the end-of-scan
    # priority block below is where that explanation — and its exit code
    # — belongs, not this generic "wrong depth" message.
    # Guarded by `not guard.containment_errors` too (pass-11 fix wave):
    # when every change directory's candidates failed containment above,
    # that already explains why nothing was scanned, and the end-of-scan
    # priority block below — which checks `containment_errors` first — is
    # where that explanation and its exit code belong, not this generic
    # "wrong depth" message.
    if scanned == 0 and not file_errors and not guard.containment_errors:
        print(
            # "none of ... at all" rather than "no tasks.md, design.md,
            # proposal.md": the bare list reads as "not all three were
            # found", which is an ordinary state (a change mid-planning
            # legitimately has only some of them) and not what this
            # failure means. It fires only when NOT ONE of them turned up
            # anywhere.
            f"none of {', '.join(SCANNED_FILENAMES)} found at all, in any "
            f"of {len(change_dirs)} non-archived change(s) under "
            f"{changes_dir} — refusing to report a clean run",
            file=sys.stderr,
        )
        sys.exit(2)

    # Priority, decided once — in `_resolve_exit_code`, not by whichever
    # failure mode happens to be hit first mid-scan (Important 8
    # established this for OSError/decode failures; Critical 2 of the
    # pass-8 fix wave extends it to classify_line's content-classification
    # abort; Important 3 of THIS pass collapses what used to be four
    # separate hand-written priority chains into the one ladder
    # `_resolve_exit_code` walks): a containment/security refusal (exit 3)
    # outranks EVERYTHING — real violations (exit 1), classification
    # aborts (exit 4), an unreadable/undecodable file (exit 2), and
    # nothing being wrong at all (exit 0) — because a symlink escape or a
    # directory-traversal attempt must never be downgraded just because
    # something else, anywhere else, was also found (Critical 2, pass-11
    # fix wave; see ProvenanceGuard's and verify_scanned_file_containment's
    # docstrings for the two-slot disagreement this adjudicates). Below
    # exit 3: real violations (exit 1) outrank classification aborts
    # (exit 4), which outrank an unreadable/undecodable file (exit 2),
    # which outrank nothing being wrong at all (exit 0). A file this
    # guard could not classify — or could not even open due to
    # containment — never masks a violation or an abort found elsewhere;
    # every one of them is still printed below regardless of which exit
    # code wins — `_resolve_exit_code` decides only the WINNING code, not
    # which channels get reported.
    exit_code = _resolve_exit_code(
        guard.containment_errors, guard.failures, guard.abort_messages, file_errors
    )

    if exit_code == 3:
        _report_containment_refusal(guard, file_errors)

    if exit_code == 1:
        print(f"\n{guard.failures} plan-provenance violation(s) found.", file=sys.stderr)
        if guard.abort_messages:
            print(
                f"{len(guard.abort_messages)} additional line(s) could not be "
                "classified (content-classification, see above); fix those "
                "independently of the violations reported here.",
                file=sys.stderr,
            )
        if file_errors:
            # Important 8: a file this guard could not classify (bad
            # encoding, unreadable, oversized) never masks a violation
            # found elsewhere — report both, exit 1 (violations win),
            # rather than one bad byte in one file silencing everything
            # a clean file already reported.
            print(
                f"{len(file_errors)} additional file(s) could not "
                "be read (environment failure, see above); fix those "
                "independently of the violations reported here.",
                file=sys.stderr,
            )
        print(
            # Minor 8 (pass-9 fix wave): "For the violations above" scopes
            # this remedy to `self.failures` specifically. Before this fix
            # the banner had no scoping phrase at all, so when an abort or
            # a file error was ALSO reported just above (both printed
            # earlier in this same block), an imperative with no subject
            # could read as covering those too — it does not: adding a
            # verified:/unverified: tag or a measured:/predicted: comment
            # fixes a stated-provenance violation; it does not resolve an
            # unclassifiable line or an unreadable file, which have their
            # own remedies (see the abort/file-error messages themselves).
            #
            # Minor (pass-17 fix wave): the banner used to offer tagging as
            # the ONLY remedy. For a claim reported because a veto withdrew
            # its quotation exemption, the contract's actual remedy is to
            # REWORD the line — an author following the old banner literally
            # would have written a measured:/predicted: tag for a number
            # that was already correctly quoted, which is the tag vocabulary
            # being used untruthfully to silence a guard. The per-claim note
            # (`_veto_note`) names the specific cause and fix; this scopes
            # the banner so it does not contradict it.
            "For the violations above: fix by adding the missing "
            "verified:/unverified: tag or measured:/predicted: comment — or, "
            "where a note above says the quotation exemption was withdrawn, "
            "by rewording the line; never suppress.",
            file=sys.stderr,
        )
        sys.exit(1)

    if exit_code == 4:
        # No provenance violations were found, but at least one line
        # could not be classified (see above) — never a clean run.
        # Critical 2, pass-8 fix wave: this is decided here, once, after
        # every file and every change directory has had a chance to be
        # scanned, rather than by an immediate `sys.exit(4)` from inside
        # `check_file`'s per-line loop (which used to abandon every
        # directory sorted after the one that aborted).
        print(
            f"\n{len(guard.abort_messages)} line(s) could not be classified "
            "(content-classification, see above) — refusing to report a "
            "clean run",
            file=sys.stderr,
        )
        if file_errors:
            print(
                f"{len(file_errors)} file(s) also could not be "
                "read (environment failure, see above).",
                file=sys.stderr,
            )
        sys.exit(4)

    if exit_code == 2:
        # No provenance violations or classification aborts were found,
        # but at least one scanned file could not be read/decoded — that is
        # still not a clean run; report it and exit 2 (environment),
        # never 0.
        #
        # Important 3 (pass-11 fix wave): this banner used to say "could
        # not be classified" — the exact phrase exit 4's banner also uses
        # for a content-classification abort. The two failures are not
        # the same thing and do not have the same remedy: this one means
        # a permission, encoding or size problem ON DISK ("fix a
        # permission, encoding or size problem"), not a markdown edit
        # ("edit the markdown" is exit 4's remedy). Reusing "classified"
        # here told an operator to go re-read prose when the actual fix
        # was a filesystem permission or an encoding issue.
        print(
            f"\n{len(file_errors)} file(s) could not be read "
            "(environment failure — a permission, encoding or size problem "
            "on disk, see above) — refusing to report a clean run",
            file=sys.stderr,
        )
        sys.exit(2)

    print(f"check-plan-provenance: {scanned} file(s) scanned, all provenance stated")


if __name__ == "__main__":
    main()
