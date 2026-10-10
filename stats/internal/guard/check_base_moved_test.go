package guard

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Every case of scripts/test-check-base-moved.sh at d71a2327, one subtest per
// ok: label. The bash harness built throwaway git repositories and ran the
// guard with stdout and stderr merged; here the guard runs in-process over
// the same repositories, each case building its own in a directory the parent
// test owns, so every subtest reading the case's one run sees it.
//
// The harness's git shims (scripts/lib/test-git-shim.sh and the hand-written
// ones beside it) become stubGit: the same bash stub on PATH, which the guard
// resolves over Env's PATH as the bash guard resolved `git` over its own.

// fxGit builds fixtures under fixtureGitEnv with gitRun's fixed identity. The
// first failure is kept and every later call skipped, so a fixture built
// inside a case's run reports to each of its checks.
type fxGit struct{ err error }

func (g *fxGit) git(dir string, args ...string) string {
	if g.err != nil {
		return ""
	}
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command(fixtureGit, args...)
	cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	var errb bytes.Buffer
	cmd.Stderr = &errb
	out, err := cmd.Output()
	if err != nil {
		g.err = fmt.Errorf("git %v: %v\n%s", args, err, errb.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// appendLine is `echo <line> >> <path>`.
func (g *fxGit) appendLine(path, line string) {
	if g.err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err == nil {
		_, err = io.WriteString(f, line+"\n")
		if cerr := f.Close(); err == nil {
			err = cerr
		}
	}
	if err != nil {
		g.err = err
	}
}

// write is `echo <line> > <path>`.
func (g *fxGit) write(path, line string) {
	if g.err == nil {
		g.err = os.WriteFile(path, []byte(line+"\n"), 0o644)
	}
}

// stubGit writes <dir>/git: when cond (a bash test over the stub's own "$@";
// has <arg> "$@" is shim_failing_git's exact-element match) holds, it marks
// <dir>/.fired and fails as the harness's shims did; otherwise it execs the
// git on the test process's PATH, resolved here once, so a stub directory
// already on PATH can never be the "real" git.
func stubGit(t *testing.T, dir, cond, message string) {
	t.Helper()
	writeExec(t, dir+"/git", fmt.Sprintf(`#!/usr/bin/env bash
has() { local m="$1" a; shift; for a in "$@"; do [ "$a" = "$m" ] && return 0; done; return 1; }
if %s; then
  : > "${0%%/*}/.fired"
  echo "fatal: %s" >&2
  exit 128
fi
exec %q "$@"
`, cond, message, fixtureGit))
}

// guardResult is one in-process run: stdout, stderr, and the two merged in
// write order with trailing newlines stripped, as the harness's
// OUT="$(guard 2>&1)" read them.
type guardResult struct {
	rc               int
	out, stdout, err string
}

func runGuard(name string, args []string, env Env) guardResult {
	fn := Registry[name]
	if fn == nil {
		return guardResult{rc: -1, out: name + " is not registered"}
	}
	var all, o, e bytes.Buffer
	rc := fn(args, env, io.MultiWriter(&o, &all), io.MultiWriter(&e, &all))
	return guardResult{rc, strings.TrimRight(all.String(), "\n"), o.String(), e.String()}
}

// baseRefUsageWant is base_ref_usage_message from scripts/lib/base-ref-usage.sh at d71a2327
// plus the newline the heredoc ends in.
func baseRefUsageWant(script string) string {
	return "usage: " + script + " <worktree> <base-ref> <recorded-merge-base|->\n" +
		"  <base-ref>  the base branch name, bare (main) or remote-tracking\n" +
		"              (origin/main). The guard prefers refs/remotes/origin/<base-ref>\n" +
		"              when it resolves, so a bare name is never tested against a\n" +
		"              stale local branch.\n"
}

type bmCheck struct {
	label string
	ok    func(r guardResult, fx *bmFx) bool
}

func bmRC(label string, want int) bmCheck {
	return bmCheck{label, func(r guardResult, _ *bmFx) bool { return r.rc == want }}
}

func bmPrefix(label, p string) bmCheck {
	return bmCheck{label, func(r guardResult, _ *bmFx) bool { return strings.HasPrefix(r.out, p) }}
}

func bmHas(label, s string) bmCheck {
	return bmCheck{label, func(r guardResult, _ *bmFx) bool { return strings.Contains(r.out, s) }}
}

func bmLacks(label, s string) bmCheck {
	return bmCheck{label, func(r guardResult, _ *bmFx) bool { return !strings.Contains(r.out, s) }}
}

// noVerdict is the harness's `case "$OUT" in <verdict>*) fail`.
func noVerdict(label string, prefixes ...string) bmCheck {
	return bmCheck{label, func(r guardResult, _ *bmFx) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(r.out, p) {
				return false
			}
		}
		return true
	}}
}

// shimFired is assert_shim_fired.
func shimFired(label string) bmCheck {
	return bmCheck{label + ": shim actually intercepted the call", func(_ guardResult, fx *bmFx) bool {
		_, err := os.Stat(fx.stub + "/.fired")
		return err == nil
	}}
}

// bmFx is one case's sandbox: repo is new_repo's REPO (a copy of the
// template), recorded its RECORDED_BASE, stub the case's shim directory.
type bmFx struct {
	dir, repo, recorded, stub string
	g                         fxGit
}

// advanceBase is advance_base: one commit per file on main, then back to demo.
func (fx *bmFx) advanceBase(files ...string) {
	fx.g.git(fx.repo, "checkout", "-q", "main")
	for _, f := range files {
		fx.g.appendLine(fx.repo+"/"+f, "moved")
		fx.g.git(fx.repo, "add", f)
		fx.g.git(fx.repo, "commit", "-qm", "advance: "+f)
	}
	fx.g.git(fx.repo, "checkout", "-q", "demo")
}

// withOrigin is new_repo_with_origin, built fresh: a clone carries its
// origin's absolute path, so it cannot be copied from a template.
func (fx *bmFx) withOrigin() {
	origin := fx.dir + "/origin"
	fx.repo = fx.dir + "/clone"
	fx.g.git("", "init", "-q", "--bare", "-b", "main", origin)
	fx.g.git("", "clone", "-q", origin, fx.repo)
	fx.g.write(fx.repo+"/base.txt", "base")
	fx.g.write(fx.repo+"/shared.txt", "shared-base")
	fx.g.git(fx.repo, "add", "base.txt", "shared.txt")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.g.git(fx.repo, "push", "-q", "origin", "main")
	fx.recorded = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "checkout", "-q", "-b", "demo")
}

