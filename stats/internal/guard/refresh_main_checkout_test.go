package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-refresh-main-checkout.sh at c4f26c84, one subtest each,
// plus the cannot-answer exit the port names. Each case builds a repo at
// commit A, then moves refs/heads/main to a later commit B with update-ref
// -- exactly what a landing worktree's fast-forward does to the main
// checkout -- so the index and worktree lag the branch pointer.

// staleRepo is the harness's stale_repo: the checkout's index and worktree
// at A, refs/heads/main at B.
func staleRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := dir + "/repo"
	gitRun(t, "", "init", "-q", "-b", "main", repo)
	writeFile(t, repo+"/f.txt", "a\n")
	gitRun(t, repo, "add", "f.txt")
	gitRun(t, repo, "commit", "-q", "-m", "A")
	writeFile(t, repo+"/f.txt", "b\n")
	writeFile(t, repo+"/g.txt", "new\n")
	gitRun(t, repo, "add", "f.txt", "g.txt")
	gitRun(t, repo, "commit", "-q", "-m", "B")
	var g fxGit
	b := g.git(repo, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}
	gitRun(t, repo, "reset", "-q", "--hard", "HEAD~1")
	gitRun(t, repo, "update-ref", "refs/heads/main", b)
	return repo
}

func rmcShort(t *testing.T, repo, rev string) string {
	t.Helper()
	var g fxGit
	s := g.git(repo, "rev-parse", "--short", rev)
	if g.err != nil {
		t.Fatal(g.err)
	}
	return s
}

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

	// Cases 1 and 2: a stale index is refreshed, and a second run finds it
	// current.
	t.Run("stale index refreshed, then current", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		if cuStatus(t, repo) == "" {
			t.Fatal("fixture: expected a stale status")
		}
		tip, old := rmcShort(t, repo, "HEAD"), rmcShort(t, repo, "HEAD~1")
		r := run(repo, "main")
		if want := "REFRESH-DONE: " + repo + " index was the tree of " + old + "; now at " + tip + "\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
		if s := cuStatus(t, repo); s != "" {
			t.Fatalf("status not clean: %q", s)
		}
		if body, _ := os.ReadFile(repo + "/f.txt"); string(body) != "b\n" || !cuExists(repo+"/g.txt") {
			t.Fatal("worktree not at B")
		}
		r = run(repo, "main")
		if want := "REFRESH-CURRENT: " + repo + " already at " + tip + "\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("want %q; got %+v", want, r)
		}
	})

	t.Run("unstaged edit refused and kept", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		writeFile(t, repo+"/f.txt", "a\nedited\n")
		refused(t, run(repo, "main"), repo, "has unstaged changes")
		if body, _ := os.ReadFile(repo + "/f.txt"); !strings.Contains(string(body), "edited") {
			t.Fatal("edit lost")
		}
	})

	t.Run("real staged work refused and kept", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		writeFile(t, repo+"/h.txt", "real\n")
		gitRun(t, repo, "add", "h.txt")
		refused(t, run(repo, "main"), repo, "has staged changes that match no recent main tip")
		if !cuExists(repo + "/h.txt") {
			t.Fatal("staged file lost")
		}
	})

	t.Run("other branch refused", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		gitRun(t, repo, "checkout", "-q", "-b", "other")
		refused(t, run(repo, "main"), repo, "is on other, not main")
	})

	t.Run("detached refused", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		gitRun(t, repo, "checkout", "-q", "--detach")
		refused(t, run(repo, "main"), repo, "is detached")
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

	t.Run("untracked survives", func(t *testing.T) {
		t.Parallel()
		repo := staleRepo(t)
		writeFile(t, repo+"/untracked.txt", "loose\n")
		if r := run(repo, "main"); r.rc != 0 || !cuExists(repo+"/untracked.txt") {
			t.Fatalf("untracked file lost; %+v", r)
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
