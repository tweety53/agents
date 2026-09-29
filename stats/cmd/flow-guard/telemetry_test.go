package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
)

func TestOutcomeForUsesTheGuardsOwnCannotAnswerCode(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		code int
		want string
	}{
		{"check-plan-shape", 0, "clear"},
		{"check-plan-shape", 1, "fired"},
		{"check-plan-shape", 2, "cannot-answer"},
		{"run-reproducer", 2, "fired"},
		{"run-reproducer", 4, "cannot-answer"},
		{"mutate-and-verify", 4, "cannot-answer"},
	} {
		if got := outcomeFor(tc.name, tc.code); got != tc.want {
			t.Errorf("outcomeFor(%q, %d) = %q, want %q", tc.name, tc.code, got, tc.want)
		}
	}
}

func TestRecordsAddrOrder(t *testing.T) {
	t.Parallel()
	env := func(m map[string]string) func(string) string { return func(k string) string { return m[k] } }
	if got := recordsAddr(env(nil)); got != "http://127.0.0.1:4173" {
		t.Errorf("default = %q", got)
	}
	if got := recordsAddr(env(map[string]string{"FLOW_ADDR": "http://a/"})); got != "http://a" {
		t.Errorf("FLOW_ADDR = %q", got)
	}
	if got := recordsAddr(env(map[string]string{"FLOW_ADDR": "http://a", "FLOW_RECORDS_ADDR": "http://r"})); got != "http://r" {
		t.Errorf("FLOW_RECORDS_ADDR = %q, want it to win", got)
	}
}

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return dir
}

// TestProjectKeyMatchesFallback pins this package's stdlib copy of the
// project-key derivation to internal/fallback's, which every flow CLI write
// uses -- the two must name the same project or no guard run is ever
// recorded. A test file may import fallback: _test.go files are outside the
// flow-guard cache key.
func TestProjectKeyMatchesFallback(t *testing.T) {
	t.Parallel()
	dir := gitRepo(t)
	sub := filepath.Join(dir, "sub")
	if err := exec.Command("mkdir", "-p", sub).Run(); err != nil {
		t.Fatal(err)
	}
	want, _, err := fallback.ProjectKey(sub)
	if err != nil {
		t.Fatalf("fallback.ProjectKey: %v", err)
	}
	got, err := projectKey(t.Context(), sub)
	if err != nil {
		t.Fatalf("projectKey: %v", err)
	}
	if got != want {
		t.Fatalf("projectKey = %q, fallback.ProjectKey = %q", got, want)
	}
}

func TestRecordRunPostsTheRun(t *testing.T) {
	t.Parallel()
	dir := gitRepo(t)
	key, err := projectKey(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}
	var gotPath string
	var got guardRun
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	recordRun(func(k string) string { return env[k] }, dir, "check-plan-shape", 1, 1500*time.Millisecond, srv.Client())

	if gotPath != "/api/v1/guard-runs/"+key {
		t.Errorf("path = %q, want the project's guard-runs route", gotPath)
	}
	if got.Guard != "check-plan-shape" || got.ExitCode != 1 || got.Outcome != "fired" || got.DurationMS != 1500 || got.Worktree != dir {
		t.Errorf("body = %+v", got)
	}
}

func TestRecordRunHonoursTheOptOut(t *testing.T) {
	t.Parallel()
	var called atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL, "FLOW_GUARD_TELEMETRY": "off"}
	recordRun(func(k string) string { return env[k] }, gitRepo(t), "check-plan-shape", 0, time.Millisecond, srv.Client())
	if called.Load() {
		t.Fatal("FLOW_GUARD_TELEMETRY=off still posted a run")
	}
}

// TestRecordRunIsBoundedByItsTimeout: a daemon that never answers costs a
// guard at most telemetryTimeout, never its answer.
func TestRecordRunIsBoundedByItsTimeout(t *testing.T) {
	t.Parallel()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)

	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	start := time.Now()
	recordRun(func(k string) string { return env[k] }, gitRepo(t), "check-plan-shape", 0, 0, srv.Client())
	if elapsed := time.Since(start); elapsed > 3*telemetryTimeout {
		t.Fatalf("recordRun took %v against a hung daemon, want it bounded near %v", elapsed, telemetryTimeout)
	}
}

func TestDispatchedOnlyNamesRegisteredGuards(t *testing.T) {
	t.Parallel()
	if _, ok := dispatched(nil); ok {
		t.Error("no args dispatched")
	}
	if _, ok := dispatched([]string{"no-such-guard"}); ok {
		t.Error("an unknown guard dispatched")
	}
	if name, ok := dispatched([]string{"check-cleanup-complete", "x"}); !ok || name != "check-cleanup-complete" {
		t.Errorf("dispatched = %q, %v", name, ok)
	}
}
