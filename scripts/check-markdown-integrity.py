#!/usr/bin/env python3
r"""check-markdown-integrity.py — fail when this repository's own skill and
rule Markdown carries structural damage a rendering-agnostic reader would
mis-read: a code span broken by a misunderstood backslash escape, a
blockquote whose first line lost its `>` marker, a paragraph torn apart
mid-sentence, or a paragraph that promises content with a trailing colon
and then supplies none.

Exit codes:
  0  clean — every scanned file is structurally sound.
  1  violations found — one or more of the four signals below fired.
     Each is reported as `file:line: <signal>: <message>`. Fix the named
     line; there is no suppression marker and no mechanism for one.
  2  cannot answer — the project root does not exist, an argument this
     guard does not accept was given, or a file inside the scanned scope
     exists but cannot be read (permission denied, not valid UTF-8).

Scope: `skills/**/*.md` and `rules/*.mdc` under the given project root
(default: the repository this script lives in) — the files every
`/myflow-*` run loads, and where torn text changes an agent's behaviour
rather than merely reading badly. Nothing else is scanned: not `docs/`,
not `openspec/`, not the repository root's own Markdown.

=============================================================================
THE BLOCK MODEL
=============================================================================

A file is walked line by line into a flat sequence of typed blocks —
`frontmatter`, `fence`, `thematic_break`, `table`, `heading`, `list_item`,
`blockquote`, `paragraph` — following the shape `check-plan-provenance.py`'s
own classifier uses: a block ends at a blank line, at a line that opens a
new block outright (a fence, a heading, a thematic break, a new list
marker, a table header), or at EOF; content inside a fence is never
reclassified — it is consumed whole, from the opening delimiter run to a
closing run of the same character at least as long, exactly as CommonMark's
own fenced code block rule requires.

**Continuation is the load-bearing rule this guard got wrong on its first
calibration pass against this repository's own tree, and is documented
here because of it.** A `paragraph` or `list_item` block does not end just
because a later physical line, taken alone, would classify differently in
isolation — an ordinary wrapped sentence ("...before it altered, elided,
or reflowed; and") or a second line of a bulleted item ("  unaffected")
looks, read on its own, exactly like a paragraph line with no marker of
its own. CommonMark itself treats these as lazy continuation lines of the
block already open, and the first version of this guard's block walker did
not: it re-classified every physical line independently, which reads a
routine two-line sentence as two separate one-line blocks and flags the
first one as torn. The fix is `_open_or_continue` below: once a
`paragraph` or `list_item` block is open, an ordinary line continues it
unless the line itself opens something structurally distinct (a fence, a
heading, a thematic break, a blockquote, a **new** list marker, or a table
header confirmed by lookahead at the separator row beneath it) — the same
distinction CommonMark's own "lazy continuation" rule draws.

This is deliberately a coarser grammar than `check-plan-provenance.py`'s
container-prefix stripper: that guard must resolve a fence nested behind
blockquote and list markers because plan prose commonly nests that way;
this guard's scope is skill and rule prose, where that nesting is rare, so
a fence is recognised only at its own line start (up to three leading
spaces, matching CommonMark) and a line already inside one is never
reconsidered by any other rule.

**YAML frontmatter** (a `---` delimiter as the file's first line, closed
by the next line that is itself a bare `---`/`***`/`___` rule) is consumed
as its own `frontmatter` block before the general walk begins, and is
never scanned for any signal: it is not prose, and its own closing `---`
line is not a torn paragraph. **A thematic break** (a bare line of three
or more `-`, `*` or `_` characters, elsewhere in the file — a horizontal
rule) is its own single-line `thematic_break` block for the same reason:
it is punctuation-free by construction and is not a paragraph that failed
to end.

Signals 1 and 3 (broken code span, unterminated paragraph) scan only
`paragraph`-kind blocks — a fence, a table cell, a heading and a list item
are structurally excluded, not merely skipped by a heuristic, which is
what lets a code fixture inside a fenced example read however it needs to
without tripping the guard that reads this file's own prose. Signal 1
scans a whole paragraph block's text at once, not line by line: a code
span may itself wrap across a soft line break inside one paragraph (`` `a
count line and one `finding-reproducer: F<n>` `` split across two physical
lines is one valid span, not two lines each carrying half a broken one),
and CommonMark's own delimiter-matching rule does not stop at the newline
either. Signal 2 (orphaned blockquote body) reads the transition between a
`blockquote` block and the block immediately before it. Signal 4 (dangling
promise) reads the transition between a `paragraph` block and whatever
block follows it, or its absence.

=============================================================================
SIGNAL 1 — BROKEN CODE SPAN
=============================================================================

A CommonMark code span opens at a run of N backticks and closes at the
next run of exactly N backticks; backslash is not an escape inside a code
span, so `` `foo\`bar` `` does not read as one span containing a literal
backtick — the backslash is inert, the second run (length 1) closes the
span early, and the run after `bar` is left with no partner at all. This
guard finds that condition without hand-coding the escape shape directly:
it collects every maximal run of backtick characters in a paragraph
block's text, and greedily pairs each run with the nearest following run
of the same length (CommonMark's own nearest-match rule); a run that never
gets a partner leaves the guard unable to see where its span was meant to
close, which is signal 1 in both of its native shapes at once — a stray
backtick with no partner anywhere in the block, and a backslash-escape
attempt that steals an early close and orphans the real one. A backtick
run that is not length 1 is exempt from the escape story but still needs a
partner of its own length: `` ``get `foo` done`` `` pairs the two
double-backtick runs with each other and the two single backticks in
between with each other, so the literal backtick inside is read correctly
and nothing fires.

=============================================================================
SIGNAL 4 IS STRUCTURAL, NEVER PHRASE-BASED
=============================================================================

Signal 4 fires only when a paragraph's last line ends in a colon AND the
next block, if any, is a heading (or there is no next block at all — end
of file). It does not look at what the sentence says. This repository
ends hundreds of sections with a citation of the shape `See **X**
(`f.md`) for why ...`, correctly, and those end in a full stop, not a
colon — the structural rule reads them as ordinary complete paragraphs
and never fires. A phrase-based rule keying on "see ... for why" or "the
reason is" would instead have to fire on all of them, and the lint policy
this repository runs under forbids adding a suppression marker to quiet
a rule that over-fires by design; the fix for an over-firing rule is
always to narrow the rule; there is no other lever. Signal 4 is
structural for that reason, deliberately, and stays that way.
"""

