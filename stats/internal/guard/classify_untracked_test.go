package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-classify-untracked.sh at c4f26c84, one subtest per harness
// case, each asserting the exit code, stdout and -- for the cases that move,
// ignore or report entries -- the state left in the worktree, the
// scratchpad and the local exclude file; an exit code alone proves nothing
// here. Two cases the port added: a listing that fails is exit 2, never
// CLEAN, and a non-ASCII screenshot is captured, not reported as an asset under
// its quoted name.

// cuFixture is the harness's new_fixture: a main checkout with a linked
// worktree at main/.worktrees/_landing-test on branch topic, one tracked
// file committed, and the exclude file and scratchpad the guard is
// contract-bound to resolve. Paths are physical, as `rev-parse
// --path-format=absolute` resolves them.
type cuFixture struct{ main, wt, exclude, scratch string }

func newCUFixture(t *testing.T) cuFixture {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	main := dir + "/main"
	gitRun(t, "", "init", "-q", "-b", "main", main)
	writeFile(t, main+"/tracked.txt", "base\n")
	gitRun(t, main, "add", "tracked.txt")
	gitRun(t, main, "commit", "-qm", "base")
	mkdir(t, main+"/.worktrees")
	wt := main + "/.worktrees/_landing-test"
	gitRun(t, main, "worktree", "add", "-q", "-b", "topic", wt)
	return cuFixture{main, wt, main + "/.git/info/exclude", main + "/.worktrees/_scratchpad"}
}

func cuStatus(t *testing.T, dir string) string {
	t.Helper()
	var g fxGit
	out := g.git(dir, "status", "--porcelain", "--untracked-files=all")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return out
}

func cuExists(p string) bool { _, err := os.Stat(p); return err == nil }

func cuExcludes(t *testing.T, exclude, line string) bool {
	t.Helper()
	body, err := os.ReadFile(exclude)
	return err == nil && strings.Contains("\n"+string(body), "\n"+line+"\n")
}

func cuEnv() Env { return Env{Getenv: os.Getenv} }

