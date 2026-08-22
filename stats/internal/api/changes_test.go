package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// fakeStore implements api.ChangeStore entirely in memory. These tests
// need no PostgreSQL at all -- exactly the payoff of defining the store
// dependency as an interface at the consumer package (go-interface-design)
// rather than depending on *store.Store directly.
type fakeStore struct {
	changes map[string]store.Change

	getErr   error
	putErr   error
	queryErr error

	// getDelay, when set, runs synchronously inside GetChange before it
	// returns -- used to hold a request in flight for the shutdown test.
	getDelay func()

	// filterFunc, when set, lets a test supply real filtering behaviour
	// for QueryChanges instead of a canned result, so a test can assert
	// the handler built the right store.Query rather than merely that it
	// forwarded some Query and displayed whatever the fake was told to
	// return.
	filterFunc func(store.Query) []store.Change

	lastQuery store.Query

	// --- stage-mark bookkeeping (internal/api/stages_test.go's fakeStore
	// methods operate on these) ---
	stageRuns      []stageRunRecord
	nextStageRunID int64

	beginStageErr     error
	endStageErr       error
	mergeMetricsErr   error
	queryStageRunsErr error

	// --- run-record bookkeeping (internal/api/records_test.go's fakeStore
	// methods operate on these) ---
	dispatches     []dispatchRecord
	findings       []findingRecord
	nextDispatchID int64

	recordDispatchErr   error
	upsertFindingErr    error
	setFindingStatusErr error
	runRecordErr        error

	// recordCalls counts every call that reached one of the record
	// methods, so a test can assert that a request rejected on its way in
	// -- a body that did not decode -- never reached the store at all,
	// rather than merely that it was answered with a 400.
	recordCalls int

	// --- statistics bookkeeping (internal/api/stats_test.go's fakeStore
	// methods operate on these) ---
	liveStateBoard      []store.LiveStateRow
	liveStateBoardErr   error
	costPerChange       []store.CostPerChangeRow
	costPerChangeErr    error
	stageLeaderboard    []store.StageLeaderboardRow
	stageLeaderboardErr error
	trendOverTime       []store.TrendPoint
	trendOverTimeErr    error
	cacheEfficiency     []store.CacheEfficiencyRow
	cacheEfficiencyErr  error
	panelEconomics      []store.PanelEconomicsRow
	panelEconomicsErr   error
	modelComparison     []store.ModelComparisonRow
	modelComparisonErr  error
	reworkRate          []store.ReworkRateRow
	reworkRateErr       error

	// lastStatsProject records the project pointer passed to whichever
	// aggregation method a stats test just called, so a test can assert
	// the handler forwarded the project filter it parsed from the request
	// (TestProjectFilterRestrictsResults / TestNoProjectFilterAggregatesAcrossProjects)
	// without needing real per-project filtering logic in this fake.
	lastStatsProject *string

	// --- project display-name resolution (task 3's resolveProjectParam,
	// stats.go) --- projectKeysByDisplayName maps a display name to the
	// project keys that resolve to it, so a test can seed a unique match,
	// an ambiguous one (two-plus keys), or leave a name absent for the
	// "unknown project" case. projectKeysByDisplayNameCalls records every
	// display name actually queried, so a test can assert that an exact
	// key -- one already carrying the derivation suffix -- never reaches
	// this method at all (spec: "no resolution query is run").
	projectKeysByDisplayName      map[string][]string
	projectKeysByDisplayNameErr   error
	projectKeysByDisplayNameCalls []string
}

func (f *fakeStore) ProjectKeysByDisplayName(_ context.Context, displayName string) ([]string, error) {
	f.projectKeysByDisplayNameCalls = append(f.projectKeysByDisplayNameCalls, displayName)
	if f.projectKeysByDisplayNameErr != nil {
		return nil, f.projectKeysByDisplayNameErr
	}
	return f.projectKeysByDisplayName[displayName], nil
}

