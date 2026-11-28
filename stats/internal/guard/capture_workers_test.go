package guard

import "testing"

// TestCaptureWorkers pins the capture suite's worker count: half the cores,
// never below 1, never above the cap.
func TestCaptureWorkers(t *testing.T) {
	t.Parallel()
	for ncpu, want := range map[int]int{1: 1, 2: 1, 4: 2, 10: 5, 12: 6, 64: 6} {
		if got := captureWorkers(ncpu); got != want {
			t.Errorf("captureWorkers(%d) = %d, want %d", ncpu, got, want)
		}
	}
}
