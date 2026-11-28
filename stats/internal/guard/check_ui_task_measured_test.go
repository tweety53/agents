package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// A task whose Files touch `ui paths` closes only over a measure record that
// exists, is no older than the task's last source commit and carries no
// flagged row; the refusal prints the command, and nothing is ticked.
func TestCloseTaskRefusesAnUnmeasuredUITask(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, true, "0")
	guardtest.WriteFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `web/**` |\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. screen\n\n**Files:** `web/a.css`\n**Tests:** none\n**After:** none\n**Size:** small\n\n- [ ] 2. engine\n\n**Files:** `core/b.go`\n**Tests:** none\n**After:** none\n**Size:** small\n")
	s1 := fx.commit(t, fx.wt, "one", "1", "web/a.css", "a\n")
	s2 := fx.commit(t, fx.wt, "two", "2", "core/b.go", "b\n")
	record := fx.wt + "/spectre/changes/demo/visual-verification/task-1-measure.json"
	command := "run: flow-guard visual-verify-pass measure " + fx.wt + " demo task-1 "
	refused := func(what, reason string) {
		t.Helper()
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
		if r.Rc != 1 || !strings.Contains(r.Err, "UI-TASK-UNMEASURED: 1") || !strings.Contains(r.Err, reason) || !strings.Contains(r.Err, command) {
			t.Fatalf("%s: exit %d, want 1 naming %q and the command\n%s", what, r.Rc, reason, r.Out+r.Err)
		}
		ctEqual(t, fx.calls(t), nil)
	}

	refused("no record", "no measure record")
	clean := `{"pairs":[{"frame":"frame-a","exits":{"compose":0}}]}`
	guardtest.WriteFile(t, record, clean)
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(record, old, old); err != nil {
		t.Fatal(err)
	}
	refused("a record older than the last commit", "older than the task's last source commit")
	guardtest.WriteFile(t, record, `{"pairs":[{"frame":"frame-a","exits":{"compose":0},"over":[{"row":3}]}]}`)
	refused("a record with a flagged row", "flagged rows remain (frame-a 1)")
	// An unmeasurable row is not a pass: it counts toward the flagged total.
	guardtest.WriteFile(t, record, `{"pairs":[{"frame":"frame-a","exits":{"compose":0},"unmeasurable":[{"row":4,"reason":"empty rect"}]}]}`)
	refused("a record with an unmeasurable row", "flagged rows remain (frame-a 1)")

	// The task that touches no `ui paths` file closes with no record at all.
	if r := fx.run(fx.wt, "demo", fx.base, "2:"+s2); r.Rc != 0 {
		t.Fatalf("a non-UI task: exit %d, want 0\n%s", r.Rc, r.Out+r.Err)
	}
	guardtest.WriteFile(t, record, clean)
	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.Rc != 0 {
		t.Fatalf("a clean record: exit %d, want 0\n%s", r.Rc, r.Out+r.Err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 0 || !strings.Contains(r.Out, "UI-TASK-MEASURED: 1") {
		t.Fatalf("check-ui-task-measured on a clean record: exit %d\n%s", r.Rc, r.Out+r.Err)
	}
	if err := os.Remove(record); err != nil {
		t.Fatal(err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 1 || !strings.Contains(r.Err, command) {
		t.Fatalf("check-ui-task-measured with no record: exit %d\n%s", r.Rc, r.Out+r.Err)
	}
}

// A cross-repository task's file is matched as `<repository name>/<file>`,
// the way `ui paths` names another repository's files.
func TestUITaskMeasuredMatchesAPeerRepositorysFile(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, false, "0").withPeer(t)
	guardtest.WriteFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `"+filepath.Base(fx.peer)+"/**` |\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. screen\n\n**Files:** `src/a.css`\n**Tests:** none\n**After:** none\n**Size:** small\n")
	fx.commit(t, fx.peer, "one", "1", "src/a.css", "a\n")
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 0 {
		t.Fatalf("without the peer worktree: exit %d, want 0 (the file is no repository's)\n%s", r.Rc, r.Out+r.Err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1", fx.peer}, fx.env); r.Rc != 1 {
		t.Fatalf("with the peer worktree: exit %d, want 1\n%s", r.Rc, r.Out+r.Err)
	}
}

// A sample or fake file under `ui paths` draws no frame: its task is no UI
// task, while a sibling task with a real ui file still is.
func TestUITaskMeasuredSkipsSamplesAndFakes(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, false, "0")
	guardtest.WriteFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `web/**` |\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. samples\n\n**Files:** `web/WorkoutSamples.kt`, `web/FakeApi.kt`\n**Tests:** none\n**After:** none\n**Size:** small\n\n- [ ] 2. screen\n\n**Files:** `web/Screen.kt`\n**Tests:** none\n**After:** none\n**Size:** small\n")
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 0 {
		t.Fatalf("samples/fake task: exit %d, want 0\n%s", r.Rc, r.Out+r.Err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "2"}, fx.env); r.Rc != 1 {
		t.Fatalf("real ui task: exit %d, want 1\n%s", r.Rc, r.Out+r.Err)
	}
}

func utmCaptureLaterPlan(t *testing.T, laterFiles string) *ctFx {
	t.Helper()
	fx := ctFixture(t, true, "0")
	guardtest.WriteFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `web/**` |\n")
	guardtest.WriteFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. screen\n\n**Files:** `web/a.css`\n**Tests:** none\n**After:** none\n**Size:** small\n\n- [ ] 2. capture\n\n**Files:** "+laterFiles+"\n**Tests:** none\n**After:** none\n**Size:** small\n")
	fx.commit(t, fx.wt, "one", "1", "web/a.css", "a\n")
	return fx
}

func TestUITaskMeasuredWaitsForALaterCaptureSpec(t *testing.T) {
	t.Parallel()
	fx := utmCaptureLaterPlan(t, "`e2e/x.spec.ts`, `e2e/x.spec.ts.mockups`")
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 0 || !strings.Contains(r.Out, "adds its capture spec in task 2") {
		t.Fatalf("later capture task: exit %d\n%s", r.Rc, r.Out+r.Err)
	}
	// The capture task itself is not exempted by its own sidecar.
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "2"}, fx.env); strings.Contains(r.Out, "adds its capture spec") {
		t.Fatalf("capture task exempted itself\n%s", r.Out+r.Err)
	}
}

func TestUITaskMeasuredWithoutALaterCaptureSpecStaysUnmeasured(t *testing.T) {
	t.Parallel()
	fx := utmCaptureLaterPlan(t, "`core/b.go`")
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.Rc != 1 || !strings.Contains(r.Err, "UI-TASK-UNMEASURED: 1") {
		t.Fatalf("no later capture task: exit %d\n%s", r.Rc, r.Out+r.Err)
	}
}
