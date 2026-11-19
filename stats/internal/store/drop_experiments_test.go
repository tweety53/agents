package store_test

import (
	"context"
	"testing"
)

// TestNoExperimentTables pins migration 0081: the allowed-pairs experiments
// and their per-role holds are gone; role pairs are edited directly.
func TestNoExperimentTables(t *testing.T) {
	t.Parallel()
	_, pool := newRecordStore(t)
	var n int
	if err := pool.QueryRow(context.Background(), `
		SELECT count(*) FROM information_schema.tables
		WHERE table_schema = current_schema() AND table_name IN ('experiments', 'experiment_roles')
	`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d experiment tables remain, want none", n)
	}
}
