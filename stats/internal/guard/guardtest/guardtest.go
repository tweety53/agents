package guardtest

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// ExecFixtures backs every executable fixture with an inode already
// assessed. macOS assesses a newly written executable on its first exec --
// 0.16-0.31s measured on this machine, serialised machine-wide (twenty fresh
// scripts took 3.06s run concurrently, 3.33s one after another) -- so the
// ~210 executable fixtures this package writes cost ~30s of wall however
// parallel it runs. The assessment holds per inode, across a rewrite with
// other content. So each distinct body gets one master inode here, every
// fixture carrying that body is a hard link to it, and a master no fixture
// links any more is rewritten for the next new body: the cost is bounded by
// the distinct bodies alive at once. Every fixture still carries exactly the
// bytes and mode its test wrote; a test changes one only through WriteExec,
// which unlinks it first.
//
// dir is the one path this package writes outside t.TempDir(): a master must
// outlive the test that first wrote its body, and a t.TempDir() is removed
// when its test ends. Main removes it after the run.
var ExecFixtures struct {
	sync.Mutex
	Dir     string            // created by Main
	masters map[string]string // body -> master inode's path in dir
	n       int               // masters ever named
}

// repoRoot is the repository a test runs in: the nearest directory above the
// test's package directory carrying scripts/flow-guard, so it holds whichever
// guard package a test binary belongs to.
var repoRoot = func() string {
	dir, _ := os.Getwd()
	for d := dir; ; d = filepath.Dir(d) {
		if core.IsFile(filepath.Join(d, "scripts/flow-guard")) {
			return d
		}
		if d == filepath.Dir(d) {
			return dir
		}
	}
}()

