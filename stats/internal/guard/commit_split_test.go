package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Every case of scripts/test-commit-split.sh, one subtest per case, each
// assertion named after the harness's ok: label. The harness built a
// sandboxed repository per case and ran the script with stdout and stderr
// merged; here commit-split runs in-process over the same repositories, with
// the real check-planning-commit-location.sh beside it as the shim exports
// FLOW_GUARD_SELF.

// splitGit runs fixture git in dir under fixtureGitEnv plus extra, failing
// the test on error, and returns stdout with trailing newlines stripped.
func splitGit(t *testing.T, dir string, extra []string, args ...string) string {
	t.Helper()
	cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(append(os.Environ(), fixtureGitEnv...), extra...)
	var errb bytes.Buffer
	cmd.Stderr = &errb
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, errb.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// splitRepo initialises main with the given branch flags, a local identity
// and a seed commit of files, and adds a linked worktree at wt on
// spectre/demo — the only place the location guard lets either guard commit.
func splitRepo(t *testing.T, main, wt string, files map[string]string, initArgs ...string) {
	t.Helper()
	splitGit(t, ".", nil, append(append([]string{"init", "-q"}, initArgs...), main)...)
	for _, kv := range [][2]string{{"user.email", "test@example.com"}, {"user.name", "Test"}, {"commit.gpgsign", "false"}} {
		splitGit(t, main, nil, "config", kv[0], kv[1])
	}
	for p, body := range files {
		writeFile(t, main+"/"+p, body)
	}
	splitGit(t, main, nil, "add", "-A")
	splitGit(t, main, nil, "commit", "-q", "-m", "seed")
	splitGit(t, main, nil, "worktree", "add", "-q", "-b", "spectre/demo", wt)
}

// runSplitGuard runs guard fn in-process from dir as its shim would, stdout
// and stderr merged in write order with trailing newlines stripped.
func runSplitGuard(t *testing.T, fn Func, shim, dir string, args ...string) (int, string) {
	t.Helper()
	self := tcfScriptsDir(t) + "/" + shim
	env := Env{Dir: dir, Getenv: func(k string) string {
		if k == "FLOW_GUARD_SELF" {
			return self
		}
		return os.Getenv(k)
	}}
	var all bytes.Buffer
	rc := fn(args, env, &all, &all)
	return rc, strings.TrimRight(all.String(), "\n")
}

type csFx struct{ main, wt string }

func csNew(t *testing.T) csFx {
	t.Helper()
	root := t.TempDir()
	fx := csFx{root + "/main", root + "/wt"}
	splitRepo(t, fx.main, fx.wt, map[string]string{
		"README.md": "seed\n", "spectre/changes/seed.md": "seed\n", "spectre/specs/seed.md": "seed\n",
	})
	return fx
}

func (fx csFx) split(t *testing.T, dir, c string) (int, string) {
	t.Helper()
	return runSplitGuard(t, commitSplit, "commit-split.sh", dir, dir, "demo", "impl: "+c, "plan: "+c)
}

func (fx csFx) subjects(t *testing.T) string { return splitGit(t, fx.wt, nil, "log", "--format=%s") }

// shaOf is the harness's `log --format='%H %s' | awk '/<subject>/'`.
func (fx csFx) shaOf(t *testing.T, subject string) string {
	for _, l := range strings.Split(splitGit(t, fx.wt, nil, "log", "--format=%H %s"), "\n") {
		if sha, s, _ := strings.Cut(l, " "); s == subject {
			return sha
		}
	}
	return ""
}

// files is `git show --name-only --format= <sha>`, one entry per path; an
// empty sha (no such commit) lists nothing.
func (fx csFx) files(t *testing.T, sha string) map[string]bool {
	set := map[string]bool{}
	if sha == "" {
		return set
	}
	for _, p := range strings.Split(splitGit(t, fx.wt, nil, "show", "--name-only", "--format=", sha), "\n") {
		set[p] = true
	}
	return set
}

func TestCommitSplit(t *testing.T) {
	t.Parallel()
	check := func(t *testing.T, ok bool, label, format string, a ...any) {
		t.Helper()
		if !ok {
			t.Errorf(label+": "+format, a...)
		}
	}
	has := strings.Contains

	t.Run("1 both commits happen when both staging areas have changes", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/src.md", "impl change\n")
		writeFile(t, fx.wt+"/spectre/changes/plan.md", "plan change\n")
		rc, out := fx.split(t, fx.wt, "case1")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case1"), "implementation commit made", "%s", s)
		check(t, has(s, "plan: case1"), "planning commit made", "%s", s)
		n := len(strings.Split(splitGit(t, fx.wt, nil, "log", "--oneline"), "\n"))
		check(t, n == 3, "exactly two new commits on top of seed", "got %d commits", n)
	})

	t.Run("2 implementation commit skipped when only planning paths changed", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/spectre/changes/only.md", "plan only\n")
		rc, out := fx.split(t, fx.wt, "case2")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, !has(s, "impl: case2"), "implementation commit skipped", "%s", s)
		check(t, has(s, "plan: case2"), "planning commit made", "%s", s)
	})

	t.Run("3 planning commit skipped when only implementation paths changed", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/only.md", "impl only\n")
		rc, out := fx.split(t, fx.wt, "case3")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case3"), "implementation commit made", "%s", s)
		check(t, !has(s, "plan: case3"), "planning commit skipped", "%s", s)
	})

	t.Run("4 a tracked symlink at spectre/ stops the split with git's exit 128", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		if err := os.RemoveAll(fx.wt + "/spectre"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("docs", fx.wt+"/spectre"); err != nil {
			t.Fatal(err)
		}
		splitGit(t, fx.wt, nil, "add", "-A")
		splitGit(t, fx.wt, nil, "commit", "-q", "-m", "spectre becomes a tracked symlink")
		before := splitGit(t, fx.wt, nil, "rev-parse", "HEAD")
		writeFile(t, fx.wt+"/only.md", "impl change\n")
		rc, out := fx.split(t, fx.wt, "case4")
		check(t, rc == 128, "exits 128", "rc=%d out=%s", rc, out)
		check(t, splitGit(t, fx.wt, nil, "rev-parse", "HEAD") == before, "no commit made", "out=%s", out)
		check(t, has(out, "symbolic"), "git's own message surfaces", "got %s", out)
	})

	t.Run("5 a capability spec under spectre/specs/ is implementation", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/spectre/specs/greeting.md", "spec change\n")
		rc, out := fx.split(t, fx.wt, "case5")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case5"), "capability spec committed as implementation", "%s", s)
		check(t, !has(s, "plan: case5"), "planning commit skipped", "%s", s)
		// Resolved by SUBJECT, not by HEAD: with the exclusion pathspec
		// widened back to spectre/ the implementation commit is SKIPPED and
		// HEAD is the planning commit — which carries the spec too, so a
		// HEAD-relative assertion would pass for the wrong reason.
		sha := fx.shaOf(t, "impl: case5")
		check(t, fx.files(t, sha)["spectre/specs/greeting.md"], "spec file is in the implementation commit", "sha=%q", sha)
	})

	t.Run("6 link.md is implementation, its siblings planning", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		d := fx.wt + "/spectre/changes/kan-363/"
		for _, f := range []string{"link", "proposal", "design", "tasks"} {
			writeFile(t, d+f+".md", f+"\n")
		}
		rc, out := fx.split(t, fx.wt, "case6")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case6"), "implementation commit made", "%s", s)
		check(t, has(s, "plan: case6"), "planning commit made", "%s", s)
		impl, plan := fx.files(t, fx.shaOf(t, "impl: case6")), fx.files(t, fx.shaOf(t, "plan: case6"))
		const p = "spectre/changes/kan-363/"
		check(t, impl[p+"link.md"], "link.md is in the implementation commit", "%v", impl)
		check(t, !impl[p+"proposal.md"] && !impl[p+"design.md"] && !impl[p+"tasks.md"],
			"no planning file in the implementation commit", "%v", impl)
		check(t, !plan[p+"link.md"], "link.md is not in the planning commit", "%v", plan)
		for _, f := range []string{"proposal", "design", "tasks"} {
			check(t, plan[p+f+".md"], f+".md is in the planning commit", "%v", plan)
		}
	})

	t.Run("6b link.md under a dot directory is not re-added, as the shell glob skipped it", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/spectre/changes/.hidden/link.md", "link\n")
		rc, out := fx.split(t, fx.wt, "case6b")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, !has(s, "impl: case6b"), "implementation commit skipped", "%s", s)
		plan := fx.files(t, fx.shaOf(t, "plan: case6b"))
		check(t, plan["spectre/changes/.hidden/link.md"], "dot-directory link.md lands as planning", "%v", plan)
	})

	t.Run("7 docs/research/ is implementation", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/docs/research/note.md", "note\n")
		rc, out := fx.split(t, fx.wt, "case7")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case7"), "implementation commit made", "%s", s)
		check(t, !has(s, "plan: case7"), "planning commit skipped", "%s", s)
		sha := fx.shaOf(t, "impl: case7")
		check(t, fx.files(t, sha)["docs/research/note.md"], "docs/research file is in the implementation commit", "sha=%q", sha)
	})

	t.Run("8 the main checkout is refused before anything is staged", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.main+"/only.md", "impl change\n")
		writeFile(t, fx.main+"/spectre/changes/plan.md", "plan change\n")
		before := splitGit(t, fx.main, nil, "rev-parse", "HEAD")
		rc, out := fx.split(t, fx.main, "case8")
		check(t, rc == 1, "main checkout exits 1", "rc=%d out=%s", rc, out)
		check(t, has(out, "PLANNING-COMMIT-MAIN-CHECKOUT"), "guard's own line surfaces", "got %s", out)
		check(t, splitGit(t, fx.main, nil, "rev-parse", "HEAD") == before, "no commit made", "a commit landed")
		check(t, splitGit(t, fx.main, nil, "diff", "--cached", "--name-only") == "", "nothing staged", "index touched")
	})

	t.Run("usage: fewer than four arguments cannot answer", func(t *testing.T) {
		t.Parallel()
		rc, out := runSplitGuard(t, commitSplit, "commit-split.sh", t.TempDir(), "a", "b", "c")
		check(t, rc == 2 && out == "usage: commit-split.sh <worktree> <name> <impl-msg> <plan-msg>", "usage", "rc=%d out=%q", rc, out)
	})
}
