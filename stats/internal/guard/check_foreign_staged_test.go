package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-foreign-staged.sh at c4f26c84, one subtest per ok:
// label. The harness built throwaway repositories and ran the guard with
// stdout and stderr captured apart; here the guard runs in-process over the
// same repositories, and each verdict is pinned whole rather than grepped.

// fsRepo is the harness's new_repo: a repository with one commit on `main`,
// so staged entries are the only variable between cases. Its path is
// physical, as the guard's verdict names it (macOS's TMPDIR is a symlink).
func fsRepo(t *testing.T, g *fxGit) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := dir + "/repo"
	g.git("", "init", "-q", "-b", "main", repo)
	g.write(repo+"/base.txt", "base")
	g.git(repo, "add", "base.txt")
	g.git(repo, "commit", "-qm", "base")
	return repo
}

func TestCheckForeignStaged(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}

	for _, c := range []struct {
		name string
		args func(t *testing.T) []string
	}{
		{"no argument: exit 2, stdout empty, usage on stderr", func(*testing.T) []string { return nil }},
		{"non-git directory: exit 2, nothing on stdout", func(t *testing.T) []string { return []string{t.TempDir()} }},
		{"unreadable path: exit 2, nothing on stdout", func(t *testing.T) []string { return []string{t.TempDir() + "/no-such-path"} }},
		{"beyond the harness: an empty argument is cd's null directory", func(*testing.T) []string { return []string{""} }},
		{"beyond the harness: two arguments are a usage error", func(t *testing.T) []string { return []string{t.TempDir(), t.TempDir()} }},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			r := runGuard("check-foreign-staged", c.args(t), env)
			if r.rc != 2 || r.stdout != "" || r.err == "" {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want 2, empty stdout, a message on stderr", r.rc, r.stdout, r.err)
			}
		})
	}
	t.Run("no argument: the usage line", func(t *testing.T) {
		t.Parallel()
		if r := runGuard("check-foreign-staged", nil, env); r.err != "usage: check-foreign-staged.sh <main-checkout>\n" {
			t.Errorf("stderr %q", r.err)
		}
	})

	for _, c := range []struct {
		name  string
		setup func(g *fxGit, repo string)
		want  string // REPO stands for the repository's physical path
	}{
		{"clean repo: STAGED-CLEAN verdict, exit 0", func(*fxGit, string) {}, "STAGED-CLEAN: REPO\n"},
		{"untracked-only repo: STAGED-CLEAN verdict, exit 0", func(g *fxGit, repo string) {
			g.write(repo+"/stray.txt", "stray")
		}, "STAGED-CLEAN: REPO\n"},
		{"staged file: one FOREIGN-STAGED line, verdict counts 1, exit 0", func(g *fxGit, repo string) {
			g.write(repo+"/new.txt", "new")
			g.git(repo, "add", "new.txt")
		}, "FOREIGN-STAGED: A  new.txt\nSTAGED-FOREIGN: REPO — 1\n"},
		{"staged+unstaged mix: only the staged entry listed, unstaged named nowhere", func(g *fxGit, repo string) {
			g.write(repo+"/a.txt", "a2")
			g.git(repo, "add", "a.txt")
			g.appendLine(repo+"/base.txt", "modified")
		}, "FOREIGN-STAGED: A  a.txt\nSTAGED-FOREIGN: REPO — 1\n"},
		{"staged deletion + addition: both listed, verdict counts 2", func(g *fxGit, repo string) {
			g.git(repo, "rm", "-q", "base.txt")
			g.write(repo+"/added.txt", "added")
			g.git(repo, "add", "added.txt")
		}, "FOREIGN-STAGED: A  added.txt\nFOREIGN-STAGED: D  base.txt\nSTAGED-FOREIGN: REPO — 2\n"},
		// Both branches ADD f.txt with different content, so the merge stops
		// on an add/add conflict, which porcelain prints as `AA f.txt` — both
		// sides are staged, which is exactly the shape this guard lists.
		{"unmerged entry: AA listed, verdict counts 1", func(g *fxGit, repo string) {
			g.git(repo, "checkout", "-qb", "side")
			g.write(repo+"/f.txt", "side")
			g.git(repo, "add", "f.txt")
			g.git(repo, "commit", "-qm", "side")
			g.git(repo, "checkout", "-q", "main")
			g.write(repo+"/f.txt", "main")
			g.git(repo, "add", "f.txt")
			g.git(repo, "commit", "-qm", "main")
			if g.err == nil {
				g.git(repo, "merge", "side")
				g.err = nil // the conflict is the fixture
			}
		}, "FOREIGN-STAGED: AA f.txt\nSTAGED-FOREIGN: REPO — 1\n"},
		// `git add -N` stages an index entry, so it is index work like any
		// other stage — but porcelain prints its index code blank, ` A
		// pending.txt`, which a first-column filter alone would hide (panel
		// finding F2, round 1).
		{"intent-to-add entry: the A-index line listed, verdict counts 1", func(g *fxGit, repo string) {
			g.write(repo+"/pending.txt", "pending")
			g.git(repo, "add", "-N", "pending.txt")
		}, "FOREIGN-STAGED:  A pending.txt\nSTAGED-FOREIGN: REPO — 1\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			var g fxGit
			repo := fsRepo(t, &g)
			c.setup(&g, repo)
			if g.err != nil {
				t.Fatal(g.err)
			}
			r := runGuard("check-foreign-staged", []string{repo}, env)
			if want := strings.ReplaceAll(c.want, "REPO", repo); r.rc != 0 || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want 0 and %q", r.rc, r.stdout, r.err, want)
			}
		})
	}

	t.Run("physical path: a symlinked argument resolves to the real path in the verdict", func(t *testing.T) {
		t.Parallel()
		var g fxGit
		repo := fsRepo(t, &g)
		g.write(repo+"/new.txt", "new")
		g.git(repo, "add", "new.txt")
		if g.err != nil {
			t.Fatal(g.err)
		}
		link := filepath.Dir(repo) + "/via-link"
		if err := os.Symlink(repo, link); err != nil {
			t.Fatal(err)
		}
		r := runGuard("check-foreign-staged", []string{link}, env)
		if want := "FOREIGN-STAGED: A  new.txt\nSTAGED-FOREIGN: " + repo + " — 1\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q; want 0 and %q", r.rc, r.stdout, want)
		}
	})

	t.Run("beyond the harness: a relative argument resolves against the caller's cwd", func(t *testing.T) {
		t.Parallel()
		var g fxGit
		repo := fsRepo(t, &g)
		if g.err != nil {
			t.Fatal(g.err)
		}
		r := runGuard("check-foreign-staged", []string{"repo"}, Env{Getenv: os.Getenv, Dir: filepath.Dir(repo)})
		if want := "STAGED-CLEAN: " + repo + "\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q; want 0 and %q", r.rc, r.stdout, want)
		}
	})
}
