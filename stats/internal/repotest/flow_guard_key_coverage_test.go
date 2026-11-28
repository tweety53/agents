package repotest

// TestFlowGuardKeyCoversEveryDependency pins scripts/flow-guard's cache key to
// the packages the cmd/flow-guard binary is actually built from. The key hashes
// the files its flow_guard_files globs name; a package of this module that the
// binary imports but no glob covers is an edit that never changes the key, so
// every later call execs a stale cached binary. The test parses the globs out
// of the script and fails naming each dependency of `go list -deps
// ./cmd/flow-guard` under this module that none of them matches.

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const statsModule = "github.com/tweety53/agents/stats"

// flowGuardKeyGlobs returns the stats-relative globs flow_guard_files hashes.
func flowGuardKeyGlobs(t *testing.T) []string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "scripts", "flow-guard"))
	if err != nil {
		t.Fatal(err)
	}
	body := regexp.MustCompile(`(?s)flow_guard_files\(\) \{(.*?)\n\}`).FindSubmatch(b)
	if body == nil {
		t.Fatal("scripts/flow-guard: no flow_guard_files function")
	}
	var globs []string
	for _, m := range regexp.MustCompile(`"\$1"/([^\s;"}]+)`).FindAllSubmatch(body[1], -1) {
		globs = append(globs, string(m[1]))
	}
	if len(globs) == 0 {
		t.Fatal("scripts/flow-guard: flow_guard_files names no globs")
	}
	return globs
}

func TestFlowGuardKeyCoversEveryDependency(t *testing.T) {
	globs := flowGuardKeyGlobs(t)
	cmd := exec.Command(goBinary(t), "list", "-deps", "-f", "{{.ImportPath}}", "./cmd/flow-guard")
	cmd.Dir = filepath.Join(repoRoot(t), "stats")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	for _, pkg := range strings.Fields(string(out)) {
		rel, ok := strings.CutPrefix(pkg, statsModule+"/")
		if !ok {
			continue
		}
		covered := false
		for _, g := range globs {
			if m, _ := filepath.Match(g, rel+"/x.go"); m {
				covered = true
				break
			}
		}
		if !covered {
			t.Errorf("scripts/flow-guard's cache key does not hash %s, which cmd/flow-guard imports (globs %v)", rel, globs)
		}
	}
}
