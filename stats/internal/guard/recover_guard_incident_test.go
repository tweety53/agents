package guard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// The cases of scripts/test-recover-guard-incident.sh at c5379c0a that
// outlived the planning-file restore (KAN-1073), one subtest per ok: label,
// plus port: rows pinning every refusal's stderr line and every dry-run/apply
// stdout byte for byte, every refusal leaving `status --porcelain`, HEAD and
// REVERT_HEAD as it found them, and the repo-dir resolution the harness never
// reached. new_repo is built once and copied per case, and the guard runs
// in-process over the copy: the guard's --apply mutates git.

// rgiState is what a case observes of its repository, read once before the
// run and once after it, so no parallel subtest runs git on the repo.
type rgiState struct {
	revert                        bool
	porcelain, head, stash, index string
}

type rgiFx struct {
	dir, repo, cwd string // repo is physical, as the harness's pwd -P REPO
	g              fxGit
	reflog         string // `git reflog -g HEAD -n 15` before the run
	before, after  rgiState
}

// out is a git read under the environment the guard's own git runs in, so
// the reflog it captures is the one the guard prints.
func (fx *rgiFx) out(args ...string) (string, bool) {
	cmd := exec.Command(fixtureGit, append([]string{"-C", fx.repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	b, err := cmd.Output()
	return string(b), err == nil
}

func (fx *rgiFx) state() rgiState {
	var s rgiState
	_, s.revert = fx.out("rev-parse", "-q", "--verify", "REVERT_HEAD")
	s.porcelain, _ = fx.out("status", "--porcelain", "-uall")
	s.head, _ = fx.out("rev-parse", "HEAD")
	s.stash, _ = fx.out("stash", "list", "--format=%H %gs")
	s.index, _ = fx.out("diff", "--cached")
	return s
}

// startRevert is start_conflicting_revert: REVERT_HEAD set, failure ignored.
func (fx *rgiFx) startRevert() {
	if fx.g.err != nil {
		return
	}
	cmd := exec.Command(fixtureGit, "-C", fx.repo, "revert", "--no-commit", "HEAD~1")
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	_ = cmd.Run()
}

func (fx *rgiFx) write(rel, body string) {
	if fx.g.err != nil {
		return
	}
	p := filepath.Join(fx.repo, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		fx.g.err = err
		return
	}
	fx.g.err = os.WriteFile(p, []byte(body), 0o644)
}

type rgiCheck struct {
	label string
	ok    func(r guardResult, fx *rgiFx) bool
}

func rgiRC(label string, rc int) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return r.rc == rc }}
}

func rgiNoStdout(label string) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return r.stdout == "" }}
}

func rgiRevert(label string, want bool) rgiCheck {
	return rgiCheck{label, func(_ guardResult, fx *rgiFx) bool { return fx.after.revert == want }}
}

// rgiUnchanged is the Review Focus row: a refusal leaves the repository's
// porcelain, HEAD, REVERT_HEAD, index and stash list as it found them.
func rgiUnchanged(label string) rgiCheck {
	return rgiCheck{label, func(_ guardResult, fx *rgiFx) bool { return fx.after == fx.before }}
}

// rgiErrIs pins the whole stderr, built from the case's fixture.
func rgiErrIs(label string, want func(fx *rgiFx) string) rgiCheck {
	return rgiCheck{label, func(r guardResult, fx *rgiFx) bool { return r.err == want(fx) }}
}

