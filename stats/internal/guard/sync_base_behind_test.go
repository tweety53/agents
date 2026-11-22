package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sbbFixture is TestCheckBaseBehind's shape with the base behind: a bare origin on main, feature/x
// at main's first commit, a clone on spectre/c at origin/feature/x with that base recorded, and main
// one commit ahead.
func sbbFixture(t *testing.T) (g *fxGit, seed, origin, wt string) {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	g = &fxGit{}
	seed, origin, wt = root+"/seed", root+"/origin.git", root+"/wt"
	g.copyShape(seed, repoShape{Branch: "main", Commit: "one"})
	g.git(seed, "branch", "feature/x")
	g.git("", "clone", "-q", "--bare", seed, origin)
	g.git("", "clone", "-q", origin, wt)
	g.git(wt, "checkout", "-q", "-b", "spectre/c", "origin/feature/x")
	g.git(wt, "config", "branch.spectre/c.flowBase", "feature/x")
	g.git(wt, "config", "user.name", "test")
	g.git(wt, "config", "user.email", "test@example.com")
	g.git(seed, "commit", "-q", "--allow-empty", "-m", "two")
	g.git(seed, "push", "-q", origin, "main")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return g, seed, origin, wt
}

// TestSyncBaseBehindSyncsTheBase (KAN-1156): the guard alone does what check-base-behind printed two
// commands for — the base gains the tracked branch's commit on origin, the change branch is rebased
// onto it, and check-base-behind then reports the base current. No raw git push is run by the caller.
func TestSyncBaseBehindSyncsTheBase(t *testing.T) {
	t.Parallel()
	g, _, origin, wt := sbbFixture(t)
	env := Env{Getenv: os.Getenv}
	if r := runGuard("check-base-behind", []string{wt}, env); r.rc != 1 {
		t.Fatalf("fixture is not behind: rc %d %q", r.rc, r.stdout)
	}
	r := runGuard("sync-base-behind", []string{wt}, env)
	if r.rc != 0 || !strings.Contains(r.stdout, "BASE-SYNCED: "+wt+" — feature/x now holds origin/main") {
		t.Fatalf("rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	if r := runGuard("check-base-behind", []string{wt}, env); r.rc != 0 {
		t.Errorf("still behind after the sync: rc %d %q", r.rc, r.stdout)
	}
	if log := g.git(origin, "log", "--format=%s", "-1", "feature/x"); log != "Merge main into feature/x" {
		t.Errorf("origin/feature/x tip %q", log)
	}
	if g.git(wt, "merge-base", "--is-ancestor", "origin/feature/x", "HEAD"); g.err != nil {
		t.Errorf("the change branch was not rebased onto the synced base: %v", g.err)
	}
	// Nothing behind: a no-op, exit 0.
	if r := runGuard("sync-base-behind", []string{wt}, env); r.rc != 0 || strings.Contains(r.stdout, "BASE-SYNCED") {
		t.Errorf("second run: rc %d %q", r.rc, r.stdout)
	}
	if r := runGuard("sync-base-behind", nil, env); r.rc != 2 {
		t.Errorf("no argument: rc %d", r.rc)
	}
}

// TestSyncBaseBehindStopsOnAMergeConflict: base and tracked branch edit one line differently — exit 1,
// BASE-CONFLICT, and origin's base untouched.
func TestSyncBaseBehindStopsOnAMergeConflict(t *testing.T) {
	t.Parallel()
	g, seed, origin, wt := sbbFixture(t)
	g.git(seed, "checkout", "-q", "feature/x")
	os.WriteFile(seed+"/f.txt", []byte("base\n"), 0o644)
	g.git(seed, "add", "f.txt")
	g.git(seed, "commit", "-q", "-m", "base edit")
	g.git(seed, "push", "-q", origin, "feature/x")
	g.git(seed, "checkout", "-q", "main")
	os.WriteFile(seed+"/f.txt", []byte("main\n"), 0o644)
	g.git(seed, "add", "f.txt")
	g.git(seed, "commit", "-q", "-m", "main edit")
	g.git(seed, "push", "-q", origin, "main")
	if g.err != nil {
		t.Fatal(g.err)
	}
	before := g.git(origin, "rev-parse", "feature/x")
	r := runGuard("sync-base-behind", []string{wt}, Env{Getenv: os.Getenv})
	if r.rc != 1 || !strings.Contains(r.stdout, "BASE-CONFLICT") || g.git(origin, "rev-parse", "feature/x") != before {
		t.Fatalf("rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
}
