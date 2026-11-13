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
	if g := suites["guard-tests"]; !strings.Contains(g, "go test -race -short -count=1 ./internal/guard/...") {
		t.Errorf("## test guard-tests line %q does not run go test -race -short over ./internal/guard/...", g)
	}
	goLine := suites["stats-go"]
	if !strings.Contains(goLine, "$(go list ./... | grep -v '/internal/guard$')") || strings.Contains(goLine, "go test ./...") {
		t.Errorf("## test Go line %q does not draw its packages from go list with /internal/guard filtered out", goLine)
	}
}

// TestTestListRunsTheModTests pins the kan-1104 self-review's F12: the
// `flow-task-list` mod's tests run in the `## test` list, on the harness's own
// runner, as a suite of their own declaring the mod's directory as its inputs.
func TestTestListRunsTheModTests(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(repoRoot(t), ".flow/project.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, section, ok := strings.Cut(string(b), "\n## test\n")
	if !ok {
		t.Fatal(".flow/project.md has no ## test section")
	}
	section, _, _ = strings.Cut(section, "\n```\n")
	want := "# suite: flow-task-list\n" +
		"# inputs: mods/flow-task-list/\n" +
		"scripts/flow-guard unchanged mods/flow-task-list/ || claude plugin test mods/flow-task-list\n"
	if !strings.Contains(section+"\n", want) {
		t.Errorf("## test fence lacks the flow-task-list suite:\n%s", want)
	}
}

// TestTestListRunsTheSmokeSet pins smoke-is-short-on-the-existing-lines and
// keep-rows-are-arguments-on-the-line: both Go lines of `## test` run under
// -short with the keep list flow-guard smoke-keep prints, and the SPA and mod
// lines carry neither.
func TestTestListRunsTheSmokeSet(t *testing.T) {
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
		} else if suite != "" && !strings.HasPrefix(l, "#") && l != "" {
			suites[suite] = l
		}
	}
	keep := "keep=$(scripts/flow-guard smoke-keep stats/go.mod=stats/ stats/go.sum=stats/" +
		" hooks/=stats/internal/repotest/ hooks/=stats/internal/setuptest/" +
		" scripts/=stats/internal/repotest/ scripts/=stats/internal/setuptest/" +
		" setup.sh=stats/internal/setuptest/"
	ordered := []string{keep, ") && cd stats && ", `FLOW_SMOKE_KEEP="$keep" GOMAXPROCS=`, " -race -short -count=1 "}
	for _, name := range []string{"guard-tests", "stats-go"} {
		line, at := suites[name], 0
		for _, part := range ordered {
			i := strings.Index(line[at:], part)
			if i < 0 {
				t.Errorf("## test %s line %q lacks %q after offset %d", name, line, part, at)
				break
			}
			at += i + len(part)
		}
	}
	for _, name := range []string{"stats-spa", "flow-task-list"} {
		if line := suites[name]; line == "" || strings.Contains(line, "smoke-keep") || strings.Contains(line, "-short") {
			t.Errorf("## test %s line %q is missing or carries smoke-keep or -short", name, line)
		}
	}
}
