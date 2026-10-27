package guard

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"
)

// checkMarkdownIntegrity is scripts/check-markdown-integrity.sh's guard. The
// shim's header states the contract — the seven signals, the scope and the
// exit codes; this comment is the mechanism.
//
// # The block model
//
// A file is walked line by line into a flat sequence of typed blocks —
// frontmatter, fence, thematic_break, html_comment, table, heading,
// list_item, blockquote, paragraph — following the shape planprovenance.go's
// own classifier uses: a block ends at a blank line, at a line that opens a
// new block outright (a fence, a heading, a thematic break, a new list
// marker, a table header), or at EOF; content inside a fence is never
// reclassified — it is consumed whole, from the opening delimiter run to a
// closing run of the same character at least as long, exactly as
// CommonMark's own fenced code block rule requires.
//
// Continuation is the load-bearing rule this guard got wrong on its first
// calibration pass against this repository's own tree, and is documented here
// because of it. A paragraph or list_item block does not end just because a
// later physical line, taken alone, would classify differently in isolation —
// an ordinary wrapped sentence ("...before it altered, elided, or reflowed;
// and") or a second line of a bulleted item ("  unaffected") looks, read on
// its own, exactly like a paragraph line with no marker of its own. CommonMark
// itself treats these as lazy continuation lines of the block already open,
// and the first version of this guard's block walker did not: it
// re-classified every physical line independently, which reads a routine
// two-line sentence as two separate one-line blocks and flags the first one
// as torn. The fix is the continuation branch in mdiParseBlocks: once a
// paragraph or list_item block is open, an ordinary line continues it unless
// the line itself opens something structurally distinct (a fence, a heading,
// a thematic break, a blockquote, a new list marker, or a table header
// confirmed by lookahead at the separator row beneath it) — the same
// distinction CommonMark's own "lazy continuation" rule draws.
//
// This is deliberately a coarser grammar than planprovenance.go's
// container-prefix stripper: that guard must resolve a fence nested behind
// blockquote and list markers because plan prose commonly nests that way;
// this guard's scope is skill and rule prose, where that nesting is rare, so
// a fence is recognised only at its own line start (up to three leading
// spaces, matching CommonMark) and a line already inside one is never
// reconsidered by any other rule.
//
// YAML frontmatter (a `---` delimiter as the file's first line, closed by the
// next line that is itself a bare `---`/`***`/`___` rule) is consumed as its
// own frontmatter block before the general walk begins, and is never scanned
// for any signal: it is not prose, and its own closing `---` line is not a
// torn paragraph. A thematic break (a bare line of three or more `-`, `*` or
// `_` characters, elsewhere in the file — a horizontal rule) is its own
// single-line thematic_break block for the same reason: it is
// punctuation-free by construction and is not a paragraph that failed to end.
// A standalone HTML comment — a whole line of `<!-- … -->`, such as the
// `<!-- core -->` markers setup.sh renders against or the `<!-- rule: … -->`
// labels it writes — is its own html_comment block on the same grounds: it
// renders to nothing, it cannot carry terminal punctuation, and a paragraph
// does not continue through it. Only a comment that opens and closes on one
// line qualifies; a multi-line comment is left to the ordinary walk.
//
// Signals 1 and 3 (broken code span, unterminated paragraph) scan only
// paragraph-kind blocks — a fence, a table cell, a heading and a list item
// are structurally excluded, not merely skipped by a heuristic, which is what
// lets a code fixture inside a fenced example read however it needs to
// without tripping the guard that reads this file's own prose. Signal 1 scans
// a whole paragraph block's text at once, not line by line: a code span may
// itself wrap across a soft line break inside one paragraph, and CommonMark's
// own delimiter-matching rule does not stop at the newline either. Signal 2
// (orphaned blockquote body) reads the transition between a blockquote block
// and the block immediately before it. Signal 4 (dangling promise) reads the
// transition between a paragraph block and whatever block follows it, or its
// absence.
// Signal 5 (missing final newline) reads the raw text: a non-empty file
// whose last byte is not `\n`. Signal 6 (consecutive blank lines) reads the
// physical lines: a blank line directly after a blank line, neither inside a
// fence block, reported at the second (kan-1034: both slipped past review).
//
// # Signal 1 — broken code span
//
// A CommonMark code span opens at a run of N backticks and closes at the next
// run of exactly N backticks; backslash is not an escape inside a code span,
// so a single-backtick span written foo\`bar does not read as one span containing a literal backtick
// — the backslash is inert, the second run (length 1) closes the span early,
// and the run after `bar` is left with no partner at all. This guard finds
// that condition without hand-coding the escape shape directly: it collects
// every maximal run of backtick characters in a paragraph block's text, and
// greedily pairs each run with the nearest following run of the same length
// (CommonMark's own nearest-match rule); a run that never gets a partner
// leaves the guard unable to see where its span was meant to close, which is
// signal 1 in both of its native shapes at once — a stray backtick with no
// partner anywhere in the block, and a backslash-escape attempt that steals
// an early close and orphans the real one. A backtick run that is not length
// 1 is exempt from the escape story but still needs a partner of its own
// length: a double-backtick span around get `foo` done pairs the two
// double-backtick runs with each other and the two single backticks in between with each other, so the
// literal backtick inside is read correctly and nothing fires.
//
// # Signal 4 is structural, never phrase-based
//
// Signal 4 fires only when a paragraph's last line ends in a colon AND the
// next block, if any, is a heading (or there is no next block at all — end of
// file). It does not look at what the sentence says. This repository ends
// hundreds of sections with a citation of the shape `See **X** (f.md) for why
// ...`, correctly, and those end in a full stop, not a colon — the structural
// rule reads them as ordinary complete paragraphs and never fires. A
// phrase-based rule keying on "see ... for why" or "the reason is" would
// instead have to fire on all of them, and the lint policy this repository
// runs under forbids adding a suppression marker to quiet a rule that
// over-fires by design; the fix for an over-firing rule is always to narrow
// the rule; there is no other lever. Signal 4 is structural for that reason,
// deliberately, and stays that way.
//
// # Python parity
//
// The port reproduces what the .py read: universal newlines then
// str.splitlines (so \r, \v, \f, \x1c–\x1e, \x85, U+2028 and U+2029 break
// lines too), Unicode `\s`/`\d`/strip in the patterns, a strict UTF-8 decode
// whose failure is reported in Python's own words, os.walk's view of
// symlinks (a symlinked file is scanned, a symlinked directory is not
// entered), and paths printed as the root argument spelled them. Two
// departures: with no argument the root is FLOW_GUARD_REPO_ROOT, which the
// shim exports (the .py derived it from its own location), and an unreadable
// rules/ directory is the cannot-read refusal (exit 2) where the .py died
// with a traceback (exit 1).
func init() { Registry["check-markdown-integrity"] = checkMarkdownIntegrity }

