package repotest

// TestFlowStopGuardHook pins hooks/flow-stop-guard.py, the Stop hook that
// refuses a /flow turn end mid-phase (kan-1024). Each subtest starts a stub
// store answering /api/v1/stage-runs and /api/v1/changes/find, writes the
// change's tasks.md under the stage row's repoRoot, and pipes a Stop payload
// to the hook with HOME in a temp dir (the fingerprint lives under it).

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const stopChange = "fixture-change"

type stopFixture struct {
	state    string // record state; "" = no record
	open     bool   // last stage run has no endedAt
	ticked   int
	total    int
	noRun    bool
	status   int    // store status; 0 = 200
	lastUser string // last operator message
	lastTool string // last assistant tool use
}

func runStopGuard(t *testing.T, f stopFixture, home string, active bool) string {
	t.Helper()
	root := t.TempDir()
	plan := filepath.Join(root, "spectre", "changes", stopChange)
	if err := os.MkdirAll(plan, 0o755); err != nil {
		t.Fatal(err)
	}
	var tasks strings.Builder
	for i := 1; i <= f.total; i++ {
		mark := " "
		if i <= f.ticked {
			mark = "x"
		}
		fmt.Fprintf(&tasks, "- [%s] %d. task\n  - [ ] step\n", mark, i)
	}
	if err := os.WriteFile(filepath.Join(plan, "tasks.md"), []byte(tasks.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	ended := `,"endedAt":"2026-10-27T10:00:00Z"`
	if f.open {
		ended = ""
	}
	row := fmt.Sprintf(`{"stageRunId":7,"command":"/flow","stage":"flow.implement","startedAt":"2026-10-27T09:00:00Z"%s,"changeName":%q,"projectKey":"p","repoRoot":%q}`, ended, stopChange, root)
	if f.noRun {
		row = ""
	}
	recs := ""
	if f.state != "" {
		recs = fmt.Sprintf(`{"projectKey":"p","name":%q,"state":%q,"reviewBeforeLanding":true}`, stopChange, f.state)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f.status != 0 {
			w.WriteHeader(f.status)
			return
		}
		switch r.URL.Path {
		case "/api/v1/stage-runs":
			_, _ = fmt.Fprintf(w, `{"stageRuns":[%s]}`, row)
		case "/api/v1/changes/find":
			_, _ = fmt.Fprintf(w, `{"records":[%s]}`, recs)
		default:
			w.WriteHeader(404)
		}
	}))
	t.Cleanup(srv.Close)

	user := f.lastUser
	if user == "" {
		user = "/flow " + stopChange
	}
	transcript := fmt.Sprintf(`{"type":"user","message":{"content":%q}}`+"\n", user)
	if f.lastTool != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":%q,"input":{}}]}}`+"\n", f.lastTool)
	}
	tp := filepath.Join(t.TempDir(), "sess-1.jsonl")
	if err := os.WriteFile(tp, []byte(transcript), 0o644); err != nil {
		t.Fatal(err)
	}
	stdin := fmt.Sprintf(`{"session_id":"sess-1","transcript_path":%q,"stop_hook_active":%t}`, tp, active)
	out, errOut, rc := runHook(t, "flow-stop-guard.py", []byte(stdin), "FLOW_ADDR="+srv.URL, "HOME="+home)
	if rc != 0 || errOut != "" {
		t.Fatalf("rc=%d stderr=%s", rc, errOut)
	}
	return out
}

func TestFlowStopGuardHook(t *testing.T) {
	t.Parallel()
	block := func(t *testing.T, out, needle string) {
		t.Helper()
		if !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, needle) {
			t.Errorf("expected block naming %q, got %q", needle, out)
		}
	}
	allow := func(t *testing.T, out string) {
		t.Helper()
		if out != "" {
			t.Errorf("expected allow (no output), got %q", out)
		}
	}
	started := func(ticked, total int, open bool) stopFixture {
		return stopFixture{state: "STARTED", ticked: ticked, total: total, open: open}
	}
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"a: open stage blocks naming it", func(t *testing.T) {
			block(t, runStopGuard(t, started(3, 7, true), t.TempDir(), false), "flow.implement")
		}},
		{"b: STARTED all ticked blocks naming flow.write-in-progress", func(t *testing.T) {
			block(t, runStopGuard(t, started(7, 7, false), t.TempDir(), false), "flow.write-in-progress")
		}},
		{"c: STARTED no tasks allows", func(t *testing.T) {
			allow(t, runStopGuard(t, started(0, 0, false), t.TempDir(), false))
		}},
		{"d: STARTED 0/7 no open stage allows", func(t *testing.T) {
			allow(t, runStopGuard(t, started(0, 7, false), t.TempDir(), false))
		}},
		{"e: IN_PROGRESS allows", func(t *testing.T) {
			allow(t, runStopGuard(t, stopFixture{state: "IN_PROGRESS", ticked: 7, total: 7, open: true}, t.TempDir(), false))
		}},
		{"f: FINISHED allows", func(t *testing.T) {
			allow(t, runStopGuard(t, stopFixture{state: "FINISHED", ticked: 7, total: 7, open: true}, t.TempDir(), false))
		}},
		{"g: AskUserQuestion allows", func(t *testing.T) {
			f := started(7, 7, true)
			f.lastTool = "AskUserQuestion"
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"h: operator said stop allows", func(t *testing.T) {
			f := started(7, 7, true)
			f.lastUser = "stop here"
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"i: stop_hook_active with unchanged fingerprint allows", func(t *testing.T) {
			home := t.TempDir()
			block(t, runStopGuard(t, started(3, 7, true), home, false), "flow.implement")
			allow(t, runStopGuard(t, started(3, 7, true), home, true))
		}},
		{"j: stop_hook_active after a stage ended blocks", func(t *testing.T) {
			home := t.TempDir()
			block(t, runStopGuard(t, started(7, 7, true), home, false), "flow.implement")
			block(t, runStopGuard(t, started(7, 7, false), home, true), "flow.write-in-progress")
		}},
		{"k: store 500 allows", func(t *testing.T) {
			f := started(7, 7, true)
			f.status = 500
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"l: no /flow row allows", func(t *testing.T) {
			f := started(7, 7, true)
			f.noRun = true
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}
