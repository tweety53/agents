package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// TestBuildSlotPromptGolden pins every slot's built prompt byte for byte — each reviewer template on
// a fix round with every flag set, and mutation — against testdata/build-slot-prompt/<id>.golden,
// with the skill, worktree and diff directories written as <SKILL>, <WT> and <DIFF>.
func TestBuildSlotPromptGolden(t *testing.T) {
	t.Parallel()
	skill, _ := filepath.Abs("../../../skills/flow")
	for _, id := range []string{"primary", "principles", "failure-modes", "security", "test-sanity", "call-efficiency", "mutation"} {
		wt := t.TempDir()
		d := bspDiffDir(t)
		args := []string{skill, wt, "2", id, "-diff", d + "/fix-round-2.diff", "-panel", "primary," + id}
		if id == "mutation" {
			args = append(args, "-copies", "/w-mutation-2", "-tests", "/w/foo_test.go")
		} else {
			args = append(args, "-artifacts", "/c/proposal.md,/c/design.md", "-standards", "/s/std.md",
				"-global-constraints", "/g/gc.md", "-worktrees", "/w,/f", "-repo", "gymie", "-fix-report", "/r/panel-fix-report-2.md")
		}
		r := runGuard("build-slot-prompt", args, core.Env{})
		if r.rc != 0 {
			t.Fatalf("%s: rc %d: %s", id, r.rc, r.out)
		}
		got, _ := os.ReadFile(strings.TrimSpace(r.stdout))
		s := strings.NewReplacer(skill, "<SKILL>", wt, "<WT>", d, "<DIFF>").Replace(string(got))
		golden := filepath.Join("testdata/build-slot-prompt", id+".golden")
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatal(err)
		}
		if s != string(want) {
			t.Errorf("%s: built prompt differs from %s", id, golden)
		}
	}
}
