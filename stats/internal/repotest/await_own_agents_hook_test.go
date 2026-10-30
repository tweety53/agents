package repotest

// TestAwaitOwnAgentsHook pins hooks/await-own-agents.py, the SubagentStop
// hook that keeps a subagent from ending its turn while an agent it launched
// still runs: the wait would outlive its 5-minute prompt cache (kan-1068).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAwaitOwnAgentsHook(t *testing.T) {
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
	if rc != 0 || !strings.Contains(out, `"decision": "block"`) || !strings.Contains(out, "await-sub-agents.sh run1 run2") {
		t.Errorf("running children: rc %d, out %q; want a block naming await-sub-agents.sh run1 run2", rc, out)
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