func newFakeStore() *fakeStore {
	return &fakeStore{changes: map[string]store.Change{}}
}

func changeKey(project, name string) string { return project + "/" + name }

func (f *fakeStore) GetChange(_ context.Context, projectKey, name string) (store.Change, error) {
	if f.getDelay != nil {
		f.getDelay()
	}
	if f.getErr != nil {
		return store.Change{}, f.getErr
	}
	c, ok := f.changes[changeKey(projectKey, name)]
	if !ok {
		return store.Change{}, fmt.Errorf("%w: %s/%s", store.ErrChangeNotFound, projectKey, name)
	}
	return c, nil
}

func (f *fakeStore) PutChange(_ context.Context, c store.Change) error {
	if f.putErr != nil {
		return f.putErr
	}
	f.changes[changeKey(c.ProjectKey, c.Name)] = c
	return nil
}

func (f *fakeStore) QueryChanges(_ context.Context, q store.Query) ([]store.Change, int, error) {
	f.lastQuery = q
	if f.queryErr != nil {
		return nil, 0, f.queryErr
	}
	if f.filterFunc != nil {
		result := f.filterFunc(q)
		return result, len(result), nil
	}
	var all []store.Change
	for _, c := range f.changes {
		all = append(all, c)
	}
	return all, len(all), nil
}

var _ api.ChangeStore = (*fakeStore)(nil)

// --- test helpers ---------------------------------------------------------

func newTestServer(t *testing.T, fs *fakeStore) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, fs, fs, fs, fs, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func doGet(t *testing.T, ts *httptest.Server, path string) (int, string) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(b)
}

func doPut(t *testing.T, ts *httptest.Server, path string, body []byte) (int, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, ts.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build PUT %s: %v", path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(b)
}

func changeBody(t *testing.T, state string) []byte {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"state":     state,
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"updatedBy": "tester",
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return body
}

type listChangesResponseTest struct {
	Total   int `json:"total"`
	Changes []struct {
		ProjectKey string `json:"projectKey"`
		Name       string `json:"name"`
		State      string `json:"state"`
	} `json:"changes"`
}

