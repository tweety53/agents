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

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
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
		meta                      string
	}{
		// The row is ended under its dispatcher's token: the main thread's, else the parent agent
		// the sidecar names -- the token the begin hook recorded it under.
		{"a nested agent's child", "Key: task-4-implementer\n", "done\n", "-key task-4-implementer -session-token cc-s1-a9 -outcome completed", "", `{"parentAgentId":"a9"}`},
		{"implementer", "Role: implementer\nKey: task-3-implementer\n", "status: DONE\ncommit: 3 " + sha + "\n", "-key task-3-implementer -session-token cc-s1-main -outcome completed -commit " + sha, "", ""},
		{"multi-repo implementer", "Key: task-3-implementer\n", "status: DONE\ncommit: 3 /wt/a=" + sha + ",/wt/b=" + strings.Repeat("b", 40) + "\n", "-outcome completed -commit " + sha, "", ""},
		// A task-<ids>-reviewer key, the retired gated review's, ends as any other row does.
		{"task reviewer key", "Key: task-1+2-reviewer\n", "task 1: clean — x\ntask 2: fix — y\n", "-key task-1+2-reviewer -session-token cc-s1-main -outcome completed", "", ""},
		{"no final message", "Key: panel-1-primary\n", "", "", "", ""},
		{"final message only in the hook payload", "Key: panel-1-primary\n", "", "-key panel-1-primary -session-token cc-s1-main -outcome completed", "done\n", ""},
		{"loop key", "Key: pipeline-fix-3-fix\n", "done", "", "", ""},
		{"no key", "hello", "done", "", "", ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			log := filepath.Join(tmp, "flow.log")
			guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
			tr := filepath.Join(tmp, "agent.jsonl")
			content := `{"type":"user","message":{"role":"user","content":` + jsonString(c.prompt) + `}}` + "\n"
			if c.final != "" {
				content += `{"type":"assistant","message":{"content":[{"type":"text","text":` + jsonString(c.final) + `}]}}` + "\n"
			}
			guardtest.WriteFile(t, tr, content)
			if c.meta != "" {
				guardtest.WriteFile(t, filepath.Join(tmp, "agent.meta.json"), c.meta)
			}
			env := core.Env{Getenv: func(k string) string {
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
// ends on a tool call with no text part; its message is the final message, so its commit line
// closes the row with that commit.
func TestEndDispatchRowReadsTheHandbackReport(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	log := filepath.Join(tmp, "flow.log")
	guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	tr := filepath.Join(tmp, "agent.jsonl")
	sha := strings.Repeat("c", 40)
	guardtest.WriteFile(t, tr, `{"type":"user","message":{"role":"user","content":"Key: operator-3\n"}}`+"\n"+
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"commit: request-9 `+sha+`\n"}}]}}`+"\n")
	env := core.Env{Getenv: func(k string) string {
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
	if got, _ := os.ReadFile(log); !strings.Contains(string(got), "-outcome completed -commit "+sha) {
		t.Errorf("flow called with %q, want -outcome completed -commit %s", got, sha)
	}
}

// TestEndDispatchRowEndsAResumedChildsRow: a child resumed by a SendMessage whose message carries
// its own `Key:` line ends that resume's row with the resume's commit — never re-ends the row its
// first prompt named, whose end the first stop already wrote.
func TestEndDispatchRowEndsAResumedChildsRow(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t2"}]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	log := filepath.Join(tmp, "flow.log")
	guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	tr := filepath.Join(tmp, "agent.jsonl")
	resumeSha := strings.Repeat("b", 40)
	guardtest.WriteFile(t, tr, `{"type":"user","message":{"content":"Role: implementer\nKey: task-3-implementer\n"}}`+"\n"+
		`{"type":"assistant","message":{"content":[{"type":"text","text":"commit: 3 `+strings.Repeat("a", 40)+`\n"}]}}`+"\n"+
		`{"type":"user","message":{"content":[{"type":"tool_result","content":"Key: task-9-implementer\n"}]}}`+"\n"+
		`{"type":"user","message":{"content":"The coordinator sent a message:\nKey: task-3-implementer-resume-1\nFIX INPUT: close-task exit 1\n"}}`+"\n"+
		`{"type":"assistant","message":{"content":[{"type":"text","text":"commit: 3 `+resumeSha+`\n"}]}}`+"\n")
	env := core.Env{Getenv: func(k string) string {
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
	if got, _ := os.ReadFile(log); !strings.Contains(string(got), "-key task-3-implementer-resume-1 ") || !strings.Contains(string(got), "-commit "+resumeSha) {
		t.Errorf("flow called with %q, want the resume row ended with its commit", got)
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
	guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	handback := func(key string) string {
		return `{"type":"user","message":{"role":"user","content":"Key: ` + key + `\n"}}` + "\n" +
			`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"status: DONE"}}]}}` + "\n"
	}
	write := func(name, content string, age time.Duration) {
		f := filepath.Join(tmp, name)
		guardtest.WriteFile(t, f, content)
		old := time.Now().Add(-age)
		if err := os.Chtimes(f, old, old); err != nil {
			t.Fatal(err)
		}
	}
	write("agent-old.jsonl", handback("task-1-implementer"), time.Hour)
	write("agent-young.jsonl", handback("task-2-implementer"), time.Minute)
	write("agent-noreport.jsonl", `{"type":"user","message":{"role":"user","content":"Key: task-3-implementer\n"}}`+"\n", time.Hour)
	write("agent-swept.jsonl", handback("task-4-implementer"), time.Hour)
	guardtest.WriteFile(t, filepath.Join(tmp, "agent-swept.jsonl.swept"), "")
	env := core.Env{Getenv: func(k string) string {
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

// TestSweepRetriesATransientMiss: flowd naming no change for the session leaves no sidecar, so
// a later sweep still closes the row.
func TestSweepRetriesATransientMiss(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[]}`))
	}))
	t.Cleanup(srv.Close)
	tmp := t.TempDir()
	f := filepath.Join(tmp, "agent-a.jsonl")
	guardtest.WriteFile(t, f, `{"type":"user","message":{"role":"user","content":"Key: panel-0-x\n"}}`+"\n"+
		`{"type":"assistant","message":{"content":[{"type":"tool_use","name":"SubagentHandback","input":{"message":"done"}}]}}`+"\n")
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(f, old, old); err != nil {
		t.Fatal(err)
	}
	env := core.Env{Getenv: func(k string) string {
		if k == "FLOW_ADDR" {
			return srv.URL
		}
		return os.Getenv(k)
	}, Dir: tmp}
	var o, e bytes.Buffer
	sweepDispatchRows([]string{tmp, "s1", tmp}, env, &o, &e)
	if _, err := os.Stat(f + ".swept"); err == nil {
		t.Error("a transient miss wrote the .swept sidecar; the row would never be retried")
	}
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// TestEndDispatchRowEndsAKeylessAgentsRow: a transcript with no Key: line is ended under the key
// the begin hook derived from its meta sidecar (request 211).
func TestEndDispatchRowEndsAKeylessAgentsRow(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t1"}]}`))
	}))
	t.Cleanup(srv.Close)
	for name, c := range map[string]struct{ description, agentType, prompt, want string }{
		"explore word": {"plan-1 round — explore-2 survey", "", "look", "-key explore-2 "},
		"tool use id":  {"live check", "", "look", "-key agent-tu1 "},
		"row key word": {"panel-0-primary+principles", "", "look", "-key panel-0-primary+principles "},
		// A planner's Explore agent described `plan-1 …` is begun agent-<toolUseId>, never the
		// planner's own plan-1 row (request 286).
		"planning agent under plan-1": {"plan-1 explore runner seam", "Explore", "look", "-key agent-tu1 "},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tmp := t.TempDir()
			log := filepath.Join(tmp, "flow.log")
			guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
			tr := filepath.Join(tmp, "agent-a1.jsonl")
			guardtest.WriteFile(t, tr, `{"type":"user","message":{"content":`+jsonString(c.prompt)+`}}`+"\n")
			guardtest.WriteFile(t, filepath.Join(tmp, "agent-a1.meta.json"), `{"description":`+jsonString(c.description)+`,"agentType":`+jsonString(c.agentType)+`,"toolUseId":"tu1"}`)
			env := core.Env{Getenv: func(k string) string {
				switch k {
				case "FLOW_ADDR":
					return srv.URL
				case "PATH":
					return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
				}
				return os.Getenv(k)
			}, Dir: tmp}
			var o, e bytes.Buffer
			endDispatchRow([]string{tr, "s1", tmp, "done"}, env, &o, &e)
			if got, _ := os.ReadFile(log); !strings.Contains(string(got), c.want) {
				t.Errorf("flow called with %q, want %q (out %s)", got, c.want, o.String())
			}
		})
	}
}

