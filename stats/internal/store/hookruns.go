package store

import (
	"context"
	"fmt"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
)

// RecordHookRuns inserts runs into hook_runs (0114_hook_runs.sql) in one
// statement; a run with no hook name or a negative duration is refused with
// the whole batch.
func (s *Store) RecordHookRuns(ctx context.Context, runs []records.HookRun) error {
	if len(runs) == 0 {
		return nil
	}
	hooks, us, at := make([]string, len(runs)), make([]int64, len(runs)), make([]time.Time, len(runs))
	for i, r := range runs {
		hooks[i], us[i], at[i] = r.Hook, r.DurationUS, r.At
	}
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO hook_runs (hook, duration_us, ran_at)
		SELECT * FROM unnest($1::text[], $2::bigint[], $3::timestamptz[])
	`, hooks, us, at); err != nil {
		return fmt.Errorf("store: record hook runs: %w", err)
	}
	return nil
}