func matchesFilters(c store.Change, filters []store.Filter) bool {
	for _, f := range filters {
		var got string
		switch f.Field {
		case "project":
			got = c.ProjectKey
		case "name":
			got = c.Name
		case "state":
			got = string(c.State)
		default:
			return false
		}
		want, _ := f.Value.(string)
		switch f.Op {
		case store.OpEq:
			if got != want {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// --- tests -----------------------------------------------------------------

// TestServerBindsLoopbackOnly asserts api.New refuses a non-loopback host
// before ever opening a listener, and accepts a loopback one. Mutation
// check performed by hand: with the cfg.Validate() call removed from
// api.New, the "non-loopback host is refused" subtest fails (err == nil),
// confirming the test actually exercises the guard rather than passing
// vacuously.
func TestServerBindsLoopbackOnly(t *testing.T) {
	t.Run("non-loopback host is refused", func(t *testing.T) {
		cfg := config.Config{Host: "0.0.0.0", Port: 0, DSN: "unused"}
		srv, err := api.New(cfg, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil)
		if err == nil {
			t.Fatal("expected an error for a non-loopback host, got nil")
		}
		if !errors.Is(err, config.ErrNonLoopbackHost) {
			t.Fatalf("expected config.ErrNonLoopbackHost, got %v", err)
		}
		if srv != nil {
			t.Fatal("expected a nil server when configuration is refused")
		}
	})

	t.Run("a hostname is refused, not resolved", func(t *testing.T) {
		// "localhost" usually resolves to a loopback address, but
		// Validate must answer on the literal string alone -- resolving
		// it would make the answer depend on this machine's own
		// name-resolution configuration.
		cfg := config.Config{Host: "localhost", Port: 0, DSN: "unused"}
		if _, err := api.New(cfg, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil); err == nil {
			t.Fatal("expected an error for a hostname, got nil")
		}
	})

	t.Run("loopback host is accepted", func(t *testing.T) {
		cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
		srv, err := api.New(cfg, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil)
		if err != nil {
			t.Fatalf("unexpected error for a loopback host: %v", err)
		}
		if srv == nil {
			t.Fatal("expected a non-nil server")
		}
	})
}

// TestPutChangeReturnsMonotonicRefusalDistinctly asserts a monotonic
// refusal from the store comes back as a status distinguishable from both
// a client mistake (400) and a store-side failure (500/503) -- the
// distinction task 5's CLI depends on to decide between "report the
// refusal" and "take the journal fallback". Mutation check performed by
// hand: collapsing mapStoreError's ErrMonotonicViolation case into the
// default 500 branch makes this test fail, confirming it actually pins
// the status rather than merely accepting whatever the handler returns.
func TestPutChangeReturnsMonotonicRefusalDistinctly(t *testing.T) {
	fs := newFakeStore()
	fs.putErr = store.ErrMonotonicViolation
	ts := newTestServer(t, fs)

	status, body := doPut(t, ts, "/api/v1/changes/proj/chg", changeBody(t, "IN_PROGRESS"))

	if status != http.StatusConflict {
		t.Fatalf("monotonic refusal: got status %d, want %d", status, http.StatusConflict)
	}
	if status == http.StatusInternalServerError || status == http.StatusServiceUnavailable {
		t.Fatalf("monotonic refusal must not share a status class with a store-unreachable failure, got %d", status)
	}
	if !strings.Contains(body, "refused") && !strings.Contains(body, "backwards") {
		t.Fatalf("expected the monotonic refusal message in the body, got %q", body)
	}
}

// TestGetChangeUnknownReturns404 asserts an unknown change is reported as
// 404, and that the same handler returns 200 for a change that does
// exist -- so this cannot pass merely because the handler returns 404 for
// every request.
func TestGetChangeUnknownReturns404(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	status, body := doGet(t, ts, "/api/v1/changes/proj/does-not-exist")
	if status != http.StatusNotFound {
		t.Fatalf("got status %d, want 404", status)
	}
	if !strings.Contains(body, "not found") {
		t.Fatalf("expected a not-found message in the body, got %q", body)
	}

	fs.changes[changeKey("proj", "exists")] = store.Change{
		ProjectKey: "proj", Name: "exists", State: store.StateStarted,
		UpdatedAt: time.Now(), UpdatedBy: "tester",
	}
	status, _ = doGet(t, ts, "/api/v1/changes/proj/exists")
	if status != http.StatusOK {
		t.Fatalf("existing change: got status %d, want 200", status)
	}
}

// TestPutChangeRejectsUnknownField asserts a PUT body naming a field
// outside changeDTO's vocabulary is refused with 400 and a message naming
// the field -- the request-body half of the state file contract's
// closed-schema rule, enforced by dec.DisallowUnknownFields() in put().
// Mutation check performed by hand: removing that call from put() makes
// this test fail (the request instead succeeds, PutChange is called, and
// the unknown field is silently dropped), confirming the test actually
// exercises the guard rather than passing regardless of it.
func TestPutChangeRejectsUnknownField(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	body := []byte(`{"state":"STARTED","updatedAt":"2026-08-13T00:00:00Z","updatedBy":"tester","bogusField":"x"}`)
	status, respBody := doPut(t, ts, "/api/v1/changes/proj/chg", body)

	if status != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", status)
	}
	if !strings.Contains(respBody, "bogusField") {
		t.Fatalf("expected the rejected field named in the body, got %q", respBody)
	}
	if strings.Contains(respBody, "invalid character") {
		t.Fatalf("unknown-field response should not read like a JSON-syntax rejection, got %q", respBody)
	}
	if _, stored := fs.changes[changeKey("proj", "chg")]; stored {
		t.Fatal("a rejected body must not reach PutChange")
	}
}

// TestPutChangeRejectsMalformedJSON asserts a body that is not valid JSON
// at all is refused with 400 and a JSON-syntax-specific message -- "invalid
// character", what encoding/json's own SyntaxError says -- rather than the
// unknown-field test's "bogusField". Both tests share the handler's
// "malformed request body: " prefix, so asserting only that prefix would
// make the two interchangeable and prove neither is exercising its own
// code path; asserting the decoder's actual, distinct error text is what
// tells them apart.
func TestPutChangeRejectsMalformedJSON(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	status, respBody := doPut(t, ts, "/api/v1/changes/proj/chg", []byte(`{not json`))

	if status != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", status)
	}
	if !strings.Contains(respBody, "invalid character") {
		t.Fatalf("expected a JSON-syntax-specific message, got %q", respBody)
	}
	if strings.Contains(respBody, "bogusField") || strings.Contains(respBody, "unknown field") {
		t.Fatalf("malformed-JSON response should not read like an unknown-field rejection, got %q", respBody)
	}
	if _, stored := fs.changes[changeKey("proj", "chg")]; stored {
		t.Fatal("a rejected body must not reach PutChange")
	}
}

// TestListChangesFiltersByProjectAndState asserts the handler both builds
// the right store.Query from the request (asserted directly against
// fs.lastQuery) and returns only what the store's own filtering produced
// (asserted against the response body) -- either check alone could pass
// for the wrong reason: a fake that ignores the query but happens to
// return one row, or a Query built correctly but never actually applied.
func TestListChangesFiltersByProjectAndState(t *testing.T) {
	fs := newFakeStore()
	all := []store.Change{
		{ProjectKey: "proj-a", Name: "one", State: store.StateStarted},
		{ProjectKey: "proj-a", Name: "two", State: store.StateFinished},
		{ProjectKey: "proj-b", Name: "three", State: store.StateStarted},
	}
	fs.filterFunc = func(q store.Query) []store.Change {
		var out []store.Change
		for _, c := range all {
			if matchesFilters(c, q.Filters) {
				out = append(out, c)
			}
		}
		return out
	}

	ts := newTestServer(t, fs)
	status, body := doGet(t, ts, "/api/v1/changes?project=proj-a&state=STARTED")
	if status != http.StatusOK {
		t.Fatalf("got status %d, want 200", status)
	}

	var resp listChangesResponseTest
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, body)
	}
	if len(resp.Changes) != 1 || resp.Changes[0].Name != "one" {
		t.Fatalf("expected exactly change %q, got %+v", "one", resp.Changes)
	}

	wantFields := map[string]string{"project": "proj-a", "state": "STARTED"}
	if len(fs.lastQuery.Filters) != len(wantFields) {
		t.Fatalf("expected %d filters, got %d: %+v", len(wantFields), len(fs.lastQuery.Filters), fs.lastQuery.Filters)
	}
	for _, f := range fs.lastQuery.Filters {
		want, ok := wantFields[f.Field]
		if !ok || f.Value != want {
			t.Fatalf("unexpected filter %+v", f)
		}
	}
}

