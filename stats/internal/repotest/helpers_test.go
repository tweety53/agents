package repotest

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// runHook execs `python3 <repo>/hooks/<hook>` with payload on stdin and HOME set
// to a fresh temp dir; env entries are appended last, so they override.
func runHook(t *testing.T, hook string, payload []byte, env ...string) (stdout, stderr string, rc int) {
	t.Helper()
	cmd := exec.Command("python3", filepath.Join(repoRoot(t), "hooks", hook))
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Env = append(append(append(os.Environ(), "HOME="+t.TempDir()), goCacheEnv()...), env...)
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
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
