package plans

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

func TestDiffGroups(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	diff := filepath.Join(t.TempDir(), "final-review.diff")
	body := "# worktree: /w — merge base abc\n" +
		"diff --git a/pkg/foo.go b/pkg/foo.go\n--- a/pkg/foo.go\n+++ b/pkg/foo.go\n+a\n+b\n-c\n" +
		"diff --git a/tests/foo_test.go b/tests/foo_test.go\n--- a/tests/foo_test.go\n+++ b/tests/foo_test.go\n+t\n" +
		"diff --git a/pkg/README.md b/pkg/README.md\n--- a/pkg/README.md\n+++ b/pkg/README.md\n+d\n" +
		"diff --git a/ops/x.yml b/ops/x.yml\n--- a/ops/x.yml\n+++ b/ops/x.yml\n+y\n"
	if err := os.WriteFile(diff, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	r := runGuard("diff-groups", []string{wt, diff}, core.Env{})
	if r.Rc != 0 {
		t.Fatalf("rc %d: %s", r.Rc, r.Out)
	}
	got, _ := os.ReadFile(strings.TrimSuffix(diff, ".diff") + ".groups.md")
	want := "## docs and config\n\n- /w/ops/x.yml (+1 -0)\n\n" +
		"## pkg\n\n- /w/pkg/README.md (+1 -0)\n- /w/pkg/foo.go (+2 -1)\n- /w/tests/foo_test.go (+1 -0)\n\n"
	if string(got) != want {
		t.Errorf("groups:\n%s\nwant:\n%s", got, want)
	}
	if r := runGuard("diff-groups", []string{wt, "/nope"}, core.Env{}); r.Rc != 2 || !strings.Contains(r.Out, "diff-groups:") {
		t.Errorf("missing diff: rc %d", r.Rc)
	}
}

func TestBuildSlotPromptNamesDiffGroups(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../../skills/flow")
	wt := t.TempDir()
	sdd := filepath.Join(wt, ".superpowers/sdd")
	_ = os.MkdirAll(sdd, 0o755)
	_ = os.WriteFile(filepath.Join(sdd, "final-review.groups.md"), []byte("## x\n"), 0o644)
	diff := filepath.Join(sdd, "final-review.diff")
	r := runGuard("build-slot-prompt", []string{skill, wt, "1", "mutation", "-diff", diff, "-copies", "/w", "-panel", "mutation"}, core.Env{})
	if r.Rc != 0 {
		t.Fatalf("rc %d: %s", r.Rc, r.Out)
	}
	got, _ := os.ReadFile(filepath.Join(sdd, "slot-prompt-1-mutation.md"))
	if !strings.Contains(string(got), "**Diff groups:** "+filepath.Join(sdd, "final-review.groups.md")) {
		t.Errorf("prompt lacks the Diff groups line")
	}
}
