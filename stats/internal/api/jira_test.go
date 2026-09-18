package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
)

// jiraConfigFor points cfg.Jira at atlassian, a scripted stand-in for the
// Atlassian REST API, so this file's tests drive POST
// /api/v1/jira/transition through the real api.New wiring (the real
// internal/jira client included) with no database and no new injection
// surface. jiraClientConfiguredServer builds the daemon around that
// configuration; a nil atlassian builds an unconfigured daemon.
func jiraTestServer(t *testing.T, atlassian http.HandlerFunc) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	if atlassian != nil {
		upstream := httptest.NewServer(atlassian)
		t.Cleanup(upstream.Close)
		cfg.Jira = config.JiraConfig{Site: upstream.URL, Email: "ops@example.com", APIToken: "token"}
	}
	srv, err := api.New(cfg, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil)
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func postTransition(t *testing.T, ts *httptest.Server, body string) (int, string) {
	t.Helper()
	resp, err := http.Post(ts.URL+"/api/v1/jira/transition", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/v1/jira/transition: %v", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.StatusCode, string(raw)
}

// happyAtlassian answers a clean To Do → In Progress path.
func happyAtlassian(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if strings.HasSuffix(req.URL.Path, "/transitions") && req.Method == http.MethodPost {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasSuffix(req.URL.Path, "/transitions") {
		fmt.Fprintf(w, `{"transitions":[{"id":"21","to":{"name":"In Progress"}}]}`)
		return
	}
	fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
}

func TestJiraTransitionHandler(t *testing.T) {
	ts := jiraTestServer(t, happyAtlassian)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"In Progress"}`)
	if status != http.StatusOK {
		t.Fatalf("status = %d, body %s, want 200", status, body)
	}
	var got struct {
		Key    string `json:"key"`
		Status string `json:"status"`
		Moved  bool   `json:"moved"`
	}
	if err := json.Unmarshal([]byte(body), &got); err != nil {
		t.Fatalf("decode response %s: %v", body, err)
	}
	if got.Key != "KAN-571" || got.Status != "In Progress" || !got.Moved {
		t.Fatalf("response %+v, want KAN-571 moved to In Progress", got)
	}
}

func TestJiraTransitionUnrecognizedStatus(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"fields":{"status":{"name":"Blocked"}}}`)
	}
	ts := jiraTestServer(t, h)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"Done"}`)
	if status != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, body %s, want 422", status, body)
	}
	if !strings.Contains(body, "Blocked") {
		t.Fatalf("body %s does not name the unrecognized status", body)
	}
}

func TestJiraTransitionUpstreamNotFound(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}
	ts := jiraTestServer(t, h)

	status, body := postTransition(t, ts, `{"key":"KAN-9999","target":"In Progress"}`)
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, body %s, want 404", status, body)
	}
}

func TestJiraTransitionNoTransitionOffered(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(req.URL.Path, "/transitions") {
			fmt.Fprintf(w, `{"transitions":[{"id":"11","to":{"name":"To Do"}}]}`)
			return
		}
		fmt.Fprintf(w, `{"fields":{"status":{"name":"To Do"}}}`)
	}
	ts := jiraTestServer(t, h)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"Done"}`)
	if status != http.StatusConflict {
		t.Fatalf("status = %d, body %s, want 409", status, body)
	}
	if !strings.Contains(body, "no transition") {
		t.Fatalf("body %s does not name the missing transition", body)
	}
}

func TestJiraTransitionUpstreamAuthRejected(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}
	ts := jiraTestServer(t, h)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"In Progress"}`)
	if status != http.StatusBadGateway {
		t.Fatalf("status = %d, body %s, want 502", status, body)
	}
	if !strings.Contains(body, "authentication rejected") {
		t.Fatalf("body %s does not name the credential failure", body)
	}
}

// A permanently-transient upstream burns the daemon's whole retry ladder
// (real backoff, ~4s) and answers 504 -- the one mapped failure the caller
// may treat as "try again later".
func TestJiraTransitionUpstreamTransientExhausted(t *testing.T) {
	h := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	ts := jiraTestServer(t, h)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"In Progress"}`)
	if status != http.StatusGatewayTimeout {
		t.Fatalf("status = %d, body %s, want 504", status, body)
	}
	if !strings.Contains(body, "transient failure on every attempt") {
		t.Fatalf("body %s does not name the exhausted ladder", body)
	}
}

func TestJiraTransitionNotConfigured(t *testing.T) {
	ts := jiraTestServer(t, nil)

	status, body := postTransition(t, ts, `{"key":"KAN-571","target":"In Progress"}`)
	if status != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, body %s, want 503", status, body)
	}
	if !strings.Contains(body, "FLOWD_JIRA_SITE") {
		t.Fatalf("body %s does not name the configuration", body)
	}
}

func TestJiraTransitionBadBody(t *testing.T) {
	ts := jiraTestServer(t, happyAtlassian)

	cases := []struct {
		name string
		body string
		want string
	}{
		{"empty key", `{"key":"","target":"In Progress"}`, "key is required"},
		{"unknown target", `{"key":"KAN-571","target":"Blocked"}`, "accepted"},
		{"unknown field", `{"key":"KAN-571","target":"Done","extra":1}`, "malformed"},
		{"not json", `nope`, "malformed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, body := postTransition(t, ts, tc.body)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, body %s, want 400", status, body)
			}
			if !strings.Contains(body, tc.want) {
				t.Fatalf("body %s does not name %q", body, tc.want)
			}
		})
	}
}
