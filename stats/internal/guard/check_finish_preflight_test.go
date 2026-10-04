package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-finish-preflight.sh at d71a2327, one
// subtest per ok: label, over the same main-checkout-plus-linked-worktree
// repositories, built fresh per case: a linked worktree records absolute
// paths, so it cannot be copied from a template. The guard runs in-process
// with FLOW_GUARD_WORKTREE_LOCATION set as the shim exports it, so the
// stray-worktree assertion execs this checkout's real
// scripts/check-worktree-location.sh —
// except case 20, whose root carries a non-executable copy. The shared
// plumbing (fxGit, stubGit, runGuard, the bm* checks) is
// check_base_moved_test.go's.

// fpRepoRoot is this checkout's root, the tree the bash guard sat in.
var fpRepoRoot = func() string {
	root, err := filepath.Abs("../../..")
	if err != nil {
		panic(err)
	}
	return root
}()

// fpFx is one case's sandbox: main is new_repo's MAIN_REPO, repo its linked
// worktree REPO on demo/branch, root the checkout whose scripts/ holds the
// check-worktree-location.sh the guard execs.
type fpFx struct {
	bmFx
	main, root string
}

// newRepo is new_repo: MAIN_REPO on main with file.txt in one commit, and
// REPO a linked worktree of it under MAIN_REPO/.worktrees/ — in-tree, so
// check-worktree-location reports no stray — on demo/branch at that commit.
func (fx *fpFx) newRepo(cloneOf string) {
	fx.main = fx.dir + "/main"
	if cloneOf == "" {
		fx.g.git("", "init", "-q", "-b", "main", fx.main)
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
	fx.repo = fx.main + "/.worktrees/demo"
	fx.g.git(fx.main, "worktree", "add", "-q", "-b", "demo/branch", fx.repo, "main")
}

// withOrigin is new_repo_with_origin: newRepo over a clone of a bare origin.
func (fx *fpFx) withOrigin() {
	origin := fx.dir + "/origin"
	fx.g.git("", "init", "-q", "--bare", "-b", "main", origin)
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

// fpEnv is the process's environment with FLOW_GUARD_WORKTREE_LOCATION set to
// root's scripts/check-worktree-location.sh, as the shim exports it, and PATH
// led by stub when there is one.
func fpEnv(dir, root, stub string) Env {
	lookup := func(k string) (string, bool) {
		switch {
		case k == "FLOW_GUARD_WORKTREE_LOCATION":
			return root + "/scripts/check-worktree-location.sh", true
		case k == "PATH" && stub != "":
			return stub + ":" + os.Getenv("PATH"), true
		}
		return os.LookupEnv(k)
	}
	return Env{Dir: dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
}

func TestCheckFinishPreflight(t *testing.T) {
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
			bmRC("merged clean: exit 0", 0)}},
		{setup: func(fx *fpFx) { merged(fx); fx.g.write(fx.repo+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("merged but dirty -> REFUSE", "REFUSE"),
			bmRC("merged dirty: exit 0 — the verdict carries the answer", 0)}},
		{setup: func(fx *fpFx) { fx.newRepo(""); fx.work(); fx.g.write(fx.repo+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("unmerged and dirty -> RUN1", "RUN1")}},
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
				return r.err == baseRefUsageWant("check-finish-preflight.sh")
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
			git := envGit(Env{Dir: fx.dir, Getenv: os.Getenv})
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
		// into a REFUSE. 19: a stray worktree still does.
		{setup: func(fx *fpFx) { merged(fx); fx.g.git(fx.main, "checkout", "-q", "-b", "other") }, checks: []bmCheck{
			bmPrefix("main checkout off base -> RUN2", "RUN2"),
			bmRC("main checkout off base: exit 0", 0)}},
		{setup: func(fx *fpFx) { merged(fx); fx.g.appendLine(fx.main+"/file.txt", "dirty") }, checks: []bmCheck{
			bmPrefix("main checkout has tracked changes -> RUN2", "RUN2")}},
		{setup: func(fx *fpFx) {
			merged(fx)
			fx.g.git(fx.main, "worktree", "add", "-q", "--detach", fpStray(fx.dir), "main")
		}, checks: []bmCheck{
			bmPrefix("stray worktree -> REFUSE", "REFUSE"),
			bmHas("stray worktree: REFUSE names the reason", "stray worktree"),
			{"stray worktree: REFUSE names the offending path", func(r guardResult, fx *bmFx) bool {
				return strings.Contains(r.out, fpStray(fx.dir))
			}}}},
		// 20: the assertion's own dependency unusable is "cannot determine
		// anything", never an implicit pass (KAN-462 panel fix, F2).
		{setup: merged, prep: func(t *testing.T, fx *fpFx) {
			fx.root = fx.dir + "/root"
			b, err := os.ReadFile(fpRepoRoot + "/scripts/check-worktree-location.sh")
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, fx.root+"/scripts/check-worktree-location.sh", string(b))
		}, checks: []bmCheck{
			bmRC("check-worktree-location.sh not executable -> exit 2", 2),
			noVerdict("check-worktree-location.sh not executable: emits no verdict line", notVerdict...),
			{"check-worktree-location.sh not executable: names the failure", func(r guardResult, _ *bmFx) bool {
				i := strings.Index(r.out, "check-worktree-location.sh")
				return i >= 0 && strings.Contains(r.out[i:], "missing or not executable")
			}}}},
		// Beyond the harness: check-worktree-location's other exits, each as
		// the bash handled it, through a stub in the fixture's root.
		{setup: merged, prep: fpLocation("printf 'STRAY: /x;y (b)\\nnoise\\nSTRAY: /z (detached)\\n'; exit 1"), checks: []bmCheck{
			{"port: stray lines are joined with '; ', a ';' inside a path included", func(r guardResult, _ *bmFx) bool {
				return r.rc == 0 && r.stdout == "REFUSE: stray worktree(s): /x; y (b); /z (detached)\n"
			}}}},
		{setup: merged, prep: fpLocation("echo not-a-stray; exit 1"), checks: []bmCheck{
			{"port: exit 1 with no STRAY line exits 1 with nothing printed, as set -e did", func(r guardResult, _ *bmFx) bool {
				return r.rc == 1 && r.out == ""
			}}}},
		{setup: merged, prep: fpLocation("echo out; printf 'boom\\n\\n' >&2; exit 2"), checks: []bmCheck{
			{"port: exit 2 relays its stderr once, trailing newlines collapsed", func(r guardResult, _ *bmFx) bool {
				return r.rc == 2 && r.stdout == "" && r.err == "boom\n"
			}}}},
		{setup: merged, prep: fpLocation("echo out; echo err >&2; exit 3"), checks: []bmCheck{
			{"port: any other exit is named, its output dropped", func(r guardResult, fx *bmFx) bool {
				return r.rc == 2 && r.stdout == "" && strings.HasPrefix(r.err,
					"check-finish-preflight: check-worktree-location.sh exited 3 against ") &&
					strings.HasSuffix(r.err, " — cannot determine anything\n")
			}}}},
		{setup: merged, prep: fpLocation("kill -TERM $$"), checks: []bmCheck{
			{"port: a location check killed by a signal is named by bash's 128+n", func(r guardResult, _ *bmFx) bool {
				return r.rc == 2 && strings.HasPrefix(r.err, "check-finish-preflight: check-worktree-location.sh exited 143 against ")
			}}}},
		// Run without the shim, the sibling's location is unknown: refused by
		// name, never looked up at a path the guard made up.
		{setup: merged, run: func(fx *fpFx) (guardResult, error) {
			lookup := func(k string) (string, bool) {
				if k == "FLOW_GUARD_WORKTREE_LOCATION" {
					return "", false
				}
				return os.LookupEnv(k)
			}
			env := Env{Dir: fx.dir, LookupEnv: lookup, Getenv: func(k string) string { v, _ := lookup(k); return v }}
			return runGuard("check-finish-preflight", []string{fx.repo, "main", fx.recorded}, env), nil
		}, checks: []bmCheck{
			{"port: an unset FLOW_GUARD_WORKTREE_LOCATION is refused by name", func(r guardResult, _ *bmFx) bool {
				return r.rc == 2 && r.stdout == "" &&
					r.err == "check-finish-preflight: FLOW_GUARD_WORKTREE_LOCATION is unset — run scripts/check-finish-preflight.sh, which sets it\n"
			}}}},
		// A relative worktree argument resolves against the working directory.
		{setup: merged, args: func(fx *fpFx) []string { return []string{"main/.worktrees/demo", "main", fx.recorded} },
			checks: []bmCheck{
				{"port: a relative worktree resolves against the working directory", func(r guardResult, _ *bmFx) bool {
					return r.rc == 0 && strings.HasPrefix(r.stdout, "RUN2: ")
				}}}},
	}

	for i, c := range cases {
		// The sandbox belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		fx := &fpFx{bmFx: bmFx{dir: t.TempDir()}, root: fpRepoRoot}
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
			return runGuard("check-finish-preflight", args, fpEnv(fx.dir, fx.root, fx.stub)), nil
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

	// Beyond the harness (panel review of task 12): the dirty-entry count and
	// the main checkout's physical path (`pwd -P`) reach the verdict line,
	// so each is run against the bash at d71a2327 live on the same fixture.
	bashDir := filepath.Join(t.TempDir(), "scripts")
	for _, rel := range []string{"check-finish-preflight.sh", "check-worktree-location.sh", "lib/resolve-remote-base.sh"} {
		src, err := exec.Command(fixtureGit, "-C", fpRepoRoot, "show", "d71a2327:scripts/"+rel).Output()
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(bashDir, rel), string(src))
		if err := os.Chmod(filepath.Join(bashDir, rel), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	parity := []struct {
		label string
		setup func(fx *fpFx) (dir string, args []string)
	}{
		{"port: the dirty-entry count matches the bash", func(fx *fpFx) (string, []string) {
			merged(fx)
			fx.g.write(fx.repo+"/one.txt", "dirty")
			fx.g.write(fx.repo+"/two.txt", "dirty")
			return fx.dir, []string{fx.repo, "main", fx.recorded}
		}},
		{"port: the main checkout is resolved to its physical path, as pwd -P gave it", func(fx *fpFx) (string, []string) {
			merged(fx)
			fx.g.git(fx.main, "worktree", "add", "-q", "--detach", fpStray(fx.dir), "main")
			// Only a relative --git-common-dir (the main checkout's own `.git`)
			// leaves the physical path to pwd -P; git prints a linked
			// worktree's already resolved.
			fx.g.write(fx.main+"/.git/info/exclude", ".worktrees/\n")
			link := fx.dir + "/link"
			if err := os.Symlink(fx.dir, link); err != nil {
				fx.g.err = err
			}
			return link, []string{"main", "main", fx.recorded}
		}},
	}
	for _, x := range parity {
		t.Run(x.label, func(t *testing.T) {
			t.Parallel()
			fx := &fpFx{bmFx: bmFx{dir: t.TempDir()}, root: fpRepoRoot}
			dir, args := x.setup(fx)
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			r := runGuard("check-finish-preflight", args, fpEnv(dir, fx.root, ""))
			cmd := exec.Command("bash", append([]string{filepath.Join(bashDir, "check-finish-preflight.sh")}, args...)...)
			cmd.Dir = dir
			var bout, berr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &bout, &berr
			brc := 0
			if err := cmd.Run(); err != nil {
				ee, ok := err.(*exec.ExitError)
				if !ok {
					t.Fatal(err)
				}
				brc = ee.ExitCode()
			}
			if r.rc != brc || r.stdout != bout.String() || r.err != berr.String() {
				t.Fatalf("port rc=%d stdout=%q stderr=%q\nbash rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err, brc, bout.String(), berr.String())
			}
		})
	}
}

// fpLocation is a prep writing <root>/scripts/check-worktree-location.sh as
// a stub running body.
func fpLocation(body string) func(t *testing.T, fx *fpFx) {
	return func(t *testing.T, fx *fpFx) {
		fx.root = fx.dir + "/root"
		writeExec(t, fx.root+"/scripts/check-worktree-location.sh", "#!/usr/bin/env bash\n"+body+"\n")
	}
}

// fpStray is case 19's worktree outside <main>/.worktrees/, physical as the
// harness's `pwd -P` made it.
func fpStray(dir string) string {
	real, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return dir + "/unresolvable"
	}
	return real + "/stray"
}
