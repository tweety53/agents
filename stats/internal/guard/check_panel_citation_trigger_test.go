package guard

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-citation-trigger.sh at c5379c0a,
// one subtest per ok: label, plus a pin of the refusal messages the harness
// never asserted word for word. The guard runs in-process over a copy of one
// template repository per case.
//
// The harness's source-text and `bash -x` trace checks proved properties of
// the bash call sites: that every git call ran the binary resolved once from
// PATH, never a shadowable bare `git`, and that the rev-parse gate and the
// committed-paths diff carried --end-of-options with this run's own merge
// base. Here the same properties are observed where they now live — at the
// git the guard actually ran: a recording git first on the guard's PATH logs
// every argv it receives, so a call that bypassed the resolved binary, or
// dropped the flag, shows in the log. The two line numbers the trace labels
// carried name bash lines that no longer exist and are dropped from the
// labels. The harness's mutation proof broke the `.mdc?$` pattern in a
// sandboxed copy; its two labels here pin what that proof established — the
// anchored suffix alone decides case 2's verdict.

// cpctRecorder writes <dir>/git: it appends its argv, one line per call, to
// <dir>/calls, then runs the real git.
func cpctRecorder(t *testing.T, dir string) {
	t.Helper()
	writeExec(t, dir+"/git", fmt.Sprintf(`#!/usr/bin/env bash
printf '%%s\n' "$*" >> "${0%%/*}/calls"
exec %q "$@"
`, fixtureGit))
}

// cpctShadow writes <dir>/git: a hostile git that marks <dir>/.fired and
// strips --end-of-options before running the real one.
func cpctShadow(t *testing.T, dir string) {
	t.Helper()
	writeExec(t, dir+"/git", fmt.Sprintf(`#!/usr/bin/env bash
: > "${0%%/*}/.fired"
args=(); for a in "$@"; do [ "$a" = --end-of-options ] || args+=("$a"); done
exec %q "${args[@]}"
`, fixtureGit))
}