from __future__ import annotations

import os
import re
import sys
from dataclasses import dataclass, field
from typing import List, Optional

# TERMINAL_CHARS — the set of characters that legitimately end a paragraph
# line (signal 3) or the prose line immediately preceding a blockquote
# (signal 2). A colon is included here on purpose: a paragraph ending in a
# colon is not itself unterminated (signal 3 does not fire on it), and
# prose ending in a colon immediately before a blockquote is the ordinary
# way to introduce a quotation, not an orphaned marker (signal 2 does not
# fire on it either). Signal 4 is the one place a trailing colon matters,
# and it asks a different question — not "is this terminated" but "is
# nothing following what the colon promised".
TERMINAL_CHARS = frozenset('.:?!)|"\'`')

# EMPHASIS_MARKERS — Markdown's own inline-emphasis closers (`**bold**`,
# `*italic*`, `__bold__`, `_italic_`). A sentence very commonly ends
# "...done.**" — the terminal punctuation sits *inside* the emphasis span,
# followed by its own closing marker. Stripping a trailing run of these
# characters before testing for terminal punctuation is what keeps that
# ordinary, correct shape from reading as unterminated; it is not an
# attempt to parse emphasis spans in general; a paragraph that ends with a
# bare, unmatched `*` and nothing else is stripped down to nothing here
# and reported (an empty result is treated as "no terminal punctuation
# found", not as an exemption).
EMPHASIS_MARKERS = frozenset("*_")

FENCE_OPEN_RE = re.compile(r'^ {0,3}(`{3,}|~{3,})(.*)$')
FENCE_CLOSE_RE = re.compile(r'^ {0,3}(`+|~+)[ \t]*$')
THEMATIC_BREAK_RE = re.compile(r'^ {0,3}(-{3,}|\*{3,}|_{3,})[ \t]*$')
HEADING_RE = re.compile(r'^ {0,3}#{1,6}(\s|$)')
LIST_ITEM_RE = re.compile(r'^\s*([-*+]|\d+[.)])\s+')
BLOCKQUOTE_RE = re.compile(r'^\s*>(\s|$)')
TABLE_SEPARATOR_RE = re.compile(r'^ {0,3}\|?\s*:?-{1,}:?\s*(\|\s*:?-{1,}:?\s*)*\|?\s*$')
INDENTED_CODE_RE = re.compile(r'^ {4,}\S')

