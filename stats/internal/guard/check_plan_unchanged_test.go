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

// Every case of scripts/test-check-plan-unchanged.sh at c9b8fde8, one subtest
// per ok: label. The harness built a throwaway repository per case and ran
// the real guard with stdout and stderr merged; here the guard runs
// in-process over a copy of one repository built once, each case's steps
// run in order once, and each labelled check reads its step's result.
//
// The scoping the cases pin: the guard asserts the change's plan tree
// (`spectre/changes/<name>/`) and nothing else, because a gated per-task
// reviewer flies beside a working implementer whose commits and source edits
// legitimately move the rest of the worktree. Cases 6 and 7 pin exactly that;
// cases 3-5 are the failure shapes the guard exists for, case 3 verbatim the
// KAN-635 incident.

// pucRepo is make_repo: one committed plan file and one committed source
// file, built once; every case copies it.
var pucRepo = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp(execFixtures.dir, "puc-repo")
	if err != nil {
		return "", err
	}
	g := &fxGit{}
	g.git(dir, "init", "-q")
	g.write(dir+"/src.txt", "source: original")
	if err := os.MkdirAll(dir+"/spectre/changes/demo", 0o755); err != nil {
		return "", err
	}
	g.write(dir+"/spectre/changes/demo/tasks.md", "plan: original")
	g.git(dir, "add", "-A")
	g.git(dir, "commit", "-q", "-m", "original")
	return dir, g.err
})

// pucStep is one line of a harness case: an action on the fixture, or a
// guard run (label set, want its expected exit), or an expect_names check on
// the last run's output (needle set).
type pucStep struct {
	act    func(g *fxGit, wt string)
	label  string
	args   func(wt, snap string) []string
	want   int
	needle string
}

func pucRun(mode string, want int, label string) pucStep {
	return pucStep{label: label, want: want, args: func(wt, snap string) []string {
		return []string{mode, wt, "demo", snap}
	}}
}

func pucArgs(label string, want int, args func(wt, snap string) []string) pucStep {
	return pucStep{label: label, want: want, args: args}
}

func pucNames(label, needle string) pucStep { return pucStep{label: label, needle: needle} }

func pucAct(f func(g *fxGit, wt string)) pucStep { return pucStep{act: f} }

func pucEditPlan(g *fxGit, wt string) {
	g.appendLine(wt+"/spectre/changes/demo/tasks.md", "plan: edited")
}

func pucRemovePlan(g *fxGit, wt string) {
	g.git(wt, "rm", "-rq", "spectre")
	g.git(wt, "commit", "-q", "-m", "plan-removed")
}

type pucResult struct {
	code int
	out  string
}

func pucExec(root string, steps []pucStep, notRepo bool, snapRel string) ([]pucResult, error) {
	wt := filepath.Join(root, "wt")
	if notRepo {
		if err := os.Mkdir(wt, 0o755); err != nil {
			return nil, err
		}
	} else {
		src, err := pucRepo()
		if err != nil {
			return nil, err
		}
		if out, err := exec.Command("cp", "-Rp", src, wt).CombinedOutput(); err != nil {
			return nil, fmt.Errorf("cp: %v\n%s", err, out)
		}
	}
	snapDir := filepath.Join(root, "snap")
	if err := os.Mkdir(snapDir, 0o755); err != nil {
		return nil, err
	}
	snap := filepath.Join(snapDir, snapRel)
	g := &fxGit{}
	res := make([]pucResult, len(steps))
	var last string
	for i, s := range steps {
		switch {
		case s.act != nil:
			s.act(g, wt)
		case s.args != nil:
			var out bytes.Buffer
			guard, ok := Registry["check-plan-unchanged"]
			if !ok {
				return nil, fmt.Errorf("check-plan-unchanged is not registered")
			}
			res[i].code = guard(s.args(wt, snap), Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: root}, &out, &out)
			last = out.String()
			res[i].out = last
		default:
			res[i].out = last
		}
		if g.err != nil {
			return nil, g.err
		}
	}
	return res, nil
}

