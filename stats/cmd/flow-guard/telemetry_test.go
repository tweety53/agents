package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
	"github.com/tweety53/agents/stats/internal/guard"
)

func TestOutcomeForUsesTheGuardsOwnCannotAnswerCode(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		code  int
		first string
		want  string
	}{
		{"check-references", 0, "", "clear"},
		{"check-references", 1, "", "fired"},
		{"check-references", 2, "", "cannot-answer"},
		{"run-reproducer", 2, "", "fired"},
		{"run-reproducer", 4, "", "cannot-answer"},
		// KAN-1110: check-unfinished-work exits 0 on its OUTSTANDING verdict.
		{"check-unfinished-work", 0, "OUTSTANDING: task 3 has no implementer row", "fired"},
		{"check-unfinished-work", 0, "CLEAR: nothing outstanding", "clear"},
		{"check-references", 0, "OUTSTANDING: not this guard's verdict", "clear"},
	} {
		if got := outcomeFor(tc.name, tc.code, tc.first); got != tc.want {
			t.Errorf("outcomeFor(%q, %d, %q) = %q, want %q", tc.name, tc.code, tc.first, got, tc.want)
		}
	}
}

// TestFirstLineKeepsTheFirstLine pins the tee main.go reads the verdict
// through: the first line, across writes, and every byte passed on.
func TestFirstLineKeepsTheFirstLine(t *testing.T) {
	t.Parallel()
	var sink bytes.Buffer
	f := &firstLine{w: &sink}
	for _, p := range []string{"OUTST", "ANDING: x\nsecond\n", "third\n"} {
		if _, err := f.Write([]byte(p)); err != nil {
			t.Fatal(err)
		}
	}
	if string(f.line) != "OUTSTANDING: x" || sink.String() != "OUTSTANDING: x\nsecond\nthird\n" {
		t.Errorf("line %q, passed on %q", f.line, sink.String())
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
	recordRun(func(k string) string { return env[k] }, dir, "check-references", 1, "", nil, 1500*time.Millisecond, srv.Client())

	if gotPath != "/api/v1/guard-runs/"+key {
		t.Errorf("path = %q, want the project's guard-runs route", gotPath)
	}
	if got.Guard != "check-references" || got.ExitCode != 1 || got.Outcome != "fired" || got.DurationMS != 1500 || got.Worktree != dir {
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
	recordRun(func(k string) string { return env[k] }, gitRepo(t), "check-references", 0, "", nil, time.Millisecond, srv.Client())
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
	recordRun(func(k string) string { return env[k] }, gitRepo(t), "check-references", 0, "", nil, 0, srv.Client())
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
	if name, ok := dispatched([]string{"check-unfinished-work", "x"}); !ok || name != "check-unfinished-work" {
		t.Errorf("dispatched = %q, %v", name, ok)
	}
}

// TestRunAndRecordPassesTheFirstLine pins KAN-1110: the first line a guard
// prints reaches recordRun, so check-unfinished-work's exit-0 OUTSTANDING
// verdict is recorded fired. Not parallel: it swaps a registry entry.
func TestRunAndRecordPassesTheFirstLine(t *testing.T) {
	dir := gitRepo(t)
	const name = "check-unfinished-work"
	orig := registry[name]
	registry[name] = func(_ []string, _ guard.Env, stdout, _ io.Writer) int {
		fmt.Fprintln(stdout, "OUTSTANDING: task 3 has no implementer row")
		return 0
	}
	defer func() { registry[name] = orig }()
	var got guardRun
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	var stdout bytes.Buffer
	code := runAndRecord([]string{name}, &stdout, io.Discard, func(k string) string { return env[k] }, func() string { return dir }, srv.Client())
	if code != 0 || got.Guard != name || got.Outcome != "fired" {
		t.Errorf("exit %d, body %+v; want exit 0 recorded fired", code, got)
	}
}

// TestRunAndRecordPostsTheReasons pins clear-run-drops-reasons: a fired run
// posts its distinct reason ids in first-seen order, a clear run posts none.
// Not parallel: it swaps a registry entry.
func TestRunAndRecordPostsTheReasons(t *testing.T) {
	dir := gitRepo(t)
	const name = "check-references"
	orig := registry[name]
	defer func() { registry[name] = orig }()
	var body []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	getenv := func(k string) string { return env[k] }

	registry[name] = func(_ []string, e guard.Env, _, _ io.Writer) int {
		guard.Reason(e, "F6")
		guard.Reason(e, "F2")
		guard.Reason(e, "F6")
		return 1
	}
	runAndRecord([]string{name}, io.Discard, io.Discard, getenv, func() string { return dir }, srv.Client())
	var fired guardRun
	if err := json.Unmarshal(body, &fired); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if fired.Outcome != "fired" || !slices.Equal(fired.Reasons, []string{"F6", "F2"}) {
		t.Errorf("fired run posted %s; want outcome fired, reasons [F6 F2]", body)
	}

	registry[name] = func(_ []string, e guard.Env, _, _ io.Writer) int {
		guard.Reason(e, "x")
		return 0
	}
	body = nil
	runAndRecord([]string{name}, io.Discard, io.Discard, getenv, func() string { return dir }, srv.Client())
	var clear map[string]any
	if err := json.Unmarshal(body, &clear); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if _, ok := clear["reasons"]; ok {
		t.Errorf("clear run posted %s; want no reasons key", body)
	}
}

// TestGuardWorktreePrefersAWorktreeArgument pins guard-run-worktree-arg: the
// first argument that is an existing directory under a `-worktrees/<change>`
// path is the run's worktree, else the cwd.
func TestGuardWorktreePrefersAWorktreeArgument(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	wt := filepath.Join(tmp, "gymie-worktrees", "kan-1-x")
	sub := filepath.Join(wt, "sub")
	other := filepath.Join(tmp, "gymie-worktrees", "kan-3-z")
	plain := filepath.Join(tmp, "plain")
	for _, d := range []string{sub, other, plain} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	missing := filepath.Join(tmp, "gymie-worktrees", "kan-2-y")
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"worktree argument", []string{"run-verify-list", wt, "test"}, wt},
		{"a subdirectory counts", []string{"run-verify-list", sub}, sub},
		{"plain directory", []string{"recapture", plain}, "/repo"},
		{"missing worktree skipped", []string{"recapture", missing, other}, other},
		{"first match wins", []string{"recapture", wt, other}, wt},
		{"guard name never counts", []string{wt}, "/repo"},
	} {
		if got := guardWorktree(tc.args, "/repo"); got != tc.want {
			t.Errorf("%s: guardWorktree = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// TestRunAndRecordRecordsTheTargetWorktree: a guard driven from the main
// checkout at a change's worktree is recorded against that worktree. Not
// parallel: it swaps a registry entry.
func TestRunAndRecordRecordsTheTargetWorktree(t *testing.T) {
	main := gitRepo(t)
	wt := filepath.Join(t.TempDir(), "gymie-worktrees", "kan-1-x")
	if out, err := exec.Command("git", "init", "-q", wt).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	const name = "check-references"
	orig := registry[name]
	registry[name] = func(_ []string, _ guard.Env, _, _ io.Writer) int { return 0 }
	defer func() { registry[name] = orig }()
	var got guardRun
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	runAndRecord([]string{name, wt}, io.Discard, io.Discard, func(k string) string { return env[k] }, func() string { return main }, srv.Client())
	if got.Worktree != wt {
		t.Errorf("posted worktree %q, want the argument %q", got.Worktree, wt)
	}
}

// TestRunAndRecordFallsBackToTheCwd: a `-worktrees/<x>` argument that is no
// git checkout records the run under the cwd rather than dropping it. Not
// parallel: it swaps a registry entry.
func TestRunAndRecordFallsBackToTheCwd(t *testing.T) {
	main := gitRepo(t)
	stale := filepath.Join(t.TempDir(), "gymie-worktrees", "kan-1-x")
	if err := os.MkdirAll(stale, 0o755); err != nil {
		t.Fatal(err)
	}
	const name = "check-references"
	orig := registry[name]
	registry[name] = func(_ []string, _ guard.Env, _, _ io.Writer) int { return 0 }
	defer func() { registry[name] = orig }()
	var got guardRun
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	env := map[string]string{"FLOW_RECORDS_ADDR": srv.URL}
	runAndRecord([]string{name, stale}, io.Discard, io.Discard, func(k string) string { return env[k] }, func() string { return main }, srv.Client())
	if got.Worktree != main {
		t.Errorf("posted worktree %q, want the cwd %q", got.Worktree, main)
	}
}