# Kinds whose presence immediately after a paragraph means that paragraph
# is *usually* not torn — it is the CommonMark-legal "introduce, then
# supply" pattern this repository uses throughout: a clause ending in a
# bare noun or verb ("...run the following command") directly above a
# fenced command, or a bold pseudo-heading ("**Explore the problem
# space**") directly above the list it introduces. Calibrating this guard
# against this repository's own tree found dozens of exactly these two
# shapes, all correct, all repeated deliberately — see
# find_unterminated_paragraphs. This exemption alone is not discriminating
# enough, though: it fires on *any* paragraph in front of a fence, list
# item or table, whether or not that paragraph actually introduces it, so
# a genuinely torn sentence that merely happens to sit before an unrelated
# fence, list or table would pass uncaught. DANGLING_LAST_WORDS below is
# the second half of the test that closes that gap.
COMPLETING_KINDS = frozenset(("fence", "list_item", "table"))

# DANGLING_LAST_WORDS — the closed word classes that cannot end a
# grammatical clause on their own: articles, coordinating and common
# subordinating conjunctions, prepositions, relative pronouns, and
# auxiliary/linking verbs. An "introduce, then supply" paragraph reads as
# a complete clause even without terminal punctuation ("run the following
# command", "explore the problem space") — its last word is an ordinary
# noun or verb. A paragraph actually torn mid-clause instead trails off on
# one of these closed-class words ("...but then just stops without",
# "...because", "...and"), which is a structural fact about the word
# itself, not a judgement about the surrounding phrase. This is the
# discriminator COMPLETING_KINDS is paired against: a paragraph before a
# completing block is exempt only when its last word is *not* in this set;
# when it is, the paragraph is reported as unterminated regardless of what
# follows it. The set is deliberately closed-class (a finite, enumerable
# part of speech) rather than open-class (content words), which is what
# keeps the test structural instead of a phrase-matching heuristic.
DANGLING_LAST_WORDS = frozenset((
    # articles
    "a", "an", "the",
    # coordinating conjunctions
    "and", "but", "or", "nor", "so", "yet",
    # subordinating conjunctions
    "because", "although", "though", "while", "since", "if", "unless",
    "until", "when", "whereas", "whether", "as",
    # prepositions
    "of", "to", "in", "on", "at", "by", "with", "without", "from",
    "into", "onto", "upon", "over", "under", "above", "below",
    "between", "among", "through", "during", "before", "after",
    "about", "against", "toward", "towards", "within", "across",
    "along", "around", "behind", "beside", "beyond", "despite",
    "except", "inside", "near", "outside", "per", "throughout",
    "till", "unto", "via", "for",
    # relative pronouns / subordinators
    "which", "that", "who", "whom", "whose",
    # auxiliary / linking verbs
    "is", "are", "was", "were", "be", "been", "being", "am",
    "do", "does", "did", "have", "has", "had",
    "will", "would", "shall", "should", "can", "could", "may",
    "might", "must",
))

WORD_RE = re.compile(r"[A-Za-z']+")

# Kinds a `paragraph` or `list_item` block keeps absorbing lazy
# continuation lines into, per CommonMark's own lazy-continuation rule —
# see the module docstring's "Continuation is the load-bearing rule..."
# section for why this exists.
CONTINUABLE_KINDS = frozenset(("paragraph", "list_item"))


@dataclass
class Block:
    kind: str
    start_line: int
    end_line: int
    lines: List[str] = field(default_factory=list)


@dataclass
class Finding:
    line: int
    signal: str
    message: str


def _opens_new_block(line: str) -> Optional[str]:
    """Return the block kind `line` unconditionally opens (a fence, a
    heading, a thematic break, a blockquote, or a new list marker), or
    None if `line` is not, by itself, the start of one of those. Table
    detection is handled separately by the caller, since it needs to look
    at the FOLLOWING line (the separator row) to be sure."""
    if FENCE_OPEN_RE.match(line):
        return "fence"
    if THEMATIC_BREAK_RE.match(line):
        return "thematic_break"
    if HEADING_RE.match(line):
        return "heading"
    if BLOCKQUOTE_RE.match(line):
        return "blockquote"
    if LIST_ITEM_RE.match(line):
        return "list_item"
    return None