func TestCheckPlanUnchanged(t *testing.T) {
	t.Parallel()
	type pucCase struct {
		snapRel string
		notRepo bool
		steps   []pucStep
	}
	cases := []pucCase{
		{steps: []pucStep{
			pucRun("snapshot", 0, "case 1a: snapshot on a clean repo exits 0"),
			pucRun("verify", 0, "case 1b: verify on an untouched clean repo exits 0"),
		}},
		// A dirty state honestly carried across the dispatch is not a violation.
		{steps: []pucStep{
			pucAct(pucEditPlan),
			pucRun("snapshot", 0, "case 2a: snapshot with a dirty plan tree exits 0"),
			pucRun("verify", 0, "case 2b: verify with the same dirty plan tree exits 0"),
		}},
		// THE KAN-635 SHAPE.
		{steps: []pucStep{
			pucAct(pucEditPlan),
			pucRun("snapshot", 0, "case 3a: snapshot with a dirty plan tree exits 0"),
			pucAct(func(g *fxGit, wt string) { g.git(wt, "checkout", "--", "spectre/changes/demo") }),
			pucRun("verify", 1, "case 3b: verify after the edits are checked away exits 1"),
			pucNames("case 3c: names the destroyed plan file", "tasks.md"),
		}},
		{steps: []pucStep{
			pucRun("snapshot", 0, "case 4a: snapshot exits 0"),
			pucAct(func(g *fxGit, wt string) {
				if err := os.Remove(wt + "/spectre/changes/demo/tasks.md"); err != nil && g.err == nil {
					g.err = err
				}
			}),
			pucRun("verify", 1, "case 4b: verify after a plan deletion exits 1"),
			pucNames("case 4c: names the deleted plan file", "tasks.md"),
		}},
		// Untracked plan files count.
		{steps: []pucStep{
			pucRun("snapshot", 0, "case 5a: snapshot exits 0"),
			pucAct(func(g *fxGit, wt string) { g.write(wt+"/spectre/changes/demo/planted.md", "planted") }),
			pucRun("verify", 1, "case 5b: verify after a planted plan file exits 1"),
			pucNames("case 5c: names the planted file", "planted.md"),
		}},
		{steps: []pucStep{
			pucRun("snapshot", 0, "case 6a: snapshot exits 0"),
			pucAct(func(g *fxGit, wt string) {
				g.appendLine(wt+"/src.txt", "source: edited")
				g.git(wt, "add", "src.txt")
				g.git(wt, "commit", "-q", "-m", "implementer-work")
			}),
			pucRun("verify", 0, "case 6b: verify ignores work outside the plan tree"),
		}},
		{steps: []pucStep{
			pucRun("snapshot", 0, "case 7a: snapshot exits 0"),
			pucAct(func(g *fxGit, wt string) {
				g.write(wt+"/other.txt", "more")
				g.git(wt, "add", "other.txt")
				g.git(wt, "commit", "-q", "-m", "head-moves")
			}),
			pucRun("verify", 0, "case 7b: verify ignores a HEAD move"),
		}},
		{snapRel: "absent", steps: []pucStep{
			pucRun("verify", 2, "case 8: verify without a snapshot file exits 2"),
		}},
		{steps: []pucStep{
			pucArgs("case 9a: no arguments exits 2", 2, func(string, string) []string { return nil }),
			pucArgs("case 9b: a bad mode exits 2", 2, func(wt, snap string) []string { return []string{"check", wt, "demo", snap} }),
			pucArgs("case 9c: a non-directory worktree exits 2", 2, func(wt, snap string) []string {
				return []string{"snapshot", wt + "/no-such-dir", "demo", snap}
			}),
			pucArgs("case 9d: a change name with a separator exits 2", 2, func(wt, snap string) []string {
				return []string{"snapshot", wt, "../escape", snap}
			}),
		}},
		{snapRel: "no-dir/snapshot", steps: []pucStep{
			pucRun("snapshot", 2, "case 10: snapshot to an unwritable path exits 2"),
		}},
		{notRepo: true, steps: []pucStep{
			pucRun("snapshot", 2, "case 11: a non-repository worktree exits 2"),
		}},
		// A change with no plan tree has nothing to lose.
		{steps: []pucStep{
			pucAct(pucRemovePlan),
			pucRun("snapshot", 0, "case 12a: snapshot with no plan tree exits 0"),
			pucRun("verify", 0, "case 12b: verify with the plan tree still absent exits 0"),
		}},
		{steps: []pucStep{
			pucAct(pucRemovePlan),
			pucRun("snapshot", 0, "case 13a: snapshot with no plan tree exits 0"),
			pucAct(func(g *fxGit, wt string) {
				if err := os.MkdirAll(wt+"/spectre/changes/demo", 0o755); err != nil && g.err == nil {
					g.err = err
				}
				g.write(wt+"/spectre/changes/demo/tasks.md", "new")
			}),
			pucRun("verify", 1, "case 13b: verify after the plan tree appears exits 1"),
		}},
		// A fix run writes into spectre/changes/archive/<name>/ once
		// integrate's run 1 archived the change.
		{steps: []pucStep{
			pucAct(func(g *fxGit, wt string) {
				g.git(wt, "mv", "spectre/changes/demo", "spectre/changes/archive-tmp")
				if err := os.MkdirAll(wt+"/spectre/changes/archive", 0o755); err != nil && g.err == nil {
					g.err = err
				}
				g.git(wt, "mv", "spectre/changes/archive-tmp", "spectre/changes/archive/demo")
				g.git(wt, "commit", "-q", "-m", "archive")
				g.appendLine(wt+"/spectre/changes/archive/demo/tasks.md", "plan: appended by a fix run")
			}),
			pucRun("snapshot", 0, "case 14a: snapshot of an archived change exits 0"),
			pucAct(func(g *fxGit, wt string) { g.git(wt, "checkout", "-q", "--", "spectre/changes/archive/demo/tasks.md") }),
			pucRun("verify", 1, "case 14b: verify after the archived plan edit was destroyed exits 1"),
		}},
	}
	for _, c := range cases {
		snapRel := c.snapRel
		if snapRel == "" {
			snapRel = "snapshot"
		}
		root := t.TempDir()
		run := sync.OnceValues(func() ([]pucResult, error) { return pucExec(root, c.steps, c.notRepo, snapRel) })
		for i, s := range c.steps {
			if s.label == "" {
				continue
			}
			t.Run(s.label, func(t *testing.T) {
				t.Parallel()
				res, err := run()
				if err != nil {
					t.Fatal(err)
				}
				r := res[i]
				if s.needle != "" {
					if !strings.Contains(r.out, s.needle) {
						t.Fatalf("expected %s named in output, got:\n%s", s.needle, r.out)
					}
					return
				}
				if r.code != s.want {
					t.Fatalf("expected exit %d, got %d\n%s", s.want, r.code, r.out)
				}
			})
		}
	}
}
