package guard

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// TestCaptureWorkers pins the capture suite's worker count: half the cores,
// never below 1, never above the cap.
func TestCaptureWorkers(t *testing.T) {
	t.Parallel()
	for ncpu, want := range map[int]int{1: 1, 2: 1, 4: 2, 10: 5, 12: 6, 64: 6} {
		if got := captureWorkers(ncpu); got != want {
			t.Errorf("captureWorkers(%d) = %d, want %d", ncpu, got, want)
		}
	}
	env := Env{Getenv: os.Getenv}
	r := runGuard("capture-workers", nil, env)
	if n, err := strconv.Atoi(strings.TrimSpace(r.stdout)); r.rc != 0 || err != nil || n < 1 || n > captureWorkersCap {
		t.Errorf("rc=%d stdout=%q", r.rc, r.stdout)
	}
	if r = runGuard("capture-workers", []string{"x"}, env); r.rc != 2 {
		t.Errorf("an argument: rc=%d, want 2", r.rc)
	}
}

// vvpWorkersFlag is the worker flag prepare's shard capture carries on this
// machine, with its trailing space.
var vvpWorkersFlag = "--workers=" + strconv.Itoa(captureWorkers(runtime.NumCPU())) + " "
