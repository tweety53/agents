package repotest

// TestBaseSyncBeforeTaskOne pins the rebase onto the base that /flow's and
// /flow-fast's implementation runs make before task 1 (kan-1032: the base moved
// 13 commits during planning, and only the review panel's base sync, after
// every task, met the overlap).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// section is the text of the "## <heading>" section of a Markdown file, up to
// the next level-2 heading.
func section(t *testing.T, rel, heading string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i := strings.Index(s, "\n## "+heading)
	if i < 0 {
		t.Fatalf("%s: no %q section", rel, "## "+heading)
	}
	s = s[i+1:]
	if j := strings.Index(s[3:], "\n## "); j >= 0 {
		s = s[:j+3]
	}
	return s
}

func TestBaseSyncBeforeTaskOne(t *testing.T) {
	t.Parallel()
	const step = "**The change branch is rebased onto its base before task 1 runs**"
	impl := section(t, "skills/flow/implement.md", "1. Load context and validate the plan")
	for _, want := range []string{
		step,
		"It runs before the collision reconciliation above.",
		"**The before-review\nplanning commit comes first**",
		"sync-panel-base.sh --rebase <worktree> <merge-base>",
		"flow state add-worktree <name> <worktree> <sha>",
		"git -C <worktree> push --force-with-lease origin spectre/<name>",
	} {
		if !strings.Contains(impl, want) {
			t.Errorf("skills/flow/implement.md section 1 lacks %q", want)
		}
	}
	if strings.Contains(impl, "it never rebases") {
		t.Errorf("skills/flow/implement.md section 1 still defers the rebase to integrate")
	}
	const row = "and before the base sync before task 1 (**1. Load context and validate the plan**, `skills/flow/implement.md`) | `<project>/spectre/changes/<name>/` | `chore(spectre): plan <name> — before review` |"
	gb, err := os.ReadFile(filepath.Join(repoRoot(t), "skills/flow-contracts/git-boundaries.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gb), row) {
		t.Errorf("skills/flow-contracts/git-boundaries.md Planning commits lacks %q", row)
	}
	const cite = "rebased onto its base before task 1 runs, as the base-sync step of **1. Load context and validate the plan** (`skills/flow/implement.md`)"
	fast := strings.Join(strings.Fields(section(t, "skills/flow-fast/SKILL.md", "4. Implement")), " ")
	if !strings.Contains(fast, cite) {
		t.Errorf("skills/flow-fast/SKILL.md section 4 lacks %q", cite)
	}
}
