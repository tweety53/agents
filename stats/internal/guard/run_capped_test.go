package guard

import (
	"bufio"
	"math/rand/v2"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// rcHelperEnv names this test binary's helper mode: re-exec'd with it set,
// the binary allocates instead of running tests.
const rcHelperEnv = "RUN_CAPPED_TEST_HELPER"

func init() {
	switch os.Getenv(rcHelperEnv) {
	case "":
		return
	case "alloc":
		if os.Getenv("RUN_CAPPED_TEST_IGNORE_TERM") != "" {
			signal.Ignore(syscall.SIGTERM)
		}
		// 128 MB of random bytes, re-touched every 0.2 s for 10 s, so no page
		// of it is idle long enough to be compressed; a marker written only
		// once the allocator finishes proves the kill without a wall-clock
		// bound that load breaks (kan-1052 self-review, F4).
		b := make([]byte, 128<<20)
		for i := range b {
			b[i] = byte(rand.Uint32())
		}
		for range 50 {
			for i := 0; i < len(b); i += 4096 {
				b[i] = byte(rand.Uint32())
			}
			time.Sleep(200 * time.Millisecond)
		}
		runtime.KeepAlive(b)
		if done := os.Getenv("RUN_CAPPED_TEST_DONE"); done != "" {
			_ = os.WriteFile(done, nil, 0o644)
		}
	case "setsid":
		// A child in its own session and group, as a Playwright browser.
		self, _ := os.Executable()
		c := exec.Command(self)
		c.Env = append(os.Environ(), rcHelperEnv+"=alloc")
		c.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if c.Start() != nil {
			os.Exit(1)
		}
		_ = os.WriteFile(os.Getenv("RUN_CAPPED_TEST_PIDFILE"), []byte(strconv.Itoa(c.Process.Pid)), 0o644)
		_ = c.Wait()
	case "idle":
		// 128 MB of zero pages, each dirtied, then left idle: macOS compresses
		// them out of the resident set while the footprint keeps them.
		b := make([]byte, 128<<20)
		for i := 0; i < len(b); i += 4096 {
			b[i] = 1
		}
		for i := 0; i < len(b); i += 4096 {
			b[i] = 0
		}
		os.Stdout.WriteString("ready\n")
		for {
			time.Sleep(time.Hour)
			runtime.KeepAlive(b)
		}
	}
	os.Exit(0)
}

// rcRun runs `flow-guard run-capped args…` from dir with extra environment,
// stdin when given; it returns the exit code, stdout and stderr.
func rcRun(t *testing.T, dir, stdin string, extra []string, args ...string) (int, string, string) {
	t.Helper()
	cmd := exec.Command(guardBinary(t), append([]string{"run-capped"}, args...)...)
	cmd.Dir = dir
	cmd.Env = append(rcEnv(), extra...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, errb strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errb
	return exitCode(cmd.Run()), out.String(), errb.String()
}

// rcEnv is the test's environment outside any run-capped.
func rcEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "RUN_CAPPED_PGID=") {
			env = append(env, kv)
		}
	}
	return env
}

func rcSelf(t *testing.T) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return self
}

// rcAlive is whether rcProcs lists pid as a live process: a zombie is dead
// for this purpose, since its exit is done and only the reap is left
// (kan-1103 self-review, F1).
func rcAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	for _, p := range rcProcs() {
		if p.pid == pid {
			return !p.zombie
		}
	}
	return false
}

func rcReadPid(t *testing.T, path string) int {
	t.Helper()
	b, _ := os.ReadFile(path)
	pid, _ := strconv.Atoi(strings.TrimSpace(string(b)))
	return pid
}

