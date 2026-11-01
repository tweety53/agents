package lessons

import (
	"os"
	"path/filepath"
	"testing"
)

// writeRepo lays down one project root's briefs home with the named files.
func writeRepo(t *testing.T, briefs map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range briefs {
		path := filepath.Join(root, briefsDir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir briefs: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write brief: %v", err)
		}
	}
	return root
}

func TestResolveReadsTheBrief(t *testing.T) {
	root := writeRepo(t,
		map[string]string{"flake-bisection.md": "# Flake bisection brief\n\nStart from the smallest reproducing subset.\n"})
	res, err := Resolve([]Root{{Path: root}}, "flake-bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 1 {
		t.Fatalf("briefs = %d, want 1", len(res.Briefs))
	}
	if res.Briefs[0].Slug != "flake-bisection" {
		t.Errorf("slug = %q", res.Briefs[0].Slug)
	}
	if res.Briefs[0].Title != "Flake bisection brief" {
		t.Errorf("title = %q", res.Briefs[0].Title)
	}
}

func TestResolveMatchesAcrossFolding(t *testing.T) {
	root := writeRepo(t,
		map[string]string{"flake_bisection.md": "# Flake Bisection\n\nbody\n"})
	res, err := Resolve([]Root{{Path: root}}, "Flake-Bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 1 {
		t.Fatalf("briefs = %d, want the folded match", len(res.Briefs))
	}
}

func TestResolveAbsentDirsAreNotAnError(t *testing.T) {
	root := t.TempDir() // neither docs/briefs/ nor spectre/ exists
	res, err := Resolve([]Root{{Path: root}}, "anything")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 0 {
		t.Fatalf("got %d briefs; want none", len(res.Briefs))
	}
	if len(res.Unreadable) != 0 {
		t.Fatalf("unreadable = %v; an empty root is absent, not broken", res.Unreadable)
	}
}

func TestResolveUnreadableRootIsReported(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "gone")
	res, err := Resolve([]Root{{Path: missing}}, "topic")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Unreadable) != 1 || res.Unreadable[0] != missing {
		t.Fatalf("unreadable = %v, want [%s]", res.Unreadable, missing)
	}
}

func TestResolveEmptyTopicIsRefused(t *testing.T) {
	if _, err := Resolve(nil, "  -_ "); err == nil {
		t.Fatal("Resolve accepted a topic empty after normalization")
	}
}

// TestResolveUnreadableFileIsReported pins the per-file rule: one brief that
// cannot be read inside a readable root is an environmental
// failure reported in Unreadable, never silence a caller could mistake for
// "no lesson exists".
func TestResolveUnreadableFileIsReported(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root reads a 0o000 file; the permission denial is impossible")
	}
	root := writeRepo(t,
		map[string]string{"flake-bisection.md": "# Flake bisection brief\n\nStart small.\n"})
	briefPath := filepath.Join(root, briefsDir, "flake-bisection.md")
	for _, path := range []string{briefPath} {
		if err := os.Chmod(path, 0o000); err != nil {
			t.Fatalf("chmod %s: %v", path, err)
		}
	}

	res, err := Resolve([]Root{{Path: root}}, "flake bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 0 {
		t.Fatalf("got %d briefs; an unreadable file is not a match", len(res.Briefs))
	}
	for _, path := range []string{briefPath} {
		if !containsPath(res.Unreadable, path) {
			t.Errorf("unreadable = %v, want %s reported", res.Unreadable, path)
		}
	}
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}
