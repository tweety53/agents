package guard

import (
	"bytes"
	"regexp"
	"strings"
)

// The visual-verify guards' shared parsing: the Go twins of
// scripts/lib/strip-bom.sh, scripts/lib/sanitize-display.sh,
// scripts/lib/visual-table-cells.awk and the since-deleted
// scripts/lib/trim-glob-element.sh,
// plus the `## visual verification` section rule every one of those scripts
// carried inline. The bash libraries stay for their remaining bash callers
// (check-spec-reach.sh, check-dev-stack-fresh.sh, lib/project-section.sh);
// TestStripBOMParity, TestSanitizeDisplayParity and TestVisualTableCellsParity
// run them against these twins so a fix to one side cannot miss the other
// (kan850-helper-twins).
//
// `.flow/project.md` is tracked in the repository and editable in any pull
// request, so every input here is attacker-influenced: the twins are byte
// oriented (C locale — [[:space:]] is cSpace, tolower folds A-Z only) and
// linear in their input.

// cSpace is the C locale's [[:space:]].
const cSpace = " \t\n\v\f\r"

// stripBOM drops a leading UTF-8 BOM (EF BB BF) — only at byte 0, only once.
// A `.flow/project.md` some editor or Windows tooling saved with a BOM, with
// `## visual verification` as its first line, puts those three bytes before
// the `#` every `^##` heading anchor expects at column one, and a section that
// is genuinely present reads as absent. The BOM fix once reached one visual
// guard of three and had to be extracted a commit later — the reason every
// caller strips through this one function.
func stripBOM(b []byte) []byte {
	return bytes.TrimPrefix(b, []byte("\xef\xbb\xbf"))
}

// sanitizeDisplay is sanitize_display: text (stdin) to stdout, one line at a
// time, with every C0 control byte, DEL and backslash rendered as visible
// text, and every line — the last one included — ending in "\n". Every line a
// caller prints through it quotes a cell out of `.flow/project.md`, so a cell
// holding an escape sequence would otherwise write that sequence to the
// operator's terminal — a forged-verdict hazard. ESCAPED, NOT STRIPPED AND NOT
// REFUSED: stripping rewrites the text the operator is sent to go and fix,
// and refusing hands whoever edited the file a way to suppress a real
// finding.
//
// Not ccSanitize: the bash's awk ends a line at a NUL byte (its escape table
// starts at 1, and the one-true-awk's strings are NUL-terminated), where
// ccSanitize escapes the NUL and keeps the rest — TestSanitizeDisplayParity
// measured the difference.
func sanitizeDisplay(text string) string {
	var b strings.Builder
	for _, line := range awkRecords(text) {
		if i := strings.IndexByte(line, 0); i >= 0 {
			line = line[:i]
		}
		b.WriteString(ccSanitize(line))
		b.WriteByte('\n')
	}
	return b.String()
}

// awkRecords splits text into awk's newline-separated records: a final "\n"
// ends the last record rather than opening an empty one.
func awkRecords(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

// vtSplitCells is split_cells: a `| a | b |` row's raw cells, or nil when the
// line (CRs removed, edge whitespace trimmed) does not start with `|`. A `\|`
// or `\\` inside a cell unescapes to its second byte and does not split; any
// other backslash is kept literally; a missing trailing `|` is tolerated by
// dropping one final empty cell. One pass over the bytes — the awk's own
// per-character loop was quadratic on a padded cell and was rewritten as a
// delimiter walk; a builder is linear here without that contortion.
func vtSplitCells(line string) []string {
	line = strings.Trim(strings.ReplaceAll(line, "\r", ""), cSpace)
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

// vtTrimCell is trimcell: c with leading and trailing whitespace and
// backticks removed. (The awk avoids a `$`-anchored match, which costs
// seconds on this awk against a padded cell; strings.Trim has no such cost.)
func vtTrimCell(c string) string {
	return strings.Trim(c, cSpace+"`")
}

// vtFoldCell is foldcell: vtTrimCell, ASCII-lowercased, interior whitespace
// runs collapsed to one space — the shape a header or a `Setting`/`Command`
// name is compared in.
func vtFoldCell(c string) string {
	return strings.Join(strings.FieldsFunc(gsASCIILower(vtTrimCell(c)), func(r rune) bool {
		return r < 0x80 && strings.ContainsRune(cSpace, r)
	}), " ")
}

// trimGlobElement is trim_glob_element: a leading/trailing run of spaces,
// tabs and backticks removed — never an interior one, so a glob containing a
// space (a directory name with one) or a backtick of its own survives.
// Applied per comma-split element of `ui paths`, never to a whole cell:
// splitting first and stripping each element afterwards is design.md's
// `split-then-strip` (KAN-359), the whole fix for a value carrying more than
// one individually backticked glob. The bash version measured a quadratic
// character loop before settling on one awk pass; strings.Trim is linear.
func trimGlobElement(s string) string {
	return strings.Trim(s, " \t`")
}

// vvHeading is the scripts' VV_HEADING,
// `^##[[:space:]]+visual verification[[:space:]]*$`, matched against an
// ASCII-lowercased line (grep -i / tolower in the C locale). Not (?i): Go's
// case folding is Unicode's, and would match `ſ` for `s`.
var vvHeading = regexp.MustCompile(`^##[ \t\n\v\f\r]+visual verification[ \t\n\v\f\r]*$`)

// vvLine is one line of the section: No is its 1-based line number in the
// BOM-stripped text, Text the line as read (a CRLF file's CR kept).
type vvLine struct {
	No   int
	Text string
}

// vvLines is the BOM-stripped text's records, each ended at its first NUL
// as the scripts' grep and awk ended it — here, once, so the heading count
// and the section every caller reads agree on one file.
func vvLines(text string) []string {
	recs := awkRecords(string(stripBOM([]byte(text))))
	for i := range recs {
		recs[i], _, _ = strings.Cut(recs[i], "\x00")
	}
	return recs
}

// vvHeadingCount is how many lines of text (BOM stripped) are a
// `## visual verification` heading — more than one is an ambiguous
// declaration every caller refuses.
func vvHeadingCount(text string) int {
	n := 0
	for _, l := range vvLines(text) {
		if vvHeading.MatchString(gsASCIILower(l)) {
			n++
		}
	}
	return n
}

var vvHeadingLine = regexp.MustCompile(`^#+[ \t\n\v\f\r]`)

// vvSectionLines is the lines inside `## visual verification` under the
// visual scripts' shared awk rule: a `#+[[:space:]]` heading line is never a
// section line; the section's own heading opens the section; any other
// heading closes it — unless it is deeper than level 2 while inside, so a
// `### sub` heading keeps the section open.
func vvSectionLines(text string) []vvLine {
	var out []vvLine
	in := false
	for i, l := range vvLines(text) {
		if vvHeadingLine.MatchString(l) {
			own := vvHeading.MatchString(gsASCIILower(l))
			level := len(l) - len(strings.TrimLeft(l, "#"))
			if !(in && !own && level > 2) {
				in = own
			}
			continue
		}
		if in {
			out = append(out, vvLine{i + 1, l})
		}
	}
	return out
}