def _consume_frontmatter(lines: List[str]) -> Optional[Block]:
    """If `lines` opens with a `---` delimiter, return the `frontmatter`
    block it and its matching closing delimiter form, else None. Only
    checked once, at the very start of the file — a `---` appearing later
    is an ordinary thematic break, not frontmatter."""
    if not lines or lines[0].strip() != "---":
        return None
    for idx in range(1, len(lines)):
        if lines[idx].strip() in ("---", "***", "___"):
            return Block(
                kind="frontmatter",
                start_line=1,
                end_line=idx + 1,
                lines=lines[0:idx + 1],
            )
    return None


def parse_blocks(text: str) -> List[Block]:
    """Walk `text` into a flat sequence of typed blocks. Blank lines end
    the current block without becoming blocks of their own. A `paragraph`
    or `list_item` block absorbs ordinary continuation lines (see the
    module docstring); every other kind is single-line or, for `fence`,
    consumes up to its own closing delimiter."""
    lines = text.splitlines()
    blocks: List[Block] = []

    start_index = 0
    frontmatter = _consume_frontmatter(lines)
    if frontmatter is not None:
        blocks.append(frontmatter)
        start_index = frontmatter.end_line  # 0-based index of next line

    current: Optional[Block] = None
    in_fence = False
    fence_char = ""
    fence_len = 0

    def flush() -> None:
        nonlocal current
        if current is not None:
            blocks.append(current)
            current = None

    i = start_index
    n = len(lines)
    while i < n:
        lineno = i + 1
        raw = lines[i]

        if in_fence:
            # `in_fence` is only ever set True in the branch below, which in
            # the same breath assigns `current` to the fence `Block` it just
            # opened — the two are set together and `current` is never
            # cleared to None while `in_fence` stays True. The assert makes
            # that invariant explicit and fails fast if a future edit ever
            # breaks it, instead of the reader having to trust it silently;
            # it also narrows `current`'s type for the type checker, so the
            # two calls below need no suppression to satisfy it.
            assert current is not None, "in_fence implies a fence Block is open"
            current.lines.append(raw)
            current.end_line = lineno
            m = FENCE_CLOSE_RE.match(raw)
            if m and m.group(1)[0] == fence_char and len(m.group(1)) >= fence_len:
                in_fence = False
                flush()
            i += 1
            continue

        if raw.strip() == "":
            flush()
            i += 1
            continue

        opened = _opens_new_block(raw)

        if opened == "fence":
            flush()
            m = FENCE_OPEN_RE.match(raw)
            assert m is not None
            run = m.group(1)
            fence_char = run[0]
            fence_len = len(run)
            in_fence = True
            current = Block(kind="fence", start_line=lineno, end_line=lineno, lines=[raw])
            i += 1
            continue

        if opened in ("thematic_break", "heading"):
            flush()
            blocks.append(Block(kind=opened, start_line=lineno, end_line=lineno, lines=[raw]))
            i += 1
            continue

        if opened == "blockquote":
            if current is not None and current.kind == "blockquote":
                current.lines.append(raw)
                current.end_line = lineno
            else:
                flush()
                current = Block(kind="blockquote", start_line=lineno, end_line=lineno, lines=[raw])
            i += 1
            continue

        if opened == "list_item":
            # A fresh marker always starts a new list_item block, even if
            # one was already open — this guard does not track nesting
            # depth or item boundaries beyond "some list item is open".
            flush()
            current = Block(kind="list_item", start_line=lineno, end_line=lineno, lines=[raw])
            i += 1
            continue

        # Not an unconditional opener. If a paragraph/list_item is
        # already open, this is an ordinary lazy-continuation line.
        if current is not None and current.kind in CONTINUABLE_KINDS:
            current.lines.append(raw)
            current.end_line = lineno
            i += 1
            continue

        # If a table is already open, a line still carrying a `|` extends
        # it; one that doesn't ends it.
        if current is not None and current.kind == "table":
            if "|" in raw:
                current.lines.append(raw)
                current.end_line = lineno
                i += 1
                continue
            flush()

        # Fresh block. A `|`-bearing line only starts a table when the
        # very next line is a genuine GFM separator row — this lookahead
        # is what keeps an ordinary sentence with one stray `|` in it
        # (a shell pipe mentioned in prose, a record's own `|` field) from
        # being misread as a table header.
        if "|" in raw and i + 1 < n and TABLE_SEPARATOR_RE.match(lines[i + 1]):
            flush()
            current = Block(kind="table", start_line=lineno, end_line=lineno, lines=[raw])
            i += 1
            continue

        # A fresh, blank-line-preceded (or file-start) line indented four
        # or more spaces is a CommonMark indented code block, not prose —
        # reused as kind "fence" so it gets the same blanket exclusion
        # from every prose signal a fenced block already gets. This is
        # only ever checked here, at a fresh block boundary: an indented
        # CONTINUATION line of an already-open paragraph or list item is
        # handled above and never reaches this branch.
        if INDENTED_CODE_RE.match(raw):
            start = lineno
            block_lines: List[str] = []
            while i < n:
                candidate = lines[i]
                if candidate.strip() == "":
                    j = i + 1
                    while j < n and lines[j].strip() == "":
                        j += 1
                    if j < n and INDENTED_CODE_RE.match(lines[j]):
                        block_lines.append(candidate)
                        i += 1
                        continue
                    break
                if INDENTED_CODE_RE.match(candidate):
                    block_lines.append(candidate)
                    i += 1
                    continue
                break
            blocks.append(Block(
                kind="fence", start_line=start,
                end_line=start + len(block_lines) - 1, lines=block_lines,
            ))
            continue

        flush()
        current = Block(kind="paragraph", start_line=lineno, end_line=lineno, lines=[raw])
        i += 1

    flush()
    return blocks