func TestClassifyUntracked(t *testing.T) {
	t.Parallel()
	const asset = " (operator decides — commit deliberately or delete)"

	t.Run("missing-argument", func(t *testing.T) {
		t.Parallel()
		r := runGuard("classify-untracked", nil, cuEnv())
		if r.rc != 2 || r.stdout != "" || r.err != "usage: classify-untracked.sh <landing-worktree>\n" {
			t.Fatalf("want exit 2, usage on stderr, empty stdout; got %+v", r)
		}
	})

	// Nothing to classify: the guard downstream creates it fresh and clean.
	t.Run("absent-worktree", func(t *testing.T) {
		t.Parallel()
		r := runGuard("classify-untracked", []string{t.TempDir() + "/absent"}, cuEnv())
		if r.rc != 0 || r.out != "" {
			t.Fatalf("want exit 0, silent; got %+v", r)
		}
	})

	t.Run("not-a-directory", func(t *testing.T) {
		t.Parallel()
		f := t.TempDir() + "/file"
		writeFile(t, f, "x\n")
		r := runGuard("classify-untracked", []string{f}, cuEnv())
		if r.rc != 2 || r.stdout != "" || r.err != "classify-untracked: "+f+" is not a directory\n" {
			t.Fatalf("got %+v", r)
		}
	})

	t.Run("not-a-worktree", func(t *testing.T) {
		t.Parallel()
		d := t.TempDir()
		r := runGuard("classify-untracked", []string{d}, cuEnv())
		if r.rc != 2 || r.stdout != "" || r.err != "classify-untracked: "+d+" is not a git worktree\n" {
			t.Fatalf("got %+v", r)
		}
	})

	t.Run("clean-worktree", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 || r.out != "CLEAN" {
			t.Fatalf("want exactly CLEAN at exit 0; got %+v", r)
		}
		if cuExists(fx.scratch) {
			t.Fatalf("scratchpad created: %s", fx.scratch)
		}
	})

	t.Run("capture-moved", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.wt+"/shot.png", "shot\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 || r.out != "CAPTURED: shot.png -> "+fx.scratch+"/shot.png" {
			t.Fatalf("got %+v", r)
		}
		if !cuExists(fx.scratch+"/shot.png") || cuExists(fx.wt+"/shot.png") {
			t.Fatal("capture not moved into the scratchpad")
		}
		if s := cuStatus(t, fx.wt); s != "" {
			t.Fatalf("worktree still dirty: %q", s)
		}
	})

	t.Run("config-ignored", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.wt+"/.claude/settings.json", "cfg\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 || r.out != "IGNORED: .claude/ (local exclude)" {
			t.Fatalf("got %+v", r)
		}
		if !cuExists(fx.wt + "/.claude/settings.json") {
			t.Fatal(".claude/ contents disturbed")
		}
		if !cuExcludes(t, fx.exclude, ".claude/") {
			t.Fatalf("%s lacks .claude/", fx.exclude)
		}
		if s := cuStatus(t, fx.wt); s != "" {
			t.Fatalf("worktree still dirty: %q", s)
		}
		// A second run finds nothing, and the exclude line is never doubled.
		r = runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		body, _ := os.ReadFile(fx.exclude)
		if r.out != "CLEAN" || strings.Count(string(body), ".claude/\n") != 1 {
			t.Fatalf("second run: %+v, exclude %q", r, body)
		}
	})

	t.Run("asset-reported", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.wt+"/handoff.zip", "bundle\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 || r.out != "ASSET: handoff.zip"+asset {
			t.Fatalf("got %+v", r)
		}
		if !cuExists(fx.wt + "/handoff.zip") {
			t.Fatal("asset disturbed")
		}
		if cuStatus(t, fx.wt) == "" {
			t.Fatal("tree unexpectedly clean")
		}
	})

	t.Run("mixed-classes", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.wt+"/shot.png", "shot\n")
		writeFile(t, fx.wt+"/.claude/settings.json", "cfg\n")
		writeFile(t, fx.wt+"/handoff.zip", "bundle\n")
		writeFile(t, fx.wt+"/tracked.txt", "base\nedited\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		want := "IGNORED: .claude/ (local exclude)\n" +
			"ASSET: handoff.zip" + asset + "\n" +
			"CAPTURED: shot.png -> " + fx.scratch + "/shot.png"
		if r.rc != 0 || r.out != want {
			t.Fatalf("want one line per class:\n%s\ngot %+v", want, r)
		}
		if !cuExists(fx.scratch+"/shot.png") || !cuExcludes(t, fx.exclude, ".claude/") || !cuExists(fx.wt+"/handoff.zip") {
			t.Fatal("a class was not settled as its contract says")
		}
		if body, _ := os.ReadFile(fx.wt + "/tracked.txt"); string(body) != "base\nedited\n" {
			t.Fatalf("tracked modification disturbed: %q", body)
		}
	})

	t.Run("collision-not-clobbered", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.scratch+"/shot.png", "old\n")
		writeFile(t, fx.wt+"/shot.png", "new\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 {
			t.Fatalf("got %+v", r)
		}
		if body, _ := os.ReadFile(fx.scratch + "/shot.png"); string(body) != "old\n" {
			t.Fatal("original clobbered")
		}
		names, _ := os.ReadDir(fx.scratch)
		if len(names) != 2 {
			t.Fatalf("want both copies, got %v", names)
		}
	})

	t.Run("non-ascii-capture", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		writeFile(t, fx.wt+"/é.png", "shot\n")
		r := runGuard("classify-untracked", []string{fx.wt}, cuEnv())
		if r.rc != 0 || r.out != "CAPTURED: é.png -> "+fx.scratch+"/é.png" || !cuExists(fx.scratch+"/é.png") {
			t.Fatalf("got %+v", r)
		}
	})

	t.Run("listing-fails", func(t *testing.T) {
		t.Parallel()
		fx := newCUFixture(t)
		bin := t.TempDir()
		stubGit(t, bin, `has ls-files "$@"`, "index file corrupt")
		r := runGuard("classify-untracked", []string{fx.wt}, Env{Getenv: pathEnv(bin)})
		want := "fatal: index file corrupt\nclassify-untracked: cannot list the untracked entries of " + fx.wt + "\n"
		if r.rc != 2 || r.stdout != "" || r.err != want {
			t.Fatalf("want exit 2, no CLEAN; got %+v", r)
		}
	})
}
