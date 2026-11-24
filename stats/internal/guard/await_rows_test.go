package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// awaitRowsEnv puts a fake flow on PATH: `dispatches` prints rows.json, or rows-ended.json once a
// `dispatch end` has been called (the sweep's write).
func awaitRowsEnv(t *testing.T, rows, rowsEnded string) (Env, string) {
	t.Helper()
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "rows.json"), rows)
	writeFile(t, filepath.Join(tmp, "rows-ended.json"), rowsEnded)
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\n"+
		"if [ \"$2\" = dispatches ]; then if [ -e "+tmp+"/swept ]; then cat "+tmp+"/rows-ended.json; else cat "+tmp+"/rows.json; fi; exit 0; fi\n"+
		"touch "+tmp+"/swept\n")
	home := filepath.Join(tmp, "home")
	return Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return home
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}, Dir: tmp}, home
}

// TestAwaitRowsAllEnded: every named key ended prints one ENDED line per key, in argument order.
func TestAwaitRowsAllEnded(t *testing.T) {
	t.Parallel()
	rows := `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"b","outcome":"aborted","endedAt":"2026-01-01T00:00:00Z"},{"key":"c","endedAt":null}]`
	env, _ := awaitRowsEnv(t, rows, rows)
	var o, e bytes.Buffer
	if rc := awaitRows([]string{"/x/chg", "b", "a"}, env, &o, &e); rc != 0 || o.String() != "ENDED b aborted\nENDED a completed\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
}

// TestAwaitRowsWaiting: a key still open at the timeout prints WAITING with it, exit 0.
func TestAwaitRowsWaiting(t *testing.T) {
	t.Parallel()
	rows := `[{"key":"a","outcome":"completed","endedAt":"2026-01-01T00:00:00Z"},{"key":"c","endedAt":null}]`
	env, _ := awaitRowsEnv(t, rows, rows)
	var o, e bytes.Buffer
	start := time.Now()
	if rc := awaitRows([]string{"/x/chg", "a", "c", "-timeout", "1"}, env, &o, &e); rc != 0 || o.String() != "ENDED a completed\nWAITING c\n" {
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
	env, home := awaitRowsEnv(t,
		`[{"key":"s","agentId":"aaa","sessionToken":"t","endedAt":null}]`,
		`[{"key":"s","agentId":"aaa","sessionToken":"t","outcome":"aborted","endedAt":"2026-01-01T00:00:00Z"}]`)
	sub := filepath.Join(home, ".claude", "projects", "p", "sess", "subagents")
	writeFile(t, filepath.Join(sub, "agent-aaa.meta.json"), `{"finished":true}`)
	tr := filepath.Join(sub, "agent-aaa.jsonl")
	writeFile(t, tr, "{}\n")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(tr, old, old); err != nil {
		t.Fatal(err)
	}
	var o, e bytes.Buffer
	if rc := awaitRows([]string{"/x/chg", "s", "-timeout", "1"}, env, &o, &e); rc != 0 || o.String() != "ENDED s aborted\n" {
		t.Errorf("rc %d stdout %q err %q", rc, o.String(), e.String())
	}
}

// TestAwaitRowsUnknownKey: a key the change does not record exits 2.
func TestAwaitRowsUnknownKey(t *testing.T) {
	t.Parallel()
	env, _ := awaitRowsEnv(t, `[{"key":"a","endedAt":null}]`, `[]`)
	var o, e bytes.Buffer
	if rc := awaitRows([]string{"/x/chg", "a", "zzz"}, env, &o, &e); rc != 2 {
		t.Errorf("rc %d stdout %q", rc, o.String())
	}
}