def _backtick_runs_with_offsets(text: str):
    runs = []
    i = 0
    n = len(text)
    while i < n:
        if text[i] == "`":
            j = i
            while j < n and text[j] == "`":
                j += 1
            runs.append((i, j - i))
            i = j
        else:
            i += 1
    return runs


def _first_unmatched_backtick_offset(text: str) -> Optional[int]:
    """The character offset of the first backtick run in `text` that never
    finds a same-length partner, greedily, in CommonMark's own
    nearest-match order — or None if every run pairs off."""
    runs = _backtick_runs_with_offsets(text)
    i = 0
    n = len(runs)
    while i < n:
        start, length = runs[i]
        partner = None
        for j in range(i + 1, n):
            if runs[j][1] == length:
                partner = j
                break
        if partner is None:
            return start
        i = partner + 1
    return None


def _line_of_offset(block: Block, offset: int) -> int:
    """Map a character offset into `block`'s '\\n'-joined text back to the
    physical line number it falls on."""
    pos = 0
    for idx, line in enumerate(block.lines):
        seg_len = len(line) + 1  # + 1 for the joining '\n'
        if offset < pos + seg_len:
            return block.start_line + idx
        pos += seg_len
    return block.end_line


def _strip_trailing_emphasis(text: str) -> str:
    end = len(text)
    while end > 0 and text[end - 1] in EMPHASIS_MARKERS:
        end -= 1
    return text[:end]


def _ends_sentence(line: str) -> bool:
    stripped = _strip_trailing_emphasis(line.rstrip())
    if not stripped:
        return False
    return stripped[-1] in TERMINAL_CHARS


def _ends_on_dangling_word(line: str) -> bool:
    """True when `line`'s last word (after stripping trailing emphasis
    markers) is one of DANGLING_LAST_WORDS — a closed-class word that
    cannot end a grammatical clause by itself. Used to tell a paragraph
    that genuinely tears off mid-clause apart from one that merely
    introduces the block after it; see DANGLING_LAST_WORDS above."""
    stripped = _strip_trailing_emphasis(line.rstrip())
    words = WORD_RE.findall(stripped)
    if not words:
        return False
    return words[-1].lower() in DANGLING_LAST_WORDS


