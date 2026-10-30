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
	got := planTasks(row, false)
	want := []taskDTO{{N: 1, Title: "Ticked", State: "done"}, {N: 2, Title: "Committed", State: "done"}, {N: 3, Title: "Running", State: "in progress"}, {N: 4, Title: "Waiting", State: "pending"}}
	if len(got) != len(want) {
		t.Fatalf("planTasks = %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("task %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	row.State = store.StateFinished
	if got := planTasks(row, false); got != nil {
		t.Errorf("finished planTasks = %+v, want nil", got)
	}
}

// TestPlanTasksArchived pins that the plan still reads once integrate's run 1
// archived it to changes/archive/<name>/tasks.md.
func TestPlanTasksArchived(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "spectre", "changes", "archive", "kan-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [x] 1. One\n- [ ] 2. Two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := planTasks(store.LiveStateRow{Name: "kan-1", State: store.StateInProgress, RepoRoots: []string{root}}, false)
	if len(got) != 2 || got[0] != (taskDTO{N: 1, Title: "One", State: "done"}) || got[1] != (taskDTO{N: 2, Title: "Two", State: "pending"}) {
		t.Fatalf("archived planTasks = %+v", got)
	}
}

// TestSelfReviewTasks pins the self-review plan's read: self-review-<name>'s
// tasks.md, archived or not, done only when ticked.
func TestSelfReviewTasks(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "spectre", "changes", "archive", "self-review-kan-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [x] 1. Fix A\n- [ ] 2. Fix B\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	two := "2"
	row := store.LiveStateRow{
		Name: "kan-1", State: store.StateInProgress, RepoRoots: []string{root},
		CommittedTasks: []string{"2"},
		Subagents:      []store.ActiveDispatch{{Role: "implementer", TaskID: &two}},
	}
	got := selfReviewTasks(row, false)
	if len(got) != 2 || got[0] != (taskDTO{N: 1, Title: "Fix A", State: "done"}) || got[1] != (taskDTO{N: 2, Title: "Fix B", State: "pending"}) {
		t.Fatalf("selfReviewTasks = %+v", got)
	}
	if planTasks(row, false) != nil {
		t.Errorf("planTasks read the self-review plan as the change's own")
	}
}

// TestSelfReviewTasksFromSubRun pins that the plan is read from the
// self-review sub-run's own worktree, which is not one of the change's.
func TestSelfReviewTasksFromSubRun(t *testing.T) {
	sub := t.TempDir()
	dir := filepath.Join(sub, "spectre", "changes", "self-review-kan-1")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [ ] 1. Fix A\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	row := store.LiveStateRow{Name: "kan-1", State: store.StateInProgress, RepoRoots: []string{t.TempDir()}, SelfReviewRoots: []string{sub}}
	if got := selfReviewTasks(row, false); len(got) != 1 || got[0] != (taskDTO{N: 1, Title: "Fix A", State: "pending"}) {
		t.Fatalf("selfReviewTasks = %+v", got)
	}
}

// TestFinishedPlanReadOnlyWhenScoped pins that a finished change's plan --
// archived in its main checkout once cleanup removed the worktrees -- is read
// for a board scoped to that change and never for the whole board.
func TestFinishedPlanReadOnlyWhenScoped(t *testing.T) {
	main := t.TempDir()
	for _, name := range []string{"kan-1", "self-review-kan-1"} {
		dir := filepath.Join(main, "spectre", "changes", "archive", name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte("- [x] 1. One\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	row := store.LiveStateRow{Name: "kan-1", State: store.StateFinished, RepoRoots: []string{filepath.Join(main, "gone-worktree"), main}}
	if got := planTasks(row, true); len(got) != 1 || got[0] != (taskDTO{N: 1, Title: "One", State: "done"}) {
		t.Fatalf("scoped planTasks = %+v", got)
	}
	if got := selfReviewTasks(row, true); len(got) != 1 {
		t.Fatalf("scoped selfReviewTasks = %+v", got)
	}
	if planTasks(row, false) != nil || selfReviewTasks(row, false) != nil {
		t.Errorf("the whole board read a finished change's plan")
	}
	dtos := toStateBoardDTOs([]store.LiveStateRow{row}, true)
	if len(dtos[0].Tasks) != 1 || len(dtos[0].SelfReviewTasks) != 1 {
		t.Errorf("scoped board DTO = %+v", dtos[0])
	}
	if dtos := toStateBoardDTOs([]store.LiveStateRow{row}, false); dtos[0].Tasks != nil {
		t.Errorf("whole board DTO carries a finished plan: %+v", dtos[0].Tasks)
	}
}
