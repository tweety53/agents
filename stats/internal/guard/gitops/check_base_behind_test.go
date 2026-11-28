package gitops

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// TestCheckBaseBehind pins scripts/check-base-behind.sh's lines (kan-1007: a
// feature base behind main was papered over by merging main into the change
// branch). The fixture: a bare origin on `main` with `feature/x` at its first
// commit, a clone (origin/HEAD = main) on `spectre/c` at origin/feature/x.
func TestCheckBaseBehind(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g guardtest.FxGit
	seed, origin, wt := root+"/seed", root+"/origin.git", root+"/wt"
	g.CopyShape(seed, guardtest.RepoShape{Branch: "main", Commit: "one"})
	g.Git(seed, "branch", "feature/x")
	g.Git("", "clone", "-q", "--bare", seed, origin)
	g.Git("", "clone", "-q", origin, wt)
	g.Git(wt, "checkout", "-q", "-b", "spectre/c", "origin/feature/x")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	env := core.Env{Getenv: os.Getenv}
	run := func(args ...string) guardtest.GuardResult { return runGuard("check-base-behind", args, env) }

	if r := run(); r.Rc != 2 || r.Stdout != "" {
		t.Fatalf("(a) no argument: rc %d stdout %q", r.Rc, r.Stdout)
	}
	if r := run(t.TempDir()); r.Rc != 2 || r.Stdout != "" {
		t.Fatalf("(b) no repository: rc %d stdout %q", r.Rc, r.Stdout)
	}
	if r := run(wt); r.Rc != 0 || r.Stdout != "BASE-CURRENT: "+wt+" — no recorded base\n" {
		t.Fatalf("(c) no recorded base: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	g.Git(wt, "config", "branch.spectre/c.flowBase", "main")
	if r := run(wt); r.Rc != 0 || r.Stdout != "BASE-CURRENT: "+wt+" — main is the tracked branch\n" {
		t.Fatalf("(d) base is tracked: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	g.Git(wt, "config", "branch.spectre/c.flowBase", "feature/x")
	g.Git(seed, "commit", "-q", "--allow-empty", "-m", "two")
	g.Git(seed, "push", "-q", origin, "main")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	setFetchHead(t, wt+"/.git/FETCH_HEAD", true, 2*time.Minute) // (d)'s fetch is no longer fresh
	r := run(wt)
	lines := strings.Split(strings.TrimRight(r.Stdout, "\n"), "\n")
	if r.Rc != 1 || len(lines) != 2 ||
		lines[0] != "BASE-BEHIND: "+wt+" — origin/feature/x lacks 1 commits of origin/main; sync the base, never merge main into the change branch (kan-1007):" ||
		lines[1] != "  flow-guard sync-base-behind "+wt {
		t.Fatalf("(e) base behind: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	g.Git(wt, "config", "user.name", "test")
	g.Git(wt, "config", "user.email", "test@example.com")
	if r := runGuard("sync-base-behind", []string{wt}, env); r.Rc != 0 {
		t.Fatalf("(f) sync: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	if r := run(wt); r.Rc != 0 || r.Stdout != "BASE-CURRENT: "+wt+" — origin/feature/x holds every commit of origin/main\n" {
		t.Fatalf("(f) synced: rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
}

// TestCheckBaseBehindFetchesOncePerRepository: two worktrees of one repository share an origin,
// so the guard fetches it once, never once per worktree.
func TestCheckBaseBehindFetchesOncePerRepository(t *testing.T) {
	t.Parallel()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g guardtest.FxGit
	seed, origin, a, b := root+"/seed", root+"/origin.git", root+"/a", root+"/b"
	g.CopyShape(seed, guardtest.RepoShape{Branch: "main", Commit: "one"})
	g.Git("", "clone", "-q", "--bare", seed, origin)
	g.Git("", "clone", "-q", origin, a)
	g.Git(a, "worktree", "add", "-q", "-b", "c1", b)
	g.Git(a, "config", "branch.c1.flowBase", "main")
	g.Git(a, "checkout", "-q", "-b", "c2")
	g.Git(a, "config", "branch.c2.flowBase", "main")
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	real, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	bin, log := root+"/bin", root+"/fetches"
	if err := os.Mkdir(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	wrapper := "#!/bin/sh\ncase \" $* \" in *\" fetch \"*) echo f >> " + log + " ;; esac\nexec " + real + " \"$@\"\n"
	if err := os.WriteFile(bin+"/git", []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	env := core.Env{Getenv: func(k string) string {
		if k == "PATH" {
			return bin + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}}
	if r := runGuard("check-base-behind", []string{a, b}, env); r.Rc != 0 {
		t.Fatalf("rc %d stdout %q stderr %q", r.Rc, r.Stdout, r.Err)
	}
	if got, _ := os.ReadFile(log); strings.Count(string(got), "f") != 1 {
		t.Fatalf("fetches = %q, want exactly one", got)
	}
}