// Main is the TestMain of every guard package: m.Run under the fixtures
// below, after a guard name in os.Args has been dispatched like flow-guard.
func Main(m *testing.M) {
	// RunSibling execs this very executable with a guard's name first: in a
	// test binary that is the test binary, so it dispatches like flow-guard.
	if len(os.Args) > 1 {
		if fn, ok := core.Registry[os.Args[1]]; ok {
			cwd, _ := os.Getwd()
			os.Exit(fn(os.Args[2:], core.Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: cwd, Stdin: os.Stdin}, os.Stdout, os.Stderr))
		}
	}
	os.Setenv("FLOW_GUARD_TELEMETRY", "off")
	// A real `flow` on PATH would read the dev store, so a test would follow
	// whatever the operator has stored; a dead records address makes any
	// such read fail fast. A test wanting an answer puts its own fake `flow`
	// on PATH.
	os.Setenv("FLOW_RECORDS_ADDR", "http://127.0.0.1:1")
	// A test binary started with SIGINT or SIGHUP ignored -- a `&` job of a
	// non-interactive shell, or nohup -- hands that ignore to every child it
	// execs, so the signal tests watch a child that cannot die of the signal.
	// Catching and dropping the signal keeps this process as deaf as it
	// started, while exec resets a caught signal to its default in each child.
	for _, sig := range []os.Signal{syscall.SIGINT, syscall.SIGHUP} {
		if signal.Ignored(sig) {
			signal.Notify(make(chan os.Signal, 1), sig)
		}
	}
	dir, err := os.MkdirTemp("", "guard-exec-fixtures")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	ExecFixtures.Dir, ExecFixtures.masters = dir, map[string]string{}
	// Every allowed-pairs reader runs `flow policy get` (PolicyTable). A
	// real `flow` would answer from the dev store, or its cache; a `flow`
	// first on PATH answering TestPolicy -- the seeded pairs -- and handing
	// every other call to the next `flow` on PATH keeps every test on one
	// known table. A test wanting another answer puts its own fake `flow`
	// on PATH ahead of it.
	policyBin := filepath.Join(dir, "policy-bin")
	if err := os.MkdirAll(policyBin, 0o755); err == nil {
		err = os.WriteFile(filepath.Join(policyBin, "flow"), []byte("#!/bin/sh\n"+PolicyStub+"PATH=${PATH#*:} exec flow \"$@\"\n"), 0o755)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	os.Setenv("PATH", policyBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	// Every sibling a guard under test runs is the flow-guard these sources
	// build, as in production; TestRunSiblingSelfExec covers the default.
	bin, err := guardCacheBuilt()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	core.SiblingExe = bin
	// Every guard a test runs through scripts/flow-guard builds into, and
	// finds in, the cache above — never the operator's.
	os.Setenv("FLOW_GUARD_CACHE_DIR", filepath.Dir(filepath.Dir(bin)))
	repo := repoRoot
	before, watched := TreeStatus(repo)
	code := m.Run()
	_ = os.RemoveAll(dir)
	if watched {
		if leak := TreeLeak(repo, before); leak != "" {
			fmt.Fprint(os.Stderr, leak)
			code = 1
		}
	}
	os.Exit(code)
}

// TreeStatus is the repository's `git status --porcelain`; false when it
// cannot be read (no git work tree), which skips the tree gate.
func TreeStatus(repo string) (string, bool) {
	out, err := exec.Command("git", "-C", repo, "status", "--porcelain", "--untracked-files=normal").Output()
	return string(out), err == nil
}

// TreeLeak is the tree gate of the deleted run-guard-tests (kan-584,
// kan-1051 F5): a suite that leaves the repository's status lines changed --
// a test or a reproducer writing into the checkout -- fails, so the leak is
// caught by the run that made it, never by the next guard reading the tree.
// It compares status lines, not content: an edit restored inside the run, or
// one to a file already modified or untracked, is not caught. It returns the
// failure message, or "" when the status is unchanged.
func TreeLeak(repo, before string) string {
	after, ok := TreeStatus(repo)
	if !ok || after == before {
		return ""
	}
	return fmt.Sprintf("guard tests changed the tree of %s:\nbefore:\n%safter:\n%s", repo, before, after)
}

// shasum is the bash helper's own first tool, run on the real file, so the
// NUL-byte vector is the tool's answer rather than one typed in here.
func Shasum(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("shasum", "-a", "256", path).Output()
	if err != nil {
		t.Fatalf("shasum: %v", err)
	}
	return strings.Fields(string(out))[0]
}

func Ptr(s string) *string { return &s }

// PathEnv is an Env.Getenv whose PATH puts bin first, the way the bash
// harness prepends its stub directory; every other name reads the process.
func PathEnv(bin string) func(string) string {
	return func(k string) string {
		if k == "PATH" {
			return bin + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}
}

func Mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func WriteFile(t *testing.T, path, body string) {
	t.Helper()
	Mkdir(t, filepath.Dir(path))
	// Unlinked first: a write through a WriteExec fixture's hard link would
	// land in the shared master and change every sibling's bytes.
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// WriteExec links path to the master inode carrying body (ExecFixtures).
func WriteExec(t *testing.T, path, body string) {
	t.Helper()
	Mkdir(t, filepath.Dir(path))
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	ExecFixtures.Lock()
	defer ExecFixtures.Unlock()
	m, ok := ExecFixtures.masters[body]
	if !ok {
		m = idleMaster()
		if err := os.WriteFile(m, []byte(body), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(m, 0o755); err != nil {
			t.Fatal(err)
		}
		ExecFixtures.masters[body] = m
	}
	if err := os.Link(m, path); err != nil {
		t.Fatal(err)
	}
}

// fixtureRoot is a bash expression for the tree a fixture at rel sits in,
// derived from the script's own path: a reproducer is exec'd by its
// resolved absolute path, so $0 always carries a directory.
func fixtureRoot(rel string) string {
	return "${0%/*}/" + strings.Repeat("../", strings.Count(rel, "/"))
}

// idleMaster takes a master no fixture links any more, else names a new one.
func idleMaster() string {
	for body, m := range ExecFixtures.masters {
		var st syscall.Stat_t
		if syscall.Stat(m, &st) == nil && st.Nlink == 1 {
			delete(ExecFixtures.masters, body)
			return m
		}
	}
	ExecFixtures.n++
	return filepath.Join(ExecFixtures.Dir, strconv.Itoa(ExecFixtures.n))
}

// GuardCache is the FLOW_GUARD_CACHE_DIR every test shares whose shim only
// has to reach the guard, the binary built into it once per run. A cache of
// its own costs each such test a `go build` of flow-guard and the first exec
// of a new binary, which macOS assesses serialised machine-wide
// (ExecFixtures). The key is the sources' hash, so a shim run from any
// checkout of these sources finds the one binary -- under this process's
// locale: flow_guard_key's glob sorts by collation, so a shim run with no
// LANG (check_plan_provenance's) keys and builds a second. A test about the
// build itself -- no go, a broken go -- keeps an empty cache of its own.
// Lives in ExecFixtures.dir, which Main removes after the run.
func GuardCache(t *testing.T) string {
	t.Helper()
	return filepath.Dir(filepath.Dir(GuardBinary(t)))
}

// GuardBinary is the one flow-guard GuardCache built. A test that execs the
// binary itself takes this path, never a glob of the cache: a no-LANG shim
// sharing the cache adds a second key's binary beside it at any time.
func GuardBinary(t *testing.T) string {
	t.Helper()
	bin, err := guardCacheBuilt()
	if err != nil {
		t.Fatal(err)
	}
	return bin
}

var guardCacheBuilt = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(ExecFixtures.Dir, "flow-guard-cache")
	boot := filepath.Join(repoRoot, "scripts/flow-guard")
	// The bootstrap builds under its key, then execs the binary;
	// spec-root's verdict is not read.
	cmd := exec.Command(boot, "spec-root")
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+dir)
	out, _ := cmd.CombinedOutput()
	bins, _ := filepath.Glob(dir + "/*/flow-guard")
	if len(bins) != 1 {
		return "", fmt.Errorf("building the shared flow-guard cache left %d binaries in %s:\n%s", len(bins), dir, out)
	}
	return bins[0], nil
})

// FixtureGit is the git binary every fixture runs, and every git stub a test
// writes execs as the real one: the binary `git` on PATH runs, resolved once
// through its own --exec-path. On macOS /usr/bin/git is an xcrun trampoline
// that spawns the real binary, which doubles each call's cost -- measured
// here at 14.5ms CPU and 26ms wall against 8ms and 12ms direct -- and the
// fixtures spawn thousands. The guards under test still resolve `git` from
// PATH themselves. Plain "git" when the lookup fails.
var FixtureGit = func() string {
	out, err := exec.Command("git", "--exec-path").Output()
	if err != nil {
		return "git"
	}
	p, err := filepath.EvalSymlinks(strings.TrimSpace(string(out)) + "/git")
	if err != nil {
		return "git"
	}
	return p
}()

// FixtureGitPath is PATH with a directory holding only `git`, a link to
// FixtureGit, put first: a guard under test that resolves git over its PATH
// then skips the xcrun trampoline too, which halves each of its git calls.
func FixtureGitPath(t *testing.T) string {
	t.Helper()
	dir, err := FixtureGitDir()
	if err != nil {
		t.Fatal(err)
	}
	return dir + string(os.PathListSeparator) + os.Getenv("PATH")
}

var FixtureGitDir = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(ExecFixtures.Dir, "git-bin")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, os.Symlink(FixtureGit, filepath.Join(dir, "git"))
})