// TestSweepAbortsAbandonedVisualVerifyRows: a visual-verify row still open 30 minutes after its
// begin is ended aborted under its own token — end-dispatch-row leaves these rows to
// visual-verify-pass close, which an interrupted or superseded round never reaches, and an agent
// that never launched has no transcript to sweep. A younger row, a closed row and any other
// role's open row are left alone.
func TestSweepAbortsAbandonedVisualVerifyRows(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	log := filepath.Join(tmp, "flow.log")
	ago := func(d time.Duration) string { return time.Now().Add(-d).UTC().Format(time.RFC3339) }
	rows := `[` +
		`{"key":"visual-verify-3-gymie-s1of4","role":"visual-verifier","sessionToken":"mf-old","startedAt":"` + ago(time.Hour) + `"},` +
		`{"key":"visual-verify-gymie","role":"visual-verifier","sessionToken":"mf-old","startedAt":"` + ago(time.Hour) + `","endedAt":"` + ago(50*time.Minute) + `"},` +
		`{"key":"visual-verify-4-gymie","role":"visual-verifier","sessionToken":"mf-t1","startedAt":"` + ago(5*time.Minute) + `"},` +
		`{"key":"panel-fix-1","role":"panel-fix","sessionToken":"mf-old","startedAt":"` + ago(time.Hour) + `"}]`
	srv := fakeFlowdRows(t, func() string { return rows }, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"stageRuns":[{"changeName":"c1","sessionToken":"mf-t1"}]}`))
	})
	guardtest.WriteExec(t, filepath.Join(tmp, "bin", "flow"), "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >>"+log+"\n")
	cwd := gitWorktree(t, tmp, "repo")
	env := core.Env{Getenv: func(k string) string {
		switch k {
		case "FLOW_ADDR", "FLOW_RECORDS_ADDR":
			return srv.URL
		case "PATH":
			return filepath.Join(tmp, "bin") + ":" + os.Getenv("PATH")
		}
		return os.Getenv(k)
	}, Dir: tmp}
	var o, e bytes.Buffer
	if rc := sweepDispatchRows([]string{tmp, "s1", cwd}, env, &o, &e); rc != 0 {
		t.Fatalf("rc %d: %s", rc, e.String())
	}
	got, _ := os.ReadFile(log)
	if n := strings.Count(string(got), "record dispatch end"); n != 1 ||
		!strings.Contains(string(got), "-key visual-verify-3-gymie-s1of4 -session-token mf-old -outcome aborted") {
		t.Errorf("flow calls = %q; want exactly one end, aborting visual-verify-3-gymie-s1of4 under mf-old", got)
	}
}
