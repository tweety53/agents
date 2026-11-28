package guardtest

import (
	"bytes"
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// GitC runs git -C dir under FixtureGitEnv, returning
// stdout without its trailing newline.
func GitC(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(FixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), FixtureGitEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return strings.TrimRight(string(out), "\n")
}

// RowsEnv serves the rows from a fake flowd: rows, or rowsEnded once a `flow record dispatch
// end` -- the sweep's write, still the flow CLI on PATH -- has been called. It returns the env,
// HOME and the change's worktree.
func RowsEnv(t *testing.T, rows, rowsEnded string) (core.Env, string, string) {
	t.Helper()
	tmp := t.TempDir()
	swept := filepath.Join(tmp, "swept")
	srv := FakeFlowdRows(t, func() string {
		if _, err := os.Stat(swept); err == nil {
			return rowsEnded
		}
		return rows
	}, nil)
	WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\ntouch "+swept+"\n")
	home := filepath.Join(tmp, "home")
	return core.Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return home
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}, home, GitWorktree(t, tmp, "chg")
}
