package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTestListRunsTheGuardPackageOnce pins kan-1038's f9-guard-once: the
// guard package runs once per verify, in the `guard-tests` suite's go test
// under -race (KAN-1058), and the `stats-go` line of .flow/project.md filters
// it out of the module's package list rather than running the bare ./...
// pattern.
func TestTestListRunsTheGuardPackageOnce(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), ".flow/project.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(b), "\n## test\n")
	if !ok {
		t.Fatal(".flow/project.md has no ## test section")
	}
	section, _, _ = strings.Cut(section, "\n```\n")
	suites := map[string]string{}
	var suite string
	for _, l := range strings.Split(section, "\n") {
		if name, ok := strings.CutPrefix(l, "# suite: "); ok {
			suite = name
		} else if suite != "" && strings.Contains(l, "go test") {
			suites[suite] = l
		}
	}
	if g := suites["guard-tests"]; !strings.Contains(g, "go test -race -count=1 ./internal/guard/...") {
		t.Errorf("## test guard-tests line %q does not run go test -race over ./internal/guard/...", g)
	}
	goLine := suites["stats-go"]
	if !strings.Contains(goLine, "$(go list ./... | grep -v '/internal/guard$')") || strings.Contains(goLine, "go test ./...") {
		t.Errorf("## test Go line %q does not draw its packages from go list with /internal/guard filtered out", goLine)
	}
}
