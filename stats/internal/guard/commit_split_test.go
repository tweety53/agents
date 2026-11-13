package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// Every case of scripts/test-commit-split.sh, one subtest per case, each
// assertion named after the harness's ok: label. The harness built a
// sandboxed repository per case and ran the script with stdout and stderr
// merged; here commit-split runs in-process over the same repositories, with
// checkChangeWorktreeLocation in-process too.

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
func runSplitGuard(t *testing.T, fn Func, dir string, args ...string) (int, string) {
	t.Helper()
	env := Env{Dir: dir, Getenv: os.Getenv}
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
		"README.md": "seed\n", ".gitignore": "spectre/changes/\n", "spectre/specs/seed.md": "seed\n",
	})
	return fx
}

func (fx csFx) split(t *testing.T, dir, c string) (int, string) {
	t.Helper()
	return runSplitGuard(t, commitSplit, dir, dir, "demo", "impl: "+c, "plan: "+c)
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
	smoke.Heavy(t)
	t.Parallel()
	check := func(t *testing.T, ok bool, label, format string, a ...any) {
		t.Helper()
		if !ok {
			t.Errorf(label+": "+format, a...)
		}
	}
	has := strings.Contains

	t.Run("1 the ignored change directory is in neither commit", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/src.md", "impl change\n")
		writeFile(t, fx.wt+"/spectre/changes/demo/plan.md", "plan change\n")
		rc, out := fx.split(t, fx.wt, "case1")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case1"), "implementation commit made", "%s", s)
		check(t, !has(s, "plan: case1"), "planning commit skipped", "%s", s)
		tr := splitGit(t, fx.wt, nil, "log", "--format=%s|%(trailers:key=Flow-Key,valueonly)", "-1")
		check(t, has(tr, "impl: case1|flow.commit-two"), "implementation commit carries Flow-Key", "%s", tr)
		impl := fx.files(t, fx.shaOf(t, "impl: case1"))
		check(t, !impl["spectre/changes/demo/plan.md"], "no change-directory file in the implementation commit", "%v", impl)
		check(t, isFile(fx.wt+"/spectre/changes/demo/plan.md"), "the change directory stays on disk", "plan.md lost")
	})

	t.Run("1b an unignored change directory refuses before anything is staged", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		fx := csFx{root + "/main", root + "/wt"}
		splitRepo(t, fx.main, fx.wt, map[string]string{"README.md": "seed\n", "spectre/specs/seed.md": "seed\n"})
		writeFile(t, fx.wt+"/src.md", "impl change\n")
		writeFile(t, fx.wt+"/spectre/changes/demo/plan.md", "plan change\n")
		rc, out := fx.split(t, fx.wt, "case1b")
		check(t, rc == 2, "exit 2", "rc=%d out=%s", rc, out)
		check(t, has(out, "spectre/changes/ is not gitignored"), "the refusal names the path", "%s", out)
		check(t, fx.subjects(t) == "seed", "no commit made", "%s", fx.subjects(t))
	})

	t.Run("1c each trailer argument joins the implementation commit's trailer block", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/src.md", "impl change\n")
		rc, out := runSplitGuard(t, commitSplit, fx.wt, fx.wt, "demo", "impl: case1c", "plan: case1c", "Task-Id: 1", "Task-Id: 2")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		tr := splitGit(t, fx.wt, nil, "log", "-1", "--format=%(trailers:key=Task-Id,valueonly,separator=%x2C)|%(trailers:key=Flow-Key,valueonly)")
		check(t, tr == "1,2|flow.commit-two", "trailers", "%q", tr)
	})

	t.Run("2 nothing is committed when only the change directory changed", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/spectre/changes/demo/only.md", "plan only\n")
		rc, out := fx.split(t, fx.wt, "case2")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, s == "seed", "no commit made", "%s", s)
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

	t.Run("5 a capability spec under spectre/specs/ is implementation", func(t *testing.T) {
		t.Parallel()
		fx := csNew(t)
		writeFile(t, fx.wt+"/spectre/specs/greeting.md", "spec change\n")
		rc, out := fx.split(t, fx.wt, "case5")
		check(t, rc == 0, "exit", "rc=%d out=%s", rc, out)
		s := fx.subjects(t)
		check(t, has(s, "impl: case5"), "capability spec committed as implementation", "%s", s)
		check(t, !has(s, "plan: case5"), "planning commit skipped", "%s", s)
		sha := fx.shaOf(t, "impl: case5")
		check(t, fx.files(t, sha)["spectre/specs/greeting.md"], "spec file is in the implementation commit", "sha=%q", sha)
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
		check(t, has(out, "WORKTREE-LOCATION-MAIN-CHECKOUT"), "guard's own line surfaces", "got %s", out)
		check(t, splitGit(t, fx.main, nil, "rev-parse", "HEAD") == before, "no commit made", "a commit landed")
		check(t, splitGit(t, fx.main, nil, "diff", "--cached", "--name-only") == "", "nothing staged", "index touched")
	})

	t.Run("usage: fewer than four arguments cannot answer", func(t *testing.T) {
		t.Parallel()
		rc, out := runSplitGuard(t, commitSplit, t.TempDir(), "a", "b", "c")
		check(t, rc == 2 && out == "usage: flow-guard commit-split <worktree> <name> <impl-msg> <plan-msg> [<trailer>…]", "usage", "rc=%d out=%q", rc, out)
	})
}

// A change planned under the pre-kan-1073 tooling committed its own
// directory; reshape-branch's soft reset leaves it staged, and the ignore
// rule alone never unstages a tracked path.
func TestCommitSplitUntracksTheChangeDirectory(t *testing.T) {
	t.Parallel()
	fx := csNew(t)
	writeFile(t, fx.wt+"/spectre/changes/demo/plan.md", "plan\n")
	writeFile(t, fx.wt+"/spectre/changes/other/keep.md", "other\n")
	splitGit(t, fx.wt, nil, "add", "-f", "spectre/changes")
	splitGit(t, fx.wt, nil, "commit", "-q", "-m", "old planning commit")
	writeFile(t, fx.wt+"/spectre/changes/demo/plan.md", "plan, edited after the planning commit\n")
	splitGit(t, fx.wt, nil, "reset", "-q", "--soft", "HEAD~1")
	writeFile(t, fx.wt+"/src.md", "impl change\n")
	rc, out := fx.split(t, fx.wt, "untrack")
	if rc != 0 {
		t.Fatalf("exit: rc=%d out=%s", rc, out)
	}
	impl := fx.files(t, fx.shaOf(t, "impl: untrack"))
	if !impl["src.md"] {
		t.Errorf("src.md missing from the implementation commit: %v", impl)
	}
	if impl["spectre/changes/demo/plan.md"] {
		t.Errorf("the change's own plan.md is in the implementation commit: %v", impl)
	}
	if got := splitGit(t, fx.wt, nil, "ls-files", "spectre/changes/demo"); got != "" {
		t.Errorf("the change directory is still tracked: %q", got)
	}
	if !isFile(fx.wt + "/spectre/changes/demo/plan.md") {
		t.Errorf("plan.md lost from disk")
	}
	if !impl["spectre/changes/other/keep.md"] {
		t.Errorf("another change's tracked file was untracked: %v", impl)
	}
}