// TestRunCappedContract ports scripts/test-run-capped.sh's cases (a)–(j):
// run-capped kills a command's process tree past its memory cap. Kept
// light: cap 64 MB, the allocator 128 MB.
func TestRunCappedContract(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	self := rcSelf(t)
	alloc := rcHelperEnv + "=alloc"

	t.Run("(a) capped before completion, with a CAPPED: line", func(t *testing.T) {
		t.Parallel()
		done := t.TempDir() + "/done-a"
		rc, _, errOut := rcRun(t, t.TempDir(), "", []string{alloc, "RUN_CAPPED_TEST_DONE=" + done}, "-m", "64", "--", self)
		if rc != 137 || rmcExists(done) || !strings.Contains(errOut, "CAPPED: 64MB (tree footprint ") {
			t.Fatalf("exit %d, completed %v, stderr %q", rc, rmcExists(done), errOut)
		}
	})
	t.Run("(b) a grandchild allocator is killed too", func(t *testing.T) {
		t.Parallel()
		pidfile := t.TempDir() + "/pid"
		rc, _, _ := rcRun(t, t.TempDir(), "", []string{alloc}, "-m", "64", "--", "bash", "-c", `"$0" & echo $! >"$1"; wait`, self, pidfile)
		if gp := rcReadPid(t, pidfile); rc != 137 || gp == 0 || rcAlive(gp) {
			t.Fatalf("exit %d, grandchild %d alive %v", rc, gp, rcAlive(gp))
		}
	})
	t.Run("(c) the command's status passes through", func(t *testing.T) {
		t.Parallel()
		if rc, _, _ := rcRun(t, t.TempDir(), "", nil, "-m", "64", "--", "bash", "-c", "exit 3"); rc != 3 {
			t.Fatalf("exit %d, want 3", rc)
		}
	})
	project := func(t *testing.T, body string) string {
		p := t.TempDir()
		writeFile(t, p+"/.flow/project.md", body)
		return p
	}
	t.Run("(d) the cap comes from .flow/project.md's ## resource cap", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := rcRun(t, project(t, "## resource cap\n\n64\n"), "", []string{alloc}, "--", self)
		if rc != 137 || !strings.Contains(errOut, "CAPPED: 64MB") {
			t.Fatalf("exit %d, stderr %q", rc, errOut)
		}
	})
	t.Run("(e) a setsid child is in the tree", func(t *testing.T) {
		t.Parallel()
		pidfile := t.TempDir() + "/sid"
		rc, _, _ := rcRun(t, t.TempDir(), "", []string{rcHelperEnv + "=setsid", "RUN_CAPPED_TEST_PIDFILE=" + pidfile}, "-m", "64", "--", self)
		sp := rcReadPid(t, pidfile)
		defer func() {
			if sp > 0 {
				_ = syscall.Kill(sp, syscall.SIGKILL)
			}
		}()
		if rc != 137 || sp == 0 || rcAlive(sp) {
			t.Fatalf("exit %d, setsid child %d alive %v", rc, sp, rcAlive(sp))
		}
	})
	t.Run("(f) a nested run-capped: the outer cap covers the inner command", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := rcRun(t, t.TempDir(), "", []string{alloc}, "-m", "64", "--", guardBinary(t), "run-capped", "-m", "4096", "--", self)
		if rc != 137 || !strings.Contains(errOut, "CAPPED: 64MB") {
			t.Fatalf("exit %d, stderr %q", rc, errOut)
		}
	})
	t.Run("(g) a cap with a unit is refused, not misread", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := rcRun(t, project(t, "## resource cap\n\n8 GB\n"), "", nil, "--", "true")
		if rc != 2 || !strings.Contains(errOut, "plain integer") {
			t.Fatalf("exit %d, stderr %q", rc, errOut)
		}
	})
	t.Run("(h) project-get refusing (the key declared twice) fails loudly", func(t *testing.T) {
		t.Parallel()
		rc, _, errOut := rcRun(t, project(t, "## resource cap\n\n64\n\n## resource cap\n\n64\n"), "", nil, "--", "true")
		if rc != 2 || !strings.Contains(errOut, "project-get failed") {
			t.Fatalf("exit %d, stderr %q", rc, errOut)
		}
	})
	t.Run("(i) SIGTERM to the wrapper exits 143 and kills the command", func(t *testing.T) {
		t.Parallel()
		tp := t.TempDir() + "/tp"
		cmd := exec.Command(guardBinary(t), "run-capped", "-m", "64", "--", "bash", "-c", `echo $$ >"$0"; sleep 20`, tp)
		cmd.Env = rcEnv()
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(30 * time.Second)
		for rcReadPid(t, tp) == 0 && time.Now().Before(deadline) {
			time.Sleep(50 * time.Millisecond)
		}
		_ = cmd.Process.Signal(syscall.SIGTERM)
		rc := exitCode(cmd.Wait())
		if child := rcReadPid(t, tp); rc != 143 || child == 0 || rcAlive(child) {
			t.Fatalf("exit %d, child %d alive %v", rc, child, rcAlive(child))
		}
	})
	t.Run("(j) the command's stdin is /dev/null", func(t *testing.T) {
		t.Parallel()
		if _, out, _ := rcRun(t, t.TempDir(), "hi\n", nil, "-m", "64", "--", "bash", "-c", `read -r x && echo "got $x"`); out != "" {
			t.Fatalf("the child read stdin: %q", out)
		}
	})
}

