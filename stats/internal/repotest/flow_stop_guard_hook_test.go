package repotest

// TestFlowStopGuardHook pins hooks/flow-stop-guard.py, the Stop hook that
// refuses a /flow turn end mid-phase (kan-1024). Each subtest starts a stub
// store answering /api/v1/stage-runs and /api/v1/changes/find, writes the
// change's tasks.md under the stage row's repoRoot, and pipes a Stop payload
// to the hook with HOME in a temp dir.

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
	metaUser string // isMeta user entry after the operator message (hook feedback, skill text)
	lastText string // last assistant text
	oldText  string // assistant text from an earlier turn, before the operator message
	launched bool   // a background agent launch after the operator message
	notified bool   // that agent's task-notification
	bgBash   bool   // a Bash command moved to the background after the operator message
	bgDone   bool   // that command's task-notification
	refusal  string // error text of the turn's last tool result, an Edit's, after every other entry
	blocks   int    // refusal-stop blocks the hook already fed back since the operator message
	loopRun  bool   // an open loop stage run, newer than the change's own
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
	if f.loopRun {
		row = `{"stageRunId":8,"command":"/flow","stage":"flow.pipeline-fix","startedAt":"2026-10-27T09:30:00Z","changeName":"pipeline-fix-fixture-change-81","projectKey":"p"},` + row
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
	transcript := ""
	if f.oldText != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":%q}]}}`+"\n", f.oldText)
	}
	transcript += fmt.Sprintf(`{"type":"user","message":{"content":%q}}`+"\n", user)
	if f.metaUser != "" {
		transcript += fmt.Sprintf(`{"type":"user","isMeta":true,"message":{"content":%q}}`+"\n", f.metaUser)
	}
	if f.lastTool != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":%q,"input":{}}]}}`+"\n", f.lastTool)
	}
	if f.launched {
		transcript += `{"type":"user","toolUseResult":{"isAsync":true,"status":"async_launched","agentId":"a1b2c3d4e5f60718"}}` + "\n"
	}
	if f.notified {
		transcript += `{"type":"user","message":{"content":"<task-notification>\n<task-id>a1b2c3d4e5f60718</task-id>\n<status>completed</status>\n</task-notification>"}}` + "\n"
	}
	if f.bgBash {
		transcript += `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_bg1","name":"Bash","input":{}}]}}` + "\n"
		transcript += `{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_bg1","content":"Command running in background with ID: b7egj759e. Output is being written to: /tmp/x.output"}]}}` + "\n"
	}
	if f.bgDone {
		transcript += `{"type":"user","message":{"content":"<task-notification>\n<task-id>b7egj759e</task-id>\n<status>completed</status>\n</task-notification>"}}` + "\n"
	}
	for i := 0; i < f.blocks; i++ {
		transcript += `{"type":"user","isMeta":true,"message":{"content":"Stop hook feedback:\nflow-stop-guard: the turn ends on a refusal -- fix the cause and continue"}}` + "\n"
	}
	if f.refusal != "" {
		transcript += `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"toolu_ref1","name":"Edit","input":{}}]}}` + "\n"
		transcript += fmt.Sprintf(`{"type":"user","message":{"content":[{"type":"tool_result","tool_use_id":"toolu_ref1","is_error":true,"content":%q}]}}`+"\n", f.refusal)
	}
	if f.lastText != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":%q}]}}`+"\n", f.lastText)
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

