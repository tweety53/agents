package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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
	statsID   int64
	pairs     map[string]store.RolePolicy
	grants    map[string][]string
}

func (f *fakeExperiments) RolePairs(_ context.Context) (map[string]store.RolePolicy, error) {
	if f.pairs == nil {
		f.pairs = map[string]store.RolePolicy{}
	}
	return f.pairs, nil
}

func (f *fakeExperiments) ReplaceRolePairs(_ context.Context, role string, base, optin []string) error {
	for _, p := range append(append([]string{}, base...), optin...) {
		if !store.ValidPair(p) {
			return fmt.Errorf("%w: %q", store.ErrInvalidPair, p)
		}
	}
	if f.pairs == nil {
		f.pairs = map[string]store.RolePolicy{}
	}
	f.pairs[role] = store.RolePolicy{Base: base, Optin: optin, Marked: f.pairs[role].Marked}
	return nil
}

func (f *fakeExperiments) Grant(_ context.Context, change string, roles []string) error {
	if f.grants == nil {
		f.grants = map[string][]string{}
	}
	f.grants[change] = append(f.grants[change], roles...)
	return nil
}

func (f *fakeExperiments) Revoke(_ context.Context, change string, roles []string) error {
	if len(f.grants[change]) == 0 {
		return store.ErrGrantNotFound
	}
	delete(f.grants, change)
	return nil
}

