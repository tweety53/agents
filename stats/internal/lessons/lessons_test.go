package lessons

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeRepo lays down one project root's tree: a briefs home with the
// named files and one archived narrative per change.
func writeRepo(t *testing.T, briefs map[string]string, narratives map[string]string) string {
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
	for change, content := range narratives {
		path := filepath.Join(root, "spectre/changes/archive", change, "narrative.md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir archive: %v", err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write narrative: %v", err)
		}
	}
	return root
}

func TestResolveBriefOutranksNarrative(t *testing.T) {
	root := writeRepo(t,
		map[string]string{"flake-bisection.md": "# Flake bisection brief\n\nStart from the smallest reproducing subset.\n"},
		map[string]string{"kan-527-fix-the-flake": "# kan-527 — session narrative\n\n- re-derived the flake-bisection brief from scratch\n"})
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
	if len(res.Mentions) != 1 {
		t.Fatalf("mentions = %d, want 1", len(res.Mentions))
	}
	if res.Mentions[0].Change != "kan-527-fix-the-flake" {
		t.Errorf("change = %q", res.Mentions[0].Change)
	}
}

func TestResolveMatchesAcrossFolding(t *testing.T) {
	root := writeRepo(t,
		map[string]string{"flake_bisection.md": "# Flake Bisection\n\nbody\n"},
		nil)
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
	if len(res.Briefs) != 0 || len(res.Mentions) != 0 {
		t.Fatalf("got %d briefs, %d mentions; want none", len(res.Briefs), len(res.Mentions))
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

func TestResolveMentionLinesAreCappedAndNewestFirst(t *testing.T) {
	line := strings.Repeat("- flake bisection mention ", 1) + "%d"
	old := writeRepo(t, nil, map[string]string{"kan-100-old": strings.Repeat(line+"\n", matchLineCap+3)})
	oldPath := filepath.Join(old, "spectre/changes/archive/kan-100-old/narrative.md")
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(oldPath, past, past); err != nil {
		t.Fatalf("Chtimes old: %v", err)
	}
	new := writeRepo(t, nil, map[string]string{"kan-200-new": strings.Repeat(line+"\n", 1)})
	newPath := filepath.Join(new, "spectre/changes/archive/kan-200-new/narrative.md")
	if err := os.Chtimes(newPath, time.Now(), time.Now()); err != nil {
		t.Fatalf("Chtimes new: %v", err)
	}

	res, err := Resolve([]Root{{Path: old}, {Path: new}}, "flake bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Mentions) != 2 {
		t.Fatalf("mentions = %d, want 2", len(res.Mentions))
	}
	if res.Mentions[0].Change != "kan-200-new" {
		t.Errorf("first mention = %q, want the newest", res.Mentions[0].Change)
	}
	if got := len(res.Mentions[1].Lines); got != matchLineCap {
		t.Errorf("capped lines = %d, want %d", got, matchLineCap)
	}
}

func TestResolveNonMatchingNarrativeIsAbsent(t *testing.T) {
	root := writeRepo(t, nil, map[string]string{"kan-300": "# narrative\n\n- unrelated work only\n"})
	res, err := Resolve([]Root{{Path: root}}, "flake bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Mentions) != 0 {
		t.Fatalf("mentions = %d, want 0", len(res.Mentions))
	}
}

// TestResolveUnreadableFileIsReported pins the per-file rule: one brief or
// narrative that cannot be read inside a readable root is an environmental
// failure reported in Unreadable, never silence a caller could mistake for
// "no lesson exists".
func TestResolveUnreadableFileIsReported(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root reads a 0o000 file; the permission denial is impossible")
	}
	root := writeRepo(t,
		map[string]string{"flake-bisection.md": "# Flake bisection brief\n\nStart small.\n"},
		map[string]string{"kan-100": "- flake bisection mention\n"})
	briefPath := filepath.Join(root, briefsDir, "flake-bisection.md")
	narrPath := filepath.Join(root, "spectre/changes/archive/kan-100/narrative.md")
	for _, path := range []string{briefPath, narrPath} {
		if err := os.Chmod(path, 0o000); err != nil {
			t.Fatalf("chmod %s: %v", path, err)
		}
	}

	res, err := Resolve([]Root{{Path: root}}, "flake bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 0 || len(res.Mentions) != 0 {
		t.Fatalf("got %d briefs, %d mentions; an unreadable file is not a match", len(res.Briefs), len(res.Mentions))
	}
	for _, path := range []string{briefPath, narrPath} {
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
