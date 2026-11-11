package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunSizeStopNamesPhaseFileLength pins kan-1104: a phase file's length is never a stop.
func TestRunSizeStopNamesPhaseFileLength(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow-contracts/pipeline.md"))
	if err != nil {
		t.Fatal(err)
	}
	const start = "**A run's size is never a stop.**"
	s := string(b)
	i := strings.Index(s, start)
	if i < 0 {
		t.Fatalf("pipeline.md: no %q paragraph", start)
	}
	para := s[i:]
	if j := strings.Index(para, "\n\n"); j >= 0 {
		para = para[:j]
	}
	for _, want := range []string{"phase file", "offset chunks", "no explicit operator stop", "never handed to the\noperator as 're-run /flow to continue'"} {
		if !strings.Contains(para, want) {
			t.Errorf("pipeline.md size-stop paragraph lacks %q", want)
		}
	}
	if !strings.Contains(section(t, "skills/flow/SKILL.md", "Guardrails"), "phase file's length") {
		t.Errorf("SKILL.md Guardrails lacks %q", "phase file's length")
	}
}
