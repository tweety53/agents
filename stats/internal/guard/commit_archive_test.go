package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// caFx is a change worktree as run 1's drop step finds it: on spectre/demo
// with spectre/changes/demo and its fix sub-change spectre/changes/demo-fix-1
// committed, holding the rendered ledger (no panel record) under its
// gitignored .superpowers/sdd/ -- which the drop must not copy anywhere.
type caFx struct {
	wt string
	g  fxGit
}

func caNewFx(t *testing.T) *caFx {
	t.Helper()
	dir := t.TempDir()
	fx := &caFx{wt: dir + "/wt"}
	fx.g.git("", "init", "-q", "-b", "main", fx.wt)
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"}, {"commit.gpgsign", "false"}} {
		fx.g.git(fx.wt, "config", kv[0], kv[1])
	}
	mkdir(t, fx.wt+"/spectre/changes/demo")
	fx.g.write(fx.wt+"/spectre/changes/demo/tasks.md", "plan")
	fx.g.write(fx.wt+"/base.txt", "base")
	fx.g.write(fx.wt+"/.gitignore", ".superpowers/\n")
	fx.g.git(fx.wt, "add", "-A")
	fx.g.git(fx.wt, "commit", "-qm", "base")
	fx.g.git(fx.wt, "checkout", "-q", "-b", "spectre/demo")
	mkdir(t, fx.wt+"/spectre/changes/demo-fix-1")
	fx.g.write(fx.wt+"/spectre/changes/demo-fix-1/tasks.md", "fix plan")
	fx.g.write(fx.wt+"/spectre/changes/demo/narrative.md", "narrative")
	fx.g.git(fx.wt, "add", "-A")
	fx.g.git(fx.wt, "commit", "-qm", "work")
	mkdir(t, fx.wt+"/.superpowers/sdd/ledgers")
	fx.g.write(fx.wt+"/.superpowers/sdd/ledgers/demo.md", "ledger")
	return fx
}

// run runs commit-archive in-process.
func (fx *caFx) run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	env := Env{Dir: fx.wt, Getenv: os.Getenv}
	if len(args) == 0 {
		args = []string{fx.wt, "demo"}
	}
	var out, errb bytes.Buffer
	code := commitArchive(args, env, &out, &errb)
	return code, out.String(), errb.String()
}

// TestCommitArchiveDropsTheChange pins run 1's landing commit: the live
// change directory and its fix sub-change leave the tip in one
// `chore(spectre): drop <name>` commit, nothing lands under
// spectre/changes/archive/, and no ledger or panel record is written.
func TestCommitArchiveDropsTheChange(t *testing.T) {
	t.Parallel()
	fx := caNewFx(t)
	fx.g.write(fx.wt+"/spectre/changes/demo/scratch.patch", "untracked")
	code, out, errb := fx.run(t)
	head := fx.g.git(fx.wt, "rev-parse", "HEAD")
	if code != 0 || out != "ARCHIVE-COMMITTED: "+head+"\n" {
		t.Fatalf("exit %d, stdout %q, stderr %q; want 0 and ARCHIVE-COMMITTED: %s", code, out, errb, head)
	}
	if s := fx.g.git(fx.wt, "log", "-1", "--format=%s"); s != "chore(spectre): drop demo" {
		t.Errorf("subject %q", s)
	}
	for _, d := range []string{"spectre/changes/demo", "spectre/changes/demo-fix-1", "spectre/changes/archive"} {
		if tree := fx.g.git(fx.wt, "ls-tree", "-d", "HEAD", d); tree != "" {
			t.Errorf("%s still at HEAD: %s", d, tree)
		}
		if _, err := os.Stat(fx.wt + "/" + d); err == nil {
			t.Errorf("%s still on disk", d)
		}
	}
	if st := fx.g.git(fx.wt, "show", "--name-status", "--format=", "HEAD"); strings.Contains(st, "ledger.md") || strings.Contains(st, "panel.md") || strings.Contains(st, "\nA\t") || strings.HasPrefix(st, "A\t") {
		t.Errorf("the drop commit added something:\n%s", st)
	}
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
}

