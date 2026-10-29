package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// fakeExperiments is an in-memory api.ExperimentStore: it validates like
// the real store and answers with whatever the test configured.
type fakeExperiments struct {
	list      []store.Experiment
	openErr   error
	closeErr  error
	openOnly  bool
	closedID  int64
	closedArg [2]string
}

func (f *fakeExperiments) OpenExperiment(_ context.Context, e store.Experiment) (store.Experiment, error) {
	if err := store.ValidateExperiment(e); err != nil {
		return store.Experiment{}, err
	}
	if f.openErr != nil {
		return store.Experiment{}, f.openErr
	}
	e.ID, e.StartedAt, e.Status = 7, time.Date(2026, 10, 29, 0, 0, 0, 0, time.UTC), "open"
	return e, nil
}

func (f *fakeExperiments) ListExperiments(_ context.Context, openOnly bool) ([]store.Experiment, error) {
	f.openOnly = openOnly
	return f.list, nil
}

func (f *fakeExperiments) OpenOverrides(_ context.Context) (map[string][]string, error) {
	out := map[string][]string{}
	for _, e := range f.list {
		if e.Status == "open" {
			for r, p := range e.Cells {
				out[r] = p
			}
		}
	}
	return out, nil
}

func (f *fakeExperiments) CloseExperiment(_ context.Context, id int64, verdict, summary string) error {
	if !store.ValidVerdicts[verdict] {
		return store.ErrInvalidVerdict
	}
	if f.closeErr != nil {
		return f.closeErr
	}
	f.closedID, f.closedArg = id, [2]string{verdict, summary}
	return nil
}

func newExperimentsTestServer(t *testing.T, fe *fakeExperiments) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	fs := newFakeStore()
	srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil, api.WithExperiments(fe))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func doJSON(t *testing.T, method, url, body string) (int, string) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build %s %s: %v", method, url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(b)
}

// TestExperimentsRoutes drives every experiments endpoint through its
// success and refusal statuses.
func TestExperimentsRoutes(t *testing.T) {
	fe := &fakeExperiments{}
	ts := newExperimentsTestServer(t, fe)
	base := ts.URL + "/api/v1/experiments"

	status, body := doJSON(t, http.MethodPost, base, `{"name":"t","cells":{"explorer":["sonnet/Explore@low"]}}`)
	if status != http.StatusCreated || !strings.Contains(body, `"id":7`) {
		t.Fatalf("POST open = %d %s, want 201 with id 7", status, body)
	}
	if status, body = doJSON(t, http.MethodPost, base, `{"name":"","cells":{"explorer":["sonnet/Explore@low"]}}`); status != http.StatusBadRequest {
		t.Fatalf("POST invalid = %d %s, want 400", status, body)
	}
	if status, body = doJSON(t, http.MethodPost, base, `{"name":"t","cells":{},"bogus":1}`); status != http.StatusBadRequest {
		t.Fatalf("POST unknown field = %d %s, want 400", status, body)
	}
	fe.openErr = fmt.Errorf("%w: role %q is held by open experiment 3", store.ErrExperimentConflict, "explorer")
	status, body = doJSON(t, http.MethodPost, base, `{"name":"t","cells":{"explorer":["sonnet/Explore@low"]}}`)
	if status != http.StatusConflict || !strings.Contains(body, "held by open experiment 3") {
		t.Fatalf("POST conflict = %d %s, want 409 naming the holder", status, body)
	}

	fe.list = []store.Experiment{{ID: 3, Name: "a", Status: "open", Cells: map[string][]string{"explorer": {"sonnet/Explore@low"}}}}
	status, body = doJSON(t, http.MethodGet, base+"?open=1", "")
	if status != http.StatusOK || !fe.openOnly || !strings.Contains(body, `"id":3`) {
		t.Fatalf("GET ?open=1 = %d %s (openOnly %v), want 200 listing 3 with openOnly", status, body, fe.openOnly)
	}
	if status, _ = doJSON(t, http.MethodGet, base, ""); status != http.StatusOK || fe.openOnly {
		t.Fatalf("GET all = %d (openOnly %v), want 200 without openOnly", status, fe.openOnly)
	}

	status, body = doJSON(t, http.MethodGet, base+"/overrides", "")
	var overrides map[string][]string
	if status != http.StatusOK || json.Unmarshal([]byte(body), &overrides) != nil ||
		!reflect.DeepEqual(overrides, map[string][]string{"explorer": {"sonnet/Explore@low"}}) {
		t.Fatalf("GET overrides = %d %s, want 200 {explorer:[sonnet/Explore@low]}", status, body)
	}

	if status, body = doJSON(t, http.MethodPost, base+"/3/close", `{"verdict":"negative","summary":"s"}`); status != http.StatusOK || fe.closedID != 3 || fe.closedArg != [2]string{"negative", "s"} {
		t.Fatalf("POST close = %d %s (closed %d %v), want 200 closing 3 negative s", status, body, fe.closedID, fe.closedArg)
	}
	if status, _ = doJSON(t, http.MethodPost, base+"/3/close", `{"verdict":"great","summary":"s"}`); status != http.StatusBadRequest {
		t.Fatalf("POST close bad verdict = %d, want 400", status)
	}
	if status, _ = doJSON(t, http.MethodPost, base+"/x/close", `{"verdict":"negative","summary":"s"}`); status != http.StatusBadRequest {
		t.Fatalf("POST close bad id = %d, want 400", status)
	}
	fe.closeErr = fmt.Errorf("%w: 9", store.ErrExperimentNotOpen)
	if status, _ = doJSON(t, http.MethodPost, base+"/9/close", `{"verdict":"negative","summary":"s"}`); status != http.StatusNotFound {
		t.Fatalf("POST close not open = %d, want 404", status)
	}
}
