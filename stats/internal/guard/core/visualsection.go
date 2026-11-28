package core

import (
	"bytes"
	"regexp"
	"strings"
)

// The visual-verify guards' shared parsing: the Go twins of the bash
// libraries every visual guard once sourced, all since deleted — the last
// three by kan-919 — plus the `## visual verification` section rule every
// one of those scripts carried inline. TestStripBOMParity,
// TestSanitizeDisplayParity and TestVisualTableCellsParity run the deleted
// libraries, as they last stood, against these twins (kan850-helper-twins).
//
// `.flow/project.md` is tracked in the repository and editable in any pull
// request, so every input here is attacker-influenced: the twins are byte
// oriented (C locale — [[:space:]] is CSpace, tolower folds A-Z only) and
// linear in their input.

// CSpace is the C locale's [[:space:]].
const CSpace = " \t\n\v\f\r"

// StripBOM drops a leading UTF-8 BOM (EF BB BF) — only at byte 0, only once.
// A `.flow/project.md` some editor or Windows tooling saved with a BOM, with
// `## visual verification` as its first line, puts those three bytes before
// the `#` every `^##` heading anchor expects at column one, and a section that
// is genuinely present reads as absent. The BOM fix once reached one visual
// guard of three and had to be extracted a commit later — the reason every
// caller strips through this one function.
func StripBOM(b []byte) []byte {
	return bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))
}

// SanitizeDisplay is the bash libraries' display sanitizer: text (stdin) to
// stdout, one line at a time, with every C0 control byte, DEL and backslash rendered as visible
// text, and every line — the last one included — ending in "\n". Every line a
// caller prints through it quotes a cell out of `.flow/project.md`, so a cell
// holding an escape sequence would otherwise write that sequence to the
// operator's terminal — a forged-verdict hazard. ESCAPED, NOT STRIPPED AND NOT
// REFUSED: stripping rewrites the text the operator is sent to go and fix,
// and refusing hands whoever edited the file a way to suppress a real
// finding.
//
// Not Sanitize: the bash's awk ends a line at a NUL byte (its escape table
// starts at 1, and the one-true-awk's strings are NUL-terminated), where
// Sanitize escapes the NUL and keeps the rest — TestSanitizeDisplayParity
// measured the difference.
func SanitizeDisplay(text string) string {
	var b strings.Builder
	for _, line := range AwkRecords(text) {
		if i := strings.IndexByte(line, 0); i >= 0 {
			line = line[:i]
		}
		b.WriteString(Sanitize(line))
		b.WriteByte('\n')
	}
	return b.String()
}

// AwkRecords splits text into awk's newline-separated records: a final "\n"
// ends the last record rather than opening an empty one.
func AwkRecords(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// SplitCells is split_cells: a `| a | b |` row's raw cells, or nil when the
// line (CRs removed, edge whitespace trimmed) does not start with `|`. A `\|`
// or `\\` inside a cell unescapes to its second byte and does not split; any
// other backslash is kept literally; a missing trailing `|` is tolerated by
// dropping one final empty cell. One pass over the bytes — the awk's own
// per-character loop was quadratic on a padded cell and was rewritten as a
// delimiter walk; a builder is linear here without that contortion.
func SplitCells(line string) []string {
	line = strings.Trim(strings.ReplaceAll(line, "\r", ""), CSpace)
	if !strings.HasPrefix(line, "|") {
		return nil
	}
	rest := line[1:]
	var cells []string
	var cur strings.Builder
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		switch {
		case c == '|':
			cells = append(cells, cur.String())
			cur.Reset()
		case c == '\\' && i+1 < len(rest) && (rest[i+1] == '|' || rest[i+1] == '\\'):
			i++
			cur.WriteByte(rest[i])
		default:
			cur.WriteByte(c)
		}
	}
	cells = append(cells, cur.String())
	if len(cells) > 1 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	return cells
}

// TrimCell is trimcell: c with leading and trailing whitespace and
// backticks removed. (The awk avoids a `$`-anchored match, which costs
// seconds on this awk against a padded cell; strings.Trim has no such cost.)
func TrimCell(c string) string {
	return strings.Trim(c, CSpace+"`")
}

// FoldCell is foldcell: TrimCell, ASCII-lowercased, interior whitespace
// runs collapsed to one space — the shape a header or a `Setting`/`Command`
// name is compared in.
func FoldCell(c string) string {
	return strings.Join(strings.FieldsFunc(AwkLower(TrimCell(c)), func(r rune) bool {
		return r < 0x80 && strings.ContainsRune(CSpace, r)
	}), " ")
}

// Heading is the scripts' VV_HEADING,
// `^##[[:space:]]+visual verification[[:space:]]*$`, matched against an
// ASCII-lowercased line (grep -i / tolower in the C locale). Not (?i): Go's
// case folding is Unicode's, and would match `ſ` for `s`.
var Heading = regexp.MustCompile(`^##[ \t\n\v\f\r]+visual verification[ \t\n\v\f\r]*$`)

// Line is one line of the section: No is its 1-based line number in the
// BOM-stripped text, Text the line as read (a CRLF file's CR kept).
type Line struct {
	No   int
	Text string
}

// TextLines is the BOM-stripped text's records, each ended at its first NUL
// as the scripts' grep and awk ended it — here, once, so the heading count
// and the section every caller reads agree on one file.
func TextLines(text string) []string {
	recs := AwkRecords(string(StripBOM([]byte(text))))
	for i := range recs {
		recs[i], _, _ = strings.Cut(recs[i], "\x00")
	}
	return recs
}

// HeadingCount is how many lines of text (BOM stripped) are a
// `## visual verification` heading — more than one is an ambiguous
// declaration every caller refuses.
func HeadingCount(text string) int {
	n := 0
	for _, l := range TextLines(text) {
		if Heading.MatchString(AwkLower(l)) {
			n++
		}
	}
	return n
}

var vvHeadingLine = regexp.MustCompile(`^#+[ \t\n\v\f\r]`)

// SectionLines is the lines inside `## visual verification` under the
// visual scripts' shared awk rule: a `#+[[:space:]]` heading line is never a
// section line; the section's own heading opens the section; any other
// heading closes it — unless it is deeper than level 2 while inside, so a
// `### sub` heading keeps the section open.
func SectionLines(text string) []Line {
	var out []Line
	in := false
	for i, l := range TextLines(text) {
		if vvHeadingLine.MatchString(l) {
			own := Heading.MatchString(AwkLower(l))
			level := len(l) - len(strings.TrimLeft(l, "#"))
			if !(in && !own && level > 2) {
				in = own
			}
			continue
		}
		if in {
			out = append(out, Line{i + 1, l})
		}
	}
	return out
}
