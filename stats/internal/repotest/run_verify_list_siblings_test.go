package repotest

// TestRunVerifyListShimNamesItsSiblings pins that scripts/run-verify-list.sh
// spells every sibling script runverifylist.go joins onto its script dir, so
// check-guard-symlinks rule 2 derives it and setup.sh installs it beside the
// shim (kan-1048: run-capped.sh was never installed, and every suite failed to
// start).

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestRunVerifyListShimNamesItsSiblings(t *testing.T) {
	root := repoRoot(t)
	src, err := os.ReadFile(filepath.Join(root, "stats/internal/guard/runverifylist.go"))
	if err != nil {
		t.Fatal(err)
	}
	shim, err := os.ReadFile(filepath.Join(root, "scripts/run-verify-list.sh"))
	if err != nil {
		t.Fatal(err)
	}
	joins := regexp.MustCompile(`filepath\.Join\(scriptDir, "([^"]+\.sh)"\)`).FindAllStringSubmatch(string(src), -1)
	if len(joins) == 0 {
		t.Fatal("no filepath.Join(scriptDir, ...) in runverifylist.go")
	}
	for _, m := range joins {
		if !strings.Contains(string(shim), "$SCRIPT_DIR/"+m[1]) {
			t.Errorf("scripts/run-verify-list.sh does not name $SCRIPT_DIR/%s", m[1])
		}
	}
}
