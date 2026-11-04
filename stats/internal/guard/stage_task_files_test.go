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

// A named path already deleted -- by git rm or a plain rm -- is staged as
// a deletion, not refused by git add's pathspec check.
func TestStageTaskFilesStagesDeletedPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	wt := dir + "/wt"
	var g fxGit
	g.git("", "init", "-q", "-b", "work", wt)
	g.write(wt+"/gone-rm.txt", "base")
	g.write(wt+"/gone-plain.txt", "base")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "base")
	g.git(wt, "rm", "-q", "gone-rm.txt")
	if g.err != nil {
		t.Fatal(g.err)
	}
	if err := os.Remove(wt + "/gone-plain.txt"); err != nil {
		t.Fatal(err)
	}
	r := runGuard("stage-task-files", []string{wt, "gone-rm.txt", "gone-plain.txt"}, Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	if got, want := g.git(wt, "diff", "--cached", "--name-status"), "D\tgone-plain.txt\nD\tgone-rm.txt"; got != want {
		t.Errorf("staged:\n%s\nwant:\n%s", got, want)
	}
}

// A repository that gitignores spectre/changes/ while still tracking files
// under it (KAN-1073's agents tip before the parent's untrack commit): the
// exclude pathspec makes git 2.50.1's add fail on the ignored directory, so
// the add names the task's paths alone and a later reset unstages any
// planning path it swept in.
func TestStageTaskFilesStagesWhenTheChangeDirectoryIsIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var g fxGit
	files := stfRepo(t, &g, dir+"/wt")
	g.write(dir+"/wt/.gitignore", "spectre/changes/\n")
	mkdir(t, dir+"/wt/spectre/changes/d")
	g.write(dir+"/wt/spectre/changes/d/tasks.md", "new")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("stage-task-files", append([]string{dir + "/wt", "spectre"}, files...), Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	if staged := g.git(dir+"/wt", "diff", "--cached", "--name-only"); staged != strings.Join(files, "\n") {
		t.Errorf("staged:\n%s\nwant:\n%s", staged, strings.Join(files, "\n"))
	}
}

// A path untracked with git rm --cached and now gitignored stays on disk:
// its staged deletion needs no add, and git add refuses an ignored path
// (kan-1079 task 1: .idea/ untracked, then ignored whole).
func TestStageTaskFilesSkipsAnUntrackedIgnoredPath(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	wt := dir + "/wt"
	var g fxGit
	g.git("", "init", "-q", "-b", "work", wt)
	mkdir(t, wt+"/.idea")
	g.write(wt+"/.idea/vcs.xml", "base")
	g.write(wt+"/.gitignore", "base\n")
	g.git(wt, "add", ".")
	g.git(wt, "commit", "-qm", "base")
	g.git(wt, "rm", "-q", "--cached", ".idea/vcs.xml")
	g.write(wt+"/.gitignore", ".idea/\n")
	if g.err != nil {
		t.Fatal(g.err)
	}
	r := runGuard("stage-task-files", []string{wt, ".gitignore", ".idea/vcs.xml"}, Env{Dir: dir, Getenv: os.Getenv})
	if r.rc != 0 {
		t.Fatalf("exit %d, want 0\n%s", r.rc, r.out)
	}
	if got, want := g.git(wt, "diff", "--cached", "--name-status"), "M\t.gitignore\nD\t.idea/vcs.xml"; got != want {
		t.Errorf("staged:\n%s\nwant:\n%s", got, want)
	}
}