// TestListChangesAcceptsProjectDisplayName asserts the changes list's own
// "project" filter resolves a display name matching exactly one project to
// that project's key before the store ever sees the filter -- the same
// resolution stats views get, applied here at the list endpoint's own
// parse site (specs/myflow-stats-views/spec.md, "A display name is
// filtered on").
func TestListChangesAcceptsProjectDisplayName(t *testing.T) {
	fs := newFakeStore()
	fs.projectKeysByDisplayName = map[string][]string{"agents": {"agents-a740d89c"}}
	all := []store.Change{
		{ProjectKey: "agents-a740d89c", Name: "kan-1", State: store.StateStarted},
		{ProjectKey: "other-7c1f238a", Name: "kan-2", State: store.StateStarted},
	}
	fs.filterFunc = func(q store.Query) []store.Change {
		var out []store.Change
		for _, c := range all {
			if matchesFilters(c, q.Filters) {
				out = append(out, c)
			}
		}
		return out
	}

	ts := newTestServer(t, fs)
	status, body := doGet(t, ts, "/api/v1/changes?project=agents")
	if status != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", status, body)
	}

	var resp listChangesResponseTest
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("decode response: %v; body=%s", err, body)
	}
	if len(resp.Changes) != 1 || resp.Changes[0].ProjectKey != "agents-a740d89c" {
		t.Fatalf("expected exactly the change from agents-a740d89c, got %+v", resp.Changes)
	}

	for _, f := range fs.lastQuery.Filters {
		if f.Field == "project" && f.Value != "agents-a740d89c" {
			t.Fatalf("store received project filter %v, want resolved key %q", f.Value, "agents-a740d89c")
		}
	}
}