// wide is case 6's files: n tracked files, each touched by the base and by
// this change.
func (fx *bmFx) wide(n int) {
	var files []string
	for i := 1; i <= n; i++ {
		files = append(files, fmt.Sprintf("wide%02d.txt", i))
	}
	fx.overlap(files...)
}

// overlap commits files on both sides: added, then touched by the base and
// by this change.
func (fx *bmFx) overlap(files ...string) {
	for _, f := range files {
		fx.g.write(fx.repo+"/"+f, "wide-base")
	}
	fx.g.git(fx.repo, append([]string{"add"}, files...)...)
	fx.g.git(fx.repo, "commit", "-qm", "add wide files")
	fx.advanceBase(files...)
	for _, f := range files {
		fx.g.appendLine(fx.repo+"/"+f, "demo-wide")
		fx.g.git(fx.repo, "add", f)
	}
	fx.g.git(fx.repo, "commit", "-qm", "demo touches all wide files")
}

func TestCheckBaseMoved(t *testing.T) {
	t.Parallel()
	// new_repo, built once and copied per case: main with base.txt and
	// shared.txt in one commit, and demo checked out at it.
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", tmpl)
	g.write(tmpl+"/base.txt", "base")
	g.write(tmpl+"/shared.txt", "shared-base")
	g.git(tmpl, "add", "base.txt", "shared.txt")
	g.git(tmpl, "commit", "-qm", "base")
	recorded := g.git(tmpl, "rev-parse", "HEAD")
	g.git(tmpl, "checkout", "-q", "-b", "demo")
	if g.err != nil {
		t.Fatal(g.err)
	}
	notVerdict := []string{"CLEAR", "MOVED", "REFUSE"}

	type bmCase struct {
		fresh  bool                                // no template copy (origin cases, argument cases)
		setup  func(fx *bmFx)                      // nil: nothing beyond new_repo
		stub   string                              // stubGit condition; "" runs with no shim
		msg    string                              // the stub's fatal message
		args   func(fx *bmFx) []string             // nil: {repo, main, recorded}
		prep   func(t *testing.T, fx *bmFx)        // runs in the parent, before any subtest
		run    func(fx *bmFx) (guardResult, error) // replaces the guard run entirely
		checks []bmCheck
	}
	cases := []bmCase{
		{checks: []bmCheck{
			bmPrefix("unmoved base -> CLEAR", "CLEAR"),
			bmRC("unmoved base: exit 0", 0),
			bmHas("unmoved base: names the reason", "has not moved since the recorded merge base")}},
		{setup: func(fx *bmFx) { fx.advanceBase("unrelated1.txt", "unrelated2.txt") }, checks: []bmCheck{
			bmPrefix("moved with no overlap -> MOVED", "MOVED"),
			bmHas("moved with no overlap: names the actual commit count", "2 commits"),
			bmHas("moved with no overlap: says no overlap", "no overlap"),
			bmLacks("moved with no overlap: names no overlapping path", "overlaps:")}},
		{setup: func(fx *bmFx) {
			fx.advanceBase("shared.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-committed")
			fx.g.git(fx.repo, "add", "shared.txt")
			fx.g.git(fx.repo, "commit", "-qm", "demo touches shared, committed")
		}, checks: []bmCheck{
			bmPrefix("moved with committed overlap -> MOVED", "MOVED"),
			bmHas("moved with committed overlap: names shared.txt", "overlaps: shared.txt")}},
		{setup: func(fx *bmFx) {
			fx.advanceBase("shared.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-staged")
			fx.g.git(fx.repo, "add", "shared.txt")
		}, checks: []bmCheck{bmHas("moved with staged-only overlap: names shared.txt", "overlaps: shared.txt")}},
		{setup: func(fx *bmFx) {
			fx.advanceBase("shared.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-unstaged")
		}, checks: []bmCheck{bmHas("moved with unstaged-only overlap: names shared.txt", "overlaps: shared.txt")}},
		{setup: func(fx *bmFx) { fx.wide(11) }, checks: []bmCheck{
			bmHas("overlap over 10: carries a (+1 more) tail", "(+1 more)"),
			bmHas("overlap over 10: includes the first sorted path", "wide01.txt"),
			bmLacks("overlap over 10: eleventh path is not named individually", "wide11.txt")}},
		{setup: func(fx *bmFx) { fx.wide(10) }, checks: []bmCheck{
			bmLacks("overlap of exactly 10: no (+N more) tail", "more)"),
			{"overlap of exactly 10: all ten paths named", func(r guardResult, _ *bmFx) bool {
				for i := 1; i <= 10; i++ {
					if !strings.Contains(r.out, fmt.Sprintf("wide%02d.txt", i)) {
						return false
					}
				}
				return true
			}}}},
		{args: func(fx *bmFx) []string { return []string{fx.repo, "main", "-"} }, checks: []bmCheck{
			bmPrefix("'-' recorded merge base -> REFUSE", "REFUSE"),
			bmRC("'-' recorded merge base: exit 0", 0),
			bmHas("'-' recorded merge base: names the reason", "no merge base recorded")}},
		{args: func(fx *bmFx) []string { return []string{fx.repo, "main", strings.Repeat("0", 40)} }, checks: []bmCheck{
			bmPrefix("unresolvable recorded merge base -> REFUSE", "REFUSE"),
			bmRC("unresolvable recorded merge base: exit 0", 0)}},
		{args: func(fx *bmFx) []string { return []string{fx.repo, "no-such-base", fx.recorded} }, checks: []bmCheck{
			bmPrefix("unresolvable base ref -> REFUSE", "REFUSE"),
			bmRC("unresolvable base ref: exit 0", 0)}},
		// 9b: fails only `rev-list`, so this exercises the COUNT capture alone.
		{setup: func(fx *bmFx) { fx.advanceBase("unrelated1.txt") },
			stub: `has rev-list "$@"`, msg: "simulated pack corruption", checks: []bmCheck{
				bmRC("unreadable commit count -> exit 2", 2),
				noVerdict("unreadable commit count: emits no verdict line", notVerdict...),
				bmHas("unreadable commit count: names the failure", "check-base-moved:"),
				shimFired("unreadable commit count")}},
		// 9c: MOVED_RAW's diff carries the same range as COUNT's rev-list, so
		// the stub matches subcommand ($3) and range ($7) together.
		{setup: func(fx *bmFx) { fx.advanceBase("unrelated1.txt") },
			stub: `[ "$3" = "diff" ] && [ "$7" = "` + recorded + `..main" ]`, msg: "simulated MOVED_RAW read failure",
			checks: []bmCheck{
				bmRC("unreadable moved-paths list -> exit 2", 2),
				noVerdict("unreadable moved-paths list: emits no verdict line", notVerdict...),
				bmHas("unreadable moved-paths list: names the failure", "changed on"),
				shimFired("unreadable moved-paths list")}},
		// 9d: RECORDED_BASE..HEAD is unique to the COMMITTED_RAW capture.
		{setup: func(fx *bmFx) {
			fx.advanceBase("unrelated1.txt")
			fx.g.write(fx.repo+"/demo.txt", "work")
			fx.g.git(fx.repo, "add", "demo.txt")
			fx.g.git(fx.repo, "commit", "-qm", "demo work")
		}, stub: `has "` + recorded + `..HEAD" "$@"`, msg: "simulated COMMITTED_RAW read failure", checks: []bmCheck{
			bmRC("unreadable committed-paths list -> exit 2", 2),
			noVerdict("unreadable committed-paths list: emits no verdict line", notVerdict...),
			bmHas("unreadable committed-paths list: names the failure", "committed paths"),
			shimFired("unreadable committed-paths list")}},
		// 9e: --cached is unique to the STAGED_RAW capture.
		{setup: func(fx *bmFx) {
			fx.advanceBase("unrelated1.txt")
			fx.g.write(fx.repo+"/demo.txt", "staged")
			fx.g.git(fx.repo, "add", "demo.txt")
		}, stub: `has --cached "$@"`, msg: "simulated STAGED_RAW read failure", checks: []bmCheck{
			bmRC("unreadable staged-paths list -> exit 2", 2),
			noVerdict("unreadable staged-paths list: emits no verdict line", notVerdict...),
			bmHas("unreadable staged-paths list: names the failure", "staged paths"),
			shimFired("unreadable staged-paths list")}},
		// 9f: `-C <worktree> diff --no-renames --name-only` is the only
		// five-argument call, so the stub matches on the count.
		{setup: func(fx *bmFx) {
			fx.advanceBase("unrelated1.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "unstaged")
		}, stub: `[ "$#" -eq 5 ] && [ "$3" = "diff" ] && [ "$4" = "--no-renames" ] && [ "$5" = "--name-only" ]`,
			msg: "simulated UNSTAGED_RAW read failure", checks: []bmCheck{
				bmRC("unreadable unstaged-paths list -> exit 2", 2),
				noVerdict("unreadable unstaged-paths list: emits no verdict line", notVerdict...),
				bmHas("unreadable unstaged-paths list: names the failure", "unstaged paths"),
				shimFired("unreadable unstaged-paths list")}},
		// 9g: ^HEAD is unique to the UNCARRIED rev-list.
		{setup: func(fx *bmFx) { fx.advanceBase("unrelated1.txt") },
			stub: `[ "$3" = "rev-list" ] && [ "$7" = "^HEAD" ]`, msg: "simulated UNCARRIED read failure", checks: []bmCheck{
				bmRC("unreadable uncarried count -> exit 2", 2),
				noVerdict("unreadable uncarried count: emits no verdict line", notVerdict...),
				bmHas("unreadable uncarried count: names the failure", "uncarried commits"),
				shimFired("unreadable uncarried count")}},
		// 10: origin/main moved, local main left behind; handed the bare name.
		{fresh: true, setup: func(fx *bmFx) {
			fx.withOrigin()
			fx.g.git(fx.repo, "branch", "-q", "tmp-advance", "main")
			fx.g.git(fx.repo, "checkout", "-q", "tmp-advance")
			fx.g.appendLine(fx.repo+"/unrelated.txt", "moved")
			fx.g.git(fx.repo, "add", "unrelated.txt")
			fx.g.git(fx.repo, "commit", "-qm", "advance: unrelated.txt")
			fx.g.git(fx.repo, "push", "-q", "origin", "tmp-advance:main")
			fx.g.git(fx.repo, "checkout", "-q", "demo")
			fx.g.git(fx.repo, "branch", "-q", "-D", "tmp-advance")
			fx.g.git(fx.repo, "fetch", "-q", "origin")
		}, checks: []bmCheck{
			bmPrefix("bare name behind origin, origin moved -> MOVED", "MOVED"),
			bmHas("bare name behind origin: verdict names origin/main", "origin/main")}},
		{fresh: true, args: func(*bmFx) []string { return []string{"", "", ""} }, checks: []bmCheck{
			bmRC("missing arguments -> exit 2", 2),
			noVerdict("missing arguments: emits no verdict line", notVerdict...)}},
		{fresh: true, args: func(*bmFx) []string { return nil }, checks: []bmCheck{
			{"missing arguments: stdout is empty", func(r guardResult, _ *bmFx) bool { return r.stdout == "" }},
			{"missing arguments: stderr is exactly the usage message", func(r guardResult, _ *bmFx) bool {
				return r.err == baseRefUsageWant("check-base-moved.sh")
			}}}},
		{fresh: true, args: func(fx *bmFx) []string { return []string{fx.dir, "main", "deadbeef"} }, checks: []bmCheck{
			bmRC("non-repository -> exit 2", 2),
			noVerdict("non-repository: emits no verdict line", notVerdict...)}},
		// 13: a second stub built while the first sits on PATH still execs
		// the real git — handed the first stub's own match argument, it must
		// reach git rather than the first stub.
		{fresh: true, prep: func(t *testing.T, fx *bmFx) {
			stubGit(t, fx.dir+"/one", `has chain-marker-one "$@"`, "first shim fired")
			stubGit(t, fx.dir+"/two", `has chain-marker-two "$@"`, "second shim fired")
		}, run: func(fx *bmFx) (guardResult, error) {
			cmd := exec.Command(fx.dir+"/two/git", "chain-marker-one")
			cmd.Env = append(os.Environ(), "PATH="+fx.dir+"/one:"+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			rc := 0
			if ee, ok := err.(*exec.ExitError); ok {
				rc = ee.ExitCode()
			} else if err != nil {
				return guardResult{}, err
			}
			return guardResult{rc: rc, out: string(out)}, nil
		}, checks: []bmCheck{
			bmLacks("shim chaining: second shim's real-git did not resolve to the first shim", "first shim fired"),
			{"shim chaining: second shim exited a real-git error, not the first shim's 128", func(r guardResult, _ *bmFx) bool {
				return r.rc != 128
			}}}},
		// 14: movement entirely carried by this branch (KAN-535) is CLEAR.
		{fresh: true, setup: func(fx *bmFx) {
			fx.withOrigin()
			fx.g.git(fx.repo, "commit", "-qm", "demo note commit", "--allow-empty")
			carry := fx.g.git(fx.repo, "rev-parse", "HEAD")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-work")
			fx.g.git(fx.repo, "add", "shared.txt")
			fx.g.git(fx.repo, "commit", "-qm", "demo work on top")
			fx.g.git(fx.repo, "push", "-q", "origin", carry+":main")
			fx.g.git(fx.repo, "fetch", "-q", "origin")
		}, checks: []bmCheck{
			bmPrefix("movement entirely carried -> CLEAR", "CLEAR"),
			bmHas("movement entirely carried: names the carried movement", "already carried by this branch"),
			bmHas("movement entirely carried: verdict names the moved ref origin/main", "origin/main")}},
		// 15: movement only partly carried stays MOVED with the full count.
		{setup: func(fx *bmFx) {
			fx.g.git(fx.repo, "commit", "-qm", "carried commit", "--allow-empty")
			carried := fx.g.git(fx.repo, "rev-parse", "HEAD")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-work")
			fx.g.git(fx.repo, "add", "shared.txt")
			fx.g.git(fx.repo, "commit", "-qm", "demo work")
			fx.g.git(fx.repo, "branch", "-f", "main", carried)
			fx.g.git(fx.repo, "checkout", "-q", "main")
			fx.g.appendLine(fx.repo+"/unrelated1.txt", "unrelated")
			fx.g.git(fx.repo, "add", "unrelated1.txt")
			fx.g.git(fx.repo, "commit", "-qm", "unrelated advance")
			fx.g.git(fx.repo, "checkout", "-q", "demo")
		}, checks: []bmCheck{
			bmPrefix("movement partially carried -> MOVED", "MOVED"),
			bmHas("movement partially carried: names the full commit count", "2 commits")}},
		// Beyond the harness (Review Focus): the overlap list is in the byte
		// order the bash's LC_ALL=C `sort -u` gave, whatever the caller's
		// locale — a collating sort would put `_` and `a` before `Z`. git
		// prints a non-ASCII path quoted, so its own order puts it last where
		// the byte sort of the quoted form (a leading `"`) puts it first.
		{setup: func(fx *bmFx) { fx.overlap("a.txt", "_.txt", "Z.txt", "ä.txt") }, checks: []bmCheck{
			{"port: overlaps are listed in C byte order", func(r guardResult, _ *bmFx) bool {
				return strings.HasSuffix(r.out, `overlaps: "\303\244.txt", Z.txt, _.txt, a.txt`)
			}}}},
		// KAN-298: the usage names the base-ref rule. The harness grepped the
		// script's source; the text now lives in Go, so the printed usage is
		// read instead.
		{fresh: true, args: func(*bmFx) []string { return nil }, checks: []bmCheck{
			{"usage message states the base-ref rule", func(r guardResult, _ *bmFx) bool {
				return strings.Contains(r.err, "prefers refs/remotes/origin/<base-ref>")
			}}}},
	}

	for i, c := range cases {
		// The sandbox belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		fx := &bmFx{dir: t.TempDir(), recorded: recorded}
		if !c.fresh {
			fx.repo = fx.dir + "/repo"
			gdcCopyTree(t, tmpl, fx.repo)
		}
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
			if fx.g.err != nil {
				return guardResult{}, fx.g.err
			}
			if c.run != nil {
				return c.run(fx)
			}
			args := []string{fx.repo, "main", fx.recorded}
			if c.args != nil {
				args = c.args(fx)
			}
			env := Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: os.LookupEnv}
			if fx.stub != "" {
				env.Getenv = pathEnv(fx.stub)
			}
			return runGuard("check-base-moved", args, env), nil
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
					if !chk.ok(r, fx) {
						t.Fatalf("rc=%d out=%s", r.rc, r.out)
					}
				})
			}
		})
	}
}

