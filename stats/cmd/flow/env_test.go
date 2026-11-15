package main

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// TestRunReadsTheEnvironmentItIsGiven pins the per-call environment
// (decisions cli-per-call-env, getenv-explicit-param): run reads FLOW_*
// through the getenv it is handed, never the process environment, so a
// parallel test's state root and store address are its own. The process
// variables are decoys a leaked os.Getenv would read.
func TestRunReadsTheEnvironmentItIsGiven(t *testing.T) {
	repo := gitRepo(t)
	t.Setenv("FLOW_STATE_DIR", t.TempDir())
	t.Setenv("FLOW_RECORDS_ADDR", "http://127.0.0.1:1")
	root := t.TempDir()
	dead := deadPortURL(t)
	env := map[string]string{"FLOW_STATE_DIR": root, "FLOW_RECORDS_ADDR": dead}
	getenv := func(k string) string { return env[k] }

	key, _, err := fallback.ProjectKey(repo)
	if err != nil {
		t.Fatalf("ProjectKey: %v", err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(context.Background(), []string{"state", "dir", "-C", repo}, getenv, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("state dir exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if want := filepath.Join(root, key) + "\n"; stdout.String() != want {
		t.Errorf("state dir = %q, want %q", stdout.String(), want)
	}

	stdout.Reset()
	stderr.Reset()
	if code := run(context.Background(), []string{"suite", "list", "-C", repo, "-timeout", "500ms"}, getenv, strings.NewReader(""), &stdout, &stderr); code != 1 {
		t.Errorf("suite list against the given dead store exit = %d, want 1; stderr:\n%s", code, stderr.String())
	}
	if want := "flow: using FLOW_RECORDS_ADDR=" + dead; !strings.Contains(stderr.String(), want) {
		t.Errorf("stderr = %q, want %q", stderr.String(), want)
	}
}