// TestRunCappedKillsASessionLeaderThatIgnoresTerm: a setsid child that
// ignores SIGTERM is reparented once its parent dies of the TERM, and still
// gets the KILL — run-capped remembers every pid its tree held (kan-1103
// self-review, F1).
func TestRunCappedKillsASessionLeaderThatIgnoresTerm(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	pidfile := t.TempDir() + "/sid"
	rc, _, _ := rcRun(t, t.TempDir(), "", []string{rcHelperEnv + "=setsid", "RUN_CAPPED_TEST_PIDFILE=" + pidfile, "RUN_CAPPED_TEST_IGNORE_TERM=1"}, "-m", "64", "--", rcSelf(t))
	sp := rcReadPid(t, pidfile)
	defer func() {
		if sp > 0 {
			_ = syscall.Kill(sp, syscall.SIGKILL)
		}
	}()
	if rc != 137 || sp == 0 || rcAlive(sp) {
		t.Fatalf("exit %d, TERM-ignoring setsid child %d alive %v", rc, sp, rcAlive(sp))
	}
}

// TestRunCappedFootprintSurvivesCompression is the proof the cap reaches a
// compressed process: an allocator that dirtied 128 MB of zero pages and
// idles falls out of its own RSS as macOS compresses those pages, and the
// tree sampler still reads all of it.
func TestRunCappedFootprintSurvivesCompression(t *testing.T) {
	smoke.Heavy(t)
	if runtime.GOOS != "darwin" {
		t.Skip("memory compression of idle pages is the darwin behaviour this pins")
	}
	t.Parallel()
	cmd := exec.Command(rcSelf(t))
	cmd.Env = append(rcEnv(), rcHelperEnv+"=idle")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cmd.Process.Kill(); _ = cmd.Wait() }()
	if line, _ := bufio.NewReader(out).ReadString('\n'); line != "ready\n" {
		t.Fatalf("the allocator did not get ready: %q", line)
	}
	pid := cmd.Process.Pid
	rss := func() int {
		b, _ := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid)).Output()
		kb, _ := strconv.Atoi(strings.TrimSpace(string(b)))
		return kb
	}
	peak, now := rss(), 0
	const fall = 32 << 10 // KB
	for deadline := time.Now().Add(120 * time.Second); ; time.Sleep(time.Second) {
		now = rss()
		peak = max(peak, now)
		if now <= peak-fall {
			break
		}
		if time.Now().After(deadline) {
			t.Skipf("precondition unmet: RSS did not fall 32 MB under its %d KB peak within 120 s (now %d KB)", peak, now)
		}
	}
	if got := rcTreeFootprint(pid); got < 128<<20 {
		t.Fatalf("the tree sampler reads %d MB of a 128 MB allocator whose RSS fell from %d to %d MB", got>>20, peak>>10, now>>10)
	}
}
