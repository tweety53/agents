package api

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestPlanTasks pins how a board row's plan reads: from the first worktree
// holding the change's tasks.md, a task done when ticked or committed, in
// progress when a running subagent works on it, and none for a finished change.
func TestPlanTasks(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "spectre", "changes", "kan-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	md := "# Tasks\n\n- [x] 1. Ticked\n- [ ] 2. Committed\n  - [ ] a step\n- [ ] 3. Running\n- [ ] 4. Waiting\n"
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(md), 0o644); err != nil {
		t.Fatal(err)
	}
	three := "3"
	row := store.LiveStateRow{
		Name: "kan-1", State: store.StateInProgress,
		RepoRoots:      []string{t.TempDir(), root},
		CommittedTasks: []string{"2"},
		Subagents:      []store.ActiveDispatch{{Role: "implementer", TaskID: &three}},
	}
	got := planTasks(row)
	want := []taskDTO{{1, "Ticked", "done"}, {2, "Committed", "done"}, {3, "Running", "in progress"}, {4, "Waiting", "pending"}}
	if len(got) != len(want) {
		t.Fatalf("planTasks = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("task %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	row.State = store.StateFinished
	if got := planTasks(row); got != nil {
		t.Errorf("finished planTasks = %+v, want nil", got)
	}
}
