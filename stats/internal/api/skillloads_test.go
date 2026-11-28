package api_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/api"
	"github.com/tweety53/agents/stats/internal/config"
	"github.com/tweety53/agents/stats/internal/store"
)

// skillLoadFake is the SkillLoadStore the routes' tests wire through api.WithSkillLoads.
type skillLoadFake struct {
	path, session string
	from, to      time.Time
}

func (f *skillLoadFake) RecordSkillLoad(_ context.Context, path, sessionID string, _ time.Time) error {
	f.path, f.session = path, sessionID
	return nil
}

func (f *skillLoadFake) SkillLoadCounts(_ context.Context, from, to time.Time) ([]store.SkillLoadCount, error) {
	f.from, f.to = from, to
	return []store.SkillLoadCount{{Path: "skills/flow/fix.md", Loads: 2}}, nil
}

// TestSkillLoadRoutes pins POST /api/v1/skill-loads (204, the load reaching the store; 400 on an
// empty path) and GET /api/v1/skill-loads?from&to (the window reaching the store, its rows back).
func TestSkillLoadRoutes(t *testing.T) {
	t.Parallel()
	f := &skillLoadFake{}
	srv, err := api.New(config.Config{Host: "127.0.0.1", DSN: "unused"}, newFakeStore(), newFakeStore(), newFakeStore(), newFakeStore(), nil, api.WithSkillLoads(f))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)

	post := func(body string) int {
		resp, err := http.Post(ts.URL+"/api/v1/skill-loads", "application/json", strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp.StatusCode
	}
	if s := post(`{"path":"skills/flow/fix.md","sessionId":"s-1"}`); s != http.StatusNoContent || f.path != "skills/flow/fix.md" || f.session != "s-1" {
		t.Errorf("POST: status %d, store got %q %q", s, f.path, f.session)
	}
	if s := post(`{"path":"","sessionId":"s-1"}`); s != http.StatusBadRequest {
		t.Errorf("POST empty path: status %d, want 400", s)
	}

	from, to := "2026-10-01T00:00:00Z", "2026-11-01T00:00:00Z"
	resp, err := http.Get(ts.URL + "/api/v1/skill-loads?" + url.Values{"from": {from}, "to": {to}}.Encode())
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || string(b) != `{"rows":[{"path":"skills/flow/fix.md","loads":2,"first":"0001-01-01T00:00:00Z"}]}`+"\n" ||
		f.from.Format(time.RFC3339) != from || f.to.Format(time.RFC3339) != to {
		t.Errorf("GET: status %d body %s window %v..%v", resp.StatusCode, b, f.from, f.to)
	}
	if resp, err := http.Get(ts.URL + "/api/v1/skill-loads?from=x"); err != nil || resp.StatusCode != http.StatusBadRequest {
		t.Errorf("GET bad window: %v %v", err, resp.StatusCode)
	}
}
