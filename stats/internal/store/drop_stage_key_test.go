package store_test

import (
	"context"
	"testing"
)

// TestStageRunsCarriesNoStageKey pins migration 0048: stage_runs.stage_key and
// its index, added by 0008 and never written or read, are gone (kan-1037).
func TestStageRunsCarriesNoStageKey(t *testing.T) {
	_, pool := newRecordStore(t)
	ctx := context.Background()
	var n int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'stage_runs' AND column_name = 'stage_key'
	`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Error("stage_runs still has a stage_key column")
	}
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM pg_indexes
		WHERE schemaname = current_schema() AND indexname = 'stage_runs_stage_key'
	`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Error("index stage_runs_stage_key still exists")
	}
}
