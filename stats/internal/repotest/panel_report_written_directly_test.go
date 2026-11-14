package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPanelReportFileIsWrittenDirectly pins the REPORT FILE paragraph's ban
// on a report derived by substitution (kan-1123 self-review, F3: a bundled
// reviewer wrote two role files from one shell variable, rewriting the
// findings' own matches as well).
func TestPanelReportFileIsWrittenDirectly(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow/review-panel.md"))
	if err != nil {
		t.Fatal(err)
	}
	panel := string(b)
	i := strings.Index(panel, "> **REPORT FILE:** write your full report to")
	if i < 0 {
		t.Fatal("skills/flow/review-panel.md: no REPORT FILE paragraph")
	}
	block := panel[i:]
	if j := strings.Index(block, "\n\n"); j >= 0 {
		block = block[:j]
	}
	lines := strings.Split(block, "\n")
	for k, l := range lines {
		lines[k] = strings.TrimPrefix(l, "> ")
	}
	want := "never one derived from another role's report or from a shared text by substitution"
	if !strings.Contains(strings.Join(lines, " "), want) {
		t.Errorf("REPORT FILE paragraph does not say %q", want)
	}
}
