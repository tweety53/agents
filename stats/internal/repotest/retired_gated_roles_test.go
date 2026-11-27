package repotest

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoReaderOfTheRetiredGatedRoles: migration 0120 folded every stored row
// of the retired gated review into the panel's roles, so no tracked file names
// gated-reviewer or gated-re-reviewer -- no reader, label, fixture or test of
// them -- outside the migrations that wrote and folded those rows, the
// rationale entry recording the retirement, 0120's own test and this one.
func TestNoReaderOfTheRetiredGatedRoles(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("git", "grep", "-l", "-e", "gated-reviewer", "-e", "gated-re-reviewer", "-e", "GatedReviewer", "--",
		".",
		":!stats/internal/store/migrations/",
		":!docs/rationale/",
		":!stats/internal/store/legacyshapes_test.go",
		":!stats/internal/repotest/retired_gated_roles_test.go")
	cmd.Dir = repoRoot(t)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 { // git grep: no match
			return
		}
		t.Fatalf("git grep: %v", err)
	}
	t.Errorf("files still naming a retired gated role:\n%s", strings.TrimSpace(string(out)))
}