// FixtureGitEnv is what every fixture-building git runs under: the user's
// global config masked, as the bash harnesses' GIT_CONFIG_GLOBAL=/dev/null
// did, no `git maintenance run --auto` child spawned after each commit, and
// the files ref backend that tcfRepo.head()'s loose-ref read assumes, whatever
// init.defaultRefFormat says.
var FixtureGitEnv = []string{"GIT_CONFIG_GLOBAL=/dev/null",
	"GIT_CONFIG_COUNT=1", "GIT_CONFIG_KEY_0=maintenance.auto", "GIT_CONFIG_VALUE_0=false",
	"GIT_DEFAULT_REF_FORMAT=files"}

// GitRun runs git under FixtureGitEnv with a fixed identity, as the bash
// harness's GIT_CONFIG_GLOBAL=/dev/null plus its `config user.*` do.
func GitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	if err := fixtureGitExec(dir, args...); err != nil {
		t.Fatal(err)
	}
}

// fixtureGitExec is GitRun for code with no *testing.T: the shape templates below.
func fixtureGitExec(dir string, args ...string) error {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command(FixtureGit, args...)
	cmd.Env = append(append(os.Environ(), FixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git %v: %v\n%s", args, err, out)
	}
	return nil
}

// RepoShape is one recurring fixture repository: git init [-b Branch] [--bare], plus one
// empty commit with message Commit when Commit is set.
type RepoShape struct {
	Branch, Commit string
	Bare           bool
}

// shapeTemplates holds each shape built once per package run under ExecFixtures.dir, through
// fixtureGitExec's fixed identity, and copied per test: a `git init` plus a commit is a handful of
// execs, and the package's fixtures make one per test. A template is a directory; for the
// origin shapes it holds origin.git (bare) and clone/ (Commit pushed to it).
var shapeTemplates struct {
	sync.Mutex
	plain, origin map[RepoShape]func() (string, error)
}

func shapeTemplate(s RepoShape, withOrigin bool) (string, error) {
	shapeTemplates.Lock()
	m := &shapeTemplates.plain
	if withOrigin {
		m = &shapeTemplates.origin
	}
	if *m == nil {
		*m = map[RepoShape]func() (string, error){}
	}
	build := (*m)[s]
	if build == nil {
		build = sync.OnceValues(func() (string, error) { return buildShape(s, withOrigin) })
		(*m)[s] = build
	}
	shapeTemplates.Unlock()
	return build()
}

func buildShape(s RepoShape, withOrigin bool) (string, error) {
	dir, err := os.MkdirTemp(ExecFixtures.Dir, "shape")
	if err != nil {
		return "", err
	}
	init := func(repo string, bare bool) error {
		args := []string{"init", "-q"}
		if s.Branch != "" {
			args = append(args, "-b", s.Branch)
		}
		if bare {
			args = append(args, "--bare")
		}
		return fixtureGitExec("", append(args, repo)...)
	}
	if !withOrigin {
		if err := init(dir+"/repo", s.Bare); err != nil {
			return "", err
		}
		if s.Commit != "" {
			// ponytail: no bare shape carries a commit yet; build it through a clone when one does.
			if s.Bare {
				return "", fmt.Errorf("repoShape %+v: a bare shape with a commit is not built", s)
			}
			if err := fixtureGitExec(dir+"/repo", "commit", "-q", "--allow-empty", "-m", s.Commit); err != nil {
				return "", err
			}
		}
		return dir + "/repo", nil
	}
	if s.Commit == "" || s.Branch == "" || s.Bare {
		return "", fmt.Errorf("repoShape %+v: an origin shape names a Branch and a Commit and is not bare", s)
	}
	origin, clone := dir+"/origin.git", dir+"/clone"
	for _, step := range []func() error{
		func() error { return init(origin, true) },
		func() error { return init(clone, false) },
		func() error { return fixtureGitExec(clone, "commit", "-q", "--allow-empty", "-m", s.Commit) },
		func() error { return fixtureGitExec(clone, "remote", "add", "origin", origin) },
		func() error { return fixtureGitExec(clone, "push", "-q", "-u", "origin", s.Branch) },
		// origin/HEAD, as `git clone` records it.
		func() error { return fixtureGitExec(clone, "remote", "set-head", "origin", s.Branch) },
	} {
		if err := step(); err != nil {
			return "", err
		}
	}
	return dir, nil
}

// CopyShape copies shape s into dir: a bare shape's repository is dir itself, any other's .git
// is dir/.git; dir is created if missing and files already in it are kept. It takes no
// *testing.T, so a sync.OnceValues master can build on it.
func CopyShape(dir string, s RepoShape) error {
	tpl, err := shapeTemplate(s, false)
	if err != nil {
		return err
	}
	if s.Bare {
		return CopyTree(tpl, dir)
	}
	return CopyTree(tpl+"/.git", dir+"/.git")
}

// FixtureRepo is CopyShape, failing t on error.
func FixtureRepo(t *testing.T, dir string, s RepoShape) {
	t.Helper()
	if err := CopyShape(dir, s); err != nil {
		t.Fatal(err)
	}
}

// FixtureRepoWithOrigin copies the origin shape: a bare repository at origin and a clone at
// dir with s.Commit on s.Branch pushed to it and tracking it, origin/HEAD naming s.Branch as
// a clone's does, its origin URL rewritten to
// origin in Go. Worktrees and further commits are made after the copy: gitdir paths are
// absolute.
func FixtureRepoWithOrigin(t *testing.T, dir, origin string, s RepoShape) {
	t.Helper()
	tpl, err := shapeTemplate(s, true)
	if err == nil {
		err = CopyTree(tpl+"/origin.git", origin)
	}
	if err == nil {
		err = CopyTree(tpl+"/clone/.git", dir+"/.git")
	}
	if err != nil {
		t.Fatal(err)
	}
	cfg := dir + "/.git/config"
	b, err := os.ReadFile(cfg)
	if err != nil {
		t.Fatal(err)
	}
	from, to := "\turl = "+tpl+"/origin.git\n", "\turl = "+origin+"\n"
	if !bytes.Contains(b, []byte(from)) {
		t.Fatalf("%s carries no %q", cfg, from)
	}
	if err := os.WriteFile(cfg, bytes.Replace(b, []byte(from), []byte(to), 1), 0o644); err != nil {
		t.Fatal(err)
	}
}

// CopyTree copies src to dst — regular files, directories and symlinks — keeping each one's
// mode and mtime; dst and its directories are created when missing, other files in them kept.
func CopyTree(src, dst string) error {
	type dirTime struct {
		path string
		mt   time.Time
	}
	var dirs []dirTime
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		to := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		switch {
		case d.IsDir():
			if err := os.MkdirAll(to, 0o755); err != nil {
				return err
			}
			dirs = append(dirs, dirTime{to, info.ModTime()})
			return os.Chmod(to, info.Mode().Perm())
		case d.Type()&fs.ModeSymlink != 0:
			l, err := os.Readlink(p)
			if err != nil {
				return err
			}
			return os.Symlink(l, to)
		case d.Type().IsRegular():
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			if err := os.WriteFile(to, b, info.Mode().Perm()); err != nil {
				return err
			}
			if err := os.Chmod(to, info.Mode().Perm()); err != nil {
				return err
			}
			return os.Chtimes(to, info.ModTime(), info.ModTime())
		}
		return fmt.Errorf("copyTree: %s is neither a file, a directory nor a symlink", p)
	})
	if err != nil {
		return err
	}
	// Children set their directory's mtime as they are written, so directories go last,
	// deepest first.
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := os.Chtimes(dirs[i].path, dirs[i].mt, dirs[i].mt); err != nil {
			return err
		}
	}
	return nil
}

