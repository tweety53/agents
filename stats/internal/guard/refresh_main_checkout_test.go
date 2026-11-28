package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// behindRepo is a main checkout one commit behind its origin: a bare origin
// at B, the checkout cloned at A and never pulled. The refresh's own fetch is
// what brings origin/main to B, so every case exercises the fetch too.
func behindRepo(t *testing.T) (repo, origin string) {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	origin, repo, other := dir+"/origin.git", dir+"/repo", dir+"/other"
	guardtest.FixtureRepo(t, origin, guardtest.RepoShape{Branch: "main", Bare: true})
	guardtest.GitRun(t, "", "clone", "-q", origin, repo)
	guardtest.WriteFile(t, repo+"/f.txt", "a\n")
	guardtest.GitRun(t, repo, "add", "f.txt")
	guardtest.GitRun(t, repo, "commit", "-q", "-m", "A")
	guardtest.GitRun(t, repo, "push", "-q", "origin", "main")
	guardtest.GitRun(t, "", "clone", "-q", origin, other)
	guardtest.WriteFile(t, other+"/f.txt", "b\n")
	guardtest.WriteFile(t, other+"/g.txt", "new\n")
	guardtest.GitRun(t, other, "add", "f.txt", "g.txt")
	guardtest.GitRun(t, other, "commit", "-q", "-m", "B")
	guardtest.GitRun(t, other, "push", "-q", "origin", "main")
	return repo, origin
}

func rmcRev(t *testing.T, repo, rev string) string {
	t.Helper()
	var g guardtest.FxGit
	s := g.Git(repo, "rev-parse", "--short", rev)
	if g.Err != nil {
		t.Fatal(g.Err)
	}
	return s
}

func rmcExists(p string) bool { _, err := os.Stat(p); return err == nil }

