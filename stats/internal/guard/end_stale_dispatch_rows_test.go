package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestEndStaleDispatchRows: an open row whose agent finished and went quiet 10+ minutes ago is
// ended aborted; unfinished, recently written, id-less and a failing row never stop the others.
func TestEndStaleDispatchRows(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	sub := filepath.Join(home, ".claude", "projects", "p", "sess", "subagents")
	agent := func(id string, finished bool, age time.Duration) {
		fin := "false"
		if finished {
			fin = "true"
		}
		writeFile(t, filepath.Join(sub, "agent-"+id+".meta.json"), `{"finished":`+fin+`}`)
		tr := filepath.Join(sub, "agent-"+id+".jsonl")
		writeFile(t, tr, "{}\n")
		old := time.Now().Add(-age)
		if err := os.Chtimes(tr, old, old); err != nil {
			t.Fatal(err)
		}
	}
	agent("aaa", true, 20*time.Minute)
	agent("bbb", false, 20*time.Minute)
	agent("ccc", true, time.Minute)
	agent("ddd", true, 20*time.Minute)
	agent("eee", true, 20*time.Minute)
	rows := `[{"key":"stale","agentId":"aaa","sessionToken":"t","endedAt":null},
{"key":"alive","agentId":"bbb","sessionToken":"t","endedAt":null},
{"key":"recent","agentId":"ccc","sessionToken":"t","endedAt":null},
{"key":"noid","agentId":"","sessionToken":"t","endedAt":null},
{"key":"boom","agentId":"ddd","sessionToken":"t","endedAt":null},
{"key":"closed","agentId":"eee","sessionToken":"t","endedAt":"2026-01-01T00:00:00Z"},
{"key":"stale2","agentId":"eee","sessionToken":"t","endedAt":null}]`
	srv := fakeFlowdRows(t, func() string { return rows }, nil)
	log := filepath.Join(tmp, "flow.log")
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\n"+
		"printf '%s\\n' \"$*\" >>"+log+"\n"+
		"case \"$*\" in *-key\\ boom*) echo nope >&2; exit 1;; esac\n")
	env := Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return home
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}
	wt := gitWorktree(t, tmp, "chg")
	for i := 0; i < 2; i++ {
		var o, e bytes.Buffer
		if rc := endStaleDispatchRows([]string{wt}, env, &o, &e); rc != 0 {
			t.Fatalf("rc %d", rc)
		}
		if i == 0 && o.String() != "ENDED stale aaa\nENDED stale2 eee\n" {
			t.Errorf("stdout %q", o.String())
		}
		if i == 0 && !strings.Contains(e.String(), "boom") {
			t.Errorf("stderr %q does not name the failed row", e.String())
		}
	}
	got, _ := os.ReadFile(log)
	for _, k := range []string{"alive", "recent", "noid", "closed"} {
		if strings.Contains(string(got), "-key "+k+" ") {
			t.Errorf("ended %s: %s", k, got)
		}
	}
	if !strings.Contains(string(got), "-key stale -session-token t -outcome aborted -agent-id aaa -ended-at") {
		t.Errorf("stale row not ended aborted: %s", got)
	}
}

// TestEndStaleDispatchRowsIdempotent: a second run over rows the store now reports ended ends nothing.
func TestEndStaleDispatchRowsIdempotent(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	srv := fakeFlowdRows(t, func() string { return `[{"key":"k","agentId":"aaa","endedAt":"2026-01-01T00:00:00Z"}]` }, nil)
	env := Env{Getenv: func(k string) string {
		if k == "FLOW_ADDR" || k == "FLOW_RECORDS_ADDR" {
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}
	var o, e bytes.Buffer
	if rc := endStaleDispatchRows([]string{gitWorktree(t, tmp, "chg")}, env, &o, &e); rc != 0 || o.Len() != 0 || e.Len() != 0 {
		t.Errorf("rc %d out %q err %q", rc, o.String(), e.String())
	}
}

// TestEndStaleDispatchRowsReadsTheHandbackReport: a finished agent that delivered its report through
// SubagentHandback is ended with the outcome and commit its report carries, never aborted (request 245).
func TestEndStaleDispatchRowsReadsTheHandbackReport(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	sub := filepath.Join(home, ".claude", "projects", "p", "sess", "subagents")
	sha := strings.Repeat("a", 40)
	for id, report := range map[string]string{
		"aaa": "done\ncommit: t1 " + sha,
		"bbb": "- task 3: fix",
	} {
		writeFile(t, filepath.Join(sub, "agent-"+id+".meta.json"), `{"finished":true}`)
		tr := filepath.Join(sub, "agent-"+id+".jsonl")
		writeFile(t, tr, `{"type":"user","message":{"content":"go"}}`+"\n"+
			`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":`+jsonString(report)+`}}]}}`+"\n")
		old := time.Now().Add(-20 * time.Minute)
		if err := os.Chtimes(tr, old, old); err != nil {
			t.Fatal(err)
		}
	}
	srv := fakeFlowdRows(t, func() string {
		return `[{"key":"panel-0-primary","agentId":"aaa","sessionToken":"t","endedAt":null},
{"key":"task-3-reviewer","agentId":"bbb","sessionToken":"t","endedAt":null}]`
	}, nil)
	log := filepath.Join(tmp, "flow.log")
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\n"+
		"printf '%s\\n' \"$*\" >>"+log+"\n")
	env := Env{Getenv: func(k string) string {
		switch k {
		case "HOME":
			return home
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}
	var o, e bytes.Buffer
	endStaleDispatchRows([]string{gitWorktree(t, tmp, "chg")}, env, &o, &e)
	got, _ := os.ReadFile(log)
	for _, want := range []string{"-key panel-0-primary -session-token t -outcome completed -agent-id aaa -ended-at", "-commit " + sha,
		"-key task-3-reviewer -session-token t -outcome completed -agent-id bbb"} {
		if !strings.Contains(string(got), want) {
			t.Errorf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(string(got), "aborted") {
		t.Errorf("a handed-back agent ended aborted: %s", got)
	}
}
