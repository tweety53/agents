package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/records"
)

// TestHazardCommandRoundTrip pins the three verbs' happy paths against a
// genuine-daemon stub: add prints "recorded: hazard" and POSTs the wire
// shape; remove prints "retired: hazard <name>" and PATCHes the named row;
// hazards prints the store's array as JSON on stdout.
func TestHazardCommandRoundTrip(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		gotBody, _ = readAll(r)
		// The project segment is the -C repo's own resolved key, matched by
		// path shape rather than a literal name.
		switch {
		case r.Method == http.MethodPatch:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":7,"name":"probe","body":"b","applies":"all","active":false,"createdAt":"2026-09-09T12:00:00Z"}`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":7,"name":"probe","body":"b","applies":"all","active":true,"createdAt":"2026-09-09T12:00:00Z"}`))
		default:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[{"id":7,"name":"probe","body":"b","applies":"all","active":true,"createdAt":"2026-09-09T12:00:00Z"}]`))
		}
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"hazard", "add", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-name", "probe", "-text", "b", "-applies", "all"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hazard add exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "recorded: hazard" {
		t.Errorf("hazard add stdout = %q, want %q", got, "recorded: hazard")
	}
	if gotMethod != http.MethodPost || !strings.HasPrefix(gotPath, "/api/v1/hazards/") {
		t.Errorf("add request = %s %s, want POST under /api/v1/hazards/", gotMethod, gotPath)
	}

	var sent records.Hazard
	if err := json.Unmarshal(gotBody, &sent); err != nil {
		t.Fatalf("decode add body %s: %v", gotBody, err)
	}
	if sent.Name != "probe" || sent.Body != "b" || sent.Applies != "all" {
		t.Errorf("add body = %+v, want name probe, body b, applies all", sent)
	}

	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(),
		[]string{"hazards", "-addr", srv.URL, "-timeout", "500ms", "-C", repo},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hazards exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	var listed []records.Hazard
	if err := json.Unmarshal(stdout.Bytes(), &listed); err != nil {
		t.Fatalf("decode hazards stdout %s: %v", stdout.String(), err)
	}
	if len(listed) != 1 || listed[0].Name != "probe" {
		t.Errorf("hazards = %+v, want the probe row", listed)
	}

	stdout.Reset()
	stderr.Reset()
	code = run(context.Background(),
		[]string{"hazard", "remove", "-addr", srv.URL, "-timeout", "500ms", "-C", repo, "-name", "probe"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hazard remove exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if got := strings.TrimSpace(stdout.String()); got != "retired: hazard probe" {
		t.Errorf("hazard remove stdout = %q, want %q", got, "retired: hazard probe")
	}
}

// TestHazardAddRefusesBadFlags pins the caller-mistake refusals: a missing
// required flag and an -applies outside the closed vocabulary each exit 2
// before any store is contacted.
func TestHazardAddRefusesBadFlags(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	contacted := false
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		contacted = true
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	for _, tc := range []struct {
		name string
		args []string
	}{
		{"missing -name", []string{"hazard", "add", "-addr", srv.URL, "-C", repo, "-text", "b", "-applies", "all"}},
		{"missing -text", []string{"hazard", "add", "-addr", srv.URL, "-C", repo, "-name", "n", "-applies", "all"}},
		{"missing -applies", []string{"hazard", "add", "-addr", srv.URL, "-C", repo, "-name", "n", "-text", "b"}},
		{"bad -applies", []string{"hazard", "add", "-addr", srv.URL, "-C", repo, "-name", "n", "-text", "b", "-applies", "always"}},
		{"unknown subcommand", []string{"hazard", "retract", "-addr", srv.URL, "-C", repo}},
		{"bad -shape on hazards", []string{"hazards", "-addr", srv.URL, "-C", repo, "-shape", "bogus"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), tc.args, strings.NewReader(""), &stdout, &stderr)
			if code != 2 {
				t.Fatalf("exit = %d, want 2; stderr:\n%s", code, stderr.String())
			}
			if stderr.Len() == 0 {
				t.Errorf("stderr empty, want a caller-mistake line")
			}
		})
	}
	if contacted {
		t.Errorf("the store was reached for a caller mistake")
	}
}

// TestHazardsListFlags pins the read path's forwarding: -shape reaches the
// daemon as the shape query parameter, -all as all=true, and a -shape
// outside the vocabulary exits 2 before the store is contacted (covered in
// TestHazardAddRefusesBadFlags's bad -shape case).
func TestHazardsListFlags(t *testing.T) {
	repo := gitRepo(t)
	isolatedStateRoot(t)

	var gotQuery string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	var stdout, stderr bytes.Buffer
	code := run(context.Background(),
		[]string{"hazards", "-addr", srv.URL, "-timeout", "500ms", "-C", repo,
			"-shape", "cross-repo", "-all"},
		strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hazards exit = %d, want 0; stderr:\n%s", code, stderr.String())
	}
	if gotQuery != "all=true&shape=cross-repo" {
		t.Errorf("query = %q, want all=true&shape=cross-repo", gotQuery)
	}
	if strings.TrimSpace(stdout.String()) != "[]" {
		t.Errorf("stdout = %q, want []", stdout.String())
	}
}