func TestRefreshMainCheckout(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}
	run := func(args ...string) guardtest.GuardResult { return runGuard("refresh-main-checkout", args, env) }
	refused := func(t *testing.T, r guardtest.GuardResult, repo, reason string) {
		t.Helper()
		if want := "REFRESH-REFUSED: " + repo + " " + reason + " — nothing touched\n"; r.Rc != 1 || r.Stdout != want {
			t.Fatalf("want exit 1, stdout %q; got %+v", want, r)
		}
	}
	// unmoved asserts a refusal left local main where it was.
	unmoved := func(t *testing.T, repo, before string) {
		t.Helper()
		if got := rmcRev(t, repo, "refs/heads/main"); got != before {
			t.Fatalf("local main moved: %s -> %s", before, got)
		}
	}

	t.Run("behind is fast-forwarded, then current", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		old := rmcRev(t, repo, "HEAD")
		r := run(repo, "main")
		tip := rmcRev(t, repo, "origin/main")
		if want := "REFRESH-DONE: " + repo + " fast-forwarded " + old + " -> " + tip + "\n"; r.Rc != 0 || r.Stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" || !rmcExists(repo+"/g.txt") {
			t.Fatal("worktree not at B")
		}
		r = run(repo, "main")
		if want := "REFRESH-CURRENT: " + repo + " already at " + tip + "\n"; r.Rc != 0 || r.Stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
	})

	t.Run("untracked survives", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.WriteFile(t, repo+"/untracked.txt", "loose\n")
		if r := run(repo, "main"); r.Rc != 0 || !rmcExists(repo+"/untracked.txt") {
			t.Fatalf("untracked file lost; %+v", r)
		}
	})

	// B adds g.txt; an untracked g.txt here is what the fast-forward would
	// overwrite, so git's own merge refuses it: exit 2, no verdict, nothing
	// moved and the untracked file kept.
	t.Run("fast-forward over an untracked file cannot answer", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.WriteFile(t, repo+"/g.txt", "mine\n")
		if r := run(repo, "main"); r.Rc != 2 || r.Stdout != "" || !strings.Contains(r.Err, "g.txt") {
			t.Fatalf("want exit 2, no verdict, git's message naming g.txt; got %+v", r)
		}
		unmoved(t, repo, before)
		if body, _ := os.ReadFile(repo + "/g.txt"); string(body) != "mine\n" {
			t.Fatal("untracked file overwritten")
		}
	})

	t.Run("detached refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.GitRun(t, repo, "checkout", "-q", "--detach")
		refused(t, run(repo, "main"), repo, "is detached")
		unmoved(t, repo, before)
	})

	t.Run("other branch refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.GitRun(t, repo, "checkout", "-q", "-b", "other")
		refused(t, run(repo, "main"), repo, "is on other, not main")
		unmoved(t, repo, before)
	})

	t.Run("unstaged edit refused and kept", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.WriteFile(t, repo+"/f.txt", "a\nedited\n")
		refused(t, run(repo, "main"), repo, "has tracked changes")
		unmoved(t, repo, before)
		if body, _ := os.ReadFile(repo + "/f.txt"); !strings.Contains(string(body), "edited") {
			t.Fatal("edit lost")
		}
	})

	t.Run("staged work refused and kept", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.WriteFile(t, repo+"/h.txt", "real\n")
		guardtest.GitRun(t, repo, "add", "h.txt")
		refused(t, run(repo, "main"), repo, "has tracked changes")
		unmoved(t, repo, before)
		if !rmcExists(repo + "/h.txt") {
			t.Fatal("staged file lost")
		}
	})

	t.Run("local commits not on origin refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.WriteFile(t, repo+"/local.txt", "mine\n")
		guardtest.GitRun(t, repo, "add", "local.txt")
		guardtest.GitRun(t, repo, "commit", "-q", "-m", "local")
		before := rmcRev(t, repo, "main")
		refused(t, run(repo, "main"), repo, "has main commits origin/main lacks")
		unmoved(t, repo, before)
	})

	t.Run("no origin base refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.GitRun(t, repo, "checkout", "-q", "-b", "trunk")
		before := rmcRev(t, repo, "trunk")
		refused(t, run(repo, "trunk"), repo, "has no origin/trunk")
		unmoved(t, repo, before)
	})

	t.Run("usage", func(t *testing.T) {
		t.Parallel()
		const usage = "usage: flow-guard refresh-main-checkout <main-checkout> <base>\n"
		for _, args := range [][]string{{t.TempDir()}, {t.TempDir() + "/absent", "main"}} {
			if r := run(args...); r.Rc != 2 || r.Stdout != "" || r.Err != usage {
				t.Fatalf("%v: want exit 2 with usage; got %+v", args, r)
			}
		}
	})

	// An unborn HEAD has no tip to name: git's own message, exit 2, never a
	// verdict.
	t.Run("unborn HEAD cannot answer", func(t *testing.T) {
		t.Parallel()
		repo := t.TempDir() + "/repo"
		guardtest.FixtureRepo(t, repo, guardtest.RepoShape{Branch: "main"})
		if r := run(repo, "main"); r.Rc != 2 || r.Stdout != "" || r.Err == "" {
			t.Fatalf("want exit 2, git's message on stderr; got %+v", r)
		}
	})
}

func TestRefreshMainCheckoutRestartsFlowdWhenStatsMoved(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		exit string
		rc   int
		line string
	}{
		{"a stats/ refresh restarts flowd", "0", 0, "FLOWD-RESTARTED: "},
		{"a failed restart is exit 2", "1", 2, "FLOWD-RESTART-FAILED: "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, origin := behindRepo(t)
			dir := filepath.Dir(repo)
			other := dir + "/other"
			guardtest.WriteFile(t, other+"/stats/Makefile", "restart:\n\ttrue\n")
			guardtest.WriteFile(t, other+"/stats/cmd/flowd/main.go", "package main\n")
			guardtest.GitRun(t, other, "add", "-A")
			guardtest.GitRun(t, other, "commit", "-q", "-m", "C")
			guardtest.GitRun(t, other, "push", "-q", origin, "main")
			log := dir + "/make.log"
			guardtest.WriteExec(t, dir+"/bin/make", "#!/usr/bin/env bash\nprintf 'make %s\\n' \"$*\" >>"+log+"\nexit "+tc.exit+"\n")
			r := runGuard("refresh-main-checkout", []string{repo, "main"}, core.Env{Getenv: guardtest.PathEnv(dir + "/bin")})
			calls, _ := os.ReadFile(log)
			if r.Rc != tc.rc || !strings.Contains(r.Stdout, "REFRESH-DONE: ") || !strings.Contains(r.Stdout, tc.line) ||
				!strings.Contains(string(calls), "make -C "+repo+"/stats restart") {
				t.Fatalf("%+v\n%s", r, calls)
			}
		})
	}
}

