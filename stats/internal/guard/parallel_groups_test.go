package guard

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// pgTask is one plan task for pgSelect: open unless done, its After set and Files.
func pgTask(id string, done bool, after []string, files ...string) pdbTask {
	return pdbTask{id: id, unchecked: !done, afterSet: after, afterDeclared: true, filesPresent: true, files: files}
}

func TestParallelGroupsSelect(t *testing.T) {
	t.Parallel()
	none := func(string) bool { return false }
	for name, c := range map[string]struct {
		groups [][]string
		tasks  []pdbTask
		ui     func(string) bool
		limit  int
		want   []int
	}{
		"independent groups with disjoint files run together": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, nil, "b.go")}, none, 2, []int{1, 2}},
		"an After task still open keeps the group serial": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, []string{"1"}, "b.go")}, none, 2, []int{1}},
		"an After task already done does not": {
			[][]string{{"1"}, {"2"}, {"3"}}, []pdbTask{pgTask("1", true, nil, "a.go"), pgTask("2", false, []string{"1"}, "b.go"), pgTask("3", false, []string{"1"}, "c.go")}, none, 2, []int{2, 3}},
		"an After task inside the same group does not": {
			[][]string{{"1", "2"}, {"3"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, []string{"1"}, "b.go"), pgTask("3", false, nil, "c.go")}, none, 2, []int{1, 2}},
		"a shared file keeps the group serial": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, nil, "./a.go")}, none, 2, []int{1}},
		"a file under a directory the other group names overlaps": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "pkg/"), pgTask("2", false, nil, "pkg/x.go")}, none, 2, []int{1}},
		"a ui paths file keeps the group serial": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, nil, "web/x.tsx")}, func(f string) bool { return strings.HasPrefix(f, "web/") }, 2, []int{1}},
		"no Files: declared keeps the group serial": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), {id: "2", unchecked: true, afterDeclared: true}}, none, 2, []int{1}},
		"the limit caps the batch": {
			[][]string{{"1"}, {"2"}, {"3"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, nil, "b.go"), pgTask("3", false, nil, "c.go")}, none, 2, []int{1, 2}},
		"limit 1 is serial": {
			[][]string{{"1"}, {"2"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, nil, "b.go")}, none, 1, []int{1}},
		"the batch stops at the first group that cannot join": {
			[][]string{{"1"}, {"2"}, {"3"}}, []pdbTask{pgTask("1", false, nil, "a.go"), pgTask("2", false, []string{"1"}, "b.go"), pgTask("3", false, nil, "c.go")}, none, 3, []int{1}},
		"every group done is an empty batch": {
			[][]string{{"1"}}, []pdbTask{pgTask("1", true, nil, "a.go")}, none, 2, nil},
	} {
		if got := pgSelect(c.groups, c.tasks, c.ui, c.limit); !slices.Equal(got, c.want) {
			t.Errorf("%s: got %v, want %v", name, got, c.want)
		}
	}
}

// pgFixture is a canonical worktree on spectre/demo holding a plan of two independent tasks in two
// groups, its decision recorded.
func pgFixture(t *testing.T) (wt string, env core.Env) {
	t.Helper()
	wt = t.TempDir() + "/wt"
	guardtest.Mkdir(t, wt)
	guardtest.FixtureRepo(t, wt, guardtest.RepoShape{Branch: "spectre/demo"})
	guardtest.GitRun(t, wt, "config", "user.name", "t")
	guardtest.GitRun(t, wt, "config", "user.email", "t@t")
	guardtest.WriteFile(t, wt+"/.gitignore", ".superpowers/\nspectre/\n")
	guardtest.WriteFile(t, wt+"/base.txt", "base\n")
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "base")
	guardtest.WriteFile(t, wt+"/spectre/changes/demo/tasks.md", "# Tasks\n\n- [ ] 1. one\n\n**Files:**\n- `a.txt`\n\n**After:** none\n\n- [ ] 2. two\n\n**Files:**\n- `b.txt`\n\n**After:** none\n")
	guardtest.WriteFile(t, wt+"/.superpowers/sdd/decision.json", `{"groups":[{"tasks":["1"],"reason":"r"},{"tasks":["2"],"reason":"r"}],"panel":[],"visual":{"verify":"skipped"}}`)
	return wt, core.Env{Getenv: guardtest.PathEnv(""), Dir: wt}
}

func TestParallelGroupsNext(t *testing.T) {
	t.Parallel()
	t.Run("two independent groups", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		if r := runGuard("parallel-groups", []string{"next", wt}, env); r.Rc != 0 || r.Out != "GROUPS: 1 2" {
			t.Fatalf("exit %d, want 0 and GROUPS: 1 2\n%s", r.Rc, r.Out)
		}
	})
	t.Run("a JVM build is serial", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		guardtest.WriteFile(t, wt+"/settings.gradle.kts", "")
		if r := runGuard("parallel-groups", []string{"next", wt}, env); r.Rc != 0 || r.Out != "SERIAL: JVM build — one build at a time\nGROUPS: 1" {
			t.Fatalf("exit %d, want 0 and SERIAL then GROUPS: 1\n%s", r.Rc, r.Out)
		}
	})
	t.Run("no decision is a usage error", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		if err := os.Remove(wt + "/.superpowers/sdd/decision.json"); err != nil {
			t.Fatal(err)
		}
		if r := runGuard("parallel-groups", []string{"next", wt}, env); r.Rc != 2 {
			t.Fatalf("exit %d, want 2\n%s", r.Rc, r.Out)
		}
	})
}