// TestListChangesRejectsAmbiguousProjectDisplayName asserts a display name
// matching more than one project is refused with 400, naming the
// ambiguity and listing the candidate keys, rather than silently filtering
// on one of them.
func TestListChangesRejectsAmbiguousProjectDisplayName(t *testing.T) {
	fs := newFakeStore()
	fs.projectKeysByDisplayName = map[string][]string{
		"agents": {"agents-a740d89c", "agents-7c1f238a"},
	}
	ts := newTestServer(t, fs)

	status, body := doGet(t, ts, "/api/v1/changes?project=agents")
	if status != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400, body=%s", status, body)
	}
	if !strings.Contains(body, "agents-a740d89c") || !strings.Contains(body, "agents-7c1f238a") {
		t.Fatalf("expected both candidate keys named in the body, got %q", body)
	}
}

// TestListChangesProjectResolutionStoreFailureReturns500 asserts that a
// store failure inside ProjectKeysByDisplayName -- as opposed to a genuine
// client mistake like an ambiguous display name -- is reported as a
// logged 500 with a generic body, exactly like any other store failure in
// this handler, rather than the 400 an ordinary parse failure gets. Before
// the fix, resolveProjectParam's propagated store error was
// indistinguishable from a parse error at this call site and was reported
// as 400 with the store's own internal text leaking into the body
// (post-commit review round 2, F4).
func TestListChangesProjectResolutionStoreFailureReturns500(t *testing.T) {
	fs := newFakeStore()
	fs.projectKeysByDisplayNameErr = errors.New("store: project keys by display name \"agents\": connection refused")
	ts := newTestServer(t, fs)

	status, body := doGet(t, ts, "/api/v1/changes?project=agents")
	if status != http.StatusInternalServerError {
		t.Fatalf("got status %d, want 500, body=%s", status, body)
	}
	if strings.Contains(body, "connection refused") {
		t.Fatalf("response body leaked internal store text: %s", body)
	}
}

// TestListChangesExactProjectKeyRunsNoResolutionQuery asserts a "project"
// value already carrying the derivation suffix is used unchanged and never
// reaches ProjectKeysByDisplayName at all -- specs/myflow-stats-views/spec.md's
// "A full key is filtered on": "it is used as-is, with no resolution
// attempted".
func TestListChangesExactProjectKeyRunsNoResolutionQuery(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	status, body := doGet(t, ts, "/api/v1/changes?project=agents-a740d89c")
	if status != http.StatusOK {
		t.Fatalf("got status %d, want 200, body=%s", status, body)
	}
	if len(fs.projectKeysByDisplayNameCalls) != 0 {
		t.Fatalf("ProjectKeysByDisplayName called with %v, want no calls for an exact key", fs.projectKeysByDisplayNameCalls)
	}
	for _, f := range fs.lastQuery.Filters {
		if f.Field == "project" && f.Value != "agents-a740d89c" {
			t.Fatalf("store received project filter %v, want unchanged %q", f.Value, "agents-a740d89c")
		}
	}
}

