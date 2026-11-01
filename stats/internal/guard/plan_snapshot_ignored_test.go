package guard

import (
	"os"
	"strings"
	"testing"
)

// The plan tree is gitignored (KAN-1073): these fixtures put
// `spectre/changes/` in .gitignore, so no git listing sees the plan and
// each guard must read it from disk.

// psiRepo is a repository whose plan tree is ignored and on disk only.
func psiRepo(t *testing.T) (*fxGit, string) {
	t.Helper()
	wt := t.TempDir()
	g := &fxGit{}
	g.git("", "init", "-q", "-b", "work", wt)
	g.write(wt+"/.gitignore", "spectre/changes/\n.superpowers/")
	g.write(wt+"/a.txt", "a")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "closed clean")
	mkdir(t, wt+"/spectre/changes/demo")
	g.write(wt+"/spectre/changes/demo/tasks.md", "- [ ] 1. the first task")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return g, wt
}

// TestCheckPlanUnchangedSeesAnIgnoredPlanEdit pins the walk: an edit of an
// ignored plan's task text is a change, exit 1 naming the path. Under
// `git ls-files --exclude-standard` the ignored file is never listed, both
// dumps are empty and verify exits 0.
func TestCheckPlanUnchangedSeesAnIgnoredPlanEdit(t *testing.T) {
	g, wt := psiRepo(t)
	snap := t.TempDir() + "/snapshot"
	env := Env{Getenv: func(string) string { return "" }, Dir: wt}
	if r := runGuard("check-plan-unchanged", []string{"snapshot", wt, "demo", snap}, env); r.rc != 0 {
		t.Fatalf("snapshot exit %d\n%s", r.rc, r.out)
	}
	g.write(wt+"/spectre/changes/demo/tasks.md", "- [ ] 1. the first task, edited")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("check-plan-unchanged", []string{"verify", wt, "demo", snap}, env)
	if r.rc != 1 {
		t.Fatalf("verify exit %d, want 1\n%s", r.rc, r.out)
	}
	if !strings.Contains(r.err, "changed: spectre/changes/demo/tasks.md") {
		t.Errorf("stderr %q, want the changed path named", r.err)
	}
}

// TestCheckPlanUnchangedSeesAPlantedSymlink: a symlink planted under the
// plan tree after the snapshot is a change, as an untracked one was when git
// listed the tree.
func TestCheckPlanUnchangedSeesAPlantedSymlink(t *testing.T) {
	_, wt := psiRepo(t)
	snap := t.TempDir() + "/snapshot"
	env := Env{Getenv: func(string) string { return "" }, Dir: wt}
	if r := runGuard("check-plan-unchanged", []string{"snapshot", wt, "demo", snap}, env); r.rc != 0 {
		t.Fatalf("snapshot exit %d\n%s", r.rc, r.out)
	}
	if err := os.Symlink("tasks.md", wt+"/spectre/changes/demo/alias.md"); err != nil {
		t.Fatal(err)
	}
	if r := runGuard("check-plan-unchanged", []string{"verify", wt, "demo", snap}, env); r.rc != 1 {
		t.Fatalf("verify exit %d, want 1\n%s%s", r.rc, r.out, r.err)
	}
}

// psiLateFix runs the late-fix guard on wt with a clean close at HEAD and a
// one-line late fix on top.
func psiLateFix(t *testing.T, g *fxGit, wt string) guardResult {
	t.Helper()
	since := g.git(wt, "rev-parse", "HEAD")
	g.appendLine(wt+"/a.txt", "the late fix")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return runGuard("check-late-fix-trigger",
		[]string{"demo", wt + "/spectre/changes/demo/tasks.md", wt, since, "CLEAR:"}, lftEnv(wt, "[]"))
}

// TestCheckLateFixTriggerFindsATaskAddedSinceTheCopy pins condition 4 on an
// ignored plan: a task number tasks.md carries and the at-close copy lacks
// is scope growth (exit 4, small append); a tick since the close is not
// (exit 0, reduce).
func TestCheckLateFixTriggerFindsATaskAddedSinceTheCopy(t *testing.T) {
	for _, tc := range []struct {
		label, tasks string
		code         int
	}{
		{"a task added since the close", "- [ ] 1. the first task\n- [ ] 2. a second task", 4},
		{"a task ticked since the close", "- [x] 1. the first task", 0},
	} {
		t.Run(tc.label, func(t *testing.T) {
			g, wt := psiRepo(t)
			mkdir(t, wt+"/.superpowers/sdd/demo")
			g.write(wt+"/.superpowers/sdd/demo/tasks-at-close.md", "- [ ] 1. the first task")
			g.write(wt+"/spectre/changes/demo/tasks.md", tc.tasks)
			r := psiLateFix(t, g, wt)
			if r.rc != tc.code {
				t.Fatalf("exit %d, want %d\n%s", r.rc, tc.code, r.out)
			}
		})
	}
}

// TestCheckLateFixTriggerMissingCopyIsExit4 pins the missing at-close copy:
// condition 4 fails (exit 4 on a small delta) and stderr names the copy.
func TestCheckLateFixTriggerMissingCopyIsExit4(t *testing.T) {
	g, wt := psiRepo(t)
	r := psiLateFix(t, g, wt)
	if r.rc != 4 {
		t.Fatalf("exit %d, want 4\n%s", r.rc, r.out)
	}
	if !strings.Contains(r.err, ".superpowers/sdd/demo/tasks-at-close.md") {
		t.Errorf("stderr %q, want the missing copy named", r.err)
	}
}
