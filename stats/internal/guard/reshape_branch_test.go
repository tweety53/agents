package guard

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Every case of scripts/test-reshape-branch.sh, one subtest per case, each
// assertion named after the harness's ok: label. reshape-branch runs
// in-process, followed by commit-split exactly as integrate runs them, each
// calling check-planning-commit-location in-process.

type rbFx struct{ main, wt, base string }

// rbNew is the harness's new_repo: a repo on main with one seed commit and a
// change worktree at <main>/.worktrees/demo on spectre/demo.
func rbNew(t *testing.T) rbFx {
	t.Helper()
	main := t.TempDir() + "/repo"
	fx := rbFx{main: main, wt: main + "/.worktrees/demo"}
	splitRepo(t, fx.main, fx.wt, map[string]string{
		"src/a.txt": "seed\n", "spectre/changes/seed.md": "other change\n",
	}, "-b", "main")
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
	t.Parallel()
	check := func(t *testing.T, ok bool, label, format string, a ...any) {
		t.Helper()
		if !ok {
			t.Errorf(label+": "+format, a...)
		}
	}

	t.Run("1 planning commits survive; task and fixup commits collapse", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		w := fx.wt
		writeFile(t, w+"/spectre/changes/demo/proposal.md", "proposal\n")
		fx.commitAs(t, "planner", "chore(spectre): plan", "spectre/changes/demo")
		writeFile(t, w+"/src/a.txt", "task 1\n")
		fx.commitAs(t, "impl", "feat(src): task 1\n\nTask-Id: 1", "src")
		writeFile(t, w+"/spectre/changes/demo/link.md", "link\n")
		fx.commitAs(t, "linker", "chore(spectre): link peer", "spectre/changes/demo/link.md")
		writeFile(t, w+"/src/b.txt", "task 2\n")
		fx.commitAs(t, "impl", "feat(src): task 2\n\nTask-Id: 2", "src")
		writeFile(t, w+"/src/b.txt", "task 2\nfixup\n")
		fx.commitAs(t, "impl", "fixup! feat(src): task 2", "src")
		writeFile(t, w+"/spectre/changes/demo/tasks.md", "tasks ticked\n")
		fx.commitAs(t, "reviewer", "chore(spectre): plan", "spectre/changes/demo")
		writeFile(t, w+"/src/a.txt", "task 1\noperator edit\n")
		writeFile(t, w+"/spectre/changes/demo/narrative.md", "narrative\n")
		wantTree := fx.expectedTree(t)

		rc, out := fx.reshape(t, w, w, "demo", fx.base)
		check(t, rc == 0, "reshape exit", "rc=%d out=%s", rc, out)
		check(t, strings.Contains(out, "RESHAPED: "+w+" — 3 planning commit(s) kept on "+fx.base[:12]),
			"reshape verdict names three kept commits", "%s", out)
		if rc, out := runSplitGuard(t, commitSplit, w,
			w, "demo", "feat(src): the change", "chore(spectre): plan\n\noutstanding: none"); rc != 0 {
			t.Fatalf("commit-split rc=%d out=%s", rc, out)
		}

		subjects := splitGit(t, w, nil, "log", "--reverse", "--format=%s", fx.base+"..HEAD")
		want := "chore(spectre): plan\nchore(spectre): link peer\nchore(spectre): plan\nfeat(src): the change\nchore(spectre): plan"
		check(t, subjects == want, "planning commits kept, in order, around one implementation commit", "subjects:\n%s", subjects)
		authors := splitGit(t, w, nil, "log", "--reverse", "--format=%an", fx.base+"..HEAD~2")
		check(t, authors == "planner\nlinker\nreviewer", "planning commits keep their authors", "authors: %q", authors)
		check(t, splitGit(t, w, nil, "rev-parse", "HEAD^{tree}") == wantTree,
			"final tree equals the pre-reshape branch plus working tree", "tree differs")

		for _, c := range strings.Fields(splitGit(t, w, nil, "rev-list", fx.base+"..HEAD")) {
			paths := strings.Split(splitGit(t, w, nil, "diff-tree", "--no-commit-id", "--name-only", "-r", c), "\n")
			subject := splitGit(t, w, nil, "log", "-1", "--format=%s", c)
			for _, p := range paths {
				leak := strings.HasPrefix(subject, "chore") && !strings.HasPrefix(p, "spectre/changes/demo/") ||
					strings.HasPrefix(subject, "feat") && strings.HasPrefix(p, "spectre/changes/")
				check(t, !leak, "each commit carries only its own side of the split", "%s %q carries %s", c, subject, p)
			}
		}
		last := splitGit(t, w, nil, "show", "--name-only", "--format=", "HEAD")
		check(t, strings.Contains("\n"+last+"\n", "\nspectre/changes/demo/narrative.md\n"),
			"uncommitted planning delta is the last planning commit", "%s", last)
		check(t, strings.Contains(splitGit(t, w, nil, "log", "-1", "--format=%B", "HEAD"), "outstanding: none"),
			"last planning commit carries integrate's message", "message lost")
	})

	t.Run("1b a planning commit keeps its message byte for byte, trailing blank lines included", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		writeFile(t, fx.wt+"/spectre/changes/demo/proposal.md", "proposal\n")
		splitGit(t, fx.wt, nil, "add", "-A")
		splitGit(t, fx.wt, nil, "commit", "-q", "--cleanup=verbatim", "-m", "chore(spectre): plan\n\nbody\n\n")
		// The raw commit object, untrimmed: its message is everything after
		// the first blank line.
		msg := func() string {
			out, err := exec.Command(fixtureGit, "-C", fx.wt, "cat-file", "commit", "HEAD").Output()
			if err != nil {
				t.Fatal(err)
			}
			_, m, _ := strings.Cut(string(out), "\n\n")
			return m
		}
		want := msg()
		rc, out := fx.reshape(t, fx.wt, fx.wt, "demo", fx.base)
		check(t, rc == 0, "reshape exit", "rc=%d out=%s", rc, out)
		check(t, msg() == want, "message kept", "got %q, want %q", msg(), want)
	})

	t.Run("2 no planning commits: the reshape is a plain reset --soft", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		writeFile(t, fx.wt+"/src/a.txt", "task\n")
		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
		rc, out := fx.reshape(t, fx.wt, fx.wt, "demo", fx.base)
		check(t, rc == 0 && strings.Contains(out, "0 planning commit(s)"), "nothing kept", "rc=%d out=%s", rc, out)
		check(t, splitGit(t, fx.wt, nil, "rev-parse", "HEAD") == fx.base, "HEAD is the merge base", "HEAD moved elsewhere")
		check(t, splitGit(t, fx.wt, nil, "diff", "--cached", "--name-only", "--", "src/a.txt") == "src/a.txt",
			"task work kept staged", "task work lost from the index")
	})

	// A run 1 that stopped between `spectre archive` and commit-archive.sh
	// left the archive rename staged; a resumed reshape keeps it staged and
	// the working tree as found — only HEAD moves.
	t.Run("2b a staged archive rename survives the reshape staged", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		w := fx.wt
		writeFile(t, w+"/spectre/changes/demo/tasks.md", "- [x] 1. task\n")
		fx.commitAs(t, "planner", "chore(spectre): plan", "spectre/changes/demo")
		writeFile(t, w+"/src/a.txt", "task\n")
		fx.commitAs(t, "impl", "feat(src): task\n\nTask-Id: 1", "src")
		mkdir(t, w+"/spectre/changes/archive")
		splitGit(t, w, nil, "mv", "spectre/changes/demo", "spectre/changes/archive/demo")
		want := fx.expectedTree(t)
		rc, out := fx.reshape(t, w, w, "demo", fx.base)
		check(t, rc == 0 && strings.Contains(out, "1 planning commit(s)"), "planning commit kept", "rc=%d out=%s", rc, out)
		staged := splitGit(t, w, nil, "diff", "--cached", "--name-status", "-M", "--", "spectre/changes/")
		check(t, staged == "R100\tspectre/changes/demo/tasks.md\tspectre/changes/archive/demo/tasks.md",
			"rename still staged", "staged under spectre/changes/: %q", staged)
		check(t, fx.expectedTree(t) == want, "working tree and index as found", "the tree moved")
	})

	t.Run("3 the guard refuses a main checkout and HEAD does not move", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		before := splitGit(t, fx.main, nil, "rev-parse", "HEAD")
		rc, out := fx.reshape(t, fx.main, fx.main, "demo", fx.base)
		check(t, rc == 1, "main checkout refused with exit 1", "rc=%d out=%s", rc, out)
		check(t, strings.Contains(out, "PLANNING-COMMIT-MAIN-CHECKOUT"), "guard's own line printed", "out=%s", out)
		check(t, splitGit(t, fx.main, nil, "rev-parse", "HEAD") == before, "HEAD untouched", "HEAD moved")
	})

	t.Run("4 a base that does not resolve cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := rbNew(t)
		rc, out := fx.reshape(t, fx.wt, fx.wt, "demo", "no-such-ref")
		check(t, rc == 2, "unresolvable base exits 2", "rc=%d out=%s", rc, out)
		check(t, strings.HasSuffix(out, "\nreshape-branch.sh: base does not resolve to a commit: no-such-ref"), "named reason", "out=%q", out)
	})

	t.Run("usage: not three arguments cannot answer", func(t *testing.T) {
		t.Parallel()
		rc, out := runSplitGuard(t, reshapeBranch, t.TempDir(), "a", "b")
		check(t, rc == 2 && out == "usage: reshape-branch.sh <worktree> <name> <merge-base>", "usage", "rc=%d out=%q", rc, out)
	})
}
