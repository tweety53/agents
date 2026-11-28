package main

import (
	"bytes"
	"encoding/json"
	"errors"
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

	"github.com/tweety53/agents/stats/internal/guard/core"
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

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	return dir
}

// stateEnv is a getenv over m plus FLOW_STATE_DIR, a fresh state root, and
// the guard-runs file under it.
func stateEnv(t *testing.T, m map[string]string) (func(string) string, string) {
	t.Helper()
	root := t.TempDir()
	return func(k string) string {
		if k == "FLOW_STATE_DIR" {
			return root
		}
		return m[k]
	}, filepath.Join(root, "guard-runs.jsonl")
}

// lines is path's JSON lines, raw.
func lines(t *testing.T, path string) [][]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return bytes.Split(bytes.TrimSuffix(b, []byte("\n")), []byte("\n"))
}

// lastRun decodes path's last line.
func lastRun(t *testing.T, path string) guardRun {
	t.Helper()
	ls := lines(t, path)
	if len(ls) == 0 {
		t.Fatalf("%s holds no run", path)
	}
	var r guardRun
	if err := json.Unmarshal(ls[len(ls)-1], &r); err != nil {
		t.Fatalf("decode %s: %v", ls[len(ls)-1], err)
	}
	return r
}

func TestRecordRunAppendsTheRun(t *testing.T) {
	t.Parallel()
	dir := gitRepo(t)
	getenv, path := stateEnv(t, nil)
	recordRun(getenv, dir, "check-references", 1, "", nil, nil, 300*time.Millisecond, 1500*time.Millisecond)
	recordRun(getenv, dir, "check-references", 0, "", nil, nil, 0, time.Millisecond)
	if n := len(lines(t, path)); n != 2 {
		t.Fatalf("%d lines, want one per run", n)
	}
	var got guardRun
	if err := json.Unmarshal(lines(t, path)[0], &got); err != nil {
		t.Fatal(err)
	}
	if got.Guard != "check-references" || got.ExitCode != 1 || got.Outcome != "fired" || got.DurationMS != 1500 || got.WaitMS != 300 || got.Worktree != dir || got.RecordedAt.IsZero() {
		t.Errorf("line = %+v", got)
	}
}

func TestRecordRunHonoursTheOptOut(t *testing.T) {
	t.Parallel()
	getenv, path := stateEnv(t, map[string]string{"FLOW_GUARD_TELEMETRY": "off"})
	recordRun(getenv, gitRepo(t), "check-references", 0, "", nil, nil, 0, time.Millisecond)
	if ls := lines(t, path); ls != nil {
		t.Fatalf("FLOW_GUARD_TELEMETRY=off still appended %q", ls)
	}
}

// TestRecordRunNeverFailsTheGuard: a state root that cannot be written to
// loses the line, never the guard's answer.
func TestRecordRunNeverFailsTheGuard(t *testing.T) {
	t.Parallel()
	const name = "check-references"
	getenv := func(k string) string {
		if k == "FLOW_STATE_DIR" {
			return filepath.Join(t.TempDir(), "missing", "root")
		}
		return ""
	}
	recordRun(getenv, gitRepo(t), name, 3, "x", []string{"r"}, nil, 0, time.Millisecond)
}

// TestRunAndRecordSpawnsNothingOnExit: recording a run starts no git and
// makes no HTTP request, and the guard's exit code and stdout pass through
// unchanged. Not parallel: it swaps a registry entry and PATH.
func TestRunAndRecordSpawnsNothingOnExit(t *testing.T) {
	dir := gitRepo(t)
	bin := t.TempDir()
	spawned := filepath.Join(t.TempDir(), "spawned")
	if err := os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\necho git >> '"+spawned+"'\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	const name = "check-references"
	orig := registry[name]
	registry[name] = func(_ []string, _ core.Env, stdout, _ io.Writer) int {
		fmt.Fprintln(stdout, "VIOLATION: x")
		return 1
	}
	defer func() { registry[name] = orig }()
	getenv, path := stateEnv(t, map[string]string{"FLOW_RECORDS_ADDR": srv.URL, "FLOW_ADDR": srv.URL})
	var stdout bytes.Buffer
	code := runAndRecord([]string{name}, &stdout, io.Discard, getenv, func() string { return dir })
	if _, err := os.Stat(spawned); err == nil {
		t.Error("recording the run spawned git")
	}
	if hits.Load() != 0 {
		t.Errorf("recording the run made %d HTTP requests", hits.Load())
	}
	if code != 1 || stdout.String() != "VIOLATION: x\n" {
		t.Errorf("exit %d stdout %q, want the guard's own", code, stdout.String())
	}
	if got := lastRun(t, path); got.Guard != name || got.Worktree != dir || got.Outcome != "fired" {
		t.Errorf("line = %+v", got)
	}
}

