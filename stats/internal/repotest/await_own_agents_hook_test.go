package repotest

// TestAwaitOwnAgentsHook pins hooks/await-own-agents.py, the SubagentStop
// hook that keeps a subagent from ending its turn while an agent it launched
// still runs: the wait would outlive its 5-minute prompt cache (kan-1068).

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
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
	wait := "2" // seconds the hook waits for a running child before answering "still running"
	run := func(agent string, active bool) (string, int) {
		t.Helper()
		p, _ := json.Marshal(map[string]any{
			"hook_event_name": "SubagentStop", "agent_id": agent, "stop_hook_active": active,
			"agent_transcript_path": filepath.Join(dir, "agent-"+agent+".jsonl"),
		})
		out, _, rc := runHook(t, "await-own-agents.py", p, "AWAIT_CHILDREN_SECONDS="+wait)
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
	if rc != 0 || !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "still running (run1 run2)") {
		t.Errorf("running children: rc %d, out %q; want a block saying run1 run2 still run", rc, out)
	}
	// stop_hook_active still blocks: the wait is the point.
	if out, _ := run("planner", true); !strings.Contains(out, `"decision": "block"`) {
		t.Errorf("stop_hook_active with a running child: out %q; want a block", out)
	}
	// The hook waits for the children itself: they finish mid-wait, and the hook answers once that they have.
	go func() {
		time.Sleep(1500 * time.Millisecond)
		meta("run1", "planner", true)
		meta("run2", "planner", true)
	}()
	wait = "30"
	waited := time.Now()
	if out, _ := run("planner", false); !strings.Contains(out, "have finished (run1 run2)") || time.Since(waited) > 20*time.Second {
		t.Errorf("children finishing mid-wait: out %q after %s; want a block saying they finished", out, time.Since(waited))
	}
	wait = "2"
	meta("run1", "planner", false)
	meta("run2", "planner", false)

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

// TestAwaitOwnAgentsHookTaskStopEndsBackground pins decision
// taskstop-ends-background: a TaskStop result paired with the agent's own
// TaskStop call ends the background command it names; the same sentence in a
// Read or a Bash result ends nothing.
func TestAwaitOwnAgentsHookTaskStopEndsBackground(t *testing.T) {
	smoke.Heavy(t)
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
	call := func(id, tool string) {
		t.Helper()
		appendEntry("assistant", []any{map[string]any{"type": "tool_use", "id": id, "name": tool, "input": map[string]any{}}})
	}
	resultOf := func(id, text string) {
		t.Helper()
		appendEntry("user", []any{map[string]any{"type": "tool_result", "tool_use_id": id, "content": text}})
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

	call("b1", "Bash")
	resultOf("b1", "Command running in background with ID: bg1. Output is being written to: /x/bg1.output.")
	if out := run(); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "bg1") {
		t.Errorf("(a) running bg1: out %q; want a block naming bg1", out)
	}
	call("s1", "TaskStop")
	resultOf("s1", "{\"message\":\"Successfully stopped task: bg1 (sleep 3000)\",\"task_id\":\"bg1\",\"task_type\":\"local_bash\",\"command\":\"sleep 3000\"}")
	if out := run(); out != "" {
		t.Errorf("(b) bg1 stopped with TaskStop: out %q; want allow", out)
	}
	call("b2", "Bash")
	resultOf("b2", "Command running in background with ID: bg2.")
	call("r1", "Read")
	resultOf("r1", "Successfully stopped task: bg2 (sleep 3000)")
	call("b3", "Bash")
	resultOf("b3", "Successfully stopped task: bg2")
	if out := run(); !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "bg2") {
		t.Errorf("(c) stop only quoted: out %q; want a block naming bg2", out)
	}
}

// TestAwaitOwnAgentsHookKillsLeftoverBackground pins the stop that carries a
// final message: the background command the agent still holds is killed (only
// the processes holding its own output file) and the stop goes through; with
// no final message the stop still blocks and the command lives.
func TestAwaitOwnAgentsHookKillsLeftoverBackground(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "sess", "subagents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	startShell := func(name string) (*exec.Cmd, string) {
		t.Helper()
		out := filepath.Join(dir, name+".output")
		f, err := os.Create(out)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		c := exec.Command("sleep", "120")
		c.Stdout = f
		if err := c.Start(); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = c.Process.Kill(); _, _ = c.Process.Wait() })
		return c, out
	}
	own, ownOut := startShell("bgown")
	sibling, _ := startShell("bgsibling")

	tr := filepath.Join(dir, "agent-impl.jsonl")
	call, _ := json.Marshal(map[string]any{"type": "assistant", "message": map[string]any{"role": "assistant", "content": []any{map[string]any{"type": "tool_use", "id": "t1", "name": "Bash", "input": map[string]any{}}}}})
	res, _ := json.Marshal(map[string]any{"type": "user", "message": map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "t1", "content": "Command was moved to the background (ID: bgown) after running for 180s. Output is being written to: " + ownOut + ". You will be notified"}}}})
	if err := os.WriteFile(tr, append(append(call, '\n'), append(res, '\n')...), 0o644); err != nil {
		t.Fatal(err)
	}
	run := func(final string) string {
		t.Helper()
		p, _ := json.Marshal(map[string]any{"hook_event_name": "SubagentStop", "agent_id": "impl", "agent_transcript_path": tr, "last_assistant_message": final})
		out, _, rc := runHook(t, "await-own-agents.py", p)
		if rc != 0 {
			t.Fatalf("rc %d, out %q", rc, out)
		}
		return out
	}
	gone := func(c *exec.Cmd) bool {
		done := make(chan struct{})
		go func() { _ = c.Wait(); close(done) }()
		select {
		case <-done:
			return true
		case <-time.After(3 * time.Second):
			return false
		}
	}

	if out := run(""); !strings.Contains(out, `"decision": "block"`) {
		t.Fatalf("no final message: out %q; want a block", out)
	}
	if out := run("task 19 done\ncommit: 19 " + strings.Repeat("a", 40)); out != "" {
		t.Errorf("final message: out %q; want allow", out)
	}
	if !gone(own) {
		t.Error("the agent's own background command still runs after its stop")
	}
	if sibling.ProcessState != nil {
		t.Error("a sibling's background command was killed")
	}
}