// HermeticGitLookup is os.LookupEnv answering GIT_CONFIG_GLOBAL=/dev/null, so
// a guard that forwards it runs git without the operator's ~/.gitconfig.
func HermeticGitLookup(k string) (string, bool) {
	if k == "GIT_CONFIG_GLOBAL" {
		return "/dev/null", true
	}
	return os.LookupEnv(k)
}

// HistoricalFile is <rev>:<path> read from the repository at root — the bash
// a parity test pins its port against, as it stood at <rev>. A clone that
// does not carry <rev>, a shallow one, skips the calling test naming the
// commit, rather than failing the whole package on history it never had
// (KAN-905).
func HistoricalFile(t *testing.T, root, rev, path string) []byte {
	t.Helper()
	if exec.Command(FixtureGit, "-C", root, "cat-file", "-e", rev+"^{commit}").Run() != nil {
		t.Skipf("%s is not in this clone's history (a shallow clone?) — the parity check against the bash at it cannot run", rev)
	}
	src, err := exec.Command(FixtureGit, "-C", root, "show", rev+":"+path).Output()
	if err != nil {
		t.Fatalf("git show %s:%s: %v", rev, path, err)
	}
	return src
}

// bashUsageRe matches a shim-era usage line's script name.
var bashUsageRe = regexp.MustCompile(`usage: ([a-z][a-z0-9-]*)\.sh\b`)

