package repotest

import (
	"os/exec"
	"strings"
	"testing"
)

// TestClaudeFlagsInOneSeam: the headless Claude CLI flags are assembled in one
// non-test Go file, pjClaudeArgv's (stats/internal/guard/plans/pairjudge.go), so
// no second consumer builds its own invocation.
func TestClaudeFlagsInOneSeam(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("git", "grep", "-l", "-F", "-e", "--strict-mcp-config", "--", "*.go", ":!*_test.go")
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git grep: %v", err)
	}
	if got := strings.TrimSpace(string(out)); got != "stats/internal/guard/plans/pairjudge.go" {
		t.Errorf("files naming --strict-mcp-config, want only stats/internal/guard/plans/pairjudge.go:\n%s", got)
	}
}
