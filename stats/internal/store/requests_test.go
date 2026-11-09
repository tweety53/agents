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
// end for a request never begun is ErrInvalidRequest and writes no row
// (kan-1081 self-review, F6: such ends made untitled requests 18 and 86).
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
		if _, err := st.BeginRequest(ctx, projectKey, "kan-1061", records.RequestBegin{N: n, Title: outcome}); err != nil {
			t.Fatalf("BeginRequest %d: %v", n, err)
		}
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
		{N: 4, Outcome: "answered"},
		{N: 5, Outcome: "landed", CommitSHA: "c5", BaseSHA: "b5"},
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
		t.Fatalf("ListRequests returned %d rows, want 3 -- an end for a request never begun writes no row", len(list))
	}
}

// TestRequestBeginRefusesUntitledFirst pins request-title-required
// (kan-1081 self-review, F6): a request's first begin needs a title -- one
// carrying only a brainstorm reason is ErrInvalidRequest and writes no row
// -- while a later begin may carry only a reason and keeps the title.
func TestRequestBeginRefusesUntitledFirst(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-untitled-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1081")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 1, BrainstormReason: "r"}); !errors.Is(err, store.ErrInvalidRequest) {
		t.Errorf("untitled first begin: err = %v, want ErrInvalidRequest", err)
	}
	if list, err := st.ListRequests(ctx, projectKey, "kan-1081"); err != nil || len(list) != 0 {
		t.Fatalf("ListRequests = %+v, %v, want no row", list, err)
	}

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 2, Title: "t"}); err != nil {
		t.Fatalf("titled begin: %v", err)
	}
	got, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 2, BrainstormReason: "r"})
	if err != nil {
		t.Fatalf("reason-only begin of a begun request: %v", err)
	}
	if got.Title != "t" || got.BrainstormReason != "r" {
		t.Errorf("row = %+v, want title t kept and reason r", got)
	}
}

// TestRequestBeginFillsEmptyTitle pins the repair of an ended, untitled
// request (kan-1081 self-review, F6): a begin naming a title fills the
// empty one and leaves the outcome; a stored title is never changed after.
func TestRequestBeginFillsEmptyTitle(t *testing.T) {
	st, pool := newRecordStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-filltitle-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1081")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 18, Title: "x"}); err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	if _, err := st.EndRequest(ctx, projectKey, "kan-1081", records.RequestEnd{N: 18, Outcome: "landed", CommitSHA: "c", BaseSHA: "b"}); err != nil {
		t.Fatalf("EndRequest: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE requests SET title = '' WHERE n = 18`); err != nil {
		t.Fatalf("blank the title: %v", err)
	}

	got, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 18, Title: "t"})
	if err != nil {
		t.Fatalf("fill the empty title: %v", err)
	}
	if got.Title != "t" || got.Outcome != "landed" || got.CommitSHA != "c" || got.BaseSHA != "b" {
		t.Errorf("row = %+v, want title t, outcome landed c/b unchanged", got)
	}
	if _, err := st.BeginRequest(ctx, projectKey, "kan-1081", records.RequestBegin{N: 18, Title: "u"}); !errors.Is(err, store.ErrRequestEnded) {
		t.Errorf("change a stored title: err = %v, want ErrRequestEnded", err)
	}
	list, err := st.ListRequests(ctx, projectKey, "kan-1081")
	if err != nil || len(list) != 1 || list[0].Title != "t" {
		t.Errorf("ListRequests = %+v, %v, want title t", list, err)
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

// TestRequestWritesRefuseChangingAnEndedRequest pins sr2-ended-request-conflict
// (kan-1070 self-review, F2): once a request has an outcome, a begin with a
// different title and an end with a different outcome, commit or base are
// refused with ErrRequestEnded and change nothing, while an identical replay
// of either is accepted; an open request is still amended.
func TestRequestWritesRefuseChangingAnEndedRequest(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-ended-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1070")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 1, Title: "a"}); err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	ended, err := st.EndRequest(ctx, projectKey, "kan-1070", records.RequestEnd{N: 1, Outcome: "landed", CommitSHA: "c1", BaseSHA: "b1"})
	if err != nil {
		t.Fatalf("EndRequest: %v", err)
	}

	readBack := func() records.Request {
		t.Helper()
		list, err := st.ListRequests(ctx, projectKey, "kan-1070")
		if err != nil || len(list) == 0 || list[0].N != 1 {
			t.Fatalf("ListRequests = %+v, %v", list, err)
		}
		return list[0]
	}
	unchanged := func(what string) {
		t.Helper()
		got := readBack()
		if got.Title != "a" || got.Outcome != "landed" || got.CommitSHA != "c1" || got.BaseSHA != "b1" ||
			got.EndedAt == nil || !got.EndedAt.Equal(*ended.EndedAt) {
			t.Errorf("%s: row = %+v, want it unchanged", what, got)
		}
	}

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 1, Title: "b"}); !errors.Is(err, store.ErrRequestEnded) {
		t.Errorf("begin with a new title: err = %v, want ErrRequestEnded", err)
	}
	unchanged("begin with a new title")
	for _, in := range []records.RequestEnd{
		{N: 1, Outcome: "answered"},
		{N: 1, Outcome: "landed", CommitSHA: "c2", BaseSHA: "b1"},
	} {
		if _, err := st.EndRequest(ctx, projectKey, "kan-1070", in); !errors.Is(err, store.ErrRequestEnded) {
			t.Errorf("end %+v: err = %v, want ErrRequestEnded", in, err)
		}
		unchanged(fmt.Sprintf("end %+v", in))
	}

	for _, title := range []string{"a", ""} {
		if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 1, Title: title}); err != nil {
			t.Errorf("identical begin title %q: %v", title, err)
		}
		unchanged(fmt.Sprintf("identical begin title %q", title))
	}
	if _, err := st.EndRequest(ctx, projectKey, "kan-1070", records.RequestEnd{N: 1, Outcome: "landed", CommitSHA: "c1", BaseSHA: "b1"}); err != nil {
		t.Errorf("identical end: %v", err)
	}
	unchanged("identical end")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 2, Title: "x"}); err != nil {
		t.Fatalf("BeginRequest n 2: %v", err)
	}
	if got, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 2, Title: "y"}); err != nil || got.Title != "y" {
		t.Errorf("amend open n 2: got %+v, %v, want title y", got, err)
	}
	if got, err := st.EndRequest(ctx, projectKey, "kan-1070", records.RequestEnd{N: 2, Outcome: "answered"}); err != nil || got.Outcome != "answered" {
		t.Errorf("end open n 2: got %+v, %v, want answered", got, err)
	}
}

