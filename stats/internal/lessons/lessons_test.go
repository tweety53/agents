package lessons

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeRepo lays down one project root's tree: a briefs home with the
// named files and one landed narrative per change, in git history.
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
		landNarrative(t, root, change, content, time.Now())
	}
	return root
}

// landNarrative commits change's narrative.md into root's history and drops
// it from the tip, as integrate's run 1 does, at commit time when.
func landNarrative(t *testing.T, root, change, content string, when time.Time) {
	t.Helper()
	path := filepath.Join(root, "spectre/changes", change, "narrative.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir change: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write narrative: %v", err)
	}
	date := when.Format(time.RFC3339)
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "spectre"},
		{"commit", "-q", "-m", "plan " + change},
		{"rm", "-r", "-q", "spectre/changes/" + change},
		{"commit", "-q", "-m", "chore(spectre): drop " + change},
	} {
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+date, "GIT_AUTHOR_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
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
	old := t.TempDir()
	landNarrative(t, old, "kan-100-old", strings.Repeat(line+"\n", matchLineCap+3), time.Now().Add(-time.Hour))
	new := writeRepo(t, nil, map[string]string{"kan-200-new": strings.Repeat(line+"\n", 1)})

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

// TestResolveUnreadableFileIsReported pins the per-file rule: one brief that
// cannot be read inside a readable root is an environmental
// failure reported in Unreadable, never silence a caller could mistake for
// "no lesson exists".
func TestResolveUnreadableFileIsReported(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root reads a 0o000 file; the permission denial is impossible")
	}
	root := writeRepo(t,
		map[string]string{"flake-bisection.md": "# Flake bisection brief\n\nStart small.\n"},
		nil)
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
	if len(res.Briefs) != 0 || len(res.Mentions) != 0 {
		t.Fatalf("got %d briefs, %d mentions; an unreadable file is not a match", len(res.Briefs), len(res.Mentions))
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
