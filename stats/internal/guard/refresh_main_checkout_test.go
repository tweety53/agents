package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	gitRun(t, "", "init", "-q", "--bare", "-b", "main", origin)
	gitRun(t, "", "clone", "-q", origin, repo)
	writeFile(t, repo+"/f.txt", "a\n")
	gitRun(t, repo, "add", "f.txt")
	gitRun(t, repo, "commit", "-q", "-m", "A")
	gitRun(t, repo, "push", "-q", "origin", "main")
	gitRun(t, "", "clone", "-q", origin, other)
	writeFile(t, other+"/f.txt", "b\n")
	writeFile(t, other+"/g.txt", "new\n")
	gitRun(t, other, "add", "f.txt", "g.txt")
	gitRun(t, other, "commit", "-q", "-m", "B")
	gitRun(t, other, "push", "-q", "origin", "main")
	return repo, origin
}

func rmcRev(t *testing.T, repo, rev string) string {
	t.Helper()
	var g fxGit
	s := g.git(repo, "rev-parse", "--short", rev)
	if g.err != nil {
		t.Fatal(g.err)
	}
	return s
}

func rmcExists(p string) bool { _, err := os.Stat(p); return err == nil }

func TestRefreshMainCheckout(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	run := func(args ...string) guardResult { return runGuard("refresh-main-checkout", args, env) }
	refused := func(t *testing.T, r guardResult, repo, reason string) {
		t.Helper()
		if want := "REFRESH-REFUSED: " + repo + " " + reason + " — nothing touched\n"; r.rc != 1 || r.stdout != want {
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
		if want := "REFRESH-DONE: " + repo + " fast-forwarded " + old + " -> " + tip + "\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" || !rmcExists(repo+"/g.txt") {
			t.Fatal("worktree not at B")
		}
		r = run(repo, "main")
		if want := "REFRESH-CURRENT: " + repo + " already at " + tip + "\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
	})

	t.Run("untracked survives", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		writeFile(t, repo+"/untracked.txt", "loose\n")
		if r := run(repo, "main"); r.rc != 0 || !rmcExists(repo+"/untracked.txt") {
			t.Fatalf("untracked file lost; %+v", r)
		}
	})

	t.Run("detached refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		gitRun(t, repo, "checkout", "-q", "--detach")
		refused(t, run(repo, "main"), repo, "is detached")
		unmoved(t, repo, before)
	})

	t.Run("other branch refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		gitRun(t, repo, "checkout", "-q", "-b", "other")
		refused(t, run(repo, "main"), repo, "is on other, not main")
		unmoved(t, repo, before)
	})

	t.Run("unstaged edit refused and kept", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		before := rmcRev(t, repo, "main")
		writeFile(t, repo+"/f.txt", "a\nedited\n")
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
		writeFile(t, repo+"/h.txt", "real\n")
		gitRun(t, repo, "add", "h.txt")
		refused(t, run(repo, "main"), repo, "has tracked changes")
		unmoved(t, repo, before)
		if !rmcExists(repo + "/h.txt") {
			t.Fatal("staged file lost")
		}
	})

	t.Run("local commits not on origin refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		writeFile(t, repo+"/local.txt", "mine\n")
		gitRun(t, repo, "add", "local.txt")
		gitRun(t, repo, "commit", "-q", "-m", "local")
		before := rmcRev(t, repo, "main")
		refused(t, run(repo, "main"), repo, "has main commits origin/main lacks")
		unmoved(t, repo, before)
	})

	t.Run("no origin base refused", func(t *testing.T) {
		t.Parallel()
		repo, _ := behindRepo(t)
		gitRun(t, repo, "checkout", "-q", "-b", "trunk")
		before := rmcRev(t, repo, "trunk")
		refused(t, run(repo, "trunk"), repo, "has no origin/trunk")
		unmoved(t, repo, before)
	})

	t.Run("usage", func(t *testing.T) {
		t.Parallel()
		const usage = "usage: refresh-main-checkout.sh <main-checkout> <base>\n"
		for _, args := range [][]string{{t.TempDir()}, {t.TempDir() + "/absent", "main"}} {
			if r := run(args...); r.rc != 2 || r.stdout != "" || r.err != usage {
				t.Fatalf("%v: want exit 2 with usage; got %+v", args, r)
			}
		}
	})

	// An unborn HEAD has no tip to name: git's own message, exit 2, never a
	// verdict.
	t.Run("unborn HEAD cannot answer", func(t *testing.T) {
		t.Parallel()
		repo := t.TempDir() + "/repo"
		gitRun(t, "", "init", "-q", "-b", "main", repo)
		if r := run(repo, "main"); r.rc != 2 || r.stdout != "" || r.err == "" {
			t.Fatalf("want exit 2, git's message on stderr; got %+v", r)
		}
	})
}
