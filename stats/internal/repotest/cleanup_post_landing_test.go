package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCleanupFinishesPostLandingItems pins kan-1104: run 2 ends with no "Not done" item.
func TestCleanupFinishesPostLandingItems(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow/cleanup.md"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	for _, want := range []string{"Every post-landing item is finished before the handoff prints", "`fixed` rows", "never the way\nout"} {
		if !strings.Contains(s, want) {
			t.Errorf("cleanup.md lacks %q", want)
		}
	}
}
