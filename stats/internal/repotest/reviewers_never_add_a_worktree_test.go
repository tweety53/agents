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
	if !strings.Contains(read("skills/flow/fragments/read-only-review.md"), "`git worktree add`") {
		t.Errorf("READ-ONLY REVIEW does not forbid `git worktree add`")
	}
	mut := read("skills/flow/fragments/mutation-testing-brief.md")
	for _, want := range []string{"Never run `git worktree add`", "`git show <base>:<path>`"} {
		if !strings.Contains(mut, want) {
			t.Errorf("the mutation-testing brief lacks %q", want)
		}
	}
}
