package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// rgPlan is the change "demo"'s plan: task 1 declares a.txt and b.bin, plus
// everything under docs/ as collateral.
var rgPlan = bt(`- [ ] 1. Gate fixture

**Files:** ¤a.txt¤, ¤b.bin¤, ¤c.txt¤
**Allowed-collateral:** ¤docs/*¤
**Tests:** none
**Commit:** gate fixture
**Build:** green
`)

// rgRepo is a git repository with one base commit and, when plan is set,
// the change "demo"'s tasks.md on disk (never committed, as in a run).
func rgRepo(t *testing.T, plan bool) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	mkdir(t, wt)
	gitRun(t, wt, "init", "-q")
	writeFile(t, wt+"/base.txt", "base\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "base")
	if plan {
		writeFile(t, wt+"/spectre/changes/demo/tasks.md", rgPlan)
	}
	return wt
}

// rgCommit writes each path=body pair, commits exactly those paths, and
// returns the commit's sha.
func rgCommit(t *testing.T, wt string, files ...string) string {
	t.Helper()
	var paths []string
	for i := 0; i < len(files); i += 2 {
		writeFile(t, wt+"/"+files[i], files[i+1])
		paths = append(paths, files[i])
	}
	gitRun(t, wt, append([]string{"add", "--"}, paths...)...)
	gitRun(t, wt, "commit", "-q", "-m", "gate fixture")
	return tcfGit(t, wt, "rev-parse", "HEAD")
}

func rgLines(n int) string { return strings.Repeat("x\n", n) }

func TestCheckReviewGate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		setup func(t *testing.T) []string // the guard's arguments
		rc    int
		out   string // stdout exactly; on rc 2, a substring of stderr
	}{
		{"40 lines, all declared", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(30), "docs/x.md", rgLines(10))
			return []string{wt, "1", sha, wt, "demo"}
		}, 0, "QUIET: task 1 — 40 changed lines, every path declared\n"},
		{"more than 40 lines", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(41))
			return []string{wt, "1", sha, wt, "demo"}
		}, 0, "FIRE: task 1 — 41 changed lines (more than 40)\n"},
		{"undeclared path", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(2), "stray.txt", rgLines(1))
			return []string{wt, "1", sha, wt, "demo"}
		}, 0, "FIRE: task 1 — undeclared paths: stray.txt\n"},
		{"refused path a Files: widening declares", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(2), "c.txt", rgLines(1))
			return []string{wt, "1", sha, wt, "demo", "c.txt"}
		}, 0, "FIRE: task 1 — undeclared paths: c.txt\n"},
		{"both arms", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(50), "stray.txt", rgLines(1))
			return []string{wt, "1", sha, wt, "demo"}
		}, 0, "FIRE: task 1 — 51 changed lines (more than 40); undeclared paths: stray.txt\n"},
		{"map form sums across worktrees", func(t *testing.T) []string {
			wt, peer := rgRepo(t, true), rgRepo(t, false)
			sha := rgCommit(t, wt, "a.txt", rgLines(25))
			peerSHA := rgCommit(t, peer, "c.txt", rgLines(25))
			return []string{wt, "1", wt + "=" + sha + "," + peer + "=" + peerSHA, wt, "demo"}
		}, 0, "FIRE: task 1 — 50 changed lines (more than 40)\n"},
		{"binary counts as 0 lines", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(3), "b.bin", strings.Repeat("\x00\x01", 500))
			return []string{wt, "1", sha, wt, "demo"}
		}, 0, "QUIET: task 1 — 3 changed lines, every path declared\n"},
		{"unreadable plan", func(t *testing.T) []string {
			wt := rgRepo(t, false)
			sha := rgCommit(t, wt, "a.txt", rgLines(1))
			return []string{wt, "1", sha, wt, "demo"}
		}, 2, "check-review-gate: COULD NOT JUDGE — no tasks.md found for change 'demo'"},
		{"unresolvable commit", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			return []string{wt, "1", "0000000", wt, "demo"}
		}, 2, "check-review-gate: COULD NOT JUDGE — "},
		{"task absent from the plan", func(t *testing.T) []string {
			wt := rgRepo(t, true)
			sha := rgCommit(t, wt, "a.txt", rgLines(1))
			return []string{wt, "9", sha, wt, "demo"}
		}, 2, "check-review-gate: COULD NOT JUDGE — task 9 not found"},
		{"usage", func(t *testing.T) []string { return []string{"x"} }, 2, "check-review-gate: COULD NOT JUDGE — usage: check-review-gate.sh"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			args := c.setup(t)
			r := runGuard("check-review-gate", args, Env{Getenv: os.Getenv, Dir: args[0]})
			if r.rc != c.rc {
				t.Fatalf("exit %d, want %d\n%s", r.rc, c.rc, r.out)
			}
			if c.rc == 0 && r.stdout != c.out {
				t.Errorf("stdout %q, want %q (stderr %q)", r.stdout, c.out, r.err)
			}
			if c.rc == 2 && (r.stdout != "" || !strings.Contains(r.err, c.out)) {
				t.Errorf("stdout %q stderr %q, want empty stdout and stderr carrying %q", r.stdout, r.err, c.out)
			}
		})
	}
}
