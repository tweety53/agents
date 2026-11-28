package guardtest

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// FakeFlowdRows is a flowd answering change records' GET with rows() as their dispatches, under
// the daemon header the flow client trusts; every other route goes to next (nil: 404).
func FakeFlowdRows(t *testing.T, rows func() string, next http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.Count(strings.TrimPrefix(r.URL.Path, "/api/v1/records/"), "/") == 1 &&
			strings.HasPrefix(r.URL.Path, "/api/v1/records/") {
			w.Header().Set("Flow-Daemon", "flowd/1")
			_, _ = w.Write([]byte(`{"dispatches":` + rows() + `}`))
			return
		}
		if next == nil {
			http.NotFound(w, r)
			return
		}
		next(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// GitWorktree is a fresh git repository dir/name, so a guard derives a project key from it.
func GitWorktree(t *testing.T, dir, name string) string {
	t.Helper()
	wt := filepath.Join(dir, name)
	if out, err := exec.Command("git", "init", "-q", wt).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	return wt
}
