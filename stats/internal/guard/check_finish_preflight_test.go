package guard

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// Every case of scripts/test-check-finish-preflight.sh at d71a2327, one
// subtest per ok: label, over the same main-checkout-plus-linked-worktree
// repositories, built fresh per case: a linked worktree records absolute
// paths, so it cannot be copied from a template. The guard runs in-process.
// The shared plumbing (fxGit, stubGit, runGuard, the bm* checks) is
// check_base_moved_test.go's.

// fpFx is one case's sandbox: main is new_repo's MAIN_REPO, repo its linked
// worktree REPO on demo/branch.
type fpFx struct {
	bmFx
	main string
}

// newRepo is new_repo: MAIN_REPO on main with file.txt in one commit, and
// REPO a linked worktree of it under MAIN_REPO-worktrees/ on demo/branch at
// that commit.
func (fx *fpFx) newRepo(cloneOf string) {
	fx.main = fx.dir + "/main"
	if cloneOf == "" {
		fx.g.copyShape(fx.main, repoShape{Branch: "main"})
	} else {
		fx.g.git("", "clone", "-q", cloneOf, fx.main)
	}
	fx.g.write(fx.main+"/file.txt", "base")
	fx.g.git(fx.main, "add", "file.txt")
	fx.g.git(fx.main, "commit", "-qm", "base")
	if cloneOf != "" {
		fx.g.git(fx.main, "push", "-q", "origin", "main")
	}
	fx.recorded = fx.g.git(fx.main, "rev-parse", "HEAD")
	fx.repo = fx.main + "-worktrees/demo"
	fx.g.git(fx.main, "worktree", "add", "-q", "-b", "demo/branch", fx.repo, "main")
}

// withOrigin is new_repo_with_origin: newRepo over a clone of a bare origin.
func (fx *fpFx) withOrigin() {
	origin := fx.dir + "/origin"
	fx.g.copyShape(origin, repoShape{Branch: "main", Bare: true})
	fx.newRepo(origin)
}

// work commits new.txt on demo/branch.
func (fx *fpFx) work() {
	fx.g.write(fx.repo+"/new.txt", "work")
	fx.g.git(fx.repo, "add", "new.txt")
	fx.g.git(fx.repo, "commit", "-qm", "work")
}

// merge is merge_demo_into_main, run from MAIN_REPO, the one worktree
// allowed to hold main.
func (fx *fpFx) merge() { fx.g.git(fx.main, "merge", "-q", "--no-ff", "-m", "merge", "demo/branch") }

// mergeOrigin is merge_demo_into_origin_main: demo/branch lands on
// origin/main while local main stays at the recorded merge base.
func (fx *fpFx) mergeOrigin() {
	fx.g.git(fx.repo, "branch", "-q", "tmp-merge", "main")
	fx.g.git(fx.repo, "checkout", "-q", "tmp-merge")
	fx.g.git(fx.repo, "merge", "-q", "--no-ff", "-m", "merge", "demo/branch")
	fx.g.git(fx.repo, "push", "-q", "origin", "tmp-merge:main")
	fx.g.git(fx.repo, "checkout", "-q", "demo/branch")
	fx.g.git(fx.repo, "branch", "-q", "-D", "tmp-merge")
	fx.g.git(fx.repo, "fetch", "-q", "origin")
}

