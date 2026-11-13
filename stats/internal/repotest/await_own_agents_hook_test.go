package repotest

// TestAwaitOwnAgentsHook pins hooks/await-own-agents.py, the SubagentStop
// hook that keeps a subagent from ending its turn while an agent it launched
// still runs: the wait would outlive its 5-minute prompt cache (kan-1068).

import (
	"encoding/json"
	"github.com/tweety53/agents/stats/internal/smoke"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAwaitOwnAgentsHook(t *testing.T) {
	smoke.Heavy(t)
	dir := filepath.Join(t.TempDir(), "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	meta := func(id, parent string, finished bool) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"agentType": "Plan", "parentAgentId": parent, "finished": finished})
		if err := os.WriteFile(filepath.Join(dir, "agent-"+id+".meta.json"), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run := func(agent string, active bool) (string, int) {
		t.Helper()
		p, _ := json.Marshal(map[string]any{
			"hook_event_name": "SubagentStop", "agent_id": agent, "stop_hook_active": active,
			"agent_transcript_path": filepath.Join(dir, "agent-"+agent+".jsonl"),
		})
		out, _, rc := runHook(t, "await-own-agents.py", p)
		return out, rc
	}

	meta("planner", "", false)
	meta("done1", "planner", true)
	meta("other", "someone-else", false)
	if out, rc := run("planner", false); rc != 0 || out != "" {
		t.Errorf("no running child: rc %d, out %q; want allow", rc, out)
	}

	meta("run1", "planner", false)
	meta("run2", "planner", false)
	out, rc := run("planner", false)
	if rc != 0 || !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "flow-guard await-sub-agents run1 run2") {
		t.Errorf("running children: rc %d, out %q; want a block naming flow-guard await-sub-agents run1 run2", rc, out)
	}
	// stop_hook_active still blocks: the wait is the point, and the script returns within two minutes.
	if out, _ := run("planner", true); !strings.Contains(out, `"decision": "block"`) {
		t.Errorf("stop_hook_active with a running child: out %q; want a block", out)
	}

	// A child whose transcript went quiet for 30 minutes is dead and never traps its parent.
	for _, id := range []string{"run1", "run2"} {
		tr := filepath.Join(dir, "agent-"+id+".jsonl")
		if err := os.WriteFile(tr, []byte("{}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		old := time.Now().Add(-31 * time.Minute)
		if err := os.Chtimes(tr, old, old); err != nil {
			t.Fatal(err)
		}
	}
	if out, rc := run("planner", false); rc != 0 || out != "" {
		t.Errorf("stale children: rc %d, out %q; want allow", rc, out)
	}

	// Fails open on a payload it cannot read.
	if out, _, rc := runHook(t, "await-own-agents.py", []byte("not json")); rc != 0 || out != "" {
		t.Errorf("bad payload: rc %d, out %q; want allow", rc, out)
	}
}

// TestAwaitOwnAgentsHookWaitsOnBackgroundBash pins decision
// background-bash-blocks-stop: a subagent may not end its turn while a Bash
// command it moved to the background has no task-notification (kan-1085).
func TestAwaitOwnAgentsHookWaitsOnBackgroundBash(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	tr := filepath.Join(dir, "agent-impl.jsonl")
	appendEntry := func(role string, content any) {
		t.Helper()
		b, _ := json.Marshal(map[string]any{"type": role, "message": map[string]any{"role": role, "content": content}})
		f, err := os.OpenFile(tr, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if _, err := f.Write(append(b, '\n')); err != nil {
			t.Fatal(err)
		}
	}
	appendLine := func(content any) { t.Helper(); appendEntry("user", content) }
	call := func(id, tool string) {
		t.Helper()
		appendEntry("assistant", []any{map[string]any{"type": "tool_use", "id": id, "name": tool, "input": map[string]any{}}})
	}
	resultOf := func(id, text string) []any {
		return []any{map[string]any{"type": "tool_result", "tool_use_id": id, "content": text}}
	}
	result := func(text string) []any { return resultOf("t1", text) }
	note := func(id, status string) string {
		return "<task-notification>\n<task-id>" + id + "</task-id>\n<status>" + status + "</status>\n</task-notification>"
	}
	run := func() string {
		t.Helper()
		p, _ := json.Marshal(map[string]any{"hook_event_name": "SubagentStop", "agent_id": "impl", "stop_hook_active": false, "agent_transcript_path": tr})
		out, _, rc := runHook(t, "await-own-agents.py", p)
		if rc != 0 {
			t.Fatalf("rc %d, out %q", rc, out)
		}
		return out
	}

	call("t1", "Bash")
	appendLine(result("Command running in background with ID: bg1. Output is being written to: /x/bg1.output."))
	if out := run(); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "bg1") {
		t.Errorf("(a) running bg1: out %q; want a block naming bg1", out)
	}
	appendLine(note("bg1", "completed"))
	if out := run(); out != "" {
		t.Errorf("(b) bg1 notified: out %q; want allow", out)
	}
	call("t1", "Bash")
	appendLine(result("1 2 3\nCommand did not complete within its 600s timeout and was moved to the background (ID: bg2). Output is being written to: /x/bg2.output."))
	if out := run(); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "bg2") {
		t.Errorf("(c) running bg2: out %q; want a block naming bg2", out)
	}
	appendLine(note("bg2", "killed"))
	if out := run(); out != "" {
		t.Errorf("(d) bg2 killed: out %q; want allow", out)
	}
	// A launch sentence the agent only read -- a Read of a transcript, a Bash that printed
	// a file -- launches nothing (review R1).
	call("t3", "Read")
	appendLine(resultOf("t3", "Command running in background with ID: bgX. Output is being written to: /x/bgX.output."))
	call("t4", "Bash")
	appendLine(resultOf("t4", "{\"content\":\"Command running in background with ID: bgY.\"}"))
	call("t5", "Bash")
	appendLine(resultOf("t5", "+\tappendLine(result(\"Command was moved to the background (ID: bgZ) after running for 180s.\"))\n"+
		`{"content":"1\nCommand was moved to the background (ID: bgW) after running for 180s."}`))
	if out := run(); out != "" {
		t.Errorf("(e) launches only quoted: out %q; want allow", out)
	}
}
