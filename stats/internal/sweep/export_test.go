package sweep

import "time"

// SetClockForTest overrides s's internal clock. It exists so
// sweep_test.go (package sweep_test, exercising this package the way a
// real caller would) can pin RunOnce's cutoff computation to a fixed
// instant instead of racing the real wall clock -- the same reason
// store's own SweepAbandoned tests compute their cutoff once, up front,
// rather than re-deriving "now" inside an assertion.
func SetClockForTest(s *Sweeper, now func() time.Time) {
	s.now = now
}