// TestListChangesUnknownFieldReturns400 asserts a query field the store's
// allowlist rejects surfaces as 400, naming the field -- never a 500 and
// never a silently ignored parameter.
func TestListChangesUnknownFieldReturns400(t *testing.T) {
	fs := newFakeStore()
	fs.queryErr = fmt.Errorf("%w: filter %q not recognised; accepted: name, project, state", store.ErrUnknownField, "bogus")
	ts := newTestServer(t, fs)

	status, body := doGet(t, ts, "/api/v1/changes?bogus=x")
	if status != http.StatusBadRequest {
		t.Fatalf("got status %d, want 400", status)
	}
	if !strings.Contains(body, "bogus") {
		t.Fatalf("expected the rejected field named in the body, got %q", body)
	}
}

// TestErrorStatusMapping is the table-driven test over every typed store
// error task 4 names, asserting each maps to its own deliberate status
// rather than folding into a blanket 500 -- including an unrecognised
// error, which must still land on 500 so the table actually distinguishes
// "known and handled" from "unknown and generic".
func TestErrorStatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{"change not found", store.ErrChangeNotFound, http.StatusNotFound},
		{"monotonic violation", store.ErrMonotonicViolation, http.StatusConflict},
		{"invalid state", store.ErrInvalidState, http.StatusBadRequest},
		{"invalid main checkout path", store.ErrInvalidMainCheckoutPath, http.StatusBadRequest},
		{"duplicate repo root", store.ErrDuplicateRepoRoot, http.StatusBadRequest},
		{"unknown field", store.ErrUnknownField, http.StatusBadRequest},
		{"unknown filter op", store.ErrUnknownFilterOp, http.StatusBadRequest},
		{"invalid metrics key path", store.ErrInvalidMetricsKeyPath, http.StatusBadRequest},
		{"offset with no limit", store.ErrOffsetWithNoLimit, http.StatusBadRequest},
		{"tokens unavailable", store.ErrTokensUnavailable, http.StatusUnprocessableEntity},
		{"pricing not found", store.ErrPricingNotFound, http.StatusUnprocessableEntity},
		{"too many attempt collisions", store.ErrTooManyAttemptCollisions, http.StatusServiceUnavailable},
		{"unrecognised error", errors.New("boom"), http.StatusInternalServerError},
	}

	seen := map[int]bool{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := newFakeStore()
			fs.getErr = tc.err
			ts := newTestServer(t, fs)

			status, _ := doGet(t, ts, "/api/v1/changes/proj/chg")
			if status != tc.wantStatus {
				t.Fatalf("got status %d, want %d", status, tc.wantStatus)
			}
			seen[tc.wantStatus] = true
		})
	}

	if len(seen) < 5 {
		t.Fatalf("expected several distinct statuses across the table, saw only %d", len(seen))
	}
}

