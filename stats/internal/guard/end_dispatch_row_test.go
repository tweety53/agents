package guard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestEndDispatchRow: a stopping subagent's row is ended under the session's
// open change, its outcome and commit read off the final message; a loop key
// and a prompt with no Key: line end nothing.
func TestEndDispatchRow(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("session_id") != "s1" {
			_, _ = w.Write([]byte(`{"stageRuns":[]}`))
			return
		}
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"pipeline-fix-x-1","sessionToken":"pipeline-fix-x-1"},{"changeName":"c1","sessionToken":"mf-t1","endedAt":"2026-01-01T00:00:00Z"},{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	sha := strings.Repeat("a", 40)
	for _, c := range []struct {
		name, prompt, final, want string
		payload                   string
	}{
		{"implementer", "Role: implementer\nKey: task-3-implementer\n", "status: DONE\ncommit: 3 " + sha + "\n", "-key task-3-implementer -session-token mf-t2 -outcome completed -commit " + sha, ""},
		{"gated fix", "Key: task-1+2-reviewer\n", "task 1: clean — x\ntask 2: fix — y\n", "-key task-1+2-reviewer -session-token mf-t2 -outcome fix", ""},
		{"gated clean", "Key: task-1-reviewer\n", "task 1: clean — x\n", "-outcome clean", ""},
		{"no final message", "Key: panel-1-primary\n", "", "", ""},
		{"gated verdict only in the hook payload", "Key: task-1-reviewer\n", "", "-outcome fix", "task 1: fix — y\n"},
		{"gated verdicts as a bullet list", "Key: task-1+2-reviewer-fix-2\n", "## Verdict\nOne still needs a fix.\n\n- task 1: clean — x\n- task 2: fix — y\n", "-outcome fix", ""},
		{"loop key", "Key: pipeline-fix-3-fix\n", "done", "", ""},
		{"no key", "hello", "done", "", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			log := filepath.Join(tmp, "flow.log")
			writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
			tr := filepath.Join(tmp, "agent.jsonl")
			content := `{"type":"user","message":{"role":"user","content":` + jsonString(c.prompt) + `}}` + "\n"
			if c.final != "" {
				content += `{"type":"assistant","message":{"content":[{"type":"text","text":` + jsonString(c.final) + `}]}}` + "\n"
			}
			writeFile(t, tr, content)
			env := Env{Getenv: func(k string) string {
				switch k {
				case "FLOW_ADDR":
					return srv.URL
				case "PATH":
					return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
				}
				return os.Getenv(k)
			}, Dir: tmp}
			var o, e bytes.Buffer
			if rc := endDispatchRow([]string{tr, "s1", tmp, c.payload}, env, &o, &e); rc != 0 {
				t.Fatalf("rc %d: %s%s", rc, o.String(), e.String())
			}
			got, _ := os.ReadFile(log)
			switch {
			case c.want == "" && len(got) != 0:
				t.Errorf("ended a row: %s", got)
			case c.want != "" && !strings.Contains(string(got), c.want):
				t.Errorf("flow called with %q, want it to carry %q (out %s)", got, c.want, o.String())
			}
		})
	}
}

// TestEndDispatchRowReadsTheHandbackReport: a child that reports through the SubagentHandback tool
// ends on a tool call with no text part; its message is the final message, so a gated reviewer's
// verdict lines close the row `fix`, not `aborted`.
func TestEndDispatchRowReadsTheHandbackReport(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	log := filepath.Join(tmp, "flow.log")
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	tr := filepath.Join(tmp, "agent.jsonl")
	writeFile(t, tr, `{"type":"user","message":{"role":"user","content":"Key: task-1-reviewer\n"}}`+"\n"+
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"task 1: fix — y\n"}}]}}`+"\n")
	env := Env{Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR":
			return srv.URL
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}, Dir: tmp}
	var o, e bytes.Buffer
	if rc := endDispatchRow([]string{tr, "s1", tmp}, env, &o, &e); rc != 0 {
		t.Fatalf("rc %d: %s%s", rc, o.String(), e.String())
	}
	if got, _ := os.ReadFile(log); !strings.Contains(string(got), "-outcome fix") {
		t.Errorf("flow called with %q, want -outcome fix", got)
	}
}

// TestSweepDispatchRows: a transcript 30+ minutes old with a handback report is closed from it, once;
// a young one, one with no report and an already swept one are left alone.
func TestSweepDispatchRows(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	log := filepath.Join(tmp, "flow.log")
	writeExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	handback := func(key string) string {
		return `{"type":"user","message":{"role":"user","content":"Key: ` + key + `\n"}}` + "\n" +
			`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"status: DONE"}}]}}` + "\n"
	}
	write := func(name, content string, age time.Duration) {
		f := filepath.Join(tmp, name)
		writeFile(t, f, content)
		old := time.Now().Add(-age)
		if err := os.Chtimes(f, old, old); err != nil {
			t.Fatal(err)
		}
	}
	write("agent-old.jsonl", handback("task-1-implementer"), time.Hour)
	write("agent-young.jsonl", handback("task-2-implementer"), time.Minute)
	write("agent-noreport.jsonl", `{"type":"user","message":{"role":"user","content":"Key: task-3-implementer\n"}}`+"\n", time.Hour)
	write("agent-swept.jsonl", handback("task-4-implementer"), time.Hour)
	writeFile(t, filepath.Join(tmp, "agent-swept.jsonl.swept"), "")
	env := Env{Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR":
			return srv.URL
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}, Dir: tmp}
	for i := 0; i < 2; i++ { // the second sweep finds nothing new
		var o, e bytes.Buffer
		if rc := sweepDispatchRows([]string{tmp, "s1", tmp}, env, &o, &e); rc != 0 {
			t.Fatalf("rc %d: %s", rc, e.String())
		}
	}
	got, _ := os.ReadFile(log)
	if n := strings.Count(string(got), "record dispatch end"); n != 1 || !strings.Contains(string(got), "-key task-1-implementer") {
		t.Errorf("flow calls = %q; want exactly one end, for task-1-implementer", got)
	}
}

// TestCheckGatedVerdicts: only a gated reviewer whose report has no verdict line is refused.
func TestCheckGatedVerdicts(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, prompt, final string
		want                int
	}{
		{"gated, no verdict", "Key: task-1-reviewer\n", "all good", 1},
		{"gated, verdict", "Key: task-1-reviewer\n", "task 1: clean — x", 0},
		{"gated, bulleted verdict", "Key: task-1-reviewer\n", "- task 1: fix — x", 0},
		{"not gated", "Key: task-1-implementer\n", "all good", 0},
		{"no report yet", "Key: task-1-reviewer\n", "", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tr := filepath.Join(t.TempDir(), "agent.jsonl")
			content := `{"type":"user","message":{"role":"user","content":` + jsonString(c.prompt) + `}}` + "\n"
			if c.final != "" {
				content += `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":` + jsonString(c.final) + `}}]}}` + "\n"
			}
			writeFile(t, tr, content)
			var o, e bytes.Buffer
			if rc := checkGatedVerdicts([]string{tr}, Env{Getenv: os.Getenv}, &o, &e); rc != c.want {
				t.Errorf("rc %d, want %d (%s)", rc, c.want, o.String())
			}
		})
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