func TestRecoverGuardIncident(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	// new_repo, built once and copied per case: one tracked file three
	// commits touch, so a revert of HEAD~1 leaves REVERT_HEAD set.
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", tmpl)
	for i, body := range []string{"one", "two", "one"} {
		g.write(tmpl+"/shared.txt", body)
		g.git(tmpl, "add", "shared.txt")
		g.git(tmpl, "commit", "-qm", []string{"base", "two", "three"}[i])
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	const self = "recover-guard-incident: "

	type rgiCase struct {
		fresh  bool                  // a plain directory, no repository
		setup  func(fx *rgiFx)       // nil: new_repo only
		args   func(*rgiFx) []string // nil: {repo}
		checks []rgiCheck
	}
	defaultPlan := func(fx *rgiFx) { fx.startRevert() }
	repoArg := func(extra ...string) func(*rgiFx) []string {
		return func(fx *rgiFx) []string { return append([]string{fx.repo}, extra...) }
	}
	plan := func(fx *rgiFx) string {
		return "reflog diagnosis — the alternating reset pattern the incident showed:\n" + fx.reflog +
			"plan:\n  git -C " + fx.repo + " revert --abort\n"
	}
	cases := []rgiCase{
		// Case 1: usage-error-exit-2 — a plain directory that is not a git repo.
		{fresh: true, args: func(fx *rgiFx) []string { return []string{fx.dir} }, checks: []rgiCheck{
			rgiRC("case 1: exits 2 on a non-repo directory", 2),
			rgiNoStdout("case 1: writes nothing to stdout"),
			{"case 1: stderr carries the cause only, not the exit code", func(r guardResult, _ *rgiFx) bool {
				return !strings.HasSuffix(strings.TrimRight(r.err, "\n"), " 2")
			}},
			rgiErrIs("case 1: port: stderr is the not-a-repository line", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.repo + "\n"
			})}},
		// Case 1b: an unknown flag after repo-dir is a usage error.
		{setup: defaultPlan, args: repoArg("--bogus"), checks: []rgiCheck{
			rgiRC("case 1b: unknown flag exits 2", 2),
			rgiNoStdout("case 1b: stdout empty on the usage error"),
			rgiErrIs("case 1b: port: stderr is the unknown-option line", func(*rgiFx) string {
				return self + "unknown option: --bogus\n"
			}),
			rgiUnchanged("case 1b: refusal leaves porcelain, HEAD and REVERT_HEAD unchanged")}},
		// A path after repo-dir is a usage error: the tool restores nothing.
		{setup: defaultPlan, args: repoArg("spectre/changes"), checks: []rgiCheck{
			rgiRC("a path after repo-dir exits 2", 2),
			rgiNoStdout("a path after repo-dir: stdout empty"),
			rgiErrIs("a path after repo-dir: stderr names it", func(*rgiFx) string {
				return self + "unexpected argument: spectre/changes\n"
			}),
			rgiUnchanged("a path after repo-dir: nothing changed")}},
		// Case 2: no-revert-refusal — no revert in progress, nothing changed.
		{checks: []rgiCheck{
			rgiRC("case 2: exits 1 with no revert in progress", 1),
			rgiNoStdout("case 2: refusal leaves stdout empty"),
			rgiErrIs("case 2: port: stderr is the no-revert line", func(fx *rgiFx) string {
				return self + "no revert in progress in " + fx.repo + " (REVERT_HEAD missing) — nothing to recover\n"
			}),
			rgiUnchanged("case 2: refusal leaves porcelain, HEAD and REVERT_HEAD unchanged")}},
		// Dry-run: the reflog diagnosis and the abort plan, nothing changed.
		{setup: defaultPlan, checks: []rgiCheck{
			rgiRC("dry-run exits 0", 0),
			{"dry-run: stdout is the plan byte for byte, stderr empty", func(r guardResult, fx *rgiFx) bool {
				return r.stdout == plan(fx) && r.err == ""
			}},
			rgiUnchanged("dry-run leaves porcelain, HEAD and REVERT_HEAD unchanged")}},
		// Apply: the plan, then the abort, run.
		{setup: defaultPlan, args: func(fx *rgiFx) []string { return []string{"--apply", fx.repo} }, checks: []rgiCheck{
			rgiRC("apply exits 0", 0),
			rgiRevert("apply: the revert is aborted", false),
			{"apply: stdout is the plan then the run line", func(r guardResult, fx *rgiFx) bool {
				return r.stdout == plan(fx)+"running: git -C "+fx.repo+" revert --abort\n"
			}},
			{"apply: HEAD is what it was", func(_ guardResult, fx *rgiFx) bool { return fx.after.head == fx.before.head }}}},
		// Port: the usage paths the harness never reached.
		{fresh: true, args: func(*rgiFx) []string { return []string{"--help"} }, checks: []rgiCheck{
			rgiRC("port: --help exits 0", 0),
			{"port: --help prints the usage on stdout, stderr empty", func(r guardResult, _ *rgiFx) bool {
				return r.stdout == "usage: recover-guard-incident [--apply] [repo-dir]\n"+
					"  --apply    execute; default is a dry-run plan\n"+
					"  repo-dir   git repository, default cwd\n" && r.err == ""
			}}}},
		{setup: defaultPlan, args: repoArg("--apply"), checks: []rgiCheck{
			rgiRC("port: --apply after repo-dir is an unknown option: exits 2", 2),
			rgiErrIs("port: --apply after repo-dir: stderr names it", func(*rgiFx) string {
				return self + "unknown option: --apply\n"
			}),
			rgiUnchanged("port: --apply after repo-dir: nothing changed")}},
		{setup: func(fx *rgiFx) { fx.write("file", "") }, args: func(fx *rgiFx) []string {
			return []string{fx.repo + "/file"}
		}, checks: []rgiCheck{
			rgiRC("port: repo-dir that is a file: exits 2", 2),
			rgiNoStdout("port: repo-dir that is a file: stdout empty"),
			rgiErrIs("port: repo-dir that is a file: stderr names it as given", func(fx *rgiFx) string {
				return self + "not a directory: " + fx.repo + "/file\n"
			})}},
		// `cd "$1"` refuses these three where a lexical Join would not: an
		// empty argument (bash 5's "null directory"), a missing component
		// before `..`, and a directory without search permission.
		{setup: func(fx *rgiFx) { defaultPlan(fx); fx.cwd = fx.repo }, args: func(*rgiFx) []string { return []string{"--apply", ""} }, checks: []rgiCheck{
			rgiRC("port: empty repo-dir under --apply: exits 2", 2),
			rgiNoStdout("port: empty repo-dir under --apply: stdout empty"),
			rgiErrIs("port: empty repo-dir under --apply: stderr names it", func(*rgiFx) string {
				return self + "not a directory: \n"
			}),
			rgiUnchanged("port: empty repo-dir under --apply: nothing changed")}},
		// cd resolves an absolute `<symlink>/..` lexically, to the link's
		// own parent -- here the sandbox, not the repository.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("sub/deep/x", "")
			if err := os.Symlink(fx.repo+"/sub/deep", fx.dir+"/ldeep"); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{"--apply", fx.dir + "/ldeep/.."} }, checks: []rgiCheck{
			rgiRC("port: absolute symlink/.. under --apply: exits 2", 2),
			rgiErrIs("port: absolute symlink/.. under --apply: resolved lexically", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.dir + "\n"
			}),
			rgiUnchanged("port: absolute symlink/.. under --apply: nothing changed")}},
		// When the lexical path does not exist, cd falls back to the argument
		// as given, resolved physically through the symlink -- here into the
		// repository, so the bash acts.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("a/b/x", "")
			fx.write("a/sub/x", "")
			if err := os.Symlink(fx.repo+"/a/b", fx.dir+"/yb"); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{fx.dir + "/yb/../sub"} }, checks: []rgiCheck{
			rgiRC("port: absolute symlink/../x with no lexical target falls back to the physical path: exits 0", 0)}},
		// Logical cd checks only the lexical path: `w/c/../wsub` is the
		// directory w/wsub (here a plain directory), though its physical
		// reading through the w/c link does not exist.
		{setup: func(fx *rgiFx) {
			fx.write("a/c/x", "")
			for _, d := range []string{"/w", "/other"} {
				if err := os.Mkdir(fx.dir+d, 0o755); err != nil {
					panic(err)
				}
			}
			for l, tgt := range map[string]string{"/w/c": "../repo/a/c", "/w/wsub": "../other"} {
				if err := os.Symlink(tgt, fx.dir+l); err != nil {
					panic(err)
				}
			}
		}, args: func(*rgiFx) []string { return []string{"w/c/../wsub"} }, checks: []rgiCheck{
			rgiRC("port: a lexically valid symlink/.. path is used though its physical reading is absent: exits 2", 2),
			rgiErrIs("port: a lexically valid symlink/.. path resolves lexically", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.dir + "/other\n"
			})}},
		// A prefix before a `..` that is not a directory fails the lexical
		// canonicalisation, so cd falls back to the physical path -- into
		// the repository, never the lexical lv/sub.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("a/b/x", "")
			fx.write("a/m/x", "")
			fx.write("a/sub/x", "")
			for _, d := range []string{"/lv", "/other"} {
				if err := os.Mkdir(fx.dir+d, 0o755); err != nil {
					panic(err)
				}
			}
			for l, tgt := range map[string]string{"/lv/lb": "../repo/a/b", "/lv/sub": "../other"} {
				if err := os.Symlink(tgt, fx.dir+l); err != nil {
					panic(err)
				}
			}
		}, args: func(*rgiFx) []string { return []string{"lv/lb/../m/../sub"} }, checks: []rgiCheck{
			rgiRC("port: a missing prefix before .. falls back to the physical path: exits 0", 0)}},
		{setup: func(fx *rgiFx) { fx.write("sub/x", "") }, args: func(*rgiFx) []string {
			return []string{"repo/nonexist/../sub"}
		}, checks: []rgiCheck{
			rgiRC("port: missing component before ..: exits 2", 2),
			rgiErrIs("port: missing component before ..: stderr names it as given", func(*rgiFx) string {
				return self + "not a directory: repo/nonexist/../sub\n"
			})}},
		{setup: func(fx *rgiFx) {
			if err := os.Mkdir(fx.repo+"/d", 0o600); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{fx.repo + "/d"} }, checks: []rgiCheck{
			rgiRC("port: repo-dir without search permission: exits 2", 2),
			rgiErrIs("port: repo-dir without search permission: stderr names it as given", func(fx *rgiFx) string {
				return self + "not a directory: " + fx.repo + "/d\n"
			})}},
	}

	for i, c := range cases {
		// The sandbox belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		dir, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		fx := &rgiFx{dir: dir, repo: dir, cwd: dir}
		if !c.fresh {
			fx.repo = dir + "/repo"
			gdcCopyTree(t, tmpl, fx.repo)
		}
		run := func() (guardResult, error) {
			if c.setup != nil {
				c.setup(fx)
			}
			if fx.g.err != nil {
				return guardResult{}, fx.g.err
			}
			if !c.fresh {
				fx.reflog, _ = fx.out("reflog", "-g", "HEAD", "-n", "15")
				fx.before = fx.state()
			}
			args := []string{fx.repo}
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("recover-guard-incident", args, Env{Dir: fx.cwd, Getenv: os.Getenv, LookupEnv: os.LookupEnv})
			if r.rc == -1 {
				return r, errors.New(r.out)
			}
			if !c.fresh {
				fx.after = fx.state()
			}
			return r, nil
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
						t.Fatalf("rc=%d\nstdout=%s\nstderr=%s", r.rc, r.stdout, r.err)
					}
				})
			}
		})
	}
}
