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
	g.copyShape(wt, repoShape{Branch: "work"})
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