// TestShutdownDrainsInFlight asserts Shutdown lets an in-flight request
// finish -- rather than cutting it off -- and does not return until it
// has. Mutation check performed by hand: replacing Server.Shutdown's call
// to s.httpServer.Shutdown with s.httpServer.Close (which drops
// connections immediately) makes the in-flight request fail instead of
// observing http.StatusOK, confirming this test actually distinguishes a
// graceful shutdown from an abrupt one.
func TestShutdownDrainsInFlight(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "slow")] = store.Change{
		ProjectKey: "proj", Name: "slow", State: store.StateStarted,
		UpdatedAt: time.Now(), UpdatedBy: "tester",
	}

	started := make(chan struct{})
	release := make(chan struct{})
	fs.getDelay = func() {
		close(started)
		<-release
	}

	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, fs, fs, fs, fs, nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	serveErrCh := make(chan error, 1)
	go func() { serveErrCh <- srv.Serve(ln) }()

	reqDoneCh := make(chan int, 1)
	go func() {
		resp, err := http.Get(fmt.Sprintf("http://%s/api/v1/changes/proj/slow", ln.Addr()))
		if err != nil {
			reqDoneCh <- -1
			return
		}
		defer resp.Body.Close()
		_, _ = io.Copy(io.Discard, resp.Body)
		reqDoneCh <- resp.StatusCode
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request never started")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownDone <- srv.Shutdown(ctx)
	}()

	// Give Shutdown a moment to begin refusing new connections before the
	// in-flight handler is released, so this exercises "drain what is in
	// flight" rather than "shut down after everything already finished".
	time.Sleep(50 * time.Millisecond)
	close(release)

	select {
	case status := <-reqDoneCh:
		if status != http.StatusOK {
			t.Fatalf("in-flight request did not complete successfully, got status %d", status)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight request never completed")
	}

	if err := <-shutdownDone; err != nil {
		t.Fatalf("Shutdown returned an error: %v", err)
	}
	if err := <-serveErrCh; err != nil {
		t.Fatalf("Serve returned an error: %v", err)
	}
}

// --- F1: the daemon header identifies genuine responses ---

// TestServerSetsDaemonHeaderOnEveryResponse asserts api.DaemonHeader is
// present, with api.DaemonHeaderValue, on a success response, a 4xx and a
// 5xx alike -- task 5's CLI trusts a 409 or 404 as a genuine store answer
// only when this header is present, so a response missing it on any path
// through this handler would silently reopen the gap F1 closed.
//
// Mutation check performed by hand: with the withDaemonHeader wrapper
// removed from api.New's Handler (using mux directly), every subtest here
// fails with an empty header value.
func TestServerSetsDaemonHeaderOnEveryResponse(t *testing.T) {
	t.Run("200 on a known change", func(t *testing.T) {
		fs := newFakeStore()
		fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateInProgress}
		ts := newTestServer(t, fs)

		resp, err := http.Get(ts.URL + "/api/v1/changes/proj/chg")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
			t.Errorf("%s = %q, want %q", api.DaemonHeader, got, api.DaemonHeaderValue)
		}
	})

	t.Run("404 on an unknown change", func(t *testing.T) {
		fs := newFakeStore()
		ts := newTestServer(t, fs)

		resp, err := http.Get(ts.URL + "/api/v1/changes/proj/absent")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", resp.StatusCode)
		}
		if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
			t.Errorf("%s = %q, want %q", api.DaemonHeader, got, api.DaemonHeaderValue)
		}
	})

	t.Run("409 on a monotonic refusal", func(t *testing.T) {
		fs := newFakeStore()
		fs.putErr = store.ErrMonotonicViolation
		ts := newTestServer(t, fs)

		status, _ := doPut(t, ts, "/api/v1/changes/proj/chg", changeBody(t, "IN_PROGRESS"))
		if status != http.StatusConflict {
			t.Fatalf("status = %d, want 409", status)
		}

		req, err := http.NewRequest(http.MethodPut, ts.URL+"/api/v1/changes/proj/chg", bytes.NewReader(changeBody(t, "IN_PROGRESS")))
		if err != nil {
			t.Fatalf("build request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("PUT: %v", err)
		}
		defer resp.Body.Close()
		if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
			t.Errorf("%s = %q, want %q", api.DaemonHeader, got, api.DaemonHeaderValue)
		}
	})

	t.Run("500 on an unmapped store error", func(t *testing.T) {
		fs := newFakeStore()
		fs.getErr = fmt.Errorf("boom")
		ts := newTestServer(t, fs)

		resp, err := http.Get(ts.URL + "/api/v1/changes/proj/chg")
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", resp.StatusCode)
		}
		if got := resp.Header.Get(api.DaemonHeader); got != api.DaemonHeaderValue {
			t.Errorf("%s = %q, want %q", api.DaemonHeader, got, api.DaemonHeaderValue)
		}
	})
}

