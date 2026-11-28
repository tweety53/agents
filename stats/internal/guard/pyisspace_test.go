package guard

import (
	"testing"
	"unicode"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestPyIsSpace: PyIsSpace is exactly Python's str.isspace() — the set
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
		if core.PyIsSpace(r) != want[r] {
			t.Errorf("pyIsSpace(%U) = %v, Python's isspace() %v", r, core.PyIsSpace(r), want[r])
		}
	}
}
