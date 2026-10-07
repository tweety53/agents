package guard

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every case of scripts/test-check-worktree-location.sh at 3d56584c, the
// executable statement of the KAN-462 §11 grammar, one subtest per harness
// case, plus the refusals the harness never named. Real repositories, real
// `git worktree add`; stdout and stderr are read apart, never merged: a
// refusal puts its message on stderr and must leave stdout empty. The
// harness's case 7 (a mutant guard that always prints LOCATION-OK must fail
// the suite) is covered by the exact-stdout assertions below — a mutant
// printing LOCATION-OK for a stray fails cases 3–5b.

// newLocationRepo is the harness's new_repo: a fresh repository with one
// commit, so `git worktree add` has a HEAD to branch from. The path is
// physical, as the guard names it.
func newLocationRepo(t *testing.T) string {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := dir + "/repo"
	gitRun(t, "", "init", "-q", repo)
	gitRun(t, repo, "commit", "-q", "--allow-empty", "-m", "init")
	return repo
}

// physicalTempDir is t.TempDir() in physical form, as git records a
// worktree's path at `add` time.
func physicalTempDir(t *testing.T) string {
	t.Helper()
	d, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestCheckWorktreeLocation(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	const self = "check-worktree-location: "

	verdicts := []struct {
		name   string
		setup  func(t *testing.T, repo string) string // returns the expected STRAY lines
		wantRC int
	}{
		{"case 1: ok with no worktree", func(*testing.T, string) string { return "" }, 0},
		{"case 2: ok with one under <repo>-worktrees", func(t *testing.T, r string) string {
			gitRun(t, r, "worktree", "add", "-q", r+"-worktrees/in-tree", "-b", "in-tree")
			return ""
		}, 0},
		{"case 3: stray sibling directory", func(t *testing.T, r string) string {
			d := physicalTempDir(t)
			gitRun(t, r, "worktree", "add", "-q", d+"/x", "-b", "sibling-branch")
			return "STRAY: " + d + "/x (refs/heads/sibling-branch)\n"
		}, 1},
		{"case 4: stray detached", func(t *testing.T, r string) string {
			d := physicalTempDir(t)
			gitRun(t, r, "worktree", "add", "-q", "--detach", d+"/x")
			return "STRAY: " + d + "/x (detached)\n"
		}, 1},
		{"case 5: stray path with a space", func(t *testing.T, r string) string {
			d := physicalTempDir(t)
			mkdir(t, d+"/with space")
			gitRun(t, r, "worktree", "add", "-q", "--detach", d+"/with space/x")
			return "STRAY: " + d + "/with space/x (detached)\n"
		}, 1},
		// "AT OR UNDER IS NOT STARTS WITH": <project>-worktrees-old/x
		// starts with the literal string <project>-worktrees without being
		// nested under it.
		{"case 5b: sibling directory sharing <repo>-worktrees as a string prefix", func(t *testing.T, r string) string {
			mkdir(t, r+"-worktrees-old")
			gitRun(t, r, "worktree", "add", "-q", "--detach", r+"-worktrees-old/x")
			return "STRAY: " + r + "-worktrees-old/x (detached)\n"
		}, 1},
		// Beyond the harness: two strays are counted, in porcelain order.
		{"two strays counted", func(t *testing.T, r string) string {
			d := physicalTempDir(t)
			gitRun(t, r, "worktree", "add", "-q", "--detach", d+"/a")
			gitRun(t, r, "worktree", "add", "-q", r+"-worktrees/ok", "-b", "ok")
			gitRun(t, r, "worktree", "add", "-q", d+"/b", "-b", "b")
			return "STRAY: " + d + "/a (detached)\nSTRAY: " + d + "/b (refs/heads/b)\n"
		}, 1},
	}
	for _, tc := range verdicts {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			repo := newLocationRepo(t)
			strays := tc.setup(t, repo)
			want := strays + "LOCATION-OK: " + repo + "\n"
			if tc.wantRC == 1 {
				want = strays + "LOCATION-STRAY: " + repo + " — " + strconv.Itoa(strings.Count(strays, "\n")) + "\n"
			}
			r := runGuard("check-worktree-location", []string{repo}, env)
			if r.rc != tc.wantRC || r.stdout != want || r.err != "" {
				t.Fatalf("want exit %d, stdout %q, empty stderr; got %+v", tc.wantRC, want, r)
			}
		})
	}

	// A symlinked argument is resolved to the physical root the verdict names.
	t.Run("physical path", func(t *testing.T) {
		t.Parallel()
		repo := newLocationRepo(t)
		link := t.TempDir() + "/via-link"
		if err := os.Symlink(repo, link); err != nil {
			t.Fatal(err)
		}
		r := runGuard("check-worktree-location", []string{link}, env)
		if want := "LOCATION-OK: " + repo + "\n"; r.rc != 0 || r.stdout != want || r.err != "" {
			t.Fatalf("want %q; got %+v", want, r)
		}
	})

	refusals := []struct {
		name    string
		args    func(t *testing.T) []string
		wantErr func(args []string) string
	}{
		{"no argument", func(*testing.T) []string { return nil },
			func([]string) string { return self + "usage: check-worktree-location.sh <project>\n" }},
		{"two arguments", func(*testing.T) []string { return []string{"a", "b"} },
			func([]string) string { return self + "usage: check-worktree-location.sh <project>\n" }},
		// bash's `cd ""` refuses an empty operand rather than staying put.
		{"empty argument", func(*testing.T) []string { return []string{""} },
			func([]string) string { return self + " is not a directory\n" }},
		{"missing path", func(t *testing.T) []string { return []string{t.TempDir() + "/no-such-path"} },
			func(a []string) string { return self + a[0] + " is not a directory\n" }},
		{"regular file argument", func(t *testing.T) []string {
			f := t.TempDir() + "/file"
			writeFile(t, f, "x\n")
			return []string{f}
		}, func(a []string) string { return self + a[0] + " is not a directory\n" }},
		{"case 6: exit 2 on a non-worktree argument", func(t *testing.T) []string {
			return []string{physicalTempDir(t)}
		}, func(a []string) string { return self + "cannot list the worktrees of " + a[0] + "\n" }},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args := tc.args(t)
			r := runGuard("check-worktree-location", args, env)
			if r.rc != 2 || r.stdout != "" || r.err != tc.wantErr(args) {
				t.Fatalf("want exit 2, empty stdout, stderr %q; got %+v", tc.wantErr(args), r)
			}
		})
	}
}