// BashUsage respells a bash-at-base usage line the way the guard now spells
// it, so a parity test compares everything else byte for byte.
func BashUsage(s string) string {
	return bashUsageRe.ReplaceAllString(s, "usage: flow-guard $1")
}

// GuardShim writes an executable that runs `scripts/flow-guard <name>` of
// this checkout with its arguments, for a test that hands a guard to code
// expecting one executable path — a sibling override, a PATH entry.
func GuardShim(t *testing.T, name string) string {
	t.Helper()
	boot := filepath.Join(repoRoot, "scripts/flow-guard")
	p := filepath.Join(t.TempDir(), name)
	WriteExec(t, p, "#!/bin/sh\nexec '"+boot+"' "+name+" \"$@\"\n")
	return p
}

// TestPolicy is `flow policy get` as 0061_role_pairs.sql seeds it: the
// Allowed pairs table as it stood when the pairs moved into flowd
// (testdata/policy.json, which internal/repotest's hook tests read too).
// Embedded, not read off the working directory: a test binary self-exec'd
// as a guard (RunSibling) runs from wherever its caller is, and -trimpath
// makes a runtime.Caller path module-relative, so neither locates the file
// (kan-1090 self-review planning).
//
//go:embed testdata/policy.json
var testPolicyJSON string

