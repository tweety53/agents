package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

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
