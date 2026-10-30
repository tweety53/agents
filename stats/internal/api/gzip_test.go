package api_test

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestServerGzipsAPIResponses asserts a GET under /api/ is gzipped for a
// client that accepts it and left plain for one that does not: the live
// state board's finished read is ~170 KB of repetitive JSON.
func TestServerGzipsAPIResponses(t *testing.T) {
	fs := newFakeStore()
	fs.changes[changeKey("proj", "chg")] = store.Change{ProjectKey: "proj", Name: "chg", State: store.StateInProgress}
	ts := newTestServer(t, fs)

	get := func(enc string) *http.Response {
		t.Helper()
		req, _ := http.NewRequest(http.MethodGet, ts.URL+"/api/v1/changes/proj/chg", nil)
		if enc != "" {
			req.Header.Set("Accept-Encoding", enc)
		}
		resp, err := http.DefaultTransport.RoundTrip(req)
		if err != nil {
			t.Fatalf("GET: %v", err)
		}
		t.Cleanup(func() { resp.Body.Close() })
		return resp
	}

	resp := get("gzip, deflate")
	if resp.Header.Get("Content-Encoding") != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", resp.Header.Get("Content-Encoding"))
	}
	zr, err := gzip.NewReader(resp.Body)
	if err != nil {
		t.Fatalf("gzip.NewReader: %v", err)
	}
	var body map[string]any
	if err := json.NewDecoder(zr).Decode(&body); err != nil {
		t.Fatalf("decode gzipped body: %v", err)
	}
	if body["name"] != "chg" {
		t.Errorf("body name = %v, want chg", body["name"])
	}

	if got := get("").Header.Get("Content-Encoding"); got != "" {
		t.Errorf("Content-Encoding without Accept-Encoding = %q, want none", got)
	}
}
