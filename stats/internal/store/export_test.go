package store

import "time"

// FanOutMaxForTest exposes fanOutMax to package store_test, whose black-box
// tests otherwise have no way to reach an unexported function directly.
// dispatchWindow is the test package's own minimal shape for a dispatch
// interval, converted here to the runDispatchRowRaw fanOutMax actually
// sweeps.
type DispatchWindowForTest struct {
	StartedAt time.Time
	EndedAt   *time.Time
}

func FanOutMaxForTest(windows []DispatchWindowForTest) int {
	raw := make([]runDispatchRowRaw, len(windows))
	for i, w := range windows {
		raw[i] = runDispatchRowRaw{StartedAt: w.StartedAt, EndedAt: w.EndedAt}
	}
	return fanOutMax(raw)
}
