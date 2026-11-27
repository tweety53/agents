package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// traceSpansFixture is flowd's answer for one small change: a request, two
// overlapping stage runs, one dispatch whose agent left a transcript, a
// guard run, a verify run, and a stage run still open; the dispatch and the
// first stage run carry the time split flowd stored for them.
const traceSpansFixture = `[
 {"kind":"dispatch","lane":"operator-asked operator-1","name":"operator-asked operator-1","agentId":"a1","outcome":"done","start":"2026-01-01T00:00:10Z","end":"2026-01-01T00:01:30Z","split":{"llmMs":60000,"toolMs":10000,"humanMs":0,"byTool":{"Bash git":5000,"Bash go":4000,"Read":1000}}},
 {"kind":"guard","lane":"guard","name":"check-a","outcome":"clear","start":"2026-01-01T00:00:19.5Z","end":"2026-01-01T00:00:20Z"},
 {"kind":"request","lane":"request","name":"request 1: add trace","outcome":"landed","start":"2026-01-01T00:00:00Z","end":"2026-01-01T00:03:20Z"},
 {"kind":"stage","lane":"stage","name":"flow.implement #1","outcome":"completed","start":"2026-01-01T00:00:00Z","end":"2026-01-01T00:01:40Z","split":{"llmMs":70000,"toolMs":25000,"humanMs":5000,"byTool":{"Agent":25000}}},
 {"kind":"stage","lane":"stage","name":"flow.verify #1","start":"2026-01-01T00:01:30Z"},
 {"kind":"verify","lane":"verify","name":"lint","outcome":"exit 0","start":"2026-01-01T00:01:40Z","end":"2026-01-01T00:02:10Z"}
]`

// traceTranscriptFixture is agent a1's transcript: two parallel calls, one
// running past the dispatch's end (clamped to it), one started after it
// (dropped).
const traceTranscriptFixture = `{"type":"assistant","timestamp":"2026-01-01T00:00:20Z","message":{"content":[{"type":"tool_use","id":"t1","name":"Bash","input":{"command":"go test ./..."}},{"type":"tool_use","id":"t2","name":"Read","input":{"file_path":"/a/b.go"}}]}}
{"type":"user","timestamp":"2026-01-01T00:00:21Z","message":{"content":[{"type":"tool_result","tool_use_id":"t2"}]}}
{"type":"user","timestamp":"2026-01-01T00:00:25Z","message":{"content":[{"type":"tool_result","tool_use_id":"t1"}]}}
{"type":"assistant","timestamp":"2026-01-01T00:01:25Z","message":{"content":[{"type":"tool_use","id":"t3","name":"Bash","input":{"command":"git push"}}]}}
{"type":"user","timestamp":"2026-01-01T00:01:39Z","message":{"content":[{"type":"tool_result","tool_use_id":"t3"}]}}
{"type":"assistant","timestamp":"2026-01-01T00:01:35Z","message":{"content":[{"type":"tool_use","id":"t4","name":"Bash","input":{"command":"late"}}]}}
{"type":"user","timestamp":"2026-01-01T00:01:36Z","message":{"content":[{"type":"tool_result","tool_use_id":"t4"}]}}
`

type traceEvent struct {
	Name string         `json:"name"`
	Cat  string         `json:"cat"`
	Ph   string         `json:"ph"`
	Ts   *int64         `json:"ts"`
	Dur  *int64         `json:"dur"`
	Pid  *int           `json:"pid"`
	Tid  *int           `json:"tid"`
	Args map[string]any `json:"args"`
}

// TestStatsTrace pins `flow stats trace` against a golden Chrome trace:
// the project and change reach flowd's query, every complete event carries
// name/ph/ts/dur/pid/tid with a non-negative duration, each tool span sits
// inside its dispatch, every lane is named, -no-transcripts drops the tool
// lanes, and -o writes the same JSON to a file.
func TestStatsTrace(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	root := t.TempDir()
	dir := filepath.Join(root, "proj", "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent-a1.jsonl"), []byte(traceTranscriptFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	env["FLOW_TRANSCRIPTS_DIR"] = root
	var query string
	srv := httptest.NewServer(genuineDaemon(func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Path + "?" + r.URL.RawQuery
		_, _ = w.Write([]byte(traceSpansFixture))
	}))
	defer srv.Close()
	call := func(args ...string) (int, string, string) {
		var out, errOut bytes.Buffer
		code := run(context.Background(), append([]string{"stats", "trace", "-addr", srv.URL, "-project", "p"}, args...), env.get, strings.NewReader(""), &out, &errOut)
		return code, out.String(), errOut.String()
	}
	code, out, errOut := call("mine")
	if code != 0 || query != "/api/v1/stats/trace?change=mine&project=p" {
		t.Fatalf("code %d query %q stderr %s", code, query, errOut)
	}
	golden, err := os.ReadFile("testdata/trace.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	if out != string(golden) {
		t.Errorf("trace differs from testdata/trace.golden.json; got:\n%s", out)
	}
	var doc struct {
		TraceEvents []traceEvent `json:"traceEvents"`
	}
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	lanes := map[int]string{}
	dispatch := map[string][2]int64{}
	for _, e := range doc.TraceEvents {
		if e.Ph == "M" && e.Name == "thread_name" {
			lanes[*e.Tid] = e.Args["name"].(string)
		}
	}
	for _, e := range doc.TraceEvents {
		if e.Ph != "X" {
			continue
		}
		if e.Name == "" || e.Ts == nil || e.Dur == nil || e.Pid == nil || e.Tid == nil || *e.Dur < 0 {
			t.Errorf("bad complete event %+v", e)
			continue
		}
		if lanes[*e.Tid] == "" {
			t.Errorf("event %q on unnamed tid %d", e.Name, *e.Tid)
		}
		if e.Cat == "dispatch" {
			dispatch[lanes[*e.Tid]] = [2]int64{*e.Ts, *e.Ts + *e.Dur}
		}
	}
	tools := 0
	for _, e := range doc.TraceEvents {
		if e.Cat != "tool" {
			continue
		}
		tools++
		d, ok := dispatch[strings.TrimSuffix(strings.TrimSuffix(lanes[*e.Tid], " (2)"), " tools")]
		if !ok || *e.Ts < d[0] || *e.Ts+*e.Dur > d[1] {
			t.Errorf("tool span %q [%d,+%d] outside its dispatch %v", e.Name, *e.Ts, *e.Dur, d)
		}
	}
	if tools != 3 {
		t.Errorf("tool spans = %d, want 3", tools)
	}
	if _, out, _ := call("-no-transcripts", "mine"); strings.Contains(out, `"cat": "tool"`) || strings.Contains(out, " tools") {
		t.Errorf("-no-transcripts still draws tool spans:\n%s", out)
	}
	file := filepath.Join(t.TempDir(), "t.json")
	if code, out, _ := call("-o", file, "mine"); code != 0 || !strings.Contains(out, file) {
		t.Errorf("-o = %d %q", code, out)
	}
	if b, err := os.ReadFile(file); err != nil || string(b) != string(golden) {
		t.Errorf("-o file differs from the golden: %v", err)
	}
	if code, _, _ := call(); code != 2 {
		t.Errorf("no change = %d, want 2", code)
	}
}