// KAN-926: drift landed on origin from another worktree leaves the main
// checkout holding identical uncommitted copies -- an unstaged edit, a
// staged addition -- which git's own fast-forward refuses to overwrite.
func TestRefreshMainCheckoutDiscardsLandedCopies(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	env := core.Env{Getenv: os.Getenv}

	t.Run("copies identical to origin are discarded, then fast-forwarded", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		old := rmcRev(t, repo, "HEAD")
		guardtest.WriteFile(t, repo+"/f.txt", "b\n")
		guardtest.WriteFile(t, repo+"/g.txt", "new\n")
		guardtest.GitRun(t, repo, "add", "g.txt")
		r := runGuard("refresh-main-checkout", []string{repo, "main"}, env)
		tip := rmcRev(t, repo, "origin/main")
		want := "REFRESH-DISCARDED: " + repo + " 2 tracked paths identical to origin/main\n" +
			"REFRESH-DONE: " + repo + " fast-forwarded " + old + " -> " + tip + "\n"
		if r.Rc != 0 || r.Stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
		var g guardtest.FxGit
		if s := g.Git(repo, "status", "--porcelain"); g.Err != nil || s != "" {
			t.Fatalf("checkout not clean: %q %v", s, g.Err)
		}
	})

	// One path matching origin does not carry another that does not: the
	// whole set is refused and every copy kept.
	t.Run("a copy differing from origin refuses the whole set", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		guardtest.WriteFile(t, repo+"/f.txt", "b\n")
		guardtest.WriteFile(t, repo+"/g.txt", "mine\n")
		guardtest.GitRun(t, repo, "add", "g.txt")
		r := runGuard("refresh-main-checkout", []string{repo, "main"}, env)
		if want := "REFRESH-REFUSED: " + repo + " has tracked changes — nothing touched\n"; r.Rc != 1 || r.Stdout != want {
			t.Fatalf("want exit 1, stdout %q; got %+v", want, r)
		}
		if got := rmcRev(t, repo, "main"); got != before {
			t.Fatalf("local main moved: %s -> %s", before, got)
		}
		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" {
			t.Fatal("identical copy discarded beside a differing one")
		}
	})

	refusedKept := func(t *testing.T, r guardtest.GuardResult, repo, reason string) {
		t.Helper()
		if want := "REFRESH-REFUSED: " + repo + " " + reason + " — nothing touched\n"; r.Rc != 1 || r.Stdout != want {
			t.Fatalf("want exit 1, stdout %q; got %+v", want, r)
		}
	}

	// The index differing is enough: a worktree copy matching origin does
	// not carry a staged copy that does not.
	t.Run("an index copy differing from origin refuses", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.WriteFile(t, repo+"/f.txt", "mine\n")
		guardtest.GitRun(t, repo, "add", "f.txt")
		guardtest.WriteFile(t, repo+"/f.txt", "b\n")
		refusedKept(t, runGuard("refresh-main-checkout", []string{repo, "main"}, env), repo, "has tracked changes")
		var g guardtest.FxGit
		if s := g.Git(repo, "show", ":f.txt"); g.Err != nil || s != "mine" {
			t.Fatalf("index copy lost: %q %v", s, g.Err)
		}
	})

	// A lossy clean filter makes differing bytes compare identical; any
	// filtered path refuses, so the stripped line survives.
	t.Run("a filtered path refuses", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.WriteFile(t, repo+"/.git/info/attributes", "f.txt filter=strip\n")
		guardtest.GitRun(t, repo, "config", "filter.strip.clean", "sed '/^OUT/d'")
		guardtest.WriteFile(t, repo+"/f.txt", "b\nOUT precious\n")
		refusedKept(t, runGuard("refresh-main-checkout", []string{repo, "main"}, env), repo, "has tracked changes")
		if body, _ := os.ReadFile(repo + "/f.txt"); !strings.Contains(string(body), "OUT precious") {
			t.Fatal("filtered content discarded")
		}
	})

	// Discarding is only worth it when the fast-forward then succeeds: an
	// untracked g.txt origin adds refuses before any copy is touched.
	t.Run("an untracked path origin adds refuses before discarding", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		guardtest.WriteFile(t, repo+"/f.txt", "b\n")
		guardtest.WriteFile(t, repo+"/g.txt", "mine\n")
		refusedKept(t, runGuard("refresh-main-checkout", []string{repo, "main"}, env), repo, "has untracked g.txt origin/main adds")
		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" {
			t.Fatal("copy discarded ahead of a failing fast-forward")
		}
	})
}