// TestRequestRepairWrites pins sr2-repair-commands: PutRequest overwrites an
// ended row whole and inserts a new n, refusing a malformed outcome and an
// unknown change; SetStageRunRequest moves a stage run's request, refusing
// n 0 and an unknown id.
func TestRequestRepairWrites(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-requests-repair-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1070")

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 1, Title: "wrong"}); err != nil {
		t.Fatalf("BeginRequest: %v", err)
	}
	if _, err := st.EndRequest(ctx, projectKey, "kan-1070", records.RequestEnd{N: 1, Outcome: "landed", CommitSHA: "c9", BaseSHA: "b9"}); err != nil {
		t.Fatalf("EndRequest: %v", err)
	}

	received := time.Date(2026, 10, 31, 10, 27, 31, 0, time.UTC)
	endedAt := time.Date(2026, 10, 31, 10, 28, 5, 0, time.UTC)
	if err := st.PutRequest(ctx, projectKey, "kan-1070", records.Request{N: 1, Title: "right", ReceivedAt: received, EndedAt: &endedAt, Outcome: "answered"}); err != nil {
		t.Fatalf("PutRequest overwrite: %v", err)
	}
	if err := st.PutRequest(ctx, projectKey, "kan-1070", records.Request{N: 31, Title: "moved", ReceivedAt: received, EndedAt: &endedAt, Outcome: "landed", CommitSHA: "c9", BaseSHA: "b9"}); err != nil {
		t.Fatalf("PutRequest insert: %v", err)
	}
	list, err := st.ListRequests(ctx, projectKey, "kan-1070")
	if err != nil || len(list) != 2 {
		t.Fatalf("ListRequests = %+v, %v, want 2 rows", list, err)
	}
	if r := list[0]; r.N != 1 || r.Title != "right" || r.Outcome != "answered" || r.CommitSHA != "" || r.BaseSHA != "" ||
		!r.ReceivedAt.Equal(received) || r.EndedAt == nil || !r.EndedAt.Equal(endedAt) {
		t.Errorf("row 1 = %+v, want it overwritten whole", r)
	}
	if r := list[1]; r.N != 31 || r.Title != "moved" || r.Outcome != "landed" || r.CommitSHA != "c9" || r.BaseSHA != "b9" {
		t.Errorf("row 31 = %+v, want the inserted row", r)
	}

	if err := st.PutRequest(ctx, projectKey, "kan-1070", records.Request{N: 2, Title: "t", ReceivedAt: received, EndedAt: &endedAt, Outcome: "landed"}); !errors.Is(err, store.ErrInvalidRequest) {
		t.Errorf("PutRequest landed with no commit: err = %v, want ErrInvalidRequest", err)
	}
	if err := st.PutRequest(ctx, projectKey, "no-such-change", records.Request{N: 1, Title: "t", ReceivedAt: received, EndedAt: &endedAt, Outcome: "answered"}); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("PutRequest unknown change: err = %v, want ErrChangeNotFound", err)
	}

	if _, err := st.BeginRequest(ctx, projectKey, "kan-1070", records.RequestBegin{N: 7, Title: "queued"}); err != nil {
		t.Fatalf("BeginRequest n 7: %v", err)
	}
	run, err := st.BeginStage(ctx, baseBeginInput(projectKey, "kan-1070", "/flow-interactive", "flow.sdd-tdd"))
	if err != nil {
		t.Fatalf("BeginStage: %v", err)
	}
	if err := st.SetStageRunRequest(ctx, run.ID, 7); err != nil {
		t.Fatalf("SetStageRunRequest: %v", err)
	}
	list, err = st.ListRequests(ctx, projectKey, "kan-1070")
	if err != nil || len(list) != 3 || list[1].N != 7 || list[1].State != "in progress" {
		t.Errorf("ListRequests after SetStageRunRequest = %+v, %v, want n 7 in progress", list, err)
	}
	if err := st.SetStageRunRequest(ctx, run.ID, 0); !errors.Is(err, store.ErrInvalidRequest) {
		t.Errorf("SetStageRunRequest n 0: err = %v, want ErrInvalidRequest", err)
	}
	if err := st.SetStageRunRequest(ctx, -1, 7); err == nil {
		t.Errorf("SetStageRunRequest unknown id: err = nil, want an error")
	}
}