// Cost: 1.78–2.23 s (load 8–26; no idle window) — twenty-five runs of the
// real stop hook, each a python3 start reading its own fake store; the shared
// flow-guard build is TestMain's, not this test's. Cutting a run drops a stop
// shape the gate proves.
// smoke: core — the stop gate
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
		{"g2: a background agent still running allows", func(t *testing.T) {
			f := started(3, 7, true)
			f.launched = true
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"g3: a background agent that notified its end blocks", func(t *testing.T) {
			f := started(3, 7, true)
			f.launched, f.notified = true, true
			block(t, runStopGuard(t, f, t.TempDir(), false), "flow.implement")
		}},
		{"g4: a background Bash command still running allows", func(t *testing.T) {
			f := started(3, 7, true)
			f.bgBash = true
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"g5: a background Bash command that notified its end blocks", func(t *testing.T) {
			f := started(3, 7, true)
			f.bgBash, f.bgDone = true, true
			block(t, runStopGuard(t, f, t.TempDir(), false), "flow.implement")
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
		{"j: stop_hook_active after a stage ended blocks", func(t *testing.T) {
			home := t.TempDir()
			block(t, runStopGuard(t, started(7, 7, true), home, false), "flow.implement")
			block(t, runStopGuard(t, started(7, 7, false), home, true), "flow.write-in-progress")
		}},
		{"m: isMeta Stop hook feedback then tool use blocks", func(t *testing.T) {
			f := started(3, 7, true)
			f.metaUser = "Stop hook feedback:\nflow-stop-guard: stage flow.implement is still open"
			f.lastTool = "Bash"
			block(t, runStopGuard(t, f, t.TempDir(), false), "flow.implement")
		}},
		{"n: stop not leading the operator message blocks", func(t *testing.T) {
			for _, msg := range []string{"don't stop until it's done", "set the instruction to stop for manual test after the implementation"} {
				f := started(3, 7, true)
				f.lastUser = msg
				block(t, runStopGuard(t, f, t.TempDir(), false), "flow.implement")
			}
		}},
		{"o: operator said pause or wait allows", func(t *testing.T) {
			for _, msg := range []string{"  Pause.", "wait, let me check"} {
				f := started(3, 7, true)
				f.lastUser = msg
				allow(t, runStopGuard(t, f, t.TempDir(), false))
			}
		}},
		{"p: Finished handoff without Prompt tuning blocks naming it", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", lastText: "## Finished\n\n**Change:** x\n**Learning:** /a/n.md\n"}
			block(t, runStopGuard(t, f, t.TempDir(), false), "**Prompt tuning:**")
		}},
		{"q: flow-fast report without Learning blocks naming it", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", lastText: "Landed.\n**Prompt tuning:** NONE\n"}
			block(t, runStopGuard(t, f, t.TempDir(), false), "**Learning:**")
		}},
		{"r: complete Finished handoff allows", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", lastText: "## Finished\n**Learning:** NONE — off\n**Prompt tuning:** NONE\n"}
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"s: incomplete handoff with stop_hook_active allows", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", lastText: "## Finished\n**Learning:** NONE — off\n"}
			allow(t, runStopGuard(t, f, t.TempDir(), true))
		}},
		{"t: earlier turn's incomplete handoff, then operator message, then tool_use only allows", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", oldText: "## Finished\n**Learning:** NONE — off\n", lastUser: "thanks", lastTool: "Bash"}
			allow(t, runStopGuard(t, f, t.TempDir(), false))
		}},
		{"u: Learning label mid-sentence allows", func(t *testing.T) {
			f := stopFixture{state: "FINISHED", lastText: "The template carries a **Learning:** line and more."}
			allow(t, runStopGuard(t, f, t.TempDir(), false))
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

// TestFlowStopGuardReblocksARepeatedEnd pins kan-1099: a /flow turn ended twice
// at the same open stage on a status report offering to continue, and the
// hook let the second end through because its fingerprint had not moved.
// Every repeat is refused; an operator stop passes only as an AskUserQuestion
// or a ## Question.
// smoke: core — the stop gate
// TestFlowStopGuardSkipsALoopRun: a loop's own open stage run, newer than the change's, carries the
// parent's session id; the guard still judges the STARTED change and its open stage.
func TestFlowStopGuardSkipsALoopRun(t *testing.T) {
	t.Parallel()
	out := runStopGuard(t, stopFixture{state: "STARTED", ticked: 3, total: 7, open: true, loopRun: true}, t.TempDir(), false)
	want := "/flow change fixture-change is mid-phase (stage flow.implement is still open)"
	if !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, want) {
		t.Errorf("expected block naming %q, got %q", want, out)
	}
}

func TestFlowStopGuardReblocksARepeatedEnd(t *testing.T) {
	t.Parallel()
	offer := "Stopped here because the turn has been long and each step is a real dispatch or build. Send /flow to continue."
	home := t.TempDir()
	for i, active := range []bool{false, true, true} {
		f := stopFixture{state: "STARTED", ticked: 3, total: 7, open: true, lastText: offer}
		if out := runStopGuard(t, f, home, active); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "flow.implement") {
			t.Errorf("end %d at the same open stage: expected block naming flow.implement, got %q", i+1, out)
		}
	}
	f := stopFixture{state: "STARTED", ticked: 3, total: 7, open: true, lastText: "## Question\n\nThe store needs a credential only you hold."}
	if out := runStopGuard(t, f, home, true); out != "" {
		t.Errorf("a ## Question stop: expected allow (no output), got %q", out)
	}
	f.lastText = "The report says ## Question mid-line."
	if out := runStopGuard(t, f, home, true); !strings.Contains(out, `"decision": "block"`) {
		t.Errorf("## Question mid-line: expected block, got %q", out)
	}
}