// TestCheckWorktreeLocationSiblingLayout pins the sibling layout
// (design.md: sibling-worktrees-layout): <parent>/<repo>-worktrees/ and below
// is accepted, <repo>/.worktrees/ and a prefix lookalike are STRAY, and the
// main checkout itself is never flagged.
func TestCheckWorktreeLocationSiblingLayout(t *testing.T) {
	t.Parallel()
	repo := newLocationRepo(t)
	sib := repo + "-worktrees"
	gitRun(t, repo, "worktree", "add", "-q", sib+"/x", "-b", "x")
	gitRun(t, repo, "worktree", "add", "-q", "--detach", sib+"/nested/y")
	gitRun(t, repo, "worktree", "add", "-q", repo+"/.worktrees/x", "-b", "in-repo")
	mkdir(t, repo+"-worktrees-x")
	gitRun(t, repo, "worktree", "add", "-q", "--detach", repo+"-worktrees-x/y")
	// `git worktree list` sorts linked worktrees by path: '-' before '/'.
	want := "STRAY: " + repo + "-worktrees-x/y (detached)\n" +
		"STRAY: " + repo + "/.worktrees/x (refs/heads/in-repo)\n" +
		"LOCATION-STRAY: " + repo + " — 2\n"
	r := runGuard("check-worktree-location", []string{repo}, Env{Getenv: os.Getenv})
	if r.rc != 1 || r.stdout != want || r.err != "" {
		t.Fatalf("want exit 1, stdout %q; got %+v", want, r)
	}
}
