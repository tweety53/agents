// Package sweep periodically closes stage runs abandoned by a session that
// has gone silent -- design.md's "Stage marks": "A run that dies leaves a
// stage open. The daemon sweeps stages whose session has been silent past
// a timeout and sets outcome = 'abandoned'."
//
// An abandoned stage is a statistic worth having, not an error to
// suppress: the outcome column records it exactly like any other outcome.
//
// This package owns none of the closing logic itself -- that lives in
// internal/store's SweepAbandoned (task 3), a single atomic UPDATE that is
// already safe under concurrency (see its own doc comment) and already
// idempotent (re-running it against a row it already closed matches no
// rows the second time, since its WHERE clause requires ended_at IS
// NULL). What this package adds is the daemon-facing half design.md's
// "Stage marks" describes but SweepAbandoned's signature does not: turning
// a configured silence timeout into the cutoff instant SweepAbandoned
// takes, and running that on a schedule for as long as the daemon lives --
// the same shape internal/harvest.Watcher and internal/reconcile.Reconciler
// already give their own periodic work.
package sweep

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// AbandonedSweeper is the store dependency this package needs -- defined
// here, at the consumer, per go-interface-design, so Sweeper is testable
// against a fake with no database at all. store.Store.SweepAbandoned
// (task 3) already matches this signature exactly; cmd/flowd/main.go
// wires the real store in with no adapter, exactly as it already does for
// harvest.WindowSource and harvest.HarvestSink.
type AbandonedSweeper interface {
	// SweepAbandoned closes every stage run that is still open (no end
	// mark) and started before silentBefore, setting its outcome to
	// "abandoned" and its end instant to when the sweep ran. It returns
	// how many rows it closed.
	SweepAbandoned(ctx context.Context, silentBefore time.Time) (int64, error)
}

// Sweeper periodically closes every stage run whose session has been
// silent for longer than its configured timeout, with no end mark to
// explain why.
type Sweeper struct {
	store   AbandonedSweeper
	timeout time.Duration
	logger  *slog.Logger
	// now is time.Now by default; a test may override it to make the
	// cutoff RunOnce computes deterministic instead of racing the clock.
	now func() time.Time
}

// New builds a Sweeper over store, closing any stage run that has sat open
// for longer than timeout with no end mark. timeout must be chosen well
// above this pipeline's longest ordinary stage duration -- the sweeper's
// job is to catch a session that will never send an end mark, not to
// second-guess one that is still legitimately working. logger may be nil.
func New(store AbandonedSweeper, timeout time.Duration, logger *slog.Logger) *Sweeper {
	return &Sweeper{store: store, timeout: timeout, logger: logger, now: time.Now}
}

// RunOnce performs a single sweep pass: every stage run started more than
// the sweeper's timeout before now, and still lacking an end mark, is
// closed with outcome "abandoned". It returns how many rows were closed.
//
// RunOnce is safe to call concurrently with itself (from a second
// Sweeper, in a second daemon process, sharing the same store) and safe
// to call concurrently with a live `stage end` racing to close the same
// row: SweepAbandoned's own doc comment covers why -- a plain
// `UPDATE ... WHERE ended_at IS NULL AND started_at < $1` takes Postgres's
// ordinary per-row locking and read-committed re-check semantics, with no
// separate read-then-write for either side to race against.
func (s *Sweeper) RunOnce(ctx context.Context) (int64, error) {
	cutoff := s.now().Add(-s.timeout)
	n, err := s.store.SweepAbandoned(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("sweep: %w", err)
	}
	return n, nil
}

// Run calls RunOnce every interval until ctx is done. A single failed pass
// is logged and does not stop the loop -- the daemon must keep sweeping
// future abandoned stages even if one pass hit a transient error, exactly
// as internal/harvest.Watcher.Run and internal/reconcile.Reconciler.Watch
// already do for their own periodic work.
func (s *Sweeper) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := s.RunOnce(ctx)
			if err != nil {
				s.warn("sweep: run failed", "error", err)
				continue
			}
			if n > 0 {
				s.info("sweep: closed abandoned stage runs", "count", n)
			}
		}
	}
}

func (s *Sweeper) warn(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Warn(msg, args...)
	}
}

func (s *Sweeper) info(msg string, args ...any) {
	if s.logger != nil {
		s.logger.Info(msg, args...)
	}
}