const mdiName = "check-markdown-integrity"

// mdiTerminal — the set of characters that legitimately end a paragraph line
// (signal 3) or the prose line immediately preceding a blockquote (signal 2).
// A colon is included here on purpose: a paragraph ending in a colon is not
// itself unterminated (signal 3 does not fire on it), and prose ending in a
// colon immediately before a blockquote is the ordinary way to introduce a
// quotation, not an orphaned marker (signal 2 does not fire on it either).
// Signal 4 is the one place a trailing colon matters, and it asks a different
// question — not "is this terminated" but "is nothing following what the
// colon promised".
const mdiTerminal = ".:?!)|\"'`"

// mdiEmphasis — Markdown's own inline-emphasis closers (`**bold**`,
// `*italic*`, `__bold__`, `_italic_`). A sentence very commonly ends
// "...done.**" — the terminal punctuation sits inside the emphasis span,
// followed by its own closing marker. Stripping a trailing run of these
// characters before testing for terminal punctuation is what keeps that
// ordinary, correct shape from reading as unterminated; it is not an attempt
// to parse emphasis spans in general; a paragraph that ends with a bare,
// unmatched `*` and nothing else is stripped down to nothing here and
// reported (an empty result is treated as "no terminal punctuation found",
// not as an exemption).
const mdiEmphasis = "*_"

// mdiS is Python's Unicode `\s` (str.isspace), which RE2's ASCII `\s` is not.
const mdiS = `[\t\n\v\f\r\x1c-\x1f \x{85}\x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}]`