func TestCheckPanelCitationTrigger(t *testing.T) {
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

	// cpctFx is one case's sandbox: repo a template copy, env the guard's.
	type cpctFx struct {
		dir, repo string
		g         fxGit
		env       Env
	}
	newFx := func(t *testing.T) *cpctFx {
		fx := &cpctFx{dir: t.TempDir()}
		fx.repo = fx.dir + "/repo"
		gdcCopyTree(t, tmpl, fx.repo)
		fx.env = Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: os.LookupEnv}
		return fx
	}
	// commitFile is `echo <body> > <f>; git add <f>; git commit`.
	commitFile := func(fx *cpctFx, f string) {
		fx.g.write(fx.repo+"/"+f, f)
		fx.g.git(fx.repo, "add", f)
		fx.g.git(fx.repo, "commit", "-qm", "add "+f)
	}
	run := func(t *testing.T, fx *cpctFx, args ...string) guardResult {
		t.Helper()
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
		return runGuard("check-panel-citation-trigger", args, fx.env)
	}
	rcIs := func(t *testing.T, r guardResult, want int) {
		t.Helper()
		if r.rc != want || r.stdout != "" {
			t.Fatalf("want rc=%d and no stdout: rc=%d out=%s", want, r.rc, r.out)
		}
	}
	calls := func(t *testing.T, dir string) []string {
		t.Helper()
		b, err := os.ReadFile(dir + "/calls")
		if err != nil {
			t.Fatal(err)
		}
		return lines(b)
	}
	// recorded is case 12's run: a committed notes.md, the guard run with
	// the recording git first on PATH (and, when shadow is set, a hostile
	// git after it).
	recorded := func(t *testing.T, shadow bool) (*cpctFx, guardResult) {
		fx := newFx(t)
		commitFile(fx, "notes.md")
		cpctRecorder(t, fx.dir+"/rec")
		path := fx.dir + "/rec"
		if shadow {
			cpctShadow(t, fx.dir+"/shadow")
			path += ":" + fx.dir + "/shadow"
		}
		fx.env.Getenv = pathEnv(path)
		return fx, run(t, fx, fx.repo, mb)
	}
	// F7: a git that fails only the call carrying cond's argument.
	failing := func(t *testing.T, cond, msg string, prep func(fx *cpctFx)) (*cpctFx, guardResult) {
		fx := newFx(t)
		prep(fx)
		stubGit(t, fx.dir+"/shim", cond, msg)
		fx.env.Getenv = pathEnv(fx.dir + "/shim")
		return fx, run(t, fx, fx.repo, mb)
	}
	fired := func(t *testing.T, fx *cpctFx) {
		t.Helper()
		if _, err := os.Stat(fx.dir + "/shim/.fired"); err != nil {
			t.Fatal("shim never intercepted the call — this case tested nothing")
		}
	}

	for _, tc := range []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"committed .md change -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			commitFile(fx, "notes.md")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"committed .mdc change -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			commitFile(fx, "rule.mdc")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"only non-Markdown committed changes -> exit 1", func(t *testing.T) {
			fx := newFx(t)
			commitFile(fx, "main.go")
			rcIs(t, run(t, fx, fx.repo, mb), 1)
		}},
		{"staged-only .md change -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			fx.g.write(fx.repo+"/staged.md", "# staged")
			fx.g.git(fx.repo, "add", "staged.md")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"unstaged-only .md change -> exit 0", func(t *testing.T) {
			fx := newFx(t)
			fx.g.appendLine(fx.repo+"/tracked.md", "unstaged edit")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},
		{"no changes at all -> exit 1", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.repo, mb), 1)
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
		// F2/F3: a merge base shaped like a git option is read as a
		// revision, fails to resolve, and is reported as such.
		{"flag-shaped merge base ('--evil') -> exit 2", func(t *testing.T) {
			fx := newFx(t)
			rcIs(t, run(t, fx, fx.repo, "--evil"), 2)
		}},
		{"flag-shaped merge base: reported as not resolving, not a git option error", func(t *testing.T) {
			fx := newFx(t)
			if r := run(t, fx, fx.repo, "--evil"); !strings.Contains(r.out, "does not resolve") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		// F8/F11/F13: every git call runs the binary resolved from the
		// guard's PATH, against this worktree — none bypasses it.
		{`F8/F11/F13: every real call site is literally "$GIT_BIN" -C "$WORKTREE", never bare git in any spelling`, func(t *testing.T) {
			fx, r := recorded(t, false)
			rcIs(t, r, 0)
			got := calls(t, fx.dir+"/rec")
			if len(got) != 5 {
				t.Fatalf("want the guard's 5 git calls through the resolved git, got %d: %q", len(got), got)
			}
			for _, c := range got {
				if !strings.HasPrefix(c, "-C "+fx.repo+" ") {
					t.Errorf("call not against the worktree: %q", c)
				}
			}
		}},
		{"F8/F11/F13 mutation proof: a git() shadow (space-in-parens, defined after GIT_BIN) never runs — the real call still carries the flag", func(t *testing.T) {
			fx, r := recorded(t, true)
			rcIs(t, r, 0)
			if _, err := os.Stat(fx.dir + "/shadow/.fired"); err == nil {
				t.Fatal("the shadow git ran")
			}
			want := "-C " + fx.repo + " rev-parse --verify --end-of-options " + mb + "^{commit}"
			if !strings.Contains(strings.Join(calls(t, fx.dir+"/rec"), "\n"), want) {
				t.Fatalf("no call %q", want)
			}
		}},
		{"F3/F9/F10: the guard's own rev-parse gate traces carrying the flag for this run's merge base", func(t *testing.T) {
			fx, r := recorded(t, false)
			rcIs(t, r, 0)
			want := "-C " + fx.repo + " rev-parse --verify --end-of-options " + mb + "^{commit}"
			if got := calls(t, fx.dir+"/rec"); len(got) < 2 || got[1] != want {
				t.Fatalf("want the gate call %q, got %q", want, got)
			}
		}},
		{"F2/F9/F10: the guard's own COMMITTED collection traces carrying the flag for this run's merge base", func(t *testing.T) {
			fx, r := recorded(t, false)
			rcIs(t, r, 0)
			want := "-C " + fx.repo + " diff --no-renames --name-only --end-of-options " + mb + "..HEAD"
			if got := calls(t, fx.dir+"/rec"); len(got) < 3 || got[2] != want {
				t.Fatalf("want the committed-paths call %q, got %q", want, got)
			}
		}},
		{"F7: a failing COMMITTED diff call -> exit 2", func(t *testing.T) {
			_, r := failing(t, `has `+mb+`..HEAD "$@"`, "simulated diff failure (COMMITTED)", func(fx *cpctFx) { commitFile(fx, "notes.md") })
			rcIs(t, r, 2)
		}},
		{"F7: COMMITTED failure reported with its own message", func(t *testing.T) {
			_, r := failing(t, `has `+mb+`..HEAD "$@"`, "simulated diff failure (COMMITTED)", func(fx *cpctFx) { commitFile(fx, "notes.md") })
			if !strings.Contains(r.out, "committed paths") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{"F7 COMMITTED shim: shim actually intercepted the call", func(t *testing.T) {
			fx, _ := failing(t, `has `+mb+`..HEAD "$@"`, "simulated diff failure (COMMITTED)", func(fx *cpctFx) { commitFile(fx, "notes.md") })
			fired(t, fx)
		}},
		{"F7: a failing STAGED diff call -> exit 2", func(t *testing.T) {
			_, r := failing(t, `has --cached "$@"`, "simulated diff failure (STAGED)", func(fx *cpctFx) {
				fx.g.write(fx.repo+"/notes.md", "notes")
				fx.g.git(fx.repo, "add", "notes.md")
			})
			rcIs(t, r, 2)
		}},
		{"F7: STAGED failure reported with its own message", func(t *testing.T) {
			_, r := failing(t, `has --cached "$@"`, "simulated diff failure (STAGED)", func(fx *cpctFx) {
				fx.g.write(fx.repo+"/notes.md", "notes")
				fx.g.git(fx.repo, "add", "notes.md")
			})
			if !strings.Contains(r.out, "staged paths") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{"F7 STAGED shim: shim actually intercepted the call", func(t *testing.T) {
			fx, _ := failing(t, `has --cached "$@"`, "simulated diff failure (STAGED)", func(fx *cpctFx) {
				fx.g.write(fx.repo+"/notes.md", "notes")
				fx.g.git(fx.repo, "add", "notes.md")
			})
			fired(t, fx)
		}},
		{"mutation: breaking .mdc?$ flips case 2 from pass to fail", func(t *testing.T) {
			fx := newFx(t)
			commitFile(fx, "rule.mdcx")
			rcIs(t, run(t, fx, fx.repo, mb), 1)
		}},
		{"mutation: guard restored, case 2 passes again", func(t *testing.T) {
			fx := newFx(t)
			commitFile(fx, "rule.mdcx")
			commitFile(fx, "rule.mdc")
			rcIs(t, run(t, fx, fx.repo, mb), 0)
		}},

		// The refusal messages, each what the bash printed at c5379c0a.
		{"pin: refusal messages word for word", func(t *testing.T) {
			fx := newFx(t)
			mkdir(t, fx.dir+"/plain")
			p := "check-panel-citation-trigger: "
			for _, c := range []struct {
				args []string
				want string
			}{
				{nil, p + "usage: check-panel-citation-trigger.sh <worktree> <merge-base>\n"},
				{[]string{fx.repo}, p + "usage: check-panel-citation-trigger.sh <worktree> <merge-base>\n"},
				{[]string{"/nonexist", "deadbeef"}, p + "/nonexist is not a directory — cannot determine anything\n"},
				{[]string{"plain", "deadbeef"}, p + "plain is not a git worktree — cannot determine anything\n"},
				{[]string{"repo", "--evil"}, p + "merge base '--evil' does not resolve in repo\n"},
			} {
				r := run(t, fx, c.args...)
				if r.rc != 2 || r.stdout != "" || r.err != c.want {
					t.Errorf("%q: rc=%d stdout=%q stderr=%q", c.args, r.rc, r.stdout, r.err)
				}
			}
			if r := run(t, fx, "repo", mb, "extra"); r.rc != 1 || r.out != "" {
				t.Errorf("extra argument: rc=%d out=%s", r.rc, r.out)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}