// fpEnv is the process's environment with PATH led by stub when there is
// one.
func fpEnv(dir, stub string) core.Env {
	lookup := func(k string) (string, bool) {
		switch {
		case k == "PATH" && stub != "":
			return stub + ":" + os.Getenv("PATH"), true
		}
		return os.LookupEnv(k)
	}
	return core.Env{Dir: dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
}

func TestCheckFinishPreflight(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	notVerdict := []string{"RUN1", "RUN2", "REFUSE"}
	merged := func(fx *fpFx) { fx.newRepo(""); fx.work(); fx.merge() }

	type fpCase struct {
		setup  func(fx *fpFx)                      // nil: nothing
		prep   func(t *testing.T, fx *fpFx)        // runs in the parent, before any subtest
		stub   string                              // stubGit condition; "" runs with no shim
		msg    string                              // the stub's fatal message
		args   func(fx *fpFx) []string             // nil: {repo, main, recorded}
		run    func(fx *fpFx) (guardResult, error) // replaces the guard run entirely
		checks []bmCheck
	}
	cases := []fpCase{
		// 1: HEAD is still the merge base, so the ancestor test alone says "merged".
		{setup: func(fx *fpFx) {
			fx.newRepo("")
			fx.g.write(fx.repo+"/new.txt", "staged")
			fx.g.git(fx.repo, "add", "new.txt")
		}, checks: []bmCheck{
			bmPrefix("zero-commit branch with staged work -> RUN1", "RUN1"),
			bmRC("zero-commit branch: exit 0", 0)}},
		// 1b: signal (b) needs only HEAD and the recorded merge base.
		{setup: func(fx *fpFx) {
			fx.newRepo("")
			fx.g.write(fx.repo+"/new.txt", "staged")
			fx.g.git(fx.repo, "add", "new.txt")
		}, args: func(fx *fpFx) []string { return []string{fx.repo, "no-such-base", fx.recorded} }, checks: []bmCheck{
			bmPrefix("zero-commit branch, unresolvable base ref -> RUN1", "RUN1")}},
		// 1c: every other signal points the wrong way; (b) alone stands
		// between this worktree and run 2.
		{setup: func(fx *fpFx) { fx.newRepo("") }, checks: []bmCheck{
			bmPrefix("zero-commit branch with a clean tree -> RUN1", "RUN1"),
			bmRC("zero-commit clean branch: exit 0", 0)}},
		{setup: merged, checks: []bmCheck{
			bmPrefix("merged with clean tree -> RUN2", "RUN2"),
			bmLacks("RUN2 carries no base-moved line", "CLEAR:"),
			bmRC("merged clean: exit 0", 0)}},
		{setup: func(fx *fpFx) { merged(fx); fx.g.write(fx.repo+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("merged but dirty -> REFUSE", "REFUSE"),
			bmRC("merged dirty: exit 0 — the verdict carries the answer", 0)}},
		{setup: func(fx *fpFx) { fx.newRepo(""); fx.work(); fx.g.write(fx.repo+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("unmerged and dirty -> RUN1", "RUN1"),
			bmHas("RUN1 is followed by check-base-moved's verdict line", "\nCLEAR: ")}},
		// 5: the reason, not just the prefix — `-^{commit}` would REFUSE too.
		{setup: merged, args: func(fx *fpFx) []string { return []string{fx.repo, "main", "-"} }, checks: []bmCheck{
			bmPrefix("absent merge base -> REFUSE", "REFUSE"),
			bmHas("absent merge base: the refusal says the record is absent", "no merge base recorded"),
			bmLacks("absent merge base: not reported as a ref-resolution failure", "does not resolve")}},
		{setup: merged, args: func(fx *fpFx) []string {
			return []string{fx.repo, "main", fx.g.git(fx.repo, "rev-parse", "--short", fx.recorded)}
		}, checks: []bmCheck{bmPrefix("abbreviated merge base -> RUN2", "RUN2")}},
		{setup: func(fx *fpFx) { fx.newRepo(""); fx.work() },
			args: func(fx *fpFx) []string { return []string{fx.repo, "main", strings.Repeat("0", 40)} }, checks: []bmCheck{
				bmPrefix("unresolvable merge base -> REFUSE", "REFUSE"),
				bmRC("unresolvable merge base: exit 0", 0)}},
		{setup: func(fx *fpFx) { fx.newRepo(""); fx.work() },
			args: func(fx *fpFx) []string { return []string{fx.repo, "no-such-base", fx.recorded} }, checks: []bmCheck{
				bmPrefix("unresolvable base ref -> REFUSE", "REFUSE")}},
		{args: func(fx *fpFx) []string { return []string{fx.dir, "main", "deadbeef"} }, checks: []bmCheck{
			bmRC("non-repository -> exit 2", 2),
			noVerdict("non-repository: emits no verdict line", notVerdict...)}},
		// 8b: signal (d) is the last thing between a merged branch and run 2.
		{setup: merged, stub: `has status "$@"`, msg: "simulated index.lock contention", checks: []bmCheck{
			bmRC("unreadable worktree status -> exit 2", 2),
			noVerdict("unreadable status: emits no verdict line", notVerdict...),
			bmHas("unreadable status: names the failure", "check-finish-preflight:"),
			shimFired("unreadable status")}},
		// 8c: only exit 1 from the ancestor test means "not an ancestor".
		{setup: func(fx *fpFx) { fx.newRepo(""); fx.work() },
			stub: `has merge-base "$@"`, msg: "simulated repository corruption", checks: []bmCheck{
				bmRC("merge-base failing (not 'not an ancestor') -> exit 2", 2),
				noVerdict("merge-base failure: emits no verdict line", notVerdict...),
				bmHas("merge-base failure: names the failure", "check-finish-preflight:"),
				shimFired("merge-base failure")}},
		// 8d: HEAD_SHA's own capture guard.
		{setup: func(fx *fpFx) { fx.newRepo("") },
			stub: `has 'HEAD^{commit}' "$@"`, msg: "simulated unresolvable HEAD", checks: []bmCheck{
				bmRC("unresolvable HEAD -> exit 2", 2),
				noVerdict("unresolvable HEAD: emits no verdict line", notVerdict...),
				bmHas("unresolvable HEAD: names the failure", "cannot resolve HEAD"),
				shimFired("unresolvable HEAD")}},
		{args: func(*fpFx) []string { return []string{"", "", ""} }, checks: []bmCheck{
			bmRC("missing arguments -> exit 2", 2)}},
		{args: func(*fpFx) []string { return nil }, checks: []bmCheck{
			{"missing arguments: stdout is empty", func(r guardResult, _ *bmFx) bool { return r.stdout == "" }},
			{"missing arguments: stderr is exactly the usage message", func(r guardResult, _ *bmFx) bool {
				return r.err == baseRefUsageWant("flow-guard check-finish-preflight")
			}}}},
		// 10: merged into origin/main, local main left behind; handed the
		// bare name — the KAN-88 regression.
		{setup: func(fx *fpFx) { fx.withOrigin(); fx.work(); fx.mergeOrigin() }, checks: []bmCheck{
			bmPrefix("KAN-88: bare 'main' behind origin, merged into origin/main -> RUN2", "RUN2"),
			bmHas("KAN-88: RUN2 verdict names origin/main", "origin/main")}},
		// 11: the same shape handed origin/main explicitly is unchanged.
		{setup: func(fx *fpFx) { fx.withOrigin(); fx.work(); fx.mergeOrigin() },
			args: func(fx *fpFx) []string { return []string{fx.repo, "origin/main", fx.recorded} }, checks: []bmCheck{
				bmPrefix("KAN-88: explicit origin/main -> RUN2, unchanged", "RUN2"),
				bmHas("KAN-88: explicit origin/main verdict names origin/main", "origin/main")}},
		{setup: func(fx *fpFx) { fx.withOrigin(); fx.work() }, checks: []bmCheck{
			bmPrefix("KAN-88: bare 'main' resolved to origin/main, unmerged -> RUN1", "RUN1"),
			bmHas("KAN-88: RUN1 verdict names origin/main", "origin/main")}},
		// 13 (F9): resolveRemoteBase over a dash-prefixed base ref, both ways.
		{setup: func(fx *fpFx) {
			fx.withOrigin()
			fx.g.git(fx.repo, "update-ref", "refs/remotes/origin/-weird", fx.g.git(fx.repo, "rev-parse", "HEAD"))
		}, run: func(fx *fpFx) (guardResult, error) {
			git := core.EnvGit(core.Env{Dir: fx.dir, Getenv: os.Getenv})
			return guardResult{out: resolveRemoteBase(git, fx.repo, "-weird") + "|" +
				resolveRemoteBase(git, fx.repo, "-no-such-ref")}, nil
		}, checks: []bmCheck{
			bmPrefix("F9: dash-prefixed base ref with an origin/ counterpart resolves to origin/-weird", "origin/-weird|"),
			{"F9: dash-prefixed base ref with no origin/ counterpart passes through unchanged", func(r guardResult, _ *bmFx) bool {
				return strings.HasSuffix(r.out, "|-no-such-ref")
			}}}},
		// 14 (F9): no input can make --end-of-options observable (the fixed
		// refs/remotes/origin/ prefix), so the source carries the pin. The
		// label is the harness's; resolve-remote-base.sh's function now lives
		// in resolveremotebase.go.
		{run: func(*fpFx) (guardResult, error) {
			b, err := os.ReadFile("resolveremotebase.go")
			return guardResult{out: string(b)}, err
		}, checks: []bmCheck{
			bmHas("F9: resolve-remote-base.sh's rev-parse call still carries --end-of-options", `"rev-parse", "--verify", "--end-of-options",`)}},
		// 15 (F14): the stub's match is exact equality against one argument.
		{prep: func(t *testing.T, fx *fpFx) {
			fx.stub = fx.dir + "/shim"
			stubGit(t, fx.stub, `has abc "$@"`, "should never fire")
		}, run: func(fx *fpFx) (guardResult, error) {
			out, err := exec.Command(fx.stub+"/git", "xabcx").CombinedOutput()
			rc := 0
			if ee, ok := err.(*exec.ExitError); ok {
				rc = ee.ExitCode()
			} else if err != nil {
				return guardResult{}, err
			}
			return guardResult{rc: rc, out: string(out)}, nil
		}, checks: []bmCheck{
			bmLacks("F14: an argument that merely contains the match arg is not intercepted", "should never fire"),
			{"F14: superstring argument did not exit 128 (the shim's own fatal exit)", func(r guardResult, _ *bmFx) bool { return r.rc != 128 }},
			{"F14: superstring argument left no .fired sentinel", func(_ guardResult, fx *bmFx) bool {
				_, err := os.Stat(fx.stub + "/.fired")
				return os.IsNotExist(err)
			}}}},
		// 16 (KAN-298): the harness grepped the script; the text now lives in
		// Go, so the printed usage is read instead.
		{args: func(*fpFx) []string { return nil }, checks: []bmCheck{
			{"usage message states the base-ref rule", func(r guardResult, _ *bmFx) bool {
				return strings.Contains(r.err, "prefers refs/remotes/origin/<base-ref>")
			}}}},
		// 17-18: the main checkout's branch and tracked state are not
		// asserted — run 2 only fast-forwards it — so neither turns a RUN2
		// into a REFUSE.
		{setup: func(fx *fpFx) { merged(fx); fx.g.git(fx.main, "checkout", "-q", "-b", "other") }, checks: []bmCheck{
			bmPrefix("main checkout off base -> RUN2", "RUN2"),
			bmRC("main checkout off base: exit 0", 0)}},
		{setup: func(fx *fpFx) { merged(fx); fx.g.appendLine(fx.main+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("main checkout has tracked changes -> RUN2", "RUN2")}},
		// A relative worktree argument resolves against the working directory.
		{setup: merged, args: func(fx *fpFx) []string { return []string{"main-worktrees/demo", "main", fx.recorded} },
			checks: []bmCheck{
				{"port: a relative worktree resolves against the working directory", func(r guardResult, _ *bmFx) bool {
					return r.rc == 0 && strings.HasPrefix(r.stdout, "RUN2: ")
				}}}},
	}

	for i, c := range cases {
		// The sandbox belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		fx := &fpFx{bmFx: bmFx{dir: t.TempDir()}}
		if c.stub != "" {
			fx.stub = fx.dir + "/shim"
			stubGit(t, fx.stub, c.stub, c.msg)
		}
		if c.prep != nil {
			c.prep(t, fx)
		}
		run := func() (guardResult, error) {
			if c.setup != nil {
				c.setup(fx)
			}
			var args []string
			if c.args != nil {
				args = c.args(fx)
			} else {
				args = []string{fx.repo, "main", fx.recorded}
			}
			if fx.g.err != nil {
				return guardResult{}, fx.g.err
			}
			if c.run != nil {
				return c.run(fx)
			}
			return runGuard("check-finish-preflight", args, fpEnv(fx.dir, fx.stub)), nil
		}
		// One parallel subtest runs the case, its checks nested beneath it:
		// a parallel subtest per check held a -parallel slot apiece while
		// the one running the case worked and the rest waited on it.
		t.Run(fmt.Sprint("run ", i), func(t *testing.T) {
			t.Parallel()
			r, err := run()
			for _, chk := range c.checks {
				t.Run(chk.label, func(t *testing.T) {
					if err != nil {
						t.Fatal(err)
					}
					if !chk.ok(r, &fx.bmFx) {
						t.Fatalf("rc=%d out=%s", r.rc, r.out)
					}
				})
			}
		})
	}
}
