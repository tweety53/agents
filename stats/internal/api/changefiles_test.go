package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// newChangeFilesServer wires the change-file routes against the real store.
func newChangeFilesServer(t *testing.T, st *store.Store) *httptest.Server {
	t.Helper()
	cfg := config.Config{Host: "127.0.0.1", Port: 0, DSN: "unused"}
	srv, err := api.New(cfg, st, st, st, st, st, nil, api.WithChangeFiles(st))
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func doRequest(t *testing.T, method, url string, body []byte) (int, string, []byte) {
	t.Helper()
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("build %s %s: %v", method, url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header.Get("Content-Type"), out
}

// TestChangeFilePutThenGetServesBody pins the round trip: the saved
// listing is false before a PUT, the PUT answers 204, the GET serves the
// body byte for byte as text/markdown, and the listing turns true.
func TestChangeFilePutThenGetServesBody(t *testing.T) {
	t.Parallel()
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-change-files-api-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1073", nil)
	ts := newChangeFilesServer(t, st)
	base := ts.URL + "/api/v1/changes/" + projectKey + "/kan-1073/files"

	saved := func() bool {
		t.Helper()
		status, _, body := doRequest(t, http.MethodGet, base, nil)
		if status != http.StatusOK {
			t.Fatalf("GET listing status %d, body %s", status, body)
		}
		var out struct{ Saved bool }
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("decode listing %s: %v", body, err)
		}
		return out.Saved
	}
	if saved() {
		t.Fatal("saved = true before any PUT")
	}

	want := "# plan\n\n- [ ] 1. a task\n"
	raw, _ := json.Marshal(map[string]string{"body": want})
	if status, _, body := doRequest(t, http.MethodPut, base+"/tasks.md", raw); status != http.StatusNoContent {
		t.Fatalf("PUT status %d, body %s", status, body)
	}
	status, ctype, body := doRequest(t, http.MethodGet, base+"/tasks.md", nil)
	if status != http.StatusOK {
		t.Fatalf("GET status %d, body %s", status, body)
	}
	if string(body) != want {
		t.Errorf("GET body = %q, want %q", body, want)
	}
	if ctype != "text/markdown; charset=utf-8" {
		t.Errorf("Content-Type = %q", ctype)
	}
	if !saved() {
		t.Error("saved = false after a PUT")
	}
}

// TestChangeFilePutAcceptsABodyOverTwoMiB pins the request-body cap above
// the largest change file observed: a self-review bundle past 2 MiB saves.
func TestChangeFilePutAcceptsABodyOverTwoMiB(t *testing.T) {
	t.Parallel()
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-change-files-big-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1087", nil)
	ts := newChangeFilesServer(t, st)
	url := ts.URL + "/api/v1/changes/" + projectKey + "/kan-1087/files/self-review-context.md"

	raw, _ := json.Marshal(map[string]string{"body": strings.Repeat("x", 3<<20)})
	if status, _, body := doRequest(t, http.MethodPut, url, raw); status != http.StatusNoContent {
		t.Fatalf("PUT status %d, body %.200s", status, body)
	}
	status, _, body := doRequest(t, http.MethodGet, url, nil)
	if status != http.StatusOK {
		t.Fatalf("GET status %d, body %.200s", status, body)
	}
	if len(body) != 3<<20 {
		t.Errorf("GET body length = %d, want %d", len(body), 3<<20)
	}
}

// TestChangeFileGetMissingIs404 pins both absences as 404: a path never
// saved for a known change, and a change the store does not hold.
func TestChangeFileGetMissingIs404(t *testing.T) {
	t.Parallel()
	st := newIntegrationStore(t)
	projectKey := fmt.Sprintf("proj-change-files-404-%d", time.Now().UnixNano())
	mustPutIntegrationChange(t, st, projectKey, "kan-1073", nil)
	ts := newChangeFilesServer(t, st)

	for _, url := range []string{
		ts.URL + "/api/v1/changes/" + projectKey + "/kan-1073/files/tasks.md",
		ts.URL + "/api/v1/changes/" + projectKey + "/no-such-change/files/tasks.md",
	} {
		if status, _, body := doRequest(t, http.MethodGet, url, nil); status != http.StatusNotFound {
			t.Errorf("GET %s status %d, body %s; want 404", url, status, body)
		}
	}
}
