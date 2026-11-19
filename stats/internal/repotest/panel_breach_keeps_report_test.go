package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPanelBreachKeepsACompleteReport pins the breach paragraph (kan-1121
// self-review, F3): a slot that
// returned a parsable report past its ceiling keeps that report and its
// findings, and only a slot stopped or returning none is closed timed-out and
// re-dispatched.
func TestPanelBreachKeepsACompleteReport(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow/review-panel.md"))
	if err != nil {
		t.Fatal(err)
	}
	panel := string(b)
	i := strings.Index(panel, "On a breach, in order:")
	if i < 0 {
		t.Fatal("skills/flow/review-panel.md: no breach paragraph")
	}
	para := panel[i:]
	if j := strings.Index(para, "\n\n"); j >= 0 {
		para = para[:j]
	}
	para = strings.Join(strings.Fields(para), " ")
	for _, want := range []string{
		"A slot that returned a parsable report past its ceiling is closed by that report as usual, keeps its findings, and is not re-dispatched.",
		"A slot stopped, or one that returned no parsable report, is closed `-outcome timed-out` and re-dispatched once,",
	} {
		if !strings.Contains(para, want) {
			t.Errorf("breach paragraph does not say %q", want)
		}
	}
}