var TestPolicy = strings.TrimSpace(testPolicyJSON)

// PolicyStub is the arm of a fake `flow` answering `policy get` with
// TestPolicy; every fake flow of this package's policy readers opens with it.
var PolicyStub = "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '" + TestPolicy + "'; exit 0; }\n"

// NoFlowEnv is an env whose only `flow` answers `policy get` with
// TestPolicy and fails everything else: the seeded pairs alone.
func NoFlowEnv(t *testing.T) core.Env {
	t.Helper()
	return fakeFlowEnv(t, "exit 1\n")
}

// fakeFlowEnv is an env whose PATH holds only a `flow` answering `policy
// get` with TestPolicy and running script for everything else.
func fakeFlowEnv(t *testing.T, script string, over ...string) core.Env {
	t.Helper()
	return FakeFlowEnvBare(t, PolicyStub+script, over...)
}

// PolicyFlowEnv is an env whose only `flow` answers `policy get` with
// policy and fails everything else.
func PolicyFlowEnv(t *testing.T, policy string, over ...string) core.Env {
	t.Helper()
	return FakeFlowEnvBare(t, "[ \"$1 $2\" = \"policy get\" ] && { printf '%s\\n' '"+policy+"'; exit 0; }\nexit 1\n", over...)
}

// FakeFlowEnvBare is an env whose PATH holds only a `flow` running script.
func FakeFlowEnvBare(t *testing.T, script string, over ...string) core.Env {
	t.Helper()
	bin := t.TempDir()
	WriteExec(t, bin+"/flow", "#!/bin/sh\n"+script)
	return EnvWith(t, append([]string{"PATH=" + bin}, over...))
}

// bt turns the ¤ placeholder into a backtick: a Go raw string cannot hold one.
func bt(s string) string { return strings.ReplaceAll(s, "¤", "`") }

