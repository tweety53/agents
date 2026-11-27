package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestEndSessionPermission pins how a subagent's later activity reads on the
// board: a permission notification followed by that activity no longer reads
// as awaiting permission, while an open question or an idle main thread is
// left as it was -- a subagent's tool call answers neither.
func TestEndSessionPermission(t *testing.T) {
	t.Parallel()
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-end-permission-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1")
	open := baseBeginInput(projectKey, "kan-1", "/flow", "flow.implement")
	session := fmt.Sprintf("s-perm-%d", time.Now().UnixNano())
	open.SessionID = &session
	open.StartedAt = time.Date(2026, 11, 27, 21, 0, 0, 0, time.UTC)
	if _, err := st.BeginStage(ctx, open); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	activity := func() string {
		t.Helper()
		rows, err := st.LiveStateBoard(ctx, store.Period{
			From: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			To:   time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC),
		}, &projectKey, nil)
		if err != nil || len(rows) != 1 || rows[0].Activity == nil {
			t.Fatalf("LiveStateBoard = %+v, %v; want one row with activity", rows, err)
		}
		return *rows[0].Activity + "/" + *rows[0].ActivityDetail
	}

	for _, c := range []struct{ before, want string }{
		{"permission", "working/Bash: run tests"},
		{"question", "question/x"},
		{"idle", "idle/x"},
	} {
		if err := st.SetSessionActivity(ctx, session, c.before, "x"); err != nil {
			t.Fatalf("SetSessionActivity: %v", err)
		}
		if err := st.EndSessionPermission(ctx, session, "Bash: run tests"); err != nil {
			t.Fatalf("EndSessionPermission: %v", err)
		}
		if got := activity(); got != c.want {
			t.Errorf("%s then a subagent's tool call: activity %q, want %q", c.before, got, c.want)
		}
	}
	// A session with no row yet gets none: the subagent's call creates nothing.
	if err := st.EndSessionPermission(ctx, "s-none", "Read"); err != nil {
		t.Fatalf("EndSessionPermission(no row): %v", err)
	}
}
