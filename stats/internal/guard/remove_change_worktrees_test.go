package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// rcwFx is a main checkout whose change branch spectre/demo is merged into
// main and pushed to a bare origin, with the change's apply worktree checked
// out beside it -- the state run 2 reaches its cleanup step in.
type rcwFx struct {
	dir, repo, wt, mergeBase string
	g                        fxGit
}

func rcwNewFx(t *testing.T) *rcwFx {
	t.Helper()
	dir := t.TempDir()
	fx := &rcwFx{dir: dir, repo: dir + "/repo", wt: dir + "/wt/demo"}
	fx.g.git("", "init", "-q", "--bare", "-b", "main", dir+"/origin.git")
	fx.g.git("", "init", "-q", "-b", "main", fx.repo)
	for _, kv := range [][2]string{{"user.name", "Test"}, {"user.email", "test@example.invalid"}, {"commit.gpgsign", "false"}} {
		fx.g.git(fx.repo, "config", kv[0], kv[1])
	}
	fx.g.write(fx.repo+"/.gitignore", "*.log\nbuild/\n.env")
	fx.g.write(fx.repo+"/base.txt", "base")
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.mergeBase = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "remote", "add", "origin", dir+"/origin.git")
	fx.g.git(fx.repo, "push", "-q", "-u", "origin", "main")
	fx.g.git(fx.repo, "remote", "set-head", "origin", "main")
	fx.g.git(fx.repo, "checkout", "-q", "-b", "spectre/demo")
	fx.g.write(fx.repo+"/work.txt", "work")
	fx.g.git(fx.repo, "add", "work.txt")
	fx.g.git(fx.repo, "commit", "-qm", "work")
	fx.g.git(fx.repo, "push", "-q", "origin", "spectre/demo")
	fx.g.git(fx.repo, "checkout", "-q", "main")
	fx.g.git(fx.repo, "merge", "-q", "--ff-only", "spectre/demo")
	fx.g.git(fx.repo, "push", "-q", "origin", "main")
	fx.g.git(fx.repo, "worktree", "add", "-q", fx.wt, "spectre/demo")
	return fx
}

// copy adds a detached wave-group copy of the apply worktree, as
// implement.md makes one.
func (fx *rcwFx) copy() string {
	c := fx.wt + "-wave-group-2"
	fx.g.git(fx.repo, "worktree", "add", "-q", "--detach", c, "HEAD")
	return c
}

func (fx *rcwFx) stop(t *testing.T, body string) {
	t.Helper()
	mkdir(t, fx.repo+"/.flow")
	writeFile(t, fx.repo+"/.flow/project.md", "# project\n\n## stop\n\n"+body+"\n\n## run\n\nnothing\n")
}

// run runs remove-change-worktrees in-process with the real
// check-worktree-processes.sh beside it, as the shim exports FLOW_GUARD_SELF.
func (fx *rcwFx) run(t *testing.T, tweak func(*Env), args ...string) (int, string, string) {
	t.Helper()
	if fx.g.err != nil {
		t.Fatal(fx.g.err)
	}
	self := tcfScriptsDir(t) + "/remove-change-worktrees.sh"
	env := Env{Dir: fx.dir, Getenv: func(k string) string {
		if k == "FLOW_GUARD_SELF" {
			return self
		}
		return os.Getenv(k)
	}}
	if tweak != nil {
		tweak(&env)
	}
	var out, errb bytes.Buffer
	code := removeChangeWorktrees(append([]string{fx.repo, "demo", fx.mergeBase}, args...), env, &out, &errb)
	return code, out.String(), errb.String()
}

func rcwExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (fx *rcwFx) hasRef(ref string) bool {
	return exec.Command(fixtureGit, "-C", fx.repo, "show-ref", "--verify", "--quiet", ref).Run() == nil
}

