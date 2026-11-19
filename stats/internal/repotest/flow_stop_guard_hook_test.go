package repotest

// TestFlowStopGuardHook pins hooks/flow-stop-guard.py, the Stop hook that
// refuses a handoff dropping its never-omitted lines. Each subtest writes a
// transcript and pipes a Stop payload to the hook with HOME in a temp dir.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stopFixture struct {
	lastUser string // last operator message
	lastTool string // last assistant tool use
	lastText string // last assistant text
	oldText  string // assistant text from an earlier turn, before the operator message
}

func runStopGuard(t *testing.T, f stopFixture, active bool) string {
	t.Helper()
	user := f.lastUser
	if user == "" {
		user = "/flow fixture-change"
	}
	transcript := ""
	if f.oldText != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":%q}]}}`+"\n", f.oldText)
	}
	transcript += fmt.Sprintf(`{"type":"user","message":{"content":%q}}`+"\n", user)
	if f.lastTool != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"tool_use","name":%q,"input":{}}]}}`+"\n", f.lastTool)
	}
	if f.lastText != "" {
		transcript += fmt.Sprintf(`{"type":"assistant","message":{"content":[{"type":"text","text":%q}]}}`+"\n", f.lastText)
	}
	tp := filepath.Join(t.TempDir(), "sess-1.jsonl")
	if err := os.WriteFile(tp, []byte(transcript), 0o644); err != nil {
		t.Fatal(err)
	}
	stdin := fmt.Sprintf(`{"session_id":"sess-1","transcript_path":%q,"stop_hook_active":%t}`, tp, active)
	out, errOut, rc := runHook(t, "flow-stop-guard.py", []byte(stdin), "HOME="+t.TempDir())
	if rc != 0 || errOut != "" {
		t.Fatalf("rc=%d stderr=%s", rc, errOut)
	}
	return out
}

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
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"p: Finished handoff without Prompt tuning blocks naming it", func(t *testing.T) {
			f := stopFixture{lastText: "## Finished\n\n**Change:** x\n**Learning:** /a/n.md\n"}
			block(t, runStopGuard(t, f, false), "**Prompt tuning:**")
		}},
		{"q: flow-fast report without Learning blocks naming it", func(t *testing.T) {
			f := stopFixture{lastText: "Landed.\n**Prompt tuning:** NONE\n"}
			block(t, runStopGuard(t, f, false), "**Learning:**")
		}},
		{"r: complete Finished handoff allows", func(t *testing.T) {
			f := stopFixture{lastText: "## Finished\n**Learning:** NONE — off\n**Prompt tuning:** NONE\n"}
			allow(t, runStopGuard(t, f, false))
		}},
		{"s: incomplete handoff with stop_hook_active allows", func(t *testing.T) {
			f := stopFixture{lastText: "## Finished\n**Learning:** NONE — off\n"}
			allow(t, runStopGuard(t, f, true))
		}},
		{"t: earlier turn's incomplete handoff, then operator message, then tool_use only allows", func(t *testing.T) {
			f := stopFixture{oldText: "## Finished\n**Learning:** NONE — off\n", lastUser: "thanks", lastTool: "Bash"}
			allow(t, runStopGuard(t, f, false))
		}},
		{"u: Learning label mid-sentence allows", func(t *testing.T) {
			f := stopFixture{lastText: "The template carries a **Learning:** line and more."}
			allow(t, runStopGuard(t, f, false))
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}