// --- F5: Repos is derived from worktrees, not read from the request body ---

// TestPutChangeDerivesRepoSetFromWorktrees asserts a worktrees map with two
// entries produces two store.Repo entries on the Change PutChange is
// called with, keyed by worktree path with the merge base carried through
// -- and that a null merge base survives as a nil *string, per the state
// file contract's "null is legal and means no merge base recorded".
//
// Mutation check performed by hand: with the `c.Repos, err =
// reposFromWorktrees(...)` call removed from put, this test fails with
// zero repos recorded (the pre-fix behaviour: an explicit "repos" field
// was needed on the wire, which task 5's CLI never sends).
func TestPutChangeDerivesRepoSetFromWorktrees(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	body, err := json.Marshal(map[string]any{
		"state": "IN_PROGRESS",
		"worktrees": map[string]any{
			"/Users/tester/Projects/agents-worktrees/openspec-kan-16": "5ee4c9a",
			"/Users/tester/Projects/other-worktrees/openspec-kan-16":  nil,
		},
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"updatedBy": "tester",
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	status, respBody := doPut(t, ts, "/api/v1/changes/proj/kan-16", body)
	if status != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200: %s", status, respBody)
	}

	stored, ok := fs.changes[changeKey("proj", "kan-16")]
	if !ok {
		t.Fatal("change was not stored")
	}
	if len(stored.Repos) != 2 {
		t.Fatalf("len(stored.Repos) = %d, want 2: %+v", len(stored.Repos), stored.Repos)
	}

	byRoot := map[string]*string{}
	for _, r := range stored.Repos {
		byRoot[r.RepoRoot] = r.MergeBase
	}
	mb, ok := byRoot["/Users/tester/Projects/agents-worktrees/openspec-kan-16"]
	if !ok || mb == nil || *mb != "5ee4c9a" {
		t.Errorf("repo with merge base missing or wrong: %v", byRoot["/Users/tester/Projects/agents-worktrees/openspec-kan-16"])
	}
	mb2, ok := byRoot["/Users/tester/Projects/other-worktrees/openspec-kan-16"]
	if !ok || mb2 != nil {
		t.Errorf("repo with null merge base should round-trip as nil, got %v (present=%v)", mb2, ok)
	}
}

// TestPutChangeEmptyWorktreesDerivesNoRepos asserts a change confined to
// its own project's repository -- no worktrees entries -- derives an empty
// repo set rather than an error.
func TestPutChangeEmptyWorktreesDerivesNoRepos(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	status, respBody := doPut(t, ts, "/api/v1/changes/proj/kan-16", changeBody(t, "IN_PROGRESS"))
	if status != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200: %s", status, respBody)
	}

	stored := fs.changes[changeKey("proj", "kan-16")]
	if len(stored.Repos) != 0 {
		t.Errorf("len(stored.Repos) = %d, want 0", len(stored.Repos))
	}
}

// TestPutChangeMalformedWorktreesIsRejected asserts a worktrees value that
// is not an object -- not this endpoint's own JSON decoder's job to catch,
// since worktrees is stored as json.RawMessage and passed through
// untouched until reposFromWorktrees parses it -- is reported as 400, not
// silently ignored or turned into a 500.
func TestPutChangeMalformedWorktreesIsRejected(t *testing.T) {
	fs := newFakeStore()
	ts := newTestServer(t, fs)

	body, err := json.Marshal(map[string]any{
		"state":     "IN_PROGRESS",
		"worktrees": []string{"not", "an", "object"},
		"updatedAt": time.Now().UTC().Format(time.RFC3339),
		"updatedBy": "tester",
	})
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}

	status, _ := doPut(t, ts, "/api/v1/changes/proj/kan-16", body)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", status)
	}
	if _, ok := fs.changes[changeKey("proj", "kan-16")]; ok {
		t.Error("a malformed worktrees value must not reach the store")
	}
}
