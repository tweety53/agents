package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

type fakeBottlenecks struct {
	rows    []store.BottleneckRow
	last    int
	project *string
}

func (f *fakeBottlenecks) Bottlenecks(_ context.Context, project *string, last int) ([]store.BottleneckRow, int64, error) {
	f.last, f.project = last, project
	return f.rows, 1000, nil
}

// TestBottlenecksRoute pins the ranking: -last defaults to 20, each
// dimension keeps its 10 largest, share is total over wall clock, and the
// sinks are the 5 largest across dimensions with role-class left out.
func TestBottlenecksRoute(t *testing.T) {
	f := &fakeBottlenecks{rows: []store.BottleneckRow{{Dimension: "role-class", Name: "reviewer", Runs: 1, TotalMs: 900}}}
	for i := range 12 {
		f.rows = append(f.rows, store.BottleneckRow{Dimension: "guard", Name: fmt.Sprintf("g%d", i), Runs: 1, TotalMs: int64(500 - i)})
	}
	fs := newFakeStore()
	srv, err := api.New(config.Config{Host: "127.0.0.1", DSN: "unused"}, fs, fs, fs, fs, nil, api.WithBottlenecks(f))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/api/v1/stats/bottlenecks?project=p")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got api.BottlenecksResponse
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if f.last != 20 || f.project == nil || *f.project != "p" {
		t.Errorf("store asked last=%d project=%v", f.last, f.project)
	}
	if n := len(got.Ranked["guard"]); n != 10 {
		t.Errorf("guard rows = %d, want 10", n)
	}
	if s := got.Ranked["guard"][0].Share; s != 0.5 {
		t.Errorf("share = %v, want 0.5", s)
	}
	if len(got.Sinks) != 5 || got.Sinks[0].Name != "g0" {
		t.Errorf("sinks = %+v", got.Sinks)
	}
	bad, err := http.Get(ts.URL + "/api/v1/stats/bottlenecks?last=0")
	if err != nil {
		t.Fatal(err)
	}
	bad.Body.Close()
	if bad.StatusCode != http.StatusBadRequest {
		t.Errorf("last=0 status %d, want 400", bad.StatusCode)
	}
}
