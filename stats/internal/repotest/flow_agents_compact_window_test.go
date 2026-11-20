package repotest

// TestFlowAgentsCompactAtTheContextCeiling pins autoCompactWindow on the four
// non-planner flow agent types (kan-1131, decision native-compaction).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlowAgentsCompactAtTheContextCeiling(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)
	const key = "\nautoCompactWindow: 200000\n"
	for _, n := range []string{"low", "medium", "high", "xhigh"} {
		rel := "agents/flow-" + n + ".md"
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), key) {
			t.Errorf("%s lacks the frontmatter line autoCompactWindow: 200000", rel)
		}
	}
	planners, _ := filepath.Glob(filepath.Join(root, "agents", "flow-planner-*.md"))
	if len(planners) == 0 {
		t.Fatal("no agents/flow-planner-*.md found")
	}
	for _, p := range planners {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(b), "autoCompactWindow") {
			t.Errorf("%s carries autoCompactWindow", p)
		}
	}
	b, err := os.ReadFile(filepath.Join(root, "skills/flow/implementer-brief.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "**AFTER A COMPACTION:**") {
		t.Errorf("implementer-brief.md lacks **AFTER A COMPACTION:**")
	}
}
