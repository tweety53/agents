package guard

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// reshape-branch runs in-process, followed by the implementation commit
// integrate makes, calling checkChangeWorktreeLocation.

type rbFx struct{ main, wt, base string }

// rbNew is the harness's new_repo: a repo on main with one seed commit that
// gitignores spectre/changes/, and a change worktree at
// <main>/.worktrees/demo on spectre/demo.
func rbNew(t *testing.T) rbFx {
	t.Helper()
	main := t.TempDir() + "/repo"
	fx := rbFx{main: main, wt: main + "/.worktrees/demo"}
	splitRepo(t, fx.main, fx.wt, "main", map[string]string{
		"src/a.txt": "seed\n", ".gitignore": "spectre/changes/\n",
	})
	fx.base = splitGit(t, fx.main, nil, "rev-parse", "HEAD")
	mkdir(t, fx.wt+"/spectre/changes/demo")
	return fx
}

// commitAs is a pathspec-scoped commit by <author>.
func (fx rbFx) commitAs(t *testing.T, author, msg string, paths ...string) {
	t.Helper()
	splitGit(t, fx.wt, nil, append([]string{"add", "-A", "--"}, paths...)...)
	splitGit(t, fx.wt, nil, append([]string{"-c", "user.name=" + author, "-c", "user.email=" + author + "@example.com",
		"commit", "-q", "-m", msg, "--"}, paths...)...)
}

// expectedTree is the tree the branch plus working tree holds right now.
func (fx rbFx) expectedTree(t *testing.T) string {
	t.Helper()
	gitDir := splitGit(t, fx.wt, nil, "rev-parse", "--path-format=absolute", "--git-dir")
	b, err := os.ReadFile(gitDir + "/index")
	if err != nil {
		t.Fatal(err)
	}
	idx := t.TempDir() + "/expected-index"
	writeFile(t, idx, string(b))
	env := []string{"GIT_INDEX_FILE=" + idx}
	splitGit(t, fx.wt, env, "add", "-A")
	return splitGit(t, fx.wt, env, "write-tree")
}

func (fx rbFx) reshape(t *testing.T, dir string, args ...string) (int, string) {
	t.Helper()
	return runSplitGuard(t, reshapeBranch, dir, args...)
}

