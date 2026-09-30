package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ffFx is rbNewFx plus one task commit adding task.txt, with a dirty
// planning file the fold must carry around its rebase. task is the task
// commit's sha, parent its parent's.
type ffFx struct {
	*bmFx
	task, parent, tasksMD string
}

func ffNewFx(t *testing.T) *ffFx {
	t.Helper()
	fx := &ffFx{bmFx: rbNewFx(t)}
	fx.tasksMD = fx.repo + "/spectre/changes/demo/tasks.md"
	fx.g.write(fx.repo+"/task.txt", "task")
	fx.g.git(fx.repo, "add", "task.txt")
	fx.g.git(fx.repo, "commit", "-qm", "task")
	fx.task = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.parent = fx.g.git(fx.repo, "rev-parse", "HEAD^")
	return fx
}

// later commits one more file on top of the task commit.
func (fx *ffFx) later() {
	fx.g.write(fx.repo+"/later.txt", "later")
	fx.g.git(fx.repo, "add", "later.txt")
	fx.g.git(fx.repo, "commit", "-qm", "later")
}

// run runs fold-fixup in-process with the real guard-autosquash.sh beside
// it, as the shim exports it.
func (fx *ffFx) run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	fx.g.write(fx.tasksMD, "plan edited")
	return fx.runAsIs(t, args...)
}

// runAsIs is run without re-dirtying the planning file.
func (fx *ffFx) runAsIs(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	autosquash := tcfScriptsDir(t) + "/guard-autosquash.sh"
	env := Env{Dir: fx.dir, Getenv: func(k string) string {
		if k == "FLOW_GUARD_AUTOSQUASH" {
			return autosquash
		}
		return os.Getenv(k)
	}}
	var out, errb bytes.Buffer
	code := foldFixup(args, env, &out, &errb)
	return code, out.String(), errb.String()
}

func (fx *ffFx) rebasing() bool {
	_, err := os.Stat(fx.repo + "/.git/rebase-merge")
	return err == nil
}

// assertSettled checks the fold left no fixup commit, no rebase in
// progress, and the planning edit restored with its aside popped.
func (fx *ffFx) assertSettled(t *testing.T) {
	t.Helper()
	if log := fx.g.git(fx.repo, "log", "--format=%s"); strings.Contains(log, "fixup!") {
		t.Errorf("log still carries a fixup commit:\n%s", log)
	}
	if fx.rebasing() {
		t.Error("a rebase is still in progress")
	}
	if got := rbRead(fx.tasksMD); got != "plan edited\n" {
		t.Errorf("planning file %q, want the aside restored", got)
	}
	if l := fx.g.git(fx.repo, "stash", "list"); l != "" {
		t.Errorf("stash list %q, want the aside popped", l)
	}
}

func TestFoldFixupFolds(t *testing.T) {
	t.Parallel()
	fx := ffNewFx(t)
	fx.later()
	fx.g.write(fx.repo+"/task.txt", "task fixed")
	code, out, errs := fx.run(t, fx.repo, fx.task, fx.tasksMD, "task.txt")
	if code != 0 {
		t.Fatalf("exit %d, want 0\nstdout %s\nstderr %s", code, out, errs)
	}
	head := fx.g.git(fx.repo, "rev-parse", "HEAD")
	if want := "FOLDED: " + fx.repo + " — HEAD " + head + "\n"; out != want {
		t.Errorf("stdout %q, want %q", out, want)
	}
	if got := fx.g.git(fx.repo, "log", "--format=%s", fx.parent+"..HEAD"); got != "later\ntask" {
		t.Errorf("commits since the task's parent %q, want later then task", got)
	}
	if got := fx.g.git(fx.repo, "show", "HEAD^:task.txt"); got != "task fixed" {
		t.Errorf("the task commit carries %q, want the fix folded in", got)
	}
	fx.assertSettled(t)
}

