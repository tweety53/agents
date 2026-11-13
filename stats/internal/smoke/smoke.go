// Package smoke splits a package's tests between the default test list and
// the full suite.
//
// A heavy mark is smoke.Heavy(t) as the first statement of a top-level test:
// the test steps aside in a -short run unless the run names its tests with
// -run, FLOW_VERIFY_FULL is set, or FLOW_SMOKE_KEEP names the test's file or
// a directory above it. A core tag is "// smoke: core — <capability>" as the
// last line of the comment above a test that carries no heavy mark: it keeps
// that test in every default run. The rubric for both, and the list that runs
// them, are .flow/project.md's ## test.
package smoke

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	shortRun = testing.Short
	runFlag  = func() string {
		if f := flag.Lookup("test.run"); f != nil {
			return f.Value.String()
		}
		return ""
	}
	getenv = os.Getenv
)

// Heavy skips the calling test in a -short run that neither selects tests
// with -run, sets FLOW_VERIFY_FULL, nor keeps the caller's file through
// FLOW_SMOKE_KEEP's space-separated repository-relative paths.
func Heavy(t testing.TB) {
	t.Helper()
	_, file, _, _ := runtime.Caller(1)
	if !shortRun() || runFlag() != "" || getenv("FLOW_VERIFY_FULL") != "" {
		return
	}
	rel := repoRel(file)
	for _, k := range strings.Fields(getenv("FLOW_SMOKE_KEEP")) {
		if k == rel || strings.HasPrefix(rel, strings.TrimSuffix(k, "/")+"/") {
			return
		}
	}
	t.Skip("smoke: heavy test skipped in a -short run: " + rel)
}

// repoRel is file as a repository-relative path: the module is stats/.
func repoRel(file string) string {
	_, self, _, _ := runtime.Caller(0)
	rel, err := filepath.Rel(filepath.Dir(filepath.Dir(filepath.Dir(self))), file)
	if err != nil {
		return file
	}
	return "stats/" + filepath.ToSlash(rel)
}
