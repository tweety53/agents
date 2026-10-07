package fallback_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/fallback"
)

// TestAppendJournalEntrySurvivesASlowRetire is the deterministic
// reproducer for the append-vs-retire loss that
// internal/reconcile's TestConcurrentAppendVersusRetirePreservesEveryEntry
// only catches under load. It plays internal/reconcile's retirePrefix by
// hand -- take the sidecar lock, re-read the journal, write the retained
// remainder to a temp file, rename it over the journal -- with that
// section held open for holdFor, standing in for an F_FULLFSYNC that runs
// slow under load, while an append starts inside it.
//
// An append that gives up on the lock and writes anyway lands on the inode
// the rename then unlinks: its entry is gone, with no error anywhere. The
// append must instead wait the retire out, then write to the journal the
// rename put in place.
func TestAppendJournalEntrySurvivesASlowRetire(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "held-proj", "held-chg.journal")
	body := func(by string) []byte {
		return []byte(`{"state":"STARTED","mainCheckoutPath":"/tmp/lock-test","updatedAt":"2026-08-13T10:00:00Z","updatedBy":"` + by + `"}`)
	}

	if err := fallback.AppendJournalEntry(path, "held-proj", "held-chg", body("replayed"), time.Now()); err != nil {
		t.Fatalf("seed AppendJournalEntry: %v", err)
	}

	// The retirer's critical section: lock, re-read, then (after the
	// append below has started) write the remainder and rename it in.
	unlock, err := fallback.LockJournal(path)
	if err != nil {
		t.Fatalf("LockJournal: %v", err)
	}
	current, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("re-read journal: %v", err)
	}
	consumed := len(current) // the seed entry, replayed and now retired

	done := make(chan error, 1)
	go func() {
		done <- fallback.AppendJournalEntry(path, "held-proj", "held-chg", body("appended"), time.Now())
	}()

	const holdFor = 300 * time.Millisecond
	select {
	case err := <-done:
		t.Errorf("AppendJournalEntry returned (err=%v) while the retire lock was held -- it wrote without the lock", err)
		done <- err
	case <-time.After(holdFor):
	}

	tmp := path + ".tmp-test"
	if err := os.WriteFile(tmp, current[consumed:], 0o644); err != nil {
		t.Fatalf("write temp journal: %v", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		t.Fatalf("rename temp journal: %v", err)
	}
	unlock()

	if err := <-done; err != nil {
		t.Fatalf("AppendJournalEntry: %v", err)
	}

	entries, err := fallback.ReadJournalEntries(path)
	if err != nil {
		t.Fatalf("ReadJournalEntries: %v", err)
	}
	if len(entries) != 1 || string(entries[0].Body) != string(body("appended")) {
		t.Fatalf("journal after retire = %+v, want exactly the appended entry -- it was lost to the retire's rename", entries)
	}
}
