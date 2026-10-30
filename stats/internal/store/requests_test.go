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

// TestRequestBeginUpserts pins BeginRequest's upsert (KAN-1061): the first
// begin stamps received_at and reads back queued; a second begin carrying
// only a brainstorm reason keeps the stored title and the first receipt;
// the begin bumps the change's updated_at; a non-positive n and an unknown
// change are refused.
func TestRequestBeginUpserts(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-begin-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1061")

	before := time.Now().Add(-time.Second)
	first, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: 1, Title: "a"})
	if err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	if first.ReceivedAt.IsZero() {
		t.Errorf("ReceivedAt is zero, want the store's stamp")
	}
	if first.State != "queued" {
		t.Errorf("State = %q, want queued", first.State)
	}

	second, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: 1, BrainstormReason: "r"})
	if err != nil {
		t.Fatalf("second BeginRequest: %v", err)
	}
	if second.Title != "a" {
		t.Errorf("Title = %q, want a -- an omitted title must not blank the stored one", second.Title)
	}
	if second.BrainstormReason != "r" {
		t.Errorf("BrainstormReason = %q, want r", second.BrainstormReason)
	}
	if !second.ReceivedAt.Equal(first.ReceivedAt) {
		t.Errorf("ReceivedAt = %v, want the first begin's %v", second.ReceivedAt, first.ReceivedAt)
	}

	third, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: 1, Title: "b"})
	if err != nil {
		t.Fatalf("third BeginRequest: %v", err)
	}
	if third.BrainstormReason != "r" {
		t.Errorf("BrainstormReason = %q, want r -- an omitted reason must not blank the stored one", third.BrainstormReason)
	}
	if third.Title != "b" {
		t.Errorf("Title = %q, want b", third.Title)
	}

	c, err := st.GetChange(ctx, projectKey, "kan-1061")
	if err != nil {
		t.Fatalf("GetChange: %v", err)
	}
	if c.UpdatedAt.Before(before) {
		t.Errorf("change UpdatedAt = %v, want not earlier than the begin instant %v", c.UpdatedAt, before)
	}

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: 0, Title: "x"}); !errors.Is(err, store.ErrInvalidRequest) {
		t.Errorf("BeginRequest n=0: err = %v, want ErrInvalidRequest", err)
	}
	if _, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: -1, Title: "x"}); !errors.Is(err, store.ErrInvalidRequest) {
		t.Errorf("BeginRequest n=-1: err = %v, want ErrInvalidRequest", err)
	}
	if _, err := st.BeginRequest(ctx, projectKey, "no-such-change", records.RequestBegin{N: 1, Title: "x"}); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("BeginRequest unknown change: err = %v, want ErrChangeNotFound", err)
	}
}

// TestRequestEndRecordsOutcome pins EndRequest: landed carries both shas,
// answered and dropped carry neither, any other shape is refused, and an
// end for a request never begun inserts it.
func TestRequestEndRecordsOutcome(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-end-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1061")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: 1, Title: "a"}); err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	landed, err := st.EndRequest(ctx, projectKey, "kan-1061", records.RequestEnd{N: 1, Outcome: "landed", CommitSHA: "c1", BaseSHA: "b1"})
	if err != nil {
		t.Fatalf("EndRequest landed: %v", err)
	}
	if landed.EndedAt == nil {
		t.Errorf("EndedAt is nil, want the end's stamp")
	}
	if landed.CommitSHA != "c1" || landed.BaseSHA != "b1" {
		t.Errorf("shas = %q/%q, want c1/b1", landed.CommitSHA, landed.BaseSHA)
	}
	if landed.State != "landed" || landed.Outcome != "landed" {
		t.Errorf("State/Outcome = %q/%q, want landed/landed", landed.State, landed.Outcome)
	}
	if landed.Title != "a" {
		t.Errorf("Title = %q, want a -- an end must not blank the begin's title", landed.Title)
	}

	for n, outcome := range map[int]string{2: "answered", 3: "dropped"} {
		got, err := st.EndRequest(ctx, projectKey, "kan-1061", records.RequestEnd{N: n, Outcome: outcome})
		if err != nil {
			t.Fatalf("EndRequest %s: %v", outcome, err)
		}
		if got.State != outcome {
			t.Errorf("EndRequest %s: State = %q", outcome, got.State)
		}
	}

	bad := []records.RequestEnd{
		{N: 4, Outcome: "landed", CommitSHA: "c"},
		{N: 4, Outcome: "landed", BaseSHA: "b"},
		{N: 4, Outcome: "answered", CommitSHA: "c"},
		{N: 4, Outcome: "dropped", BaseSHA: "b"},
		{N: 4, Outcome: "shipped"},
		{N: 0, Outcome: "answered"},
	}
	for _, in := range bad {
		if _, err := st.EndRequest(ctx, projectKey, "kan-1061", in); !errors.Is(err, store.ErrInvalidRequest) {
			t.Errorf("EndRequest %+v: err = %v, want ErrInvalidRequest", in, err)
		}
	}
	if _, err := st.EndRequest(ctx, projectKey, "no-such-change", records.RequestEnd{N: 1, Outcome: "answered"}); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("EndRequest unknown change: err = %v, want ErrChangeNotFound", err)
	}

	list, err := st.ListRequests(ctx, projectKey, "kan-1061")
	if err != nil {
		t.Fatalf("ListRequests: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("ListRequests returned %d rows, want 3 -- an end for a request never begun inserts it, a refused one does not", len(list))
	}
}

// TestListRequestsDerivesState pins requestState's derivation: an ended
// request reads its outcome, one a stage run carries reads in progress,
// one with neither reads queued -- returned in n order.
func TestListRequestsDerivesState(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-state-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1061")

	for _, n := range []int{3, 1, 2} {
		if _, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: n, Title: fmt.Sprintf("r%d", n)}); err != nil {
			t.Fatalf("BeginRequest %d: %v", n, err)
		}
	}
	if _, err := st.EndRequest(ctx, projectKey, "kan-1061", records.RequestEnd{N: 1, Outcome: "answered"}); err != nil {
		t.Fatalf("EndRequest: %v", err)
	}
	in := baseBeginInput(projectKey, "kan-1061", "/flow-interactive", "flow.sdd-tdd")
	in.RequestN = ptr(2)
	if _, err := st.BeginStage(ctx, in); err != nil {
		t.Fatalf("BeginStage: %v", err)
	}

	list, err := st.ListRequests(ctx, projectKey, "kan-1061")
	if err != nil {
		t.Fatalf("ListRequests: %v", err)
	}
	want := []struct {
		n     int
		state string
	}{{1, "answered"}, {2, "in progress"}, {3, "queued"}}
	if len(list) != len(want) {
		t.Fatalf("ListRequests returned %d rows, want %d", len(list), len(want))
	}
	for i, w := range want {
		if list[i].N != w.n || list[i].State != w.state {
			t.Errorf("list[%d] = n %d state %q, want n %d state %q", i, list[i].N, list[i].State, w.n, w.state)
		}
	}
}
