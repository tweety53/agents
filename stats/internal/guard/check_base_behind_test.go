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
	g.copyShape(seed, repoShape{Branch: "main", Commit: "one"})
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
		lines[2] != "  git -C "+wt+" fetch -q origin && flow-guard sync-onto-base "+wt+" origin/feature/x "+mb {
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

// TestCheckBaseBehindRefusesAnEmptyC pins kan-1087 self-review F4: the push half of the
// printed sync line, run with an empty C, must abort rather than push an empty refspec
// that deletes the remote base.
func TestCheckBaseBehindRefusesAnEmptyC(t *testing.T) {
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	var g fxGit
	seed, origin, wt := root+"/seed", root+"/origin.git", root+"/wt"
	g.copyShape(seed, repoShape{Branch: "main", Commit: "one"})
	g.git(seed, "branch", "feature/x")
	g.git("", "clone", "-q", "--bare", seed, origin)
	g.git("", "clone", "-q", origin, wt)
	g.git(wt, "checkout", "-q", "-b", "spectre/c", "origin/feature/x")
	g.git(wt, "config", "branch.spectre/c.flowBase", "feature/x")
	g.git(seed, "commit", "-q", "--allow-empty", "-m", "two")
	g.git(seed, "push", "-q", origin, "main")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("check-base-behind", []string{wt}, Env{Getenv: os.Getenv})
	lines := strings.Split(strings.TrimRight(r.stdout, "\n"), "\n")
	if r.rc != 1 || len(lines) != 3 {
		t.Fatalf("base behind: rc %d stdout %q stderr %q", r.rc, r.stdout, r.err)
	}
	i := strings.LastIndex(lines[1], " && git -C ")
	if i < 0 {
		t.Fatalf("line 1 has no push command: %q", lines[1])
	}
	push := strings.TrimPrefix(lines[1][i:], " && ")
	shells := []string{"bash"}
	if _, err := exec.LookPath("zsh"); err == nil {
		shells = append(shells, "zsh")
	}
	for _, sh := range shells {
		cmd := exec.Command(sh, "-c", "C=; "+push)
		cmd.Env = append(os.Environ(), fixtureGitEnv...)
		if out, err := cmd.CombinedOutput(); err == nil {
			t.Fatalf("%s: empty-C push exited 0\n%s", sh, out)
		}
		if _, err := exec.Command("git", "-C", origin, "rev-parse", "--verify", "-q", "refs/heads/feature/x").Output(); err != nil {
			t.Fatalf("%s: empty-C push deleted feature/x from origin", sh)
		}
	}
	if !strings.Contains(lines[1], `"${C:?}:refs/heads/feature/x"`) {
		t.Fatalf("line 1 lacks the required-C refspec: %q", lines[1])
	}
}
