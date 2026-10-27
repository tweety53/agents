package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTestListRunsTheGuardPackageOnce pins kan-1038's f9-guard-once: the
// guard package runs once per verify, in scripts/test-go-guards.sh under
// -race, and the `## test` Go line of .flow/project.md filters it out of the
// module's package list rather than running the bare ./... pattern.
func TestTestListRunsTheGuardPackageOnce(t *testing.T) {
	root := repoRoot(t)
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	var harnessLine string
	for _, l := range strings.Split(read("scripts/test-go-guards.sh"), "\n") {
		if strings.Contains(l, "go test ./internal/guard/") {
			harnessLine = l
		}
	}
	if !strings.Contains(harnessLine, "-race") {
		t.Errorf("scripts/test-go-guards.sh's go test line %q does not carry -race", harnessLine)
	}

	md := read(".flow/project.md")
	_, section, ok := strings.Cut(md, "\n## test\n")
	if !ok {
		t.Fatal(".flow/project.md has no ## test section")
	}
	section, _, _ = strings.Cut(section, "\n```\n")
	var goLine string
	for _, l := range strings.Split(section, "\n") {
		if strings.Contains(l, "go test") {
			goLine = l
		}
	}
	if !strings.Contains(goLine, "$(go list ./... | grep -v '/internal/guard$')") || strings.Contains(goLine, "go test ./...") {
		t.Errorf("## test Go line %q does not draw its packages from go list with /internal/guard filtered out", goLine)
	}
}
