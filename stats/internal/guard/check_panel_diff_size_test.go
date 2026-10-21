package guard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-check-panel-diff-size.sh at c9b8fde8, one
// subtest per ok: label, plus exact-output pins for every refusal the
// harness never reached. The harness built a sandboxed git repository per
// case and ran the guard with stdout and stderr merged; here the guard runs
// in-process over a copy of one seeded master repository built once, against
// real git.

// pdsMaster is new_repo's tree: a git repo with one seed commit.
var pdsMaster = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "check-panel-diff-size-master")
	if err != nil {
		return "", err
	}
	g := &fxGit{}
	g.git(dir, "init", "-q")
	g.write(dir+"/README.md", "seed")
	g.git(dir, "add", "-A")
	g.git(dir, "commit", "-q", "-m", "seed")
	return dir, g.err
})

// pdsLines is the harness's `lines N`.
func pdsLines(n int) string {
	var b strings.Builder
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	return b.String()
}

// pdsRepo copies the master and returns the copy and its seed commit (the
// merge base every case measures from).
func pdsRepo(t *testing.T) (string, string) {
	t.Helper()
	master, err := pdsMaster()
	if err != nil {
		t.Fatal(err)
	}
	repo := t.TempDir() + "/repo"
	if out, err := exec.Command("cp", "-R", master, repo).CombinedOutput(); err != nil {
		t.Fatalf("cp: %v\n%s", err, out)
	}
	g := &fxGit{}
	head := g.git(repo, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return repo, head
}

// pdsCommit writes body to repo/name and commits it.
func pdsCommit(t *testing.T, repo, name, body string) {
	t.Helper()
	writeFile(t, repo+"/"+name, body)
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-q", "-m", name)
}

func pdsRun(args ...string) guardResult {
	return runGuard("check-panel-diff-size", args, Env{Getenv: os.Getenv})
}

func TestCheckPanelDiffSize(t *testing.T) {
	t.Parallel()
	type check struct{ label, fail string }
	var checks []check
	want := func(label string, ok bool, r guardResult) {
		c := check{label: label}
		if !ok {
			c.fail = fmt.Sprintf("rc=%d out=%q stdout=%q stderr=%q", r.rc, r.out, r.stdout, r.err)
		}
		checks = append(checks, c)
	}

	// 1. A diff under the cap exits 0 and prints the count.
	repo, base := pdsRepo(t)
	pdsCommit(t, repo, "small.txt", pdsLines(10))
	r := pdsRun(repo, base)
	want("case 1: exits 0 for a diff under the cap", r.rc == 0, r)
	want("case 1: prints the count", strings.Contains(r.out, "10"), r)
	want("case 1: stdout is the count then the verdict", r.stdout == "10\nunder cap\n" && r.err == "", r)

	// 2. A diff over the default cap exits 1 and prints the count. 5200
	// tracks the default cap of 5000: a case whose size no longer exceeds
	// the default would pass its own name while testing nothing.
	repo, base = pdsRepo(t)
	pdsCommit(t, repo, "big.txt", pdsLines(5200))
	r = pdsRun(repo, base)
	want("case 2: exits 1 for a diff over the cap", r.rc == 1, r)
	want("case 2: prints the count", strings.Contains(r.out, "5200"), r)
	want("case 2: stdout is the count then over cap", r.stdout == "5200\nover cap\n", r)

	// 3. An explicit third argument overrides the default cap.
	repo, base = pdsRepo(t)
	pdsCommit(t, repo, "small.txt", pdsLines(10))
	r = pdsRun(repo, base, "5")
	want("case 3: explicit cap overrides the default", r.rc == 1, r)
	r = pdsRun(repo, base, "10")
	want("case 3b: a diff exactly at the cap is under it", r.rc == 0 && r.stdout == "10\nunder cap\n", r)
	r = pdsRun(repo, base, "")
	want("case 3c: an empty cap reads as the default", r.rc == 0 && r.stdout == "10\nunder cap\n", r)
	r = pdsRun(repo, base, "0010")
	want("case 3d: a zero-padded cap reads as decimal", r.rc == 0 && r.stdout == "10\nunder cap\n", r)
	r = pdsRun(repo, base, "99999999999999999999")
	want("case 3e: a cap past the integer range reads as over cap, as bash's test did", r.rc == 1 && r.stdout == "10\nover cap\n", r)

	// 4. Unstaged working-tree changes count toward the total alongside
	// committed ones (`git diff <merge-base>` never reports untracked files).
	repo, base = pdsRepo(t)
	pdsCommit(t, repo, "committed.txt", pdsLines(10))
	writeFile(t, repo+"/committed.txt", pdsLines(20))
	r = pdsRun(repo, base, "15")
	want("case 4: unstaged changes count toward the total", r.rc == 1 && r.stdout == "20\nover cap\n", r)

	// 4b. A pure rename counts as a full delete+add (--no-renames).
	repo, base = pdsRepo(t)
	gitRun(t, repo, "mv", "README.md", "RENAMED.md")
	r = pdsRun(repo, base)
	want("case 4b: a rename counts as delete plus add", r.rc == 0 && r.stdout == "2\nunder cap\n", r)

	// 5. A non-numeric cap exits 2.
	repo, base = pdsRepo(t)
	r = pdsRun(repo, base, "abc")
	want("case 5: non-numeric cap exits 2", r.rc == 2 && r.err == "check-panel-diff-size: cap is not a non-negative decimal integer: 'abc'\n" && r.stdout == "", r)
	r = pdsRun(repo, base, "-1")
	want("case 5b: a negative cap exits 2", r.rc == 2, r)

	// 6. A path that is not a git worktree exits 2.
	notRepo := t.TempDir()
	r = pdsRun(notRepo, "HEAD")
	want("case 6: non-repository path exits 2", r.rc == 2 && r.err == "check-panel-diff-size: "+notRepo+" is not a directory containing a git repository\n", r)

	// 7. The refusals the harness never reached, byte for byte.
	r = pdsRun()
	want("case 7a: no arguments prints the usage line", r.rc == 2 && r.err == "check-panel-diff-size: usage: check-panel-diff-size.sh <worktree> <merge-base> [cap]\n", r)
	r = pdsRun(notRepo + "/absent")
	want("case 7b: a missing worktree exits 2", r.rc == 2 && r.err == "check-panel-diff-size: "+notRepo+"/absent is not a directory\n", r)
	r = pdsRun(repo)
	want("case 7c: a missing merge base exits 2", r.rc == 2 && r.err == "check-panel-diff-size: merge base argument is required\n", r)
	r = pdsRun(repo, "--all")
	want("case 7d: a dash-leading merge base is read as a ref and refused", r.rc == 2 && r.err == "check-panel-diff-size: merge base does not resolve in "+repo+": '--all'\n", r)
	r = pdsRun(repo, "no-such-ref")
	want("case 7e: an unresolved merge base exits 2", r.rc == 2 && r.err == "check-panel-diff-size: merge base does not resolve in "+repo+": 'no-such-ref'\n", r)

	for _, c := range checks {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if c.fail != "" {
				t.Error(c.fail)
			}
		})
	}
}
