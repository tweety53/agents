package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

func detachedEnv(dir string) core.Env {
	return core.Env{Getenv: os.Getenv, Dir: dir}
}

// waitForFile polls for path; the test's own wait, not the guard's.
func waitForFile(t *testing.T, path string) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s never appeared", path)
}

// TestRunDetachedRecordsExit: STARTED <pid> at once, the child's output in out.log and its exit
// code written to exit when it ends.
func TestRunDetachedRecordsExit(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "st")
	var o, e bytes.Buffer
	if rc := runDetached([]string{dir, "--", "sh", "-c", "echo hello; exit 3"}, detachedEnv(t.TempDir()), &o, &e); rc != 0 {
		t.Fatalf("rc %d: %s", rc, e.String())
	}
	pid, _ := os.ReadFile(filepath.Join(dir, "pid"))
	if o.String() != "STARTED "+strings.TrimSpace(string(pid))+"\n" {
		t.Errorf("stdout %q pid %q", o.String(), pid)
	}
	waitForFile(t, filepath.Join(dir, "exit"))
	if b, _ := os.ReadFile(filepath.Join(dir, "exit")); strings.TrimSpace(string(b)) != "3" {
		t.Errorf("exit %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(dir, "out.log")); string(b) != "hello\n" {
		t.Errorf("out.log %q", b)
	}
}

// TestRunDetachedRefusesWhileRunning: a second call on a live state-dir exits 1; the child runs in
// its own session, so it outlives the starter's return.
func TestRunDetachedRefusesWhileRunning(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "st")
	marker := filepath.Join(t.TempDir(), "done")
	var o, e bytes.Buffer
	if rc := runDetached([]string{dir, "--", "sh", "-c", "sleep 1; touch " + marker}, detachedEnv(t.TempDir()), &o, &e); rc != 0 {
		t.Fatalf("rc %d", rc)
	}
	o.Reset()
	if rc := runDetached([]string{dir, "--", "true"}, detachedEnv(t.TempDir()), &o, &e); rc != 1 || !strings.Contains(e.String(), "running") {
		t.Errorf("second call rc %d err %q", rc, e.String())
	}
	waitForFile(t, marker)
	waitForFile(t, filepath.Join(dir, "exit"))
}

// TestAwaitDetachedExited: an exit file returns EXITED <code> and the log tail.
func TestAwaitDetachedExited(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var log strings.Builder
	for i := 1; i <= 50; i++ {
		log.WriteString("line" + strconv.Itoa(i) + "\n")
	}
	writeFile(t, filepath.Join(dir, "out.log"), log.String())
	writeFile(t, filepath.Join(dir, "pid"), "1\n")
	writeFile(t, filepath.Join(dir, "exit"), "2\n")
	var o, e bytes.Buffer
	if rc := awaitDetached([]string{dir, "-timeout", "1"}, detachedEnv(dir), &o, &e); rc != 0 {
		t.Fatalf("rc %d", rc)
	}
	got := o.String()
	if !strings.HasPrefix(got, "EXITED 2\nline11\n") || !strings.HasSuffix(got, "line50\n") {
		t.Errorf("stdout %q", got)
	}
}

// TestAwaitDetachedRunning: a live pid with no exit file prints RUNNING and the last 10 lines at
// the timeout, exit 0.
func TestAwaitDetachedRunning(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "out.log"), "a\nb\n")
	writeFile(t, filepath.Join(dir, "pid"), strconv.Itoa(os.Getpid())+"\n")
	var o, e bytes.Buffer
	start := time.Now()
	if rc := awaitDetached([]string{dir, "-timeout", "1"}, detachedEnv(dir), &o, &e); rc != 0 {
		t.Fatalf("rc %d", rc)
	}
	if !strings.HasPrefix(o.String(), "RUNNING ") || !strings.HasSuffix(o.String(), "s\na\nb\n") {
		t.Errorf("stdout %q", o.String())
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("waited past its timeout: %s", time.Since(start))
	}
}

// TestAwaitDetachedLost: a dead pid with no exit file is LOST, exit 1.
func TestAwaitDetachedLost(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "out.log"), "")
	writeFile(t, filepath.Join(dir, "pid"), "999999\n")
	var o, e bytes.Buffer
	if rc := awaitDetached([]string{dir, "-timeout", "1"}, detachedEnv(dir), &o, &e); rc != 1 || o.String() != "LOST\n" {
		t.Errorf("rc %d stdout %q", rc, o.String())
	}
}

// TestAwaitDetachedRecordsItsSleepAsWait: the polling sleep is the run's wait, not work.
func TestAwaitDetachedRecordsItsSleepAsWait(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "out.log"), "")
	writeFile(t, filepath.Join(dir, "pid"), strconv.Itoa(os.Getpid())+"\n")
	env := detachedEnv(dir)
	env.Reasons = &core.Reasons{}
	var o, e bytes.Buffer
	if rc := awaitDetached([]string{dir, "-timeout", "1"}, env, &o, &e); rc != 0 {
		t.Fatalf("rc %d", rc)
	}
	if w := env.Reasons.Wait(); w < 900*time.Millisecond {
		t.Errorf("wait %s, want about the 1s slept", w)
	}
}