var (
	mdiFenceOpen  = regexp.MustCompile("^ {0,3}(`{3,}|~{3,})")
	mdiFenceClose = regexp.MustCompile("^ {0,3}(`+|~+)[ \t]*$")
	mdiThematic   = regexp.MustCompile(`^ {0,3}(-{3,}|\*{3,}|_{3,})[ \t]*$`)
	mdiHeading    = regexp.MustCompile(`^ {0,3}#{1,6}(` + mdiS + `|$)`)
	mdiListItem   = regexp.MustCompile(`^` + mdiS + `*([-*+]|\p{Nd}+[.)])` + mdiS + `+`)
	mdiBlockquote = regexp.MustCompile(`^` + mdiS + `*>(` + mdiS + `|$)`)
	mdiTableSep   = regexp.MustCompile(`^ {0,3}\|?` + mdiS + `*:?-+:?` + mdiS + `*(\|` + mdiS + `*:?-+:?` + mdiS + `*)*\|?` + mdiS + `*$`)
	mdiIndented   = regexp.MustCompile(`^ {4,}\S`)
	mdiWord       = regexp.MustCompile(`[A-Za-z']+`)
)

// mdiCompleting — kinds whose presence immediately after a paragraph means
// that paragraph is usually not torn — it is the CommonMark-legal "introduce,
// then supply" pattern this repository uses throughout: a clause ending in a
// bare noun or verb ("...run the following command") directly above a fenced
// command, or a bold pseudo-heading ("**Explore the problem space**")
// directly above the list it introduces. Calibrating this guard against this
// repository's own tree found dozens of exactly these two shapes, all
// correct, all repeated deliberately — see mdiUnterminated. This exemption
// alone is not discriminating enough, though: it fires on any paragraph in
// front of a fence, list item or table, whether or not that paragraph
// actually introduces it, so a genuinely torn sentence that merely happens to
// sit before an unrelated fence, list or table would pass uncaught.
// mdiDangling below is the second half of the test that closes that gap.
var mdiCompleting = map[string]bool{"fence": true, "list_item": true, "table": true}

// mdiDangling — the closed word classes that cannot end a grammatical clause
// on their own: articles, coordinating and common subordinating
// conjunctions, prepositions, relative pronouns, and auxiliary/linking verbs.
// An "introduce, then supply" paragraph reads as a complete clause even
// without terminal punctuation ("run the following command", "explore the
// problem space") — its last word is an ordinary noun or verb. A paragraph
// actually torn mid-clause instead trails off on one of these closed-class
// words ("...but then just stops without", "...because", "...and"), which is
// a structural fact about the word itself, not a judgement about the
// surrounding phrase. This is the discriminator mdiCompleting is paired
// against: a paragraph before a completing block is exempt only when its last
// word is not in this set; when it is, the paragraph is reported as
// unterminated regardless of what follows it. The set is deliberately
// closed-class (a finite, enumerable part of speech) rather than open-class
// (content words), which is what keeps the test structural instead of a
// phrase-matching heuristic.
var mdiDangling = func() map[string]bool {
	m := map[string]bool{}
	for _, w := range strings.Fields(`
		a an the
		and but or nor so yet
		because although though while since if unless until when whereas whether as
		of to in on at by with without from into onto upon over under above below
		between among through during before after about against toward towards within
		across along around behind beside beyond despite except inside near outside per
		throughout till unto via for
		which that who whom whose
		is are was were be been being am do does did have has had
		will would shall should can could may might must`) {
		m[w] = true
	}
	return m
}()

type mdiBlock struct {
	kind       string
	start, end int // 1-based physical lines
	lines      []string
}

type mdiFinding struct {
	line            int
	signal, message string
}

func mdiIsSpace(r rune) bool { return unicode.IsSpace(r) || (r >= 0x1c && r <= 0x1f) }

func mdiStrip(s string) string { return strings.TrimFunc(s, mdiIsSpace) }

// mdiSplitLines is Python's universal-newline read followed by
// str.splitlines(): no trailing empty element.
func mdiSplitLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	var lines []string
	start := 0
	for i, r := range text {
		switch r {
		case '\n', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
			lines = append(lines, text[start:i])
			start = i + utf8.RuneLen(r)
		}
	}
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

