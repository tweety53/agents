package lessons

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestResolveIgnoresANarrativeInHistory pins that git history is not a
// lesson source (KAN-1073): a narrative that lives only in a drop commit's
// parent, as integrate's run 1 used to leave it, is never a mention.
func TestResolveIgnoresANarrativeInHistory(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "spectre/changes/kan-527/narrative.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("- flake bisection lesson\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "spectre"},
		{"commit", "-q", "-m", "plan kan-527"},
		{"rm", "-r", "-q", "spectre/changes/kan-527"},
		{"commit", "-q", "-m", "chore(spectre): drop kan-527"},
	} {
		cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@t", "-c", "commit.gpgsign=false"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	res, err := Resolve([]Root{{Path: root}}, "flake bisection")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if len(res.Briefs) != 0 || len(res.Unreadable) != 0 {
		t.Fatalf("briefs = %+v, unreadable = %v; history is not a source", res.Briefs, res.Unreadable)
	}
}
