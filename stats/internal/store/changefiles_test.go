package store_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/store"
)

// TestSaveChangeFileUpsertsOnPath pins the upsert on (change, path): a
// second save of the same path replaces the body rather than adding a row,
// and an unknown change is ErrChangeNotFound.
func TestSaveChangeFileUpsertsOnPath(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-change-files-upsert-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1073")

	if err := st.SaveChangeFile(ctx, projectKey, "kan-1073", "tasks.md", "first\n"); err != nil {
		t.Fatalf("first SaveChangeFile: %v", err)
	}
	if err := st.SaveChangeFile(ctx, projectKey, "kan-1073", "tasks.md", "second\n"); err != nil {
		t.Fatalf("second SaveChangeFile: %v", err)
	}
	body, err := st.GetChangeFile(ctx, projectKey, "kan-1073", "tasks.md")
	if err != nil {
		t.Fatalf("GetChangeFile: %v", err)
	}
	if body != "second\n" {
		t.Errorf("body = %q, want the second save's body", body)
	}
	var rows int
	if err := store.PoolOf(st).QueryRow(ctx, `
		SELECT count(*) FROM change_files cf JOIN changes c ON c.id = cf.change_id
		WHERE c.project_key = $1 AND c.name = $2`, projectKey, "kan-1073").Scan(&rows); err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if rows != 1 {
		t.Errorf("rows = %d, want 1 -- a second save of one path must upsert", rows)
	}

	if err := st.SaveChangeFile(ctx, projectKey, "no-such-change", "tasks.md", "x"); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("SaveChangeFile on an unknown change = %v, want ErrChangeNotFound", err)
	}
}

// TestGetChangeFileNotFound pins the two absences a read tells apart: a
// known change with no such file is ErrChangeFileNotFound, an unknown change
// is ErrChangeNotFound.
func TestGetChangeFileNotFound(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-change-files-missing-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1073")

	if _, err := st.GetChangeFile(ctx, projectKey, "kan-1073", "tasks.md"); !errors.Is(err, store.ErrChangeFileNotFound) {
		t.Errorf("GetChangeFile of an unsaved path = %v, want ErrChangeFileNotFound", err)
	}
	if _, err := st.GetChangeFile(ctx, projectKey, "no-such-change", "tasks.md"); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("GetChangeFile on an unknown change = %v, want ErrChangeNotFound", err)
	}
}

// TestHasChangeFilesFalseUntilSaved pins the cleanup gate's read: false for
// a change with no saved file, true after one save, and ErrChangeNotFound
// for a change the store has never heard of.
func TestHasChangeFilesFalseUntilSaved(t *testing.T) {
	st := newTestStore(t)
	ctx := context.Background()
	projectKey := fmt.Sprintf("proj-change-files-has-%d", time.Now().UnixNano())
	seedChange(t, st, projectKey, "kan-1073")

	has, err := st.HasChangeFiles(ctx, projectKey, "kan-1073")
	if err != nil || has {
		t.Fatalf("HasChangeFiles before a save = %v, %v; want false, nil", has, err)
	}
	if err := st.SaveChangeFile(ctx, projectKey, "kan-1073", "self-review-context.md", "b"); err != nil {
		t.Fatalf("SaveChangeFile: %v", err)
	}
	has, err = st.HasChangeFiles(ctx, projectKey, "kan-1073")
	if err != nil || !has {
		t.Errorf("HasChangeFiles after a save = %v, %v; want true, nil", has, err)
	}
	if _, err := st.HasChangeFiles(ctx, projectKey, "no-such-change"); !errors.Is(err, store.ErrChangeNotFound) {
		t.Errorf("HasChangeFiles on an unknown change = %v, want ErrChangeNotFound", err)
	}
}