// mdiHTMLComment is `^ {0,3}<!--(?:(?!-->).)*-->[ \t]*$`: a line that is
// nothing but an HTML comment — `<!-- core -->`, `<!-- rule: x.mdc -->`, a
// provenance banner. Punctuation-free by construction and invisible when
// rendered, so like a thematic break it is its own block and never a
// paragraph that failed to end. Deliberately single-line and anchored: a
// comment that opens on one line and closes on another is prose-adjacent
// enough to keep scanning. RE2 has no lookahead, so the first `-->` after
// the opener is found by hand.
func mdiHTMLComment(line string) bool {
	for i := 0; i < 3 && strings.HasPrefix(line, " "); i++ {
		line = line[1:]
	}
	rest, ok := strings.CutPrefix(line, "<!--")
	if !ok {
		return false
	}
	i := strings.Index(rest, "-->")
	return i >= 0 && strings.Trim(rest[i+3:], " \t") == ""
}

// mdiOpens returns the block kind line unconditionally opens (a fence, a
// heading, a thematic break, a standalone HTML comment, a blockquote, or a
// new list marker), or "". Table detection is the caller's, since it needs
// the following line (the separator row) to be sure.
func mdiOpens(line string) string {
	switch {
	case mdiFenceOpen.MatchString(line):
		return "fence"
	case mdiThematic.MatchString(line):
		return "thematic_break"
	case mdiHTMLComment(line):
		return "html_comment"
	case mdiHeading.MatchString(line):
		return "heading"
	case mdiBlockquote.MatchString(line):
		return "blockquote"
	case mdiListItem.MatchString(line):
		return "list_item"
	}
	return ""
}

// mdiParseBlocks walks text into a flat sequence of typed blocks. Blank lines
// end the current block without becoming blocks of their own. A paragraph or
// list_item block absorbs ordinary continuation lines; every other kind is
// single-line or, for fence, consumes up to its own closing delimiter.
func mdiParseBlocks(text string) []*mdiBlock {
	lines := mdiSplitLines(text)
	n := len(lines)
	var blocks []*mdiBlock
	i := 0

	// Frontmatter is only checked once, at the very start of the file — a
	// `---` appearing later is an ordinary thematic break.
	if n > 0 && mdiStrip(lines[0]) == "---" {
		for j := 1; j < n; j++ {
			if s := mdiStrip(lines[j]); s == "---" || s == "***" || s == "___" {
				blocks = append(blocks, &mdiBlock{"frontmatter", 1, j + 1, lines[:j+1]})
				i = j + 1
				break
			}
		}
	}

	var cur *mdiBlock
	inFence := false
	var fenceChar byte
	fenceLen := 0
	flush := func() {
		if cur != nil {
			blocks = append(blocks, cur)
			cur = nil
		}
	}
	open := func(kind string, lineno int, raw string) {
		cur = &mdiBlock{kind, lineno, lineno, []string{raw}}
	}
	extend := func(lineno int, raw string) {
		cur.lines = append(cur.lines, raw)
		cur.end = lineno
	}

	for ; i < n; i++ {
		lineno, raw := i+1, lines[i]

		if inFence {
			extend(lineno, raw)
			if m := mdiFenceClose.FindStringSubmatch(raw); m != nil && m[1][0] == fenceChar && len(m[1]) >= fenceLen {
				inFence = false
				flush()
			}
			continue
		}
		if mdiStrip(raw) == "" {
			flush()
			continue
		}

		switch kind := mdiOpens(raw); kind {
		case "fence":
			flush()
			run := mdiFenceOpen.FindStringSubmatch(raw)[1]
			fenceChar, fenceLen, inFence = run[0], len(run), true
			open(kind, lineno, raw)
			continue
		case "thematic_break", "heading", "html_comment":
			flush()
			blocks = append(blocks, &mdiBlock{kind, lineno, lineno, []string{raw}})
			continue
		case "blockquote":
			if cur != nil && cur.kind == "blockquote" {
				extend(lineno, raw)
			} else {
				flush()
				open(kind, lineno, raw)
			}
			continue
		case "list_item":
			// A fresh marker always starts a new list_item block, even if
			// one was already open — this guard does not track nesting
			// depth or item boundaries beyond "some list item is open".
			flush()
			open(kind, lineno, raw)
			continue
		}

		// Not an unconditional opener. If a paragraph/list_item is already
		// open, this is an ordinary lazy-continuation line.
		if cur != nil && (cur.kind == "paragraph" || cur.kind == "list_item") {
			extend(lineno, raw)
			continue
		}
		// If a table is already open, a line still carrying a `|` extends
		// it; one that doesn't ends it.
		if cur != nil && cur.kind == "table" {
			if strings.Contains(raw, "|") {
				extend(lineno, raw)
				continue
			}
			flush()
		}
		// Fresh block. A `|`-bearing line only starts a table when the very
		// next line is a genuine GFM separator row — this lookahead is what
		// keeps an ordinary sentence with one stray `|` in it (a shell pipe
		// mentioned in prose, a record's own `|` field) from being misread
		// as a table header.
		if strings.Contains(raw, "|") && i+1 < n && mdiTableSep.MatchString(lines[i+1]) {
			flush()
			open("table", lineno, raw)
			continue
		}
		// A fresh, blank-line-preceded (or file-start) line indented four or
		// more spaces is a CommonMark indented code block, not prose — reused
		// as kind "fence" so it gets the same blanket exclusion from every
		// prose signal a fenced block already gets. This is only ever
		// checked here, at a fresh block boundary: an indented continuation
		// line of an already-open paragraph or list item is handled above
		// and never reaches this branch.
		if mdiIndented.MatchString(raw) {
			var block []string
			for i < n {
				c := lines[i]
				if mdiStrip(c) == "" {
					j := i + 1
					for j < n && mdiStrip(lines[j]) == "" {
						j++
					}
					if j < n && mdiIndented.MatchString(lines[j]) {
						block = append(block, c)
						i++
						continue
					}
					break
				}
				if !mdiIndented.MatchString(c) {
					break
				}
				block = append(block, c)
				i++
			}
			blocks = append(blocks, &mdiBlock{"fence", lineno, lineno + len(block) - 1, block})
			i-- // the loop's i++ lands on the first line past the block
			continue
		}
		flush()
		open("paragraph", lineno, raw)
	}
	flush()
	return blocks
}

