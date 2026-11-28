package guard

import (
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// The change directory is gitignored (KAN-1073): the guards that once read a
// change from git — a committed directory, a drop or archive subject — read
// flowd's change records and the worktrees on disk instead, and cleanup
// refuses to remove a worktree holding the only copy of a plan flowd has not
// saved.

// TestPipelineFixHighestKReadsFlowdChanges pins pfHighestK's sources: a
// change flowd records (FINISHED included) and a directory under
// <main>-worktrees/ count, and an answer from flowd's fallback journal
// refuses exit 2.
func TestPipelineFixHighestKReadsFlowdChanges(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	guardtest.WriteFile(t, fx.states, `{"source":"store","complete":true,"records":[{"name":"pipeline-fix-x-9","state":"FINISHED"}]}`)
	guardtest.Mkdir(t, fx.repo+"-worktrees/pipeline-fix-y-11")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "3"); rc != 1 || !strings.Contains(out, "<k> is 12") {
		t.Fatalf("rc %d, want the next key 12 — one past the worktree's 11, flowd's 9 below it: %s", rc, out)
	}
	if !strings.Contains(fx.calls(t), "state list -C "+fx.repo) {
		t.Fatalf("flowd's change records were never asked for: %s", fx.calls(t))
	}
	guardtest.WriteFile(t, fx.states, `{"source":"fallback","complete":false,"records":[]}`)
	if rc, out := fx.pf(t, "start", "kan-1-demo", "12"); rc != 2 || !strings.Contains(out, "could not be read in full") {
		t.Fatalf("rc %d, want 2 on a fallback answer: %s", rc, out)
	}
}