func TestCommitArchive(t *testing.T) {
	t.Parallel()

	t.Run("committed", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		code, out, errb := fx.run(t)
		head := fx.g.git(fx.wt, "rev-parse", "HEAD")
		if code != 0 || out != "ARCHIVE-COMMITTED: "+head+"\n" {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 0 and ARCHIVE-COMMITTED: %s", code, out, errb, head)
		}
		if s := fx.g.git(fx.wt, "log", "-1", "--format=%s"); s != "chore(spectre): drop demo" {
			t.Errorf("subject %q", s)
		}
	})

	t.Run("nothing staged", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		if code, _, errb := fx.run(t); code != 0 {
			t.Fatalf("first run: exit %d, stderr %q", code, errb)
		}
		before := fx.g.git(fx.wt, "rev-parse", "HEAD")
		code, out, errb := fx.run(t)
		if code != 0 || out != "ARCHIVE-NOTHING-STAGED\n" {
			t.Fatalf("exit %d, stdout %q, stderr %q; want 0 and ARCHIVE-NOTHING-STAGED", code, out, errb)
		}
		if after := fx.g.git(fx.wt, "rev-parse", "HEAD"); after != before {
			t.Error("a commit was made with nothing staged")
		}
	})

	t.Run("wrong branch", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		fx.g.git(fx.wt, "checkout", "-q", "-b", "elsewhere")
		before := fx.g.git(fx.wt, "rev-parse", "HEAD")
		code, out, _ := fx.run(t)
		if code != 1 || out != "ARCHIVE-WRONG-BRANCH: elsewhere\n" {
			t.Fatalf("exit %d, stdout %q; want 1 and ARCHIVE-WRONG-BRANCH: elsewhere", code, out)
		}
		if after := fx.g.git(fx.wt, "rev-parse", "HEAD"); after != before {
			t.Error("the wrong branch was committed on")
		}
		if !isFile(fx.wt + "/spectre/changes/demo/tasks.md") {
			t.Error("the change directory was removed before the branch was asserted")
		}
	})

	t.Run("done-when refusal", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		mkdir(t, fx.wt+"/docs/tickets")
		fx.g.write(fx.wt+"/docs/tickets/demo.md", "## Done when\n\n`shots/27.png` re-baselined.\n")
		fx.g.git(fx.wt, "add", "docs/tickets/demo.md")
		before := fx.g.git(fx.wt, "rev-parse", "HEAD")
		code, out, _ := fx.run(t)
		if code != 1 || !strings.Contains(out, "DONE-WHEN-PATH: shots/27.png — docs/tickets/demo.md") ||
			!strings.Contains(out, "DONE-WHEN-VIOLATION: ") {
			t.Fatalf("exit %d, stdout %q; want 1 and the Done-when guard's violation lines", code, out)
		}
		if after := fx.g.git(fx.wt, "rev-parse", "HEAD"); after != before {
			t.Error("a Done-when violation was committed")
		}
		if !isFile(fx.wt + "/spectre/changes/demo/tasks.md") {
			t.Error("the change directory was removed before the Done-when check refused")
		}
	})

	t.Run("scope violation", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		fx.g.write(fx.wt+"/stray.txt", "stray")
		fx.g.git(fx.wt, "add", "stray.txt")
		before := fx.g.git(fx.wt, "rev-parse", "HEAD")
		code, out, _ := fx.run(t)
		if code != 1 || !strings.Contains(out, "OUT-OF-SCOPE") || !strings.Contains(out, "stray.txt") ||
			!strings.Contains(out, "SCOPE-VIOLATION: ") {
			t.Fatalf("exit %d, stdout %q; want 1 and the scope guard's violation lines", code, out)
		}
		if after := fx.g.git(fx.wt, "rev-parse", "HEAD"); after != before {
			t.Error("a scope violation was committed")
		}
	})

	for _, c := range []struct {
		name string
		args func(fx *caFx) []string
	}{
		{"not a worktree", func(fx *caFx) []string { return []string{fx.wt + "/..", "demo"} }},
		{"not a plain name", func(fx *caFx) []string { return []string{fx.wt, "../demo"} }},
		{"usage", func(fx *caFx) []string { return []string{fx.wt} }},
	} {
		t.Run("cannot answer: "+c.name, func(t *testing.T) {
			t.Parallel()
			fx := caNewFx(t)
			before := fx.g.git(fx.wt, "rev-parse", "HEAD")
			code, out, _ := fx.run(t, c.args(fx)...)
			if code != 2 || out != "" {
				t.Fatalf("exit %d, stdout %q; want 2 and nothing on stdout", code, out)
			}
			if after := fx.g.git(fx.wt, "rev-parse", "HEAD"); after != before {
				t.Error("a commit was made without an answer")
			}
		})
	}

	// The real shim, end to end.
	t.Run("through the shim", func(t *testing.T) {
		t.Parallel()
		fx := caNewFx(t)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		cmd := exec.Command(tcfScriptsDir(t)+"/commit-archive.sh", fx.wt, "demo")
		cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
		out, err := cmd.Output()
		head := fx.g.git(fx.wt, "rev-parse", "HEAD")
		if err != nil || string(out) != "ARCHIVE-COMMITTED: "+head+"\n" {
			t.Fatalf("shim: %v, stdout %q; want ARCHIVE-COMMITTED: %s", err, out, head)
		}
	})
}