// mdiFirstUnmatchedBacktick is the byte offset of the first backtick run in
// text that never finds a same-length partner, greedily, in CommonMark's own
// nearest-match order — or -1 if every run pairs off.
func mdiFirstUnmatchedBacktick(text string) int {
	type run struct{ at, n int }
	var runs []run
	for i := 0; i < len(text); {
		if text[i] != '`' {
			i++
			continue
		}
		j := i
		for j < len(text) && text[j] == '`' {
			j++
		}
		runs = append(runs, run{i, j - i})
		i = j
	}
	for i := 0; i < len(runs); {
		partner := -1
		for j := i + 1; j < len(runs); j++ {
			if runs[j].n == runs[i].n {
				partner = j
				break
			}
		}
		if partner < 0 {
			return runs[i].at
		}
		i = partner + 1
	}
	return -1
}

// mdiLineOfOffset maps an offset into b's "\n"-joined text back to the
// physical line it falls on.
func mdiLineOfOffset(b *mdiBlock, offset int) int {
	pos := 0
	for idx, line := range b.lines {
		if offset < pos+len(line)+1 {
			return b.start + idx
		}
		pos += len(line) + 1
	}
	return b.end
}

func mdiStripEmphasis(s string) string { return strings.TrimRight(s, mdiEmphasis) }

func mdiEndsSentence(line string) bool {
	s := mdiStripEmphasis(strings.TrimRightFunc(line, mdiIsSpace))
	return s != "" && strings.ContainsRune(mdiTerminal, rune(s[len(s)-1]))
}

// mdiEndsOnDangling is true when line's last word (after stripping trailing
// emphasis markers) is one of mdiDangling — a closed-class word that cannot
// end a grammatical clause by itself. Used to tell a paragraph that genuinely
// tears off mid-clause apart from one that merely introduces the block after
// it.
func mdiEndsOnDangling(line string) bool {
	words := mdiWord.FindAllString(mdiStripEmphasis(strings.TrimRightFunc(line, mdiIsSpace)), -1)
	return len(words) > 0 && mdiDangling[strings.ToLower(words[len(words)-1])]
}