// pgCopy creates group g's copy of wt through throwaway-worktree and commits files in it under
// Task-Id id.
func pgCopy(t *testing.T, wt string, env core.Env, g, id string, files ...string) string {
	t.Helper()
	cp := wt + "-group-" + g
	if r := runGuard("throwaway-worktree", []string{"create", wt, cp, "--sdd"}, env); r.Rc != 0 {
		t.Fatalf("create %s: %s", cp, r.Out)
	}
	for i := 0; i < len(files); i += 2 {
		guardtest.WriteFile(t, cp+"/"+files[i], files[i+1])
	}
	guardtest.GitRun(t, cp, "add", "-A")
	guardtest.GitRun(t, cp, "commit", "-q", "-m", "task "+id, "-m", "Task-Id: "+id)
	guardtest.WriteFile(t, cp+"/.superpowers/sdd/implementer-report-"+g+".md", "report\n")
	return cp
}

func TestParallelGroupsMerge(t *testing.T) {
	t.Parallel()
	t.Run("copies land in argument order and are removed", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		c2 := pgCopy(t, wt, env, "2", "2", "b.txt", "b\n")
		c3 := pgCopy(t, wt, env, "3", "3", "c.txt", "c\n")
		guardtest.GitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "task 1", "-m", "Task-Id: 1")
		r := runGuard("parallel-groups", []string{"merge", wt, c2, c3}, env)
		if r.Rc != 0 || strings.Count(r.Out, "PICKED: ") != 2 {
			t.Fatalf("exit %d, want 0 and two PICKED lines\n%s", r.Rc, r.Out)
		}
		if got := tcfGit(t, wt, "log", "--format=%s", "-3"); got != "task 3\ntask 2\ntask 1" {
			t.Errorf("log %q, want task 1, 2, 3 in order", got)
		}
		for _, cp := range []string{c2, c3} {
			if _, err := os.Stat(cp); !os.IsNotExist(err) {
				t.Errorf("%s still on disk", cp)
			}
		}
		if _, err := os.Stat(wt + "/.superpowers/sdd/implementer-report-3.md"); err != nil {
			t.Errorf("report not folded back: %v", err)
		}
	})
	t.Run("a conflict is BLOCKED, aborted, and keeps only the unmerged copies", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		c2 := pgCopy(t, wt, env, "2", "2", "b.txt", "b\n")
		c3 := pgCopy(t, wt, env, "3", "3", "b.txt", "other\n")
		r := runGuard("parallel-groups", []string{"merge", wt, c2, c3}, env)
		if r.Rc != 1 || !strings.Contains(r.Out, "BLOCKED: "+c3) {
			t.Fatalf("exit %d, want 1 and BLOCKED naming %s\n%s", r.Rc, c3, r.Out)
		}
		if got := tcfGit(t, wt, "status", "--porcelain", "--untracked-files=no"); got != "" {
			t.Errorf("canonical tree left dirty: %q", got)
		}
		if _, err := os.Stat(c2); !os.IsNotExist(err) {
			t.Errorf("merged copy %s still on disk", c2)
		}
		if _, err := os.Stat(c3); err != nil {
			t.Errorf("unmerged copy %s removed: %v", c3, err)
		}
	})
	t.Run("uncommitted work in a copy is BLOCKED before any pick", func(t *testing.T) {
		t.Parallel()
		wt, env := pgFixture(t)
		c2 := pgCopy(t, wt, env, "2", "2", "b.txt", "b\n")
		guardtest.WriteFile(t, c2+"/wip.txt", "wip\n")
		before := tcfGit(t, wt, "rev-parse", "HEAD")
		if r := runGuard("parallel-groups", []string{"merge", wt, c2}, env); r.Rc != 1 || !strings.Contains(r.Out, "BLOCKED: "+c2) {
			t.Fatalf("exit %d, want 1 and BLOCKED\n%s", r.Rc, r.Out)
		}
		if tcfGit(t, wt, "rev-parse", "HEAD") != before {
			t.Error("HEAD moved")
		}
	})
}

func TestChangeWorktreesFindsGroupCopies(t *testing.T) {
	t.Parallel()
	porcelain := "worktree /w/demo\nbranch refs/heads/spectre/demo\n\nworktree /w/demo-group-2\nbranch refs/heads/flow-copy/demo-group-2\n\n"
	if _, copies := changeWorktrees([]byte(porcelain), "demo"); !slices.Equal(copies, []string{"/w/demo-group-2"}) {
		t.Errorf("copies %v, want the group copy", copies)
	}
}
