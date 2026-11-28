package guard

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// awaitRowsEnv serves the rows from a fake flowd: rows, or rowsEnded once a `flow record dispatch
// end` -- the sweep's write, still the flow CLI on PATH -- has been called. It returns the env,
// HOME and the change's worktree.
func awaitRowsEnv(t *testing.T, rows, rowsEnded string) (core.Env, string, string) {
	t.Helper()
	tmp := t.TempDir()
	swept := filepath.Join(tmp, "swept")
	srv := guardtest.FakeFlowdRows(t, func() string {
		if _, err := os.Stat(swept); err == nil {
			return rowsEnded
		}
		return rows
	}, nil)
	guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\ntouch "+swept+"\n")
	home := filepath.Join(tmp, "home")
	return core.Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return home
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}, home, guardtest.GitWorktree(t, tmp, "chg")
}

// TestAwaitRowsAllEnded: every named key ended prints one ENDED line per key, in argument order.
func TestAwaitRowsAllEnded(t *testing.T) {
	t.Parallel()
	rows := `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"b","outcome":"aborted","endedAt":"2026-01-01T00:00:00Z"},{"key":"c","endedAt":null}]`
	env, _, wt := awaitRowsEnv(t, rows, rows)
	var o, e bytes.Buffer
	if rc := awaitRows([]string{wt, "b", "a"}, env, &o, &e); rc != 0 || o.String() != "ENDED b aborted\nENDED a completed\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
}

// TestAwaitRowsWaiting: a key still open at the timeout prints WAITING with it, exit 0.
func TestAwaitRowsWaiting(t *testing.T) {
	t.Parallel()
	rows := `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"c","endedAt":null}]`
	env, _, wt := awaitRowsEnv(t, rows, rows)
	var o, e bytes.Buffer
	start := time.Now()
	if rc := awaitRows([]string{wt, "a", "c", "-timeout", "1"}, env, &o, &e); rc != 0 || o.String() != "ENDED a completed\nWAITING c\n" {
		t.Errorf("rc %d stdout %q", rc, o.String())
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("waited past its timeout: %s", time.Since(start))
	}
}

// TestAwaitRowsSweepsStale: a row whose agent finished and went quiet is ended by the sweep inside
// the wait and reported.
func TestAwaitRowsSweepsStale(t *testing.T) {
	t.Parallel()
	env, home, wt := awaitRowsEnv(t,
		`[{"key":"s","agentId":"aaa","sessionToken":"t","endedAt":null}]`,
		`[{"key":"s","agentId":"aaa","sessionToken":"t","outcome":"aborted","endedAt":"2026-01-01T00:00:00Z"}]`)
	sub := filepath.Join(home, ".claude", "projects", "p", "sess", "subagents")
	guardtest.WriteFile(t, filepath.Join(sub, "agent-aaa.meta.json"), `{"finished":true}`)
	tr := filepath.Join(sub, "agent-aaa.jsonl")
	guardtest.WriteFile(t, tr, "{}\n")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(tr, old, old); err != nil {
		t.Fatal(err)
	}
	var o, e bytes.Buffer
	if rc := awaitRows([]string{wt, "s", "-timeout", "1"}, env, &o, &e); rc != 0 || o.String() != "ENDED s aborted\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
}

// TestAwaitRowsUnknownKey: a key the change does not record exits 2.
func TestAwaitRowsUnknownKey(t *testing.T) {
	t.Parallel()
	env, _, wt := awaitRowsEnv(t, `[{"key":"a","endedAt":null}]`, `[]`)
	var o, e bytes.Buffer
	if rc := awaitRows([]string{wt, "a", "zzz"}, env, &o, &e); rc != 2 {
		t.Errorf("rc %d stdout %q", rc, o.String())
	}
}

// TestAwaitRowsAny: -any returns as soon as one named key is ended, the still-open ones on WAITING.
func TestAwaitRowsAny(t *testing.T) {
	t.Parallel()
	rows := `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"b","endedAt":null}]`
	env, _, wt := awaitRowsEnv(t, rows, rows)
	var o, e bytes.Buffer
	start := time.Now()
	if rc := awaitRows([]string{wt, "-any", "b", "a", "-timeout", "4"}, env, &o, &e); rc != 0 || o.String() != "ENDED a completed\nWAITING b\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
	if time.Since(start) > 2*time.Second {
		t.Errorf("-any waited %s with a key already ended", time.Since(start))
	}
}