// assertUntouched is "any failed check leaves every worktree alone".
func (fx *rcwFx) assertUntouched(t *testing.T, paths ...string) {
	t.Helper()
	for _, p := range append([]string{fx.wt}, paths...) {
		if !rcwExists(p) {
			t.Errorf("%s was removed", p)
		}
	}
	if !fx.hasRef("refs/heads/spectre/demo") {
		t.Error("the local branch was deleted")
	}
	if !fx.hasRef("refs/remotes/origin/spectre/demo") {
		t.Error("the remote branch was deleted")
	}
}

func rcwLines(out, prefix string) []string {
	var got []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, prefix) {
			got = append(got, l)
		}
	}
	return got
}

func TestRemoveChangeWorktrees(t *testing.T) {
	t.Parallel()

	t.Run("clean removal", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.write(fx.wt+"/run.log", "log")
		mkdir(t, fx.wt+"/build")
		fx.g.write(fx.wt+"/build/out.o", "obj")
		// An image under a regeneratable directory is that directory's, not a
		// loose capture (KAN-860 F3).
		fx.g.write(fx.wt+"/build/icon.png", "png")
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REGENERATABLE: "); len(l) != 1 || !strings.HasSuffix(l[0], " — 3") {
			t.Errorf("REGENERATABLE lines %q, want one counting 3", l)
		}
		if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 0 {
			t.Errorf("UNCLASSIFIED lines %q, want none", l)
		}
		if l := rcwLines(out, "SKIPPED: "); len(l) != 0 {
			t.Errorf("SKIPPED lines %q, want none for an absent ## stop", l)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 2 || !strings.HasSuffix(l[0], "/wt/demo") || l[1] != "REMOVED: spectre/demo" {
			t.Errorf("REMOVED lines %q, want the worktree then the branch", l)
		}
		if l := rcwLines(out, "REMOTE-"); len(l) != 1 || l[0] != "REMOTE-DELETED: origin/spectre/demo" {
			t.Errorf("REMOTE lines %q, want REMOTE-DELETED", l)
		}
		if rcwExists(fx.wt) || fx.hasRef("refs/heads/spectre/demo") || fx.hasRef("refs/remotes/origin/spectre/demo") {
			t.Error("the worktree, the branch or the remote branch survived")
		}
		if exec.Command(fixtureGit, "--git-dir", fx.dir+"/origin.git", "show-ref", "--verify", "--quiet", "refs/heads/spectre/demo").Run() == nil {
			t.Error("origin still carries spectre/demo")
		}
	})

	t.Run("remote ref already gone", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.git("", "--git-dir", fx.dir+"/origin.git", "branch", "-D", "spectre/demo")
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOTE-"); len(l) != 1 || l[0] != "REMOTE-GONE: origin/spectre/demo" {
			t.Errorf("REMOTE lines %q, want REMOTE-GONE", l)
		}
		if fx.hasRef("refs/remotes/origin/spectre/demo") {
			t.Error("the stale tracking ref was not pruned")
		}
	})

	t.Run("a failed gate removes nothing", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.appendLine(fx.wt+"/work.txt", "dirty")
		code, out, _ := fx.run(t, nil)
		if code != 1 || len(rcwLines(out, "REFUSED: ")) != 1 || !strings.Contains(out, "check 1") {
			t.Fatalf("exit %d, stdout %q; want 1 and one check 1 REFUSED line", code, out)
		}
		if len(rcwLines(out, "REMOVED: "))+len(rcwLines(out, "REMOTE-")) != 0 {
			t.Errorf("a refusal still removed something:\n%s", out)
		}
		fx.assertUntouched(t)
	})

	t.Run("check 3 — a commit its upstream lacks", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.git(fx.wt, "branch", "-q", "--set-upstream-to=origin/spectre/demo")
		fx.g.write(fx.wt+"/late.txt", "late")
		fx.g.git(fx.wt, "add", "late.txt")
		fx.g.git(fx.wt, "commit", "-qm", "unpushed")
		code, out, _ := fx.run(t, nil)
		if code != 1 || len(rcwLines(out, "REFUSED: ")) != 1 || !strings.Contains(out, "check 3: commits not on origin/spectre/demo") {
			t.Fatalf("exit %d, stdout %q; want 1 and one check 3 REFUSED line", code, out)
		}
		fx.assertUntouched(t)
	})

	t.Run("check 6 — a live process holds the worktree", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		held := exec.Command("sleep", "60")
		held.Dir = fx.wt
		if err := held.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = held.Process.Kill(); _ = held.Wait() })
		code, out, _ := fx.run(t, nil)
		if code != 1 || len(rcwLines(out, "HELD: ")) != 1 {
			t.Fatalf("exit %d, stdout %q; want 1 and one HELD line", code, out)
		}
		if len(rcwLines(out, "REMOVED: "))+len(rcwLines(out, "REMOTE-")) != 0 {
			t.Errorf("a held worktree still removed something:\n%s", out)
		}
		fx.assertUntouched(t)
	})

	t.Run("a worktree already gone", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		if err := os.RemoveAll(fx.wt); err != nil {
			t.Fatal(err)
		}
		code, out, errb := fx.run(t, nil)
		if code != 0 {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 2 || !strings.HasSuffix(l[0], "/wt/demo — already gone") || l[1] != "REMOVED: spectre/demo" {
			t.Errorf("REMOVED lines %q, want the gone worktree then the branch", l)
		}
		if fx.hasRef("refs/heads/spectre/demo") {
			t.Error("the branch survived its gone worktree")
		}
	})

	t.Run("cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		for _, args := range [][]string{{fx.repo, "demo"}, {fx.repo, "../demo", "-"}, {fx.dir, "demo", "-"}, {fx.repo, "demo", "-", "--force"}} {
			var out, errb bytes.Buffer
			env := Env{Dir: fx.dir, Getenv: func(k string) string {
				if k == "FLOW_GUARD_SELF" {
					return tcfScriptsDir(t) + "/remove-change-worktrees.sh"
				}
				return os.Getenv(k)
			}}
			if code := removeChangeWorktrees(args, env, &out, &errb); code != 2 || out.String() != "" {
				t.Errorf("%q: exit %d, stdout %q; want 2 and nothing on stdout", args, code, out.String())
			}
		}
		fx.assertUntouched(t)
	})

	// The real shim: it exports FLOW_GUARD_SELF and asserts the sibling.
	t.Run("through the shim", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		cmd := exec.Command(tcfScriptsDir(t)+"/remove-change-worktrees.sh", fx.repo, "demo", "-")
		cmd.Dir = fx.dir
		cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
		out, err := cmd.Output()
		if err != nil || !strings.Contains(string(out), "REMOVED: spectre/demo\n") {
			t.Fatalf("shim: %v, stdout %q", err, out)
		}
	})
}

