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

// sharedGuardCache makes one temp directory and builds flow-guard into it
// once per package run, as internal/guard's guardCacheBuilt does.
var sharedGuardCache = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "repotest-flow-guard-")
	if err != nil {
		return "", err
	}
	guardCacheDir = dir
	lib, err := filepath.Abs("../../../scripts/lib/flow-guard.sh")
	if err != nil {
		return "", err
	}
	// flow_guard_exec builds under the shim's own key, then execs the
	// binary; spec-root's verdict is not read.
	cmd := exec.Command("bash", "-c", `. "$1" && flow_guard_exec spec-root 2 guard-cache:`, "_", lib)
	cmd.Env = append(append(os.Environ(), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+dir)
	out, _ := cmd.CombinedOutput()
	if bins, _ := filepath.Glob(dir + "/*/flow-guard"); len(bins) == 0 {
		return "", fmt.Errorf("building the shared flow-guard cache left no binary in %s:\n%s", dir, out)
	}
	return dir, nil
})

func TestMain(m *testing.M) {
	code := m.Run()
	if guardCacheDir != "" {
		_ = os.RemoveAll(guardCacheDir)
	}
	os.Exit(code)
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
	cmd.Env = append(append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), "FLOW_GUARD_CACHE_DIR="+cache), env...)
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
