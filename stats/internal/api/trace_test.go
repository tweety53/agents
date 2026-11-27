package api_test

import (
	"context"
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

type fakeTrace struct{ project, change string }

func (f *fakeTrace) TraceSpans(_ context.Context, project, change string) ([]store.TraceSpan, error) {
	f.project, f.change = project, change
	if change != "mine" {
		return nil, store.ErrChangeNotFound
	}
	return []store.TraceSpan{{Kind: "stage", Lane: "stage", Name: "flow.verify #1", Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}}, nil
}

// TestTraceRoute: the route passes project and change to the store and
// writes its spans, an open one without "end"; a missing parameter is 400,
// an unknown change 404.
func TestTraceRoute(t *testing.T) {
	f := &fakeTrace{}
	fs := newFakeStore()
	srv, err := api.New(config.Config{Host: "127.0.0.1", DSN: "unused"}, fs, fs, fs, fs, nil, api.WithTrace(f))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	get := func(q string) (int, string) {
		resp, err := http.Get(ts.URL + "/api/v1/stats/trace?" + q)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, strings.TrimSpace(string(b))
	}
	code, body := get("project=p&change=mine")
	if code != http.StatusOK || f.project != "p" || body != `[{"kind":"stage","lane":"stage","name":"flow.verify #1","start":"2026-01-01T00:00:00Z"}]` {
		t.Errorf("ok = %d %s (store asked %s/%s)", code, body, f.project, f.change)
	}
	if code, _ := get("project=p"); code != http.StatusBadRequest {
		t.Errorf("no change = %d, want 400", code)
	}
	if code, _ := get("project=p&change=absent"); code != http.StatusNotFound {
		t.Errorf("absent = %d, want 404", code)
	}
}
