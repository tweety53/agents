package guard

import (
	"os/exec"
	"testing"
)

// The repository's root .gitignore covers a node_modules/ at any depth (a
// vitest started at the root writes .vite/ there) and the Playwright MCP's
// screenshots, so neither is untracked for a commit chain to sweep in
// (kan-1057 self-review, F3). One check-ignore call per path: it exits 0
// when any one of several paths is ignored.
func TestRootGitignoreCoversToolCaches(t *testing.T) {
	t.Parallel()
	for _, p := range []string{"node_modules/.vite/x", "stats/web/node_modules/x", ".playwright-mcp/x.png"} {
		if err := exec.Command("git", "-C", "../../..", "check-ignore", "-q", p).Run(); err != nil {
			t.Errorf("%s is not ignored by the root .gitignore: %v", p, err)
		}
	}
}
