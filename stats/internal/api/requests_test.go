package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// BeginRequest mirrors store.Store.BeginRequest's contract: a non-positive
// n is store.ErrInvalidRequest, an unknown change store.ErrChangeNotFound,
// and a second begin keeps a stored title an empty one would blank.
func (f *fakeStore) BeginRequest(_ context.Context, projectKey, change string, in records.RequestBegin) (records.Request, error) {
	f.recordCalls++
	if in.N <= 0 {
		return records.Request{}, fmt.Errorf("%w: n must be positive", store.ErrInvalidRequest)
	}
	key := changeKey(projectKey, change)
	if _, ok := f.changes[key]; !ok {
		return records.Request{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}
	if f.requests == nil {
		f.requests = map[string]map[int]records.Request{}
	}
	if f.requests[key] == nil {
		f.requests[key] = map[int]records.Request{}
	}
	r := f.requests[key][in.N]
	r.N = in.N
	if in.Title != "" {
		r.Title = in.Title
	}
	if in.BrainstormReason != "" {
		r.BrainstormReason = in.BrainstormReason
	}
	if in.ReceivedAt != nil && r.ReceivedAt.IsZero() {
		r.ReceivedAt = *in.ReceivedAt
	}
	r.State = "queued"
	f.requests[key][in.N] = r
	return r, nil
}

// EndRequest mirrors store.Store.EndRequest's outcome checks.
func (f *fakeStore) EndRequest(_ context.Context, projectKey, change string, in records.RequestEnd) (records.Request, error) {
	f.recordCalls++
	if in.N <= 0 || (in.Outcome != "landed" && in.Outcome != "answered" && in.Outcome != "dropped") {
		return records.Request{}, fmt.Errorf("%w: n %d outcome %q", store.ErrInvalidRequest, in.N, in.Outcome)
	}
	key := changeKey(projectKey, change)
	if _, ok := f.changes[key]; !ok {
		return records.Request{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, change)
	}
	if f.requests == nil {
		f.requests = map[string]map[int]records.Request{}
	}
	if f.requests[key] == nil {
		f.requests[key] = map[int]records.Request{}
	}
	r := f.requests[key][in.N]
	r.N, r.Outcome, r.CommitSHA, r.BaseSHA, r.EndedAt, r.State = in.N, in.Outcome, in.CommitSHA, in.BaseSHA, in.EndedAt, in.Outcome
	f.requests[key][in.N] = r
	return r, nil
}

// ListRequests reads a change's requests in n order, as the store does.
func (f *fakeStore) ListRequests(_ context.Context, projectKey, change string) ([]records.Request, error) {
	if f.listRequestsErr != nil {
		return nil, f.listRequestsErr
	}
	var out []records.Request
	for _, r := range f.requests[changeKey(projectKey, change)] {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].N < out[j].N })
	return out, nil
}

// TestRequestRoutes pins the three request routes (KAN-1061): a begin and
// an end each answer 200 with the stored row, the list reads them back in
// n order, a store refusal of the record is a 400 and an unknown change a
// 404.
func TestRequestRoutes(t *testing.T) {
	ts, _ := recordTestServer(t, "proj", "kan-1")
	base := ts.URL + recordsPath("proj", "kan-1") + "/requests"

	resp, body := postJSON(t, base, map[string]any{"n": 2, "title": "second", "receivedAt": "2026-10-30T10:00:00Z"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST requests = %d (%s), want 200", resp.StatusCode, body)
	}
	var got records.Request
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if got.N != 2 || got.Title != "second" {
		t.Errorf("begin row = %+v, want n 2 titled second", got)
	}
	if resp, body := postJSON(t, base, map[string]any{"n": 1, "title": "first"}); resp.StatusCode != http.StatusOK {
		t.Fatalf("POST requests n=1 = %d (%s), want 200", resp.StatusCode, body)
	}

	resp, body = postJSON(t, base+"/end", map[string]any{"n": 1, "outcome": "landed", "commitSha": "abc1234", "baseSha": "def5678"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST requests/end = %d (%s), want 200", resp.StatusCode, body)
	}
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if got.Outcome != "landed" || got.CommitSHA != "abc1234" || got.BaseSHA != "def5678" {
		t.Errorf("end row = %+v, want landed abc1234 on def5678", got)
	}

	status, list := doGet(t, ts, recordsPath("proj", "kan-1")+"/requests")
	if status != http.StatusOK {
		t.Fatalf("GET requests = %d (%s), want 200", status, list)
	}
	var rows []records.Request
	if err := json.Unmarshal([]byte(list), &rows); err != nil {
		t.Fatalf("decode %s: %v", list, err)
	}
	if len(rows) != 2 || rows[0].N != 1 || rows[1].N != 2 || rows[0].Outcome != "landed" {
		t.Errorf("list = %+v, want requests 1 (landed) then 2", rows)
	}

	for name, tc := range map[string]struct {
		url  string
		body map[string]any
		want int
	}{
		"begin n 0":       {base, map[string]any{"n": 0, "title": "t"}, http.StatusBadRequest},
		"end bad outcome": {base + "/end", map[string]any{"n": 1, "outcome": "merged"}, http.StatusBadRequest},
		"unknown change":  {ts.URL + recordsPath("proj", "nope") + "/requests", map[string]any{"n": 1, "title": "t"}, http.StatusNotFound},
		"end unknown":     {ts.URL + recordsPath("proj", "nope") + "/requests/end", map[string]any{"n": 1, "outcome": "answered"}, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			resp, body := postJSON(t, tc.url, tc.body)
			if resp.StatusCode != tc.want {
				t.Errorf("POST = %d (%s), want %d", resp.StatusCode, body, tc.want)
			}
		})
	}
}
