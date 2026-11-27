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

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// fakePolicy is an in-memory api.PolicyStore: it validates like the real
// store and answers with whatever the test configured.
type fakePolicy struct {
	pairs  map[string]store.RolePolicy
	grants map[string][]string
}

func (f *fakePolicy) RolePairs(_ context.Context) (map[string]store.RolePolicy, error) {
	if f.pairs == nil {
		f.pairs = map[string]store.RolePolicy{}
	}
	return f.pairs, nil
}

func (f *fakePolicy) ReplaceRolePairs(_ context.Context, role string, base, optin []string) error {
	for _, p := range append(append([]string{}, base...), optin...) {
		if !store.ValidPair(p) {
			return fmt.Errorf("%w: %q", store.ErrInvalidPair, p)
		}
	}
	if f.pairs == nil {
		f.pairs = map[string]store.RolePolicy{}
	}
	f.pairs[role] = store.RolePolicy{Base: base, Optin: optin}
	return nil
}

func (f *fakePolicy) Grant(_ context.Context, change string, roles []string) error {
	if f.grants == nil {
		f.grants = map[string][]string{}
	}
	f.grants[change] = append(f.grants[change], roles...)
	return nil
}

func (f *fakePolicy) Revoke(_ context.Context, change string, roles []string) error {
	if len(f.grants[change]) == 0 {
		return store.ErrGrantNotFound
	}
	delete(f.grants, change)
	return nil
}

func (f *fakePolicy) Grants(_ context.Context, change string) ([]string, error) {
	return f.grants[change], nil
}

func newPolicyTestServer(t *testing.T, fe *fakePolicy) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	fs := newFakeStore()
	table, err := os.ReadFile("../../../skills/flow-contracts/model-policy.md")
	if err != nil {
		t.Fatalf("read model-policy table: %v", err)
	}
	srv, err := api.New(cfg, fs, fs, fs, fs, nil, api.WithPolicy(fe, table))
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

// TestRolesRoute reads GET /api/v1/roles: a leading main entry, then one
// entry per row of the real roles table in table order, its pairs from the
// store and its group -- a slim pipeline stage -- and Does line from the table.
func TestRolesRoute(t *testing.T) {
	t.Parallel()
	type entry struct {
		Role  string   `json:"role"`
		Base  []string `json:"base"`
		Optin []string `json:"optin"`
		Group string   `json:"group"`
		Does  string   `json:"does"`
	}
	fe := &fakePolicy{pairs: map[string]store.RolePolicy{
		"explorer": {Base: []string{"opus/Explore@low"}, Optin: []string{"opus/Explore@high"}},
	}}
	ts := newPolicyTestServer(t, fe)

	status, body := doJSON(t, http.MethodGet, ts.URL+"/api/v1/roles", "")
	var got []entry
	if status != http.StatusOK || json.Unmarshal([]byte(body), &got) != nil {
		t.Fatalf("GET roles = %d %s, want 200 with a JSON list", status, body)
	}
	if len(got) < 6 {
		t.Fatalf("GET roles = %s, want main plus the table's rows", body)
	}
	if got[0].Role != "main" || got[0].Group != "Session" {
		t.Errorf("entry 0 = %+v, want main in Session", got[0])
	}
	want := entry{Role: "explorer", Base: []string{"opus/Explore@low"}, Optin: []string{"opus/Explore@high"},
		Group: "Planning", Does: "Reads the codebase to answer a planner's question."}
	if !reflect.DeepEqual(got[1], want) {
		t.Errorf("entry 1 = %+v, want %+v", got[1], want)
	}
	if got[2].Role != "plan-designer" || got[2].Base == nil {
		t.Errorf("entry 2 = %+v, want plan-designer with an empty base list", got[2])
	}
	// Each role sits under the slim stage it runs in.
	groups := map[string]bool{}
	for _, e := range got {
		groups[e.Group] = true
		for role, group := range map[string]string{"decider": "Planning", "implementer": "Implementation", "panel-fix": "Review", "minor-fix": "Review", "visual-verifier": "Verification"} {
			if e.Role == role && e.Group != group {
				t.Errorf("%s in %q, want %q", role, e.Group, group)
			}
		}
	}
	if groups["Visual verification"] || !groups["Verification"] {
		t.Errorf("groups = %v, want the slim stages", groups)
	}
	seen := map[string]bool{}
	for i, e := range got {
		if e.Group == "" || seen[e.Role] {
			t.Errorf("entry %d = %+v, want a group and one entry per role", i, e)
		}
		seen[e.Role] = true
	}
}