func TestDispatchedOnlyNamesRegisteredGuards(t *testing.T) {
	t.Parallel()
	if _, ok := dispatched(nil); ok {
		t.Error("no args dispatched")
	}
	if _, ok := dispatched([]string{"hook", "enforce-agent-baseline"}); ok {
		t.Error("a hook run dispatched")
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
	registry[name] = func(_ []string, _ core.Env, stdout, _ io.Writer) int {
		fmt.Fprintln(stdout, "OUTSTANDING: task 3 has no implementer row")
		return 0
	}
	defer func() { registry[name] = orig }()
	getenv, path := stateEnv(t, nil)
	var stdout bytes.Buffer
	code := runAndRecord([]string{name}, &stdout, io.Discard, getenv, func() string { return dir })
	if got := lastRun(t, path); code != 0 || got.Guard != name || got.Outcome != "fired" {
		t.Errorf("exit %d, body %+v; want exit 0 recorded fired", code, got)
	}
}

// TestRunAndRecordRecordsTheReasons pins clear-run-drops-reasons: a fired run
// records its distinct reason ids in first-seen order, a clear run none.
// Not parallel: it swaps a registry entry.
func TestRunAndRecordRecordsTheReasons(t *testing.T) {
	dir := gitRepo(t)
	const name = "check-references"
	orig := registry[name]
	defer func() { registry[name] = orig }()
	getenv, path := stateEnv(t, nil)

	registry[name] = func(_ []string, e core.Env, _, _ io.Writer) int {
		core.Reason(e, "F6")
		core.Reason(e, "F2")
		core.Reason(e, "F6")
		return 1
	}
	runAndRecord([]string{name}, io.Discard, io.Discard, getenv, func() string { return dir })
	if fired := lastRun(t, path); fired.Outcome != "fired" || !slices.Equal(fired.Reasons, []string{"F6", "F2"}) {
		t.Errorf("fired run recorded %+v; want outcome fired, reasons [F6 F2]", fired)
	}

	registry[name] = func(_ []string, e core.Env, _, _ io.Writer) int {
		core.Reason(e, "x")
		return 0
	}
	runAndRecord([]string{name}, io.Discard, io.Discard, getenv, func() string { return dir })
	ls := lines(t, path)
	body := ls[len(ls)-1]
	var clear map[string]any
	if err := json.Unmarshal(body, &clear); err != nil {
		t.Fatalf("decode %s: %v", body, err)
	}
	if _, ok := clear["reasons"]; ok {
		t.Errorf("clear run recorded %s; want no reasons key", body)
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
	registry[name] = func(_ []string, _ core.Env, _, _ io.Writer) int { return 0 }
	defer func() { registry[name] = orig }()
	getenv, path := stateEnv(t, nil)
	runAndRecord([]string{name, wt}, io.Discard, io.Discard, getenv, func() string { return main })
	if got := lastRun(t, path); got.Worktree != wt {
		t.Errorf("recorded worktree %q, want the argument %q", got.Worktree, wt)
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
	registry[name] = func(_ []string, _ core.Env, _, _ io.Writer) int { return 0 }
	defer func() { registry[name] = orig }()
	getenv, path := stateEnv(t, nil)
	runAndRecord([]string{name, stale}, io.Discard, io.Discard, getenv, func() string { return main })
	if got := lastRun(t, path); got.Worktree != main {
		t.Errorf("recorded worktree %q, want the cwd %q", got.Worktree, main)
	}
}

// TestRecordRunRecordsTheBootstrap: FLOW_GUARD_START, scripts/flow-guard's
// $EPOCHREALTIME stamp, records bootstrapMs as the time from it to the
// guard's start (the run's end less its duration); without the stamp, or
// with one that will not parse, no bootstrapMs is recorded.
func TestRecordRunRecordsTheBootstrap(t *testing.T) {
	t.Parallel()
	dir := gitRepo(t)
	stamp := time.Now().Add(-3 * time.Second)
	for name, tc := range map[string]struct {
		start  string
		wantMs float64 // -1: absent
	}{
		"stamped":       {fmt.Sprintf("%d.%06d", stamp.Unix(), stamp.Nanosecond()/1000), 2000},
		"comma decimal": {fmt.Sprintf("%d,%06d", stamp.Unix(), stamp.Nanosecond()/1000), 2000},
		"unstamped":     {"", -1},
		"garbage":       {"soon", -1},
	} {
		getenv, path := stateEnv(t, map[string]string{"FLOW_GUARD_START": tc.start})
		recordRun(getenv, dir, "check-references", 0, "", nil, nil, 0, time.Second)
		var got map[string]any
		if err := json.Unmarshal(lines(t, path)[0], &got); err != nil {
			t.Fatal(err)
		}
		ms, ok := got["bootstrapMs"].(float64)
		switch {
		case tc.wantMs < 0 && ok:
			t.Errorf("%s: bootstrapMs = %v, want none", name, ms)
		case tc.wantMs >= 0 && (!ok || ms < tc.wantMs || ms > tc.wantMs+1000):
			t.Errorf("%s: bootstrapMs = %v (%v), want the 2 s between the stamp and the run's start", name, ms, ok)
		}
	}
}

// TestRecordRunRecordsTheHeadlessCall: a run that made a headless model call
// records it with the run.
func TestRecordRunRecordsTheHeadlessCall(t *testing.T) {
	t.Parallel()
	getenv, path := stateEnv(t, nil)
	usd := 0.05
	h := &core.Headless{Role: "decider", Model: "opus", Effort: "low", CostUSD: &usd, DurationMS: 4200}
	recordRun(getenv, gitRepo(t), "plan-check", 0, "", nil, h, 0, time.Second)
	if got := lastRun(t, path); got.Headless == nil || *got.Headless.CostUSD != usd || got.Headless.Role != "decider" || got.Headless.DurationMS != 4200 {
		t.Errorf("headless = %+v, want the decider call", got.Headless)
	}
}