func (f *fakeExperiments) Grants(_ context.Context, change string) ([]string, error) {
	return f.grants[change], nil
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
	if !openOnly {
		return f.list, nil
	}
	var open []store.Experiment
	for _, e := range f.list {
		if e.Status == "open" {
			open = append(open, e)
		}
	}
	return open, nil
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
	table, err := os.ReadFile("../../../skills/flow-contracts/model-policy.md")
	if err != nil {
		t.Fatalf("read model-policy table: %v", err)
	}
	srv, err := api.New(cfg, fs, fs, fs, fs, fs, nil, api.WithExperiments(fe, table))
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

// TestRolesRoute reads GET /api/v1/roles: a leading main entry, then one
// entry per row of the real roles table in table order, its pairs from the
// store and its group and Does from the table, an open experiment's cell on
// its role and on main, never read off a closed experiment.
func TestRolesRoute(t *testing.T) {
	type override struct {
		Pairs        []string `json:"pairs"`
		ExperimentID int64    `json:"experimentId"`
	}
	type entry struct {
		Role     string             `json:"role"`
		Base     []string           `json:"base"`
		Optin    []string           `json:"optin"`
		Marked   []store.MarkedPair `json:"marked"`
		Group    string             `json:"group"`
		Does     string             `json:"does"`
		Override *override          `json:"override"`
	}
	fe := &fakeExperiments{list: []store.Experiment{{ID: 5, Name: "x", Status: "open", Cells: map[string][]string{
		"plan-designer": {"opus/Plan@xhigh"},
		"main":          {"opus@xhigh"},
	}}}, pairs: map[string]store.RolePolicy{
		"explorer":  {Base: []string{"opus/Explore@low"}, Optin: []string{"opus/Explore@high"}, Marked: []store.MarkedPair{}},
		"panel-fix": {Base: []string{"opus/flow-low"}, Optin: []string{}, Marked: []store.MarkedPair{{Pair: "opus/flow-high", Marker: "No-progress escalation"}}},
	}}
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
	if got[0].Role != "main" || got[0].Group != "Session" || !reflect.DeepEqual(got[0].Override, &override{Pairs: []string{"opus@xhigh"}, ExperimentID: 5}) {
		t.Errorf("entry 0 = %+v, want main carrying experiment 5's override", got[0])
	}
	want := entry{Role: "explorer", Base: []string{"opus/Explore@low"}, Optin: []string{"opus/Explore@high"}, Marked: []store.MarkedPair{},
		Group: "Planning", Does: "Reads the codebase to answer a planner's question."}
	if !reflect.DeepEqual(got[1], want) {
		t.Errorf("entry 1 = %+v, want %+v", got[1], want)
	}
	if got[2].Role != "plan-designer" || !reflect.DeepEqual(got[2].Override, &override{Pairs: []string{"opus/Plan@xhigh"}, ExperimentID: 5}) || got[2].Base == nil {
		t.Errorf("entry 2 = %+v, want plan-designer carrying experiment 5's override and an empty base list", got[2])
	}
	seen := map[string]bool{}
	for i, e := range got {
		if e.Group == "" || seen[e.Role] {
			t.Errorf("entry %d = %+v, want a group and one entry per role", i, e)
		}
		seen[e.Role] = true
		if e.Role == "panel-fix" && !reflect.DeepEqual(e.Marked, fe.pairs["panel-fix"].Marked) {
			t.Errorf("panel-fix = %+v, want its marked escalation pair", e)
		}
	}
}

// TestPolicyPutRoundTrip: PUT replaces a role's base and opt-in pairs and
// answers with the whole policy, which GET then serves; a malformed pair is
// 400 and writes nothing.
func TestPolicyPutRoundTrip(t *testing.T) {
	fe := &fakeExperiments{}
	ts := newExperimentsTestServer(t, fe)
	status, body := doJSON(t, http.MethodPut, ts.URL+"/api/v1/policy/roles/planner", `{"base":["opus/flow-planner-low"],"optin":["fable/flow-planner-xhigh"]}`)
	if status != http.StatusOK || !strings.Contains(body, `"planner":{"base":["opus/flow-planner-low"],"optin":["fable/flow-planner-xhigh"]`) {
		t.Fatalf("PUT planner = %d %s", status, body)
	}
	if status, got := doJSON(t, http.MethodGet, ts.URL+"/api/v1/policy", ""); status != http.StatusOK || got != body {
		t.Fatalf("GET policy = %d %s, want %s", status, got, body)
	}
	if status, body = doJSON(t, http.MethodPut, ts.URL+"/api/v1/policy/roles/planner", `{"base":["opus"]}`); status != http.StatusBadRequest {
		t.Fatalf("PUT malformed = %d %s, want 400", status, body)
	}
	if !reflect.DeepEqual(fe.pairs["planner"].Base, []string{"opus/flow-planner-low"}) {
		t.Fatalf("a refused PUT changed planner: %+v", fe.pairs["planner"])
	}
}

// TestPolicyPutUnknownRole: a role the roles table lacks is 400, naming it.
func TestPolicyPutUnknownRole(t *testing.T) {
	fe := &fakeExperiments{}
	ts := newExperimentsTestServer(t, fe)
	status, body := doJSON(t, http.MethodPut, ts.URL+"/api/v1/policy/roles/plannr", `{"base":["opus/flow-low"]}`)
	if status != http.StatusBadRequest || !strings.Contains(body, "plannr") || fe.pairs["plannr"].Base != nil {
		t.Fatalf("PUT unknown role = %d %s, want 400 naming it and no write", status, body)
	}
}

// TestPolicyGrantEndpoints: grant, read back, revoke; a revoke of a grant
// never made is 404, a grant naming no change or an unknown role 400.
func TestPolicyGrantEndpoints(t *testing.T) {
	fe := &fakeExperiments{}
	ts := newExperimentsTestServer(t, fe)
	g := ts.URL + "/api/v1/policy/grants"
	if status, body := doJSON(t, http.MethodPut, g, `{"change":"kan-1","roles":["planner","plan-designer"]}`); status != http.StatusOK {
		t.Fatalf("PUT grant = %d %s", status, body)
	}
	if status, body := doJSON(t, http.MethodGet, g+"?change=kan-1", ""); status != http.StatusOK || body != `{"change":"kan-1","roles":["planner","plan-designer"]}`+"\n" {
		t.Fatalf("GET grants = %d %q", status, body)
	}
	if status, body := doJSON(t, http.MethodGet, g+"?change=kan-2", ""); status != http.StatusOK || !strings.Contains(body, `"roles":[]`) {
		t.Fatalf("GET grants of kan-2 = %d %s, want an empty list", status, body)
	}
	if status, body := doJSON(t, http.MethodGet, g, ""); status != http.StatusBadRequest {
		t.Fatalf("GET grants with no change = %d %s, want 400", status, body)
	}
	if status, body := doJSON(t, http.MethodDelete, g, `{"change":"kan-1","roles":["planner"]}`); status != http.StatusOK {
		t.Fatalf("DELETE grant = %d %s", status, body)
	}
	if status, body := doJSON(t, http.MethodDelete, g, `{"change":"kan-1","roles":["planner"]}`); status != http.StatusNotFound {
		t.Fatalf("DELETE grant twice = %d %s, want 404", status, body)
	}
	if status, body := doJSON(t, http.MethodPut, g, `{"change":"","roles":["planner"]}`); status != http.StatusBadRequest {
		t.Fatalf("PUT grant with no change = %d %s, want 400", status, body)
	}
	if status, body := doJSON(t, http.MethodPut, g, `{"change":"kan-1","roles":["nobody"]}`); status != http.StatusBadRequest || !strings.Contains(body, "nobody") {
		t.Fatalf("PUT grant unknown role = %d %s, want 400", status, body)
	}
}

// TestPromoteWritesBaseAndCloses: promote writes each cell but main into its
// role's base pairs, keeps the role's opt-in pairs, and closes the
// experiment positive; a closed or unknown id is 404 and writes nothing.
func TestPromoteWritesBaseAndCloses(t *testing.T) {
	fe := &fakeExperiments{
		list: []store.Experiment{{ID: 4, Name: "fable-planner", Status: "open", Cells: map[string][]string{
			"planner": {"fable/flow-planner-high"}, "main": {"opus@high"},
		}}, {ID: 7, Name: "closed-one", Status: "closed", Cells: map[string][]string{"planner": {"opus/flow-planner-high"}}}},
		pairs: map[string]store.RolePolicy{"planner": {Base: []string{"opus/flow-planner-low"}, Optin: []string{"fable/flow-planner-xhigh"}}},
	}
	ts := newExperimentsTestServer(t, fe)
	if status, body := doJSON(t, http.MethodPost, ts.URL+"/api/v1/experiments/9/promote", ""); status != http.StatusNotFound {
		t.Fatalf("promote unknown = %d %s, want 404", status, body)
	}
	if status, body := doJSON(t, http.MethodPost, ts.URL+"/api/v1/experiments/7/promote", ""); status != http.StatusNotFound || fe.closedID != 0 || fe.pairs["planner"].Base[0] != "opus/flow-planner-low" {
		t.Fatalf("promote closed = %d %s (closed %d, planner %+v), want 404 writing nothing", status, body, fe.closedID, fe.pairs["planner"])
	}
	if status, body := doJSON(t, http.MethodPost, ts.URL+"/api/v1/experiments/4/promote", ""); status != http.StatusOK {
		t.Fatalf("promote = %d %s", status, body)
	}
	want := store.RolePolicy{Base: []string{"fable/flow-planner-high"}, Optin: []string{"fable/flow-planner-xhigh"}}
	if !reflect.DeepEqual(fe.pairs["planner"], want) {
		t.Errorf("planner = %+v, want %+v", fe.pairs["planner"], want)
	}
	if _, ok := fe.pairs["main"]; ok {
		t.Error("promote wrote a main role")
	}
	if fe.closedID != 4 || fe.closedArg[0] != "positive" || !strings.Contains(fe.closedArg[1], "planner") {
		t.Errorf("closed %d %q, want 4 positive with a summary naming planner", fe.closedID, fe.closedArg)
	}
	fe.list[0].ID = 6
	if status, body := doJSON(t, http.MethodPost, ts.URL+"/api/v1/experiments/6/promote", `{"summary":"Positive: fewer rounds."}`); status != http.StatusOK || fe.closedArg[1] != "Positive: fewer rounds." {
		t.Fatalf("promote with a summary = %d %s, closed %q", status, body, fe.closedArg)
	}
}
