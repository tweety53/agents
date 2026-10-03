package guard

import (
	"os"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-docs-only.sh at 9cd35da8, one
// subtest per ok: label, plus pins of the refusal messages, the collation of
// the printed path and a failing git diff, none of which the harness
// asserted. The guard runs in-process over a copy of one template repository
// per case.
//
// The harness's structural check grepped the bash call sites for a literal
// "$GIT_BIN" -C "$WORKTREE": every git call ran the binary resolved once
// from PATH against the worktree. Here that property is observed at the git
// the guard actually ran — the recording git first on its PATH
// (cpctRecorder) logs every argv it receives. The harness's mutation proof
// sed-broke the `.mdc?$` pattern in a copy of the bash; its two labels here
// pin what that proof established — the anchored suffix alone decides case
// 1's verdict, and case 3's exit 1 does not depend on it.
func TestCheckPanelDocsOnly(t *testing.T) {
	t.Parallel()
	// new_repo, built once and copied per case: main with base.go and
	// tracked.md in one commit, the merge base every case starts from.
	// tracked.md is tracked from the start so an unstaged-only edit is a
	// modification `git diff` detects.
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", tmpl)
	g.write(tmpl+"/base.go", "base")
	g.write(tmpl+"/tracked.md", "# tracked")
	g.git(tmpl, "add", "base.go", "tracked.md")
	g.git(tmpl, "commit", "-qm", "base")
	mb := g.git(tmpl, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}

	type cpdoFx struct {
		dir, repo string
		g         fxGit
		env       Env
	}
	newFx := func(t *testing.T) *cpdoFx {
		fx := &cpdoFx{dir: t.TempDir()}
		fx.repo = fx.dir + "/repo"
		gdcCopyTree(t, tmpl, fx.repo)
		fx.env = Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
		return fx
	}
	// commit is `echo <f> > <f>` for each file, then one add and commit.
	commit := func(fx *cpdoFx, files ...string) {
		for _, f := range files {
			fx.g.write(fx.repo+"/"+f, f)
		}
		fx.g.git(fx.repo, append([]string{"add"}, files...)...)
		fx.g.git(fx.repo, "commit", "-qm", "add "+strings.Join(files, " "))
	}
	stage := func(fx *cpdoFx, f string) {
		fx.g.write(fx.repo+"/"+f, f)
		fx.g.git(fx.repo, "add", f)
	}
	run := func(t *testing.T, fx *cpdoFx, args ...string) guardResult {
		t.Helper()
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		return runGuard("check-panel-docs-only", args, fx.env)
	}
	// is asserts the exit code and the whole merged output, as the
	// harness's OUT="$(guard 2>&1)" read it, with that output on stdout.
	is := func(t *testing.T, r guardResult, rc int, out string) {
		t.Helper()
		if r.rc != rc || r.out != out || r.err != "" {
			t.Fatalf("want rc=%d out=%q: rc=%d stdout=%q stderr=%q", rc, out, r.rc, r.stdout, r.err)
		}
	}
	rcIs := func(t *testing.T, r guardResult, rc int) {
		t.Helper()
		if r.rc != rc {
			t.Fatalf("want rc=%d: rc=%d out=%s", rc, r.rc, r.out)
		}
	}
	case1 := func(t *testing.T) guardResult {
		fx := newFx(t)
		commit(fx, "notes.md")
		return run(t, fx, fx.repo, mb)
	}
	case3 := func(t *testing.T) guardResult {
		fx := newFx(t)
		commit(fx, "notes.md", "main.go")
		return run(t, fx, fx.repo, mb)
	}
	case4 := func(t *testing.T) guardResult {
		fx := newFx(t)
		commit(fx, "notes.md")
		stage(fx, "main.go")
		return run(t, fx, fx.repo, mb)
	}
	case5 := func(t *testing.T) guardResult {
		fx := newFx(t)
		commit(fx, "notes.md")
		fx.g.appendLine(fx.repo+"/base.go", "unstaged edit")
		return run(t, fx, fx.repo, mb)
	}

	for _, tc := range []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"committed .md change only -> exit 0", func(t *testing.T) { rcIs(t, case1(t), 0) }},
		{"committed .md change only -> empty stdout", func(t *testing.T) { is(t, case1(t), 0, "") }},
		{"committed .mdc change only -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			commit(fx, "rule.mdc")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"committed .mdc change only -> empty stdout", func(t *testing.T) {
			fx := newFx(t)
			commit(fx, "rule.mdc")
			is(t, run(t, fx, fx.repo, mb), 0, "")
		}},
		{"committed .md plus committed main.go -> exit 1", func(t *testing.T) { rcIs(t, case3(t), 1) }},
		{"committed .md plus committed main.go -> stdout is main.go", func(t *testing.T) { is(t, case3(t), 1, "main.go") }},
		{"committed .md plus staged main.go -> exit 1", func(t *testing.T) { rcIs(t, case4(t), 1) }},
		{"committed .md plus staged main.go -> stdout is main.go", func(t *testing.T) { is(t, case4(t), 1, "main.go") }},
		{"committed .md plus unstaged edit to base.go -> exit 1", func(t *testing.T) { rcIs(t, case5(t), 1) }},
		{"committed .md plus unstaged edit to base.go -> stdout is base.go", func(t *testing.T) { is(t, case5(t), 1, "base.go") }},
		{"staged-only .md change -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			stage(fx, "staged.md")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"unstaged-only edit to tracked.md -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			fx.g.appendLine(fx.repo+"/tracked.md", "unstaged edit")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"no changes at all -> exit 1", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.repo, mb), 1)
		}},
		{"no changes at all -> empty stdout", func(t *testing.T) {
			fx := newFx(t)
			is(t, run(t, fx, fx.repo, mb), 1, "")
		}},
		{"missing arguments -> exit 2", func(t *testing.T) {
			rcIs(t, run(t, newFx(t), "", ""), 2)
		}},
		{"worktree not a directory -> exit 2", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.dir+"/nodir", "deadbeef"), 2)
		}},
		{"worktree not a git repository -> exit 2", func(t *testing.T) {
			rcIs(t, run(t, newFx(t), t.TempDir(), "deadbeef"), 2)
		}},
		{"merge base not resolving -> exit 2", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.repo, "0000000000000000000000000000000000000000"), 2)
		}},
		{"flag-shaped merge base ('--evil') -> exit 2", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.repo, "--evil"), 2)
		}},
		{"flag-shaped merge base: reported as not resolving", func(t *testing.T) {
			fx := newFx(t)
			if r := run(t, fx, fx.repo, "--evil"); !strings.Contains(r.out, "does not resolve") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{`every real call site is literally "$GIT_BIN" -C "$WORKTREE", never bare git`, func(t *testing.T) {
			fx := newFx(t)
			commit(fx, "notes.md")
			cpctRecorder(t, fx.dir+"/rec")
			fx.env.Getenv = pathEnv(fx.dir + "/rec")
			is(t, run(t, fx, fx.repo, mb), 0, "")
			b, err := os.ReadFile(fx.dir + "/rec/calls")
			if err != nil {
				t.Fatal(err)
			}
			got := lines(b)
			if len(got) != 5 {
				t.Fatalf("want the guard's 5 git calls through the resolved git, got %d: %q", len(got), got)
			}
			for _, c := range got {
				if !strings.HasPrefix(c, "-C "+fx.repo+" ") {
					t.Errorf("call not against the worktree: %q", c)
				}
			}
		}},
		{"mutation: case 1 flips to exit 1 with .mdc?$ broken", func(t *testing.T) {
			fx := newFx(t)
			commit(fx, "notes.md.txt")
			is(t, run(t, fx, fx.repo, mb), 1, "notes.md.txt")
		}},
		{"mutation: case 3 still exits 1 with .mdc?$ broken", func(t *testing.T) {
			fx := newFx(t)
			commit(fx, "notes.mdx", "main.go")
			is(t, run(t, fx, fx.repo, mb), 1, "main.go")
		}},

		// Beyond the harness: the refusal messages, each what the bash
		// printed at 9cd35da8, on stderr with nothing on stdout.
		{"pin: refusal messages word for word", func(t *testing.T) {
			fx := newFx(t)
			mkdir(t, fx.dir+"/plain")
			p := "check-panel-docs-only: "
			for _, c := range []struct {
				args []string
				want string
			}{
				{nil, p + "usage: check-panel-docs-only.sh <worktree> <merge-base>\n"},
				{[]string{fx.repo}, p + "usage: check-panel-docs-only.sh <worktree> <merge-base>\n"},
				{[]string{"/nonexist", "deadbeef"}, p + "/nonexist is not a directory — cannot determine anything\n"},
				{[]string{"plain", "deadbeef"}, p + "plain is not a git worktree — cannot determine anything\n"},
				{[]string{"repo", "--evil"}, p + "merge base '--evil' does not resolve in repo\n"},
			} {
				r := run(t, fx, c.args...)
				if r.rc != 2 || r.stdout != "" || r.err != c.want {
					t.Errorf("%q: rc=%d stdout=%q stderr=%q", c.args, r.rc, r.stdout, r.err)
				}
			}
			commit(fx, "main.go")
			if r := run(t, fx, "repo", mb, "extra"); r.rc != 1 || r.out != "main.go" {
				t.Errorf("extra argument: rc=%d out=%s", r.rc, r.out)
			}
		}},
		// A failing git diff is the cannot-answer exit, named by its
		// collection, never a verdict.
		{"pin: a failing staged diff -> exit 2 with its own message", func(t *testing.T) {
			fx := newFx(t)
			stubGit(t, fx.dir+"/shim", `has --cached "$@"`, "simulated diff failure (STAGED)")
			fx.env.Getenv = pathEnv(fx.dir + "/shim")
			r := run(t, fx, "repo", mb)
			if r.rc != 2 || r.stdout != "" || r.err != "check-panel-docs-only: cannot list this change's staged paths in repo\n" {
				t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
			}
			if _, err := os.Stat(fx.dir + "/shim/.fired"); err != nil {
				t.Fatal("shim never intercepted the call — this case tested nothing")
			}
		}},
		// Locale-sensitive ordering: the first non-doc path is the first
		// after the bash's `sort -u` under the caller's collation — B.go
		// under C, a.go under en_US.UTF-8, as the bash printed at 9cd35da8.
		{"pin: the printed path follows the caller's collation", func(t *testing.T) {
			for _, c := range []struct{ lcAll, want string }{{"C", "B.go"}, {"en_US.UTF-8", "a.go"}} {
				fx := newFx(t)
				commit(fx, "B.go", "a.go", "_c.md")
				fx.env = envWith(t, []string{"LC_ALL=" + c.lcAll, "GIT_CONFIG_GLOBAL=/dev/null"})
				fx.env.Dir = fx.dir
				if r := run(t, fx, "repo", mb); r.rc != 1 || r.out != c.want {
					t.Errorf("LC_ALL=%s: rc=%d out=%q, want %q", c.lcAll, r.rc, r.out, c.want)
				}
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}
