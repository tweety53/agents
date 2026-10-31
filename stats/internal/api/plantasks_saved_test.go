package api

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
	"github.com/tweety53/agents/stats/internal/storetest"
)

// savedPlanStore is a real store holding change name, finished, with
// tasks.md saved in change_files as flow change save leaves it.
func savedPlanStore(t *testing.T, name, tasks string) (*statsHandler, string) {
	t.Helper()
	st := storetest.NewStore(t)
	ctx := context.Background()
	project := fmt.Sprintf("proj-saved-plan-%d", time.Now().UnixNano())
	if err := st.PutChange(ctx, store.Change{
		ProjectKey: project, MainCheckoutPath: t.TempDir(), Name: name,
		State: store.StateFinished, UpdatedAt: time.Now(), UpdatedBy: "test",
	}); err != nil {
		t.Fatalf("seed change: %v", err)
	}
	if err := st.SaveChangeFile(ctx, project, name, "tasks.md", tasks); err != nil {
		t.Fatalf("save tasks.md: %v", err)
	}
	return &statsHandler{store: st, changeFiles: st, logger: slog.New(slog.DiscardHandler)}, project
}

// TestPlanTasksFinishedReadsTheSavedPlan pins that a finished change whose
// worktrees are gone reads its plan from change_files, never git history.
func TestPlanTasksFinishedReadsTheSavedPlan(t *testing.T) {
	h, project := savedPlanStore(t, "kan-1", "- [x] 1. One\n- [ ] 2. Two\n")
	row := store.LiveStateRow{ProjectKey: project, Name: "kan-1", State: store.StateFinished, RepoRoots: []string{filepath.Join(t.TempDir(), "gone")}}
	got := planTasks(row, h.savedFile(context.Background()))
	if len(got) != 2 || got[0] != (taskDTO{N: 1, Title: "One", State: "done"}) || got[1] != (taskDTO{N: 2, Title: "Two", State: "pending"}) {
		t.Fatalf("finished planTasks = %+v", got)
	}
	row.Name = "kan-2"
	if got := planTasks(row, h.savedFile(context.Background())); got != nil {
		t.Errorf("unsaved change planTasks = %+v, want nil", got)
	}
}

// TestPlanTasksPrefersTheWorktreeOverTheSavedPlan pins the read order: a
// live tasks.md on disk wins over the saved copy.
func TestPlanTasksPrefersTheWorktreeOverTheSavedPlan(t *testing.T) {
	h, project := savedPlanStore(t, "kan-1", "- [x] 1. Saved\n")
	root := t.TempDir()
	dir := filepath.Join(root, "spectre", "changes", "kan-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] 1. Live\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	row := store.LiveStateRow{ProjectKey: project, Name: "kan-1", State: store.StateFinished, RepoRoots: []string{root}}
	if got := planTasks(row, h.savedFile(context.Background())); len(got) != 1 || got[0].Title != "Live" {
		t.Fatalf("planTasks = %+v, want the worktree's Live", got)
	}
}
