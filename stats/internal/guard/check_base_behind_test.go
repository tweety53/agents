package guard

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckBaseBehind pins scripts/check-base-behind.sh's lines (kan-1007: a
// feature base behind main was papered over by merging main into the change
// branch). The fixture: a bare origin on `main` with `feature/x` at its first
// commit, a clone (origin/HEAD = main) on `spectre/c` at origin/feature/x.
func TestCheckBaseBehind(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	seed, origin, wt := root+"/seed", root+"/origin.git", root+"/wt"
	g.git("", "init", "-q", "-b", "main", seed)
	g.git(seed, "commit", "-q", "--allow-empty", "-m", "one")
	g.git(seed, "branch", "feature/x")
	g.git("", "clone", "-q", "--bare", seed, origin)
	g.git("", "clone", "-q", origin, wt)
	g.git(wt, "checkout", "-q", "-b", "spectre/c", "origin/feature/x")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Getenv: os.Getenv}
	run := func(args ...string) guardResult { return runGuard("check-base-behind", args, env) }

	if r := run(); r.rc != 2 || r.stdout != "" {
		t.Fatalf("(a) no argument: rc %d stdout %q", r.rc, r.stdout)
	}
	if r := run(t.TempDir()); r.rc != 2 || r.stdout != "" {
		t.Fatalf("(b) no repository: rc %d stdout %q", r.rc, r.stdout)
	}
	if r := run(wt); r.rc != 0 || r.stdout != "BASE-CURRENT: "+wt+" — no recorded base\n" {
		t.Fatalf("(c) no recorded base: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	g.git(wt, "config", "branch.spectre/c.flowBase", "main")
	if r := run(wt); r.rc != 0 || r.stdout != "BASE-CURRENT: "+wt+" — main is the tracked branch\n" {
		t.Fatalf("(d) base is tracked: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	g.git(wt, "config", "branch.spectre/c.flowBase", "feature/x")
	g.git(seed, "commit", "-q", "--allow-empty", "-m", "two")
	g.git(seed, "push", "-q", origin, "main")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := run(wt)
	lines := strings.Split(strings.TrimRight(r.stdout, "\n"), "\n")
	mb := g.git(wt, "merge-base", "HEAD", "origin/feature/x")
	if r.rc != 1 || len(lines) != 3 ||
		lines[0] != "BASE-BEHIND: "+wt+" — origin/feature/x lacks 1 commits of origin/main; sync the base, never merge main into the change branch (kan-1007):" ||
		!strings.HasPrefix(lines[1], "  T=$(git -C "+wt+" merge-tree --write-tree origin/feature/x origin/main)") ||
		lines[2] != "  git -C "+wt+" fetch -q origin && sync-onto-base.sh "+wt+" origin/feature/x "+mb {
		t.Fatalf("(e) base behind: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	sync := exec.Command("bash", "-c", strings.TrimPrefix(lines[1], "  "))
	sync.Env = append(append(os.Environ(), fixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := sync.CombinedOutput(); err != nil {
		t.Fatalf("(f) sync command: %v\n%s", err, out)
	}
	if r := run(wt); r.rc != 0 || r.stdout != "BASE-CURRENT: "+wt+" — origin/feature/x holds every commit of origin/main\n" {
		t.Fatalf("(f) synced: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
}
