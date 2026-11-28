package repotest

// TestFlowGuardHookLinksOnlyTheHookFamily pins cmd/flow-guard-hook, the binary
// every Claude Code hook registration execs on each tool call, to the hook
// family and the core it stands on: a guard family it links is start-up cost
// paid per tool call for code no hook runs. The test fails naming each guard
// package of `go list -deps ./cmd/flow-guard-hook` outside guard/core and
// guard/hook.

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFlowGuardHookLinksOnlyTheHookFamily(t *testing.T) {
	cmd := exec.Command(goBinary(t), "list", "-deps", "-f", "{{.ImportPath}}", "./cmd/flow-guard-hook")
	cmd.Dir = filepath.Join(repoRoot(t), "stats")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list ./cmd/flow-guard-hook: %v\n%s", err, out)
	}
	allowed := map[string]bool{"internal/guard/core": true, "internal/guard/hook": true}
	for _, pkg := range strings.Fields(string(out)) {
		rel, ok := strings.CutPrefix(pkg, statsModule+"/")
		if !ok || !strings.HasPrefix(rel, "internal/guard") || allowed[rel] {
			continue
		}
		t.Errorf("cmd/flow-guard-hook links %s; a hook binary links only guard/core and guard/hook", rel)
	}
}