// bmNewFx is new_repo in its own directory: main with base.txt and
// shared.txt in one commit, demo checked out at it.
func bmNewFx(t *testing.T) *bmFx {
	t.Helper()
	fx := &bmFx{dir: t.TempDir()}
	fx.repo = fx.dir + "/repo"
	fx.g.git("", "init", "-q", "-b", "main", fx.repo)
	fx.g.write(fx.repo+"/base.txt", "base")
	fx.g.write(fx.repo+"/shared.txt", "shared-base")
	fx.g.git(fx.repo, "add", "base.txt", "shared.txt")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.recorded = fx.g.git(fx.repo, "rev-parse", "HEAD")
	fx.g.git(fx.repo, "checkout", "-q", "-b", "demo")
	return fx
}

// TestBaseMovedFullOverlap pins every line check-base-moved prints, exactly,
// as it printed them before baseMoved was split out of it -- the printed line
// keeps its 10-path cut -- and that the verdict a caller reads carries the
// full sorted overlap the line cuts.
func TestBaseMovedFullOverlap(t *testing.T) {
	t.Parallel()
	var wide []string
	for i := 1; i <= 11; i++ {
		wide = append(wide, fmt.Sprintf("wide%02d.txt", i))
	}
	cases := []struct {
		label    string
		setup    func(fx *bmFx)
		args     func(fx *bmFx) []string // nil: {repo, main, recorded}
		code     int
		out, err string // $REPO is the repository, $DIR its parent
		kind     string // the verdict's kind; "" on exit 2
		ref      string // the verdict's resolved base ref; "" before it is resolved
		overlap  []string
	}{
		{"unmoved", nil, nil, 0, "CLEAR: $REPO — main has not moved since the recorded merge base\n", "", "CLEAR", "main", nil},
		{"moved but carried", func(fx *bmFx) {
			fx.advanceBase("unrelated1.txt")
			fx.g.git(fx.repo, "merge", "-q", "--ff-only", "main")
		}, nil, 0, "CLEAR: $REPO — the 1 commits main gained since the recorded merge base are all already carried by this branch — nothing to rebase; merge base now $MAIN\n", "", "CLEAR", "main", nil},
		{"moved, no overlap", func(fx *bmFx) { fx.advanceBase("unrelated1.txt", "unrelated2.txt") }, nil, 0,
			"MOVED: $REPO — 2 commits on main since the recorded merge base; no overlap with this change's paths\n", "", "MOVED", "main", nil},
		{"moved, one overlap", func(fx *bmFx) {
			fx.advanceBase("shared.txt")
			fx.g.appendLine(fx.repo+"/shared.txt", "demo-unstaged")
		}, nil, 0, "MOVED: $REPO — 1 commits on main since the recorded merge base; overlaps: shared.txt\n", "", "MOVED", "main", []string{"shared.txt"}},
		{"moved, eleven overlaps", func(fx *bmFx) { fx.wide(11) }, nil, 0,
			"MOVED: $REPO — 11 commits on main since the recorded merge base; overlaps: " + strings.Join(wide[:10], ", ") + " (+1 more)\n", "", "MOVED", "main", wide},
		{"no recorded merge base", nil, func(fx *bmFx) []string { return []string{fx.repo, "main", "-"} }, 0,
			"REFUSE: no merge base recorded for $REPO — cannot tell whether the base has moved\n", "", "REFUSE", "", nil},
		{"unresolvable recorded merge base", nil, func(fx *bmFx) []string { return []string{fx.repo, "main", "nope"} }, 0,
			"REFUSE: recorded merge base 'nope' does not resolve in $REPO\n", "", "REFUSE", "", nil},
		{"unresolvable base ref", nil, func(fx *bmFx) []string { return []string{fx.repo, "no-such-base", fx.recorded} }, 0,
			"REFUSE: base ref 'no-such-base' does not resolve in $REPO — cannot tell whether the base has moved\n", "", "REFUSE", "no-such-base", nil},
		{"not a directory", nil, func(fx *bmFx) []string { return []string{fx.repo + "/missing", "main", fx.recorded} }, 2,
			"", "check-base-moved: $REPO/missing is not a directory — cannot determine anything\n", "", "", nil},
		{"not a git worktree", func(fx *bmFx) { mkdir(t, fx.dir+"/plain") }, func(fx *bmFx) []string { return []string{fx.dir + "/plain", "main", fx.recorded} }, 2,
			"", "check-base-moved: $DIR/plain is not a git worktree — cannot determine anything\n", "", "", nil},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			fx := bmNewFx(t)
			if c.setup != nil {
				c.setup(fx)
			}
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			args := []string{fx.repo, "main", fx.recorded}
			if c.args != nil {
				args = c.args(fx)
			}
			subst := strings.NewReplacer("$REPO", fx.repo, "$DIR", fx.dir, "$REC", fx.recorded, "$MAIN", fx.g.git(fx.repo, "rev-parse", "main"))
			env := Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: os.LookupEnv}
			r := runGuard("check-base-moved", args, env)
			if r.rc != c.code || r.stdout != subst.Replace(c.out) || r.err != subst.Replace(c.err) {
				t.Fatalf("got exit %d stdout %q stderr %q\nwant exit %d stdout %q stderr %q", r.rc, r.stdout, r.err, c.code, subst.Replace(c.out), subst.Replace(c.err))
			}
			v := baseMoved(env, args[0], args[1], args[2], io.Discard)
			if line := strings.TrimSuffix(subst.Replace(c.out), "\n"); v.code != c.code || v.line != line || v.kind != c.kind || v.ref != c.ref ||
				strings.Join(v.overlap, ",") != strings.Join(c.overlap, ",") {
				t.Fatalf("verdict %+v, want code %d line %q kind %q ref %q overlap %v", v, c.code, line, c.kind, c.ref, c.overlap)
			}
		})
	}
}
