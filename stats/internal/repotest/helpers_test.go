package repotest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// repoRoot is the repository root: this package sits at stats/internal/repotest.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

// goCacheEnv keeps the Go build and module caches where they are once HOME
// moves to a temp dir, so a hook that builds flow-guard does not start cold.
func goCacheEnv() []string {
	var env []string
	if d, err := os.UserCacheDir(); err == nil && os.Getenv("GOCACHE") == "" {
		env = append(env, "GOCACHE="+filepath.Join(d, "go-build"))
	}
	if h, err := os.UserHomeDir(); err == nil && os.Getenv("GOMODCACHE") == "" && os.Getenv("GOPATH") == "" {
		env = append(env, "GOMODCACHE="+filepath.Join(h, "go", "pkg", "mod"))
	}
	return env
}

// guardCacheDir is the directory sharedGuardCache made, "" until then.
var guardCacheDir string

// pkgDir holds the package run's shared executables: the stub trampoline
// writeStub links to and the bootstrap copy copyBootstrap links to. macOS
// checks every newly written executable on its first exec (0.2-0.5s, and
// serialized across parallel tests), so each test linking one already-run
// file instead of writing its own is what keeps the hook tests cheap.
var pkgDir string

// sharedGuardCache makes one temp directory and builds flow-guard into it
// once per package run, as internal/guard's guardCacheBuilt does.
var sharedGuardCache = sync.OnceValues(func() (string, error) {
	dir := filepath.Join(pkgDir, "guard-cache")
	if err := os.Mkdir(dir, 0o755); err != nil {
		return "", err
	}
	guardCacheDir = dir
	boot := filepath.Join(pkgDir, "scripts", "flow-guard")
	// The bootstrap builds under its key, then execs the binary;
	// spec-root's verdict is not read.
	cmd := exec.Command(boot, "spec-root")
	cmd.Env = append(append(os.Environ(), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+dir)
	out, _ := cmd.CombinedOutput()
	if bins, _ := filepath.Glob(dir + "/*/flow-guard"); len(bins) == 0 {
		return "", fmt.Errorf("building the shared flow-guard cache left no binary in %s:\n%s", dir, out)
	}
	return dir, nil
})

// TestMain builds the shared flow-guard cache before any test runs, so its
// build is not charged to whichever hook test calls runHook first (kan-1124,
// slowest-threshold).
func TestMain(m *testing.M) {
	code, err := runMain(m)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if pkgDir != "" {
		_ = os.RemoveAll(pkgDir)
	}
	os.Exit(code)
}

func runMain(m *testing.M) (int, error) {
	var err error
	if pkgDir, err = os.MkdirTemp("", "repotest-"); err != nil {
		return 2, err
	}
	// The bootstrap copy is what every fixture links; it resolves its checkout
	// from its own path, so it is a copy beside a link to the real stats/.
	boot, err := os.ReadFile("../../../scripts/flow-guard")
	if err != nil {
		return 2, err
	}
	stats, err := filepath.Abs("../..")
	if err != nil {
		return 2, err
	}
	if err := os.Mkdir(filepath.Join(pkgDir, "scripts"), 0o755); err != nil {
		return 2, err
	}
	for _, f := range []struct {
		path string
		body []byte
	}{
		{filepath.Join(pkgDir, "scripts", "flow-guard"), boot},
		{stubTrampoline(), []byte("#!/bin/sh\nexec /bin/sh \"$0.body\" \"$@\"\n")},
	} {
		if err := os.WriteFile(f.path, f.body, 0o755); err != nil {
			return 2, err
		}
	}
	if err := os.Symlink(stats, filepath.Join(pkgDir, "stats")); err != nil {
		return 2, err
	}
	if _, err := sharedGuardCache(); err != nil {
		return 2, err
	}
	// Run the trampoline once so its first-exec check is not charged to a test;
	// with no stub.body beside it, it exits non-zero.
	_ = exec.Command(stubTrampoline()).Run()
	return m.Run(), nil
}

// stubTrampoline is the one executable every writeStub stub links to.
func stubTrampoline() string { return filepath.Join(pkgDir, "stub") }

// writeStub makes path an executable running body under /bin/sh: body lands in
// path.body and path is a symlink to the package's one trampoline, which execs
// "$0.body" — a stub costs no first-exec check of a newly written file.
func writeStub(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path+".body", []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(stubTrampoline(), path); err != nil {
		t.Fatal(err)
	}
}

// runHook execs `python3 <repo>/hooks/<hook>` with payload on stdin and HOME set
// to a fresh temp dir; env entries are appended last, so they override. An
// absolute hook path is run as given (a hook reached through an install symlink).
// FLOW_GUARD_CACHE_DIR points every call at one flow-guard cache built once per
// package run: under load, each call building flow-guard into its own HOME
// could turn an allow into a deny (kan-1045 self-review, F-E).
func runHook(t *testing.T, hook string, payload []byte, env ...string) (stdout, stderr string, rc int) {
	t.Helper()
	if !filepath.IsAbs(hook) {
		hook = filepath.Join(repoRoot(t), "hooks", hook)
	}
	cache, err := sharedGuardCache()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("python3", hook)
	cmd.Stdin = bytes.NewReader(payload)
	// FLOW_HOOK_INCIDENT_CMD=true keeps a deny's incident write off the dev store; a caller's
	// env, appended after it, may point it at a stub. FLOW_HOOK_UNBOUNDED=1 lifts the hooks'
	// wall-clock bounds: under the full -race suite a starved stub timed out silently (kan-1060).
	// FLOW_ADDR and FLOW_RECORDS_ADDR point at a dead port so no hook reads the live dev store: an
	// inherited FLOW_RECORDS_ADDR let an open experiment's overrides flip a policy case (kan-1060).
	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache, "FLOW_HOOK_INCIDENT_CMD=true", "FLOW_HOOK_UNBOUNDED=1", "FLOW_ADDR=http://127.0.0.1:1", "FLOW_RECORDS_ADDR=http://127.0.0.1:1"), env...)
	cmd.Env = append(cmd.Env, "PATH="+policyFlowDir(t)+string(os.PathListSeparator)+lastEnv(cmd.Env, "PATH"))
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err = cmd.Run()
	var ee *exec.ExitError
	switch {
	case errors.As(err, &ee):
		rc = ee.ExitCode()
	case err != nil:
		t.Fatalf("run %s: %v", hook, err)
	}
	return out.String(), errb.String(), rc
}

