package repotest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestGoSourcesHashSeesGoEdits pins the kan-1053 self-review fix, F1: the
// UI-test stack's `fingerprint` row could not see a Go or seed change, so a
// stack built before a Go-side fix read fresh. The behaviour half runs
// stats/Makefile's go-sources-hash target in a scratch git repository; the
// wiring half asserts, on source text, that ui-test-up records the hash only
// after the seed, ui-test-down removes it, and the fingerprint row compares it.
// smoke: core — the UI-test stack's fingerprint sees a Go edit
func TestGoSourcesHashSeesGoEdits(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	mk := filepath.Join(root, "stats", "Makefile")
	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	appendTo := func(name, line string) {
		t.Helper()
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if _, err := f.WriteString(line); err != nil {
			t.Fatal(err)
		}
	}
	rm := func(name string) {
		t.Helper()
		if err := os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Fatal(err)
		}
	}
	sh(t, dir, nil, "git", "init", "-q")
	write(".gitignore", "ign/\n")
	write("a.go", "package a\n")
	write("go.mod", "module a\n")
	write("go.sum", "x v1 h1:x\n")
	sh(t, dir, nil, "git", "add", ".")
	sh(t, dir, nil, "git", "-c", "user.email=t@t", "-c", "user.name=t", "commit", "-qm", "init")

	hash := func() string {
		t.Helper()
		cmd := exec.Command("make", "-s", "--no-print-directory", "-f", mk, "-C", dir, "go-sources-hash")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("make go-sources-hash: %v\n%s", err, out)
		}
		return strings.TrimSpace(string(out))
	}
	base := hash()
	if base == "" {
		t.Fatal("go-sources-hash printed nothing")
	}

	appendTo("a.go", "// x\n")
	if hash() == base {
		t.Error("an uncommitted edit of a tracked .go left the hash unchanged")
	}
	sh(t, dir, nil, "git", "checkout", "--", "a.go")
	if got := hash(); got != base {
		t.Errorf("reverting the edit: hash %q, want the base %q", got, base)
	}
	write("b.go", "package a\n")
	if hash() == base {
		t.Error("an untracked .go left the hash unchanged")
	}
	rm("b.go")
	appendTo("go.sum", "y v1 h1:y\n")
	if hash() == base {
		t.Error("a go.sum edit left the hash unchanged")
	}
	sh(t, dir, nil, "git", "checkout", "--", "go.sum")
	write("ign/c.go", "package c\n")
	write("w.ts", "export {}\n")
	if got := hash(); got != base {
		t.Errorf("a gitignored .go and a .ts moved the hash: %q, want %q", got, base)
	}

	src, err := os.ReadFile(mk)
	if err != nil {
		t.Fatal(err)
	}
	up, down := recipe(string(src), "ui-test-up:"), recipe(string(src), "ui-test-down:")
	rec := strings.Index(up, "go-sources-hash > $(UITEST_GO_HASH).new")
	seed := strings.Index(up, "go run ./cmd/uitest-seed")
	pub := strings.Index(up, "mv -f $(UITEST_GO_HASH).new $(UITEST_GO_HASH)")
	if rec < 0 || seed < 0 || pub < 0 || pub < seed || rec > seed {
		t.Errorf("ui-test-up must record the hash into $(UITEST_GO_HASH).new before the build and mv it into place after the seed (record %d, seed %d, publish %d)", rec, seed, pub)
	}
	if !strings.Contains(down, "$(UITEST_GO_HASH)") {
		t.Error("ui-test-down does not remove $(UITEST_GO_HASH)")
	}

	proj, err := os.ReadFile(filepath.Join(root, ".flow", "project.md"))
	if err != nil {
		t.Fatal(err)
	}
	var row string
	for _, l := range strings.Split(string(proj), "\n") {
		if strings.HasPrefix(l, "| `fingerprint` | ") {
			row = l
		}
	}
	if !strings.Contains(row, `go-sources-hash \| cmp -s - /tmp/flow-uitest-go.sha256`) {
		t.Errorf("the fingerprint row does not compare the recorded Go hash: %q", row)
	}
}

// recipe is the rule starting at the line that begins with target, up to the
// next blank line.
func recipe(makefile, target string) string {
	i := strings.Index(makefile, "\n"+target)
	if i < 0 {
		return ""
	}
	rest := makefile[i+1:]
	if j := strings.Index(rest, "\n\n"); j >= 0 {
		rest = rest[:j]
	}
	return rest
}
