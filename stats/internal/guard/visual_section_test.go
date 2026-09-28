package guard

import (
	"reflect"
	"testing"
)

// TestVisualSection pins vvHeadingCount, vvSectionLines and trimGlobElement:
// the `## visual verification` heading count, the section's line set under
// the visual scripts' shared awk rule, and a glob element's edge trim.
func TestVisualSection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		text  string
		count int
		lines []vvLine
	}{
		{"section at file end", "# P\n\n## visual verification\n| a | b |\nlast", 1,
			[]vvLine{{4, "| a | b |"}, {5, "last"}}},
		{"closed by ## next", "## visual verification\nin\n## next\nout\n", 1,
			[]vvLine{{2, "in"}}},
		{"### sub stays inside", "## visual verification\na\n### sub\nb\n## next\nc\n", 1,
			[]vvLine{{2, "a"}, {4, "b"}}},
		{"# top closes it", "## visual verification\na\n# top\nb\n", 1,
			[]vvLine{{2, "a"}}},
		{"heading case and trailing spaces", "##   Visual VERIFICATION \t\nx\n", 1,
			[]vvLine{{2, "x"}}},
		{"CRLF", "## visual verification\r\n| a | b |\r\n## next\r\nz\r\n", 1,
			[]vvLine{{2, "| a | b |\r"}}},
		{"BOM", "\xef\xbb\xbf## visual verification\nx\n", 1,
			[]vvLine{{2, "x"}}},
		{"two sections", "## visual verification\na\n## other\nb\n## Visual Verification\nc\n", 2,
			[]vvLine{{2, "a"}, {6, "c"}}},
		{"no section", "# P\n## lint\nx\n", 0, nil},
		{"deeper own heading is not the section", "### visual verification\nx\n", 0, nil},
		{"no space after ## is not a heading match", "##visual verification\nx\n", 0, nil},
		{"#-line without a space is a section line", "## visual verification\n#tag\n", 1,
			[]vvLine{{2, "#tag"}}},
		{"non-ASCII fold is not ASCII", "## vİsual verification\nx\n", 0, nil},
		{"NUL ends the heading line", "## visual verification\x00x\n| a | b |\n", 1,
			[]vvLine{{2, "| a | b |"}}},
		{"NUL ends a section line", "## visual verification\n| ui paths | `a/**`\x00 |\n", 1,
			[]vvLine{{2, "| ui paths | `a/**`"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := vvHeadingCount(tc.text); got != tc.count {
				t.Errorf("vvHeadingCount = %d, want %d", got, tc.count)
			}
			if got := vvSectionLines(tc.text); !reflect.DeepEqual(got, tc.lines) {
				t.Errorf("vvSectionLines = %#v, want %#v", got, tc.lines)
			}
		})
	}
	for in, want := range map[string]string{
		" `a/**` ":        "a/**",
		"\t` my dir/*`\t": "my dir/*",
		"```":             "",
		"a`b":             "a`b",
		"\r`x`":           "\r`x",
	} {
		if got := trimGlobElement(in); got != want {
			t.Errorf("trimGlobElement(%q) = %q, want %q", in, got, want)
		}
	}
}