// sh runs a fixture command, failing the test on error.
func sh(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestRunHookSharesOneGuardCache pins the kan-1045 self-review fix, F-E:
// every runHook call sees one FLOW_GUARD_CACHE_DIR, already holding a
// built flow-guard.
func TestRunHookSharesOneGuardCache(t *testing.T) {
	script := filepath.Join(t.TempDir(), "env.py")
	if err := os.WriteFile(script, []byte("import os\nprint(os.environ.get(\"FLOW_GUARD_CACHE_DIR\", \"\"))\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	first, _, _ := runHook(t, script, nil)
	second, _, _ := runHook(t, script, nil)
	dir := strings.TrimSpace(first)
	if dir == "" || first != second {
		t.Fatalf("FLOW_GUARD_CACHE_DIR = %q then %q; want one shared, non-empty directory", first, second)
	}
	if bins, _ := filepath.Glob(filepath.Join(dir, "*", "flow-guard")); len(bins) == 0 {
		t.Fatalf("no flow-guard built under %s", dir)
	}
}

// incidentStub is an argv-logging FLOW_HOOK_INCIDENT_CMD exiting rc: each call's argv lands in
// the returned log, one argument per line, calls separated by a "--" line.
func incidentStub(t *testing.T, rc int) (cmd, log string) {
	t.Helper()
	dir := t.TempDir()
	cmd, log = filepath.Join(dir, "flow"), filepath.Join(dir, "argv")
	writeStub(t, cmd, fmt.Sprintf("printf '%%s\\n' \"$@\" >>%q\necho -- >>%q\necho noise\nexit %d\n", log, log, rc))
	return cmd, log
}

// incidentCalls reads incidentStub's log as one argv per call.
func incidentCalls(t *testing.T, log string) [][]string {
	t.Helper()
	b, err := os.ReadFile(log)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var calls [][]string
	for _, c := range strings.Split(strings.TrimSuffix(string(b), "--\n"), "--\n") {
		calls = append(calls, strings.Split(strings.TrimSuffix(c, "\n"), "\n"))
	}
	return calls
}

// argAfter is the argument following flag in argv, "" when flag is absent.
func argAfter(argv []string, flag string) string {
	for i := 0; i+1 < len(argv); i++ {
		if argv[i] == flag {
			return argv[i+1]
		}
	}
	return ""
}

// lastEnv is the value env's last k= entry sets, the one exec uses.
func lastEnv(env []string, k string) string {
	v := ""
	for _, kv := range env {
		if rest, ok := strings.CutPrefix(kv, k+"="); ok {
			v = rest
		}
	}
	return v
}

// policyFlowDir holds a `flow` that answers `policy get` with the seeded
// pairs (internal/guard/testdata/policy.json) and hands every other call to
// the next `flow` on PATH -- a test's own fake, or none: check-loop-dispatch
// reads the pairs through `flow policy get` (kan-1103), and the dev store's
// pairs, or a `flow` built before them, must never decide a hook case.
func policyFlowDir(t *testing.T) string {
	t.Helper()
	fixture, err := filepath.Abs("../guard/testdata/policy.json")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	writeStub(t, filepath.Join(dir, "flow"), "[ \"$1 $2\" = \"policy get\" ] && exec cat '"+fixture+"'\nPATH=${PATH#*:} exec flow \"$@\"\n")
	return dir
}
