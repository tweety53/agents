package api_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/records"
	"github.com/tweety53/agents/stats/internal/store"
)

// lockedStore serialises fakeStore's dispatch reads and writes, which the
// long-poll handler and a closing POST make from two goroutines; reads
// signals each RunRecord, so a test knows the waiter has read before it
// closes a row.
type lockedStore struct {
	*fakeStore
	mu    sync.Mutex
	reads chan struct{}
}

func (l *lockedStore) RecordDispatch(ctx context.Context, p, c string, in records.Dispatch) (records.Dispatch, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.fakeStore.RecordDispatch(ctx, p, c, in)
}

func (l *lockedStore) EndDispatch(ctx context.Context, p, c string, in records.DispatchEnd) (records.Dispatch, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.fakeStore.EndDispatch(ctx, p, c, in)
}

func (l *lockedStore) RunRecord(ctx context.Context, p, c string) (records.Run, error) {
	l.mu.Lock()
	run, err := l.fakeStore.RunRecord(ctx, p, c)
	l.mu.Unlock()
	select {
	case l.reads <- struct{}{}:
	default:
	}
	return run, err
}

// waitTestServer serves one change, kan-1, holding open rows for keys.
func waitTestServer(t *testing.T, keys ...string) (*httptest.Server, *lockedStore) {
	t.Helper()
	fs := newFakeStore()
	fs.changes[changeKey("proj", "kan-1")] = store.Change{ProjectKey: "proj", Name: "kan-1", State: store.StateInProgress}
	ls := &lockedStore{fakeStore: fs, reads: make(chan struct{}, 1)}
	srv, err := api.New(config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}, fs, fs, fs, ls, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	for _, k := range keys {
		body := dispatchBody("implementer", "opus")
		body["key"] = k
		if resp, b := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches", body); resp.StatusCode != http.StatusCreated {
			t.Fatalf("POST dispatches %s = %d (%s)", k, resp.StatusCode, b)
		}
	}
	return ts, ls
}

type waitResult struct {
	status int
	rows   []records.Dispatch
	took   time.Duration
}

// waitGet issues the long-poll and returns its answer on a channel.
func waitGet(t *testing.T, url string) <-chan waitResult {
	t.Helper()
	out := make(chan waitResult, 1)
	go func() {
		start := time.Now()
		resp, err := http.Get(url)
		if err != nil {
			out <- waitResult{status: -1}
			return
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		var rows []records.Dispatch
		_ = json.Unmarshal(b, &rows)
		out <- waitResult{status: resp.StatusCode, rows: rows, took: time.Since(start)}
	}()
	return out
}

// awaitRead blocks until the waiter has read the store once.
func awaitRead(t *testing.T, ls *lockedStore) {
	t.Helper()
	select {
	case <-ls.reads:
	case <-time.After(5 * time.Second):
		t.Fatal("the wait route never read the store")
	}
}

func endKey(t *testing.T, ts *httptest.Server, key string) {
	t.Helper()
	if resp, b := postJSON(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/end", dispatchEndBody(key)); resp.StatusCode != http.StatusOK {
		t.Fatalf("POST dispatches/end %s = %d (%s)", key, resp.StatusCode, b)
	}
}

func endedKeys(rows []records.Dispatch) map[string]bool {
	m := map[string]bool{}
	for _, r := range rows {
		m[r.Key] = r.EndedAt != nil
	}
	return m
}

// TestWaitDispatchesWakesOnEnd: a long-poll on an open row answers the
// moment that row is closed -- well inside the handler's own recheck
// interval, so the wakeup, not the recheck, is what answered.
func TestWaitDispatchesWakesOnEnd(t *testing.T) {
	t.Parallel()
	ts, ls := waitTestServer(t, "a")
	got := waitGet(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/wait?keys=a&timeout=15")
	awaitRead(t, ls)
	time.Sleep(50 * time.Millisecond)
	start := time.Now()
	endKey(t, ts, "a")
	r := <-got
	if r.status != http.StatusOK || !endedKeys(r.rows)["a"] {
		t.Fatalf("status %d rows %+v, want 200 with a ended", r.status, r.rows)
	}
	if since := time.Since(start); since > 2*time.Second {
		t.Errorf("answered %s after the close, want an immediate wakeup", since)
	}
}

// TestWaitDispatchesAnyVsAll: mode=any answers once one named row closes;
// mode=all keeps waiting while another is open, answering at its timeout
// with the rows as they stand.
func TestWaitDispatchesAnyVsAll(t *testing.T) {
	t.Parallel()
	ts, ls := waitTestServer(t, "a", "b")
	anyCh := waitGet(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/wait?keys=a,b&mode=any&timeout=15")
	awaitRead(t, ls)
	time.Sleep(50 * time.Millisecond)
	endKey(t, ts, "a")
	if r := <-anyCh; r.status != http.StatusOK || r.took > 3*time.Second || !endedKeys(r.rows)["a"] || endedKeys(r.rows)["b"] {
		t.Fatalf("mode=any: status %d after %s rows %+v, want 200 at once with a ended, b open", r.status, r.took, r.rows)
	}
	r := <-waitGet(t, ts.URL+recordsPath("proj", "kan-1")+"/dispatches/wait?keys=a,b&mode=all&timeout=1")
	if r.status != http.StatusOK || r.took < 900*time.Millisecond || endedKeys(r.rows)["b"] {
		t.Fatalf("mode=all: status %d after %s rows %+v, want 200 at its 1 s timeout with b open", r.status, r.took, r.rows)
	}
}

// TestWaitDispatchesAnswersAtOnce: a wait already satisfied, or naming a key
// the change does not record, answers without waiting; a bad mode or no
// keys is 400.
func TestWaitDispatchesAnswersAtOnce(t *testing.T) {
	t.Parallel()
	ts, _ := waitTestServer(t, "a")
	endKey(t, ts, "a")
	base := ts.URL + recordsPath("proj", "kan-1") + "/dispatches/wait?"
	for _, q := range []string{"keys=a&timeout=15", "keys=a,zzz&mode=all&timeout=15"} {
		if r := <-waitGet(t, base+q); r.status != http.StatusOK || r.took > 2*time.Second {
			t.Errorf("%s: status %d after %s, want 200 at once", q, r.status, r.took)
		}
	}
	for _, q := range []string{"keys=a&mode=some", "keys=", "keys=a&timeout=x"} {
		if r := <-waitGet(t, base+q); r.status != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", q, r.status)
		}
	}
}
