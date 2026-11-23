package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// A process whose cwd is inside the worktree is stopped; one outside it, and this test's own process, are not.
func TestStopWorktreeProcesses(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof is not on PATH")
	}
	root := t.TempDir()
	wt, other := filepath.Join(root, "wt"), filepath.Join(root, "other")
	for _, d := range []string{filepath.Join(wt, "src", "app"), other} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	start := func(dir string) *exec.Cmd {
		c := exec.Command("sleep", "300")
		c.Dir = dir
		if err := c.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Process.Kill(); _ = c.Wait() })
		return c
	}
	inside, outside := start(filepath.Join(wt, "src", "app")), start(other)
	reaped := make(chan struct{})
	go func() { _ = inside.Wait(); close(reaped) }()
	time.Sleep(200 * time.Millisecond)
	var out, errb bytes.Buffer
	env := Env{Getenv: os.Getenv}
	if code := stopWorktreeProcesses([]string{wt}, env, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	select {
	case <-reaped:
	case <-time.After(10 * time.Second):
		t.Fatalf("the process inside the worktree is still alive\n%s", out.String())
	}
	if syscall.Kill(outside.Process.Pid, 0) != nil {
		t.Error("the process outside the worktree was stopped")
	}
	if !strings.Contains(out.String(), "STOPPED: ") || !strings.Contains(out.String(), "STOP-WORKTREE-PROCESSES-OK: 1 stopped") {
		t.Errorf("output %q", out.String())
	}
	if code := stopWorktreeProcesses(nil, env, &out, &errb); code != 2 {
		t.Errorf("no argument: exit %d, want 2", code)
	}
}