func TestFoldFixupEmptyDrop(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		label   string
		later   bool
		subject string // the commits since the task's parent, newest first
	}{
		{"tip", false, ""},
		{"non-tip", true, "later"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			fx := ffNewFx(t)
			if tc.later {
				fx.later()
			}
			if err := os.Remove(fx.repo + "/task.txt"); err != nil {
				t.Fatal(err)
			}
			code, out, errs := fx.run(t, fx.repo, fx.task, fx.tasksMD, "task.txt")
			if code != 0 {
				t.Fatalf("exit %d, want 0\nstdout %s\nstderr %s", code, out, errs)
			}
			head := fx.g.git(fx.repo, "rev-parse", "HEAD")
			if want := "DROPPED: " + fx.repo + " — the fold emptied " + fx.task + "; HEAD " + head + "\n"; out != want {
				t.Errorf("stdout %q, want %q", out, want)
			}
			if got := fx.g.git(fx.repo, "log", "--format=%s", fx.parent+"..HEAD"); got != tc.subject {
				t.Errorf("commits since the task's parent %q, want %q", got, tc.subject)
			}
			if _, err := os.Stat(fx.repo + "/task.txt"); !os.IsNotExist(err) {
				t.Errorf("task.txt still present: %v", err)
			}
			fx.assertSettled(t)
		})
	}
}

func TestFoldFixupConflictLeftInProgress(t *testing.T) {
	t.Parallel()
	fx := ffNewFx(t)
	// later rewrites the task's own line, so a fix of that line cannot
	// replay onto the task commit without a hand resolution.
	fx.g.write(fx.repo+"/task.txt", "later")
	fx.g.git(fx.repo, "commit", "-qam", "later")
	fx.g.write(fx.repo+"/task.txt", "fixed")
	code, out, errs := fx.run(t, fx.repo, fx.task, fx.tasksMD, "task.txt")
	if code != 3 {
		t.Fatalf("exit %d, want 3\nstdout %s\nstderr %s", code, out, errs)
	}
	if !strings.HasPrefix(out, "CONFLICT: "+fx.repo+" — unmerged: task.txt;") {
		t.Errorf("stdout %q, want the CONFLICT line naming task.txt", out)
	}
	if !fx.rebasing() {
		t.Fatal("no rebase in progress: the conflict was resolved rather than left")
	}
	if top := fx.g.git(fx.repo, "stash", "list", "--format=%gs"); !strings.Contains(top, apaMarker) {
		t.Errorf("stash list %q, want the aside kept", top)
	}

	// --finish before the resolution still answers conflict.
	if code, out, errs := fx.runAsIs(t, "--finish", fx.repo, fx.task, fx.tasksMD); code != 3 {
		t.Fatalf("--finish mid-conflict: exit %d, want 3\nstdout %s\nstderr %s", code, out, errs)
	}

	// The hand resolution: take the replayed side at each stop, then
	// continue, until the rebase has finished.
	for i := 0; fx.rebasing() && i < 4; i++ {
		fx.g.git(fx.repo, "checkout", "--theirs", "task.txt")
		fx.g.git(fx.repo, "add", "task.txt")
		// A continue that stops at the next conflict exits 1: not a
		// fixture failure, so it runs outside fx.g.
		cont := exec.Command(fixtureGit, "-C", fx.repo, "-c", "core.editor=true", "rebase", "--continue")
		cont.Env = append(os.Environ(), fixtureGitEnv...)
		_ = cont.Run()
	}
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	code, out, errs = fx.runAsIs(t, "--finish", fx.repo, fx.task, fx.tasksMD)
	if code != 0 {
		t.Fatalf("--finish: exit %d, want 0\nstdout %s\nstderr %s", code, out, errs)
	}
	head := fx.g.git(fx.repo, "rev-parse", "HEAD")
	if want := "FOLDED: " + fx.repo + " — HEAD " + head + "\n"; out != want {
		t.Errorf("stdout %q, want %q", out, want)
	}
	if got := fx.g.git(fx.repo, "show", "HEAD^:task.txt"); got != "fixed" {
		t.Errorf("the task commit carries %q, want the fix folded in", got)
	}
	fx.assertSettled(t)
}

