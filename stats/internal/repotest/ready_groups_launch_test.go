package repotest

// TestGroupBoundaryLaunchesFromTheReadySet pins that implementer groups launch
// from flow-guard ready-groups's printed set (kan-1131), never from plan order.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGroupBoundaryLaunchesFromTheReadySet(t *testing.T) {
	t.Parallel()
	const cmd = "flow-guard ready-groups <changeRoot>/tasks.md <canonical-worktree> <name>"
	for _, rel := range []string{"skills/flow/sdd-dispatch.md", "skills/flow/implement.md"} {
		b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
		if err != nil {
			t.Fatal(err)
		}
		s := strings.Join(strings.Fields(string(b)), " ")
		if !strings.Contains(s, cmd) {
			t.Errorf("%s lacks %q", rel, cmd)
		}
		for _, gone := range []string{"A fourth or later ready group", "readies group N+2"} {
			if strings.Contains(s, gone) {
				t.Errorf("%s still carries %q", rel, gone)
			}
		}
	}
}
