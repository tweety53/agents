package guard

import (
	"os"
	"path/filepath"
	"testing"
	"unicode"

	"github.com/tweety53/agents/stats/internal/guard/core"
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

// TestBucRealpath: a missing tail is kept below its resolved parent, a dangling symlink resolves
// to its target, and `..` after a symlink is taken from the link's target, as the kernel does.
func TestBucRealpath(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"real/sub", "other"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for link, target := range map[string]string{"link": "real/sub", "dangling": "other/new.txt"} {
		if err := os.Symlink(target, filepath.Join(root, link)); err != nil {
			t.Fatal(err)
		}
	}
	b := &buc{env: core.Env{Dir: root}}
	for in, want := range map[string]string{
		"link/missing/f.txt": root + "/real/sub/missing/f.txt",
		"dangling":           root + "/other/new.txt",
		"link/../x":          root + "/real/x",
		root + "/link/./y/":  root + "/real/sub/y",
	} {
		if got := b.realpath(in); got != want {
			t.Errorf("realpath(%q) = %q, want %q", in, got, want)
		}
	}
}
