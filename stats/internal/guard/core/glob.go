package core

import (
	"regexp"
	"strings"
)

// TrimGlobElement is trim_glob_element: a leading/trailing run of spaces,
// tabs and backticks removed — never an interior one, so a glob containing a
// space (a directory name with one) or a backtick of its own survives.
// Applied per comma-split element of `ui paths`, never to a whole cell:
// splitting first and stripping each element afterwards is design.md's
// `split-then-strip` (KAN-359), the whole fix for a value carrying more than
// one individually backticked glob. The bash version measured a quadratic
// character loop before settling on one awk pass; strings.Trim is linear.
func TrimGlobElement(s string) string {
	return strings.Trim(s, " \t`")
}

// Cells maps each two-cell row's folded name inside the `## visual
// verification` section to its trimmed value, the first row of a name winning.
func Cells(text string) map[string]string {
	cells := map[string]string{}
	for _, l := range SectionLines(text) {
		if c := SplitCells(l.Text); len(c) == 2 {
			if k := FoldCell(c[0]); cells[k] == "" {
				cells[k] = TrimCell(c[1])
			}
		}
	}
	return cells
}

// GlobRegexp is the bash's glob_to_ere plus glob_match's anchoring: a
// leading `./` and then a leading `/` stripped from the glob; `**` becomes
// `.*` (spans `/`); a bare `*` becomes `[^/]*` (one segment); `?` becomes
// `[^/]`; every listed metacharacter is escaped; anything else is literal.
//
// Byte-wise, as the bash matched under LC_ALL=C (c-locale-table-semantics):
// both the glob and the path go through BytesAsRunes, one rune per byte,
// so `?` is one byte and a non-UTF-8 byte is an ordinary character.
func GlobRegexp(glob string) *regexp.Regexp {
	glob = strings.TrimPrefix(strings.TrimPrefix(glob, "./"), "/")
	var b strings.Builder
	b.WriteString("^")
	for i := 0; i < len(glob); i++ {
		switch c := glob[i]; c {
		case '*':
			if i+1 < len(glob) && glob[i+1] == '*' {
				b.WriteString(".*")
				i++
				continue
			}
			b.WriteString("[^/]*")
		case '?':
			b.WriteString("[^/]")
		case '.', '+', '(', ')', '{', '}', '|', '^', '$', '\\', '[', ']':
			b.WriteByte('\\')
			b.WriteByte(c)
		default:
			b.WriteRune(rune(c))
		}
	}
	b.WriteString("$")
	return regexp.MustCompile(b.String())
}

// BytesAsRunes maps each byte of s to the rune of the same value, so a
// regexp over the result matches byte by byte.
func BytesAsRunes(s string) string {
	r := make([]rune, len(s))
	for i := 0; i < len(s); i++ {
		r[i] = rune(s[i])
	}
	return string(r)
}
