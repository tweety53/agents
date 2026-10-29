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
	"github.com/tweety53/agents/stats/internal/modelpolicy/embedded"
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
	statsID   int64
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

func (f *fakeExperiments) ExperimentStats(_ context.Context, id int64) (store.ExperimentStatsResult, error) {
	if id != 3 {
		return store.ExperimentStatsResult{}, fmt.Errorf("%w: %d", store.ErrExperimentNotFound, id)
	}
	f.statsID = id
	return store.ExperimentStatsResult{Experiment: store.Experiment{ID: 3}, Runs: store.ExperimentRuns{Before: 4, After: 2},
		Metrics: []store.ExperimentMetric{{Name: "tokens", Series: []store.ExperimentWeek{}}}}, nil
}

func newExperimentsTestServer(t *testing.T, fe *fakeExperiments) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	fs := newFakeStore()
	srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil, api.WithExperiments(fe, embedded.Table))
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
	if status, body = doJSON(t, http.MethodPost, base, `{"name":"t","cells":{"plan-desginer":["opus/Plan@xhigh"]}}`); status != http.StatusBadRequest || !strings.Contains(body, "plan-desginer") {
		t.Fatalf("POST unknown role = %d %s, want 400 naming the role", status, body)
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

// TestExperimentStatsRoute serves GET /api/v1/experiments/{id}/stats:
// 200 with the store's figures, 404 for an unknown id, 400 for a
// non-integer one.
func TestExperimentStatsRoute(t *testing.T) {
	fe := &fakeExperiments{}
	ts := newExperimentsTestServer(t, fe)
	base := ts.URL + "/api/v1/experiments"

	status, body := doJSON(t, http.MethodGet, base+"/3/stats", "")
	if status != http.StatusOK || fe.statsID != 3 || !strings.Contains(body, `"runs":{"before":4,"after":2}`) || !strings.Contains(body, `"name":"tokens"`) {
		t.Fatalf("GET stats = %d %s, want 200 with experiment 3's figures", status, body)
	}
	if status, _ = doJSON(t, http.MethodGet, base+"/9/stats", ""); status != http.StatusNotFound {
		t.Fatalf("GET stats unknown = %d, want 404", status)
	}
	if status, _ = doJSON(t, http.MethodGet, base+"/x/stats", ""); status != http.StatusBadRequest {
		t.Fatalf("GET stats bad id = %d, want 400", status)
	}
}

// TestRolesRoute reads GET /api/v1/roles over the real embedded table: a
// leading main entry, then one entry per table row in table order, an open
// experiment's cell on its role's unmarked row and on main, never on a
// marked row, and never read off a closed experiment.
func TestRolesRoute(t *testing.T) {
	type override struct {
		Pairs        []string `json:"pairs"`
		ExperimentID int64    `json:"experimentId"`
	}
	type entry struct {
		Role     string    `json:"role"`
		Pairs    []string  `json:"pairs"`
		Marker   string    `json:"marker"`
		Override *override `json:"override"`
	}
	fe := &fakeExperiments{list: []store.Experiment{{ID: 5, Name: "x", Status: "open", Cells: map[string][]string{
		"plan-designer": {"opus/Plan@xhigh"},
		"implementer":   {"opus/flow-medium"},
		"main":          {"opus@xhigh"},
	}}}}
	ts := newExperimentsTestServer(t, fe)

	status, body := doJSON(t, http.MethodGet, ts.URL+"/api/v1/roles", "")
	var got []entry
	if status != http.StatusOK || json.Unmarshal([]byte(body), &got) != nil {
		t.Fatalf("GET roles = %d %s, want 200 with a JSON list", status, body)
	}
	if !fe.openOnly {
		t.Fatal("roles read every experiment, want open ones only")
	}
	if len(got) < 6 {
		t.Fatalf("GET roles = %s, want main plus the table's rows", body)
	}
	if !reflect.DeepEqual(got[0], entry{Role: "main", Pairs: []string{}, Override: &override{Pairs: []string{"opus@xhigh"}, ExperimentID: 5}}) {
		t.Errorf("entry 0 = %+v, want main carrying experiment 5's override", got[0])
	}
	if got[1].Role != "explorer" || got[1].Override != nil {
		t.Errorf("entry 1 = %+v, want explorer, no override", got[1])
	}
	if got[2].Role != "plan-designer" || got[2].Override == nil ||
		!reflect.DeepEqual(*got[2].Override, override{Pairs: []string{"opus/Plan@xhigh"}, ExperimentID: 5}) {
		t.Errorf("entry 2 = %+v, want plan-designer carrying experiment 5's override", got[2])
	}
	// Look the implementer rows up by role and marker, so a row added to the
	// table elsewhere never shifts what this asserts; table order still holds
	// (the unmarked row, then panel-fix, then the marked escalation row).
	at := func(role, marker string) int {
		for i, e := range got {
			if e.Role == role && e.Marker == marker {
				return i
			}
		}
		t.Fatalf("GET roles = %s, want a %s row marked %q", body, role, marker)
		return -1
	}
	unmarked, fix, marked := at("implementer", ""), at("panel-fix", ""), at("implementer", "No-progress escalation")
	if got[unmarked].Override == nil {
		t.Errorf("entry %d = %+v, want implementer's unmarked row carrying the override", unmarked, got[unmarked])
	}
	if got[marked].Override != nil {
		t.Errorf("entry %d = %+v, want implementer's marked row without an override", marked, got[marked])
	}
	if !(unmarked < fix && fix < marked) {
		t.Errorf("implementer %d, panel-fix %d, marked implementer %d: want table order", unmarked, fix, marked)
	}
}