func TestReshapeBranch(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	check := func(t *testing.T, ok bool, label, format string, a ...any) {
		t.Helper()
		if !ok {
			t.Errorf(label+": "+format, a...)
		}
	}

	t.Run("1 every commit folds; the landed branch is one implementation commit", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		w := fx.wt
		writeFile(t, w+"/spectre/changes/demo/proposal.md", "proposal\n")
		writeFile(t, w+"/src/a.txt", "task 1\n")
		fx.commitAs(t, "impl", "feat(src): task 1\n\nTask-Id: 1", "src")
		writeFile(t, w+"/src/b.txt", "task 2\n")
		fx.commitAs(t, "impl", "feat(src): task 2\n\nTask-Id: 2", "src")
		writeFile(t, w+"/src/b.txt", "task 2\nfixup\n")
		fx.commitAs(t, "impl", "fixup! feat(src): task 2", "src")
		writeFile(t, w+"/src/a.txt", "task 1\noperator edit\n")
		writeFile(t, w+"/spectre/changes/demo/narrative.md", "narrative\n")
		wantTree := fx.expectedTree(t)

		rc, out := fx.reshape(t, w, w, "demo", fx.base)
		check(t, rc == 0, "reshape exit", "rc=%d out=%s", rc, out)
		check(t, strings.HasSuffix(out, "\nRESHAPED: "+w+" — reset to "+fx.base[:12]), "reshape verdict", "%q", out)
		check(t, splitGit(t, w, nil, "rev-parse", "HEAD") == fx.base, "HEAD is the merge base", "HEAD moved elsewhere")
		splitGit(t, w, nil, "add", "-A")
		splitGit(t, w, nil, "commit", "-q", "-m", "feat(src): the change")

		subjects := splitGit(t, w, nil, "log", "--reverse", "--format=%s", fx.base+"..HEAD")
		check(t, subjects == "feat(src): the change", "one implementation commit and no planning commit", "subjects:\n%s", subjects)
		check(t, splitGit(t, w, nil, "rev-parse", "HEAD^{tree}") == wantTree,
			"final tree equals the pre-reshape branch plus working tree", "tree differs")
		impl := splitGit(t, w, nil, "show", "--name-only", "--format=", "HEAD")
		check(t, impl == "src/a.txt\nsrc/b.txt", "implementation commit carries src alone", "%q", impl)
		check(t, isFile(w+"/spectre/changes/demo/proposal.md") && isFile(w+"/spectre/changes/demo/narrative.md"),
			"the ignored change directory stays on disk", "change directory lost")
	})

	t.Run("2 no planning commits: HEAD at the base, task work staged", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		writeFile(t, fx.wt+"/src/a.txt", "task\n")
		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
		rc, out := fx.reshape(t, fx.wt, fx.wt, "demo", fx.base)
		check(t, rc == 0 && strings.HasSuffix(out, "\nRESHAPED: "+fx.wt+" — reset to "+fx.base[:12]), "reshape verdict", "rc=%d out=%s", rc, out)
		check(t, splitGit(t, fx.wt, nil, "rev-parse", "HEAD") == fx.base, "HEAD is the merge base", "HEAD moved elsewhere")
		check(t, splitGit(t, fx.wt, nil, "diff", "--cached", "--name-only", "--", "src/a.txt") == "src/a.txt",
			"task work kept staged", "task work lost from the index")
	})

	t.Run("3 the guard refuses a main checkout and HEAD does not move", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		before := splitGit(t, fx.main, nil, "rev-parse", "HEAD")
		rc, out := fx.reshape(t, fx.main, fx.main, "demo", fx.base)
		check(t, rc == 1, "main checkout refused with exit 1", "rc=%d out=%s", rc, out)
		check(t, strings.Contains(out, "WORKTREE-LOCATION-MAIN-CHECKOUT"), "guard's own line printed", "out=%s", out)
		check(t, splitGit(t, fx.main, nil, "rev-parse", "HEAD") == before, "HEAD untouched", "HEAD moved")
	})

	t.Run("3b a merged-in branch's commits stop the reshape and HEAD does not move", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		w := fx.wt
		splitGit(t, fx.main, nil, "checkout", "-q", "-b", "develop")
		writeFile(t, fx.main+"/src/d.txt", "develop\n")
		splitGit(t, fx.main, nil, "add", "-A", "--", "src")
		splitGit(t, fx.main, nil, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "-m", "feat(src): develop work")
		writeFile(t, w+"/src/a.txt", "task\n")
		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
		splitGit(t, w, nil, "-c", "user.name=impl", "-c", "user.email=impl@example.com", "merge", "-q", "--no-ff", "-m", "chore: merge develop", "develop")
		before := splitGit(t, w, nil, "rev-parse", "HEAD")
		rc, out := fx.reshape(t, w, w, "demo", fx.base)
		check(t, rc == 1, "a foreign merge exits 1", "rc=%d out=%s", rc, out)
		check(t, strings.Contains(out, "RESHAPE-FOREIGN-MERGE: "+w+" — 1 commit(s)") && strings.Contains(out, "feat(src): develop work"),
			"the verdict names the merged-in commit", "out=%s", out)
		check(t, !strings.Contains(out, "feat(src): task"), "the change's own commit is not named", "out=%s", out)
		check(t, splitGit(t, w, nil, "rev-parse", "HEAD") == before, "HEAD untouched", "HEAD moved")
	})

	t.Run("3c a sync merge of the base the merge base moved to reshapes", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		w := fx.wt
		writeFile(t, fx.main+"/src/m.txt", "base moved\n")
		splitGit(t, fx.main, nil, "add", "-A", "--", "src")
		splitGit(t, fx.main, nil, "-c", "user.name=dev", "-c", "user.email=dev@example.com", "commit", "-q", "-m", "feat(src): base work")
		moved := splitGit(t, fx.main, nil, "rev-parse", "HEAD")
		writeFile(t, w+"/src/a.txt", "task\n")
		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
		splitGit(t, w, nil, "-c", "user.name=impl", "-c", "user.email=impl@example.com", "merge", "-q", "--no-ff", "-m", "chore: merge main", "main")
		rc, out := fx.reshape(t, w, w, "demo", moved)
		check(t, rc == 0 && strings.HasSuffix(out, "\nRESHAPED: "+w+" — reset to "+moved[:12]), "reshape verdict", "rc=%d out=%s", rc, out)
	})

	t.Run("4 a base that does not resolve cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		rc, out := fx.reshape(t, fx.wt, fx.wt, "demo", "no-such-ref")
		check(t, rc == 2, "unresolvable base exits 2", "rc=%d out=%s", rc, out)
		check(t, strings.HasSuffix(out, "\nreshape-branch: base does not resolve to a commit: no-such-ref"), "named reason", "out=%q", out)
	})

	t.Run("usage: not three arguments cannot answer", func(t *testing.T) {
		t.Parallel()
		rc, out := runSplitGuard(t, reshapeBranch, t.TempDir(), "a", "b")
		check(t, rc == 2 && out == "usage: flow-guard reshape-branch <worktree> <name> <merge-base>", "usage", "rc=%d out=%q", rc, out)
	})
}

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

// splitRepo copies an unborn repository on branch (git's default when empty)
// to main, gives it a local identity and a seed commit of files, and adds a linked worktree at wt on
// spectre/demo — the only place the location guard lets either guard commit.
func splitRepo(t *testing.T, main, wt, branch string, files map[string]string) {
	t.Helper()
	fixtureRepo(t, main, repoShape{Branch: branch})
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