def _block_resumes_sentence(block: Optional["Block"]) -> bool:
    """True when `block` picks the sentence back up rather than starting
    a new one — the shape a paragraph takes when it deliberately spans an
    embedded fence, list, or table ("...so a tag can run past the first
    `>` — and" / a code example / "exited 0 with the bare claim
    unreported..."). This is the second half of the dangling-word test:
    a paragraph ending on a dangling word is only torn when nothing after
    the completing block resumes it.

    `None` (the completing block was the last block in the file) never
    resumes anything — there is nothing left to supply the rest of the
    sentence. A `heading` or `list_item` never resumes one either: both
    are marked structure (a `#` line, a `-`/`1.` marker) that always
    opens something new, never bare continuation prose.

    Otherwise the check is the case of the block's own first letter.
    CommonMark prose convention capitalizes the first word of a new
    sentence; a block resuming a sentence torn off mid-clause instead
    opens on whatever the sentence's next word already was — lower case,
    a digit, or inline markup such as a backtick or quote mark, none of
    which is an upper-case letter. So "not upper-case" is the structural
    stand-in for "resumes," and "upper-case" for "starts fresh" — no word
    list is involved, only the character's own case. A leading `>`
    blockquote marker is stripped first so the check looks at the quoted
    text itself rather than the marker punctuation."""
    if block is None or block.kind in ("heading", "list_item"):
        return False
    text = block.lines[0]
    quote_marker = BLOCKQUOTE_RE.match(text)
    if quote_marker:
        text = text[quote_marker.end():]
    stripped = text.lstrip()
    if not stripped:
        return False
    return not stripped[0].isupper()


def find_broken_code_spans(blocks: List[Block]) -> List[Finding]:
    findings = []
    for block in blocks:
        if block.kind != "paragraph":
            continue
        text = "\n".join(block.lines)
        offset = _first_unmatched_backtick_offset(text)
        if offset is not None:
            findings.append(Finding(
                _line_of_offset(block, offset), "broken code span",
                "a backtick run in this paragraph has no same-length "
                "partner, so a code span closes early or never closes at all",
            ))
    return findings


def find_orphaned_blockquotes(blocks: List[Block]) -> List[Finding]:
    findings = []
    for idx, block in enumerate(blocks):
        if block.kind != "blockquote":
            continue
        if idx == 0:
            continue
        prev = blocks[idx - 1]
        if prev.kind in ("blockquote", "heading") or prev.kind in COMPLETING_KINDS:
            # A heading is marked, not un-marked prose, and heading text
            # routinely carries no terminal punctuation of its own ("## Next
            # heading") — that is normal heading style, not a sentence torn
            # from the blockquote that follows it. Signal 2 exists to catch
            # prose that lost its own `>` marker; a heading was never
            # eligible to carry one, so it is excluded outright rather than
            # tested against TERMINAL_CHARS at all.
            #
            # A fence, list_item, or table is marked structure too, for the
            # same reason COMPLETING_KINDS already excludes them from signal
            # 3's "torn paragraph" test: a closing ```/~~~ delimiter, a tight
            # list item's own text, or a table's separator row is not prose
            # that could have lost a `>` marker — it is a different kind of
            # block entirely, one whose last line routinely carries no
            # terminal punctuation as a matter of its own syntax. Reusing
            # COMPLETING_KINDS here rather than a second hand-written list
            # keeps the two exemptions — this one and signal 3's — from
            # drifting apart the way the digit-bound pattern once did.
            continue
        if prev.end_line + 1 != block.start_line:
            continue
        if _ends_sentence(prev.lines[-1]):
            continue
        findings.append(Finding(
            prev.end_line, "orphaned blockquote body",
            "un-marked prose sits immediately before a blockquote and does "
            "not end a sentence — this line may have lost its own `>` marker",
        ))
    return findings


