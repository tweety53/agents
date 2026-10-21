package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-guard-autosquash.sh at c9b8fde8, one subtest per
// ok: label. The harness ran the real guard over throwaway repositories with
// stdout and stderr captured apart; here the guard runs in-process over a
// copy of one two-commit repository built once, each case taking its own.

// gaRepo is new_repo: two commits on the default branch. The root commit
// plays the fixup target's ancestor; HEAD the branch tip.
var gaRepo = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "ga-repo")
	if err != nil {
		return "", err
	}
	g := &fxGit{}
	g.git(dir, "init", "-q")
	g.write(dir+"/one.txt", "one")
	g.git(dir, "add", "one.txt")
	g.git(dir, "commit", "-qm", "one")
	g.write(dir+"/two.txt", "two")
	g.git(dir, "add", "two.txt")
	g.git(dir, "commit", "-qm", "two")
	return dir, g.err
})

// gaOrphan is orphan_commit: a commit on an orphan branch, so its sha is a
// real commit object that is nevertheless no ancestor of the default branch
// -- the incident's shape. The branch is switched back by name: a fresh
// repository's HEAD has no checkout reflog, so `git checkout -` cannot
// resolve.
func gaOrphan(g *fxGit, repo string) string {
	back := g.git(repo, "rev-parse", "--abbrev-ref", "HEAD")
	g.git(repo, "checkout", "-q", "--orphan", "stray")
	g.git(repo, "rm", "-qrf", ".")
	g.write(repo+"/stray.txt", "stray")
	g.git(repo, "add", "stray.txt")
	g.git(repo, "commit", "-qm", "stray")
	sha := g.git(repo, "rev-parse", "HEAD")
	g.git(repo, "checkout", "-q", back)
	return sha
}

func TestGuardAutosquash(t *testing.T) {
	t.Parallel()
	type gaCase struct {
		label string
		// setup builds the case in repo (a copy of gaRepo) and sandbox (an
		// empty directory), returning the guard's arguments and the needle
		// stderr must name ("" for none).
		setup    func(g *fxGit, repo, sandbox string) (args []string, needle string)
		want     int
		emptyOut bool // stdout must say nothing a caller could read as a verdict
	}
	rootSHA := func(g *fxGit, repo string) string { return g.git(repo, "rev-parse", "HEAD~1") }
	headSHA := func(g *fxGit, repo string) string { return g.git(repo, "rev-parse", "HEAD") }
	tasks := func(g *fxGit, sandbox, body string) string {
		p := sandbox + "/tasks.md"
		if g.err == nil {
			g.err = os.WriteFile(p, []byte(body), 0o644)
		}
		return p
	}
	cases := []gaCase{
		{label: "targets accepts a genuine ancestor", want: 0,
			setup: func(g *fxGit, repo, _ string) ([]string, string) {
				return []string{"targets", repo, rootSHA(g, repo), headSHA(g, repo)}, ""
			}},
		{label: "targets refuses a non-ancestor sha and names it", want: 1,
			setup: func(g *fxGit, repo, _ string) ([]string, string) {
				stray := gaOrphan(g, repo)
				return []string{"targets", repo, stray}, stray
			}},
		{label: "targets reports a non-resolving sha as unresolved", want: 1,
			setup: func(g *fxGit, repo, _ string) ([]string, string) {
				return []string{"targets", repo, "deadbeefcafe1234"}, "does not resolve as a commit object"
			}},
		{label: "after accepts a clean branch", want: 0,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				root := rootSHA(g, repo)
				p := tasks(g, sandbox, fmt.Sprintf("# plan\n\n- task 1 commit %s\n<!-- baseline %s -->\n", headSHA(g, repo), root))
				return []string{"after", repo, root, p}, ""
			}},
		{label: "after refuses a base that is no ancestor of HEAD", want: 1,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				stray := gaOrphan(g, repo)
				return []string{"after", repo, stray, tasks(g, sandbox, "")}, stray
			}},
		// The amended sha still exists as an object (cat-file -e alone would
		// pass it) but the branch no longer reaches it: the two-check
		// design's reason.
		{label: "after refuses a stale tasks.md sha and names it", want: 1,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				stale := headSHA(g, repo)
				g.git(repo, "commit", "-q", "--amend", "-m", "two-amended")
				return []string{"after", repo, rootSHA(g, repo), tasks(g, sandbox, "- task 5 baseline "+stale+"\n")}, stale
			}},
		{label: "after reports a hex-shaped token that resolves to no commit", want: 1,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				return []string{"after", repo, rootSHA(g, repo), tasks(g, sandbox, "recorded off the defaced token\n")}, "defaced"
			}},
		// A sha glued into a longer hex run is reported, never silently
		// skipped for exceeding the sha-length window.
		{label: "after reports a long hex run", want: 1,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				const long = "a12345678901234567890123456789012345678901"
				return []string{"after", repo, rootSHA(g, repo), tasks(g, sandbox, "- baseline "+long+"\n")}, long
			}},
		// A directory passes a readability test but not a regular-file test;
		// the sweep refuses, never fails open.
		{label: "after exits 2 on a tasks-md directory", want: 2, emptyOut: true,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				return []string{"after", repo, rootSHA(g, repo), sandbox}, ""
			}},
		{label: "after exits 2 on an unreadable tasks.md", want: 2, emptyOut: true,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				p := tasks(g, sandbox, "- baseline deadbeef\n")
				if g.err == nil {
					g.err = os.Chmod(p, 0)
				}
				return []string{"after", repo, rootSHA(g, repo), p}, "unreadable"
			}},
		{label: "after exits 2 on a missing tasks.md", want: 2, emptyOut: true,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				return []string{"after", repo, rootSHA(g, repo), sandbox + "/absent-tasks.md"}, ""
			}},
		{label: "targets exits 2 outside a git repository", want: 2,
			setup: func(g *fxGit, repo, sandbox string) ([]string, string) {
				return []string{"targets", sandbox, rootSHA(g, repo)}, ""
			}},
		{label: "no subcommand exits 2", want: 2,
			setup: func(*fxGit, string, string) ([]string, string) { return nil, "" }},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			src, err := gaRepo()
			if err != nil {
				t.Fatal(err)
			}
			root := t.TempDir()
			repo, sandbox := filepath.Join(root, "repo"), filepath.Join(root, "sandbox")
			if out, err := exec.Command("cp", "-Rp", src, repo).CombinedOutput(); err != nil {
				t.Fatalf("cp: %v\n%s", err, out)
			}
			mkdir(t, sandbox)
			g := &fxGit{}
			args, needle := c.setup(g, repo, sandbox)
			if g.err != nil {
				t.Fatal(g.err)
			}
			guard, ok := Registry["guard-autosquash"]
			if !ok {
				t.Fatal("guard-autosquash is not registered")
			}
			var stdout, stderr bytes.Buffer
			rc := guard(args, Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: root}, &stdout, &stderr)
			if rc != c.want || !strings.Contains(stderr.String(), needle) || (c.emptyOut && stdout.Len() > 0) {
				t.Fatalf("rc=%d (want %d) needle=%q\nstdout:\n%s\nstderr:\n%s", rc, c.want, needle, stdout.String(), stderr.String())
			}
		})
	}
}
