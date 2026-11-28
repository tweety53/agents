package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestPreExistingFailure pins the sweep's classifier: a failure is
// pre-existing only when no named path is in the change's diff and the
// command fails on the merge-base too. Fixture: main carries a.txt and a
// marker file; the change branch adds b.txt.
func TestPreExistingFailure(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	wt := root + "/wt"
	g.copyShape(wt, repoShape{Branch: "main", Commit: "one"})
	g.write(wt+"/a.txt", "a")
	g.write(wt+"/marker", "m")
	g.git(wt, "add", "-A")
	g.git(wt, "commit", "-q", "-m", "base")
	g.git(wt, "checkout", "-q", "-b", "change")
	g.write(wt+"/b.txt", "b")
	g.git(wt, "add", "-A")
	g.git(wt, "commit", "-q", "-m", "change")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := core.Env{Getenv: os.Getenv}
	run := func(args ...string) guardResult { return runGuard("pre-existing-failure", args, env) }

	if r := run(wt, "main", "true"); r.rc != 2 || r.stdout != "" {
		t.Fatalf("no path: rc %d stdout %q", r.rc, r.stdout)
	}
	if r := run(t.TempDir(), "main", "true", "a.txt"); r.rc != 2 || r.stdout != "" {
		t.Fatalf("no repository: rc %d stdout %q", r.rc, r.stdout)
	}
	if r := run(wt, "main", "false", "./b.txt"); r.rc != 1 || r.stdout != "CAUSED: b.txt is in the change's diff\n" {
		t.Fatalf("touched path: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	mb := g.git(wt, "rev-parse", "main")
	if r := run(wt, "main", "test ! -e b.txt", "a.txt"); r.rc != 1 || r.stdout != "CAUSED: passes on the merge-base "+mb+"\n" {
		t.Fatalf("passes on base: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	r := run(wt, "main", "test ! -e marker", "a.txt")
	if r.rc != 0 || r.stdout != "PRE-EXISTING: fails on the merge-base "+mb+" (exit 1) and no named path is in the change's diff\n" {
		t.Fatalf("pre-existing: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	if list := g.git(wt, "worktree", "list", "--porcelain"); strings.Count(list, "worktree ") != 1 {
		t.Fatalf("merge-base worktree left behind:\n%s", list)
	}

	// A file renamed out of a named path still counts as in the diff.
	g.git(wt, "config", "diff.renames", "true")
	g.git(wt, "mv", "a.txt", "moved.txt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	if r := run(wt, "main", "false", "a.txt"); r.rc != 1 || r.stdout != "CAUSED: a.txt is in the change's diff\n" {
		t.Fatalf("renamed path: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
}
