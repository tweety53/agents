package guard

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// openRequestsEnv serves body as the change's request rows from a fake flowd.
func openRequestsEnv(t *testing.T, body string) (core.Env, string) {
	t.Helper()
	tmp := t.TempDir()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/chg/requests") {
			w.Header().Set("Flow-Daemon", "flowd/1")
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return core.Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return filepath.Join(tmp, "home")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}, gitWorktree(t, tmp, "chg")
}

// TestCheckOpenRequestsListsOpen: a begun, never-ended request is listed and exits 1.
func TestCheckOpenRequestsListsOpen(t *testing.T) {
	t.Parallel()
	env, wt := openRequestsEnv(t, `[{"n":1,"title":"a","outcome":"landed"},{"n":2,"title":"fix b","outcome":""},{"n":3,"title":"c","outcome":"dropped"}]`)
	var o, e bytes.Buffer
	if rc := checkOpenRequests([]string{wt}, env, &o, &e); rc != 1 || o.String() != "OPEN 2 fix b\n" {
		t.Fatalf("rc=%d out=%q err=%q", rc, o.String(), e.String())
	}
}

// TestCheckOpenRequestsNoneOpen: every request ended exits 0 silently.
func TestCheckOpenRequestsNoneOpen(t *testing.T) {
	t.Parallel()
	env, wt := openRequestsEnv(t, `[{"n":1,"title":"a","outcome":"answered"}]`)
	var o, e bytes.Buffer
	if rc := checkOpenRequests([]string{wt}, env, &o, &e); rc != 0 || o.String() != "" {
		t.Fatalf("rc=%d out=%q err=%q", rc, o.String(), e.String())
	}
}

// TestCheckOpenRequestsUsage: no worktree is a usage error.
func TestCheckOpenRequestsUsage(t *testing.T) {
	t.Parallel()
	var o, e bytes.Buffer
	if rc := checkOpenRequests(nil, core.Env{Getenv: os.Getenv}, &o, &e); rc != 2 {
		t.Fatalf("rc=%d", rc)
	}
}