// mdiResumes is true when b picks the sentence back up rather than starting
// a new one — the shape a paragraph takes when it deliberately spans an
// embedded fence, list, or table ("...so a tag can run past the first `>` —
// and" / a code example / "exited 0 with the bare claim unreported..."). This
// is the second half of the dangling-word test: a paragraph ending on a
// dangling word is only torn when nothing after the completing block resumes
// it.
//
// nil (the completing block was the last block in the file) never resumes
// anything — there is nothing left to supply the rest of the sentence. A
// heading or list_item never resumes one either: both are marked structure
// (a `#` line, a `-`/`1.` marker) that always opens something new, never bare
// continuation prose.
//
// Otherwise the check is the case of the block's own first letter.
// CommonMark prose convention capitalizes the first word of a new sentence;
// a block resuming a sentence torn off mid-clause instead opens on whatever
// the sentence's next word already was — lower case, a digit, or inline
// markup such as a backtick or quote mark, none of which is an upper-case
// letter. So "not upper-case" is the structural stand-in for "resumes," and
// "upper-case" for "starts fresh" — no word list is involved, only the
// character's own case. A leading `>` blockquote marker is stripped first so
// the check looks at the quoted text itself rather than the marker
// punctuation.
func mdiResumes(b *mdiBlock) bool {
	if b == nil || b.kind == "heading" || b.kind == "list_item" {
		return false
	}
	text := b.lines[0]
	if loc := mdiBlockquote.FindStringIndex(text); loc != nil {
		text = text[loc[1]:]
	}
	text = strings.TrimLeftFunc(text, mdiIsSpace)
	if text == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(text)
	return !unicode.IsUpper(r)
}

func mdiAt(blocks []*mdiBlock, i int) *mdiBlock {
	if i < len(blocks) {
		return blocks[i]
	}
	return nil
}

