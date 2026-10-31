package guard

import (
	"bytes"
	"os"
	"testing"
)

// cwlFx is a main checkout on `main` with linked worktrees on spectre/demo,
// spectre/other and a detached HEAD.
func cwlFx(t *testing.T) (sandbox, main string) {
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

// cwlRun calls checkChangeWorktreeLocation directly and pins its exit code
// and its whole stdout; a cannot-answer exit must say why on stderr.
func cwlRun(t *testing.T, env Env, rc int, want string, args ...string) {
	t.Helper()
	var out, errb bytes.Buffer
	got := checkChangeWorktreeLocation(args, env, &out, &errb)
	if got != rc || out.String() != want {
		t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", got, out.String(), errb.String(), rc, want)
	}
	if rc == 2 && errb.Len() == 0 {
		t.Fatal("cannot answer with nothing on stderr")
	}
}

func TestChangeWorktreeLocationRefusesAMainCheckout(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	sandbox, main := cwlFx(t)
	t.Run("main checkout on its default branch is refused and is also the wrong branch", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, env, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\nPLANNING-COMMIT-WRONG-BRANCH: "+main+" on main — expected spectre/demo\n", main, "demo")
	})
	t.Run("main checkout on spectre/<name> is still refused, with only the main-checkout line", func(t *testing.T) {
		t.Parallel()
		_, main := cwlFx(t)
		var g fxGit
		g.git(main+"/.worktrees/demo", "checkout", "-q", "--detach")
		g.git(main, "checkout", "-q", "spectre/demo")
		if g.err != nil {
			t.Fatal(g.err)
		}
		cwlRun(t, env, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\n", main, "demo")
	})
	t.Run("linked worktree of another change is refused", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, env, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/other on spectre/other — expected spectre/demo\n", main+"/.worktrees/other", "demo")
	})
	t.Run("detached worktree is refused", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, env, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/detached on detached — expected spectre/demo\n", main+"/.worktrees/detached", "demo")
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
			cwlRun(t, env, 2, "", c.args...)
		})
	}
}

func TestChangeWorktreeLocationAdmitsAChangeWorktree(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	_, main := cwlFx(t)
	t.Run("linked worktree on spectre/<name> passes", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, env, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo on spectre/demo\n", main+"/.worktrees/demo", "demo")
	})
	t.Run("a subdirectory of the change worktree passes", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, env, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo/sub on spectre/demo\n", main+"/.worktrees/demo/sub", "demo")
	})
	t.Run("a relative worktree resolves against the caller's cwd and is printed as given", func(t *testing.T) {
		t.Parallel()
		cwlRun(t, Env{Getenv: os.Getenv, Dir: main}, 0, "PLANNING-COMMIT-LOCATION-OK: .worktrees/demo on spectre/demo\n", ".worktrees/demo", "demo")
	})
}
