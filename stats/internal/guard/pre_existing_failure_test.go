package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
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
	var g guardtest.FxGit
	wt := root + "/wt"
	g.CopyShape(wt, guardtest.RepoShape{Branch: "main", Commit: "one"})
	g.Write(wt+"/a.txt", "a")
	g.Write(wt+"/marker", "m")
	g.Git(wt, "add", "-A")
	g.Git(wt, "commit", "-q", "-m", "base")
	g.Git(wt, "checkout", "-q", "-b", "change")
	g.Write(wt+"/b.txt", "b")
	g.Git(wt, "add", "-A")
	g.Git(wt, "commit", "-q", "-m", "change")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	env := core.Env{Getenv: os.Getenv}
	run := func(args ...string) guardtest.GuardResult { return runGuard("pre-existing-failure", args, env) }

	if r := run(wt, "main", "true"); r.Rc != 2 || r.Stdout != "" {
		t.Fatalf("no path: rc %d stdout %q", r.Rc, r.Stdout)
	}
	if r := run(t.TempDir(), "main", "true", "a.txt"); r.Rc != 2 || r.Stdout != "" {
		t.Fatalf("no repository: rc %d stdout %q", r.Rc, r.Stdout)
	}
	if r := run(wt, "main", "false", "./b.txt"); r.Rc != 1 || r.Stdout != "CAUSED: b.txt is in the change's diff\n" {
		t.Fatalf("touched path: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	mb := g.Git(wt, "rev-parse", "main")
	if r := run(wt, "main", "test ! -e b.txt", "a.txt"); r.Rc != 1 || r.Stdout != "CAUSED: passes on the merge-base "+mb+"\n" {
		t.Fatalf("passes on base: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	r := run(wt, "main", "test ! -e marker", "a.txt")
	if r.Rc != 0 || r.Stdout != "PRE-EXISTING: fails on the merge-base "+mb+" (exit 1) and no named path is in the change's diff\n" {
		t.Fatalf("pre-existing: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	if list := g.Git(wt, "worktree", "list", "--porcelain"); strings.Count(list, "worktree ") != 1 {
		t.Fatalf("merge-base worktree left behind:\n%s", list)
	}

	// A file renamed out of a named path still counts as in the diff.
	g.Git(wt, "config", "diff.renames", "true")
	g.Git(wt, "mv", "a.txt", "moved.txt")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	if r := run(wt, "main", "false", "a.txt"); r.Rc != 1 || r.Stdout != "CAUSED: a.txt is in the change's diff\n" {
		t.Fatalf("renamed path: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
}