func mdiScan(text string) []mdiFinding {
	blocks := mdiParseBlocks(text)
	var out []mdiFinding
	// Signal 1 — broken code span.
	for _, b := range blocks {
		if b.kind != "paragraph" {
			continue
		}
		if off := mdiFirstUnmatchedBacktick(strings.Join(b.lines, "\n")); off >= 0 {
			out = append(out, mdiFinding{mdiLineOfOffset(b, off), "broken code span",
				"a backtick run in this paragraph has no same-length partner, so a code span closes early or never closes at all"})
		}
	}
	// Signal 2 — orphaned blockquote body.
	for idx, b := range blocks {
		if b.kind != "blockquote" || idx == 0 {
			continue
		}
		prev := blocks[idx-1]
		// A heading is marked, not un-marked prose, and heading text
		// routinely carries no terminal punctuation of its own ("## Next
		// heading") — that is normal heading style, not a sentence torn from
		// the blockquote that follows it. Signal 2 exists to catch prose
		// that lost its own `>` marker; a heading was never eligible to
		// carry one, so it is excluded outright.
		//
		// A fence, list_item, or table is marked structure too, for the same
		// reason mdiCompleting already excludes them from signal 3's "torn
		// paragraph" test: a closing ```/~~~ delimiter, a tight list item's
		// own text, or a table's separator row is not prose that could have
		// lost a `>` marker. Reusing mdiCompleting here rather than a second
		// hand-written list keeps the two exemptions from drifting apart.
		if prev.kind == "blockquote" || prev.kind == "heading" || mdiCompleting[prev.kind] {
			continue
		}
		if prev.end+1 != b.start || mdiEndsSentence(prev.lines[len(prev.lines)-1]) {
			continue
		}
		out = append(out, mdiFinding{prev.end, "orphaned blockquote body",
			"un-marked prose sits immediately before a blockquote and does not end a sentence — this line may have lost its own `>` marker"})
	}
	// Signal 3 — unterminated paragraph. A paragraph directly introducing a
	// fence, a list, or a table is not torn — the block that follows
	// supplies what the paragraph led into. mdiDangling narrows that: the
	// exemption applies only when the paragraph's own last word could
	// plausibly end a clause, unless the block past the fence/list/table
	// picks the sentence back up (mdiResumes).
	for idx, b := range blocks {
		if b.kind != "paragraph" {
			continue
		}
		last := b.lines[len(b.lines)-1]
		if mdiEndsSentence(last) {
			continue
		}
		if nxt := mdiAt(blocks, idx+1); nxt != nil && mdiCompleting[nxt.kind] {
			if !mdiEndsOnDangling(last) || mdiResumes(mdiAt(blocks, idx+2)) {
				continue
			}
		}
		out = append(out, mdiFinding{b.end, "unterminated paragraph",
			"this paragraph's last line ends with no terminal punctuation"})
	}
	// Signal 4 — dangling promise.
	for idx, b := range blocks {
		if b.kind != "paragraph" {
			continue
		}
		if !strings.HasSuffix(mdiStripEmphasis(strings.TrimRightFunc(b.lines[len(b.lines)-1], mdiIsSpace)), ":") {
			continue
		}
		if nxt := mdiAt(blocks, idx+1); nxt == nil || nxt.kind == "heading" {
			out = append(out, mdiFinding{b.end, "dangling promise",
				"a colon promises content, and the block ends with nothing following it"})
		}
	}
	// Signal 5 — missing final newline.
	lines := mdiSplitLines(text)
	if text != "" && !strings.HasSuffix(text, "\n") {
		out = append(out, mdiFinding{len(lines), "missing final newline",
			"the file does not end in a newline"})
	}
	// Signal 6 — consecutive blank lines outside a fence.
	inFence := func(line int) bool {
		for _, b := range blocks {
			if b.kind == "fence" && line >= b.start && line <= b.end {
				return true
			}
		}
		return false
	}
	for i := 1; i < len(lines); i++ {
		if mdiStrip(lines[i]) == "" && mdiStrip(lines[i-1]) == "" && !inFence(i) && !inFence(i+1) {
			out = append(out, mdiFinding{i + 1, "consecutive blank lines",
				"two blank lines in a row; one separates blocks"})
		}
	}
	// Signal 7 — a handoff template's next command not alone in its own fence
	// (kan-1040): inside a fenced template, `Next:` (and an optional `/clear`)
	// is followed by a three-backtick fence holding one `/flow` line. A template
	// whose outer fence is too short closes on that inner fence, so its block
	// ends before the command and fires here too.
	for _, b := range blocks {
		if b.kind != "fence" {
			continue
		}
		for i, l := range b.lines {
			if l != "Next:" {
				continue
			}
			j := i + 1
			if j < len(b.lines) && b.lines[j] == "/clear" {
				j++
			}
			if j+2 >= len(b.lines) || b.lines[j] != "```" || !strings.HasPrefix(b.lines[j+1], "/flow") || b.lines[j+2] != "```" {
				out = append(out, mdiFinding{b.start + i, "unfenced next command",
					"a template's next command sits alone in a ``` fence inside a ```` template fence"})
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].line < out[j].line })
	return out
}

// mdiPyJoin is os.path.join for one relative component: the left side kept
// as spelled, unlike filepath.Join, so printed paths match the .py's.
func mdiPyJoin(a, b string) string {
	if a == "" || strings.HasSuffix(a, "/") {
		return a + b
	}
	return a + "/" + b
}

// mdiOSError renders err as Python's OSError str: `[Errno N] Strerror: 'path'`.
func mdiOSError(err error, path string) string {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return fmt.Sprintf("[Errno %d] %s: '%s'", int(errno), psStrerror(errno), path)
	}
	return err.Error()
}

// mdiDecodeError renders the first invalid UTF-8 sequence in b as Python's
// UnicodeDecodeError str, or "" when b is valid.
func mdiDecodeError(b []byte) string {
	for i := 0; i < len(b); {
		r, size := utf8.DecodeRune(b[i:])
		if r != utf8.RuneError || size > 1 {
			i += size
			continue
		}
		c := b[i]
		if c < 0xc2 || c > 0xf4 {
			return fmt.Sprintf("'utf-8' codec can't decode byte 0x%02x in position %d: invalid start byte", c, i)
		}
		need := 2
		if c >= 0xf0 {
			need = 4
		} else if c >= 0xe0 {
			need = 3
		}
		j := i + 1
		for j < len(b) && j < i+need && b[j]&0xc0 == 0x80 {
			j++
		}
		if j == len(b) && j < i+need {
			if j-i == 1 {
				return fmt.Sprintf("'utf-8' codec can't decode byte 0x%02x in position %d: unexpected end of data", c, i)
			}
			return fmt.Sprintf("'utf-8' codec can't decode bytes in position %d-%d: unexpected end of data", i, j-1)
		}
		if j-i == 1 {
			return fmt.Sprintf("'utf-8' codec can't decode byte 0x%02x in position %d: invalid continuation byte", c, i)
		}
		return fmt.Sprintf("'utf-8' codec can't decode bytes in position %d-%d: invalid continuation byte", i, j-1)
	}
	return ""
}