// TestFlowStopGuardRefusesAStopOnARefusal pins the stop a /flow run made three
// times (gymie kan-1091 the last): the review panel's merger dispatch was
// refused over an `-unread` report, the run edited the reviewer's report to
// pass the check, the auto-mode classifier denied the edit, and the turn ended
// on a report to the operator -- let through by a background `sleep` that had
// not notified, and the two times before by a question. A refusal's cure is in
// the run's reach, so no exemption covers a turn whose last tool result is one;
// the fourth such end since the operator's message passes, so a real dead end
// still surfaces.
// smoke: core — the stop gate
func TestFlowStopGuardRefusesAStopOnARefusal(t *testing.T) {
	t.Parallel()
	const denied = "Permission for this action was denied by the Claude Code auto mode classifier. Reason: [Undermining Oversight]."
	const fix = "fix the cause and continue"
	stopping := "I'm stopping: the review panel merger is blocked, and the auto-mode classifier denied my edit of the reviewer's report."
	refused := func(text string) stopFixture {
		return stopFixture{state: "STARTED", ticked: 3, total: 7, open: true, refusal: denied, lastText: text}
	}
	blocks := func(t *testing.T, f stopFixture) {
		t.Helper()
		if out := runStopGuard(t, f, t.TempDir(), false); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, fix) {
			t.Errorf("expected block saying %q, got %q", fix, out)
		}
	}
	allows := func(t *testing.T, f stopFixture) {
		t.Helper()
		if out := runStopGuard(t, f, t.TempDir(), false); out != "" {
			t.Errorf("expected allow (no output), got %q", out)
		}
	}
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"i: the failure, behind a pending background command, blocks", func(t *testing.T) {
			f := refused(stopping)
			f.bgBash = true
			blocks(t, f)
		}},
		{"i: a question on a refusal blocks", func(t *testing.T) {
			blocks(t, refused("## Question\n\nThe merger refuses to start. How should I proceed?"))
		}},
		{"i: a blocked dispatch and a guard refusal block", func(t *testing.T) {
			for _, text := range []string{
				"PreToolUse:Agent hook error: Dispatch blocked: flow refused this dispatch's begin:\nflow: -key \"panel-1-merger\": round 1 leaves panel-report-1-failure-modes.md unread",
				"Exit code 1\nLOOP-DISPATCH-REFUSED: OPEN — pipeline-fix-70-plan has not been closed",
			} {
				f := refused(stopping)
				f.refusal = text
				blocks(t, f)
			}
		}},
		{"ii: a no-honest-recommendation question allows", func(t *testing.T) {
			f := refused("## Question\n\nTwo records match this change name and nothing recommends one. Which is it?")
			f.refusal = ""
			allows(t, f)
		}},
		{"iii: a finished handoff allows", func(t *testing.T) {
			f := refused("## Finished\n**Learning:** NONE — off\n**Prompt tuning:** NONE\n")
			f.state = "FINISHED"
			allows(t, f)
		}},
		{"an operator stop allows", func(t *testing.T) {
			f := refused(stopping)
			f.lastUser = "stop here"
			allows(t, f)
		}},
		{"a plain failing command is not a refusal", func(t *testing.T) {
			f := refused("## Question\n\nThe store needs a credential only you hold.")
			f.refusal = "Exit code 255\nPermission denied (publickey)."
			allows(t, f)
		}},
		{"the third consecutive block still blocks, the fourth end passes", func(t *testing.T) {
			f := refused(stopping)
			f.blocks = 2
			blocks(t, f)
			f.blocks = 3
			f.bgBash = true // the end then passes as it did before the rule
			allows(t, f)
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}
