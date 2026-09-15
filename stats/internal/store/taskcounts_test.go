package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// TestRecordTaskCountReturnsRecordedRow covers the write path: the
// returned row carries what was recorded plus the store's own stamps (id,
// observed_at), and a second insert is a second row -- no upsert, one row
// per observation, the property the planned/appended derivation reads the
// series by.
func TestRecordTaskCountReturnsRecordedRow(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-task-counts-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-415")

	first, err := st.RecordTaskCount(ctx, projectKey, "kan-415", records.TaskCount{TotalTasks: 22})
	if err != nil {
		t.Fatalf("RecordTaskCount: %v", err)
	}
	if first.ID == 0 {
		t.Errorf("first.ID = 0, want the recorded row's id")
	}
	if first.TotalTasks != 22 {
		t.Errorf("first.TotalTasks = %d, want 22", first.TotalTasks)
	}
	if first.ObservedAt.IsZero() {
		t.Errorf("first.ObservedAt is zero, want the store's stamp")
	}

	second, err := st.RecordTaskCount(ctx, projectKey, "kan-415", records.TaskCount{TotalTasks: 46})
	if err != nil {
		t.Fatalf("second RecordTaskCount: %v", err)
	}
	if second.ID == first.ID {
		t.Errorf("second insert returned the first row's id %d, want a new row", first.ID)
	}

	counts, err := st.ListTaskCounts(ctx, projectKey, "kan-415")
	if err != nil {
		t.Fatalf("ListTaskCounts: %v", err)
	}
	if len(counts) != 2 {
		t.Fatalf("ListTaskCounts returned %d rows, want 2", len(counts))
	}
	if counts[0].TotalTasks != 22 || counts[1].TotalTasks != 46 {
		t.Errorf("counts = [%d, %d], want [22, 46] oldest first", counts[0].TotalTasks, counts[1].TotalTasks)
	}
	if counts[0].ObservedAt.After(counts[1].ObservedAt) {
		t.Errorf("counts[0].ObservedAt %v after counts[1].ObservedAt %v, want oldest first",
			counts[0].ObservedAt, counts[1].ObservedAt)
	}
}

// TestRecordTaskCountRefusesNonPositiveTotal: a total that counts no plan
// is refused before the store is touched, so the route answers 400, not
// 500 -- the CHECK constraint is the backstop, this check is the message.
func TestRecordTaskCountRefusesNonPositiveTotal(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-task-counts-zero-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-415")

	for _, total := range []int{0, -3} {
		if _, err := st.RecordTaskCount(ctx, projectKey, "kan-415", records.TaskCount{TotalTasks: total}); !errors.Is(err, store.ErrInvalidTaskCount) {
			t.Errorf("RecordTaskCount(total %d) error = %v, want ErrInvalidTaskCount", total, err)
		}
	}
}

// TestRecordTaskCountUnknownChange: recording against a change the store
// never heard of is ErrChangeNotFound, the same answer RecordDecision
// gives, so the CLI can distinguish a mistyped name from an unreachable
// store.
func TestRecordTaskCountUnknownChange(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-task-counts-missing-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-415")

	if _, err := st.RecordTaskCount(ctx, projectKey, "kan-other", records.TaskCount{TotalTasks: 22}); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("RecordTaskCount error = %v, want ErrChangeNotFound", err)
	}
}

// TestListTaskCountsChangeWithNone: a change with no observations lists
// empty, never an error -- the same empty-slice rule LiveStateBoard
// follows.
func TestListTaskCountsChangeWithNone(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-task-counts-empty-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-415")

	counts, err := st.ListTaskCounts(ctx, projectKey, "kan-415")
	if err != nil {
		t.Fatalf("ListTaskCounts: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("ListTaskCounts returned %d rows, want 0", len(counts))
	}
}