func checkMarkdownIntegrity(args []string, env Env, stdout, stderr io.Writer) int {
	const usage = " — this guard takes at most one, an optional project root\n"
	var root string
	switch {
	case len(args) > 1:
		fmt.Fprint(stderr, mdiName+": unknown argument"+usage)
		return 2
	case len(args) == 1 && strings.HasPrefix(args[0], "-"):
		fmt.Fprintf(stderr, mdiName+": unknown argument '%s'"+usage, args[0])
		return 2
	case len(args) == 1:
		root = args[0]
	default:
		if root = env.Getenv("FLOW_GUARD_REPO_ROOT"); root == "" {
			fmt.Fprint(stderr, mdiName+": FLOW_GUARD_REPO_ROOT is unset — run scripts/check-markdown-integrity.sh, which sets it\n")
			return 2
		}
	}
	// fsPath resolves a printed path against the guard's working directory.
	fsPath := func(p string) string {
		if filepath.IsAbs(p) {
			return p
		}
		return filepath.Join(env.Dir, p)
	}
	isDir := func(p string) bool {
		st, err := os.Stat(fsPath(p))
		return err == nil && st.IsDir()
	}
	if !isDir(root) {
		fmt.Fprintf(stderr, mdiName+": cannot answer — scope root does not exist: %s\n", root)
		return 2
	}

	// Scope: skills/**/*.md and rules/*.mdc. A walk as os.walk does it: an
	// unreadable subdirectory is skipped, a symlinked directory is listed
	// but never entered, a symlinked file is scanned.
	var files []string
	var walk func(dir string)
	walk = func(dir string) {
		ents, _ := os.ReadDir(fsPath(dir))
		for _, e := range ents {
			p := dir + "/" + e.Name()
			if e.Type()&fs.ModeSymlink != 0 {
				if !isDir(p) && strings.HasSuffix(p, ".md") {
					files = append(files, p)
				}
			} else if e.IsDir() {
				walk(p)
			} else if strings.HasSuffix(p, ".md") {
				files = append(files, p)
			}
		}
	}
	if skills := mdiPyJoin(root, "skills"); isDir(skills) {
		walk(skills)
	}
	if rules := mdiPyJoin(root, "rules"); isDir(rules) {
		ents, err := os.ReadDir(fsPath(rules))
		if err != nil {
			fmt.Fprintf(stderr, mdiName+": cannot read %s: %s\n", rules, mdiOSError(err, rules))
			return 2
		}
		for _, e := range ents {
			p := rules + "/" + e.Name()
			if st, err := os.Stat(fsPath(p)); err == nil && st.Mode().IsRegular() && strings.HasSuffix(p, ".mdc") {
				files = append(files, p)
			}
		}
	}
	sort.Strings(files)

	var report []string
	for _, p := range files {
		b, err := os.ReadFile(fsPath(p))
		if err != nil {
			fmt.Fprintf(stderr, mdiName+": cannot read %s: %s\n", p, mdiOSError(err, p))
			return 2
		}
		if msg := mdiDecodeError(b); msg != "" {
			fmt.Fprintf(stderr, mdiName+": cannot read %s: %s\n", p, msg)
			return 2
		}
		for _, f := range mdiScan(string(b)) {
			report = append(report, fmt.Sprintf("%s:%d: %s: %s\n", p, f.line, f.signal, f.message))
		}
	}
	for _, line := range report {
		fmt.Fprint(stdout, line)
	}
	if len(report) > 0 {
		return 1
	}
	return 0
}
