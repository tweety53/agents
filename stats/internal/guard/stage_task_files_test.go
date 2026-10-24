package guard

import (
	"os"
	"strings"
	"testing"
)

// stfRepo is a fixture worktree shaped like kan-1012's task 2: a tracked file
// edited, new files under directories git has never seen, and an edited
// spectre/changes/ file the task does not name.
func stfRepo(t *testing.T, g *fxGit, wt string) []string {
	t.Helper()
	g.git("", "init", "-q", "-b", "work", wt)
	for _, d := range []string{"/src/app/template/src/test/x", "/src/app/template/src/main/k/l", "/spectre/changes/c"} {
		mkdir(t, wt+d)
	}
	g.write(wt+"/src/app/template/src/test/x/T.kt", "base")
	g.write(wt+"/spectre/changes/c/tasks.md", "base")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "base")
	g.write(wt+"/src/app/template/src/test/x/T.kt", "edited")
	g.write(wt+"/spectre/changes/c/tasks.md", "ticked")
	g.write(wt+"/src/app/template/src/main/k/A.kt", "new")
	g.write(wt+"/src/app/template/src/main/k/l/B.kt", "new")
	return []string{"src/app/template/src/main/k/A.kt", "src/app/template/src/main/k/l/B.kt", "src/app/template/src/test/x/T.kt"}
}

// Every named path is staged, the new ones under unseen directories
// included, and the unnamed planning edit is not. With the excludes after
// the paths, git 2.50.1 stages T.kt alone on this layout and exits 0.
func TestStageTaskFilesStagesNewFilesUnderNewDirectories(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	files := stfRepo(t, &g, dir+"/wt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("stage-task-files", append([]string{dir + "/wt"}, files...), Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	staged := g.git(dir+"/wt", "diff", "--cached", "--name-only")
	if staged != strings.Join(files, "\n") {
		t.Errorf("staged:\n%s\nwant:\n%s", staged, strings.Join(files, "\n"))
	}
}

// An add that exits 0 having staged nothing is refused, naming each path.
func TestStageTaskFilesRefusesAPathLeftUnstaged(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	files := stfRepo(t, &g, dir+"/wt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	writeExec(t, dir+"/bin/git", "#!/usr/bin/env bash\nfor a in \"$@\"; do [ \"$a\" = add ] && exit 0; done\nexec "+fixtureGit+" \"$@\"\n")
	env := Env{Dir: dir, Getenv: func(k string) string {
		if k == "PATH" {
			return dir + "/bin:" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}}
	r := runGuard("stage-task-files", append([]string{dir + "/wt"}, files...), env)
	if r.rc != 1 {
		t.Fatalf("exit %d, want 1\n%s", r.rc, r.out)
	}
	for _, want := range []string{"?? src/app/template/src/main/k/A.kt", "?? src/app/template/src/main/k/l/B.kt", " M src/app/template/src/test/x/T.kt"} {
		if !strings.Contains(r.out, want) {
			t.Errorf("output lacks %q:\n%s", want, r.out)
		}
	}
}

// A named planning path git status would quote -- spaced here -- is still
// recognised as planning and skipped, not reported unstaged.
func TestStageTaskFilesSkipsAQuotedPlanningPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	files := stfRepo(t, &g, dir+"/wt")
	g.write(dir+"/wt/spectre/changes/c/a b.md", "new")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("stage-task-files", append([]string{dir + "/wt", "spectre/changes/c/a b.md"}, files...), Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
}