def find_unterminated_paragraphs(blocks: List[Block]) -> List[Finding]:
    findings = []
    for idx, block in enumerate(blocks):
        if block.kind != "paragraph":
            continue
        if _ends_sentence(block.lines[-1]):
            continue
        # A paragraph directly introducing a fence, a list, or a table is
        # not torn — the block that follows supplies what the paragraph
        # led into (a command, a set of steps, a set of rows), the same
        # "introduce, then supply" shape signal 4 already exempts for a
        # trailing colon. See COMPLETING_KINDS above for the calibration
        # evidence this exemption is drawn from. That alone is too coarse,
        # though: it would exempt a paragraph that merely happens to sit in
        # front of a fence/list/table even when the two are unrelated and
        # the paragraph really is torn. DANGLING_LAST_WORDS narrows it —
        # the exemption applies only when the paragraph's own last word
        # could plausibly end a clause; a paragraph trailing off on a
        # closed-class word (a preposition, conjunction, article or
        # auxiliary) is reported regardless of what follows it — *unless*
        # the block past the fence/list/table picks the sentence back up,
        # which is the "spans an embedded block" shape (a sentence that
        # deliberately breaks around a code example and resumes after it,
        # e.g. "...so a tag can run past the first `>` — and" / an
        # example / "exited 0 with the bare claim unreported..."). A
        # dangling last word is torn prose only when nothing resumes it;
        # see _block_resumes_sentence for the structural test.
        nxt = blocks[idx + 1] if idx + 1 < len(blocks) else None
        introduces_next = nxt is not None and nxt.kind in COMPLETING_KINDS
        if introduces_next:
            if not _ends_on_dangling_word(block.lines[-1]):
                continue
            nxt2 = blocks[idx + 2] if idx + 2 < len(blocks) else None
            if _block_resumes_sentence(nxt2):
                continue
        findings.append(Finding(
            block.end_line, "unterminated paragraph",
            "this paragraph's last line ends with no terminal punctuation",
        ))
    return findings


def find_dangling_promises(blocks: List[Block]) -> List[Finding]:
    findings = []
    for idx, block in enumerate(blocks):
        if block.kind != "paragraph":
            continue
        last = _strip_trailing_emphasis(block.lines[-1].rstrip())
        if not last.endswith(":"):
            continue
        nxt = blocks[idx + 1] if idx + 1 < len(blocks) else None
        if nxt is None or nxt.kind == "heading":
            findings.append(Finding(
                block.end_line, "dangling promise",
                "a colon promises content, and the block ends with nothing "
                "following it",
            ))
    return findings


def scan_file(path: str) -> List[Finding]:
    with open(path, "r", encoding="utf-8") as fh:
        text = fh.read()
    blocks = parse_blocks(text)
    findings: List[Finding] = []
    findings.extend(find_broken_code_spans(blocks))
    findings.extend(find_orphaned_blockquotes(blocks))
    findings.extend(find_unterminated_paragraphs(blocks))
    findings.extend(find_dangling_promises(blocks))
    findings.sort(key=lambda f: f.line)
    return findings


def collect_scanned_files(root: str) -> List[str]:
    files: List[str] = []
    skills_dir = os.path.join(root, "skills")
    if os.path.isdir(skills_dir):
        for dirpath, _dirnames, filenames in os.walk(skills_dir):
            for name in filenames:
                if name.endswith(".md"):
                    files.append(os.path.join(dirpath, name))
    rules_dir = os.path.join(root, "rules")
    if os.path.isdir(rules_dir):
        for name in os.listdir(rules_dir):
            if name.endswith(".mdc"):
                candidate = os.path.join(rules_dir, name)
                if os.path.isfile(candidate):
                    files.append(candidate)
    files.sort()
    return files


def resolve_default_root() -> str:
    script_dir = os.path.dirname(os.path.abspath(__file__))
    return os.path.dirname(script_dir)


def main(argv: List[str]) -> int:
    if len(argv) > 1:
        sys.stderr.write(
            "check-markdown-integrity: unknown argument — this guard takes "
            "at most one, an optional project root\n"
        )
        return 2

    if len(argv) == 1:
        arg = argv[0]
        if arg.startswith("-"):
            sys.stderr.write(
                f"check-markdown-integrity: unknown argument '{arg}' — "
                "this guard takes at most one, an optional project root\n"
            )
            return 2
        root = arg
    else:
        root = resolve_default_root()

    if not os.path.isdir(root):
        sys.stderr.write(
            f"check-markdown-integrity: cannot answer — scope root does not "
            f"exist: {root}\n"
        )
        return 2

    files = collect_scanned_files(root)

    all_findings = []
    for path in files:
        try:
            findings = scan_file(path)
        except (OSError, UnicodeDecodeError) as exc:
            sys.stderr.write(
                f"check-markdown-integrity: cannot read {path}: {exc}\n"
            )
            return 2
        for finding in findings:
            all_findings.append((path, finding))

    if not all_findings:
        return 0

    all_findings.sort(key=lambda pair: (pair[0], pair[1].line))
    for path, finding in all_findings:
        print(f"{path}:{finding.line}: {finding.signal}: {finding.message}")
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
