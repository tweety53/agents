package repotest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReviewersNeverAddAWorktree pins the ban on a review slot adding a git
// worktree (kan-1084 self-review, F3: a mutation slot's detached worktree
// outlived its dispatch and was removed by hand).
func TestReviewersNeverAddAWorktree(t *testing.T) {
	t.Parallel()
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	brief := read("skills/flow/gated-reviewer-brief.md")
	i := strings.Index(brief, "**READ-ONLY REVIEW:**")
	if i < 0 {
		t.Fatal("skills/flow/gated-reviewer-brief.md: no READ-ONLY REVIEW block")
	}
	block := brief[i:]
	if j := strings.Index(block, "\n\n"); j >= 0 {
		block = block[:j]
	}
	if !strings.Contains(block, "`git worktree add`") {
		t.Errorf("READ-ONLY REVIEW does not forbid `git worktree add`")
	}
	panel := read("skills/flow/review-panel.md")
	i = strings.Index(panel, "### The mutation-testing brief")
	if i < 0 {
		t.Fatal("skills/flow/review-panel.md: no mutation-testing brief")
	}
	mut := panel[i:]
	if j := strings.Index(mut[4:], "\n### "); j >= 0 {
		mut = mut[:j+4]
	}
	for _, want := range []string{"Never run `git worktree add`", "`git show <base>:<path>`"} {
		if !strings.Contains(mut, want) {
			t.Errorf("the mutation-testing brief lacks %q", want)
		}
	}
}

// TestDocsOnlyReductionDropsSanityEfficiency pins KAN-1110: a docs-only panel
// drops the test-sanity+call-efficiency bundle and keeps mutation+failure-modes.
func TestDocsOnlyReductionDropsSanityEfficiency(t *testing.T) {
	t.Parallel()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow/review-panel.md"))
	if err != nil {
		t.Fatal(err)
	}
	panel := string(b)
	i := strings.Index(panel, "### The always-on extra dispatches")
	if i < 0 {
		t.Fatal("skills/flow/review-panel.md: no always-on extra dispatches section")
	}
	extras := panel[i:]
	if j := strings.Index(extras[4:], "\n### "); j >= 0 {
		extras = extras[:j+4]
	}
	extras = strings.Join(strings.Fields(extras), " ")
	if want := "The docs-only reduction drops `test-sanity+call-efficiency`"; !strings.Contains(extras, want) {
		t.Errorf("The always-on extra dispatches lacks %q", want)
	}
}
