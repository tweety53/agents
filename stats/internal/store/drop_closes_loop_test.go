package store_test

import (
	"context"
	"os"
	"testing"
)

// TestNoDispatchClosesLoopColumn pins migration 0107: dispatches.closes_loop
// is gone, and the migration re-runs on a store that already lacks it.
func TestNoDispatchClosesLoopColumn(t *testing.T) {
	t.Parallel()
	_, pool := newRecordStore(t)
	ctx := context.Background()
	sql, err := os.ReadFile("migrations/0107_drop_dispatch_closes_loop.sql")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("0107 is not idempotent: %v", err)
	}
	var n int
	if err := pool.QueryRow(ctx, `
		SELECT count(*) FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'dispatches' AND column_name = 'closes_loop'
	`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("dispatches.closes_loop remains")
	}
}
