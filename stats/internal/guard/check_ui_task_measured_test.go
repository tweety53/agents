package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// A task whose Files touch `ui paths` closes only over a measure record that
// exists, is no older than the task's last source commit and carries no
// flagged row; the refusal prints the command, and nothing is ticked or pushed.
func TestCloseTaskRefusesAnUnmeasuredUITask(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, true, "0")
	writeFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `web/**` |\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. screen\n\n**Files:** `web/a.css`\n**Tests:** none\n**After:** none\n**Size:** small\n\n- [ ] 2. engine\n\n**Files:** `core/b.go`\n**Tests:** none\n**After:** none\n**Size:** small\n")
	s1 := fx.commit(t, fx.wt, "one", "1", "web/a.css", "a\n")
	s2 := fx.commit(t, fx.wt, "two", "2", "core/b.go", "b\n")
	record := fx.wt + "/spectre/changes/demo/visual-verification/task-1-measure.json"
	command := "run: flow-guard visual-verify-pass measure " + fx.wt + " demo task-1 "
	refused := func(what, reason string) {
		t.Helper()
		r := fx.run(fx.wt, "demo", fx.base, "1:"+s1)
		if r.rc != 1 || !strings.Contains(r.err, "UI-TASK-UNMEASURED: 1") || !strings.Contains(r.err, reason) || !strings.Contains(r.err, command) {
			t.Fatalf("%s: exit %d, want 1 naming %q and the command\n%s", what, r.rc, reason, r.out+r.err)
		}
		ctEqual(t, fx.calls(t), nil)
	}

	refused("no record", "no measure record")
	clean := `{"pairs":[{"frame":"frame-a","exits":{"compose":0}}]}`
	writeFile(t, record, clean)
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(record, old, old); err != nil {
		t.Fatal(err)
	}
	refused("a record older than the last commit", "older than the task's last source commit")
	writeFile(t, record, `{"pairs":[{"frame":"frame-a","exits":{"compose":0},"over":[{"row":3}]}]}`)
	refused("a record with a flagged row", "flagged rows remain (frame-a 1)")

	// The task that touches no `ui paths` file closes with no record at all.
	if r := fx.run(fx.wt, "demo", fx.base, "2:"+s2); r.rc != 0 {
		t.Fatalf("a non-UI task: exit %d, want 0\n%s", r.rc, r.out+r.err)
	}
	writeFile(t, record, clean)
	if r := fx.run(fx.wt, "demo", fx.base, "1:"+s1); r.rc != 0 {
		t.Fatalf("a clean record: exit %d, want 0\n%s", r.rc, r.out+r.err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.rc != 0 || !strings.Contains(r.out, "UI-TASK-MEASURED: 1") {
		t.Fatalf("check-ui-task-measured on a clean record: exit %d\n%s", r.rc, r.out+r.err)
	}
	if err := os.Remove(record); err != nil {
		t.Fatal(err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.rc != 1 || !strings.Contains(r.err, command) {
		t.Fatalf("check-ui-task-measured with no record: exit %d\n%s", r.rc, r.out+r.err)
	}
}

// A cross-repository task's file is matched as `<repository name>/<file>`,
// the way `ui paths` names another repository's files.
func TestUITaskMeasuredMatchesAPeerRepositorysFile(t *testing.T) {
	t.Parallel()
	fx := ctFixture(t, false, "0").withPeer(t)
	writeFile(t, fx.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | `"+filepath.Base(fx.peer)+"/**` |\n")
	writeFile(t, fx.wt+"/spectre/changes/demo/tasks.md", "# demo\n\n- [ ] 1. screen\n\n**Files:** `src/a.css`\n**Tests:** none\n**After:** none\n**Size:** small\n")
	fx.commit(t, fx.peer, "one", "1", "src/a.css", "a\n")
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1"}, fx.env); r.rc != 0 {
		t.Fatalf("without the peer worktree: exit %d, want 0 (the file is no repository's)\n%s", r.rc, r.out+r.err)
	}
	if r := runGuard("check-ui-task-measured", []string{fx.wt, "1", fx.peer}, fx.env); r.rc != 1 {
		t.Fatalf("with the peer worktree: exit %d, want 1\n%s", r.rc, r.out+r.err)
	}
}
