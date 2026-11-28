package gitops

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// sbbFixture is TestCheckBaseBehind's shape with the base behind: a bare origin on main, feature/x
// at main's first commit, a clone on spectre/c at origin/feature/x with that base recorded, and main
// one commit ahead.
func sbbFixture(t *testing.T) (g *guardtest.FxGit, seed, origin, wt string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g = &guardtest.FxGit{}
	seed, origin, wt = root+"/seed", root+"/origin.git", root+"/wt"
	g.CopyShape(seed, guardtest.RepoShape{Branch: "main", Commit: "one"})
	g.Git(seed, "branch", "feature/x")
	g.Git("", "clone", "-q", "--bare", seed, origin)
	g.Git("", "clone", "-q", origin, wt)
	g.Git(wt, "checkout", "-q", "-b", "spectre/c", "origin/feature/x")
	g.Git(wt, "config", "branch.spectre/c.flowBase", "feature/x")
	g.Git(wt, "config", "user.name", "test")
	g.Git(wt, "config", "user.email", "test@example.com")
	g.Git(seed, "commit", "-q", "--allow-empty", "-m", "two")
	g.Git(seed, "push", "-q", origin, "main")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	return g, seed, origin, wt
}

// TestSyncBaseBehindSyncsTheBase (KAN-1156): the guard alone does what check-base-behind printed two
// commands for — the base gains the tracked branch's commit on origin, the change branch is rebased
// onto it, and check-base-behind then reports the base current. No raw git push is run by the caller.
func TestSyncBaseBehindSyncsTheBase(t *testing.T) {
	t.Parallel()
	g, _, origin, wt := sbbFixture(t)
	env := core.Env{Getenv: os.Getenv}
	if r := runGuard("check-base-behind", []string{wt}, env); r.Rc != 1 {
		t.Fatalf("fixture is not behind: rc %d %q", r.Rc, r.Stdout)
	}
	r := runGuard("sync-base-behind", []string{wt}, env)
	if r.Rc != 0 || !strings.Contains(r.Stdout, "BASE-SYNCED: "+wt+" — feature/x now holds origin/main") {
		t.Fatalf("rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	if r := runGuard("check-base-behind", []string{wt}, env); r.Rc != 0 {
		t.Errorf("still behind after the sync: rc %d %q", r.Rc, r.Stdout)
	}
	if log := g.Git(origin, "log", "--format=%s", "-1", "feature/x"); log != "Merge main into feature/x" {
		t.Errorf("origin/feature/x tip %q", log)
	}
	if g.Git(wt, "merge-base", "--is-ancestor", "origin/feature/x", "HEAD"); g.Err != nil {
		t.Errorf("the change branch was not rebased onto the synced base: %v", g.Err)
	}
	// Nothing behind: a no-op, exit 0.
	if r := runGuard("sync-base-behind", []string{wt}, env); r.Rc != 0 || strings.Contains(r.Stdout, "BASE-SYNCED") {
		t.Errorf("second run: rc %d %q", r.Rc, r.Stdout)
	}
	if r := runGuard("sync-base-behind", nil, env); r.Rc != 2 {
		t.Errorf("no argument: rc %d", r.Rc)
	}
}

// TestSyncBaseBehindStopsOnAMergeConflict: base and tracked branch edit one line differently — exit 1,
// BASE-CONFLICT, and origin's base untouched.
func TestSyncBaseBehindStopsOnAMergeConflict(t *testing.T) {
	t.Parallel()
	g, seed, origin, wt := sbbFixture(t)
	g.Git(seed, "checkout", "-q", "feature/x")
	os.WriteFile(seed+"/f.txt", []byte("base\n"), 0o644)
	g.Git(seed, "add", "f.txt")
	g.Git(seed, "commit", "-q", "-m", "base edit")
	g.Git(seed, "push", "-q", origin, "feature/x")
	g.Git(seed, "checkout", "-q", "main")
	os.WriteFile(seed+"/f.txt", []byte("main\n"), 0o644)
	g.Git(seed, "add", "f.txt")
	g.Git(seed, "commit", "-q", "-m", "main edit")
	g.Git(seed, "push", "-q", origin, "main")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	before := g.Git(origin, "rev-parse", "feature/x")
	r := runGuard("sync-base-behind", []string{wt}, core.Env{Getenv: os.Getenv})
	if r.Rc != 1 || !strings.Contains(r.Stdout, "BASE-CONFLICT") || g.Git(origin, "rev-parse", "feature/x") != before {
		t.Fatalf("rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
}
