package guard

import (
	"os"
	"testing"
)

// Every case of scripts/test-check-planning-commit-location.sh at c4f26c84, one subtest
// per run label, each pinning the guard's stdout whole where the harness
// globbed it: 0 with the OK line in a linked worktree on spectre/<name>; 1
// with the main-checkout line in a main checkout (on any branch,
// spectre/<name> included); 1 with the wrong-branch line in a linked
// worktree on another branch or a detached HEAD; both lines together when
// both hold; 2 with nothing on stdout on bad arguments, a missing directory
// or a directory that is not a git work tree.

// pclFx is the harness's sandbox: a main checkout on `main` with linked
// worktrees on spectre/demo, spectre/other and a detached HEAD.
func pclFx(t *testing.T) (sandbox, main string) {
	t.Helper()
	sandbox = t.TempDir()
	main = sandbox + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", main)
	g.write(main+"/README.md", "seed")
	g.git(main, "add", "-A")
	g.git(main, "commit", "-q", "-m", "seed")
	g.git(main, "worktree", "add", "-q", "-b", "spectre/demo", main+"/.worktrees/demo")
	g.git(main, "worktree", "add", "-q", "-b", "spectre/other", main+"/.worktrees/other")
	g.git(main, "worktree", "add", "-q", "--detach", main+"/.worktrees/detached")
	if g.err != nil {
		t.Fatal(g.err)
	}
	mkdir(t, sandbox+"/plain")
	mkdir(t, main+"/.worktrees/demo/sub")
	return sandbox, main
}

func TestCheckPlanningCommitLocation(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	sandbox, main := pclFx(t)
	run := func(t *testing.T, rc int, want string, args ...string) {
		t.Helper()
		r := runGuard("check-planning-commit-location", args, env)
		if r.rc != rc || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", r.rc, r.stdout, r.err, rc, want)
		}
		if rc == 2 && r.err == "" {
			t.Fatal("cannot answer with nothing on stderr")
		}
	}

	// Every case reads the fixture as pclFx left it; the main checkout moved
	// onto spectre/<name> below gets a fixture of its own.
	t.Run("linked worktree on spectre/<name> passes", func(t *testing.T) {
		t.Parallel()
		run(t, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo on spectre/demo\n", main+"/.worktrees/demo", "demo")
	})
	// The harness's "main checkout on its default branch is refused" and "…
	// is also the wrong branch": both lines, in this order.
	t.Run("main checkout on its default branch is refused and is also the wrong branch", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\nPLANNING-COMMIT-WRONG-BRANCH: "+main+" on main — expected spectre/demo\n", main, "demo")
	})
	t.Run("linked worktree of another change is refused", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/other on spectre/other — expected spectre/demo\n", main+"/.worktrees/other", "demo")
	})
	t.Run("detached worktree is refused", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/detached on detached — expected spectre/demo\n", main+"/.worktrees/detached", "demo")
	})
	t.Run("a subdirectory of the change worktree passes", func(t *testing.T) {
		t.Parallel()
		run(t, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo/sub on spectre/demo\n", main+"/.worktrees/demo/sub", "demo")
	})
	t.Run("beyond the harness: a relative worktree resolves against the caller's cwd and is printed as given", func(t *testing.T) {
		t.Parallel()
		r := runGuard("check-planning-commit-location", []string{".worktrees/demo", "demo"}, Env{Getenv: os.Getenv, Dir: main})
		if want := "PLANNING-COMMIT-LOCATION-OK: .worktrees/demo on spectre/demo\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q; want 0 and %q", r.rc, r.stdout, want)
		}
	})

	// The main checkout moved onto spectre/<name> is still the main checkout,
	// and the only line is the main-checkout one.
	t.Run("main checkout on spectre/<name> is still refused, with only the main-checkout line", func(t *testing.T) {
		t.Parallel()
		_, main := pclFx(t)
		var g fxGit
		g.git(main+"/.worktrees/demo", "checkout", "-q", "--detach")
		g.git(main, "checkout", "-q", "spectre/demo")
		if g.err != nil {
			t.Fatal(g.err)
		}
		run(t, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\n", main, "demo")
	})

	for _, c := range []struct {
		name string
		args []string
	}{
		{"no arguments cannot answer", nil},
		{"one argument cannot answer", []string{main}},
		{"empty name cannot answer", []string{main, ""}},
		{"missing directory cannot answer", []string{sandbox + "/nope", "demo"}},
		{"non-git directory cannot answer", []string{sandbox + "/plain", "demo"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			run(t, 2, "", c.args...)
		})
	}
	t.Run("beyond the harness: the usage line", func(t *testing.T) {
		t.Parallel()
		r := runGuard("check-planning-commit-location", nil, env)
		if want := "check-planning-commit-location.sh: usage: check-planning-commit-location.sh <worktree> <name>\n"; r.err != want {
			t.Errorf("stderr %q, want %q", r.err, want)
		}
	})
}