func TestRemoveChangeWorktreesDisclose(t *testing.T) {
	t.Parallel()

	t.Run("an unclassified entry", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.g.write(fx.wt+"/.env", "SECRET=1")
		code, out, _ := fx.run(t, nil)
		if code != 3 || len(rcwLines(out, "DISCLOSE: ")) == 0 {
			t.Fatalf("exit %d, stdout %q; want 3 and DISCLOSE lines", code, out)
		}
		if l := rcwLines(out, "UNCLASSIFIED: "); len(l) != 1 || !strings.HasSuffix(l[0], " — .env") {
			t.Errorf("UNCLASSIFIED lines %q, want .env", l)
		}
		fx.assertUntouched(t)

		code, out, errb := fx.run(t, nil, "--proceed")
		if code != 0 || rcwExists(fx.wt) || len(rcwLines(out, "DISCLOSE: ")) != 0 {
			t.Fatalf("--proceed: exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
	})

	t.Run("a wave-group copy", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		copyWT := fx.copy()
		fx.g.write(copyWT+"/scratch.txt", "scratch")
		code, out, _ := fx.run(t, nil)
		if code != 3 {
			t.Fatalf("exit %d, stdout %q; want 3", code, out)
		}
		d := strings.Join(rcwLines(out, "DISCLOSE: "), "\n")
		if !strings.Contains(d, "-wave-group-2") || !strings.Contains(d, "scratch.txt") || !strings.Contains(d, "work") {
			t.Errorf("DISCLOSE lines do not carry the copy's status and log:\n%s", d)
		}
		fx.assertUntouched(t, copyWT)

		code, out, errb := fx.run(t, nil, "--proceed")
		if code != 0 || rcwExists(fx.wt) || rcwExists(copyWT) {
			t.Fatalf("--proceed: exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if l := rcwLines(out, "REMOVED: "); len(l) != 3 || !strings.HasSuffix(l[0], "-wave-group-2") {
			t.Errorf("REMOVED lines %q, want the copy first, then the worktree and the branch", l)
		}
	})
}

func TestRemoveChangeWorktreesStopCommand(t *testing.T) {
	t.Parallel()

	t.Run("no fence: skipped", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		marker := fx.dir + "/ran"
		fx.stop(t, "Never stop the dev stack; an operator runs `touch "+marker+"` by hand.")
		code, out, errb := fx.run(t, nil)
		if code != 0 || rcwExists(fx.wt) {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if rcwExists(marker) {
			t.Error("prose outside a fence was run")
		}
		// KAN-860 F4: the skip is reported, never silent.
		if l := rcwLines(out, "SKIPPED: "); len(l) != 1 || l[0] != "SKIPPED: check 5 — ## stop declares no fenced command" {
			t.Errorf("SKIPPED lines %q, want the one check 5 skip", l)
		}
	})

	t.Run("fenced: every fence, read as ## worktree setup reads it", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		marker := fx.dir + "/ran"
		// KAN-860 F9: an indented fence and a second fence, as kwFenced reads them.
		fx.stop(t, "\t```bash\n\techo a >> "+marker+"\n\t```\n\nthen\n\n```bash\necho b >> "+marker+"\n```")
		code, out, errb := fx.run(t, nil)
		if code != 0 || rcwExists(fx.wt) {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if got := rbRead(marker); got != "a\nb\n" {
			t.Errorf("marker %q, want both fences run in order", got)
		}
		if l := rcwLines(out, "SKIPPED: "); len(l) != 0 {
			t.Errorf("SKIPPED lines %q, want none", l)
		}
	})

	t.Run("fenced: run, nothing else", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		marker, other := fx.dir+"/ran", fx.dir+"/other"
		fx.stop(t, "Stop it with:\n\n```bash\necho ran >> "+marker+"\n```\n\nNever `touch "+other+"`.")
		code, out, errb := fx.run(t, nil)
		if code != 0 || rcwExists(fx.wt) {
			t.Fatalf("exit %d\nstdout:\n%s\nstderr:\n%s", code, out, errb)
		}
		if got := rbRead(marker); got != "ran\n" {
			t.Errorf("marker %q, want the fenced command run exactly once", got)
		}
		if rcwExists(other) {
			t.Error("prose outside the fence was run")
		}
	})

	t.Run("fenced: a failing command is a failed check", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.stop(t, "```bash\nexit 3\n```")
		code, out, _ := fx.run(t, nil)
		if code != 1 || !strings.Contains(out, "check 5") {
			t.Fatalf("exit %d, stdout %q; want 1 and a check 5 refusal", code, out)
		}
		fx.assertUntouched(t)
	})

	t.Run("fenced: the 60-second bound", func(t *testing.T) {
		t.Parallel()
		fx := rcwNewFx(t)
		fx.stop(t, "```bash\nsleep 30\n```")
		expired := make(chan time.Time)
		close(expired)
		start := time.Now()
		code, out, _ := fx.run(t, func(e *Env) { e.SurvivorsExpire, e.SurvivorsKillGrace = expired, 100*time.Millisecond })
		if code != 1 || !strings.Contains(out, "check 5") || !strings.Contains(out, "timed out") {
			t.Fatalf("exit %d, stdout %q; want 1 and a check 5 timeout refusal", code, out)
		}
		if d := time.Since(start); d > 20*time.Second {
			t.Errorf("the bound did not cut the command short: %v", d)
		}
		fx.assertUntouched(t)
	})
}
