package guard

import (
	"os"
	"testing"
	"unicode"
)

// TestPyIsSpace: pyIsSpace is exactly Python's str.isspace() — the set
// `[c for c in map(chr, range(0x110000)) if c.isspace()]` under CPython 3.14.
func TestPyIsSpace(t *testing.T) {
	want := map[rune]bool{'\u0085': true, '\u00a0': true, '\u1680': true, '\u2028': true, '\u2029': true,
		'\u202f': true, '\u205f': true, '\u3000': true}
	for r := rune('\t'); r <= '\r'; r++ {
		want[r] = true
	}
	for r := rune('\x1c'); r <= ' '; r++ {
		want[r] = true
	}
	for r := rune('\u2000'); r <= '\u200a'; r++ {
		want[r] = true
	}
	for r := rune(0); r <= unicode.MaxRune; r++ {
		if pyIsSpace(r) != want[r] {
			t.Errorf("pyIsSpace(%U) = %v, Python's isspace() %v", r, pyIsSpace(r), want[r])
		}
	}
}

// TestEabReadPickSplitsOnPythonSpace: the pick line is cut as Python's split(maxsplit=3), so an
// ideographic space separates fields as an ASCII one does.
func TestEabReadPickSplitsOnPythonSpace(t *testing.T) {
	p := t.TempDir() + "/pick"
	if err := os.WriteFile(p, []byte("opus flow-low src　why　\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := eabReadPick(p); got != "src: why" {
		t.Errorf("eabReadPick = %q, want %q", got, "src: why")
	}
}