// awaitRowsGitEnv is a change worktree chg whose flowd serves rows.json as it stands and hands
// every other route -- the wait route -- to wait.
func awaitRowsGitEnv(t *testing.T, wait http.HandlerFunc, rows string) (core.Env, string, string) {
	t.Helper()
	tmp := t.TempDir()
	guardtest.WriteFile(t, filepath.Join(tmp, "rows.json"), rows)
	srv := guardtest.FakeFlowdRows(t, func() string {
		b, _ := os.ReadFile(filepath.Join(tmp, "rows.json"))
		return string(b)
	}, wait)
	return core.Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return filepath.Join(tmp, "home")
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}, guardtest.GitWorktree(t, tmp, "chg"), tmp
}

// TestAwaitRowsLongPollWakes: with flowd's wait route answering the moment the row closes,
// await-rows reports it at once rather than on its next 5 s poll.
func TestAwaitRowsLongPollWakes(t *testing.T) {
	t.Parallel()
	closeRow := make(chan struct{})
	var tmp string
	gotQuery := make(chan string, 8)
	wait := func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chg/dispatches/wait") {
			http.NotFound(w, r)
			return
		}
		gotQuery <- r.URL.RawQuery
		select {
		case <-closeRow:
			guardtest.WriteFile(t, filepath.Join(tmp, "rows.json"), `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"}]`)
		case <-r.Context().Done():
		}
		_, _ = w.Write([]byte(`[]`))
	}
	env, wt, dir := awaitRowsGitEnv(t, wait, `[{"key":"a","endedAt":null}]`)
	tmp = dir
	go func() { time.Sleep(300 * time.Millisecond); close(closeRow) }()
	var o, e bytes.Buffer
	start := time.Now()
	if rc := awaitRows([]string{wt, "-any", "a", "-timeout", "30"}, env, &o, &e); rc != 0 || o.String() != "ENDED a completed\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
	if took := time.Since(start); took > 3*time.Second {
		t.Errorf("reported %s after start, want the long-poll's wakeup", took)
	}
	select {
	case q := <-gotQuery:
		if !strings.Contains(q, "mode=any") || !strings.Contains(q, "keys=a") {
			t.Errorf("wait query %q, want keys=a and mode=any", q)
		}
	default:
		t.Error("await-rows never called the wait route")
	}
}

// TestAwaitRowsFallsBackToPolling: a flowd answering the wait route with no JSON array (an older
// daemon's SPA page) or not at all leaves await-rows polling, its contract unchanged.
func TestAwaitRowsFallsBackToPolling(t *testing.T) {
	t.Parallel()
	spa := func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<!doctype html>")) }
	dead := func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) }
	for name, wait := range map[string]http.HandlerFunc{"spa": spa, "dead": dead} {
		env, wt, _ := awaitRowsGitEnv(t, wait, `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"c","endedAt":null}]`)
		var o, e bytes.Buffer
		start := time.Now()
		if rc := awaitRows([]string{wt, "a", "c", "-timeout", "1"}, env, &o, &e); rc != 0 || o.String() != "ENDED a completed\nWAITING c\n" {
			t.Errorf("%s: rc %d stdout %q err %q", name, rc, o.String(), e.String())
		}
		if took := time.Since(start); took > 3*time.Second {
			t.Errorf("%s: waited %s past its 1 s timeout", name, took)
		}
	}
}

// TestAwaitRowsRecordsItsWait: the time await-rows blocks between reads -- its long-poll or, with
// no wait route, its sleep -- lands in the run's collector as wait.
func TestAwaitRowsRecordsItsWait(t *testing.T) {
	t.Parallel()
	env := core.Env{Reasons: &core.Reasons{}}
	awaitRowsBlock(env, "", 50*time.Millisecond)
	if w := env.Reasons.Wait(); w < 50*time.Millisecond {
		t.Errorf("wait = %s, want the 50 ms sleep", w)
	}
}
