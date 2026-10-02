package guard

import (
	"os"
	"path/filepath"
	"testing"
)

// Every case of scripts/test-check-main-checkout-drift.sh at c4f26c84, the executable
// statement of the KAN-647 grammar, one subtest per harness case. stdout and
// stderr are read apart, never merged: a refusal puts its message on stderr
// and must leave stdout empty.

// newDriftRepo is the harness's new_repo: one commit on main and a bare
// `origin` whose HEAD points back at main, the shape every real clone's
// default-branch resolution depends on. The path is physical, as the guard
// names it.
func newDriftRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := dir + "/repo"
	gitRun(t, "", "init", "-q", "-b", "main", repo)
	writeFile(t, repo+"/base.txt", "base\n")
	gitRun(t, repo, "add", "base.txt")
	gitRun(t, repo, "commit", "-qm", "base")
	gitRun(t, "", "clone", "-q", "--bare", repo, dir+"/origin.git")
	gitRun(t, repo, "remote", "add", "origin", dir+"/origin.git")
	gitRun(t, repo, "fetch", "-q", "origin")
	gitRun(t, repo, "remote", "set-head", "origin", "-a")
	return repo
}

func TestCheckMainCheckoutDrift(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	const usage = "usage: check-main-checkout-drift.sh <main-checkout>\n"

	refusals := []struct {
		name    string
		args    func(t *testing.T) []string
		wantErr func(args []string) string
	}{
		{"no argument", func(*testing.T) []string { return nil }, func([]string) string { return usage }},
		{"two arguments", func(*testing.T) []string { return []string{"a", "b"} }, func([]string) string { return usage }},
		{"non-git directory", func(t *testing.T) []string {
			d, _ := filepath.EvalSymlinks(t.TempDir())
			return []string{d}
		}, func(a []string) string { return "check-main-checkout-drift: " + a[0] + " is not a git repository\n" }},
		{"unreadable path", func(t *testing.T) []string { return []string{t.TempDir() + "/no-such-path"} },
			func(a []string) string { return "check-main-checkout-drift: " + a[0] + " is not a directory\n" }},
		// `cd` rejects an existing file differently from a missing path, but
		// both are refusals.
		{"regular file argument", func(t *testing.T) []string {
			f := t.TempDir() + "/file"
			writeFile(t, f, "x\n")
			return []string{f}
		}, func(a []string) string { return "check-main-checkout-drift: " + a[0] + " is not a directory\n" }},
		// `symbolic-ref -d` removes the symref itself; `update-ref -d` on a
		// symref dereferences and would delete the branch instead.
		{"no origin/HEAD", func(t *testing.T) []string {
			r := newDriftRepo(t)
			gitRun(t, r, "symbolic-ref", "-d", "refs/remotes/origin/HEAD")
			return []string{r}
		}, func(a []string) string {
			return "check-main-checkout-drift: " + a[0] + " has no resolvable refs/remotes/origin/HEAD — the default branch cannot be named\n"
		}},
		// Deleting the branch a symref points at leaves the symref behind,
		// naming a branch that resolves to nothing.
		{"dangling origin/HEAD", func(t *testing.T) []string {
			r := newDriftRepo(t)
			gitRun(t, r, "update-ref", "-d", "refs/remotes/origin/main")
			return []string{r}
		}, func(a []string) string {
			return "check-main-checkout-drift: " + a[0] + "'s refs/remotes/origin/HEAD dangles — main does not resolve\n"
		}},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := tc.args(t)
			r := runGuard("check-main-checkout-drift", args, env)
			if r.rc != 2 || r.stdout != "" || r.err != tc.wantErr(args) {
				t.Fatalf("want exit 2, empty stdout, stderr %q; got %+v", tc.wantErr(args), r)
			}
		})
	}

	verdicts := []struct {
		name  string
		setup func(t *testing.T, repo string) string // returns the argument
		want  func(repo string) string
	}{
		{"clean repo", func(_ *testing.T, r string) string { return r },
			func(r string) string { return "DRIFT-CLEAN: " + r + " — main\n" }},
		{"untracked-only repo", func(t *testing.T, r string) string {
			writeFile(t, r+"/stray.txt", "stray\n")
			return r
		}, func(r string) string { return "DRIFT-CLEAN: " + r + " — main\n" }},
		{"foreign branch", func(t *testing.T, r string) string {
			gitRun(t, r, "checkout", "-qb", "feature")
			return r
		}, func(r string) string { return "DRIFT-BRANCH: " + r + " — on feature, not main\n" }},
		{"detached HEAD", func(t *testing.T, r string) string {
			gitRun(t, r, "checkout", "-q", "--detach", "HEAD")
			return r
		}, func(r string) string { return "DRIFT-BRANCH: " + r + " — on (detached HEAD), not main\n" }},
		{"unstaged drift", func(t *testing.T, r string) string {
			writeFile(t, r+"/base.txt", "base\nreverted\n")
			return r
		}, func(r string) string { return "DRIFT-DIRTY: " + r + " — 1 tracked entries\n" }},
		{"staged + unstaged mix", func(t *testing.T, r string) string {
			writeFile(t, r+"/staged.txt", "staged\n")
			gitRun(t, r, "add", "staged.txt")
			writeFile(t, r+"/base.txt", "base\nreverted\n")
			return r
		}, func(r string) string { return "DRIFT-DIRTY: " + r + " — 2 tracked entries\n" }},
		{"foreign branch + reverted content", func(t *testing.T, r string) string {
			gitRun(t, r, "checkout", "-qb", "kan-527")
			writeFile(t, r+"/base.txt", "base\nreverted\n")
			return r
		}, func(r string) string {
			return "DRIFT-BRANCH: " + r + " — on kan-527, not main\nDRIFT-DIRTY: " + r + " — 1 tracked entries\n"
		}},
		// A symlinked argument resolves to the real path in the verdict.
		{"physical path", func(t *testing.T, r string) string {
			gitRun(t, r, "checkout", "-qb", "feature")
			link := t.TempDir() + "/via-link"
			if err := os.Symlink(r, link); err != nil {
				t.Fatal(err)
			}
			return link
		}, func(r string) string { return "DRIFT-BRANCH: " + r + " — on feature, not main\n" }},
	}
	for _, tc := range verdicts {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newDriftRepo(t)
			arg := tc.setup(t, repo)
			r := runGuard("check-main-checkout-drift", []string{arg}, env)
			if r.rc != 0 || r.stdout != tc.want(repo) || r.err != "" {
				t.Fatalf("want exit 0, stdout %q; got %+v", tc.want(repo), r)
			}
		})
	}
}
