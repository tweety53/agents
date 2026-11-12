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