func TestFoldFixupGuardRefusal(t *testing.T) {
	t.Parallel()
	fx := ffNewFx(t)
	// A commit on main resolves but is no ancestor of demo's HEAD.
	fx.advanceBase("unrelated1.txt")
	stray := fx.g.git(fx.repo, "rev-parse", "main")
	head := fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.write(fx.repo+"/task.txt", "task fixed")
	code, out, errs := fx.run(t, fx.repo, stray, fx.tasksMD, "task.txt")
	if code != 1 {
		t.Fatalf("exit %d, want 1\nstdout %s\nstderr %s", code, out, errs)
	}
	if !strings.Contains(errs, "guard-autosquash: ") || !strings.Contains(errs, stray) {
		t.Errorf("stderr %q, want guard-autosquash's refusal naming %s", errs, stray)
	}
	if got := fx.g.git(fx.repo, "rev-parse", "HEAD"); got != head {
		t.Errorf("HEAD moved to %s: a refused fold committed", got)
	}
}

// ffForeign pushes a marker stash as another worktree's aside would: the
// stash list is shared by every worktree of the repository.
func (fx *ffFx) ffForeign() string {
	if fx.g.err == nil {
		fx.g.err = os.MkdirAll(fx.repo+"/spectre/changes/other", 0o755)
	}
	fx.g.write(fx.repo+"/spectre/changes/other/tasks.md", "other worktree's plan")
	fx.g.git(fx.repo, "stash", "push", "--include-untracked", "-m", apaMarker+": planning paths set aside for a rebase", "--", "spectre/changes/other")
	return fx.g.git(fx.repo, "rev-parse", "stash@{0}")
}

func TestFoldFixupNeverPopsAnotherAside(t *testing.T) {
	t.Parallel()
	t.Run("a clean fold leaves another worktree's aside on top", func(t *testing.T) {
		t.Parallel()
		fx := ffNewFx(t)
		fx.later()
		other := fx.ffForeign()
		fx.g.write(fx.repo+"/task.txt", "task fixed")
		code, out, errs := fx.runAsIs(t, fx.repo, fx.task, fx.tasksMD, "task.txt")
		if code != 0 {
			t.Fatalf("exit %d, want 0\nstdout %s\nstderr %s", code, out, errs)
		}
		if top := fx.g.git(fx.repo, "rev-parse", "stash@{0}"); top != other {
			t.Errorf("top stash %s, want the other worktree's %s left in place", top, other)
		}
		if _, err := os.Stat(fx.repo + "/spectre/changes/other/tasks.md"); err == nil {
			t.Errorf("the other worktree's planning file was popped into this one")
		}
	})
	t.Run("--finish refuses when another aside sits on top of its own", func(t *testing.T) {
		t.Parallel()
		fx := ffNewFx(t)
		fx.g.write(fx.repo+"/task.txt", "later")
		fx.g.git(fx.repo, "commit", "-qam", "later")
		fx.g.write(fx.repo+"/task.txt", "fixed")
		if code, out, errs := fx.run(t, fx.repo, fx.task, fx.tasksMD, "task.txt"); code != 3 {
			t.Fatalf("exit %d, want 3\nstdout %s\nstderr %s", code, out, errs)
		}
		for i := 0; fx.rebasing() && i < 4; i++ {
			fx.g.git(fx.repo, "checkout", "--theirs", "task.txt")
			fx.g.git(fx.repo, "add", "task.txt")
			cont := exec.Command(fixtureGit, "-C", fx.repo, "-c", "core.editor=true", "rebase", "--continue")
			cont.Env = append(os.Environ(), fixtureGitEnv...)
			_ = cont.Run()
		}
		other := fx.ffForeign()
		code, out, errs := fx.runAsIs(t, "--finish", fx.repo, fx.task, fx.tasksMD)
		if code != 2 || !strings.Contains(errs, "not the top stash") {
			t.Fatalf("--finish: exit %d stdout %q stderr %q, want 2 naming its own aside", code, out, errs)
		}
		if top := fx.g.git(fx.repo, "rev-parse", "stash@{0}"); top != other {
			t.Errorf("top stash %s, want the other worktree's %s untouched", top, other)
		}
	})
}