// TestPolicyPutRoundTrip: PUT replaces a role's base and opt-in pairs and
// answers with the whole policy, which GET then serves; a malformed pair is
// 400 and writes nothing.
func TestPolicyPutRoundTrip(t *testing.T) {
	t.Parallel()
	fe := &fakePolicy{}
	ts := newPolicyTestServer(t, fe)
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

// TestPolicyPostRoles: one edit of two roles writes both roles' pairs and
// answers with the whole policy; a refused edit writes nothing.
func TestPolicyPostRoles(t *testing.T) {
	t.Parallel()
	fe := &fakePolicy{}
	ts := newPolicyTestServer(t, fe)
	status, body := doJSON(t, http.MethodPost, ts.URL+"/api/v1/policy/roles", `{"roles":{"planner":{"base":["opus/flow-planner-low","opus/flow-planner-high"],"optin":[]},"explorer":{"base":["opus/Explore@low"],"optin":[]}}}`)
	if status != http.StatusOK || !strings.Contains(body, `"explorer":{"base":["opus/Explore@low"]`) || !strings.Contains(body, `"planner":{"base":["opus/flow-planner-low","opus/flow-planner-high"]`) {
		t.Fatalf("POST roles = %d %s", status, body)
	}
	if len(fe.pairs["planner"].Base) != 2 || len(fe.pairs["explorer"].Base) != 1 {
		t.Fatalf("pairs = %+v", fe.pairs)
	}
	fe2 := &fakePolicy{}
	ts2 := newPolicyTestServer(t, fe2)
	for _, bad := range []string{`{"roles":{}}`, `{"roles":{"main":{"base":["opus/x"]}}}`, `{"roles":{"explorer":{"base":[]}}}`, `{"roles":{"explorer":{"base":["opus"]}}}`} {
		if status, body := doJSON(t, http.MethodPost, ts2.URL+"/api/v1/policy/roles", bad); status != http.StatusBadRequest {
			t.Fatalf("POST %s = %d %s, want 400", bad, status, body)
		}
	}
	if len(fe2.pairs) != 0 {
		t.Fatalf("a refused edit wrote %+v", fe2.pairs)
	}
	if status, body := doJSON(t, http.MethodGet, ts.URL+"/api/v1/experiments", ""); status != http.StatusNotFound {
		t.Fatalf("GET experiments = %d %s, want 404: the route is gone", status, body)
	}
}

// TestPolicyPutUnknownRole: a role the roles table lacks is 400, naming it.
func TestPolicyPutUnknownRole(t *testing.T) {
	t.Parallel()
	fe := &fakePolicy{}
	ts := newPolicyTestServer(t, fe)
	status, body := doJSON(t, http.MethodPut, ts.URL+"/api/v1/policy/roles/plannr", `{"base":["opus/flow-low"]}`)
	if status != http.StatusBadRequest || !strings.Contains(body, "plannr") || fe.pairs["plannr"].Base != nil {
		t.Fatalf("PUT unknown role = %d %s, want 400 naming it and no write", status, body)
	}
}

// TestPolicyGrantEndpoints: grant, read back, revoke; a revoke of a grant
// never made is 404, a grant naming no change or an unknown role 400.
func TestPolicyGrantEndpoints(t *testing.T) {
	t.Parallel()
	fe := &fakePolicy{}
	ts := newPolicyTestServer(t, fe)
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