// has is the harness's `case "$OUT" in *a*b*)`: every part, in order.
func has(s string, parts ...string) bool {
	for _, p := range parts {
		i := strings.Index(s, p)
		if i < 0 {
			return false
		}
		s = s[i+len(p):]
	}
	return true
}

// FxGit builds fixtures under FixtureGitEnv with GitRun's fixed identity. The
// first failure is kept and every later call skipped, so a fixture built
// inside a case's run reports to each of its checks.
type FxGit struct{ Err error }

func (g *FxGit) Git(dir string, args ...string) string {
	if g.Err != nil {
		return ""
	}
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command(FixtureGit, args...)
	cmd.Env = append(append(os.Environ(), FixtureGitEnv...),
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	var errb bytes.Buffer
	cmd.Stderr = &errb
	out, err := cmd.Output()
	if err != nil {
		g.Err = fmt.Errorf("git %v: %v\n%s", args, err, errb.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// CopyShape is CopyShape (helpers_test.go) into dir, its error held like a git step's.
func (g *FxGit) CopyShape(dir string, s RepoShape) {
	if g.Err == nil {
		g.Err = CopyShape(dir, s)
	}
}

// appendLine is `echo <line> >> <path>`.
func (g *FxGit) AppendLine(path, line string) {
	if g.Err != nil {
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
		g.Err = err
	}
}

// write is `echo <line> > <path>`.
func (g *FxGit) Write(path, line string) {
	if g.Err == nil {
		g.Err = os.WriteFile(path, []byte(line+"\n"), 0o644)
	}
}

// StubGit writes <dir>/git: when cond (a bash test over the stub's own "$@";
// has <arg> "$@" is shim_failing_git's exact-element match) holds, it marks
// <dir>/.fired and fails as the harness's shims did; otherwise it execs the
// git on the test process's PATH, resolved here once, so a stub directory
// already on PATH can never be the "real" git.
func StubGit(t *testing.T, dir, cond, message string) {
	t.Helper()
	WriteExec(t, dir+"/git", fmt.Sprintf(`#!/usr/bin/env bash
has() { local m="$1" a; shift; for a in "$@"; do [ "$a" = "$m" ] && return 0; done; return 1; }
if %s; then
  : > "${0%%/*}/.fired"
  echo "fatal: %s" >&2
  exit 128
fi
exec %q "$@"
`, cond, message, FixtureGit))
}

// GuardResult is one in-process run: stdout, stderr, and the two merged in
// write order with trailing newlines stripped, as the harness's
// OUT="$(guard 2>&1)" read them.
type GuardResult struct {
	Rc               int
	Out, Stdout, Err string
}

// lockedWriter serialises writes: a guard's child process has its stdout and
// stderr copied by two goroutines, and both land in one merged buffer.
type lockedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (l lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

// runFunc runs the guard fn, looked up under name, in process: its exit and
// its merged, stdout and stderr output.
func RunFunc(name string, fn core.Func, args []string, env core.Env) GuardResult {
	if fn == nil {
		return GuardResult{Rc: -1, Out: name + " is not registered"}
	}
	var mu sync.Mutex
	var all, o, e bytes.Buffer
	// A guard that hands both writers to a child process gets two copying
	// goroutines, so the merged buffer they share takes a lock.
	rc := fn(args, env, lockedWriter{&mu, io.MultiWriter(&o, &all)}, lockedWriter{&mu, io.MultiWriter(&e, &all)})
	return GuardResult{rc, strings.TrimRight(all.String(), "\n"), o.String(), e.String()}
}

// EnvWith is an Env over the test process's own environment and working
// directory with over's NAME=value pairs laid on top, the same pairs
// bashLibEnv hands the bash side.
func EnvWith(t *testing.T, over []string) core.Env {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	for _, kv := range over {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	look := func(k string) (string, bool) {
		if v, ok := m[k]; ok {
			return v, true
		}
		return os.LookupEnv(k)
	}
	return core.Env{Dir: cwd, LookupEnv: look, Getenv: func(k string) string { v, _ := look(k); return v }}
}
